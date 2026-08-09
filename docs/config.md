# Local configuration

Config is optional. The app works with none. Use it to add your own playlists on
top of the upstream lists, or to pin a theme, player, or API mode.

## Files (merged in order)

| Path | Scope | Tracked? |
|------|-------|----------|
| `$XDG_CONFIG_HOME/iptv-tui/config.json` (usually `~/.config/iptv-tui/config.json`) | per user | n/a |
| `./iptv.local.json` | this repo | gitignored |

Later files win for scalar settings (`theme`, `player`, `api`). The `sources`
lists are appended. A malformed file is reported on stderr and skipped; it never
crashes the app.

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

- `sources`: extra M3U/M3U8 playlists, added after iptv-org and Free-TV, and
  de-duplicated by URL. They flow through the same parser, so their `group-title`
  tags become navigable groups like the built-ins.
- `theme`: any name from `iptv -themes`, or `auto`.
- `player`: `mpv`, `vlc`, or `ffplay`.
- `api`: `true` ingests the built-in iptv-org source from its JSON API (stable
  IDs, categories) instead of M3U. It affects only that built-in source. Free-TV
  and your `sources` are unchanged.

## Precedence

For theme and player, the order is flag, then environment, then config, then the
default.

```
iptv -theme gruvbox          # flag wins
IPTV_THEME=nord iptv         # environment, if no flag
# then config.theme, then auto
```

## Adding sources without editing JSON

You can add and remove sources from inside the app. Press `s` in the group list
to open the sources manager, then `a` to add a URL or `d` to remove one you
added. Additions are written to `~/.config/iptv-tui/config.json` for you.
