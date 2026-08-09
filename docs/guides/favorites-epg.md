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
(`~/.cache/iptv` by default). They survive restarts. Deleting that file resets
them.

## Program guide (EPG)

The detail card shows the current and next programme for a channel when guide
data is available. When you open a country that has a known guide, the app fetches
its XMLTV data in the background and matches it to channels by their tvg-id.

Notes and limits:

- Coverage is per country and best-effort. Guides come from a public XMLTV source
  (epgshare01), and not every country or channel has data.
- A very large country guide is skipped rather than blocking the interface.
- Matching depends on tvg-id. API mode (`iptv-tui -api`) gives channels stable
  iptv-org IDs, which can improve matching for some sources.

If a channel shows no guide, either its country has no published guide, its
tvg-id does not match, or the fetch was skipped for size. The channel still plays;
only the now/next lines are absent.
