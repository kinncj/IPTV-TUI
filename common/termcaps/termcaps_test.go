package termcaps

import (
	"strings"
	"testing"

	"github.com/muesli/termenv"
)

func clearEnv(t *testing.T) {
	for _, k := range []string{"LC_ALL", "LC_CTYPE", "LANG", "COLORTERM", "TERM_PROGRAM",
		"GHOSTTY_RESOURCES_DIR", "TMUX", "HERDR", "HERDR_SESSION", "TERM"} {
		t.Setenv(k, "")
	}
}

func TestUnicodeDetection(t *testing.T) {
	clearEnv(t)
	t.Setenv("LANG", "en_US.UTF-8")
	if !Detect().Unicode {
		t.Error("UTF-8 LANG should report unicode")
	}

	clearEnv(t)
	t.Setenv("LANG", "C")
	if Detect().Unicode {
		t.Error("C locale should not report unicode")
	}
}

func TestTrueColorAndGhostty(t *testing.T) {
	clearEnv(t)
	t.Setenv("COLORTERM", "truecolor")
	c := Detect()
	if !c.TrueColor || c.Profile != termenv.TrueColor {
		t.Errorf("COLORTERM=truecolor should force truecolor profile, got %v", c.Profile)
	}

	clearEnv(t)
	t.Setenv("TERM_PROGRAM", "ghostty")
	c = Detect()
	if !c.Ghostty || !c.TrueColor {
		t.Errorf("ghostty should be detected as truecolor, got %+v", c)
	}
}

func TestMultiplexerPrecedence(t *testing.T) {
	clearEnv(t)
	t.Setenv("TMUX", "/tmp/tmux-1000/default,1,0")
	if got := Detect().Multiplexer; got != "tmux" {
		t.Errorf("multiplexer = %q, want tmux", got)
	}

	// herdr wins over tmux when both are present.
	t.Setenv("HERDR", "1")
	if got := Detect().Multiplexer; got != "herdr" {
		t.Errorf("multiplexer = %q, want herdr", got)
	}
}

func TestSummary(t *testing.T) {
	clearEnv(t)
	t.Setenv("COLORTERM", "truecolor")
	t.Setenv("LANG", "en_US.UTF-8")
	t.Setenv("TERM_PROGRAM", "ghostty")
	s := Detect().Summary()
	for _, want := range []string{"truecolor", "unicode", "ghostty"} {
		if !strings.Contains(s, want) {
			t.Errorf("summary %q missing %q", s, want)
		}
	}
}
