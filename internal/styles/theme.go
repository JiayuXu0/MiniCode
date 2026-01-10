package styles

import "github.com/charmbracelet/lipgloss"

// Theme 定义应用主题
type Theme struct {
	// 基础颜色
	Primary    lipgloss.Color
	Secondary  lipgloss.Color
	Accent     lipgloss.Color
	Background lipgloss.Color
	Foreground lipgloss.Color

	// 语义颜色
	Success lipgloss.Color
	Warning lipgloss.Color
	Error   lipgloss.Color
	Muted   lipgloss.Color
	Dim     lipgloss.Color
	Content lipgloss.Color

	// 角色颜色
	UserColor      lipgloss.Color
	AssistantColor lipgloss.Color
	ToolColor      lipgloss.Color
}

// DefaultDarkTheme 默认暗色主题（紫橙色系）
var DefaultDarkTheme = Theme{
	Primary:    lipgloss.Color("#BD93F9"), // 亮紫色
	Secondary:  lipgloss.Color("#FF79C6"), // 粉红色
	Accent:     lipgloss.Color("#FFB86C"), // 橙色
	Background: lipgloss.Color("#282A36"),
	Foreground: lipgloss.Color("#F8F8F2"),

	Success: lipgloss.Color("#50FA7B"), // 亮绿
	Warning: lipgloss.Color("#FFB86C"), // 橙色
	Error:   lipgloss.Color("#FF5555"), // 红色
	Muted:   lipgloss.Color("#6272A4"), // 灰蓝
	Dim:     lipgloss.Color("#44475A"), // 深灰
	Content: lipgloss.Color("#8BE9FD"), // 青色

	UserColor:      lipgloss.Color("#50FA7B"), // 亮绿
	AssistantColor: lipgloss.Color("#BD93F9"), // 亮紫
	ToolColor:      lipgloss.Color("#FFB86C"), // 橙色
}

// DefaultLightTheme 默认亮色主题（蓝绿色系，与暗色主题形成对比）
var DefaultLightTheme = Theme{
	Primary:    lipgloss.Color("#0969DA"), // 蓝色
	Secondary:  lipgloss.Color("#1F6FEB"), // 亮蓝色
	Accent:     lipgloss.Color("#CF222E"), // 红色
	Background: lipgloss.Color("#FFFFFF"),
	Foreground: lipgloss.Color("#1F2328"),

	Success: lipgloss.Color("#1A7F37"), // 深绿
	Warning: lipgloss.Color("#9A6700"), // 深黄
	Error:   lipgloss.Color("#CF222E"), // 红色
	Muted:   lipgloss.Color("#656D76"), // 灰色
	Dim:     lipgloss.Color("#D0D7DE"), // 浅灰
	Content: lipgloss.Color("#424A53"), // 深灰

	UserColor:      lipgloss.Color("#1A7F37"), // 绿色
	AssistantColor: lipgloss.Color("#0969DA"), // 蓝色
	ToolColor:      lipgloss.Color("#9A6700"), // 橙黄色
}
