package tui

import (
	"context"
	"errors"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"charm.land/fantasy"
	"github.com/JiayuXu0/MiniCode/internal/permission"
)

// Update 实现 tea.Model 接口，处理所有消息
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// 如果有待确认的权限请求，优先处理
		if m.permPending != nil {
			return m.handlePermissionKeyMsg(msg)
		}
		return m.handleKeyMsg(msg)

	case tea.MouseMsg:
		return m.handleMouseMsg(msg)

	case permission.RequestMsg:
		// 收到权限请求，显示对话框
		m.permPending = msg.Req
		m.permFocusIndex = 0
		return m, nil

	case streamTextMsg:
		return m.handleStreamTextMsg(msg)

	case streamReasoningMsg:
		return m.handleStreamReasoningMsg(msg)

	case streamToolCallMsg:
		return m.handleStreamToolCallMsg(msg)

	case streamToolResultMsg:
		return m.handleStreamToolResultMsg(msg)

	case streamTokenUpdateMsg:
		m.totalTokens = msg.tokens
		return m, nil

	case streamDoneMsg:
		return m.handleStreamDoneMsg(msg)

	case tickMsg:
		return m.handleTickMsg()

	case tea.WindowSizeMsg:
		return m.handleWindowSizeMsg(msg)
	}

	// 默认传给 textarea
	var cmd tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)
	return m, cmd
}

// handleKeyMsg 处理键盘事件
func (m *Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Ctrl+C 直接退出
	if msg.Type == tea.KeyCtrlC {
		return m, tea.Quit
	}

	// Esc 键：流式中取消，否则退出
	if msg.Type == tea.KeyEsc {
		if m.streaming && m.cancelFunc != nil {
			m.cancelFunc()
			m.streaming = false
			m.waiting = false
			m.cancelFunc = nil
			m.streamParts = nil
			m.viewport.SetContent(m.renderMessages())
			return m, nil
		}
		return m, tea.Quit
	}

	// Tab 键切换焦点
	if msg.Type == tea.KeyTab {
		if m.focus == focusTextarea {
			m.focus = focusViewport
			m.textarea.Blur()
		} else {
			m.focus = focusTextarea
			m.textarea.Focus()
		}
		return m, nil
	}

	// Ctrl+T 切换主题
	if msg.Type == tea.KeyCtrlT {
		m.toggleTheme()
		m.viewport.SetContent(m.renderMessages())
		return m, nil
	}

	// viewport 模式下的按键处理
	if m.focus == focusViewport {
		switch msg.Type {
		case tea.KeyUp, tea.KeyPgUp:
			m.viewport.LineUp(1)
			return m, nil
		case tea.KeyDown, tea.KeyPgDown:
			m.viewport.LineDown(1)
			return m, nil
		case tea.KeyHome:
			m.viewport.GotoTop()
			return m, nil
		case tea.KeyEnd:
			m.viewport.GotoBottom()
			return m, nil
		}
		return m, nil
	}

	// textarea 模式下的 Enter 键处理
	if msg.Type == tea.KeyEnter && !msg.Alt {
		userInput := strings.TrimSpace(m.textarea.Value())
		if userInput == "" || m.waiting {
			return m, nil
		}

		// 添加用户消息到历史
		m.history = append(m.history, fantasy.NewUserMessage(userInput))
		m.textarea.Reset()
		m.waiting = true
		m.streaming = true
		m.streamParts = nil
		m.totalTokens = 0
		m.viewport.SetContent(m.renderMessages())
		m.viewport.GotoBottom()

		return m, tea.Batch(m.startStream(userInput), m.tick())
	}

	// 其他按键传给 textarea
	if !m.waiting {
		var cmd tea.Cmd
		m.textarea, cmd = m.textarea.Update(msg)
		return m, cmd
	}
	return m, nil
}

// handleMouseMsg 处理鼠标事件
func (m *Model) handleMouseMsg(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// handleStreamTextMsg 处理流式文本消息
func (m *Model) handleStreamTextMsg(msg streamTextMsg) (tea.Model, tea.Cmd) {
	m.appendStreamContent(partTypeText, msg.delta)
	return m, nil
}

// handleStreamReasoningMsg 处理流式思考消息
func (m *Model) handleStreamReasoningMsg(msg streamReasoningMsg) (tea.Model, tea.Cmd) {
	m.appendStreamContent(partTypeReasoning, msg.delta)
	return m, nil
}

// appendStreamContent 追加流式内容（合并相同类型的连续片段）
func (m *Model) appendStreamContent(pType streamPartType, delta string) {
	if len(m.streamParts) > 0 && m.streamParts[len(m.streamParts)-1].partType == pType {
		m.streamParts[len(m.streamParts)-1].content += delta
	} else {
		m.streamParts = append(m.streamParts, streamPart{
			partType: pType,
			content:  delta,
		})
	}
	m.needsRefresh = true
}

// handleStreamToolCallMsg 处理工具调用消息
func (m *Model) handleStreamToolCallMsg(msg streamToolCallMsg) (tea.Model, tea.Cmd) {
	m.streamParts = append(m.streamParts, streamPart{
		partType: partTypeToolCall,
		tool: &toolCallInfo{
			name:   msg.name,
			input:  msg.input,
			status: "running",
		},
	})
	m.needsRefresh = true
	return m, nil
}

// handleStreamToolResultMsg 处理工具结果消息
func (m *Model) handleStreamToolResultMsg(msg streamToolResultMsg) (tea.Model, tea.Cmd) {
	for i := len(m.streamParts) - 1; i >= 0; i-- {
		if m.streamParts[i].partType == partTypeToolCall &&
			m.streamParts[i].tool != nil &&
			m.streamParts[i].tool.name == msg.name &&
			m.streamParts[i].tool.status == "running" {
			m.streamParts[i].tool.output = msg.content
			if msg.isError {
				m.streamParts[i].tool.status = "error"
			} else {
				m.streamParts[i].tool.status = "done"
			}
			break
		}
	}
	m.needsRefresh = true
	return m, nil
}

// handleStreamDoneMsg 处理流式完成消息
func (m *Model) handleStreamDoneMsg(msg streamDoneMsg) (tea.Model, tea.Cmd) {
	m.streaming = false
	m.waiting = false
	m.cancelFunc = nil
	m.totalTokens = msg.totalTokens

	if msg.err != nil {
		// 忽略取消错误，其他错误不处理（历史已经有用户消息了）
		if errors.Is(msg.err, context.Canceled) {
			// 取消时，移除最后添加的用户消息
			if len(m.history) > 0 {
				m.history = m.history[:len(m.history)-1]
			}
		}
	} else if msg.result != nil {
		// 直接把 Agent 返回的消息追加到历史
		for _, step := range msg.result.Steps {
			m.history = append(m.history, step.Messages...)
		}
	}

	m.streamParts = nil
	m.viewport.SetContent(m.renderMessages())
	m.viewport.GotoBottom()
	return m, nil
}

// handleTickMsg 处理定时器消息（批量刷新 + spinner 动画）
func (m *Model) handleTickMsg() (tea.Model, tea.Cmd) {
	if m.streaming {
		m.spinnerFrame = (m.spinnerFrame + 1) % len(spinnerFrames)
		if m.needsRefresh {
			m.viewport.SetContent(m.renderMessages())
			m.viewport.GotoBottom()
			m.needsRefresh = false
		}
		return m, m.tick()
	}
	return m, nil
}

// handleWindowSizeMsg 处理窗口大小变化
func (m *Model) handleWindowSizeMsg(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.width = msg.Width
	m.height = msg.Height

	contentHeight := m.height - m.textarea.Height() - 10
	if contentHeight < 3 {
		contentHeight = 3
	}

	innerWidth := m.width - 6
	if innerWidth < 10 {
		innerWidth = 10
	}

	m.viewport.Width = innerWidth
	m.viewport.Height = contentHeight
	m.textarea.SetWidth(innerWidth)
	m.viewport.SetContent(m.renderMessages())
	return m, nil
}

// handlePermissionKeyMsg 处理权限对话框的按键事件
func (m *Model) handlePermissionKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "left", "h":
		// 向左切换按钮
		m.permFocusIndex = (m.permFocusIndex - 1 + 3) % 3
	case "right", "l", "tab":
		// 向右切换按钮
		m.permFocusIndex = (m.permFocusIndex + 1) % 3
	case "enter":
		// 确认选择
		var resp permission.Response
		switch m.permFocusIndex {
		case 0:
			resp = permission.Granted
		case 1:
			resp = permission.Denied
		case 2:
			resp = permission.Persistent
		}
		m.permService.Respond(m.permPending, resp)
		m.permPending = nil
	case "escape":
		// Esc 键拒绝
		m.permService.Respond(m.permPending, permission.Denied)
		m.permPending = nil
	}
	return m, nil
}
