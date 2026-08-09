# Architecture

Clean Architecture. Every dependency points inward toward the domain. The domain
(`catalog`) imports nothing from the app.

```
                         app/main.go            composition root
                     flags, wiring, Program
                            |
        +-------------------+--------------------+
        v                   v                    v
   app/tui             common/source        common/player
 Bubble Tea MVU        fetch + cache        Player interface,
 (presentation)        (infrastructure)     mpv/vlc/ffplay
        |                   |
        v                   v
   common/probe        common/m3u
   reachability        parser
        |                   |
        +---------+---------+
                  v
           common/catalog     domain: Channel / Group / Catalog
             (no I/O)         imports nothing internal
```

## Data flow

1. `main` asks `source.Loader` for each playlist (cache first, 12h freshness).
   With `-api`, the iptv-org portion comes from `common/iptvorg` instead.
2. Bytes go through `m3u.Parse` into `[]catalog.Channel`, stamped with provenance.
3. `catalog.Build` groups and sorts them into a `Catalog`. `BuildByCategory`
   produces the alternate grouping for the `g` toggle.
4. `tui.New` renders groups. Opening one lists its channels and starts the probe
   commands, which stream results back as messages. A supported country also
   triggers an EPG fetch.
5. Enter on a channel calls the selected `player.Player`. `t` runs mpv inline via
   `tea.ExecProcess`. Neither blocks the UI on playback.

## Key points

Presentation owns no rules. `app/tui` navigates and renders. Parsing, grouping,
fetching, probing, playing, and persistence all live in `common/*` behind
interfaces or pure functions.

`Player` is an interface (`Name`, `Available`, `Play`) with adapters for mpv,
vlc, and ffplay, plus a builder for inline terminal playback. Adding a player is
one file; the TUI does not change.

Probe results are shared state, not list data. A mutex-guarded `statusStore` is
read by list items at render time, so async probe updates appear without
rebuilding the list.

Failure is degraded, not fatal. A dead source still lets the others load. Network
down falls back to a stale cache. No player produces a status message, not a
crash. A malformed config file is reported and skipped.

## Testing

- `common/m3u`, `common/catalog`, `common/iptvorg`, `common/epg`,
  `common/config`, `common/state`, `common/termcaps`: unit tests.
- `app/tui`: a render test builds the model and renders every screen at several
  sizes, which catches layout and rune-handling bugs.
- `tests/features`: an end-to-end test wiring parse into build into probe-icon
  assertions, with no network or terminal.
- `make test/race` guards the concurrent probe and sweep paths.
