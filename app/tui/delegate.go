package tui

import (
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// rowDelegate renders a two-line item (title + subtitle) with a full-width
// highlight and a left accent bar on the selected row — the MovieBox list look.
type rowDelegate struct{}

func (rowDelegate) Height() int                         { return 2 }
func (rowDelegate) Spacing() int                        { return 1 }
func (rowDelegate) Update(tea.Msg, *list.Model) tea.Cmd { return nil }

func (rowDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	width := m.Width()
	if width <= 0 {
		return
	}
	selected := index == m.Index()

	var icon, title, sub string
	iconColor := cMuted
	fav := false
	switch v := item.(type) {
	case groupItem:
		icon = glBullet()
		iconColor = cPurple
		if v.icon != "" {
			icon = v.icon
		}
		if v.iconColor != "" {
			iconColor = v.iconColor
		}
		title = v.group.Name
		sub = pluralChannels(len(v.group.Channels))
	case channelItem:
		st := v.store.get(v.ch.URL)
		icon = statusIcon(st)
		iconColor = statusColor(st)
		title = v.ch.Name
		sub = v.ch.Source
		fav = v.favorite()
	default:
		return
	}

	bg := cBg
	titleColor := cFg
	if selected {
		bg = cBgElev
		titleColor = cPurple
	}

	bar := "  "
	if selected {
		bar = lipgloss.NewStyle().Foreground(cPurple).Background(bg).Render(glBar() + " ")
	}

	iconR := lipgloss.NewStyle().Foreground(iconColor).Background(bg).Render(icon)
	titleR := lipgloss.NewStyle().Foreground(titleColor).Background(bg).Bold(selected).Render(title)
	subR := lipgloss.NewStyle().Foreground(cMuted).Background(bg).Render(sub)

	star := ""
	if fav {
		star = lipgloss.NewStyle().Foreground(cYellow).Background(bg).Render(" " + glStar())
	}

	line := lipgloss.NewStyle().Width(width).Background(bg)
	top := line.Render(bar + iconR + " " + titleR + star)
	bottom := line.Render(strings.Repeat(" ", 4) + subR)

	io.WriteString(w, top+"\n"+bottom)
}
