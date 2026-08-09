package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Block letters, 5 rows tall, matching the chunky pixel-font logos in the
// reference TUIs.
var glyphs = map[rune][]string{
	'I': {"█████", "  █  ", "  █  ", "  █  ", "█████"},
	'P': {"████ ", "█   █", "████ ", "█    ", "█    "},
	'T': {"█████", "  █  ", "  █  ", "  █  ", "  █  "},
	'V': {"█   █", "█   █", "█   █", " █ █ ", "  █  "},
}

const (
	glyphH   = 5
	glyphW   = 5
	glyphGap = 2
)

// logo renders `word` as an animated, 3D-extruded block logo. `frame` advances
// the diagonal sheen that sweeps across the letters; pass a steady counter from
// the tick loop to animate it.
func logo(word string, frame int) string {
	// 1. Rasterize the word into a boolean grid of filled cells.
	w := len(word)*glyphW + max(0, len(word)-1)*glyphGap
	fill := make([][]bool, glyphH)
	for y := range fill {
		fill[y] = make([]bool, w)
	}
	col := 0
	for i, r := range word {
		g, ok := glyphs[r]
		if !ok {
			g = []string{"     ", "     ", "     ", "     ", "     "}
		}
		for y := 0; y < glyphH; y++ {
			// Index by rune, not byte: '█' is multi-byte in UTF-8.
			for x, ch := range []rune(g[y]) {
				if ch == '█' && col+x < w {
					fill[y][col+x] = true
				}
			}
		}
		col += glyphW
		if i < len(word)-1 {
			col += glyphGap
		}
	}

	// 2. Composite a canvas one cell larger to hold the drop-shadow extrusion
	//    (offset down-right), then paint front faces with the moving sheen.
	ch, cw := glyphH+1, w+1
	var b strings.Builder
	for y := 0; y < ch; y++ {
		for x := 0; x < cw; x++ {
			front := y < glyphH && x < w && fill[y][x]
			shadow := y > 0 && x > 0 && (y-1) < glyphH && (x-1) < w && fill[y-1][x-1]

			block := glBlock()
			switch {
			case front:
				color := logoGradient[mod(x-y+frame, len(logoGradient))]
				b.WriteString(lipgloss.NewStyle().Foreground(color).Background(cBg).Render(block))
			case shadow:
				b.WriteString(lipgloss.NewStyle().Foreground(cShadow).Background(cBg).Render(block))
			default:
				b.WriteString(canvasStyle.Render(" "))
			}
		}
		if y < ch-1 {
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
