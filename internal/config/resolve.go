package config

import (
	"context"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// ResolveVariables 解析配置中的变量
// 支持:
//   - $VAR 或 ${VAR} - 环境变量
//   - $(command) - 命令替换
func ResolveVariables(value string) string {
	// 1. 解析 ${VAR} 格式
	value = resolveBraceVars(value)

	// 2. 解析 $VAR 格式
	value = resolveSimpleVars(value)

	// 3. 解析 $(command) 格式
	value = resolveCommands(value)

	return value
}

// resolveBraceVars 解析 ${VAR} 格式
func resolveBraceVars(value string) string {
	re := regexp.MustCompile(`\$\{([^}]+)\}`)
	return re.ReplaceAllStringFunc(value, func(match string) string {
		// 提取变量名
		varName := match[2 : len(match)-1]
		if envValue := os.Getenv(varName); envValue != "" {
			return envValue
		}
		return match // 保持原样
	})
}

// resolveSimpleVars 解析 $VAR 格式
func resolveSimpleVars(value string) string {
	re := regexp.MustCompile(`\$([A-Za-z_][A-Za-z0-9_]*)`)
	return re.ReplaceAllStringFunc(value, func(match string) string {
		varName := match[1:]
		if envValue := os.Getenv(varName); envValue != "" {
			return envValue
		}
		return match
	})
}

// resolveCommands 解析 $(command) 格式
func resolveCommands(value string) string {
	re := regexp.MustCompile(`\$\(([^)]+)\)`)
	return re.ReplaceAllStringFunc(value, func(match string) string {
		cmd := match[2 : len(match)-1]
		return executeCommand(cmd)
	})
}

// executeCommand 执行命令并返回输出
func executeCommand(command string) string {
	// 设置超时
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 使用 shell 执行
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(output))
}
