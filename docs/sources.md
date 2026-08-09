# Sources of truth

Playlists are never hand-maintained. The upstream repos are canonical. The app
fetches from them and rebuilds everything on demand.

## Primary (fetched today)

| Repo | What we use | URL |
|------|-------------|-----|
| [iptv-org/iptv](https://github.com/iptv-org/iptv) | `index.country.m3u`, all channels grouped by country | `https://iptv-org.github.io/iptv/index.country.m3u` |
| [Free-TV/IPTV](https://github.com/Free-TV/IPTV) | `playlist.m3u8`, curated working HD | `https://raw.githubusercontent.com/Free-TV/IPTV/master/playlist.m3u8` |

Defined in `common/source/source.go` (`Defaults`). Changing a source is a
one-line edit there.

This project is not affiliated with either upstream and is not responsible for
the streams they list. See [DISCLAIMER.md](../DISCLAIMER.md).

## Wider iptv-org ecosystem

These are the authoritative upstreams for the data behind the playlists. Some are
used, others are candidates for richer metadata and guide data.

| Repo | Purpose | Use here |
|------|---------|----------|
| [iptv-org/api](https://github.com/iptv-org/api) | JSON API generated from the database | used by `-api` mode (`common/iptvorg`) |
| [iptv-org/epg](https://github.com/iptv-org/epg) | Program guides (XMLTV) | roadmap; the detail card currently uses epgshare01 |
| [iptv-org/database](https://github.com/iptv-org/database) | Canonical channel and stream data (CSV) | roadmap: logos, categories |
| [iptv-org/sdk](https://github.com/iptv-org/sdk) | Typed client for the API | roadmap: structured ingestion |
| [iptv-org/awesome-iptv](https://github.com/iptv-org/awesome-iptv) | Curated resources | discover additional sources |
| [iptv-org/community](https://github.com/iptv-org/community) | Discussions and issues | upstream data fixes |

## Rebuilding local playlists

The app is also the rebuild tool. It always pulls fresh from the repos:

```bash
make playlists/rebuild       # ./playlists/all.m3u + ./playlists/countries/*.m3u
# or
./iptv -export ./playlists   # same, to any directory
./iptv -refresh              # refresh the cache without exporting files
```

`-export` forces a fresh download regardless of cache age, so exported files
reflect current upstream. The app itself caches for 12 hours, which `-refresh`
overrides.

## API ingestion

The iptv-org API path is available with `-api` or `"api": true`. It joins
`channels.json`, `streams.json`, and `countries.json` into channels with stable
IDs, proper country names, and categories (`common/iptvorg`). It replaces only
the built-in iptv-org M3U. Free-TV and user sources still flow through the M3U
parser, so it cannot break user lists. The M3U path stays the default, which is
simple and has no extra dependencies.
