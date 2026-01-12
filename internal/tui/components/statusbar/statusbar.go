package statusbar

import (
	"fmt"
	"strings"
	"time"

	"github.com/JiayuXu0/MiniCode/internal/styles"

	"github.com/charmbracelet/lipgloss"
)

// Status 表示状态
type Status int

const (
	StatusReady Status = iota
	StatusStreaming
	StatusThinking
	StatusError
	StatusDisconnected
)

// String 返回状态的字符串表示
func (s Status) String() string {
	switch s {
	case StatusReady:
		return "Ready"
	case StatusStreaming:
		return "Streaming..."
	case StatusThinking:
		return "Thinking..."
	case StatusError:
		return "Error"
	case StatusDisconnected:
		return "Disconnected"
	default:
		return "Unknown"
	}
}

// spinnerFrames 动画帧
var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// Model 是状态栏的数据模型
type Model struct {
	styles *styles.Styles
	width  int

	// 状态数据
	ModelName    string
	TotalTokens  int
	MessageCount int
	Status       Status
	ErrorMsg     string

	// 会话信息
	SessionID  string
	StartTime  time.Time
	LastActive time.Time

	// 动画
	spinnerFrame int

	// 配置
	ShowSessionID bool
	ShowDuration  bool
}

// New 创建状态栏
func New(s *styles.Styles) Model {
	return Model{
		styles:       s,
		ModelName:    "claude-sonnet",
		Status:       StatusReady,
		StartTime:    time.Now(),
		LastActive:   time.Now(),
		ShowDuration: true,
	}
}

// SetStyles 更新样式引用
func (m *Model) SetStyles(s *styles.Styles) {
	if s == nil {
		return
	}
	m.styles = s
}

// SetWidth 设置宽度
func (m *Model) SetWidth(w int) {
	m.width = w
}

// SetModel 设置模型名称
func (m *Model) SetModel(name string) {
	m.ModelName = name
}

// AddTokens 添加 Token 计数（增量）
func (m *Model) AddTokens(n int) {
	if n <= 0 {
		return
	}
	m.TotalTokens += n
	m.LastActive = time.Now()
}

// SetTokens 设置 Token 计数（绝对值）
func (m *Model) SetTokens(n int) {
	if n < 0 {
		n = 0
	}
	m.TotalTokens = n
	m.LastActive = time.Now()
}

// ResetTokens 重置 Token 计数
func (m *Model) ResetTokens() {
	m.TotalTokens = 0
}

// IncrementMessages 增加消息计数
func (m *Model) IncrementMessages() {
	m.MessageCount++
}

// ResetMessages 重置消息计数
func (m *Model) ResetMessages() {
	m.MessageCount = 0
}

// SetStatus 设置状态
func (m *Model) SetStatus(s Status) {
	if m.Status == s {
		return // 状态未变，不更新
	}
	m.Status = s
	if s != StatusError {
		m.ErrorMsg = ""
	}
}

// SetError 设置错误状态
func (m *Model) SetError(msg string) {
	m.Status = StatusError
	m.ErrorMsg = msg
}

// SetSessionID 设置会话 ID
func (m *Model) SetSessionID(id string) {
	m.SessionID = id
	m.ShowSessionID = true
}

// Tick 更新动画帧
func (m *Model) Tick() {
	if m.Status == StatusStreaming || m.Status == StatusThinking {
		m.spinnerFrame = (m.spinnerFrame + 1) % len(spinnerFrames)
	}
}

// View 渲染状态栏
func (m Model) View() string {
	if m.width == 0 {
		return ""
	}

	// 紧凑模式（宽度 < 60）
	if m.width < 60 {
		return m.viewCompact()
	}

	// 标准模式（宽度 60-100）
	if m.width < 100 {
		return m.viewStandard()
	}

	// 完整模式（宽度 >= 100）
	return m.viewFull()
}

// viewCompact 紧凑视图：模型 | 状态
func (m Model) viewCompact() string {
	theme := m.styles.Theme()

	modelStyle := lipgloss.NewStyle().Foreground(theme.Primary).Bold(true)
	left := modelStyle.Render(truncateString(m.ModelName, 10))

	status := m.renderStatus()

	content := left + " │ " + status
	return m.wrapBar(content)
}

// viewStandard 标准视图：模型 | Tokens | 状态
func (m Model) viewStandard() string {
	theme := m.styles.Theme()
	statsStyle := lipgloss.NewStyle().Foreground(theme.Muted)

	parts := []string{
		m.renderModel(),
		statsStyle.Render(fmt.Sprintf("Tokens: %s", formatNumber(m.TotalTokens))),
		m.renderStatus(),
	}

	return m.wrapBar(strings.Join(parts, " │ "))
}

// viewFull 完整视图：模型 | Tokens | Messages | 时长 | 状态
func (m Model) viewFull() string {
	theme := m.styles.Theme()
	parts := make([]string, 0, 5)

	// 1. 模型名称
	parts = append(parts, m.renderModel())

	// 2. Token 统计
	statsStyle := lipgloss.NewStyle().Foreground(theme.Muted)
	tokenStr := fmt.Sprintf("Tokens: %s", formatNumber(m.TotalTokens))
	parts = append(parts, statsStyle.Render(tokenStr))

	// 3. 消息数量
	msgStr := fmt.Sprintf("Messages: %d", m.MessageCount)
	parts = append(parts, statsStyle.Render(msgStr))

	// 4. 会话 ID（可选）
	if m.ShowSessionID && m.SessionID != "" {
		sessionStr := fmt.Sprintf("Session: %s", truncateID(m.SessionID, 8))
		parts = append(parts, statsStyle.Render(sessionStr))
	}

	// 5. 会话时长（可选）
	if m.ShowDuration {
		duration := time.Since(m.StartTime)
		durationStr := formatDuration(duration)
		parts = append(parts, statsStyle.Render(durationStr))
	}

	// 6. 状态指示
	parts = append(parts, m.renderStatus())

	// 组合
	separator := statsStyle.Render(" │ ")
	content := strings.Join(parts, separator)

	return m.wrapBar(content)
}

// renderModel 渲染模型名称
func (m Model) renderModel() string {
	theme := m.styles.Theme()
	modelStyle := lipgloss.NewStyle().Foreground(theme.Primary).Bold(true)
	return modelStyle.Render(m.ModelName)
}

// renderStatus 渲染状态部分
func (m Model) renderStatus() string {
	var icon, text string
	var color lipgloss.Color

	theme := m.styles.Theme()

	switch m.Status {
	case StatusReady:
		icon = "⚡"
		text = "Ready"
		color = theme.Success
	case StatusStreaming:
		icon = spinnerFrames[m.spinnerFrame]
		text = "Streaming..."
		color = theme.Warning
	case StatusThinking:
		icon = spinnerFrames[m.spinnerFrame]
		text = "Thinking..."
		color = theme.Accent
	case StatusError:
		icon = "✗"
		if m.ErrorMsg != "" {
			text = truncateString(m.ErrorMsg, 20)
		} else {
			text = "Error"
		}
		color = theme.Error
	case StatusDisconnected:
		icon = "○"
		text = "Disconnected"
		color = theme.Muted
	}

	style := lipgloss.NewStyle().Foreground(color)
	return style.Render(icon + " " + text)
}

// wrapBar 包装状态栏
func (m Model) wrapBar(content string) string {
	barStyle := lipgloss.NewStyle().
		Width(m.width).
		Padding(0, 1).
		Background(lipgloss.Color("#1a1a2e"))

	return barStyle.Render(content)
}

// formatNumber 格式化数字（添加千位分隔符）
func formatNumber(n int) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}
	if n < 1000000 {
		return fmt.Sprintf("%d,%03d", n/1000, n%1000)
	}
	return fmt.Sprintf("%d,%03d,%03d", n/1000000, (n/1000)%1000, n%1000)
}

// formatDuration 格式化时长
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm%ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	hours := int(d.Hours())
	mins := int(d.Minutes()) % 60
	return fmt.Sprintf("%dh%dm", hours, mins)
}

// truncateID 截断 ID
func truncateID(id string, maxLen int) string {
	if len(id) <= maxLen {
		return id
	}
	return id[:maxLen]
}

// truncateString 截断字符串
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
