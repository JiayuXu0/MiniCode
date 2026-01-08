package tui

import "charm.land/fantasy"

// === 流式消息类型 ===

// streamTextMsg 流式文本片段
type streamTextMsg struct {
	delta string
}

// streamReasoningMsg 思考过程片段
type streamReasoningMsg struct {
	delta string
}

// streamToolCallMsg 工具调用开始
type streamToolCallMsg struct {
	name  string
	input string
}

// streamToolResultMsg 工具调用结果
type streamToolResultMsg struct {
	name    string
	content string
	isError bool
}

// streamDoneMsg 流式完成
type streamDoneMsg struct {
	err         error
	result      *fantasy.AgentResult
	totalTokens int
}

// tickMsg 动画 tick
type tickMsg struct{}

// streamTokenUpdateMsg 实时 token 更新
type streamTokenUpdateMsg struct {
	tokens int
}

// === 辅助类型 ===

// streamPartType 流式内容类型
type streamPartType int

const (
	partTypeReasoning streamPartType = iota
	partTypeToolCall
	partTypeText
)

// streamPart 流式内容片段（按时间顺序存储）
type streamPart struct {
	partType streamPartType
	content  string        // reasoning/text 的内容
	tool     *toolCallInfo // tool call 的信息
}

// toolCallInfo 工具调用信息（用于流式显示）
type toolCallInfo struct {
	name   string
	input  string
	output string
	status string // "running", "done", "error"
}

// focus 焦点类型
type focus int

const (
	focusTextarea focus = iota
	focusViewport
)
