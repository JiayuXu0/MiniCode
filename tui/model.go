package tui

import (
	"context"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"charm.land/fantasy"
)

// Model 是 TUI 的状态模型
type Model struct {
	// UI 组件
	textarea textarea.Model
	viewport viewport.Model

	// 消息历史（用于 API 调用和 UI 渲染）
	history []fantasy.Message

	// Agent 相关
	agent      fantasy.Agent      // Agent 实例
	program    *tea.Program       // Program 引用，用于流式回调发送消息
	cancelFunc context.CancelFunc // 取消函数

	// 状态标志
	waiting      bool // 是否等待响应
	streaming    bool // 是否正在流式输出
	needsRefresh bool // 是否需要刷新界面

	// 流式内容（临时存储，流式完成后清空）
	streamParts []streamPart

	// 界面尺寸
	width  int
	height int

	// 焦点状态
	focus focus

	// 动画和统计
	spinnerFrame int // 动画帧索引
	totalTokens  int // Token 统计
}

// New 创建新的 TUI 实例
func New(agent fantasy.Agent) *Model {
	ta := textarea.New()
	ta.Placeholder = ""
	ta.Prompt = ""
	ta.ShowLineNumbers = false
	ta.SetHeight(1)
	ta.SetWidth(80)
	ta.Focus()

	vp := viewport.New(80, 20)
	vp.MouseWheelEnabled = true

	return &Model{
		textarea: ta,
		viewport: vp,
		agent:    agent,
		focus:    focusTextarea,
	}
}

// SetProgram 设置 Program 引用，用于流式回调
func (m *Model) SetProgram(p *tea.Program) {
	m.program = p
}

// Init 实现 tea.Model 接口
func (m *Model) Init() tea.Cmd {
	return textarea.Blink
}
