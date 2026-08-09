#!/usr/bin/env bash
# Cross-compile release binaries + stage the man page into dist/.
# Produces the assets uploaded to the GitHub release and consumed by the AUR
# -bin PKGBUILD (iptv-tui_linux_amd64, iptv-tui_linux_arm64, iptv-tui.1).
set -euo pipefail

cd "$(dirname "$0")/.."
ver="${1:-dev}"
dist="dist"
mkdir -p "$dist"

ldflags="-s -w -X main.version=${ver}"

build() {
	local arch="$1" out="$2"
	echo "building linux/${arch} -> ${out}"
	CGO_ENABLED=0 GOOS=linux GOARCH="$arch" \
		go build -trimpath -ldflags "$ldflags" -o "$dist/$out" ./app
}

build amd64 "iptv-tui_linux_amd64"
build arm64 "iptv-tui_linux_arm64"

cp packaging/iptv-tui.1 "$dist/iptv-tui.1"

echo
echo "release assets in $dist/:"
ls -la "$dist"
echo
echo "sha256:"
( cd "$dist" && sha256sum iptv-tui_linux_amd64 iptv-tui_linux_arm64 iptv-tui.1 )
