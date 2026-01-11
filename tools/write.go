package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"charm.land/fantasy"
	"github.com/JiayuXu0/MiniCode/internal/permission"
)

type WriteInput struct {
	Path    string `json:"path" description:"The path to the file to write"`
	Content string `json:"content" description:"The content to write to the file"`
	Append  bool   `json:"append,omitempty" description:"If true, append to file instead of overwriting (default false)"`
}

func NewWriteTool(perm *permission.Service) fantasy.AgentTool {
	return fantasy.NewAgentTool(
		"write",
		"Write content to a file. Creates the file and parent directories if they don't exist. Use append=true to append instead of overwrite.",
		func(ctx context.Context, input WriteInput, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			if input.Path == "" {
				return fantasy.ToolResponse{}, fmt.Errorf("path is required")
			}
			if input.Content == "" {
				return fantasy.ToolResponse{}, fmt.Errorf("content is required")
			}

			// 请求权限
			desc := permission.GetWriteDescription(input.Path, len(input.Content))
			if !perm.Request("write", input.Path, desc) {
				return fantasy.NewTextErrorResponse("Permission denied by user"), nil
			}

			// 确保父目录存在
			dir := filepath.Dir(input.Path)
			if dir != "." && dir != "/" {
				if err := os.MkdirAll(dir, 0755); err != nil {
					return fantasy.NewTextErrorResponse(fmt.Sprintf("Failed to create directory %s: %v", dir, err)), nil
				}
			}

			// 检查文件是否存在（用于输出信息）
			_, fileExists := os.Stat(input.Path)
			existed := fileExists == nil

			// 确定写入模式
			var flag int
			if input.Append {
				flag = os.O_WRONLY | os.O_CREATE | os.O_APPEND
			} else {
				flag = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
			}

			// 写入文件
			file, err := os.OpenFile(input.Path, flag, 0644)
			if err != nil {
				return fantasy.NewTextErrorResponse(fmt.Sprintf("Failed to open file: %v", err)), nil
			}
			defer file.Close()

			n, err := file.WriteString(input.Content)
			if err != nil {
				return fantasy.NewTextErrorResponse(fmt.Sprintf("Failed to write file: %v", err)), nil
			}

			// 构建输出
			var action string
			if input.Append {
				action = "Appended to"
			} else if existed {
				action = "Overwrote"
			} else {
				action = "Created"
			}

			return fantasy.NewTextResponse(fmt.Sprintf("%s %s (%d bytes written)", action, input.Path, n)), nil
		},
	)
}
