# Favorites, recent, and the guide

## Favorites

Press `f` on a channel to favorite it, and `f` again to remove it. Favorites get
a star in the list and in the detail card.

Favorited channels also appear as a "Favorites" group pinned to the top of the
group list, so you can reach them without hunting through countries.

## Recently played

Every channel you play is recorded. The most recent ones appear as a "Recent"
group at the top of the group list, newest first. The list is de-duplicated by
URL and capped, so it stays short.

## Where this is stored

Favorites and history persist to `state.json` in the cache directory
(`~/.config/iptv-tui` by default). They survive restarts. Deleting that file resets
them.

## Program guide (EPG)

The detail card shows the current and next programme for a channel when guide
data is available. When you open a country with a known guide, the app fetches
its XMLTV data in the background and matches it to channels, first by tvg-id and,
failing that, by a normalized channel name (so guides that use a different id
scheme still match).

Coverage depends on the guide source:

- By default, guides come from a public XMLTV source (epgshare01), per country.
  It uses its own channel ids and lineups, so it matches some channels and misses
  others. When it matches, you get real now/next; when it doesn't, the lines are
  simply absent.
- For accurate, full coverage, point at a guide whose channel ids line up with
  the playlists. Set `epg_url` in `~/.config/iptv-tui/config.json` to a single
  XMLTV file (`.xml` or `.xml.gz`). The natural fit is a self-hosted
  [iptv-org/epg](https://github.com/iptv-org/epg) guide, whose ids match the
  iptv-org tvg-ids exactly.

```json
{ "epg_url": "https://your-host/guide.xml.gz" }
```

A very large guide is skipped rather than blocking the interface. If a channel
shows no guide, either the source has no data for it or the ids and names do not
match. The channel still plays; only the now/next lines are absent.
