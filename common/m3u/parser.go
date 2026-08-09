// Package m3u parses extended M3U/M3U8 playlists into catalog channels. It is a
// pure function of its input bytes — no networking, no filesystem.
package m3u

import (
	"bufio"
	"bytes"
	"regexp"
	"strings"

	"github.com/kinncj/iptv/common/catalog"
)

var attrRe = regexp.MustCompile(`([a-zA-Z0-9_-]+)="([^"]*)"`)

// Parse reads an extended M3U playlist and returns its channels. Malformed
// entries are skipped rather than failing the whole parse. The source label is
// stamped onto every channel for provenance.
func Parse(data []byte, source string) []catalog.Channel {
	var channels []catalog.Channel
	sc := bufio.NewScanner(bytes.NewReader(data))
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var pending *catalog.Channel
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		switch {
		case line == "" || strings.HasPrefix(line, "#EXTM3U"):
			continue
		case strings.HasPrefix(line, "#EXTINF"):
			ch := parseExtinf(line)
			ch.Source = source
			pending = &ch
		case strings.HasPrefix(line, "#"):
			// #EXTVLCOPT and friends — ignored for v1.
			continue
		default:
			if pending != nil {
				pending.URL = line
				if pending.Name == "" {
					pending.Name = line
				}
				channels = append(channels, *pending)
				pending = nil
			}
		}
	}
	return channels
}

func parseExtinf(line string) catalog.Channel {
	var ch catalog.Channel

	// Attributes (tvg-id, tvg-logo, group-title, ...) live before the trailing
	// comma; the display name is everything after the last comma.
	if i := strings.LastIndex(line, ","); i >= 0 {
		ch.Name = strings.TrimSpace(line[i+1:])
	}
	for _, m := range attrRe.FindAllStringSubmatch(line, -1) {
		switch m[1] {
		case "group-title":
			ch.Group = m[2]
		case "tvg-id":
			ch.TvgID = m[2]
		case "tvg-logo":
			ch.Logo = m[2]
		}
	}
	return ch
}
