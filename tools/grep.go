package tools

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"charm.land/fantasy"
	"github.com/bmatcuk/doublestar/v4"
)

type GrepInput struct {
	Pattern string `json:"pattern" description:"The regex pattern to search for in file contents"`
	Path    string `json:"path,omitempty" description:"The directory to search in (defaults to current directory)"`
	Include string `json:"include,omitempty" description:"File pattern to include (e.g., *.go, *.{js,ts})"`
}

type GrepMatch struct {
	File    string `json:"file"`
	Line    int    `json:"line"`
	Content string `json:"content"`
}

func Grep(ctx context.Context, input GrepInput, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
	if input.Pattern == "" {
		return fantasy.ToolResponse{}, fmt.Errorf("pattern is required")
	}

	// 编译正则表达式
	re, err := regexp.Compile(input.Pattern)
	if err != nil {
		return fantasy.NewTextErrorResponse(fmt.Sprintf("invalid regex pattern: %v", err)), nil
	}

	// 默认搜索当前目录
	searchPath := input.Path
	if searchPath == "" {
		searchPath = "."
	}

	// 转换为绝对路径
	absPath, err := filepath.Abs(searchPath)
	if err != nil {
		return fantasy.NewTextErrorResponse(fmt.Sprintf("cannot resolve path: %v", err)), nil
	}

	// 检查目录是否存在
	info, err := os.Stat(absPath)
	if err != nil {
		return fantasy.NewTextErrorResponse(fmt.Sprintf("directory not found: %v", err)), nil
	}
	if !info.IsDir() {
		return fantasy.NewTextErrorResponse(fmt.Sprintf("path is not a directory: %s", absPath)), nil
	}

	// 获取要搜索的文件列表
	var files []string
	if input.Include != "" {
		// 使用 include 模式匹配文件
		pattern := filepath.Join(absPath, "**", input.Include)
		files, err = doublestar.FilepathGlob(pattern)
		if err != nil {
			return fantasy.NewTextErrorResponse(fmt.Sprintf("glob pattern error: %v", err)), nil
		}
	} else {
		// 遍历所有文件
		err = filepath.Walk(absPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // 忽略错误，继续遍历
			}
			if !info.IsDir() {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			return fantasy.NewTextErrorResponse(fmt.Sprintf("walk directory error: %v", err)), nil
		}
	}

	// 搜索文件内容
	var matches []GrepMatch
	const maxMatches = 100 // 限制最大匹配数

	for _, file := range files {
		if len(matches) >= maxMatches {
			break
		}

		fileMatches, err := searchFile(file, re, maxMatches-len(matches))
		if err != nil {
			continue // 忽略无法读取的文件
		}
		matches = append(matches, fileMatches...)
	}

	if len(matches) == 0 {
		return fantasy.NewTextResponse(fmt.Sprintf("No matches found for pattern '%s'", input.Pattern)), nil
	}

	// 构建输出
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d matches for pattern '%s':\n\n", len(matches), input.Pattern))

	currentFile := ""
	for _, m := range matches {
		relPath, _ := filepath.Rel(absPath, m.File)
		if relPath == "" {
			relPath = m.File
		}

		if relPath != currentFile {
			if currentFile != "" {
				sb.WriteString("\n")
			}
			sb.WriteString(fmt.Sprintf("%s:\n", relPath))
			currentFile = relPath
		}
		sb.WriteString(fmt.Sprintf("  %4d: %s\n", m.Line, strings.TrimSpace(m.Content)))
	}

	if len(matches) >= maxMatches {
		sb.WriteString(fmt.Sprintf("\n... (showing first %d matches)\n", maxMatches))
	}

	return fantasy.NewTextResponse(sb.String()), nil
}

func searchFile(path string, re *regexp.Regexp, limit int) ([]GrepMatch, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var matches []GrepMatch
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() && len(matches) < limit {
		lineNum++
		line := scanner.Text()
		if re.MatchString(line) {
			matches = append(matches, GrepMatch{
				File:    path,
				Line:    lineNum,
				Content: line,
			})
		}
	}

	return matches, scanner.Err()
}

func NewGrepTool() fantasy.AgentTool {
	return fantasy.NewParallelAgentTool(
		"grep",
		"Search file contents using regular expressions. Returns matching lines with file paths and line numbers.",
		Grep,
	)
}
