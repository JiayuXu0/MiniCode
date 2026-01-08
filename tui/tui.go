package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"charm.land/fantasy"
)

type chatMessage struct {
	role    string
	content string
}

type agentResponseMsg struct {
	result    *fantasy.AgentResult // 完整的 AgentResult，包含所有 Steps
	userInput string
	err       error
}

type model struct {
	textarea  textarea.Model
	viewport  viewport.Model
	messages  []chatMessage
	waiting   bool
	width     int
	height    int
	agentFunc func(input string, history []fantasy.Message) tea.Cmd
	agentMsgs []fantasy.Message
}

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4"))

	userStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")).
			Bold(true)

	assistantStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#874BFD"))

	reasoningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")).
			Italic(true)

	toolCallStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFA500")).
			Bold(true)

	toolResultStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888"))

	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4"))
)

func NewTUI(agent fantasy.Agent) model {
	ta := textarea.New()
	ta.Placeholder = ""
	ta.Prompt = ""
	ta.ShowLineNumbers = false
	ta.SetHeight(1)
	ta.SetWidth(80)
	ta.Focus()

	vp := viewport.New(80, 20)

	agentFunc := func(input string, history []fantasy.Message) tea.Cmd {
		return func() tea.Msg {
			result, err := agent.Generate(context.Background(), fantasy.AgentCall{
				Messages: history,
				Prompt:   input,
			})
			if err != nil {
				return agentResponseMsg{err: err, userInput: input}
			}
			return agentResponseMsg{result: result, userInput: input}
		}
	}

	return model{
		textarea:  ta,
		viewport:  vp,
		agentFunc: agentFunc,
	}
}

func (m model) Init() tea.Cmd {
	return textarea.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC || msg.Type == tea.KeyEsc {
			return m, tea.Quit
		}

		if msg.Type == tea.KeyEnter && !msg.Alt {
			userInput := strings.TrimSpace(m.textarea.Value())
			if userInput == "" || m.waiting {
				return m, nil
			}

			m.messages = append(m.messages, chatMessage{role: "user", content: userInput})
			m.textarea.Reset()
			m.waiting = true
			m.viewport.SetContent(m.renderMessages())
			m.viewport.GotoBottom()

			msgsCopy := make([]fantasy.Message, len(m.agentMsgs))
			copy(msgsCopy, m.agentMsgs)

			return m, m.agentFunc(userInput, msgsCopy)
		}

		// 其他按键传给 textarea
		if !m.waiting {
			var cmd tea.Cmd
			m.textarea, cmd = m.textarea.Update(msg)
			return m, cmd
		}
		return m, nil

	case agentResponseMsg:
		m.waiting = false
		if msg.err != nil {
			m.messages = append(m.messages, chatMessage{
				role:    "assistant",
				content: fmt.Sprintf("Error: %v", msg.err),
			})
		} else {
			// 遍历所有 step，提取显示内容
			for _, step := range msg.result.Steps {
				// 显示推理过程
				if reasoningText := step.Content.ReasoningText(); reasoningText != "" {
					m.messages = append(m.messages, chatMessage{
						role:    "reasoning",
						content: reasoningText,
					})
				}

				// 显示工具调用
				for _, tc := range step.Content.ToolCalls() {
					m.messages = append(m.messages, chatMessage{
						role:    "tool_call",
						content: fmt.Sprintf("%s(%s)", tc.ToolName, tc.Input),
					})
				}

				// 显示工具结果
				for _, tr := range step.Content.ToolResults() {
					m.messages = append(m.messages, chatMessage{
						role:    "tool_result",
						content: fmt.Sprintf("[%s] %v", tr.ToolName, tr.Result),
					})
				}

				// 显示文本回复
				if text := step.Content.Text(); text != "" {
					m.messages = append(m.messages, chatMessage{
						role:    "assistant",
						content: text,
					})
				}
			}

			// 正确累积历史：用户消息 + 所有 step 的 messages（包含工具调用链）
			m.agentMsgs = append(m.agentMsgs, fantasy.NewUserMessage(msg.userInput))
			for _, step := range msg.result.Steps {
				m.agentMsgs = append(m.agentMsgs, step.Messages...)
			}
		}
		m.viewport.SetContent(m.renderMessages())
		m.viewport.GotoBottom()
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// 计算 viewport 高度：总高度 - title(1) - textarea高度 - textarea边框(2) - viewport边框(2) - status(1) - 额外间距
		contentHeight := m.height - m.textarea.Height() - 10
		if contentHeight < 3 {
			contentHeight = 3
		}

		// 边框占用2个字符(左右各1)，内边距再留2个字符
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

	var cmd tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	title := titleStyle.Render("MiniCode")

	messagesBox := borderStyle.
		Width(m.width - 2).
		Height(m.viewport.Height + 2).
		Render(m.viewport.View())

	inputBox := borderStyle.
		Width(m.width - 2).
		Height(m.textarea.Height() + 2).
		Render(m.textarea.View())

	var status string
	if m.waiting {
		status = "Thinking..."
	} else {
		status = "Enter 发送, Ctrl+C 退出"
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		messagesBox,
		inputBox,
		status,
	)
}

func (m model) renderMessages() string {
	if len(m.messages) == 0 {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")).
			Italic(true).
			Render("开始对话吧...")
	}

	var sb strings.Builder
	for _, msg := range m.messages {
		switch msg.role {
		case "user":
			sb.WriteString(userStyle.Render("You: "))
			sb.WriteString(msg.content)
			sb.WriteString("\n\n")
		case "assistant":
			sb.WriteString(assistantStyle.Render("Assistant: "))
			sb.WriteString(msg.content)
			sb.WriteString("\n\n")
		case "reasoning":
			sb.WriteString(reasoningStyle.Render("💭 Thinking: "))
			sb.WriteString(reasoningStyle.Render(msg.content))
			sb.WriteString("\n\n")
		case "tool_call":
			sb.WriteString(toolCallStyle.Render("🔧 Tool: "))
			sb.WriteString(msg.content)
			sb.WriteString("\n")
		case "tool_result":
			sb.WriteString(toolResultStyle.Render("   → "))
			sb.WriteString(toolResultStyle.Render(msg.content))
			sb.WriteString("\n\n")
		}
	}
	return sb.String()
}
