// Package resolve turns page URLs that players can't open directly (YouTube
// links, in particular) into a direct stream URL using yt-dlp. Free-TV lists a
// number of channels as YouTube live URLs; ffplay, ffmpeg, and the built-in
// player cannot open those, so they are resolved first.
package resolve

import (
	"context"
	"errors"
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

// Direct returns a directly-playable stream URL for url. When url does not need
// resolving it is returned unchanged with a nil error. When it does need
// resolving and that fails, the original url is returned together with an error
// describing why, so the caller can both fall back and tell the user.
func Direct(url string) (string, error) {
	if !Needed(url) {
		return url, nil
	}
	bin := tool()
	if bin == "" {
		return url, errors.New("install yt-dlp to play YouTube channels")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	// -f b selects a single muxed stream (audio+video in one URL), which the
	// single-URL players can open; -g prints the resolved URL. For live channels
	// this is the HLS manifest. --geo-bypass spoofs the country via
	// X-Forwarded-For, which frees some geo-blocked streams with no VPN; it does
	// nothing for stricter blocks, but never hurts.
	cmd := exec.CommandContext(ctx, bin, "--no-warnings", "--geo-bypass", "-f", "b", "-g", url)
	out, err := cmd.Output()
	if err != nil {
		return url, resolveError(err)
	}
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if s := strings.TrimSpace(line); s != "" {
			return s, nil
		}
	}
	return url, errors.New("resolver returned no stream URL")
}

// resolveError extracts a short reason from yt-dlp's stderr when available.
func resolveError(err error) error {
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		for _, line := range strings.Split(string(ee.Stderr), "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "ERROR:") {
				return errors.New(strings.TrimSpace(strings.TrimPrefix(line, "ERROR:")))
			}
		}
	}
	return err
}
