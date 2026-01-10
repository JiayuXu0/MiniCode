package styles

import "github.com/charmbracelet/lipgloss"

// Styles 包含所有预定义样式
type Styles struct {
	theme Theme

	// 文本样式
	Title          lipgloss.Style
	Subtitle       lipgloss.Style
	UserLabel      lipgloss.Style
	AssistantLabel lipgloss.Style
	MutedText      lipgloss.Style

	// 容器样式
	Box         lipgloss.Style
	ToolBox     lipgloss.Style
	ThinkingBox lipgloss.Style
	InputBox    lipgloss.Style

	// 工具相关样式
	ToolHeader  lipgloss.Style
	ToolContent lipgloss.Style

	// 思考样式
	Reasoning lipgloss.Style

	// 状态样式
	Success lipgloss.Style
	Warning lipgloss.Style
	Error   lipgloss.Style
	Spinner lipgloss.Style

	// 占位符样式
	Placeholder lipgloss.Style
}

// NewStyles 创建样式实例
func NewStyles(theme Theme) *Styles {
	s := &Styles{theme: theme}
	s.init()
	return s
}

func (s *Styles) init() {
	// 文本样式
	s.Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(s.theme.Primary)

	s.Subtitle = lipgloss.NewStyle().
		Foreground(s.theme.Muted).
		Italic(true)

	s.UserLabel = lipgloss.NewStyle().
		Bold(true).
		Foreground(s.theme.UserColor)

	s.AssistantLabel = lipgloss.NewStyle().
		Foreground(s.theme.AssistantColor)

	s.MutedText = lipgloss.NewStyle().
		Foreground(s.theme.Muted)

	// 容器样式
	s.Box = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(s.theme.Primary)

	s.ToolBox = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(s.theme.ToolColor).
		Padding(0, 1)

	s.ThinkingBox = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(s.theme.Muted).
		Padding(0, 1)

	s.InputBox = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(s.theme.Primary)

	// 工具相关样式
	s.ToolHeader = lipgloss.NewStyle().
		Foreground(s.theme.ToolColor).
		Bold(true)

	s.ToolContent = lipgloss.NewStyle().
		Foreground(s.theme.Content)

	// 思考样式
	s.Reasoning = lipgloss.NewStyle().
		Foreground(s.theme.Muted).
		Italic(true)

	// 状态样式
	s.Success = lipgloss.NewStyle().
		Foreground(s.theme.Success)

	s.Warning = lipgloss.NewStyle().
		Foreground(s.theme.Warning)

	s.Error = lipgloss.NewStyle().
		Foreground(s.theme.Error)

	s.Spinner = lipgloss.NewStyle().
		Foreground(s.theme.Accent)

	// 占位符样式
	s.Placeholder = lipgloss.NewStyle().
		Foreground(s.theme.Muted).
		Italic(true)
}

// Theme 返回当前主题
func (s *Styles) Theme() Theme {
	return s.theme
}

// BoxWithFocus 返回带焦点状态的边框样式
func (s *Styles) BoxWithFocus(focused bool) lipgloss.Style {
	if focused {
		return s.Box.BorderForeground(s.theme.Warning)
	}
	return s.Box.BorderForeground(s.theme.Dim)
}

// InputBoxWithFocus 返回带焦点状态的输入框样式
func (s *Styles) InputBoxWithFocus(focused bool) lipgloss.Style {
	if focused {
		return s.InputBox.BorderForeground(s.theme.Primary)
	}
	return s.InputBox.BorderForeground(s.theme.Dim)
}
