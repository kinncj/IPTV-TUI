// Package player launches an external media player on a stream URL. The Player
// interface lets the TUI stay ignorant of which binary plays the stream;
// concrete players (mpv, vlc, ffplay) are selected by availability.
package player

import (
	"fmt"
	"os/exec"
)

// Player plays a stream URL in a detached process.
type Player interface {
	Name() string
	Available() bool
	// Play starts playback and returns without waiting for the player to exit.
	Play(url string) error
}

// Detect returns the available players in preference order (mpv, vlc, ffplay).
// mpv is first: it handles HLS reconnection and codecs better than VLC for the
// flaky streams these playlists contain.
func Detect() []Player {
	all := []Player{mpv{}, vlc{}, ffplay{}}
	var out []Player
	for _, p := range all {
		if p.Available() {
			out = append(out, p)
		}
	}
	return out
}

// ByName returns the named player if installed.
func ByName(name string) (Player, bool) {
	for _, p := range Detect() {
		if p.Name() == name {
			return p, true
		}
	}
	return nil, false
}

// HasMPV reports whether mpv is installed (required for inline terminal playback).
func HasMPV() bool { return lookPath("mpv") }

// TerminalCmd builds an mpv command that renders video inline in the current
// terminal using the given video output (kitty|sixel|tct). The command is NOT
// started — the caller hands the terminal to it (e.g. tea.ExecProcess) and mpv
// runs in the foreground until the user quits it.
//
// Performance matters here: the kitty/sixel outputs are fine with hardware
// decoding, but the character-cell output (tct) is CPU-bound, so we cap its
// resolution and let mpv drop frames to stay responsive.
func TerminalCmd(url, vo string) *exec.Cmd {
	if vo == "" {
		vo = "tct"
	}
	args := []string{
		"--vo=" + vo,
		"--hwdec=auto-safe",          // auto-pick a safe HW decoder per platform (vaapi/nvdec/vulkan/videotoolbox/d3d11va); falls back to software
		"--profile=fast",             // low-latency decode/render path
		"--cache=yes",                // buffer network streams to avoid stalls
		"--demuxer-readahead-secs=3", //
		"--framedrop=vo",             // drop frames rather than lag behind
		"--really-quiet",
		"--user-agent=VLC/3.0 LibVLC/3.0",
	}
	switch vo {
	case "kitty":
		// Shared memory transfer is much faster than escape-code streaming.
		args = append(args, "--vo-kitty-use-shm=yes")
	case "tct", "sixel":
		// Character/sixel cells are expensive per frame — render smaller and let
		// the terminal upscale, and don't chase 60fps.
		args = append(args, "--vf=scale=640:-2", "--hwdec=no")
	}
	args = append(args, url)
	return exec.Command("mpv", args...)
}

func lookPath(bin string) bool {
	_, err := exec.LookPath(bin)
	return err == nil
}

func start(bin string, args ...string) error {
	cmd := exec.Command(bin, args...)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("launch %s: %w", bin, err)
	}
	// Reap the process asynchronously so it doesn't become a zombie; we don't
	// block the UI on player lifetime.
	go func() { _ = cmd.Wait() }()
	return nil
}

type mpv struct{}

func (mpv) Name() string    { return "mpv" }
func (mpv) Available() bool { return lookPath("mpv") }
func (mpv) Play(url string) error {
	return start("mpv",
		"--force-window=immediate",
		"--keep-open=no",
		"--user-agent=VLC/3.0 LibVLC/3.0",
		url,
	)
}

type vlc struct{}

func (vlc) Name() string    { return "vlc" }
func (vlc) Available() bool { return lookPath("vlc") }
func (vlc) Play(url string) error {
	return start("vlc", "--play-and-exit", url)
}

type ffplay struct{}

func (ffplay) Name() string    { return "ffplay" }
func (ffplay) Available() bool { return lookPath("ffplay") }
func (ffplay) Play(url string) error {
	return start("ffplay", "-autoexit", "-loglevel", "warning", url)
}
