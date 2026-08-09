// Package config loads optional, gitignored user configuration: extra playlist
// sources and theme/player preferences. The upstream repos remain the source of
// truth; this only augments them. Config is never required — the app works with
// zero config.
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Source is a user-added playlist (same shape as source.Source, kept separate to
// avoid a dependency cycle).
type Source struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// Config is the merged user configuration.
type Config struct {
	Theme   string   `json:"theme,omitempty"`
	Player  string   `json:"player,omitempty"`
	API     bool     `json:"api,omitempty"`     // ingest built-in iptv-org from its JSON API
	EPGURL  string   `json:"epg_url,omitempty"` // custom XMLTV guide URL
	Sources []Source `json:"sources,omitempty"`
}

// Dir is the per-user directory for user-changeable files, in the OS config
// location (~/.config/iptv-tui on Linux, ~/Library/Application Support/iptv-tui
// on macOS). Everything a user edits or the app persists on their behalf lives
// here, so a downloaded binary works without a repo checkout.
func Dir() string {
	if d, err := os.UserConfigDir(); err == nil {
		return filepath.Join(d, "iptv-tui")
	}
	return "."
}

// UserPath is the config file the app reads and writes: <Dir>/config.json.
func UserPath() string { return filepath.Join(Dir(), "config.json") }

// StatePath is where favorites and last-played persist: <Dir>/state.json.
func StatePath() string { return filepath.Join(Dir(), "state.json") }

// DefaultPaths are searched and merged in order (later files win for scalars,
// sources are appended):
//
//	<OS config dir>/iptv-tui/config.json   (per-user, where the app writes)
//	./iptv.local.json                      (optional, gitignored, dev override)
func DefaultPaths() []string {
	paths := []string{UserPath()}
	if lp := LocalPath(); lp != paths[0] {
		paths = append(paths, lp)
	}
	return paths
}

// Load merges the given config files. Missing files are skipped silently; a
// malformed file returns an error naming it. Returned slice lists files applied.
func Load(paths ...string) (Config, []string, error) {
	var cfg Config
	var loaded []string

	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			continue // absent config is normal
		}
		var c Config
		if err := json.Unmarshal(data, &c); err != nil {
			return cfg, loaded, &LoadError{Path: p, Err: err}
		}
		cfg.merge(c)
		loaded = append(loaded, p)
	}
	return cfg, loaded, nil
}

// LocalPath is the repo-local, gitignored config the TUI reads and writes.
func LocalPath() string { return "iptv.local.json" }

// LoadFile reads a single config file (absent → empty config, no error).
func LoadFile(path string) (Config, error) {
	var c Config
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return c, nil
		}
		return c, err
	}
	if err := json.Unmarshal(data, &c); err != nil {
		return c, &LoadError{Path: path, Err: err}
	}
	return c, nil
}

// Save writes cfg to path as indented JSON, creating parent dirs.
func Save(path string, cfg Config) error {
	if dir := filepath.Dir(path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

// AddSource appends a source to the local config file (dedup by URL) and saves.
func AddSource(path string, s Source) error {
	cfg, err := LoadFile(path)
	if err != nil {
		return err
	}
	for _, x := range cfg.Sources {
		if x.URL == s.URL {
			return nil // already present
		}
	}
	cfg.Sources = append(cfg.Sources, s)
	return Save(path, cfg)
}

// RemoveSource deletes the source with the given URL from the local config file.
func RemoveSource(path, url string) error {
	cfg, err := LoadFile(path)
	if err != nil {
		return err
	}
	kept := cfg.Sources[:0]
	for _, x := range cfg.Sources {
		if x.URL != url {
			kept = append(kept, x)
		}
	}
	cfg.Sources = kept
	return Save(path, cfg)
}

func (c *Config) merge(o Config) {
	if o.Theme != "" {
		c.Theme = o.Theme
	}
	if o.Player != "" {
		c.Player = o.Player
	}
	if o.API {
		c.API = true
	}
	if o.EPGURL != "" {
		c.EPGURL = o.EPGURL
	}
	for _, s := range o.Sources {
		if s.URL != "" {
			c.Sources = append(c.Sources, s)
		}
	}
}

// LoadError names the file that failed to parse.
type LoadError struct {
	Path string
	Err  error
}

func (e *LoadError) Error() string { return "config " + e.Path + ": " + e.Err.Error() }
func (e *LoadError) Unwrap() error { return e.Err }
