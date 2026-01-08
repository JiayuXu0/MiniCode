package tui

import (
	"context"
	"fmt"
	"time"

	"charm.land/fantasy"
	tea "github.com/charmbracelet/bubbletea"
)

// startStream 启动流式请求
func (m *Model) startStream(input string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithCancel(context.Background())
		m.cancelFunc = cancel

		result, err := m.agent.Stream(ctx, fantasy.AgentStreamCall{
			Messages: m.history,
			Prompt:   input,

			// 文本片段回调
			OnTextDelta: func(id, text string) error {
				m.program.Send(streamTextMsg{delta: text})
				return nil
			},

			// 推理过程回调
			OnReasoningDelta: func(id, text string) error {
				m.program.Send(streamReasoningMsg{delta: text})
				return nil
			},

			// 工具调用回调
			OnToolCall: func(tc fantasy.ToolCallContent) error {
				m.program.Send(streamToolCallMsg{
					name:  tc.ToolName,
					input: tc.Input,
				})
				return nil
			},

			// 工具结果回调
			OnToolResult: func(tr fantasy.ToolResultContent) error {
				isError := tr.Result != nil && tr.Result.GetType() == fantasy.ToolResultContentTypeError
				m.program.Send(streamToolResultMsg{
					name:    tr.ToolName,
					content: fmt.Sprintf("%v", tr.Result),
					isError: isError,
				})
				return nil
			},

			// 流结束回调，实时更新 token
			OnStreamFinish: func(usage fantasy.Usage, _ fantasy.FinishReason, _ fantasy.ProviderMetadata) error {
				m.program.Send(streamTokenUpdateMsg{tokens: int(usage.TotalTokens)})
				return nil
			},
		})

		// 从最终结果获取 token 统计
		var totalTokens int
		if result != nil {
			totalTokens = int(result.TotalUsage.TotalTokens)
		}

		return streamDoneMsg{
			err:         err,
			result:      result,
			totalTokens: totalTokens,
		}
	}
}

// tick 动画定时器（50ms 刷新一次，20fps）
func (m *Model) tick() tea.Cmd {
	return tea.Tick(25*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg{}
	})
}
