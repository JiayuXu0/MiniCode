package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"charm.land/fantasy"
	"charm.land/fantasy/providers/openaicompat"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/joho/godotenv"
)

const (
	baseURL      = "https://open.bigmodel.cn/api/coding/paas/v4"
	modelID      = "glm-4.7"
	systemPrompt = "你是一个有帮助的 AI 助手，请用中文回答用户的问题。"
)

type GlobInput struct {
	Pattern string `json:"pattern" description:"The glob pattern to match files (e.g., *.go, **/*.json)"`
	Path    string `json:"path,omitempty" description:"The directory to search in (defaults to current directory)"`
}

func glob(ctx context.Context, input GlobInput, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
	// 默认搜索当前目录
	searchPath := input.Path
	if searchPath == "" {
		searchPath = "."
	}

	// 转换为绝对路径
	absPath, err := filepath.Abs(searchPath)
	if err != nil {
		return fantasy.NewTextErrorResponse(fmt.Sprintf("无法解析路径: %v", err)), nil
	}

	// 检查目录是否存在
	info, err := os.Stat(absPath)
	if err != nil {
		return fantasy.NewTextErrorResponse(fmt.Sprintf("目录不存在: %v", err)), nil
	}
	if !info.IsDir() {
		return fantasy.NewTextErrorResponse(fmt.Sprintf("路径不是目录: %s", absPath)), nil
	}

	// 构建完整的 glob pattern
	pattern := filepath.Join(absPath, input.Pattern)

	// 使用 doublestar 进行匹配（支持 ** 递归）
	matches, err := doublestar.FilepathGlob(pattern)
	if err != nil {
		return fantasy.NewTextErrorResponse(fmt.Sprintf("glob 匹配失败: %v", err)), nil
	}

	// 如果没有匹配结果
	if len(matches) == 0 {
		return fantasy.NewTextResponse(fmt.Sprintf("没有找到匹配 '%s' 的文件", input.Pattern)), nil
	}

	// 将绝对路径转换为相对路径（相对于搜索目录）
	var relPaths []string
	for _, match := range matches {
		relPath, err := filepath.Rel(absPath, match)
		if err != nil {
			relPath = match // 如果转换失败，使用绝对路径
		}
		relPaths = append(relPaths, relPath)
	}

	// 构建返回结果
	result := fmt.Sprintf("找到 %d 个匹配的文件:\n%s", len(relPaths), strings.Join(relPaths, "\n"))
	return fantasy.NewTextResponse(result), nil
}

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

	GlobTool := fantasy.NewAgentTool(
		"glob",
		"Match and list filesystem paths using wildcard patterns"+
			"(e.g., *.py, data/**/*.json) within a specified directory, "+
			"returning a list of matched file/directory paths"+
			" (optionally recursive).",
		glob,
	)
	agent := fantasy.NewAgent(model, fantasy.WithSystemPrompt(systemPrompt),
		fantasy.WithTools(GlobTool))
	result, err := agent.Generate(ctx, fantasy.AgentCall{Prompt: prompt})
	if err != nil {
		fmt.Fprintf(os.Stderr, "生成响应失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(result.Response.Content.Text())
}
