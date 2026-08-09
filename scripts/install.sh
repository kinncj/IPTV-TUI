#!/usr/bin/env bash
# Build and install the iptv TUI into ~/.local/bin (or $PREFIX/bin).
set -euo pipefail

cd "$(dirname "$0")/.."

PREFIX="${PREFIX:-$HOME/.local}"
BINDIR="$PREFIX/bin"

command -v go >/dev/null || { echo "go is required (1.26+)"; exit 1; }

echo "building..."
make build-tui

mkdir -p "$BINDIR"
install -m 0755 iptv "$BINDIR/iptv"
echo "installed -> $BINDIR/iptv"

if ! command -v mpv >/dev/null && ! command -v vlc >/dev/null && ! command -v ffplay >/dev/null; then
	echo "warning: no player found — install mpv (recommended), vlc, or ffplay"
fi

case ":$PATH:" in
	*":$BINDIR:"*) ;;
	*) echo "note: $BINDIR is not on PATH; add it to use 'iptv' directly" ;;
esac
