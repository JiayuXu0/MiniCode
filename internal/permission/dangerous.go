package permission

import (
	"fmt"
	"strings"
)

// DangerousCommands 危险命令前缀列表
var DangerousCommands = []string{
	"rm ",
	"rm -",
	"rmdir",
	"del ",
	"deltree",
	"format",
	"mkfs",
	"> /dev/",
	"dd ",
	"chmod 777",
	"chmod -R",
	"chown",
	"sudo ",
	"su ",
	"kill ",
	"killall",
	"pkill",
	"shutdown",
	"reboot",
	"halt",
	"poweroff",
	"git push --force",
	"git push -f",
	"git reset --hard",
	"drop table",
	"drop database",
	"truncate table",
	"curl | sh",
	"curl | bash",
	"wget | sh",
	"wget | bash",
}

// IsDangerousCommand 检查命令是否危险
func IsDangerousCommand(cmd string) bool {
	cmdLower := strings.ToLower(strings.TrimSpace(cmd))
	for _, dangerous := range DangerousCommands {
		if strings.Contains(cmdLower, strings.ToLower(dangerous)) {
			return true
		}
	}
	return false
}

// GetCommandDescription 获取命令的危险描述
func GetCommandDescription(cmd string) string {
	cmdLower := strings.ToLower(cmd)

	switch {
	case strings.Contains(cmdLower, "rm ") || strings.Contains(cmdLower, "rm -"):
		return "This operation will delete files or directories and may be irreversible."
	case strings.Contains(cmdLower, "git push --force") || strings.Contains(cmdLower, "git push -f"):
		return "This operation will force push and may overwrite remote history."
	case strings.Contains(cmdLower, "git reset --hard"):
		return "This operation will discard all uncommitted changes."
	case strings.Contains(cmdLower, "chmod"):
		return "This operation will modify file permissions."
	case strings.Contains(cmdLower, "chown"):
		return "This operation will change file ownership."
	case strings.Contains(cmdLower, "sudo"):
		return "This operation requires administrator privileges."
	case strings.Contains(cmdLower, "kill") || strings.Contains(cmdLower, "pkill"):
		return "This operation will terminate processes."
	case strings.Contains(cmdLower, "dd "):
		return "This operation performs low-level disk operations."
	default:
		return "This operation may modify system state."
	}
}

// GetWriteDescription 获取写入操作的描述
func GetWriteDescription(filePath string, contentLen int) string {
	return fmt.Sprintf("Write %d bytes to: %s", contentLen, filePath)
}

// GetEditDescription 获取编辑操作的描述
func GetEditDescription(filePath string) string {
	return "Edit file: " + filePath
}
