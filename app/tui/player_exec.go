package tui

import (
	"io"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/kinncj/IPTV-TUI/common/player"
)

// playerModal is a Bubble Tea ExecCommand that paints a themed player screen and
// runs mpv as a centered modal inside it. Bubble Tea releases the terminal for
// the duration, so the browser is frozen behind; this makes the handoff read as
// a modal player rather than the TUI disappearing. A live overlay on top of the
// running list is not possible: terminal video and the TUI cannot share the
// screen.
type playerModal struct {
	title string
	url   string
	vo    string
	w, h  int

	in   io.Reader
	out  io.Writer
	errw io.Writer
}

func (p *playerModal) SetStdin(r io.Reader)  { p.in = r }
func (p *playerModal) SetStdout(w io.Writer) { p.out = w }
func (p *playerModal) SetStderr(w io.Writer) { p.errw = w }

func (p *playerModal) Run() error {
	p.drawChrome()

	cmd := player.TerminalCmd(p.url, p.vo)
	cmd.Stdin = p.in
	cmd.Stdout = p.out
	cmd.Stderr = p.errw
	err := cmd.Run()

	// Clear scrollback + screen and restore the cursor before Bubble Tea resumes.
	io.WriteString(p.out, "\x1b[?25h\x1b[2J\x1b[3J\x1b[H")
	return err
}

// drawChrome fills the screen with the theme background and draws a title bar,
// so mpv's centered modal sits on a themed player screen instead of blank space.
func (p *playerModal) drawChrome() {
	if p.out == nil {
		return
	}
	w, h := p.w, p.h
	if w < 1 {
		w = 80
	}
	if h < 1 {
		h = 24
	}

	// Hide cursor, clear.
	io.WriteString(p.out, "\x1b[?25l\x1b[2J\x1b[H")

	// Fill the background so the modal's margins are themed, not terminal default.
	bg := lipgloss.NewStyle().Background(cBg).Width(w)
	var fill strings.Builder
	for i := 0; i < h; i++ {
		fill.WriteString(bg.Render(""))
		if i < h-1 {
			fill.WriteString("\n")
		}
	}
	io.WriteString(p.out, fill.String())

	// Title/help bar at the top.
	io.WriteString(p.out, "\x1b[H")
	left := glPlay() + " " + truncate(p.title, max(1, w-40))
	right := "f fullscreen  " + glDot() + "  q quit"
	gap := w - lipgloss.Width(left) - lipgloss.Width(right) - 2
	if gap < 1 {
		gap = 1
	}
	line := left + strings.Repeat(" ", gap) + right
	bar := lipgloss.NewStyle().Background(cBgElev).Foreground(cPurple).Bold(true).
		Width(w).Padding(0, 1).Render(line)
	io.WriteString(p.out, bar)

	// Park the cursor just under the bar; mpv paints its modal from here.
	io.WriteString(p.out, "\x1b[3;1H")
}
