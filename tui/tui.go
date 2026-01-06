package tui

import (
	"context"
	"fmt"
	"strings"

	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/fantasy"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/glamour"
)

// chatMessage 表示一条聊天消息
type chatMessage struct {
	role    string // "user" 或 "assistant"
	content string
}

// agentResponseMsg 是 AI 响应的消息类型
type agentResponseMsg struct {
	content   string
	userInput string // 用于更新历史
	err       error
}

// focus 表示当前焦点位置
type focus int

const (
	focusTextarea focus = iota
	focusViewport
)

// model 是 TUI 的状态
type model struct {
	textarea   textarea.Model
	viewport   viewport.Model
	messages   []chatMessage
	waiting    bool
	agent      fantasy.Agent
	agentMsgs  []fantasy.Message
	width      int
	height     int
	focus      focus // 当前焦点
	mdRenderer *glamour.TermRenderer
}

// 样式定义 (深色主题)
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	userStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")).
			Bold(true)

	assistantStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#874BFD"))

	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4"))

	// 焦点状态的边框样式（更亮）
	focusedBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#04B575"))

	// 非焦点状态的边框样式（暗淡）
	blurredBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#555555"))
)

// NewTUI 创建新的 TUI model
func NewTUI(agent fantasy.Agent) model {
	ta := textarea.New()
	ta.Placeholder = "输入消息... (Enter 发送, Shift+Enter 换行)"
	ta.Focus()
	ta.CharLimit = 4000
	ta.SetHeight(3)
	ta.ShowLineNumbers = false
	ta.SetStyles(textarea.DefaultDarkStyles())

	vp := viewport.New(viewport.WithWidth(80), viewport.WithHeight(20))
	vp.SoftWrap = true          // 启用自动换行
	vp.MouseWheelEnabled = true // 启用鼠标滚轮
	vp.MouseWheelDelta = 3      // 每次滚动 3 行

	// 创建 Markdown 渲染器（深色主题）
	renderer, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(76), // 默认宽度，后续会根据窗口调整
	)

	return model{
		textarea:   ta,
		viewport:   vp,
		messages:   []chatMessage{},
		agent:      agent,
		agentMsgs:  []fantasy.Message{},
		focus:      focusTextarea, // 默认焦点在输入框
		mdRenderer: renderer,
	}
}

// Init 实现 tea.Model 接口
func (m model) Init() tea.Cmd {
	return textarea.Blink
}

// Update 处理所有消息
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "tab":
			// Tab 切换焦点
			if m.focus == focusTextarea {
				m.focus = focusViewport
				m.textarea.Blur()
			} else {
				m.focus = focusTextarea
				m.textarea.Focus()
			}
			return m, nil
		case "enter":
			// Shift+Enter 换行，普通 Enter 发送
			if msg.Mod&tea.ModShift != 0 {
				// 让 textarea 处理换行
				break
			}

			userInput := strings.TrimSpace(m.textarea.Value())
			if userInput == "" || m.waiting {
				return m, nil
			}

			m.messages = append(m.messages, chatMessage{
				role:    "user",
				content: userInput,
			})
			m.textarea.Reset()
			m.waiting = true
			m.viewport.SetContent(m.renderMessages())
			m.viewport.GotoBottom()

			return m, m.sendToAgent(userInput)
		}

	case agentResponseMsg:
		m.waiting = false
		if msg.err != nil {
			m.messages = append(m.messages, chatMessage{
				role:    "assistant",
				content: fmt.Sprintf("Error: %v", msg.err),
			})
		} else {
			m.messages = append(m.messages, chatMessage{
				role:    "assistant",
				content: msg.content,
			})
			// 更新历史消息（在主线程中更新，避免并发问题）
			m.agentMsgs = append(m.agentMsgs, fantasy.NewUserMessage(msg.userInput))
			m.agentMsgs = append(m.agentMsgs, fantasy.Message{
				Role:    fantasy.MessageRoleAssistant,
				Content: []fantasy.MessagePart{fantasy.TextPart{Text: msg.content}},
			})
		}
		m.viewport.SetContent(m.renderMessages())
		m.viewport.GotoBottom()
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		headerHeight := 3
		inputHeight := 5
		contentHeight := m.height - headerHeight - inputHeight

		if contentHeight < 5 {
			contentHeight = 5
		}

		m.viewport.SetWidth(m.width - 4)
		m.viewport.SetHeight(contentHeight)
		m.textarea.SetWidth(m.width - 4)

		// 更新 Markdown 渲染器宽度
		contentWidth := m.width - 8 // 减去边框和内边距
		if contentWidth < 40 {
			contentWidth = 40
		}
		m.mdRenderer, _ = glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(contentWidth),
		)

		m.viewport.SetContent(m.renderMessages())
		return m, nil
	}

	// 鼠标滚轮事件始终传递给 viewport
	switch msg.(type) {
	case tea.MouseWheelMsg:
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	default:
		// 其他事件根据焦点更新对应组件
		if m.focus == focusTextarea {
			m.textarea, cmd = m.textarea.Update(msg)
			cmds = append(cmds, cmd)
		} else {
			m.viewport, cmd = m.viewport.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

// sendToAgent 发送消息给 Agent
func (m *model) sendToAgent(input string) tea.Cmd {
	// 复制当前历史消息，避免并发问题
	msgsCopy := make([]fantasy.Message, len(m.agentMsgs))
	copy(msgsCopy, m.agentMsgs)

	return func() tea.Msg {
		agentCall := fantasy.AgentCall{
			Messages: msgsCopy,
			Prompt:   input,
		}

		result, err := m.agent.Generate(context.Background(), agentCall)
		if err != nil {
			return agentResponseMsg{err: err, userInput: input}
		}

		responseText := result.Response.Content.Text()
		return agentResponseMsg{content: responseText, userInput: input}
	}
}

// View 渲染界面
func (m model) View() tea.View {
	if m.width == 0 {
		return tea.NewView("Loading...")
	}

	title := titleStyle.Render("MiniCode")
	header := lipgloss.NewStyle().
		Width(m.width).
		Align(lipgloss.Center).
		Render(title)

	// 根据焦点选择边框样式
	var msgBorder, inputBorder lipgloss.Style
	if m.focus == focusViewport {
		msgBorder = focusedBorderStyle
		inputBorder = blurredBorderStyle
	} else {
		msgBorder = blurredBorderStyle
		inputBorder = focusedBorderStyle
	}

	messagesBox := msgBorder.
		Width(m.width - 2).
		Height(m.viewport.Height() + 2).
		Render(m.viewport.View())

	inputBox := inputBorder.
		Width(m.width - 2).
		Render(m.textarea.View())

	var status string
	if m.waiting {
		status = "  Thinking..."
	} else if m.focus == focusViewport {
		status = "  Tab 切换到输入框, ↑↓ 滚动, Ctrl+C 退出"
	} else {
		status = "  Enter 发送, Shift+Enter 换行, Tab 切换, Ctrl+C 退出"
	}

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		messagesBox,
		inputBox,
		status,
	)

	v := tea.NewView(content)
	v.MouseMode = tea.MouseModeCellMotion // 启用鼠标点击和滚轮
	return v
}

// renderMessages 渲染消息历史
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
			sb.WriteString(assistantStyle.Render("Assistant:"))
			sb.WriteString("\n")
			// 使用 glamour 渲染 Markdown
			if m.mdRenderer != nil {
				rendered, err := m.mdRenderer.Render(msg.content)
				if err == nil {
					sb.WriteString(strings.TrimSpace(rendered))
				} else {
					sb.WriteString(msg.content)
				}
			} else {
				sb.WriteString(msg.content)
			}
			sb.WriteString("\n\n")
		}
	}
	return sb.String()
}
