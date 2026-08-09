package tui

import (
	"strings"
	"testing"
)

func TestLogoRenders(t *testing.T) {
	ApplyTheme("tokyonight", true, true)

	for _, word := range []string{"IPTV", "I", "PTV", "XYZ" /* unknown glyphs */} {
		for _, frame := range []int{0, 3, 7, 100} {
			out := logo(word, frame)
			if out == "" {
				t.Fatalf("logo(%q,%d) empty", word, frame)
			}
			// One row per glyph height plus the extrusion row.
			if got := strings.Count(out, "\n") + 1; got != glyphH+1 {
				t.Errorf("logo(%q) has %d rows, want %d", word, got, glyphH+1)
			}
		}
	}
}

func TestLogoASCIIFallback(t *testing.T) {
	ApplyTheme("tokyonight", true, false) // unicode off
	out := logo("IPTV", 0)
	if strings.Contains(out, "█") {
		t.Errorf("ascii fallback should not contain block runes:\n%s", out)
	}
	if !strings.Contains(out, "#") {
		t.Errorf("ascii fallback should use '#':\n%s", out)
	}
	ApplyTheme("tokyonight", true, true) // restore
}
