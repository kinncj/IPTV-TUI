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
	API     bool     `json:"api,omitempty"` // ingest built-in iptv-org from its JSON API
	Sources []Source `json:"sources,omitempty"`
}

// DefaultPaths are searched and merged in order (later files win for scalars,
// sources are appended):
//
//	$XDG_CONFIG_HOME/iptv/config.json   (per-user)
//	./iptv.local.json                   (repo-local, gitignored)
func DefaultPaths() []string {
	var paths []string
	if dir, err := os.UserConfigDir(); err == nil {
		paths = append(paths, filepath.Join(dir, "iptv", "config.json"))
	}
	paths = append(paths, "iptv.local.json")
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
