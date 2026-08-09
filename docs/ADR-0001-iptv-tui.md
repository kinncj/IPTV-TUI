# ADR-0001: IPTV browsing TUI

## Title
A Go + Bubble Tea TUI to browse public IPTV playlists by country and launch a player.

## Context
Two public playlists are the practical source of free channels: iptv-org
(exhaustive, ~15k channels, every country) and Free-TV (curated HD, ~2k). Both group
channels by country via `group-title`.

The problem is not *finding* channels, it is that a large fraction of iptv-org
URLs are dead or geo-blocked. Pointing VLC at the raw list produces "unable to
open MRL" errors with no way to know which channel will work. A prior approach
generated static per-country `.m3u` files with Python; it organized the data but
did nothing about reachability, so the core pain remained.

## Goals
- Browse country → channel in a fast keyboard-driven TUI.
- Show per-stream reachability (✓ / ⚠ / ✗) before launch.
- Launch the best available player (mpv preferred; vlc/ffplay fallback).
- Work offline once primed; survive a dead upstream source.

## Non-goals (v1)
- EPG / program-guide overlay.
- Recording / time-shift.
- Persisted favorites and history (structure allows adding later).
- In-TUI playback (we shell out to a real player).

## Proposal
A single Go module, BusinessRepo-structured, using Bubble Tea for the UI.

Layering (Clean Architecture; dependencies point inward):

| Layer          | Package          | Responsibility                                  |
|----------------|------------------|-------------------------------------------------|
| Domain         | `common/catalog` | Entities `Channel`/`Group`/`Catalog`; no I/O.   |
| Domain service | `common/m3u`     | Pure playlist parsing (bytes → channels).       |
| Infrastructure | `common/source`  | Fetch playlists + on-disk cache.                |
| Infrastructure | `common/probe`   | Bounded-concurrency reachability checks.        |
| Infrastructure | `common/player`  | `Player` interface + mpv/vlc/ffplay adapters.   |
| Presentation   | `app/tui`        | Bubble Tea model/update/view; navigation only.  |
| Composition    | `app/main.go`    | Wire the above; flags; run the program.         |

Reachability is probed lazily: only when a country is opened, and only for its
channels, via a ranged GET (`Range: bytes=0-1`) through a 12-way semaphore with a
6s timeout. Results live in a mutex-guarded store that list items read on every
render, so probe updates surface without rebuilding the list.

## Alternatives
1. **Static per-country `.m3u` files (the Python approach).** Rejected: no
   reachability signal, which leaves the actual problem unsolved.
2. Probe all ~17k channels at startup. Rejected: slow, heavy, and wasteful,
   most are never viewed. Lazy per-group probing costs a small burst instead.
3. **Embed a player (libmpv/ffmpeg bindings).** Rejected for v1: large cgo
   surface and packaging cost; shelling out to a real player is simpler and lets
   the user keep their configured mpv/vlc.
4. **TUI framework: tview/tcell.** Rejected: the user's other TUIs (maple,
   Heimdall) standardize on Bubble Tea; consistency wins.

## Trade-offs and Risks
- **Probe ≠ playability.** A probe can pass and playback still fail (auth, codec)
  or fail while a stream plays (some CDNs reject ranged GETs). Documented; the
  icon is a hint, not a guarantee.
- **Burst of HTTP requests** when opening a large country. Bounded by the
  semaphore; verdicts are cached so re-entry is free.
- **Upstream drift.** Source URLs can change; they are isolated in
  `source.Defaults()` and cached, so a change is a one-line fix and a dead fetch
  degrades to cache rather than crashing.

## Impact
- New standalone binary `iptv`; no service, no persistent state beyond a cache.
- FinOps: zero infra cost (local CLI). Egress is a few MB of playlist per 12h.
- Ops: no daemon; failure modes are "network down" (→ cache) and "no player"
  (→ status message), both handled without crashing.

## Decision
Adopt the layered Go + Bubble Tea TUI above. Retire the Python generator.

## Next Steps
- Done (v1): animated 3D logo, player pill + switch toast, theme packs with
  OS light/dark follow, terminal capability detection (truecolor/degrade,
  Ghostty/herdr/tmux, unicode fallback), `-export` playlist rebuild, optional
  gitignored config for extra sources.
- Done (v1.1): **inline terminal playback** (mpv kitty/sixel/tct via
  `tea.ExecProcess`); **favorites + last-played** persistence (`common/state`,
  `state.json`) surfaced as Favorites/Recent groups; **in-TUI source manager**
  (`s`, writes `iptv.local.json`, reloads); **EPG now/next overlay** in the
  detail card (`common/epg`, per-country XMLTV from epgshare01, size-guarded).
- Done (v1.2): iptv-org API ingestion (`-api`, `common/iptvorg`), opt-in,
  swaps only the built-in iptv-org M3U for the JSON API (stable IDs, categories,
  country names); user + Free-TV lists untouched. **Probe flags**
  (`-probe-concurrency`, `-probe-timeout`) and a batched **`-probe-all`**
  background sweep. **CI** (`.github/workflows/ci.yml` → `make ci` = lint +
  race tests + build). Built-in sources remain non-removable in the manager.
- v1.3 (in progress): category grouping done. `g` toggles country/category
  (`catalog.BuildByCategory` + `Flatten`, uses the API-ingested `Category`).
  **AUR release** prepared as **IPTV TUI**: AGPL-3.0, `iptv-tui` (source) +
  `iptv-tui-bin` (prebuilt, per-arch) under `packaging/aur/`, cross-compile via
  `make release`, checksums via `packaging/aur/gen-pkgbuild.sh`, runbook in
  `packaging/RELEASE.md`, man page `packaging/iptv-tui.1`, `-version` flag.
- v1.3 remaining: iptv-org **sdk/database** for logos (inline via kitty/sixel);
  native EPG via `iptv-org/epg`.
