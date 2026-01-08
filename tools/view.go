package tools

import (
	"context"
	"fmt"
	"os"
	"strings"

	"charm.land/fantasy"
)

type ViewInput struct {
	Path      string `json:"path" description:"The path to the file to view"`
	StartLine int    `json:"start_line,omitempty" description:"The line number to start from (1-based, default 1)"`
	EndLine   int    `json:"end_line,omitempty" description:"The line number to end at (1-based, inclusive, default: start_line + 200)"`
}

func View(ctx context.Context, input ViewInput, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
	if input.Path == "" {
		return fantasy.ToolResponse{}, fmt.Errorf("path is required")
	}

	content, err := os.ReadFile(input.Path)
	if err != nil {
		return fantasy.NewTextErrorResponse(fmt.Sprintf("Failed to read file: %v", err)), nil
	}

	lines := strings.Split(string(content), "\n")
	totalLines := len(lines)

	// 处理默认值（1-based）
	startLine := input.StartLine
	if startLine <= 0 {
		startLine = 1
	}
	endLine := input.EndLine
	if endLine <= 0 {
		endLine = startLine + 200 - 1 // 默认显示 200 行
	}

	// 边界检查
	if startLine > totalLines {
		return fantasy.NewTextResponse(fmt.Sprintf("File %s has only %d lines, start_line %d is out of range", input.Path, totalLines, startLine)), nil
	}
	if endLine > totalLines {
		endLine = totalLines
	}
	if startLine > endLine {
		return fantasy.NewTextErrorResponse("start_line cannot be greater than end_line"), nil
	}

	// 转换为 0-based 索引
	start := startLine - 1
	end := endLine

	// 构建带行号的输出
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("File: %s (lines %d-%d of %d)\n\n", input.Path, startLine, endLine, totalLines))
	for i := start; i < end; i++ {
		sb.WriteString(fmt.Sprintf("%4d\t%s\n", i+1, lines[i]))
	}

	return fantasy.NewTextResponse(sb.String()), nil
}

func NewViewTool() fantasy.AgentTool {
	return fantasy.NewAgentTool(
		"view",
		"Read and display file contents with line numbers. Use start_line and end_line (1-based) to view specific portions. Examples: view lines 10-50, view from line 100.",
		View,
	)
}
