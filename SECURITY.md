# Security policy

## Reporting a vulnerability

Report suspected vulnerabilities privately. Do not open a public issue for a
security problem.

- Use GitHub's private advisory form: https://github.com/kinncj/IPTV-TUI/security/advisories/new
- Or email kinncj@gmail.com with "IPTV TUI security" in the subject.

Please include what you found, the steps to reproduce it, and the version
(`iptv-tui -version`). You will get an acknowledgement, and a fix or a decision
once the report is assessed.

## Scope

The app fetches remote playlists and program guides over HTTPS and launches an
external player on stream URLs. Useful things to look at:

- parsing of untrusted M3U and XMLTV input (`common/m3u`, `common/epg`),
- how stream URLs are passed to the player process (`common/player`),
- handling of the cache and config files on disk (`common/source`,
  `common/config`, `common/state`).

## Out of scope

The app plays third-party streams it does not control. The safety, content, and
availability of those streams are not part of this project. See
[DISCLAIMER.md](DISCLAIMER.md).

## Supported versions

Fixes land on the latest release. Older versions are not maintained.
