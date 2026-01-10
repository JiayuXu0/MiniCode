package styles

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Gradient 表示颜色渐变
type Gradient struct {
	colors []lipgloss.Color
}

// NewGradient 创建渐变
func NewGradient(colors ...lipgloss.Color) *Gradient {
	return &Gradient{colors: colors}
}

// At 获取渐变中某个位置的颜色 (t: 0.0 ~ 1.0)
func (g *Gradient) At(t float64) lipgloss.Color {
	if len(g.colors) == 0 {
		return lipgloss.Color("")
	}
	if len(g.colors) == 1 {
		return g.colors[0]
	}

	t = math.Max(0, math.Min(1, t))

	// 计算在哪两个颜色之间
	segments := float64(len(g.colors) - 1)
	segment := int(t * segments)
	if segment >= len(g.colors)-1 {
		segment = len(g.colors) - 2
	}

	// 在两个颜色之间插值
	localT := (t * segments) - float64(segment)

	c1 := g.colors[segment]
	c2 := g.colors[segment+1]

	return interpolateColor(c1, c2, localT)
}

// parseHexColor 解析十六进制颜色
func parseHexColor(hex string) (r, g, b uint8) {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return 0, 0, 0
	}

	rVal, _ := strconv.ParseUint(hex[0:2], 16, 8)
	gVal, _ := strconv.ParseUint(hex[2:4], 16, 8)
	bVal, _ := strconv.ParseUint(hex[4:6], 16, 8)

	return uint8(rVal), uint8(gVal), uint8(bVal)
}

// interpolateColor 插值两个颜色
func interpolateColor(c1, c2 lipgloss.Color, t float64) lipgloss.Color {
	r1, g1, b1 := parseHexColor(string(c1))
	r2, g2, b2 := parseHexColor(string(c2))

	r := uint8(float64(r1) + t*(float64(r2)-float64(r1)))
	g := uint8(float64(g1) + t*(float64(g2)-float64(g1)))
	b := uint8(float64(b1) + t*(float64(b2)-float64(b1)))

	return lipgloss.Color(fmt.Sprintf("#%02X%02X%02X", r, g, b))
}

// RenderGradientText 渲染渐变文本
func RenderGradientText(text string, g *Gradient) string {
	if len(text) == 0 {
		return ""
	}

	runes := []rune(text)
	result := ""

	for i, r := range runes {
		var t float64
		if len(runes) > 1 {
			t = float64(i) / float64(len(runes)-1)
		}
		color := g.At(t)
		style := lipgloss.NewStyle().Foreground(color)
		result += style.Render(string(r))
	}

	return result
}
