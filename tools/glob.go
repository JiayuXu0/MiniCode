package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"charm.land/fantasy"
	"github.com/bmatcuk/doublestar/v4"
)

type GlobInput struct {
	Pattern string `json:"pattern" description:"The glob pattern to match files (e.g., *.go, **/*.json)"`
	Path    string `json:"path,omitempty" description:"The directory to search in (defaults to current directory)"`
}

func Glob(ctx context.Context, input GlobInput, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
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

// NewGlobTool 创建并返回 glob 工具
func NewGlobTool() fantasy.AgentTool {
	return fantasy.NewAgentTool(
		"glob",
		"Match and list filesystem paths using wildcard patterns"+
			"(e.g., *.py, data/**/*.json) within a specified directory, "+
			"returning a list of matched file/directory paths"+
			" (optionally recursive).",
		Glob,
	)
}
