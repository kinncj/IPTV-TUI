# AUR packaging

Two packages, both installing the binary as `/usr/bin/iptv-tui` and a man page.
They `provides`/`conflicts` each other.

- `iptv-tui-bin`: prebuilt per-arch binaries from the GitHub release.
- `iptv-tui`: builds from the tagged source tarball (`makedepends=go`).

License is MIT; the license file is installed under
`/usr/share/licenses/iptv-tui*/LICENSE`.

## Manual publish

```bash
make aur/release VERSION=1.3.0
```

This builds the release assets, regenerates `iptv-tui-bin/PKGBUILD` with real
checksums, and pushes to the AUR (via `packaging/aur/publish.sh`). It assumes the
matching GitHub release already exists, otherwise the published package fails to
build on users' machines. See [../RELEASE.md](../RELEASE.md) for the full order.

## Automatic publish from CI

The release workflow (`.github/workflows/release.yml`) can publish the AUR
package when a GitHub release is published. It is off until you opt in:

1. Add a repository variable `ENABLE_AUR` set to `true`.
2. Add a repository secret `AUR_SSH_PRIVATE_KEY` containing the private key whose
   public half is registered on your AUR account.

The CI job downloads `SHA256SUMS` from the release, regenerates the PKGBUILD, and
deploys with `KSXGitHub/github-actions-deploy-aur`. Prereleases are skipped.

## Checksums

`gen-pkgbuild.sh` fills `sha256sums` either from a `SHA256SUMS` file (set `SUMS`,
which CI does) or by hashing the local `dist/` assets. Regenerate `.SRCINFO`
locally with `makepkg --printsrcinfo > .SRCINFO` if you edit the PKGBUILD by hand.
