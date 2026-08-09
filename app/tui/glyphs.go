package tui

import "github.com/kinncj/IPTV-TUI/common/probe"

// gl returns the unicode glyph when the terminal supports UTF-8, else an ASCII
// fallback — so the UI stays legible on a bare tty.
func gl(unicode, ascii string) string {
	if unicodeOK {
		return unicode
	}
	return ascii
}

func glChevron() string { return gl("❯", ">") }
func glBar() string     { return gl("▌", "|") }
func glBullet() string  { return gl("▪", "*") }
func glCaret() string   { return gl("▌", "_") }
func glArrow() string   { return gl("→", "->") }
func glPlay() string    { return gl("▶", ">") }
func glDot() string     { return gl("·", "-") }
func glStar() string    { return gl("★", "*") }
func glBlock() string   { return gl("█", "#") }
func glSbUp() string    { return gl("▲", "^") }
func glSbDown() string  { return gl("▼", "v") }
func glSbThumb() string { return gl("█", "|") }
func glSbTrack() string { return gl("░", ":") }

// statusIcon renders a probe status, unicode-aware.
func statusIcon(s probe.Status) string {
	if unicodeOK {
		return s.Icon()
	}
	switch s {
	case probe.Checking:
		return "~"
	case probe.OK:
		return "+"
	case probe.Blocked:
		return "!"
	case probe.Dead:
		return "x"
	default:
		return " "
	}
}
