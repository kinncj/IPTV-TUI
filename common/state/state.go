// Package state persists per-user runtime state — favorites and recently played
// channels — as JSON under the cache dir. It is best-effort: a missing or
// unreadable file yields empty state, and save errors are the caller's to log.
package state

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

const maxRecent = 30

// Entry identifies a channel well enough to re-list and re-play it.
type Entry struct {
	Name   string `json:"name"`
	URL    string `json:"url"`
	Group  string `json:"group"`
	Source string `json:"source,omitempty"`
}

// State is the persisted user state.
type State struct {
	mu        sync.Mutex
	path      string
	Favorites []Entry `json:"favorites"`
	Recent    []Entry `json:"recent"`
}

// Load reads state from path (absent file → empty state).
func Load(path string) *State {
	s := &State{path: path}
	if data, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(data, s)
	}
	return s
}

// Save writes state to its path, creating parent dirs.
func (s *State) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.path == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o644)
}

// IsFavorite reports whether url is favorited.
func (s *State) IsFavorite(url string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return indexByURL(s.Favorites, url) >= 0
}

// ToggleFavorite adds or removes e from favorites, returning the new state.
func (s *State) ToggleFavorite(e Entry) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if i := indexByURL(s.Favorites, e.URL); i >= 0 {
		s.Favorites = append(s.Favorites[:i], s.Favorites[i+1:]...)
		return false
	}
	s.Favorites = append(s.Favorites, e)
	return true
}

// PushRecent records e as most-recently played (dedup by URL, capped).
func (s *State) PushRecent(e Entry) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if i := indexByURL(s.Recent, e.URL); i >= 0 {
		s.Recent = append(s.Recent[:i], s.Recent[i+1:]...)
	}
	s.Recent = append([]Entry{e}, s.Recent...)
	if len(s.Recent) > maxRecent {
		s.Recent = s.Recent[:maxRecent]
	}
}

// Snapshot returns copies of favorites and recent for read-only UI use.
func (s *State) Snapshot() (favorites, recent []Entry) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]Entry(nil), s.Favorites...), append([]Entry(nil), s.Recent...)
}

func indexByURL(es []Entry, url string) int {
	for i, e := range es {
		if e.URL == url {
			return i
		}
	}
	return -1
}
