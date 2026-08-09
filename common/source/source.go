// Package source fetches upstream playlists and caches them on disk so the TUI
// starts fast and works offline once primed. It is infrastructure: it depends on
// the network and filesystem, and returns raw bytes for m3u.Parse to interpret.
package source

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Source is a named upstream playlist.
type Source struct {
	Name string
	URL  string
}

// The upstream projects are the single source of truth. Playlists are always
// rebuilt from these; nothing is hand-maintained locally.
//
// Primary playlist sources (fetched by Defaults):
//   - iptv-org/iptv      https://github.com/iptv-org/iptv      (index.country.m3u)
//   - Free-TV/IPTV       https://github.com/Free-TV/IPTV       (playlist.m3u8)
//
// Wider iptv-org ecosystem (reference; not yet fetched — see docs/sources.md):
//   - iptv-org/database  https://github.com/iptv-org/database   canonical channel data
//   - iptv-org/api       https://github.com/iptv-org/api        JSON API over the database
//   - iptv-org/epg       https://github.com/iptv-org/epg        program guides (XMLTV)
//   - iptv-org/sdk       https://github.com/iptv-org/sdk        typed client for the API
//   - iptv-org/awesome-iptv, iptv-org/community                 curated resources
const (
	iptvOrgCountryM3U = "https://iptv-org.github.io/iptv/index.country.m3u"
	freeTVPlaylist    = "https://raw.githubusercontent.com/Free-TV/IPTV/master/playlist.m3u8"
)

// Defaults are the two upstream playlists we ship with: iptv-org (exhaustive,
// grouped by country) and Free-TV (curated HD).
func Defaults() []Source {
	return []Source{
		{Name: "iptv-org", URL: iptvOrgCountryM3U},
		{Name: "free-tv", URL: freeTVPlaylist},
	}
}

// Loader fetches sources with an on-disk cache.
type Loader struct {
	CacheDir string
	MaxAge   time.Duration
	Client   *http.Client
}

// NewLoader returns a Loader caching under dir with a 12h freshness window.
func NewLoader(dir string) *Loader {
	return &Loader{
		CacheDir: dir,
		MaxAge:   12 * time.Hour,
		Client:   &http.Client{Timeout: 60 * time.Second},
	}
}

// Load returns the playlist bytes for s, serving a fresh cache file when one
// exists and falling back to a stale cache file if the network fails.
func (l *Loader) Load(s Source, forceRefresh bool) ([]byte, error) {
	path := l.cachePath(s)

	if !forceRefresh {
		if b, ok := l.freshCache(path); ok {
			return b, nil
		}
	}

	b, err := l.download(s.URL)
	if err != nil {
		if stale, ok := l.anyCache(path); ok {
			return stale, nil // degraded but usable
		}
		return nil, err
	}

	if err := os.MkdirAll(l.CacheDir, 0o755); err == nil {
		_ = os.WriteFile(path, b, 0o644)
	}
	return b, nil
}

func (l *Loader) download(url string) ([]byte, error) {
	resp, err := l.Client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch %s: status %d", url, resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

func (l *Loader) cachePath(s Source) string {
	sum := sha1.Sum([]byte(s.URL))
	return filepath.Join(l.CacheDir, fmt.Sprintf("%s-%s.m3u", s.Name, hex.EncodeToString(sum[:6])))
}

func (l *Loader) freshCache(path string) ([]byte, bool) {
	fi, err := os.Stat(path)
	if err != nil || time.Since(fi.ModTime()) > l.MaxAge {
		return nil, false
	}
	return l.anyCache(path)
}

func (l *Loader) anyCache(path string) ([]byte, bool) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}
	return b, true
}
