# Playing channels

There are two ways to play a channel.

## In an external window

Press enter on a channel. The app launches your player (mpv by default, then vlc,
then ffplay) in its own window and keeps running. Press `p` to cycle which player
is used; the footer shows the current one.

## Inline, in the terminal

Press `t`. If more than one backend is available, a small chooser appears so you
can pick how to play. There are two backends, and they work differently.

### mpv (external)

Sharp video, hardware-decoded. mpv takes its own alternate screen to paint
frames, the same way `less` or `vim` do, so while it plays mpv owns the screen:
the browser is not drawn behind the video. It opens as a centered modal at 78% of
the terminal; press `f` to toggle between that and fullscreen (live on the kitty
and sixel outputs). Press `q` in mpv to quit it and return to the browser. mpv
shows a short banner with the channel name and controls.

### built-in (in the TUI)

Renders the video with the app itself, in a modal with the channel list still
visible above and below it. It needs ffmpeg, which decodes the video (frames come
to the app) and plays the audio.

There are two renderers, chosen automatically:

- On terminals with the kitty graphics protocol (Ghostty, kitty, WezTerm) it
  draws a real image, so the video is sharp.
- Everywhere else it falls back to truecolor half-blocks, which is blocky (each
  character is two vertical pixels) but works in any terminal.

Force one with `-inline kitty` or `-inline halfblock`, or `IPTV_INLINE`. While it
plays:

- `f` toggles between the modal size and nearly-full.
- `esc` closes the player and returns to the browser.
- `q` quits the whole app.

The key rule is consistent with the rest of the app: `esc` goes back, `q` quits.
(mpv is the exception, because it uses its own keys, where `q` quits mpv.)

The video output is chosen for your terminal:

- kitty graphics on Ghostty, kitty, and WezTerm (sharp, pixel-based)
- sixel on terminals that support it (foot, and others)
- truecolor half-blocks (`tct`) everywhere else, which is coarse but works in any
  terminal

You can force the output:

```bash
iptv-tui -vo kitty      # or sixel, or tct
IPTV_TERM_VO=tct iptv-tui
```

### Performance

Inline playback uses hardware decoding where the platform provides it (VAAPI,
NVDEC, or Vulkan on Linux, VideoToolbox on macOS, D3D11VA on Windows), and falls
back to software decoding otherwise. The kitty and sixel outputs are fine at
normal sizes. The `tct` output is CPU-bound because it repaints character cells
every frame, so it renders at a reduced size and drops frames to stay responsive.

## Over SSH

Yes, inline playback can work over SSH, with two things to understand.

The stream is decoded on the remote machine (the one running iptv-tui), and the
resulting cells or images are streamed to your local terminal. So:

- Your local terminal decides what output works. If it is Ghostty or kitty, kitty
  graphics can render over SSH. Any truecolor terminal can render `tct`.
- Auto-detection runs on the remote side, where variables like `TERM_PROGRAM` and
  `KITTY_WINDOW_ID` usually are not forwarded. So it commonly falls back to `tct`,
  which is safe but coarse. If your local terminal supports kitty graphics and you
  are not inside tmux, force it: `iptv-tui -vo kitty`.
- Audio plays on the remote machine's audio device, not your local speakers.
  There is no audio forwarding built in. Over SSH you will see video locally and
  hear nothing locally unless you set up your own audio forwarding.
- Streaming frames over SSH uses bandwidth. `tct` is the lightest (text cells);
  kitty and sixel send more data.

## Under tmux (or herdr)

The picture varies by protocol:

- kitty graphics generally do not pass through tmux, so `-vo kitty` inside tmux
  usually shows nothing. Run outside tmux for kitty graphics.
- sixel passes through recent tmux (3.4+) with passthrough enabled; try
  `-vo sixel`.
- `tct` always works, since it is just colored text.

The palette itself is handled separately and does survive multiplexers: the app
forces the detected truecolor profile so herdr and tmux do not downgrade it.

## Practical recommendations

- Local Ghostty, kitty, or WezTerm: just press `t`. You get kitty graphics.
- Over SSH to a box with a GPU, local terminal is kitty-capable, no tmux:
  `iptv-tui -vo kitty` for video (audio stays on the remote host).
- Inside tmux, or an unknown terminal: `iptv-tui -vo tct`, or use the external
  window with enter instead.
