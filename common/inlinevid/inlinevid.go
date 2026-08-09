// Package inlinevid renders a video stream inside the terminal as truecolor
// half-block cells, so it can be composited into a running TUI (unlike mpv,
// which takes its own alternate screen). It shells out to ffmpeg for decoding:
// video frames come back as raw RGB on a pipe, and audio is played by a separate
// mpv process when available.
//
// The picture is cell-resolution (each character is two vertical pixels via the
// upper-half-block glyph), so it is blocky, like mpv's tct output. The point is
// that we own the frames and can draw them anywhere, including a modal over the
// live list.
package inlinevid

import (
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// Available reports whether ffmpeg is installed (required to decode).
func Available() bool { _, err := exec.LookPath("ffmpeg"); return err == nil }

func hasMPV() bool { _, err := exec.LookPath("mpv"); return err == nil }

func isHTTP(url string) bool {
	return strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")
}

// Stream is a running decode. Frame() blocks until the next frame is ready, so
// reading in a loop is naturally paced to the requested frame rate.
type Stream struct {
	cols, rows int
	w, h       int // pixel dimensions: w = cols, h = rows*2
	video      *exec.Cmd
	audio      *exec.Cmd
	out        io.ReadCloser
	buf        []byte
}

// Open starts decoding url scaled to fit cols x rows character cells at fps. It
// also starts audio playback via mpv when available (best effort).
func Open(url string, cols, rows, fps int) (*Stream, error) {
	if cols < 2 {
		cols = 2
	}
	if rows < 1 {
		rows = 1
	}
	if fps < 1 {
		fps = 15
	}
	w, h := cols, rows*2

	vf := fmt.Sprintf(
		"fps=%d,scale=%d:%d:force_original_aspect_ratio=decrease,pad=%d:%d:(ow-iw)/2:(oh-ih)/2:color=black",
		fps, w, h, w, h)
	args := []string{
		"-hide_banner", "-loglevel", "error", "-nostdin",
	}
	// -user_agent is an HTTP-protocol option; adding it for a non-HTTP input
	// (e.g. a local file) makes ffmpeg fail, so only set it for URLs.
	if isHTTP(url) {
		args = append(args, "-user_agent", "VLC/3.0 LibVLC/3.0")
	}
	args = append(args,
		"-i", url,
		"-an", "-map", "0:v:0",
		"-vf", vf,
		"-pix_fmt", "rgb24", "-f", "rawvideo", "pipe:1",
	)
	video := exec.Command("ffmpeg", args...)
	out, err := video.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := video.Start(); err != nil {
		return nil, err
	}

	s := &Stream{cols: cols, rows: rows, w: w, h: h, video: video, out: out, buf: make([]byte, w*h*3)}

	// Audio through mpv (no video). Best effort: a missing mpv or audio track
	// just means a silent picture.
	if hasMPV() {
		a := exec.Command("mpv", "--no-video", "--no-terminal", "--really-quiet",
			"--cache=yes", "--user-agent=VLC/3.0 LibVLC/3.0", url)
		if a.Start() == nil {
			s.audio = a
		}
	}
	return s, nil
}

// Frame reads the next frame and returns it as `rows` strings of half-block
// cells. It blocks until the frame arrives, and returns io.EOF when the stream
// ends.
func (s *Stream) Frame() ([]string, error) {
	if _, err := io.ReadFull(s.out, s.buf); err != nil {
		return nil, err
	}
	return renderHalfBlocks(s.buf, s.w, s.h, s.cols, s.rows), nil
}

// Close stops decoding and audio.
func (s *Stream) Close() {
	if s.video != nil && s.video.Process != nil {
		_ = s.video.Process.Kill()
		go func() { _ = s.video.Wait() }()
	}
	if s.audio != nil && s.audio.Process != nil {
		_ = s.audio.Process.Kill()
		go func() { _ = s.audio.Wait() }()
	}
}

// renderHalfBlocks turns an RGB frame (w x h pixels) into `rows` strings of
// `cols` cells. Each cell is the upper-half-block glyph with the foreground set
// to the top pixel and the background to the bottom pixel, packing two vertical
// pixels into one character row.
func renderHalfBlocks(buf []byte, w, h, cols, rows int) []string {
	out := make([]string, rows)
	var b strings.Builder
	for y := 0; y < rows; y++ {
		b.Reset()
		top := y * 2
		bot := y*2 + 1
		for x := 0; x < cols; x++ {
			ti := (top*w + x) * 3
			tr, tg, tb := buf[ti], buf[ti+1], buf[ti+2]
			var br, bg, bb byte
			if bot < h {
				bi := (bot*w + x) * 3
				br, bg, bb = buf[bi], buf[bi+1], buf[bi+2]
			}
			fmt.Fprintf(&b, "\x1b[38;2;%d;%d;%d;48;2;%d;%d;%dm▀", tr, tg, tb, br, bg, bb)
		}
		b.WriteString("\x1b[0m")
		out[y] = b.String()
	}
	return out
}
