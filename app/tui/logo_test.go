package tui

import (
	"strings"
	"testing"
)

func TestLogoRenders(t *testing.T) {
	ApplyTheme("tokyonight", true, true)

	for _, frame := range []int{0, 3, 7, 100} {
		out := logo(frame)
		if out == "" {
			t.Fatalf("logo(%d) empty", frame)
		}
		// ANSI Shadow art is 6 rows tall.
		if got := strings.Count(out, "\n") + 1; got != len(ansiShadow) {
			t.Errorf("logo has %d rows, want %d", got, len(ansiShadow))
		}
	}
}

func TestLogoASCIIFallback(t *testing.T) {
	ApplyTheme("tokyonight", true, false) // unicode off
	out := logo(0)
	if strings.Contains(out, "█") || strings.Contains(out, "╗") {
		t.Errorf("ascii fallback should not contain block/box runes:\n%s", out)
	}
	if !strings.Contains(out, "IPTV") {
		t.Errorf("ascii fallback should still show the wordmark:\n%s", out)
	}
	ApplyTheme("tokyonight", true, true) // restore
}
