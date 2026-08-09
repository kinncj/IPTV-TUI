# Getting started

## Install

On Arch Linux, install from the AUR (package name IPTV TUI). It installs the
binary to `/usr/bin/iptv-tui`:

```bash
paru -S iptv-tui-bin     # prebuilt
# or
paru -S iptv-tui         # build from source
```

On any Linux or macOS, use the install script. It downloads the release binary
for your OS and architecture to `/usr/local/bin`, using sudo only if that
directory is not writable by you:

```bash
curl -fsSL https://raw.githubusercontent.com/kinncj/IPTV-TUI/main/scripts/install.sh | sh
```

To install elsewhere (no sudo):

```bash
curl -fsSL .../scripts/install.sh | sh -s -- --install-location ~/.local/bin
```

From source, with Go 1.26+:

```bash
git clone https://github.com/kinncj/IPTV-TUI
cd iptv
make tui/build     # produces ./iptv
```

## A player

Playback shells out to a player you already have. Install at least one:

- mpv (recommended, and required for inline terminal playback)
- vlc
- ffplay (from ffmpeg)

## First run

```bash
iptv-tui
```

On first launch it downloads the two upstream playlists and caches them for 12
hours. Press enter on the splash to browse countries. Move with the arrow keys or
`j`/`k`, press enter on a country to see its channels, and press enter on a
channel to play it in a window, or `t` to play it inline in the terminal.

Press `?` at any time for the full keybindings.

## Where things are stored

- Cached playlists and your favorites and history live under the cache directory
  (`$XDG_CACHE_HOME/iptv`, usually `~/.cache/iptv`).
- Optional config lives at `~/.config/iptv/config.json` or `./iptv.local.json`.

See [sources and config](sources.md) for how to add your own lists.
