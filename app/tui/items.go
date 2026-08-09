package tui

import (
	"sync"

	"github.com/charmbracelet/lipgloss"
	"github.com/kinncj/IPTV-TUI/common/catalog"
	"github.com/kinncj/IPTV-TUI/common/probe"
	"github.com/kinncj/IPTV-TUI/common/state"
)

// statusStore holds live probe results keyed by stream URL. List items read from
// it on every render, so probe updates surface without rebuilding the list.
type statusStore struct {
	mu sync.RWMutex
	m  map[string]probe.Status
}

func newStatusStore() *statusStore {
	return &statusStore{m: make(map[string]probe.Status)}
}

func (s *statusStore) get(url string) probe.Status {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.m[url]
}

func (s *statusStore) set(url string, st probe.Status) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[url] = st
}

// groupItem is a country/category row (or a synthetic Favorites/Recent group).
type groupItem struct {
	group     catalog.Group
	icon      string         // overrides the default bullet (synthetic groups)
	iconColor lipgloss.Color // zero value → default accent
}

func (g groupItem) Title() string       { return g.group.Name }
func (g groupItem) Description() string { return pluralChannels(len(g.group.Channels)) }
func (g groupItem) FilterValue() string { return g.group.Name }

// channelItem is a single stream row; its title carries the live probe icon and
// a favorite marker.
type channelItem struct {
	ch    catalog.Channel
	store *statusStore
	st    *state.State
}

func (c channelItem) Title() string       { return c.ch.Name }
func (c channelItem) Description() string { return c.ch.Source }
func (c channelItem) FilterValue() string { return c.ch.Name }

func (c channelItem) favorite() bool {
	return c.st != nil && c.st.IsFavorite(c.ch.URL)
}

func entryOf(ch catalog.Channel) state.Entry {
	return state.Entry{Name: ch.Name, URL: ch.URL, Group: ch.Group, Source: ch.Source}
}

func channelOf(e state.Entry) catalog.Channel {
	return catalog.Channel{Name: e.Name, URL: e.URL, Group: e.Group, Source: e.Source}
}

// entriesGroup builds a synthetic group from persisted entries.
func entriesGroup(name string, entries []state.Entry) catalog.Group {
	chs := make([]catalog.Channel, len(entries))
	for i, e := range entries {
		chs[i] = channelOf(e)
	}
	return catalog.Group{Name: name, Channels: chs}
}

func pluralChannels(n int) string {
	if n == 1 {
		return "1 channel"
	}
	return itoa(n) + " channels"
}
