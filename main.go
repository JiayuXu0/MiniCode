package main

import (
	"context"
	"fmt"
	"log"
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
	if err := godotenv.Load(); err != nil {
		log.Printf("警告: 无法加载 .env 文件: %v", err)
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "错误: 请设置 OPENAI_API_KEY 环境变量")
		os.Exit(1)
	}

	ctx := context.Background()

	provider, err := openaicompat.New(
		openaicompat.WithBaseURL(baseURL),
		openaicompat.WithAPIKey(apiKey),
		openaicompat.WithName("zai"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "创建 provider 失败: %v\n", err)
		os.Exit(1)
	}

	model, err := provider.LanguageModel(ctx, modelID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "获取模型失败: %v\n", err)
		os.Exit(1)
	}

	agent := fantasy.NewAgent(model,
		fantasy.WithSystemPrompt(systemPrompt),
		fantasy.WithTools(
			tools.NewGlobTool(),
			tools.NewViewTool(),
			tools.NewGrepTool(),
		),
	)

	m := tui.NewTUI(agent)
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Printf("运行错误: %v", err)
		os.Exit(1)
	}
}
