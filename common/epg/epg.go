// Package epg fetches and parses XMLTV program guides and answers now/next
// queries. It matches channels by their guide id and, failing that, by a
// normalized channel name, so it works with guides that use different id
// schemes. It is optional and best-effort: any failure yields no guide rather
// than an error the user must act on.
//
// Guide source: a country's file from epgshare01 by default, or a single custom
// XMLTV URL (config "epg_url"), which is the way to point at a self-hosted
// iptv-org/epg guide whose ids line up with the playlists.
package epg

import (
	"compress/gzip"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const base = "https://epgshare01.online/epgshare01/"

// maxBytes skips guides too large to fetch responsively.
const maxBytes = 60 << 20

// ErrTooLarge means the guide exceeded maxBytes and was skipped.
var ErrTooLarge = errors.New("epg guide too large")

// countryFile maps an iptv-org country (group) name to its epgshare01 filename.
var countryFile = map[string]string{
	"United Kingdom": "epg_ripper_UK1.xml.gz",
	"Brazil":         "epg_ripper_BR1.xml.gz",
	"Germany":        "epg_ripper_DE1.xml.gz",
	"Italy":          "epg_ripper_IT1.xml.gz",
	"Spain":          "epg_ripper_ES1.xml.gz",
	"France":         "epg_ripper_FR1.xml.gz",
	"Canada":         "epg_ripper_CA2.xml.gz",
	"Australia":      "epg_ripper_AU1.xml.gz",
	"Portugal":       "epg_ripper_PT1.xml.gz",
	"Netherlands":    "epg_ripper_NL1.xml.gz",
	"Poland":         "epg_ripper_PL1.xml.gz",
	"Turkey":         "epg_ripper_TR1.xml.gz",
	"Mexico":         "epg_ripper_MX1.xml.gz",
	"Argentina":      "epg_ripper_AR1.xml.gz",
}

// Supported reports whether a guide can be loaded for country (a per-country
// file exists, or a custom URL is provided).
func Supported(country, customURL string) bool {
	if customURL != "" {
		return true
	}
	_, ok := countryFile[country]
	return ok
}

// Programme is a single scheduled show.
type Programme struct {
	Start, Stop time.Time
	Title       string
}

// Guide indexes programmes by channel id and maps normalized channel names to
// ids, so a lookup can fall back from id to name.
type Guide struct {
	byChannel map[string][]Programme
	nameToID  map[string]string
}

// NowNext returns the currently-airing and next programmes for a channel,
// matched by id first and by normalized name second.
func (g *Guide) NowNext(id, name string, now time.Time) (current, next *Programme) {
	if g == nil {
		return nil, nil
	}
	ps := g.byChannel[id]
	if len(ps) == 0 && name != "" {
		key := normalize(name)
		if cid, ok := g.nameToID[key]; ok {
			ps = g.byChannel[cid]
		} else if len(key) >= 3 {
			// Fall back to a suffix match, which handles guides that prefix the
			// name (epgshare01 uses "<city>/<uf> <name>") without the false hits
			// a plain substring match would cause. Prefer the shortest fit.
			bestLen := 1 << 30
			for gname, cid := range g.nameToID {
				if strings.HasSuffix(gname, key) && len(gname) < bestLen {
					ps = g.byChannel[cid]
					bestLen = len(gname)
				}
			}
		}
	}
	for i := range ps {
		p := &ps[i]
		switch {
		case !p.Start.After(now) && p.Stop.After(now):
			current = p
		case p.Start.After(now):
			if next == nil || p.Start.Before(next.Start) {
				next = p
			}
		}
	}
	return current, next
}

// Load fetches and parses a guide for country (or customURL when set), keeping
// programmes whose channel matches a wanted id or a wanted normalized name.
func Load(ctx context.Context, country, customURL string, wantIDs, wantNames map[string]bool) (*Guide, error) {
	url := customURL
	if url == "" {
		file, ok := countryFile[country]
		if !ok {
			return nil, nil // no guide for this country; not an error
		}
		url = base + file
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("epg: status %d", resp.StatusCode)
	}
	if resp.ContentLength > maxBytes {
		return nil, ErrTooLarge
	}

	var r io.Reader = io.LimitReader(resp.Body, maxBytes)
	if strings.HasSuffix(url, ".gz") {
		gz, err := gzip.NewReader(r)
		if err != nil {
			return nil, err
		}
		defer gz.Close()
		r = gz
	}
	return parse(r, wantIDs, wantNames)
}

// parse streams an XMLTV document. It reads channel display-names first (they
// precede programmes), then keeps programmes for channels matching a wanted id
// or name.
func parse(r io.Reader, wantIDs, wantNames map[string]bool) (*Guide, error) {
	type xmlChannel struct {
		ID    string   `xml:"id,attr"`
		Names []string `xml:"display-name"`
	}
	type xmlProgramme struct {
		Start   string `xml:"start,attr"`
		Stop    string `xml:"stop,attr"`
		Channel string `xml:"channel,attr"`
		Title   string `xml:"title"`
	}

	g := &Guide{byChannel: map[string][]Programme{}, nameToID: map[string]string{}}
	keep := map[string]bool{} // channel ids whose programmes we retain

	dec := xml.NewDecoder(r)
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return g, err
		}
		se, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		switch se.Name.Local {
		case "channel":
			var c xmlChannel
			if dec.DecodeElement(&c, &se) != nil {
				continue
			}
			wanted := wantIDs[c.ID]
			for _, n := range c.Names {
				nn := normalize(n)
				if nn == "" {
					continue
				}
				if _, exists := g.nameToID[nn]; !exists {
					g.nameToID[nn] = c.ID
				}
				if wantNames[nn] {
					wanted = true
				}
				// Suffix match: a guide name that ends with a wanted name (the
				// epgshare01 "<city>/<uf> <name>" case).
				if !wanted {
					for wk := range wantNames {
						if len(wk) >= 3 && strings.HasSuffix(nn, wk) {
							wanted = true
							break
						}
					}
				}
			}
			if wanted {
				keep[c.ID] = true
			}
		case "programme":
			var xp xmlProgramme
			if dec.DecodeElement(&xp, &se) != nil {
				continue
			}
			if len(wantIDs) > 0 || len(wantNames) > 0 {
				if !keep[xp.Channel] && !wantIDs[xp.Channel] {
					continue
				}
			}
			start, err1 := parseTime(xp.Start)
			stop, err2 := parseTime(xp.Stop)
			if err1 != nil || err2 != nil {
				continue
			}
			g.byChannel[xp.Channel] = append(g.byChannel[xp.Channel], Programme{
				Start: start, Stop: stop, Title: xp.Title,
			})
		}
	}
	return g, nil
}

var (
	parenRe  = regexp.MustCompile(`[\(\[].*?[\)\]]`)
	nonAlnum = regexp.MustCompile(`[^a-z0-9]+`)
	qualRe   = regexp.MustCompile(`\b(hd|fhd|uhd|sd|4k|8k|hevc|h265|h264)\b`)
)

// NameKey is the exported normalized name key, so callers build the same keys
// used for matching.
func NameKey(s string) string { return normalize(s) }

// normalize reduces a channel name to a comparable key: lowercase, without
// parenthetical notes, quality tags, or punctuation.
func normalize(s string) string {
	s = strings.ToLower(s)
	s = parenRe.ReplaceAllString(s, " ")
	s = qualRe.ReplaceAllString(s, " ")
	s = nonAlnum.ReplaceAllString(s, "")
	return s
}

// parseTime parses XMLTV timestamps like "20240115120000 +0000".
func parseTime(s string) (time.Time, error) {
	if len(s) >= 20 {
		return time.Parse("20060102150405 -0700", s)
	}
	return time.Parse("20060102150405", s)
}
