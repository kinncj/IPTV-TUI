package tui

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
	"github.com/kinncj/IPTV-TUI/common/probe"
)

var errNoPlayer = errors.New("no media player installed")

// --- layout dimensions ------------------------------------------------------

func (m Model) contentHeight() int {
	// top margin (1) + header (2) + blank (1) + footer (1) = 5 lines of chrome.
	h := m.height - 5
	if h < 3 {
		h = 3
	}
	return h
}

func (m Model) listDims() (int, int) {
	w := m.width - 2 // leave a column for the scrollbar + gap
	if w < 10 {
		w = 10
	}
	return w, m.contentHeight()
}

func (m Model) leftPaneWidth() int {
	w := m.width * 3 / 5
	if w < 24 {
		w = m.width
	}
	return w
}

func (m Model) channelListWidth() int {
	w := m.leftPaneWidth() - 2
	if w < 10 {
		w = 10
	}
	return w
}

// --- top-level view ---------------------------------------------------------

func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}
	var content string
	switch {
	case m.vid != nil:
		content = m.videoView()
	case m.picking:
		content = m.chooserView()
	case m.showHelp:
		content = m.helpView()
	case m.state == viewHome:
		content = m.homeView()
	case m.state == viewSources:
		content = m.sourcesView()
	case m.state == viewChannels:
		content = m.channelsView()
	default:
		content = m.groupsView()
	}
	return canvasStyle.Width(m.width).Height(m.height).Render(content)
}

// --- home splash ------------------------------------------------------------

func (m Model) homeView() string {
	// Dim, letter-spaced tagline with a blinking cursor — statusline style.
	cursor := " "
	if (m.frame/7)%2 == 0 {
		cursor = "_"
	}
	tagline := tagStyle.Render(letterSpace("your terminal streamer")) +
		lipgloss.NewStyle().Foreground(cCyan).Background(cBg).Bold(true).Render(cursor)

	hint := mutedStyle.Render("press ") +
		lipgloss.NewStyle().Foreground(cPurple).Background(cBg).Render("enter") +
		mutedStyle.Render(" to browse "+itoa(len(m.cat.Groups))+" countries")

	block := lipgloss.JoinVertical(lipgloss.Center,
		logo(m.frame),
		"",
		"",
		tagline,
		"",
		hint,
	)

	centered := lipgloss.Place(m.width, m.height-1,
		lipgloss.Center, lipgloss.Center, block,
		lipgloss.WithWhitespaceBackground(cBg))

	foot := m.foot([][2]string{{"enter", "browse"}, {"?", "help"}, {"q", "quit"}})
	return centered + "\n" + foot
}

// foot builds the footer for the current player/toast state.
func (m Model) foot(keys [][2]string) string {
	return footer(m.width, keys, m.playerName(), m.toast, m.toast != "")
}

// --- country list -----------------------------------------------------------

func (m Model) groupsView() string {
	title := "Countries"
	if m.byCategoryOn {
		title = "Categories"
	}
	hdr := header(title, m.rightLabel(&m.groups, m.groupCountLabel()), m.width)
	body := m.renderList(&m.groups, m.groups.Width())
	foot := m.foot([][2]string{
		{"enter", "open"}, {"g", "group"}, {"/", "filter"}, {"s", "sources"}, {"R", "reload"}, {"?", "help"}, {"q", "quit"},
	})
	return assemble(m.width, hdr, body, foot)
}

// --- channel list + detail --------------------------------------------------

func (m Model) channelsView() string {
	count := pluralChannels(len(m.channels.VisibleItems()))
	hdr := header(m.curGroup, m.rightLabel(&m.channels, count), m.width)

	left := m.renderList(&m.channels, m.leftPaneWidth())
	right := m.detailPane()
	row := lipgloss.JoinHorizontal(lipgloss.Top, left, canvasStyle.Render(" "), right)

	keys := [][2]string{{"enter", "window"}}
	if m.termVO != "" {
		keys = append(keys, [2]string{"t", "here"})
	}
	keys = append(keys,
		[2]string{"f", "fav"}, [2]string{"esc", "back"}, [2]string{"r", "re-probe"},
		[2]string{"p", "player"}, [2]string{"/", "filter"}, [2]string{"?", "help"}, [2]string{"q", "quit"})
	foot := m.foot(keys)
	return assemble(m.width, hdr, row, foot)
}

func (m Model) detailPane() string {
	w := m.width - m.leftPaneWidth() - 1
	h := m.contentHeight()
	sel, ok := m.channels.SelectedItem().(channelItem)
	if !ok {
		empty := lipgloss.NewStyle().Foreground(cMuted).Background(cBgElev).Render("no channel selected")
		return cardStyle.Width(w - 2).Height(h - 2).Render(empty)
	}

	info := detailInfo{favorite: sel.favorite()}
	if m.guide != nil && sel.ch.TvgID != "" {
		now, next := m.guide.NowNext(sel.ch.TvgID, time.Now())
		if now != nil {
			info.epgNow = now.Title + "  (until " + now.Stop.Local().Format("15:04") + ")"
		}
		if next != nil {
			info.epgNext = next.Title + "  (" + next.Start.Local().Format("15:04") + ")"
		}
	}
	return detailCard(sel.ch, m.store.get(sel.ch.URL), info, w, h)
}

// --- help overlay -----------------------------------------------------------

func (m Model) helpView() string {
	section := func(s string) string { return sectionStyle.Render(s) }
	row := func(k, d string) string {
		key := lipgloss.NewStyle().Foreground(cOrange).Background(cBg).Width(12).Render(k)
		return key + mutedStyle.Render(d)
	}

	body := lipgloss.JoinVertical(lipgloss.Left,
		section("Navigation"),
		row("↑ / ↓", "move selection"),
		row("enter", "open country / play channel"),
		row("g", "group by country / category"),
		row("s", "manage sources"),
		row("esc", "go back"),
		row("/", "filter the current list"),
		"",
		section("Playback"),
		row("enter", "play in a window ("+m.playerName()+")"),
		row("t", "play in the terminal (choose mpv or built-in)"),
		row("p", "cycle player (mpv "+glArrow()+" vlc "+glArrow()+" ffplay)"),
		row("f", "toggle favorite"),
		row("r", "re-probe the current country"),
		row("R", "reload playlists from source"),
		"",
		section("Legend"),
		row(okGlyph(), "reachable"),
		row(warnGlyph(), "blocked (often geo-blocked)"),
		row(deadGlyph(), "dead (timeout / error)"),
		"",
		section("System"),
		row("?", "toggle this help"),
		row("q", "quit"),
	)

	title := lipgloss.NewStyle().Foreground(cPurple).Background(cBg).Bold(true).Render("Keybindings")
	panel := overlayStyle.Render(lipgloss.JoinVertical(lipgloss.Left, title, "", body))

	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center, panel,
		lipgloss.WithWhitespaceBackground(cBg))
}

// --- helpers ----------------------------------------------------------------

// renderList draws a list's bare rows inside a fixed box and joins a scrollbar.
func (m Model) renderList(l *list.Model, paneWidth int) string {
	h := m.contentHeight()
	view := lipgloss.NewStyle().Width(l.Width()).Height(h).MaxHeight(h).
		Background(cBg).Render(l.View())
	sb := scrollbar(h, len(l.VisibleItems()), l.Index())
	return lipgloss.JoinHorizontal(lipgloss.Top, view, canvasStyle.Render(" "), sb)
}

// rightLabel shows the live filter query while filtering, else the given label.
func (m Model) rightLabel(l *list.Model, fallback string) string {
	switch {
	case l.FilterState() == list.Filtering:
		return "/" + l.FilterInput.Value() + glCaret()
	case l.IsFiltered():
		return fmt.Sprintf("%d matches", len(l.VisibleItems()))
	default:
		return fallback
	}
}

// assemble stacks the standard chrome: top margin, header, blank, body, footer.
func assemble(width int, hdr, body, foot string) string {
	blank := canvasStyle.Width(width).Render("")
	return lipgloss.JoinVertical(lipgloss.Left, blank, hdr, blank, body, foot)
}

// letterSpace inserts a space between characters for an airy tagline.
func letterSpace(s string) string {
	r := []rune(s)
	var b strings.Builder
	for i, c := range r {
		if i > 0 {
			b.WriteString(" ")
		}
		b.WriteRune(c)
	}
	return b.String()
}

func termNote(vo string) string {
	if vo == "" {
		return " (needs mpv)"
	}
	return " (" + vo + ")"
}

func okGlyph() string {
	return lipgloss.NewStyle().Foreground(cGreen).Background(cBg).Render(statusIcon(probe.OK))
}
func warnGlyph() string {
	return lipgloss.NewStyle().Foreground(cYellow).Background(cBg).Render(statusIcon(probe.Blocked))
}
func deadGlyph() string {
	return lipgloss.NewStyle().Foreground(cRed).Background(cBg).Render(statusIcon(probe.Dead))
}
