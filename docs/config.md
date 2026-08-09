# Local configuration

Config is **optional** — the app works with zero config. Use it to add your own
playlists on top of the upstream source-of-truth repos, or to pin a theme/player.

## Files (merged in order)

| Path | Scope | Tracked? |
|------|-------|----------|
| `$XDG_CONFIG_HOME/iptv/config.json` (usually `~/.config/iptv/config.json`) | per-user | n/a |
| `./iptv.local.json` | this repo | **gitignored** |

Later files win for scalar settings (`theme`, `player`); `sources` are appended.
A malformed file is reported on stderr and skipped — it never crashes the app.

## Format

```json
{
  "theme": "catppuccin",
  "player": "mpv",
  "api": false,
  "sources": [
    { "name": "my-extra-list", "url": "https://example.com/playlist.m3u8" },
    { "name": "regional",      "url": "https://example.com/regional.m3u" }
  ]
}
```

- **`sources`** — extra M3U/M3U8 playlists, added *after* iptv-org and Free-TV.
  De-duplicated by URL. They flow through the same parser, so their
  `group-title` tags become navigable groups just like the built-ins.
- **`theme`** — any name from `iptv -themes` (or `auto`).
- **`player`** — `mpv` | `vlc` | `ffplay`.
- **`api`** — `true` ingests the built-in iptv-org source from its JSON API
  (stable IDs, categories) instead of M3U. Affects only that built-in source;
  Free-TV and your `sources` are unchanged.

## Precedence

For theme and player: **flag > environment > config > default**.

```
iptv -theme gruvbox          # flag wins
IPTV_THEME=nord iptv         # env, if no flag
# else config.theme, else auto
```

## Roadmap

Adding/removing sources from *inside* the TUI (a small form that writes
`iptv.local.json`) is planned — see ADR-0001 next steps. Today, edit the JSON
and relaunch (or press `-refresh` semantics via restart).
