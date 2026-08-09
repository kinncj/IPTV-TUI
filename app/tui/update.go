package tui

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kinncj/IPTV-TUI/common/catalog"
	"github.com/kinncj/IPTV-TUI/common/player"
)

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Help overlay swallows everything except its own dismissal.
	if m.showHelp {
		switch msg.String() {
		case "?", "esc", "q", "enter":
			m.showHelp = false
		}
		return m, nil
	}

	// The source manager captures its own keys.
	if m.state == viewSources {
		return m.handleSourcesKey(msg)
	}

	// While filtering, let the list consume every key.
	if m.focusedFiltering() {
		return m.delegateToList(msg)
	}

	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit

	case "?":
		m.showHelp = true
		return m, nil

	case "p":
		m.cyclePlayer()
		return m, m.animate() // drive the toast timer

	case "r":
		if m.state == viewChannels {
			return m, m.reprobeCurrent()
		}
		return m, nil

	case "R":
		if cmd := m.reloadCmd(); cmd != nil {
			m.toast = "reloading from source…"
			m.toastTTL = 40
			return m, tea.Batch(cmd, m.animate())
		}
		return m, nil

	case "f":
		return m.toggleFavorite()

	case "t":
		return m.playHere()

	case "s":
		if m.state == viewGroups {
			return m.openSources()
		}

	case "g":
		if m.state == viewGroups {
			return m.toggleGrouping()
		}

	case "enter":
		return m.handleEnter()

	case "esc", "left", "h":
		switch m.state {
		case viewChannels:
			m.state = viewGroups
		case viewGroups:
			m.state = viewHome
			return m, m.animate() // resume logo animation
		}
		return m, nil
	}

	if m.state == viewHome {
		// Any other key leaves the splash.
		m.state = viewGroups
		return m, nil
	}
	return m.delegateToList(msg)
}

func (m Model) handleEnter() (tea.Model, tea.Cmd) {
	switch m.state {
	case viewHome:
		m.state = viewGroups
		return m, nil

	case viewGroups:
		sel, ok := m.groups.SelectedItem().(groupItem)
		if !ok {
			return m, nil
		}
		m.curGroup = sel.group.Name
		m.guide = nil // guide is per-group
		items := make([]list.Item, len(sel.group.Channels))
		for i, ch := range sel.group.Channels {
			items[i] = channelItem{ch: ch, store: m.store, st: m.st}
		}
		m.channels.SetItems(items)
		m.channels.ResetSelected()
		m.channels.ResetFilter()
		m.state = viewChannels
		return m, tea.Batch(m.probeGroup(sel.group), m.loadGuide(sel.group))

	case viewChannels:
		sel, ok := m.channels.SelectedItem().(channelItem)
		if !ok {
			return m, nil
		}
		m.recordRecent(sel.ch)
		return m, m.launch(sel)
	}
	return m, nil
}

// toggleFavorite adds/removes the selected channel from favorites and persists.
func (m Model) toggleFavorite() (tea.Model, tea.Cmd) {
	if m.state != viewChannels || m.st == nil {
		return m, nil
	}
	sel, ok := m.channels.SelectedItem().(channelItem)
	if !ok {
		return m, nil
	}
	nowFav := m.st.ToggleFavorite(entryOf(sel.ch))
	_ = m.st.Save()
	m.refreshGroups()
	if nowFav {
		m.toast = glStar() + " favorited"
	} else {
		m.toast = "removed from favorites"
	}
	m.toastTTL = 16
	return m, m.animate()
}

// playHere renders the selected stream inline in the terminal via mpv.
func (m Model) playHere() (tea.Model, tea.Cmd) {
	if m.state != viewChannels {
		return m, nil
	}
	sel, ok := m.channels.SelectedItem().(channelItem)
	if !ok {
		return m, nil
	}
	if m.termVO == "" {
		m.toast = "inline playback needs mpv"
		m.toastTTL = 18
		return m, m.animate()
	}
	m.recordRecent(sel.ch)
	name := sel.ch.Name
	cmd := player.TerminalCmd(sel.ch.URL, m.termVO, name)
	return m, tea.ExecProcess(cmd, func(err error) tea.Msg {
		return termDoneMsg{channel: name, err: err}
	})
}

func (m *Model) recordRecent(ch catalog.Channel) {
	if m.st == nil {
		return
	}
	m.st.PushRecent(entryOf(ch))
	_ = m.st.Save()
}

func (m Model) launch(item channelItem) tea.Cmd {
	name := item.ch.Name
	if len(m.players) == 0 {
		return func() tea.Msg { return launchResultMsg{channel: name, err: errNoPlayer} }
	}
	p := m.players[m.playerI]
	url := item.ch.URL
	return func() tea.Msg {
		return launchResultMsg{channel: name, err: p.Play(url)}
	}
}

func (m *Model) cyclePlayer() {
	if len(m.players) == 0 {
		m.toast = "no player installed"
		m.toastTTL = 20
		return
	}
	m.playerI = (m.playerI + 1) % len(m.players)
	name := m.players[m.playerI].Name()
	m.status = "player: " + name
	m.toast = glPlay() + " player " + glArrow() + " " + name
	m.toastTTL = 18 // ~1.6s at 90ms/tick
}

func (m Model) reprobeCurrent() tea.Cmd {
	for _, g := range m.cat.Groups {
		if g.Name == m.curGroup {
			for _, ch := range g.Channels {
				m.store.set(ch.URL, 0)
			}
			return m.probeGroup(g)
		}
	}
	return nil
}

// toggleGrouping switches the country/category grouping of the group list.
func (m Model) toggleGrouping() (tea.Model, tea.Cmd) {
	m.byCategoryOn = !m.byCategoryOn
	if m.byCategoryOn {
		m.cat = m.byCategory
		m.toast = "grouped by category"
	} else {
		m.cat = m.byCountry
		m.toast = "grouped by country"
	}
	m.refreshGroups()
	m.groups.ResetSelected()
	m.toastTTL = 16
	return m, m.animate()
}

func (m Model) focusedFiltering() bool {
	switch m.state {
	case viewGroups:
		return m.groups.FilterState() == list.Filtering
	case viewChannels:
		return m.channels.FilterState() == list.Filtering
	}
	return false
}
