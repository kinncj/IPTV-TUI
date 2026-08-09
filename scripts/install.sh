#!/bin/sh
# SPDX-License-Identifier: MIT
# Copyright (c) 2026 Kinn Coelho Juliao <kinncj@gmail.com>
#
# Install IPTV TUI from GitHub Releases.
#
#   curl -fsSL https://raw.githubusercontent.com/kinncj/IPTV-TUI/main/scripts/install.sh | sh
#   curl -fsSL .../scripts/install.sh | sh -s -- --install-location ~/.local/bin
#
# The binary installs to the system bin dir by default (/usr/local/bin),
# elevating with sudo only if that dir is not writable. Override with
# --install-location <dir> or $IPTV_BIN_DIR.
#
# Environment overrides:
#   IPTV_VERSION   release tag to install (default: latest)
#   IPTV_BIN_DIR   install directory (default: /usr/local/bin)
#   IPTV_REPO      source owner/repo (default: kinncj/IPTV-TUI)
set -eu

REPO="${IPTV_REPO:-kinncj/IPTV-TUI}"
VERSION="${IPTV_VERSION:-latest}"

err()  { echo "install: $*" >&2; exit 1; }
need() { command -v "$1" >/dev/null 2>&1 || err "missing required tool: $1"; }

usage() {
  cat <<EOF
Install IPTV TUI.

Usage: install.sh [--install-location DIR] [--install-deps]

  --install-location DIR   install into DIR (default: /usr/local/bin)
  --install-deps           also install mpv and ffmpeg with your package manager
  -h, --help               show this help
EOF
}

INSTALL_LOCATION=""
INSTALL_DEPS=0
while [ $# -gt 0 ]; do
  case "$1" in
    --install-location) INSTALL_LOCATION="${2:-}"; shift 2 ;;
    --install-location=*) INSTALL_LOCATION="${1#*=}"; shift ;;
    --install-deps) INSTALL_DEPS=1; shift ;;
    -h|--help) usage; exit 0 ;;
    *) err "unknown option: $1" ;;
  esac
done

need curl
need uname

# --- detect platform -------------------------------------------------------
os=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$os" in
  linux) os=linux ;;
  darwin) os=darwin ;;
  *) err "unsupported OS: $os" ;;
esac

arch=$(uname -m)
case "$arch" in
  x86_64|amd64) arch=amd64 ;;
  arm64|aarch64) arch=arm64 ;;
  *) err "unsupported architecture: $arch" ;;
esac

# --- resolve version -------------------------------------------------------
if [ "$VERSION" = "latest" ]; then
  VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
    | grep '"tag_name"' | head -1 | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')
  [ -n "$VERSION" ] || err "could not resolve the latest release for ${REPO}"
fi
echo "install: ${REPO} ${VERSION} (${os}/${arch})"

base="https://github.com/${REPO}/releases/download/${VERSION}"
asset="iptv-tui_${os}_${arch}"

# --- choose an install dir (system bin by default) -------------------------
if [ -n "$INSTALL_LOCATION" ]; then
  bindir="$INSTALL_LOCATION"
elif [ -n "${IPTV_BIN_DIR:-}" ]; then
  bindir="$IPTV_BIN_DIR"
else
  bindir="/usr/local/bin"
fi

SUDO=""
if [ ! -d "$bindir" ]; then
  mkdir -p "$bindir" 2>/dev/null || { command -v sudo >/dev/null 2>&1 && SUDO="sudo" && $SUDO mkdir -p "$bindir"; } \
    || err "cannot create ${bindir} (pass --install-location to a writable dir)"
fi
if [ -z "$SUDO" ] && [ ! -w "$bindir" ]; then
  command -v sudo >/dev/null 2>&1 && SUDO="sudo" || err "${bindir} is not writable and sudo is unavailable (use --install-location)"
fi

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

# --- checksum verification (best effort) -----------------------------------
sums=""
if curl -fsSL "${base}/SHA256SUMS" -o "${tmp}/SHA256SUMS" 2>/dev/null; then
  sums="${tmp}/SHA256SUMS"
fi
verify() { # <file> <asset-name>
  [ -n "$sums" ] || { echo "  (no SHA256SUMS published; skipping verification)"; return 0; }
  want=$(grep " $2\$" "$sums" 2>/dev/null | awk '{print $1}' | head -1)
  [ -n "$want" ] || { echo "  (no checksum entry for $2; skipping)"; return 0; }
  if command -v sha256sum >/dev/null 2>&1; then
    got=$(sha256sum "$1" | awk '{print $1}')
  else
    got=$(shasum -a 256 "$1" | awk '{print $1}')
  fi
  [ "$want" = "$got" ] || err "checksum mismatch for $2"
  echo "  verified $2"
}

# --- install ---------------------------------------------------------------
echo "downloading ${asset}"
curl -fSL "${base}/${asset}" -o "${tmp}/${asset}" || err "download failed: ${asset}"
verify "${tmp}/${asset}" "${asset}"
chmod +x "${tmp}/${asset}"
dest="${bindir}/iptv-tui"
$SUDO mv "${tmp}/${asset}" "$dest" || err "failed to install ${dest}"
echo "  installed ${dest}"

# Man page (best effort; never fail the install over a doc). `man iptv-tui`.
mandir="$(dirname "$bindir")/share/man/man1"
if curl -fsSL "${base}/iptv-tui.1" -o "${tmp}/iptv-tui.1" 2>/dev/null; then
  verify "${tmp}/iptv-tui.1" "iptv-tui.1" || true
  if $SUDO mkdir -p "$mandir" 2>/dev/null && $SUDO mv "${tmp}/iptv-tui.1" "${mandir}/iptv-tui.1" 2>/dev/null; then
    echo "  installed ${mandir}/iptv-tui.1  (man iptv-tui)"
  fi
fi

# --- runtime dependencies --------------------------------------------------
# The binary is static, but at runtime the app needs a player and ffmpeg:
#   mpv     recommended player, and the audio for the built-in inline player
#   ffmpeg  the built-in inline video player (and provides ffplay)
detect_pm() {
  if command -v brew >/dev/null 2>&1; then echo brew
  elif command -v pacman >/dev/null 2>&1; then echo pacman
  elif command -v dnf >/dev/null 2>&1; then echo dnf
  elif command -v zypper >/dev/null 2>&1; then echo zypper
  elif command -v apt-get >/dev/null 2>&1; then echo apt
  elif command -v emerge >/dev/null 2>&1; then echo emerge
  else echo ""; fi
}

gentoo_atom() { # map a package name to its Gentoo atom
  case "$1" in
    mpv)     echo "media-video/mpv" ;;
    ffmpeg)  echo "media-video/ffmpeg" ;;
    yt-dlp)  echo "net-misc/yt-dlp" ;;
    *)       echo "$1" ;;
  esac
}

install_cmd() { # <pm> <pkgs...>
  pm="$1"; shift; pkgs="$*"
  case "$pm" in
    brew)   echo "brew install ${pkgs}" ;;
    pacman) echo "sudo pacman -S --needed ${pkgs}" ;;
    dnf)    echo "sudo dnf install ${pkgs}" ;;
    zypper) echo "sudo zypper install ${pkgs}" ;;
    apt)    echo "sudo apt-get update && sudo apt-get install -y ${pkgs}" ;;
    emerge) atoms=""; for p in $pkgs; do atoms="${atoms} $(gentoo_atom "$p")"; done
            echo "sudo emerge --ask${atoms}" ;;
    *)      echo "" ;;
  esac
}

pm_note() {
  case "$1" in
    dnf)    echo "  (Fedora: full ffmpeg may need RPM Fusion, https://rpmfusion.org)" ;;
    zypper) echo "  (openSUSE: ffmpeg usually comes from the Packman repository)" ;;
    emerge) echo "  (Gentoo: these compile from source and can take a while)" ;;
  esac
}

have_player=0
if command -v mpv >/dev/null 2>&1 || command -v vlc >/dev/null 2>&1 || command -v ffplay >/dev/null 2>&1; then
  have_player=1
fi
have_ffmpeg=0
command -v ffmpeg >/dev/null 2>&1 && have_ffmpeg=1
have_ytdlp=0
command -v yt-dlp >/dev/null 2>&1 && have_ytdlp=1

missing=""
[ "$have_player" = 0 ] && missing="mpv"
[ "$have_ffmpeg" = 0 ] && missing="${missing:+$missing }ffmpeg"
[ "$have_ytdlp" = 0 ] && missing="${missing:+$missing }yt-dlp"

echo
if [ -z "$missing" ]; then
  echo "Runtime dependencies present."
else
  pm=$(detect_pm)
  cmd=$(install_cmd "$pm" $missing)
  echo "Runtime dependencies to install: ${missing}"
  echo "  mpv     recommended player, plus audio for the built-in inline player"
  echo "  yt-dlp  play YouTube channels (Free-TV lists some, marked with a Y)"
  echo "  ffmpeg  the built-in inline video player (and ffplay)"
  if [ -n "$cmd" ] && [ "$INSTALL_DEPS" = 1 ] && [ "$pm" != emerge ]; then
    echo "Installing: ${cmd}"
    sh -c "$cmd" || echo "  install failed; run it yourself: ${cmd}"
  elif [ -n "$cmd" ]; then
    echo "Install them with:"
    echo "  ${cmd}"
    pm_note "$pm"
    [ "$INSTALL_DEPS" = 1 ] && echo "  (skipped auto-install for this package manager; run the command above)"
    echo "Or re-run with --install-deps to install automatically."
  else
    echo "Install mpv and ffmpeg with your system's package manager, then run iptv-tui."
  fi
fi

echo
echo "Done. Ensure ${bindir} is on your PATH."
echo "Try:  iptv-tui        (or: man iptv-tui)"
