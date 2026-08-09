package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kinncj/IPTV-TUI/common/inlinevid"
	"github.com/kinncj/IPTV-TUI/common/resolve"
)

// playerOption is a terminal-player backend the user can pick for `t`.
type playerOption struct {
	key   string
	label string
	desc  string
}

// playerOptions lists the available inline backends, most-preferred first.
func (m Model) playerOptions() []playerOption {
	var opts []playerOption
	if m.termVO != "" { // mpv is installed
		opts = append(opts, playerOption{
			key:   "mpv",
			label: "mpv (external)",
			desc:  "Sharp video. Takes over the terminal while it plays; press q in mpv to return.",
		})
	}
	if inlinevid.Available() {
		opts = append(opts, playerOption{
			key:   "builtin",
			label: "built-in (in the TUI)",
			desc:  "Blocky video in a modal; the browser stays visible behind it. Needs ffmpeg.",
		})
	}
	return opts
}

func modalBorder() lipgloss.Border {
	if unicodeOK {
		return lipgloss.RoundedBorder()
	}
	return lipgloss.NormalBorder()
}

// --- chooser ----------------------------------------------------------------

func (m Model) handlePickKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.picking = false
		return m, nil
	case "q", "ctrl+c":
		return m, tea.Quit
	case "up", "k":
		if m.pickIdx > 0 {
			m.pickIdx--
		}
		return m, nil
	case "down", "j":
		if m.pickIdx < len(m.pickOpts)-1 {
			m.pickIdx++
		}
		return m, nil
	case "enter":
		return m.playWith(m.pickOpts[m.pickIdx].key)
	}
	return m, nil
}

func (m Model) playWith(key string) (tea.Model, tea.Cmd) {
	m.picking = false
	// Resolve page URLs (YouTube, etc.) to a direct stream first, off the UI
	// goroutine, then play with the chosen backend.
	if resolve.Needed(m.vidURL) {
		m.toast = "resolving stream…"
		m.toastTTL = 40
	}
	url, backend, title := m.vidURL, key, m.vidTitle
	return m, tea.Batch(m.animate(), func() tea.Msg {
		return playResolvedMsg{url: resolve.Direct(url), backend: backend, title: title}
	})
}

func (m Model) chooserView() string {
	base := m.channelsView()

	title := lipgloss.NewStyle().Foreground(cPurple).Background(cBgElev).Bold(true).
		Render(glPlay() + " play " + truncate(m.vidTitle, 34))
	rows := []string{title, ""}
	for i, o := range m.pickOpts {
		sel := i == m.pickIdx
		bar := "  "
		nameStyle := lipgloss.NewStyle().Background(cBgElev).Foreground(cFg)
		if sel {
			bar = lipgloss.NewStyle().Foreground(cPurple).Background(cBgElev).Render(glBar() + " ")
			nameStyle = nameStyle.Foreground(cPurple).Bold(true)
		}
		rows = append(rows,
			bar+nameStyle.Render(o.label),
			lipgloss.NewStyle().Foreground(cMuted).Background(cBgElev).Render("    "+o.desc),
			"")
	}
	rows = append(rows, lipgloss.NewStyle().Foreground(cMuted).Background(cBgElev).
		Render("enter select   ·   esc cancel   ·   q quit"))

	box := lipgloss.NewStyle().Background(cBgElev).
		Border(modalBorder()).BorderForeground(cPurple).BorderBackground(cBg).
		Padding(1, 2).Render(lipgloss.JoinVertical(lipgloss.Left, rows...))
	return overlayCenter(base, box, m.width, m.height)
}

// --- inline video -----------------------------------------------------------

func (m Model) startInline() (tea.Model, tea.Cmd) {
	return m.newStream()
}

// newStream (re)opens the inline decoder at the current size and bumps the
// generation so any in-flight frame from a previous stream is ignored.
func (m Model) newStream() (tea.Model, tea.Cmd) {
	if m.vid != nil {
		m.vid.Close()
		m.vid = nil
	}
	mode := m.inlineMode
	if mode == "" {
		mode = inlinevid.ModeHalfBlock
	}
	fps := 15
	if mode == inlinevid.ModeKitty {
		fps = 12 // graphics frames are heavier; ease the bandwidth
	}
	cols, rows := m.videoCells()
	s, err := inlinevid.Open(m.vidURL, cols, rows, fps, mode)
	if err != nil {
		m.closeVid()
		m.toast = "could not start ffmpeg"
		m.toastTTL = 20
		return m, m.animate()
	}
	m.vid = s
	m.vidRows = nil
	m.vidPrelude = ""
	m.vidGen++
	return m, m.nextVidFrame()
}

// videoCells is the character size of the video area (inside the modal border),
// ~80% of the screen, or nearly full when toggled.
func (m Model) videoCells() (int, int) {
	pw, ph := 80, 78
	if m.vidFull {
		pw, ph = 98, 94
	}
	cols := m.width*pw/100 - 2
	rows := m.height*ph/100 - 4 // border + title + controls
	if cols < 8 {
		cols = 8
	}
	if rows < 4 {
		rows = 4
	}
	return cols, rows
}

func (m Model) nextVidFrame() tea.Cmd {
	s := m.vid
	gen := m.vidGen
	return func() tea.Msg {
		f, err := s.Next()
		if err != nil {
			return vidEndMsg{err: err, gen: gen}
		}
		return vidFrameMsg{prelude: f.Prelude, rows: f.Rows, gen: gen}
	}
}

func (m *Model) closeVid() {
	if m.vid != nil {
		m.vid.Close()
		m.vid = nil
	}
	m.vidGen++ // invalidate any in-flight frame
	m.vidRows = nil
	m.vidPrelude = ""
	m.vidFull = false
}

// kittyPlaceholder is U+10EEEE, the kitty unicode graphics placeholder. Rows of
// these have a visible width of one cell each, which lipgloss cannot measure, so
// vwidth counts them explicitly.
const kittyPlaceholder = "\U0010EEEE"

func vwidth(s string) int {
	if n := strings.Count(s, kittyPlaceholder); n > 0 {
		return n
	}
	return lipgloss.Width(s)
}

func centerTo(s string, w int) string {
	sw := lipgloss.Width(s)
	if sw >= w {
		return s
	}
	left := (w - sw) / 2
	pad := lipgloss.NewStyle().Background(cBg)
	return pad.Render(strings.Repeat(" ", left)) + s + pad.Render(strings.Repeat(" ", w-sw-left))
}

func (m Model) handleVideoKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.closeVid()
		return m, nil
	case "q", "ctrl+c":
		m.closeVid()
		return m, tea.Quit
	case "f":
		m.vidFull = !m.vidFull
		return m.newStream()
	}
	return m, nil
}

func (m Model) videoView() string {
	base := m.channelsView()

	sizeLabel := "fullscreen"
	if m.vidFull {
		sizeLabel = "modal"
	}
	titleStyle := lipgloss.NewStyle().Foreground(cPurple).Background(cBg).Bold(true)
	title := titleStyle.Render(glPlay() + " " + truncate(m.vidTitle, max(10, m.width/2)))
	controls := lipgloss.NewStyle().Foreground(cMuted).Background(cBg).
		Render("f " + sizeLabel + "   ·   esc back   ·   q quit")

	// Kitty graphics: the placeholder rows carry the image; do not wrap them in
	// a lipgloss border (it cannot measure them). Prepend the image transmit.
	if m.inlineMode == inlinevid.ModeKitty && len(m.vidRows) > 0 {
		vcols := vwidth(m.vidRows[0])
		lines := make([]string, 0, len(m.vidRows)+2)
		lines = append(lines, centerTo(title, vcols))
		lines = append(lines, m.vidRows...)
		lines = append(lines, centerTo(controls, vcols))
		block := strings.Join(lines, "\n")
		return m.vidPrelude + overlayCenter(base, block, m.width, m.height)
	}

	body := strings.Join(m.vidRows, "\n")
	if len(m.vidRows) == 0 {
		body = lipgloss.NewStyle().Foreground(cMuted).Background(cBg).Render("buffering…")
	}
	box := lipgloss.NewStyle().
		Border(modalBorder()).BorderForeground(cPurple).BorderBackground(cBg).
		Background(cBg).Render(body)
	block := lipgloss.JoinVertical(lipgloss.Center, title, box, controls)
	return overlayCenter(base, block, m.width, m.height)
}

// overlayCenter replaces the vertically-centered band of `base` with `over`,
// each line padded to full width so it sits centered. Rows above and below keep
// showing the base (the live list). Widths are measured with vwidth so kitty
// placeholder rows are handled.
func overlayCenter(base, over string, width, height int) string {
	baseLines := strings.Split(base, "\n")
	for len(baseLines) < height {
		baseLines = append(baseLines, "")
	}
	overLines := strings.Split(over, "\n")
	start := (height - len(overLines)) / 2
	if start < 0 {
		start = 0
	}
	pad := func(n int) string {
		if n < 0 {
			n = 0
		}
		return lipgloss.NewStyle().Background(cBg).Render(strings.Repeat(" ", n))
	}
	for i, ol := range overLines {
		row := start + i
		if row < 0 || row >= height {
			continue
		}
		ow := vwidth(ol)
		left := (width - ow) / 2
		baseLines[row] = pad(left) + ol + pad(width-left-ow)
	}
	return strings.Join(baseLines[:height], "\n")
}
