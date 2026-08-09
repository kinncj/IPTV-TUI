# Themes and terminal support

## Choosing a theme

The default is `auto`, which follows your terminal's background: a dark theme on a
dark background, a light theme on a light one.

Named themes: tokyonight (default dark), catppuccin, gruvbox, nord, everforest,
and the distro flavors arch, cachyos, and omarchy. There are light variants too
(for example tokyonight-light, catppuccin-latte).

List them:

```bash
iptv-tui -themes
```

Pick one, in order of precedence:

```bash
iptv-tui -theme gruvbox        # flag wins
IPTV_THEME=nord iptv-tui       # then environment
# then "theme" in config, then auto
```

## How color depth is handled

The same palette is defined once in full-color hex. The terminal library
downgrades it to whatever your terminal supports:

- Truecolor terminals (Ghostty, kitty, WezTerm) get the full 24-bit palette.
- 256-color and 16-color terminals get the nearest match automatically.

Truecolor is forced on when detected, so a multiplexer that under-reports color
(herdr, tmux) does not knock the app down to 256 colors.

## Non-UTF-8 terminals

If the locale is not UTF-8, the app swaps its block characters, box borders, and
status icons for ASCII equivalents, so the layout still reads correctly. The
animated logo falls back to a plain wordmark.

## What gets detected

At startup the app checks:

- color depth (truecolor, 256, or basic),
- dark or light background,
- whether the locale is UTF-8,
- the terminal program (for example Ghostty), and
- whether it is running under a multiplexer (tmux or herdr).

These drive the theme resolution, the glyph set, and the default inline video
output. You can override the video output with `-vo`; see
[playing channels](playback.md).
