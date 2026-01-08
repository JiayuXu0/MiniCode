package tools

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"charm.land/fantasy"
)

type BashInput struct {
	Command string `json:"command" description:"The bash command to execute"`
	Timeout int    `json:"timeout,omitempty" description:"Timeout in seconds (default 30, max 300)"`
}

const (
	defaultTimeout = 30
	maxTimeout     = 300
	maxOutputLen   = 10000 // 最大输出长度
)

// 危险命令黑名单
var dangerousCommands = []string{
	"rm -rf /",
	"rm -rf /*",
	"mkfs",
	"dd if=",
	":(){:|:&};:",
	"> /dev/sda",
	"chmod -R 777 /",
}

func Bash(ctx context.Context, input BashInput, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
	if input.Command == "" {
		return fantasy.ToolResponse{}, fmt.Errorf("command is required")
	}

	// 检查危险命令
	cmdLower := strings.ToLower(input.Command)
	for _, dangerous := range dangerousCommands {
		if strings.Contains(cmdLower, dangerous) {
			return fantasy.NewTextErrorResponse(fmt.Sprintf("Dangerous command blocked: %s", input.Command)), nil
		}
	}

	// 设置超时
	timeout := input.Timeout
	if timeout <= 0 {
		timeout = defaultTimeout
	}
	if timeout > maxTimeout {
		timeout = maxTimeout
	}

	// 创建带超时的 context
	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	// 执行命令
	cmd := exec.CommandContext(ctx, "bash", "-c", input.Command)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	// 构建输出
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("$ %s\n\n", input.Command))

	// 处理输出
	output := stdout.String()
	errOutput := stderr.String()

	if output != "" {
		if len(output) > maxOutputLen {
			output = output[:maxOutputLen] + fmt.Sprintf("\n... (output truncated, showing first %d chars)", maxOutputLen)
		}
		sb.WriteString(output)
	}

	if errOutput != "" {
		if len(errOutput) > maxOutputLen {
			errOutput = errOutput[:maxOutputLen] + fmt.Sprintf("\n... (stderr truncated, showing first %d chars)", maxOutputLen)
		}
		if output != "" {
			sb.WriteString("\n")
		}
		sb.WriteString("STDERR:\n")
		sb.WriteString(errOutput)
	}

	// 处理错误
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			sb.WriteString(fmt.Sprintf("\n\nCommand timed out after %d seconds", timeout))
		} else if exitErr, ok := err.(*exec.ExitError); ok {
			sb.WriteString(fmt.Sprintf("\n\nExit code: %d", exitErr.ExitCode()))
		} else {
			sb.WriteString(fmt.Sprintf("\n\nError: %v", err))
		}
	} else {
		sb.WriteString("\n\nExit code: 0")
	}

	return fantasy.NewTextResponse(sb.String()), nil
}

func NewBashTool() fantasy.AgentTool {
	return fantasy.NewAgentTool(
		"bash",
		"Execute a bash command and return the output. Use for system commands, file operations, git, etc. Timeout default 30s, max 300s.",
		Bash,
	)
}
