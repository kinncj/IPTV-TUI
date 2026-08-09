// Package termcaps detects terminal capabilities so the UI can be fancy when the
// environment is fancy and still look good when it isn't. It centralizes all the
// env-sniffing in one place; the UI reads a plain Caps struct.
package termcaps

import (
	"os"
	"strings"

	"github.com/muesli/termenv"
)

// Caps describes what the current terminal can do.
type Caps struct {
	Profile       termenv.Profile // color depth (TrueColor / ANSI256 / ANSI)
	TrueColor     bool
	DarkBG        bool   // terminal background is dark (follows OS/terminal theme)
	Unicode       bool   // locale is UTF-8 (block chars, borders, icons safe)
	TermProgram   string // e.g. "ghostty", "wezterm", "kitty"
	Multiplexer   string // "tmux", "herdr", or ""
	Ghostty       bool
	KittyGraphics bool // supports the kitty graphics protocol (pixel-perfect inline video)
	Sixel         bool // supports sixel graphics
}

// TerminalVO picks the best mpv terminal video output for inline playback:
// kitty graphics where available (crisp), then sixel, else truecolor half-blocks
// (works in any terminal). Audio plays regardless.
func (c Caps) TerminalVO() string {
	switch {
	case c.KittyGraphics:
		return "kitty"
	case c.Sixel:
		return "sixel"
	default:
		return "tct"
	}
}

// Detect inspects the environment and the terminal.
func Detect() Caps {
	out := termenv.NewOutput(os.Stdout)

	c := Caps{
		Profile:     out.ColorProfile(),
		DarkBG:      out.HasDarkBackground(),
		Unicode:     unicode(),
		TermProgram: strings.ToLower(os.Getenv("TERM_PROGRAM")),
		Multiplexer: multiplexer(),
	}

	c.TrueColor = c.Profile == termenv.TrueColor ||
		hasAny(strings.ToLower(os.Getenv("COLORTERM")), "truecolor", "24bit")
	c.Ghostty = c.TermProgram == "ghostty" || os.Getenv("GHOSTTY_RESOURCES_DIR") != ""

	term := strings.ToLower(os.Getenv("TERM"))
	c.KittyGraphics = c.Ghostty || c.TermProgram == "wezterm" ||
		os.Getenv("KITTY_WINDOW_ID") != "" || strings.Contains(term, "kitty")
	c.Sixel = c.TermProgram == "wezterm" || c.TermProgram == "mintty" ||
		hasAny(term, "foot", "sixel", "mlterm", "yaft")

	// Ghostty (and anything advertising truecolor) is genuinely truecolor even
	// when a multiplexer downgrades the reported profile.
	if c.Ghostty || c.TrueColor {
		c.TrueColor = true
		c.Profile = termenv.TrueColor
	}
	return c
}

func (c Caps) Summary() string {
	depth := "ansi"
	switch c.Profile {
	case termenv.TrueColor:
		depth = "truecolor"
	case termenv.ANSI256:
		depth = "256"
	}
	bg := "dark"
	if !c.DarkBG {
		bg = "light"
	}
	parts := []string{depth, bg}
	if c.Unicode {
		parts = append(parts, "unicode")
	} else {
		parts = append(parts, "ascii")
	}
	if c.Ghostty {
		parts = append(parts, "ghostty")
	} else if c.TermProgram != "" {
		parts = append(parts, c.TermProgram)
	}
	if c.Multiplexer != "" {
		parts = append(parts, c.Multiplexer)
	}
	return strings.Join(parts, " · ")
}

func unicode() bool {
	for _, k := range []string{"LC_ALL", "LC_CTYPE", "LANG"} {
		v := strings.ToUpper(os.Getenv(k))
		if strings.Contains(v, "UTF-8") || strings.Contains(v, "UTF8") {
			return true
		}
	}
	return false
}

func multiplexer() string {
	if os.Getenv("HERDR") != "" || os.Getenv("HERDR_SESSION") != "" ||
		strings.Contains(strings.ToLower(os.Getenv("TERM")), "herdr") {
		return "herdr"
	}
	if os.Getenv("TMUX") != "" {
		return "tmux"
	}
	return ""
}

func hasAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}
