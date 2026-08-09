# IPTV TUI

A terminal UI to browse free IPTV playlists by country or category and play the
channels. It probes each stream's reachability as you browse, so dead and
geo-blocked channels are visible before you open them.

The app is a browser and launcher. It hosts nothing. It reads public M3U
playlists and hands a stream URL to a player you already have (mpv, vlc, or
ffplay). The playlists come from two independent community projects:

- iptv-org/iptv: https://github.com/iptv-org/iptv
- Free-TV/IPTV: https://github.com/Free-TV/IPTV

This project is not affiliated with either, and is not responsible for the
streams they list. See [DISCLAIMER.md](DISCLAIMER.md).

## Install

Arch Linux (AUR):

```bash
paru -S iptv-tui-bin     # prebuilt binary, installs to /usr/bin/iptv-tui
```

Any Linux or macOS, with the install script (downloads the release binary to
`/usr/local/bin`, using sudo only if that directory needs it):

```bash
curl -fsSL https://raw.githubusercontent.com/kinncj/IPTV-TUI/main/scripts/install.sh | sh
# install somewhere else:
curl -fsSL .../scripts/install.sh | sh -s -- --install-location ~/.local/bin
```

From source:

```bash
make tui/run                 # build ./iptv and run
# or
make tui/build && ./iptv
```

Requirements to build from source: Go 1.26+. For playback you need a player:
mpv (recommended, and required for inline terminal playback), vlc, or ffplay.

## Screens

- Home: an animated block-logo splash. Press enter to browse.
- Countries or categories: a filterable group list with a live scrollbar.
  Favorites and recently played appear on top once you have some.
- Channels: a two-pane view with the channel list on the left and a detail card
  on the right (reachability, favorite, EPG now and next, group, source, tvg-id,
  url).
- Sources: an in-app manager to add or remove your own playlists.
- Help: press `?` for the keybindings overlay.

Each channel shows a reachability icon: `✓` reachable, `⚠` blocked (usually
geo-blocked), `✗` dead, `…` checking. This is how you land on a working stream
on the first try instead of hitting "unable to open" in a player.

## Playing

Press enter to play in an external player window, or `t` to play inline in the
terminal with mpv. The inline output is chosen for your terminal: kitty graphics
on Ghostty, kitty, and WezTerm; sixel where supported; truecolor blocks
everywhere else. Force it with `-vo kitty|sixel|tct` or `IPTV_TERM_VO`, which is
useful over SSH or tmux where auto-detection cannot see your local terminal.

## Keys

| Key             | Action                                     |
|-----------------|--------------------------------------------|
| `↑`/`↓` `j`/`k` | move                                       |
| `enter`         | open a group, or play in a window          |
| `t`             | play inline in the terminal (mpv)          |
| `f`             | toggle favorite                            |
| `g`             | switch grouping: country or category       |
| `esc` `h`       | back                                       |
| `/`             | filter the current list                    |
| `r`             | re-probe the current group                 |
| `R`             | reload playlists from source               |
| `s`             | open the sources manager                   |
| `p`             | cycle the external player                  |
| `?`             | toggle help                                |
| `q` `ctrl+c`    | quit                                       |

In the sources manager: `a` adds a playlist URL, `d` removes a user source, `esc`
goes back. Added sources are written to the gitignored `iptv.local.json`, and the
catalog reloads from source. The built-in sources (iptv-org, Free-TV) cannot be
removed.

## Flags

```
iptv-tui                     browse (12h cached playlists)
iptv-tui -version            print the version
iptv-tui -refresh            force a re-download, then browse
iptv-tui -player vlc         prefer a specific player
iptv-tui -vo tct             force the inline video output (SSH/tmux)
iptv-tui -theme catppuccin   pick a theme (default: auto, follows OS light/dark)
iptv-tui -themes             list themes
iptv-tui -api                ingest the iptv-org source from its JSON API
iptv-tui -probe-concurrency 12
iptv-tui -probe-timeout 6s
iptv-tui -probe-all          probe every channel in the background
iptv-tui -cache DIR          override the cache directory
iptv-tui -export DIR         rebuild all.m3u + per-country playlists, then exit
```

### iptv-org API mode

By default the built-in iptv-org source is the country-grouped M3U. With `-api`
(or `"api": true` in config) it comes from the iptv-org JSON API instead, which
joins channels, streams, and countries for stable channel IDs (better EPG
matching), proper country names, and categories. This swaps only the built-in
iptv-org list. Free-TV and your own lists are untouched, so turning it on cannot
break your sources.

## Themes and terminal support

The palette follows your terminal's light or dark background by default (`auto`).
Named themes: tokyonight (default), catppuccin, gruvbox, nord, everforest, and
the distro flavors arch, cachyos, and omarchy, plus light variants. Pick one with
`-theme`, `IPTV_THEME`, or the config file.

Truecolor terminals (Ghostty, kitty, WezTerm) get the full palette, forced on so
it survives multiplexers like herdr and tmux that under-report it. Lower-color
terminals get the same palette downgraded to 256 or 16 colors. Terminals without
UTF-8 fall back to ASCII glyphs and borders.

## Local config

Config is optional. Add your own playlists on top of the upstream lists, or pin a
theme and player, through a gitignored file. See [docs/config.md](docs/config.md).

```json
// ./iptv.local.json (gitignored) or ~/.config/iptv/config.json
{ "theme": "catppuccin", "player": "mpv",
  "sources": [ { "name": "my-list", "url": "https://example.com/list.m3u8" } ] }
```

Precedence for theme and player is flag, then environment, then config, then the
default.

## Rebuild playlists from source

The app is also the playlist builder. It pulls fresh from the upstream repos and
writes standard M3U files any player can read:

```bash
make playlists/rebuild       # ./playlists/all.m3u + ./playlists/countries/*.m3u
```

## Make targets

Targets are slash-namespaced:

```
make tui/build          build ./iptv
make tui/run            build and run
make tui/install        go install into GOBIN
make playlists/rebuild  export playlists from upstream
make test               unit tests
make test/race          tests with the race detector
make test/acceptance    end-to-end pipeline tests
make lint               go vet + gofmt check
make ci                 lint + race tests + build
make release/build      cross-compile binaries + man + LICENSE into dist/
make gh/release         create the GitHub release and upload assets
make aur/pkgbuild       regenerate the -bin PKGBUILD from dist/
make aur/release        build, regenerate, and push iptv-tui-bin to the AUR
```

Release and AUR targets take `VERSION=x.y.z` (semver). See
[packaging/RELEASE.md](packaging/RELEASE.md).

## Architecture

Clean Architecture; dependencies point inward toward the domain. See
[docs/ADR-0001-iptv-tui.md](docs/ADR-0001-iptv-tui.md) and
[docs/architecture.md](docs/architecture.md).

```
app/
  main.go        composition root (wiring, flags, config, -export)
  tui/           Bubble Tea: model, update, view, theme, logo, delegate, sources
common/
  catalog/       domain model, Channel/Group/Catalog (no I/O)
  m3u/           M3U parser
  iptvorg/       iptv-org JSON API ingestion (opt-in)
  source/        playlist fetch + disk cache
  probe/         stream reachability checks
  player/        Player interface + mpv/vlc/ffplay, inline terminal playback
  export/        regenerate playlists from a catalog
  config/        gitignored user config
  state/         persisted favorites and last-played
  epg/           XMLTV program guide, now and next
  termcaps/      terminal capability detection
tests/features/  acceptance tests
docs/            ADR, architecture, sources, config
packaging/       man page, release build, AUR PKGBUILDs
```

## Content and legal

The streams and their legality are the responsibility of the upstream list
maintainers and the original broadcasters, not of this project. Nothing is hosted
here. Full notice in [DISCLAIMER.md](DISCLAIMER.md).

## Contributing and security

See [CONTRIBUTING.md](CONTRIBUTING.md), [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md),
and [SECURITY.md](SECURITY.md). Report security issues privately, not as a public
issue.

## License

MIT. See [LICENSE](LICENSE).
