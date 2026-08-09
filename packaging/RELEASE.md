# Releasing IPTV TUI (GitHub + AUR)

Two AUR packages ship:

- `iptv-tui` builds from the tagged source tarball (`makedepends=go`).
- `iptv-tui-bin` installs prebuilt per-arch binaries from the GitHub release.

Both install the binary as `/usr/bin/iptv-tui` plus a man page, and they
`provides`/`conflicts` each other. License is MIT.

Versions are semver. The release and AUR make targets require `VERSION=x.y.z` and
reject anything that is not a bare semantic version.

## Automated path (recommended)

The release workflow does most of this. See
[packaging/aur/README.md](aur/README.md) for the AUR opt-in.

1. Tag and push:

   ```bash
   git tag v1.0.0 && git push origin v1.0.0
   ```

2. Publish a GitHub release for that tag (in the UI, or `gh release create v1.0.0
   --generate-notes`). The workflow builds the binaries, attaches them with
   `SHA256SUMS`, and, if `ENABLE_AUR` is set, pushes `iptv-tui-bin` to the AUR.

## Manual path

Use this if you are not running the workflow.

1. Commit, push, and tag:

   ```bash
   git add -A && git commit -m "iptv-tui v1.0.0"
   git push origin main
   git tag v1.0.0 && git push origin v1.0.0
   ```

2. Build binaries and publish the release:

   ```bash
   make gh/release VERSION=1.0.0
   ```

   This cross-compiles into `dist/` (linux and darwin, amd64 and arm64), plus the
   man page, LICENSE, and `SHA256SUMS`, then uploads them to the `v1.0.0` release.

3. Push the AUR package:

   ```bash
   make aur/release VERSION=1.0.0
   ```

   This regenerates `iptv-tui-bin/PKGBUILD` with the release checksums and pushes
   it. The GitHub release must exist first, or the published package fails to
   build on users' machines.

## The source package

`iptv-tui` builds from the tag tarball. After bumping its `pkgver`, regenerate
its `.SRCINFO`:

```bash
cd packaging/aur/iptv-tui && makepkg --printsrcinfo > .SRCINFO
```

The tarball uses `sha256sums=('SKIP')`, which is fine for GitHub archive tags.

## Verify

```bash
makepkg -si            # in a package dir, installs locally
iptv-tui -version
man iptv-tui
```
