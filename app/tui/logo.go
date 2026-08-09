package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ansiShadow is "IPTV TUI" in the FIGlet "ANSI Shadow" font — the same 3D
// block/shadow style used by the statusline installer splash. The solid blocks
// (█) catch the animated sheen; the line-drawing edges (═ ║ ╗ …) are the 3D
// shadow and are tinted darker for depth.
var ansiShadow = []string{
	`██╗██████╗ ████████╗██╗   ██╗    ████████╗██╗   ██╗██╗`,
	`██║██╔══██╗╚══██╔══╝██║   ██║    ╚══██╔══╝██║   ██║██║`,
	`██║██████╔╝   ██║   ██║   ██║       ██║   ██║   ██║██║`,
	`██║██╔═══╝    ██║   ╚██╗ ██╔╝       ██║   ██║   ██║██║`,
	`██║██║        ██║    ╚████╔╝        ██║   ╚██████╔╝██║`,
	`╚═╝╚═╝        ╚═╝     ╚═══╝         ╚═╝    ╚═════╝ ╚═╝`,
}

func isShadowEdge(r rune) bool {
	switch r {
	case '═', '║', '╗', '╔', '╝', '╚', '╣', '╠', '╦', '╩', '╬':
		return true
	}
	return false
}

// logo renders the animated "IPTV TUI" splash. `frame` advances the diagonal
// sheen sweeping across the solid blocks. On non-UTF-8 terminals it falls back
// to a plain bold wordmark.
func logo(frame int) string {
	if !unicodeOK {
		return lipgloss.NewStyle().Foreground(cPurple).Background(cBg).Bold(true).Render("IPTV  TUI")
	}

	var b strings.Builder
	for y, line := range ansiShadow {
		for x, r := range []rune(line) {
			switch {
			case r == '█':
				color := logoGradient[mod(x-y+frame, len(logoGradient))]
				b.WriteString(lipgloss.NewStyle().Foreground(color).Background(cBg).Render("█"))
			case isShadowEdge(r):
				b.WriteString(lipgloss.NewStyle().Foreground(cShadow).Background(cBg).Render(string(r)))
			default:
				b.WriteString(canvasStyle.Render(" "))
			}
		}
		if y < len(ansiShadow)-1 {
			b.WriteString("\n")
		}
	}
	return b.String()
}

func mod(a, n int) int {
	m := a % n
	if m < 0 {
		m += n
	}
	return m
}
