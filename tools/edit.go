package tools

import (
	"context"
	"fmt"
	"os"
	"strings"

	"charm.land/fantasy"
)

type EditInput struct {
	Path       string `json:"path" description:"The path to the file to edit"`
	OldString  string `json:"old_string" description:"The exact string to find and replace"`
	NewString  string `json:"new_string" description:"The string to replace with"`
	ReplaceAll bool   `json:"replace_all,omitempty" description:"If true, replace all occurrences (default false, replace first only)"`
}

func Edit(ctx context.Context, input EditInput, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
	if input.Path == "" {
		return fantasy.ToolResponse{}, fmt.Errorf("path is required")
	}
	if input.OldString == "" {
		return fantasy.ToolResponse{}, fmt.Errorf("old_string is required")
	}
	if input.OldString == input.NewString {
		return fantasy.NewTextErrorResponse("old_string and new_string are the same, no change needed"), nil
	}

	// 读取文件
	content, err := os.ReadFile(input.Path)
	if err != nil {
		return fantasy.NewTextErrorResponse(fmt.Sprintf("Failed to read file: %v", err)), nil
	}

	originalContent := string(content)

	// 检查是否存在要替换的字符串
	count := strings.Count(originalContent, input.OldString)
	if count == 0 {
		return fantasy.NewTextErrorResponse(fmt.Sprintf("old_string not found in %s", input.Path)), nil
	}

	// 如果不是 ReplaceAll 且有多个匹配，提示用户
	if !input.ReplaceAll && count > 1 {
		return fantasy.NewTextErrorResponse(fmt.Sprintf(
			"Found %d occurrences of old_string. Use replace_all=true to replace all, or provide more context to make it unique.",
			count,
		)), nil
	}

	// 执行替换
	var newContent string
	var replacedCount int
	if input.ReplaceAll {
		newContent = strings.ReplaceAll(originalContent, input.OldString, input.NewString)
		replacedCount = count
	} else {
		newContent = strings.Replace(originalContent, input.OldString, input.NewString, 1)
		replacedCount = 1
	}

	// 写入文件
	err = os.WriteFile(input.Path, []byte(newContent), 0644)
	if err != nil {
		return fantasy.NewTextErrorResponse(fmt.Sprintf("Failed to write file: %v", err)), nil
	}

	// 构建输出
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Edited %s: replaced %d occurrence(s)\n\n", input.Path, replacedCount))

	// 显示修改摘要
	oldLines := len(strings.Split(input.OldString, "\n"))
	newLines := len(strings.Split(input.NewString, "\n"))
	sb.WriteString(fmt.Sprintf("-%d lines\n+%d lines", oldLines, newLines))

	return fantasy.NewTextResponse(sb.String()), nil
}

func NewEditTool() fantasy.AgentTool {
	return fantasy.NewAgentTool(
		"edit",
		"Edit a file by replacing a specific string with a new string. Requires exact match of old_string. Use replace_all=true to replace all occurrences.",
		Edit,
	)
}
