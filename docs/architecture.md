# Architecture

Clean Architecture. Every arrow points inward toward the domain; the domain
(`catalog`) imports nothing from the app.

```
                         ┌───────────────────────────┐
                         │        app/main.go         │  composition root
                         │  flags · wiring · Program  │
                         └────────────┬──────────────┘
                                      │ constructs
              ┌───────────────────────┼────────────────────────┐
              ▼                       ▼                         ▼
     ┌─────────────────┐    ┌──────────────────┐      ┌──────────────────┐
     │    app/tui      │    │  common/source   │      │  common/player   │
     │ Bubble Tea MVU  │    │  fetch + cache   │      │ Player interface │
     │ (presentation)  │    │  (infrastructure)│      │  mpv/vlc/ffplay  │
     └───────┬─────────┘    └────────┬─────────┘      └──────────────────┘
             │ uses                  │ returns bytes
             ▼                       ▼
     ┌─────────────────┐    ┌──────────────────┐
     │  common/probe   │    │   common/m3u     │
     │ reachability    │    │  pure parser     │
     └───────┬─────────┘    └────────┬─────────┘
             │                       │ produces
             └───────────┬───────────┘
                         ▼
                ┌──────────────────┐
                │  common/catalog  │  domain — Channel / Group / Catalog
                │   (no I/O)       │  imports nothing internal
                └──────────────────┘
```

## Data flow

1. `main` asks `source.Loader` for each default playlist (cache-first, 12h TTL).
2. Bytes go through `m3u.Parse` → `[]catalog.Channel`, stamped with provenance.
3. `catalog.Build` groups + sorts them into a `Catalog`.
4. `tui.New` renders countries; opening one lists its channels and fires
   `probe` commands (bounded concurrency) that stream `probeResultMsg` back.
5. `enter` on a channel calls the selected `player.Player.Play(url)` — a detached
   process; the TUI never blocks on playback.

## Key design points

- **Presentation owns no rules.** `app/tui` only navigates and renders; parsing,
  grouping, fetching, probing, and launching are all in `common/*` behind
  interfaces or pure functions.
- **`Player` is an interface** (`Name/Available/Play`) with three adapters.
  Adding a player is one file; the TUI is unchanged. (Open/Closed.)
- **Probe results are shared state, not list data.** A mutex-guarded
  `statusStore` is read by list items at render time, so async probe updates
  appear without rebuilding the list — the UI stays responsive during a sweep.
- **Failure is degraded, not fatal.** Dead source → other source still loads;
  network down → stale cache; no player → status message, no crash.

## Testing strategy

- `common/m3u`, `common/catalog`: unit tests (pure, deterministic).
- `tests/features`: acceptance test wiring parse → build → probe-icon assertions
  against realistic multi-source fixtures, no network or TTY.
- `make test-race` guards the concurrent probe path.
