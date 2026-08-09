// Package resolve turns page URLs that players can't open directly (YouTube
// links, in particular) into a direct stream URL using yt-dlp. Free-TV lists a
// number of channels as YouTube live URLs; ffplay, ffmpeg, and the built-in
// player cannot open those, so they are resolved first.
package resolve

import (
	"context"
	"os/exec"
	"strings"
	"time"
)

// Needed reports whether url is a page URL that must be resolved to a stream.
func Needed(url string) bool {
	return strings.Contains(url, "youtube.com") ||
		strings.Contains(url, "youtu.be") ||
		strings.Contains(url, "dailymotion.com")
}

// Available reports whether a resolver (yt-dlp, or youtube-dl) is installed.
func Available() bool { return tool() != "" }

func tool() string {
	for _, t := range []string{"yt-dlp", "youtube-dl"} {
		if _, err := exec.LookPath(t); err == nil {
			return t
		}
	}
	return ""
}

// Direct returns a directly-playable stream URL for url. If url does not need
// resolving, or no resolver is installed, or resolution fails, the original url
// is returned unchanged, so callers can always use the result.
func Direct(url string) string {
	if !Needed(url) {
		return url
	}
	bin := tool()
	if bin == "" {
		return url
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// -f b selects a single muxed stream (audio+video in one URL), which the
	// single-URL players can open; -g prints the resolved URL. For live channels
	// this is the HLS manifest.
	out, err := exec.CommandContext(ctx, bin, "-q", "--no-warnings", "-f", "b", "-g", url).Output()
	if err != nil {
		return url
	}
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if s := strings.TrimSpace(line); s != "" {
			return s
		}
	}
	return url
}
