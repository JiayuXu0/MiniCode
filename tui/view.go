package tui

import (
	"fmt"
	"strings"

	"charm.land/fantasy"
	"github.com/charmbracelet/lipgloss"
)

// View 实现 tea.Model 接口，渲染界面
func (m *Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	title := titleStyle.Render("MiniCode")

	// 根据焦点状态设置边框颜色
	viewportBorder := borderStyle
	inputBorder := borderStyle
	if m.focus == focusViewport {
		viewportBorder = viewportBorder.BorderForeground(colorWarning)
		inputBorder = inputBorder.BorderForeground(colorDim)
	} else {
		viewportBorder = viewportBorder.BorderForeground(colorDim)
		inputBorder = inputBorder.BorderForeground(colorPrimary)
	}

	messagesBox := viewportBorder.
		Width(m.width - 2).
		Height(m.viewport.Height + 2).
		Render(m.viewport.View())

	inputBox := inputBorder.
		Width(m.width - 2).
		Height(m.textarea.Height() + 2).
		Render(m.textarea.View())

	// 状态栏
	status := m.renderStatus()

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		messagesBox,
		inputBox,
		status,
	)
}

// renderStatus 渲染状态栏
func (m *Model) renderStatus() string {
	var status string
	if m.streaming {
		spinner := spinnerFrames[m.spinnerFrame]
		status = fmt.Sprintf("  %s Streaming... | Esc cancel", spinner)
	} else if m.focus == focusViewport {
		status = "  ↑↓ scroll | Tab switch | Ctrl+C quit"
	} else {
		status = "  Enter send | Tab switch | Ctrl+C quit"
	}

	// 始终显示 token 统计
	if m.totalTokens > 0 {
		status += fmt.Sprintf(" | Tokens: %d", m.totalTokens)
	}
	return status
}

// renderMessages 渲染消息列表
func (m *Model) renderMessages() string {
	if len(m.history) == 0 && !m.streaming {
		return placeholderStyle.Render("开始对话吧...")
	}

	var sb strings.Builder

	// 渲染历史消息（从 fantasy.Message 中提取内容）
	for _, msg := range m.history {
		switch msg.Role {
		case fantasy.MessageRoleUser:
			sb.WriteString(userStyle.Render("You: "))
			sb.WriteString(getMessageText(msg))
			sb.WriteString("\n\n")

		case fantasy.MessageRoleAssistant:
			// 提取 reasoning、tool calls、text
			for _, part := range msg.Content {
				switch p := part.(type) {
				case fantasy.ReasoningPart:
					if p.Text != "" {
						sb.WriteString(m.renderThinkingBox(p.Text))
						sb.WriteString("\n")
					}
				case fantasy.ToolCallPart:
					sb.WriteString(m.renderToolBox(p.ToolName, p.Input, "", ""))
					sb.WriteString("\n")
				case fantasy.TextPart:
					if p.Text != "" {
						sb.WriteString(assistantStyle.Render("Assistant: "))
						sb.WriteString(p.Text)
						sb.WriteString("\n\n")
					}
				}
			}

		case fantasy.MessageRoleTool:
			// Tool 结果 - 从历史中查找工具名
			for _, part := range msg.Content {
				if tr, ok := part.(fantasy.ToolResultPart); ok {
					toolName := m.findToolName(tr.ToolCallID)
					sb.WriteString(m.renderToolBox(toolName, "", fmt.Sprintf("%v", tr.Output), "done"))
					sb.WriteString("\n")
				}
			}
		}
	}

	// 渲染当前流式内容（按时间顺序）
	if m.streaming {
		for i, part := range m.streamParts {
			switch part.partType {
			case partTypeReasoning:
				if part.content != "" {
					sb.WriteString(m.renderThinkingBox(part.content))
					sb.WriteString("\n")
				}
			case partTypeToolCall:
				if part.tool != nil {
					sb.WriteString(m.renderToolBox(part.tool.name, part.tool.input, part.tool.output, part.tool.status))
					sb.WriteString("\n")
				}
			case partTypeText:
				if part.content != "" {
					sb.WriteString(assistantStyle.Render("Assistant: "))
					sb.WriteString(part.content)
					// 只在最后一个 text part 显示光标
					if m.isLastTextPart(i) {
						sb.WriteString("█")
					}
					sb.WriteString("\n")
				}
			}
		}
	}

	return sb.String()
}

// getMessageText 从 Message 中提取文本内容
func getMessageText(msg fantasy.Message) string {
	for _, part := range msg.Content {
		if tp, ok := part.(fantasy.TextPart); ok {
			return tp.Text
		}
	}
	return ""
}

// findToolName 根据 ToolCallID 从历史中查找工具名
func (m *Model) findToolName(toolCallID string) string {
	for _, msg := range m.history {
		for _, part := range msg.Content {
			if tc, ok := part.(fantasy.ToolCallPart); ok && tc.ToolCallID == toolCallID {
				return tc.ToolName
			}
		}
	}
	return toolCallID // fallback
}

// isLastTextPart 检查是否是最后一个文本片段
func (m *Model) isLastTextPart(index int) bool {
	for j := index + 1; j < len(m.streamParts); j++ {
		if m.streamParts[j].partType == partTypeText {
			return false
		}
	}
	return true
}

// renderToolBox 统一渲染工具调用方框
// status: "running", "done", "error", 或空字符串（历史记录）
func (m *Model) renderToolBox(name, input, output, status string) string {
	var content strings.Builder
	lineWidth := m.viewport.Width - 10

	// 状态图标 + 工具名
	var header string
	switch status {
	case "running":
		header = fmt.Sprintf("⏳ %s", name)
	case "done":
		header = fmt.Sprintf("✓ %s", name)
	case "error":
		header = fmt.Sprintf("✗ %s", name)
	default:
		header = fmt.Sprintf("Tool: %s", name)
	}
	content.WriteString(toolHeaderStyle.Render(header))

	// 输入参数
	if input != "" {
		content.WriteString("\n")
		content.WriteString(toolContentStyle.Render(truncateLine(input, lineWidth)))
	}

	// 输出结果（限制行数）
	if output != "" {
		content.WriteString("\n")
		lines := strings.Split(output, "\n")
		displayLines := lines
		if len(lines) > toolBoxMaxLines {
			displayLines = lines[:toolBoxMaxLines]
		}
		for i, line := range displayLines {
			content.WriteString(toolContentStyle.Render(truncateLine(line, lineWidth)))
			if i < len(displayLines)-1 {
				content.WriteString("\n")
			}
		}
		if len(lines) > toolBoxMaxLines {
			content.WriteString("\n")
			content.WriteString(toolContentStyle.Render(fmt.Sprintf("... (%d more lines)", len(lines)-toolBoxMaxLines)))
		}
	}

	boxWidth := m.viewport.Width - 4
	if boxWidth < 20 {
		boxWidth = 20
	}
	return toolBoxStyle.Width(boxWidth).Render(content.String())
}

// renderThinkingBox 渲染思考区域（限制最大行数，显示最新内容）
func (m *Model) renderThinkingBox(content string) string {
	// 计算每行最大字符数
	lineWidth := m.viewport.Width - 10
	if lineWidth < 20 {
		lineWidth = 20
	}

	// 将内容按行宽自动换行
	lines := wrapText(content, lineWidth)

	var sb strings.Builder

	// 如果行数超过限制，只显示最后几行
	startLine := 0
	if len(lines) > thinkingMaxLines {
		startLine = len(lines) - thinkingMaxLines
		sb.WriteString(reasoningStyle.Render(fmt.Sprintf("... (%d lines above)\n", startLine)))
	}

	// 渲染可见行
	for i := startLine; i < len(lines); i++ {
		sb.WriteString(reasoningStyle.Render(lines[i]))
		if i < len(lines)-1 {
			sb.WriteString("\n")
		}
	}

	boxWidth := m.viewport.Width - 4
	if boxWidth < 20 {
		boxWidth = 20
	}

	return thinkingBoxStyle.Width(boxWidth).Render(sb.String())
}
