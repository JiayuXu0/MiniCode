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

const systemPrompt = `You are a helpful coding assistant with file system tools.

Available tools:
- glob: Find files by pattern
- view: Read file contents
- grep: Search file contents
- bash: Execute shell commands
- write: Write content to files
- edit: Edit files by replacing text

When asked about code or files, use tools to gather information.
You may need multiple tool calls. Respond in the user's language.`

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

	m := tui.New(agent, permService)
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

	// 优先使用配置中的 zhipu provider（智谱 GLM）
	if providerCfg, ok := cfg.GetProvider("zhipu"); ok {
		return createZhipuModel(ctx, providerCfg, modelName)
	}

	// 回退到环境变量
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("no API key found. Set it in config or OPENAI_API_KEY env var")
	}

	// 使用默认配置创建智谱模型
	providerCfg := config.ProviderConfig{
		APIKey:  apiKey,
		BaseURL: "https://open.bigmodel.cn/api/coding/paas/v4",
		Name:    "zai",
	}
	return createZhipuModel(ctx, providerCfg, modelName)
}

// createZhipuModel 创建智谱 GLM 模型（使用 openaicompat）
func createZhipuModel(ctx context.Context, providerCfg config.ProviderConfig, modelName string) (fantasy.LanguageModel, error) {
	if providerCfg.APIKey == "" {
		return nil, fmt.Errorf("zhipu API key is required")
	}

	// 默认值
	baseURL := providerCfg.BaseURL
	if baseURL == "" {
		baseURL = "https://open.bigmodel.cn/api/coding/paas/v4"
	}

	name := providerCfg.Name
	if name == "" {
		name = "zai"
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
		return nil, fmt.Errorf("failed to get model: %w", err)
	}

	return model, nil
}
