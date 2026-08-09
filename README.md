# iptv

A terminal UI to browse free IPTV playlists by country and launch a player —
built to make the dead/geo-blocked streams in these public lists *visible before
you click them*.

Source of truth (never hand-maintained — always rebuilt from upstream):
- [iptv-org/iptv](https://github.com/iptv-org/iptv) — exhaustive, every country
- [Free-TV/IPTV](https://github.com/Free-TV/IPTV) — curated, working HD

See [docs/sources.md](docs/sources.md) for the full iptv-org ecosystem
(database / api / epg / sdk) and the rebuild flow.

## Screens

- **Home** — animated 3D block-logo splash, `enter` to browse.
- **Countries** — filterable groups with a live scrollbar; **Favorites** and
  **Recent** appear on top once you have some.
- **Channels** — two-pane: channel list on the left, a detail card on the right
  (status, favorite, **EPG now/next**, group, source, tvg-id, url).
- **Sources** — in-TUI manager (`s`) to add/remove your own playlists.
- **Help** — `?` toggles a keybindings overlay.

Each channel shows a live reachability icon: `✓` reachable · `⚠` blocked
(usually geo-blocked) · `✗` dead · `…` checking — so you launch a working stream
on the first try instead of hitting VLC's "unable to open MRL".

**Play in a window** (`enter`) or **inline in the terminal** (`t`, needs mpv) —
kitty-graphics on Ghostty/kitty/WezTerm, sixel where supported, truecolor blocks
everywhere else. **Favorite** with `f`; favorites and recently-played persist.

## Install

**Arch Linux (AUR)** — package name **IPTV TUI**:

```bash
paru -S iptv-tui-bin     # prebuilt binary
# or
paru -S iptv-tui         # build from source (needs go)
```

Installs `/usr/bin/iptv-tui` and a man page (`man iptv-tui`). See
[packaging/RELEASE.md](packaging/RELEASE.md) for the release/publish flow.

**From source:**

```bash
make run-tui        # build ./iptv and run
# or
make build-tui && ./iptv
```

## Requirements

- Go 1.26+ (to build from source)
- A player: **mpv** (preferred, required for inline terminal playback), **vlc**, or **ffplay**

## Keys

| Key             | Action                                     |
|-----------------|--------------------------------------------|
| `↑`/`↓` `j`/`k` | move                                       |
| `enter`         | open group / play channel in a window      |
| `t`             | play channel inline in the terminal (mpv)  |
| `f`             | toggle favorite                            |
| `g`             | switch grouping: country ↔ category        |
| `esc` `h`       | back                                       |
| `/`             | filter the current list                    |
| `r`             | re-probe the current country               |
| `R`             | reload playlists from source               |
| `s`             | open the sources manager                   |
| `p`             | cycle player (mpv → vlc → ffplay)          |
| `?`             | toggle help                                |
| `q` `ctrl+c`    | quit                                       |

In the **sources manager**: `a` add a playlist URL, `d` remove a user source,
`esc` back. Added sources are written to the gitignored `iptv.local.json` and
the catalog reloads from source. **Built-in sources (iptv-org, Free-TV) can't be
removed** — `d` on a built-in is refused.

## Flags

```
./iptv                      # browse (12h cached playlists)
./iptv -refresh             # force re-download, then browse
./iptv -player vlc          # prefer a specific player
./iptv -theme catppuccin    # pick a theme (default: auto, follows OS light/dark)
./iptv -themes              # list available themes
./iptv -api                 # ingest the iptv-org source from its JSON API (richer metadata)
./iptv -probe-concurrency N # max concurrent reachability probes (default 12)
./iptv -probe-timeout 6s    # per-probe timeout
./iptv -probe-all           # probe every channel in the background at startup
./iptv -cache DIR           # override cache directory
./iptv -export DIR          # rebuild all.m3u + countries/*.m3u from upstream, then exit
```

### iptv-org API mode (`-api`)

By default the built-in iptv-org source is the country-grouped **M3U**. With
`-api` (or `"api": true` in config) it's ingested from the **iptv-org JSON API**
instead — joining `channels`/`streams`/`countries` for **stable channel IDs**
(better EPG matching), proper country names, and categories. This swaps **only**
the built-in iptv-org list; **Free-TV and your own added lists are untouched**, so
turning it on can't break your sources.

## Themes & terminal

The palette **follows your terminal's light/dark background by default** (`auto`).
Named packs: `tokyonight` (default), `catppuccin`, `gruvbox`, `nord`,
`everforest`, plus distro flavors `arch`, `cachyos`, `omarchy`, and light
variants. Pick with `-theme`, `IPTV_THEME`, or the config file.

Capability-aware, so it's fancy when your tools are and still fancy when they're
not:
- **Truecolor** (Ghostty, kitty, WezTerm, …) gets the full 24-bit palette; it's
  forced on so it survives multiplexers (**herdr**, tmux) that under-report it.
- Lower-color terminals get the same palette **auto-downgraded** to 256/16 colors.
- **Non-UTF-8** terminals fall back to ASCII glyphs and borders.

## Local config (extra lists)

Add your own playlists on top of the upstream repos via an optional, gitignored
config — see [docs/config.md](docs/config.md):

```json
// ./iptv.local.json  (gitignored)  or  ~/.config/iptv/config.json
{ "theme": "catppuccin", "player": "mpv",
  "sources": [ { "name": "my-list", "url": "https://example.com/list.m3u8" } ] }
```

## Rebuild playlists from source

The app is also the playlist builder — it pulls fresh from the repos and writes
standard M3U files any other player can use:

```bash
make rebuild                # -> ./playlists/all.m3u + ./playlists/countries/*.m3u
```

## Make targets

```
make build-tui        build ./iptv
make run-tui          build and run
make refresh          run with -refresh
make rebuild          export playlists from upstream into ./playlists
make test             unit tests
make test-race        unit tests + race detector
make test-acceptance  end-to-end pipeline tests
make lint             go vet + gofmt check
make clean            remove binary and cache
```

## Architecture

Clean Architecture; dependencies point inward toward the domain. See
[docs/ADR-0001-iptv-tui.md](docs/ADR-0001-iptv-tui.md) and
[docs/architecture.md](docs/architecture.md).

```
app/
  main.go        composition root (wiring, flags, config, -export)
  tui/           Bubble Tea: model/update/view, theme, logo, delegate, components, glyphs
common/
  catalog/       domain model — Channel, Group, Catalog (no I/O)
  m3u/           pure M3U parser
  source/        playlist fetch + disk cache (source-of-truth URLs live here)
  probe/         stream reachability checks
  player/        Player interface + mpv/vlc/ffplay
  export/        regenerate on-disk playlists from a catalog
  config/        optional gitignored user config (extra sources, theme, player)
  termcaps/      terminal capability detection (color depth, dark/light, unicode)
tests/features/  acceptance tests
docs/            ADR + architecture + sources + config
```

## Notes

- Reachability is best-effort: a probe can pass yet the stream still fail in the
  player (and vice versa) due to geo-blocks, auth, or codec quirks.
- A dead upstream source won't sink the app — the other still loads, and a stale
  cache is used if the network is down.
