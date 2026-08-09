package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kinncj/iptv/common/catalog"
	"github.com/kinncj/iptv/common/player"
	"github.com/kinncj/iptv/common/state"
)

type fakePlayer struct{ name string }

func (f fakePlayer) Name() string    { return f.name }
func (fakePlayer) Available() bool   { return true }
func (fakePlayer) Play(string) error { return nil }

func fakePlayers(names ...string) []player.Player {
	var ps []player.Player
	for _, n := range names {
		ps = append(ps, fakePlayer{name: n})
	}
	return ps
}

func testModel() Model {
	ApplyTheme("tokyonight", true, true)
	cat := catalog.Build([]catalog.Channel{
		{Name: "Globo", URL: "http://a/g.m3u8", Group: "Brazil", Source: "iptv-org"},
		{Name: "SBT", URL: "http://a/s.m3u8", Group: "Brazil", Source: "free-tv"},
		{Name: "Telefe", URL: "http://a/t.m3u8", Group: "Argentina", Source: "iptv-org"},
	})
	return New(cat, nil)
}

func send(m Model, msg tea.Msg) Model {
	next, _ := m.Update(msg)
	return next.(Model)
}

func key(s string) tea.KeyMsg {
	switch s {
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
	}
}

// TestViewNeverPanics renders every screen at a realistic size. This is the
// guard that catches rendering bugs like the logo rune/byte overflow.
func TestViewNeverPanics(t *testing.T) {
	m := testModel()
	m = send(m, tea.WindowSizeMsg{Width: 120, Height: 40})

	checks := []struct {
		name string
		do   func(Model) Model
	}{
		{"home", func(m Model) Model { return m }},
		{"groups", func(m Model) Model { return send(m, key("enter")) }},
		{"channels", func(m Model) Model { return send(send(m, key("enter")), key("enter")) }},
		{"help", func(m Model) Model { return send(m, key("?")) }},
	}
	for _, c := range checks {
		got := c.do(m).View()
		if strings.TrimSpace(got) == "" {
			t.Errorf("%s: View() rendered empty", c.name)
		}
	}
}

// TestViewSurvivesTinyTerminal guards the dimension math against underflow.
func TestViewSurvivesTinyTerminal(t *testing.T) {
	m := testModel()
	for _, sz := range []tea.WindowSizeMsg{{Width: 1, Height: 1}, {Width: 20, Height: 6}, {Width: 300, Height: 90}} {
		mm := send(m, sz)
		_ = mm.View()                     // must not panic
		_ = send(mm, key("enter")).View() // groups
	}
}

// TestFavoritesAndSourcesRender exercises the v1.1 views and state paths.
func TestFavoritesAndSourcesRender(t *testing.T) {
	ApplyTheme("tokyonight", true, true)
	cat := catalog.Build([]catalog.Channel{
		{Name: "Globo", URL: "http://a/g.m3u8", Group: "Brazil", Source: "iptv-org"},
	})
	st := state.Load("")
	st.ToggleFavorite(state.Entry{Name: "Globo", URL: "http://a/g.m3u8", Group: "Brazil"})

	m := New(cat, fakePlayers("mpv")).
		WithState(st).
		WithTerminalPlayback("tct").
		WithSources([]SourceInfo{{Name: "iptv-org", URL: "http://x", Builtin: true}})
	m = send(m, tea.WindowSizeMsg{Width: 120, Height: 40})

	// Favorites should be the first group now.
	if got := m.groups.VisibleItems()[0].(groupItem).group.Name; got != "Favorites" {
		t.Errorf("first group = %q, want Favorites", got)
	}

	// Open a country, then render channels (favorite star path) + toggle.
	m = send(send(m, key("enter")), key("enter"))
	if strings.TrimSpace(m.channelsView()) == "" {
		t.Error("channels view empty")
	}
	m = send(m, key("t")) // inline play path (mpv present)

	// Sources manager: open, enter add-mode, render.
	m = send(m, key("esc")) // back to groups
	m = send(m, key("s"))
	if m.state != viewSources {
		t.Fatalf("state = %v, want viewSources", m.state)
	}
	if strings.TrimSpace(m.sourcesView()) == "" {
		t.Error("sources view empty")
	}
	m = send(m, key("a"))
	if !m.srcAdding {
		t.Error("'a' should enter add mode")
	}
	_ = m.sourcesView() // must not panic in add mode
}

// TestBuiltinSourceNotRemovable guards the promise that built-in sources can't
// be deleted from the in-TUI manager.
func TestBuiltinSourceNotRemovable(t *testing.T) {
	ApplyTheme("tokyonight", true, true)
	cat := catalog.Build([]catalog.Channel{{Name: "X", URL: "u", Group: "G"}})
	m := New(cat, nil).WithSources([]SourceInfo{
		{Name: "iptv-org", URL: "http://builtin", Builtin: true},
	})
	m = send(m, tea.WindowSizeMsg{Width: 100, Height: 30})
	m = send(m, key("enter")) // groups
	m = send(m, key("s"))     // sources manager
	if m.state != viewSources {
		t.Fatalf("expected sources view")
	}
	// selection is the built-in; pressing 'd' must not remove it
	before := len(m.srcRows)
	m = send(m, key("d"))
	if len(m.srcRows) != before {
		t.Errorf("built-in source was removed (rows %d → %d)", before, len(m.srcRows))
	}
	if m.toast == "" {
		t.Errorf("expected a toast explaining built-ins can't be removed")
	}
}

// TestGroupingToggle covers the v1.3 country/category grouping switch.
func TestGroupingToggle(t *testing.T) {
	ApplyTheme("tokyonight", true, true)
	cat := catalog.Build([]catalog.Channel{
		{Name: "A", Group: "Brazil", Category: "news", URL: "u1"},
		{Name: "B", Group: "Japan", Category: "movies", URL: "u2"},
	})
	m := New(cat, nil)
	m = send(m, tea.WindowSizeMsg{Width: 110, Height: 30})
	m = send(m, key("enter")) // into groups view

	if m.byCategoryOn {
		t.Fatal("should start grouped by country")
	}
	m = send(m, key("g"))
	if !m.byCategoryOn {
		t.Fatal("'g' should switch to category grouping")
	}
	if !strings.Contains(m.groupsView(), "Categories") {
		t.Error("header should read Categories in category mode")
	}
	names := map[string]bool{}
	for _, it := range m.groups.VisibleItems() {
		names[it.(groupItem).group.Name] = true
	}
	if !names["news"] || !names["movies"] {
		t.Errorf("category groups missing: %v", names)
	}
	// Toggle back.
	m = send(m, key("g"))
	if m.byCategoryOn {
		t.Error("second 'g' should return to country grouping")
	}
}

func TestPlayerToastOnCycle(t *testing.T) {
	ApplyTheme("tokyonight", true, true)
	cat := catalog.Build([]catalog.Channel{{Name: "X", URL: "u", Group: "G"}})
	m := New(cat, fakePlayers("mpv", "vlc"))
	m = send(m, tea.WindowSizeMsg{Width: 100, Height: 30})

	m = send(m, key("p"))
	if !strings.Contains(m.toast, "vlc") {
		t.Errorf("toast should announce the new player, got %q", m.toast)
	}
	if m.toastTTL <= 0 {
		t.Errorf("toast should be live after switch")
	}
}
