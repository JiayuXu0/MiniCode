package tui

import "strings"

// truncateLine 截断过长的行
func truncateLine(s string, maxLen int) string {
	if maxLen <= 0 {
		maxLen = 50
	}
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// wrapText 将文本按指定宽度自动换行
func wrapText(text string, width int) []string {
	if width <= 0 {
		width = 80
	}

	var lines []string
	// 先按原有换行符分割
	paragraphs := strings.Split(text, "\n")

	for _, para := range paragraphs {
		if len(para) == 0 {
			lines = append(lines, "")
			continue
		}

		// 按宽度分割每个段落（支持中文）
		runes := []rune(para)
		for len(runes) > 0 {
			end := width
			if end > len(runes) {
				end = len(runes)
			}
			lines = append(lines, string(runes[:end]))
			runes = runes[end:]
		}
	}

	return lines
}
