package tools

import (
	"context"
	"fmt"
	"os"
	"strings"

	"charm.land/fantasy"
)

type ViewInput struct {
	Path   string `json:"path" description:"The path to the file to view"`
	Offset int    `json:"offset,omitempty" description:"The line number to start viewing from (0-based)"`
	Limit  int    `json:"limit,omitempty" description:"The maximum number of lines to view (default 1000)"`
}

func (v *ViewInput) ApplyDefaults() {
	if v.Limit <= 0 {
		v.Limit = 1000
	}
}

func View(ctx context.Context, input ViewInput, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
	if input.Path == "" {
		return fantasy.ToolResponse{}, fmt.Errorf("path is required")
	}
	if input.Offset < 0 {
		return fantasy.ToolResponse{}, fmt.Errorf("offset must be non-negative")
	}
	input.ApplyDefaults()

	content, err := os.ReadFile(input.Path)
	if err != nil {
		return fantasy.ToolResponse{}, err
	}

	lines := strings.Split(string(content), "\n")
	totalLines := len(lines)

	// 计算实际的起止行
	start := input.Offset
	if start >= totalLines {
		return fantasy.NewTextResponse(fmt.Sprintf("File %s has only %d lines, offset %d is out of range", input.Path, totalLines, start)), nil
	}

	end := start + input.Limit
	if end > totalLines {
		end = totalLines
	}

	// 构建带行号的输出
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("File: %s (lines %d-%d of %d)\n\n", input.Path, start+1, end, totalLines))
	for i := start; i < end; i++ {
		sb.WriteString(fmt.Sprintf("%4d\t%s\n", i+1, lines[i]))
	}
	fmt.Printf("Viewing %s from line %d to %d", input.Path, start+1, end)
	return fantasy.NewTextResponse(sb.String()), nil
}

func NewViewTool() fantasy.AgentTool {
	return fantasy.NewParallelAgentTool(
		"view",
		"Read and display the contents of a file with line numbers. Use offset and limit to view specific portions of large files.",
		View,
	)
}
