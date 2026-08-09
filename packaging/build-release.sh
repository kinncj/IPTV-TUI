#!/usr/bin/env bash
# Cross-compile release binaries, stage the man page + LICENSE, and write
# SHA256SUMS into dist/. These are the assets uploaded to the GitHub release and
# consumed by scripts/install.sh and the AUR -bin PKGBUILD.
#
# Asset names follow <name>_<os>_<arch>: iptv-tui_linux_amd64, etc.
set -euo pipefail

cd "$(dirname "$0")/.."
ver="${1:-dev}"
dist="dist"
rm -rf "$dist"
mkdir -p "$dist"

ldflags="-s -w -X main.version=${ver}"

build() {
	local os="$1" arch="$2" out="iptv-tui_${1}_${2}"
	echo "building ${os}/${arch} -> ${out}"
	CGO_ENABLED=0 GOOS="$os" GOARCH="$arch" \
		go build -trimpath -ldflags "$ldflags" -o "$dist/$out" ./app
}

build linux amd64
build linux arm64
build darwin amd64
build darwin arm64

cp packaging/iptv-tui.1 "$dist/iptv-tui.1"
cp LICENSE "$dist/LICENSE"

# Checksums over every asset, for the installer to verify.
( cd "$dist" && sha256sum iptv-tui_* iptv-tui.1 LICENSE > SHA256SUMS )

echo
echo "release assets in $dist/:"
ls -la "$dist"
echo
cat "$dist/SHA256SUMS"
