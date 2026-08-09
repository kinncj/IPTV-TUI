// Command iptv is a terminal UI for browsing IPTV playlists (iptv-org, Free-TV)
// by country and launching a media player, with live per-stream reachability
// probing so dead/geo-blocked channels are visible before you try them.
//
// This file is the composition root: it wires infrastructure (source loader,
// player detection) to the domain (catalog) and the presentation layer (tui).
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kinncj/iptv/app/tui"
	"github.com/kinncj/iptv/common/catalog"
	"github.com/kinncj/iptv/common/config"
	"github.com/kinncj/iptv/common/export"
	"github.com/kinncj/iptv/common/iptvorg"
	"github.com/kinncj/iptv/common/m3u"
	"github.com/kinncj/iptv/common/player"
	"github.com/kinncj/iptv/common/source"
	"github.com/kinncj/iptv/common/state"
	"github.com/kinncj/iptv/common/termcaps"
)

// version is set at build time via -ldflags "-X main.version=…".
var version = "dev"

func main() {
	showVersion := flag.Bool("version", false, "print version and exit")
	refresh := flag.Bool("refresh", false, "force re-download of playlists, ignoring cache")
	cacheDir := flag.String("cache", defaultCacheDir(), "playlist cache directory")
	preferred := flag.String("player", "", "preferred player: mpv|vlc|ffplay (default: auto)")
	exportDir := flag.String("export", "", "rebuild playlists from source into DIR (all.m3u + countries/) and exit")
	theme := flag.String("theme", "", "color theme: auto|"+strings.Join(tui.ThemeNames(), "|"))
	listThemes := flag.Bool("themes", false, "list available themes and exit")
	apiFlag := flag.Bool("api", false, "ingest the built-in iptv-org source from its JSON API (richer metadata) instead of M3U")
	probeConc := flag.Int("probe-concurrency", 12, "max concurrent reachability probes")
	probeTimeout := flag.Duration("probe-timeout", 6*time.Second, "per-probe timeout")
	probeAll := flag.Bool("probe-all", false, "probe every channel in the background at startup")
	flag.Parse()

	if *showVersion {
		fmt.Println("iptv-tui", version)
		return
	}

	if *listThemes {
		fmt.Println(strings.Join(tui.ThemeNames(), "\n"))
		return
	}

	// Optional, gitignored user config: extra sources + theme/player preferences.
	cfg, _, cfgErr := config.Load(config.DefaultPaths()...)
	if cfgErr != nil {
		fmt.Fprintln(os.Stderr, "iptv:", cfgErr) // bad config is worth surfacing
	}

	useAPI := *apiFlag || cfg.API

	// Upstream repos are the source of truth; user config only adds more. The API
	// path replaces ONLY the built-in iptv-org M3U — user and Free-TV lists are
	// untouched, so enabling it can never break a user's own lists.
	sources := mergeSources(source.Defaults(), cfg.Sources)
	m3uSources := sources
	if useAPI {
		m3uSources = withoutSource(sources, source.Defaults()[0].URL) // drop iptv-org M3U
	}

	// Rebuilding local playlists always pulls fresh from the upstream repos.
	forceFresh := *refresh || *exportDir != ""
	cat, err := assembleCatalog(m3uSources, *cacheDir, forceFresh, useAPI)
	if err != nil {
		fmt.Fprintln(os.Stderr, "iptv:", err)
		os.Exit(1)
	}
	if cat.Total == 0 {
		fmt.Fprintln(os.Stderr, "iptv: no channels loaded (network down and no cache?)")
		os.Exit(1)
	}

	if *exportDir != "" {
		n, err := export.Write(cat, *exportDir)
		if err != nil {
			fmt.Fprintln(os.Stderr, "iptv: export:", err)
			os.Exit(1)
		}
		fmt.Printf("rebuilt %d playlists (%s channels, %d countries) in %s\n",
			n, comma(cat.Total), len(cat.Groups), *exportDir)
		return
	}

	// Precedence: flag > env > config > default.
	themeName := pick(*theme, os.Getenv("IPTV_THEME"), cfg.Theme, "auto")
	playerName := pick(*preferred, cfg.Player)
	players := orderPlayers(player.Detect(), playerName)

	// Detect terminal capabilities and apply the theme before first render, so
	// truecolor survives multiplexers and the palette follows the OS light/dark.
	caps := termcaps.Detect()
	lipgloss.SetColorProfile(caps.Profile) // keep truecolor under multiplexers
	tui.ApplyTheme(themeName, caps.DarkBG, caps.Unicode)

	// Inline terminal playback requires mpv; pick the best video output.
	termVO := ""
	if player.HasMPV() {
		termVO = caps.TerminalVO()
	}

	// Persisted favorites / last-played and a source reloader for the `R` key.
	// The reloader honors the API choice and re-reads user config each time, so
	// sources added in the TUI show up after a reload.
	st := state.Load(filepath.Join(*cacheDir, "state.json"))
	reload := func() (catalog.Catalog, error) {
		latest, _, _ := config.Load(config.DefaultPaths()...)
		all := mergeSources(source.Defaults(), latest.Sources)
		m3u := all
		if useAPI {
			m3u = withoutSource(all, source.Defaults()[0].URL)
		}
		return assembleCatalog(m3u, *cacheDir, true, useAPI)
	}

	model := tui.New(cat, players).
		WithState(st).
		WithTerminalPlayback(termVO).
		WithProbe(*probeConc, *probeTimeout).
		WithProbeAll(*probeAll).
		WithReload(reload).
		WithSources(builtinSources(useAPI))

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "iptv:", err)
		os.Exit(1)
	}
}

// builtinSources describes the non-removable upstream sources for the in-TUI
// manager. With the API on, the iptv-org entry reflects the API endpoint.
func builtinSources(api bool) []tui.SourceInfo {
	var out []tui.SourceInfo
	for i, s := range source.Defaults() {
		info := tui.SourceInfo{Name: s.Name, URL: s.URL, Builtin: true}
		if api && i == 0 { // the iptv-org default
			info.Name = "iptv-org (api)"
			info.URL = "https://iptv-org.github.io/api/"
		}
		out = append(out, info)
	}
	return out
}

// pick returns the first non-empty value.
func pick(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

// mergeSources appends user sources to the defaults, de-duplicating by URL.
func mergeSources(defaults []source.Source, extra []config.Source) []source.Source {
	seen := map[string]bool{}
	for _, s := range defaults {
		seen[s.URL] = true
	}
	out := append([]source.Source(nil), defaults...)
	for _, s := range extra {
		if s.URL == "" || seen[s.URL] {
			continue
		}
		seen[s.URL] = true
		name := s.Name
		if name == "" {
			name = "custom"
		}
		out = append(out, source.Source{Name: name, URL: s.URL})
	}
	return out
}

// assembleCatalog fetches the M3U sources (cached) and, when api is set, the
// iptv-org JSON API, merging everything into one catalog.
func assembleCatalog(m3uSources []source.Source, cacheDir string, refresh, api bool) (catalog.Catalog, error) {
	loader := source.NewLoader(cacheDir)
	var channels []catalog.Channel
	var firstErr error

	for _, s := range m3uSources {
		data, err := loader.Load(s, refresh)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue // one dead source shouldn't sink the whole app
		}
		channels = append(channels, m3u.Parse(data, s.Name)...)
	}

	if api {
		apiCh, err := iptvorg.Fetch(loader, refresh)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
		} else {
			channels = append(channels, apiCh...)
		}
	}

	if len(channels) == 0 && firstErr != nil {
		return catalog.Catalog{}, firstErr
	}
	return catalog.Build(channels), nil
}

// withoutSource returns sources with any entry matching url removed.
func withoutSource(sources []source.Source, url string) []source.Source {
	out := sources[:0:0]
	for _, s := range sources {
		if s.URL != url {
			out = append(out, s)
		}
	}
	return out
}

// orderPlayers puts the preferred player first if it is installed.
func orderPlayers(players []player.Player, preferred string) []player.Player {
	if preferred == "" || len(players) == 0 {
		return players
	}
	out := make([]player.Player, 0, len(players))
	var rest []player.Player
	for _, p := range players {
		if p.Name() == preferred {
			out = append(out, p)
		} else {
			rest = append(rest, p)
		}
	}
	return append(out, rest...)
}

func comma(n int) string {
	s := fmt.Sprintf("%d", n)
	if n < 1000 {
		return s
	}
	var out []byte
	for i := 0; i < len(s); i++ {
		if i > 0 && (len(s)-i)%3 == 0 {
			out = append(out, ',')
		}
		out = append(out, s[i])
	}
	return string(out)
}

func defaultCacheDir() string {
	if d, err := os.UserCacheDir(); err == nil {
		return filepath.Join(d, "iptv")
	}
	return filepath.Join(os.TempDir(), "iptv-cache")
}
