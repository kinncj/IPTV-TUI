// Package export regenerates on-disk M3U playlists from an in-memory catalog.
// The upstream repos (iptv-org, Free-TV) are the source of truth; this package
// is how we rebuild local playlist files from them on demand.
package export

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/kinncj/iptv/common/catalog"
)

var unsafeName = regexp.MustCompile(`[/\\:*?"<>|]+`)

// Write regenerates playlists under dir:
//
//	dir/all.m3u              every channel, grouped by country
//	dir/countries/<Name>.m3u one file per group
//
// It returns the number of files written.
func Write(cat catalog.Catalog, dir string) (int, error) {
	countryDir := filepath.Join(dir, "countries")
	if err := os.MkdirAll(countryDir, 0o755); err != nil {
		return 0, err
	}

	files := 0

	var all strings.Builder
	all.WriteString("#EXTM3U\n")
	for _, g := range cat.Groups {
		var one strings.Builder
		one.WriteString("#EXTM3U\n")
		for _, ch := range g.Channels {
			entry := format(ch)
			all.WriteString(entry)
			one.WriteString(entry)
		}
		fn := filepath.Join(countryDir, sanitize(g.Name)+".m3u")
		if err := os.WriteFile(fn, []byte(one.String()), 0o644); err != nil {
			return files, err
		}
		files++
	}

	if err := os.WriteFile(filepath.Join(dir, "all.m3u"), []byte(all.String()), 0o644); err != nil {
		return files, err
	}
	files++
	return files, nil
}

func format(ch catalog.Channel) string {
	return fmt.Sprintf("#EXTINF:-1 tvg-id=%q tvg-logo=%q group-title=%q,%s\n%s\n",
		ch.TvgID, ch.Logo, ch.Group, ch.Name, ch.URL)
}

func sanitize(name string) string {
	name = unsafeName.ReplaceAllString(name, "-")
	name = strings.Trim(strings.TrimSpace(name), ".-")
	if name == "" {
		return "Uncategorized"
	}
	return name
}
