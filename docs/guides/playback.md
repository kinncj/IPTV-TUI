# Playing channels

There are two ways to play a channel.

## In an external window

Press enter on a channel. The app launches your player (mpv by default, then vlc,
then ffplay) in its own window and keeps running. Press `p` to cycle which player
is used; the footer shows the current one.

## Inline, in the terminal

Press `t`. The video renders inside the terminal itself, no separate window. This
needs mpv. It opens as a centered modal that takes most of the terminal but not
all of it. Press `f` while it plays to toggle between fullscreen and the modal
size, and `q` to quit and return to the browser. mpv shows a short banner at the
top with the channel name and these controls.

A note on what "inline" means here. A terminal video player takes its own
alternate screen to paint frames, the same way `less` or `vim` do. So while a
channel plays, mpv owns the screen: the browser is not drawn behind the video,
and the video cannot be framed by the app or overlaid on the live list. The modal
is mpv rendering the video at 78% of the terminal, centered, on its own screen.
`f` cycles that geometry between 78% and fullscreen. The resize is live on the
kitty and sixel outputs, which redraw each frame; on the plain `tct` output the
change may only apply cleanly on some terminals. When you quit mpv, the browser
returns exactly where you left it.

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
