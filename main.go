package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"charm.land/fantasy"
	"charm.land/fantasy/providers/openaicompat"
	"github.com/joho/godotenv"

	"github.com/JiayuXu0/MiniCode/tools"
)

const (
	baseURL      = "https://open.bigmodel.cn/api/coding/paas/v4"
	modelID      = "glm-4.7"
	systemPrompt = "你是一个有帮助的 AI 助手，请用中文回答用户的问题。"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("警告: 无法加载 .env 文件: %v", err)
	}

	if len(os.Args) < 2 {
		fmt.Println("用法: go run . <你的问题>")
		os.Exit(1)
	}

	prompt := os.Args[1]
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
		fantasy.WithTools(tools.NewGlobTool()),
	)
	result, err := agent.Generate(ctx, fantasy.AgentCall{Prompt: prompt})
	if err != nil {
		fmt.Fprintf(os.Stderr, "生成响应失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(result.Response.Content.Text())
}
