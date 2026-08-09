package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kinncj/iptv/common/config"
)

// openSources builds the source list (built-ins + user config) and shows it.
func (m Model) openSources() (tea.Model, tea.Cmd) {
	m.refreshSrcRows()
	m.srcIdx = 0
	m.state = viewSources
	return m, nil
}

// refreshSrcRows reloads the row list from the built-ins plus the local config.
func (m *Model) refreshSrcRows() {
	rows := append([]SourceInfo(nil), m.builtins...)
	if cfg, err := config.LoadFile(m.localPath); err == nil {
		for _, s := range cfg.Sources {
			name := s.Name
			if name == "" {
				name = "custom"
			}
			rows = append(rows, SourceInfo{Name: name, URL: s.URL, Builtin: false})
		}
	}
	m.srcRows = rows
	if m.srcIdx >= len(rows) {
		m.srcIdx = len(rows) - 1
	}
	if m.srcIdx < 0 {
		m.srcIdx = 0
	}
}

func (m Model) handleSourcesKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.srcAdding {
		switch msg.Type {
		case tea.KeyEnter:
			url := strings.TrimSpace(m.srcInput.Value())
			m.srcAdding = false
			m.srcInput.Blur()
			m.srcInput.SetValue("")
			if url == "" {
				return m, nil
			}
			if err := config.AddSource(m.localPath, config.Source{Name: "custom", URL: url}); err != nil {
				m.toast = "add failed"
				m.toastTTL = 20
				return m, m.animate()
			}
			m.refreshSrcRows()
			m.toast = "source added — reloading"
			m.toastTTL = 40
			return m, tea.Batch(m.reloadCmd(), m.animate())
		case tea.KeyEsc:
			m.srcAdding = false
			m.srcInput.Blur()
			m.srcInput.SetValue("")
			return m, nil
		}
		var cmd tea.Cmd
		m.srcInput, cmd = m.srcInput.Update(msg)
		return m, cmd
	}

	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc", "s", "q":
		m.state = viewGroups
		return m, nil
	case "up", "k":
		if m.srcIdx > 0 {
			m.srcIdx--
		}
		return m, nil
	case "down", "j":
		if m.srcIdx < len(m.srcRows)-1 {
			m.srcIdx++
		}
		return m, nil
	case "a":
		m.srcAdding = true
		m.srcInput.Focus()
		return m, textinput.Blink
	case "d":
		if m.srcIdx < 0 || m.srcIdx >= len(m.srcRows) {
			return m, nil
		}
		row := m.srcRows[m.srcIdx]
		if row.Builtin {
			m.toast = "can't remove a built-in source"
			m.toastTTL = 18
			return m, m.animate()
		}
		if err := config.RemoveSource(m.localPath, row.URL); err != nil {
			m.toast = "remove failed"
			m.toastTTL = 20
			return m, m.animate()
		}
		m.refreshSrcRows()
		m.toast = "source removed — reloading"
		m.toastTTL = 40
		return m, tea.Batch(m.reloadCmd(), m.animate())
	}
	return m, nil
}

func (m Model) sourcesView() string {
	hdr := header("Sources", itoa(len(m.srcRows))+" configured", m.width)

	var rows []string
	for i, s := range m.srcRows {
		selected := i == m.srcIdx
		bg := cBg
		if selected {
			bg = cBgElev
		}
		tag := "user "
		tagColor := cCyan
		if s.Builtin {
			tag = "built"
			tagColor = cMuted
		}
		bar := "  "
		if selected {
			bar = lipgloss.NewStyle().Foreground(cPurple).Background(bg).Render(glBar() + " ")
		}
		badge := lipgloss.NewStyle().Foreground(tagColor).Background(bg).Render("[" + tag + "]")
		name := lipgloss.NewStyle().Foreground(cFg).Background(bg).Bold(selected).
			Render(pad(s.Name, 14))
		url := lipgloss.NewStyle().Foreground(cMuted).Background(bg).Render(truncate(s.URL, maxInt(10, m.width-30)))
		line := lipgloss.NewStyle().Width(m.width).Background(bg).Render(bar + badge + " " + name + " " + url)
		rows = append(rows, line)
	}

	body := lipgloss.NewStyle().Height(m.contentHeight()).MaxHeight(m.contentHeight()).
		Background(cBg).Render(lipgloss.JoinVertical(lipgloss.Left, rows...))

	if m.srcAdding {
		prompt := lipgloss.NewStyle().Foreground(cPurple).Background(cBg).Bold(true).Render("add source  ") +
			lipgloss.NewStyle().Background(cBg).Render(m.srcInput.View())
		body = lipgloss.JoinVertical(lipgloss.Left, prompt, "", body)
		body = lipgloss.NewStyle().Height(m.contentHeight()).MaxHeight(m.contentHeight()).
			Background(cBg).Render(body)
	}

	foot := m.foot([][2]string{
		{"a", "add"}, {"d", "remove"}, {"esc", "back"}, {"q", "quit"},
	})
	return assemble(m.width, hdr, body, foot)
}

func pad(s string, n int) string {
	r := []rune(s)
	if len(r) >= n {
		return truncate(s, n)
	}
	return s + strings.Repeat(" ", n-len(r))
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
