package main

import (
	"context"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"charm.land/fantasy"
	"charm.land/fantasy/providers/openaicompat"
	"github.com/joho/godotenv"

	"github.com/JiayuXu0/MiniCode/tools"
	"github.com/JiayuXu0/MiniCode/tui"
)

const (
	baseURL      = "https://open.bigmodel.cn/api/coding/paas/v4"
	modelID      = "glm-4.7"
	systemPrompt = `You are a helpful coding assistant with file system tools.

					Available tools:
					- glob: Find files by pattern
					- view: Read file contents
					- grep: Search file contents

					When asked about code or files, use tools to gather information.
					You may need multiple tool calls. Respond in the user's language.`
)

func main() {
	_ = godotenv.Load() // 忽略 .env 不存在的情况

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "Error: OPENAI_API_KEY environment variable is required")
		os.Exit(1)
	}

	ctx := context.Background()

	provider, err := openaicompat.New(
		openaicompat.WithBaseURL(baseURL),
		openaicompat.WithAPIKey(apiKey),
		openaicompat.WithName("zai"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create provider: %v\n", err)
		os.Exit(1)
	}

	model, err := provider.LanguageModel(ctx, modelID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get model: %v\n", err)
		os.Exit(1)
	}

	agent := fantasy.NewAgent(model,
		fantasy.WithSystemPrompt(systemPrompt),
		fantasy.WithTools(
			tools.NewGlobTool(),
			tools.NewViewTool(),
			tools.NewGrepTool(),
			tools.NewBashTool(),
			tools.NewWriteTool(),
			tools.NewEditTool(),
		),
	)

	m := tui.New(agent)
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	m.SetProgram(p) // 传递 Program 引用，用于流式回调

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Runtime error: %v\n", err)
		os.Exit(1)
	}
}
