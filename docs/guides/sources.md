# Sources and config

## The built-in sources

Two upstream lists ship with the app and cannot be removed:

- iptv-org/iptv: exhaustive, every country. https://github.com/iptv-org/iptv
- Free-TV/IPTV: curated, working HD. https://github.com/Free-TV/IPTV

They are the source of truth. The app fetches them fresh and caches them; press
`R` in the group list to reload from source, or run with `-refresh`.

## Adding your own lists

From inside the app, press `s` in the country list to open the sources manager.
It lists the built-in sources (marked `built`) and your own (marked `user`).

- `a` prompts for a playlist URL and adds it.
- `d` removes the selected user source. Built-in sources are refused.
- `esc` goes back.

Additions are written to `iptv.local.json` in the current directory (gitignored),
and the catalog reloads from source so the new channels appear.

You can also edit config by hand. Two files are read and merged, in this order:

1. `~/.config/iptv/config.json` (per user)
2. `./iptv.local.json` (repo-local, gitignored)

```json
{
  "theme": "catppuccin",
  "player": "mpv",
  "api": false,
  "sources": [
    { "name": "my-list", "url": "https://example.com/playlist.m3u8" }
  ]
}
```

Extra sources are added after the built-ins and de-duplicated by URL. A malformed
config file is reported on stderr and skipped; it never stops the app. Full field
reference in [docs/config.md](../config.md).

## Grouping by country or category

Press `g` in the group list to switch between grouping by country and grouping by
category. Category data is richest with API mode (below); with plain M3U most
channels have no category and fall under "Uncategorized".

## API mode

By default the built-in iptv-org source is the country-grouped M3U file. With
`-api` (or `"api": true` in config), it is ingested from the iptv-org JSON API
instead. The API path joins channels, streams, and countries into channels with:

- stable channel IDs, which improve program-guide matching,
- proper country names, and
- categories, which make category grouping useful.

API mode replaces only the built-in iptv-org list. Free-TV and any sources you
added stay on the M3U path and are merged in as before, so enabling it cannot
break your own lists.

## Rebuilding playlist files

To turn the current catalog into plain M3U files for other players:

```bash
make playlists/rebuild      # ./playlists/all.m3u + ./playlists/countries/*.m3u
# or
iptv-tui -export ./somewhere
```

This always pulls fresh from the upstream repos.
