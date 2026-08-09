#!/usr/bin/env bash
# Push iptv-tui-bin to the AUR. Assumes:
#   - packaging/aur/iptv-tui-bin/{PKGBUILD,.SRCINFO} are generated for this
#     version (run `make aur/pkgbuild VERSION=x.y.z` first, which this target
#     depends on), and
#   - the matching GitHub release with the binaries already exists, otherwise
#     the published package will fail to build on users' machines.
set -euo pipefail

ver="${1:?usage: publish.sh <version, e.g. 1.3.0>}"
pkg="iptv-tui-bin"
root="$(cd "$(dirname "$0")/../.." && pwd)"
src="$root/packaging/aur/$pkg"

[ -f "$src/PKGBUILD" ] || { echo "missing $src/PKGBUILD, run make aur/pkgbuild VERSION=$ver" >&2; exit 1; }

work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT

echo "cloning ssh://aur@aur.archlinux.org/$pkg.git"
git clone "ssh://aur@aur.archlinux.org/$pkg.git" "$work"

cp "$src/PKGBUILD" "$work/PKGBUILD"

# Regenerate .SRCINFO from the PKGBUILD when makepkg is available (authoritative),
# otherwise ship the committed one.
if command -v makepkg >/dev/null 2>&1; then
	( cd "$work" && makepkg --printsrcinfo > .SRCINFO )
else
	cp "$src/.SRCINFO" "$work/.SRCINFO"
fi

cd "$work"
if git diff --quiet && git diff --cached --quiet; then
	echo "nothing to publish: AUR already matches this PKGBUILD"
	exit 0
fi
git add PKGBUILD .SRCINFO
git commit -m "$pkg $ver"
git push
echo "published $pkg $ver to the AUR"
