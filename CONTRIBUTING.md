# Contributing

Thanks for your interest in IPTV TUI. This is a Go and Bubble Tea terminal app.

## Before you start

- By contributing you agree to license your work under the project's
  [MIT license](LICENSE).
- Be aware of what this project is and is not: it browses and launches
  third-party playlists, it does not host streams. See [DISCLAIMER.md](DISCLAIMER.md).
  Pull requests that add hosting, scraping of paywalled content, or bundled
  stream URLs will not be merged.

## Development

Requirements: Go 1.26+, and a player (mpv, vlc, or ffplay) to test playback.

```bash
make tui/run          # build and run
make test             # unit tests
make test/race        # tests with the race detector
make lint             # go vet + gofmt check
make ci               # the full gate: lint + race tests + build
```

Run `make ci` before opening a pull request. It must pass.

## Layout

The code follows Clean Architecture; dependencies point inward toward the
domain. See [docs/architecture.md](docs/architecture.md) and
[docs/ADR-0001-iptv-tui.md](docs/ADR-0001-iptv-tui.md).

- `common/` holds the domain and infrastructure packages (no UI).
- `app/tui/` is the Bubble Tea presentation layer.
- `app/main.go` wires everything together.

## Pull requests

- Keep changes focused. One concern per pull request.
- Add or update tests for behavior you change. The TUI has a render test that
  exercises every screen; keep it passing.
- Match the surrounding code style. `gofmt` is enforced by `make lint`.
- Write a clear description of what changed and why.

## Reporting bugs and asking for features

Open an issue with the version (`iptv-tui -version`), your terminal and OS, and
steps to reproduce. For playback problems, note which player you used and whether
the stream opened in that player directly.

For security issues, do not open a public issue. See [SECURITY.md](SECURITY.md).
