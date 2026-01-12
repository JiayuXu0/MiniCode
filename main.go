package main

import (
	"context"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"charm.land/fantasy"
	"charm.land/fantasy/providers/openaicompat"
	"github.com/joho/godotenv"

	"github.com/JiayuXu0/MiniCode/internal/config"
	"github.com/JiayuXu0/MiniCode/internal/permission"
	"github.com/JiayuXu0/MiniCode/tools"
	"github.com/JiayuXu0/MiniCode/tui"
)

const systemPrompt = `你是工控智科的大模型，是一个有帮助的编码与问题解决助手，具备文件系统工具能力。

Available tools:
- glob: Find files by pattern
- view: Read file contents
- grep: Search file contents
- bash: Execute shell commands
- write: Write content to files
- edit: Edit files by replacing text

When asked about code or files, use tools to gather information.
You may need multiple tool calls. Respond in the user's language.
Do not make up facts. If you are unsure or lack information, say so clearly and ask for the needed details.`

func main() {
	_ = godotenv.Load() // 忽略 .env 不存在的情况

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()

	// 从配置创建 model
	model, err := createModel(ctx, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating model: %v\n", err)
		os.Exit(1)
	}

	// 创建权限服务
	permService := permission.NewService()

	agent := fantasy.NewAgent(model,
		fantasy.WithSystemPrompt(systemPrompt),
		fantasy.WithTools(
			tools.NewGlobTool(),
			tools.NewViewTool(),
			tools.NewGrepTool(),
			tools.NewBashTool(permService),
			tools.NewWriteTool(permService),
			tools.NewEditTool(permService),
		),
	)

	m := tui.New(agent, permService, cfg.GetDefaultModel())
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())

	// 设置 Program 引用
	m.SetProgram(p)
	permService.SetProgram(p)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Runtime error: %v\n", err)
		os.Exit(1)
	}
}

// createModel 根据配置创建语言模型
func createModel(ctx context.Context, cfg *config.Config) (fantasy.LanguageModel, error) {
	modelName := cfg.GetDefaultModel()

	// 按优先级尝试不同的 provider
	providers := []string{"zhipu", "openrouter", "bailian", "openai"}

	for _, providerName := range providers {
		if providerCfg, ok := cfg.GetProvider(providerName); ok {
			switch providerName {
			case "zhipu":
				return createOpenAICompatModel(ctx, providerCfg, modelName, "https://open.bigmodel.cn/api/coding/paas/v4", "zai")
			case "openrouter":
				return createOpenAICompatModel(ctx, providerCfg, modelName, "https://openrouter.ai/api/v1", "openrouter")
			case "bailian":
				return createOpenAICompatModel(ctx, providerCfg, modelName, "https://dashscope.aliyuncs.com/compatible-mode/v1", "bailian")
			case "openai":
				return createOpenAICompatModel(ctx, providerCfg, modelName, "https://api.openai.com/v1", "openai")
			}
		}
	}

	// 回退到环境变量
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("no API key found. Add a provider in minicode.json or set OPENAI_API_KEY env var")
	}

	// 使用默认配置创建智谱模型
	providerCfg := config.ProviderConfig{
		APIKey:  apiKey,
		BaseURL: "https://open.bigmodel.cn/api/coding/paas/v4",
		Name:    "zai",
	}
	return createOpenAICompatModel(ctx, providerCfg, modelName, "https://open.bigmodel.cn/api/coding/paas/v4", "zai")
}

// createOpenAICompatModel 创建 OpenAI 兼容的模型
// 支持：智谱 GLM、OpenRouter、百炼、OpenAI 等
func createOpenAICompatModel(ctx context.Context, providerCfg config.ProviderConfig, modelName, defaultBaseURL, defaultName string) (fantasy.LanguageModel, error) {
	if providerCfg.APIKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	// 使用配置的 base_url，如果没有则使用默认值
	baseURL := providerCfg.BaseURL
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	// 使用配置的 name，如果没有则使用默认值
	name := providerCfg.Name
	if name == "" {
		name = defaultName
	}

	// 创建 openaicompat provider
	provider, err := openaicompat.New(
		openaicompat.WithBaseURL(baseURL),
		openaicompat.WithAPIKey(providerCfg.APIKey),
		openaicompat.WithName(name),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider: %w", err)
	}

	model, err := provider.LanguageModel(ctx, modelName)
	if err != nil {
		return nil, fmt.Errorf("failed to get model %s: %w", modelName, err)
	}

	return model, nil
}
