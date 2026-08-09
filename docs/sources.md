# Sources of truth

Playlists are **never hand-maintained**. The upstream repos are canonical; the
app fetches from them and rebuilds everything on demand.

## Primary (fetched today)

| Repo | What we use | URL |
|------|-------------|-----|
| [iptv-org/iptv](https://github.com/iptv-org/iptv) | `index.country.m3u` — all channels grouped by country | `https://iptv-org.github.io/iptv/index.country.m3u` |
| [Free-TV/IPTV](https://github.com/Free-TV/IPTV) | `playlist.m3u8` — curated working HD | `https://raw.githubusercontent.com/Free-TV/IPTV/master/playlist.m3u8` |

Defined in `common/source/source.go` (`Defaults`). Changing a source is a
one-line edit there.

## Wider iptv-org ecosystem (reference / roadmap)

Not fetched yet, but these are the authoritative upstreams for the data behind
the playlists — candidates for richer metadata, guide, and search:

| Repo | Purpose | Possible use here |
|------|---------|-------------------|
| [iptv-org/database](https://github.com/iptv-org/database) | Canonical channel/stream data (CSV) | authoritative names, logos, categories, country |
| [iptv-org/api](https://github.com/iptv-org/api) | JSON API generated from the database | fetch metadata without scraping M3U |
| [iptv-org/sdk](https://github.com/iptv-org/sdk) | Typed client for the API | structured ingestion instead of M3U parsing |
| [iptv-org/epg](https://github.com/iptv-org/epg) | Program guides (XMLTV) | now/next overlay in the detail pane |
| [iptv-org/awesome-iptv](https://github.com/iptv-org/awesome-iptv) | Curated resources | discover additional sources |
| [iptv-org/community](https://github.com/iptv-org/community) | Discussions / issues | upstream data fixes |

## Rebuilding local playlists

The app is the rebuild tool — it always pulls fresh from the repos:

```bash
make rebuild                 # -> ./playlists/all.m3u + ./playlists/countries/*.m3u
# or:
./iptv -export ./playlists   # same, to any directory
./iptv -refresh              # refresh the TUI's cache without exporting files
```

`-export` forces a fresh download regardless of cache age, so exported files
always reflect current upstream. The TUI itself caches for 12h and can be forced
with `-refresh`.

## API ingestion (`-api`, implemented)

The **iptv-org API** path is available via `-api` / `"api": true`. It joins
`channels.json` + `streams.json` + `countries.json` into enriched channels with
**stable IDs**, proper country names, and categories (`common/iptvorg`). It
replaces **only** the built-in iptv-org M3U — Free-TV and user sources still flow
through the M3U parser — so it can't break user lists. The M3U path remains the
default (simple, dependency-free).

Still roadmap: `database`/`sdk` for logos and richer typing; guide via
`iptv-org/epg` (the detail-card EPG currently uses epgshare01).
