package tui

import "github.com/charmbracelet/lipgloss"

// === 显示常量 ===

const (
	toolBoxMaxLines  = 5 // 工具调用框最大显示行数
	thinkingMaxLines = 6 // 思考区域最大显示行数
)

// === 动画帧 ===

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// === 颜色定义 ===

var (
	colorPrimary   = lipgloss.Color("#7D56F4") // 主色调（紫色）
	colorSecondary = lipgloss.Color("#874BFD") // 次色调
	colorSuccess   = lipgloss.Color("#04B575") // 成功色（绿色）
	colorWarning   = lipgloss.Color("#FFA500") // 警告色（橙色）
	colorMuted     = lipgloss.Color("#666666") // 暗淡色
	colorDim       = lipgloss.Color("#444444") // 更暗色
	colorContent   = lipgloss.Color("#888888") // 内容色
)

// === 样式定义 ===

var (
	// 标题样式
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary)

	// 用户消息样式
	userStyle = lipgloss.NewStyle().
			Foreground(colorSuccess).
			Bold(true)

	// 助手消息样式
	assistantStyle = lipgloss.NewStyle().
			Foreground(colorSecondary)

	// 思考过程样式
	reasoningStyle = lipgloss.NewStyle().
			Foreground(colorMuted).
			Italic(true)

	// 工具框样式
	toolBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorWarning).
			Padding(0, 1)

	// 工具标题样式
	toolHeaderStyle = lipgloss.NewStyle().
			Foreground(colorWarning).
			Bold(true)

	// 工具内容样式
	toolContentStyle = lipgloss.NewStyle().
				Foreground(colorContent)

	// 思考框样式
	thinkingBoxStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorMuted).
				Padding(0, 1)

	// 边框样式
	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorPrimary)

	// 占位符样式
	placeholderStyle = lipgloss.NewStyle().
				Foreground(colorMuted).
				Italic(true)
)
