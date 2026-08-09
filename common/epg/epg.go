// Package epg fetches and parses XMLTV program guides (from epgshare01) and
// answers now/next queries by channel id. It is optional and best-effort: any
// failure yields no guide rather than an error the user must act on.
package epg

import (
	"compress/gzip"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

const base = "https://epgshare01.online/epgshare01/"

// maxBytes skips guides too large to fetch responsively (e.g. the ~59MB US
// file); those countries simply show no EPG rather than freezing the UI.
const maxBytes = 40 << 20

// ErrTooLarge means the guide exceeded maxBytes and was skipped.
var ErrTooLarge = errors.New("epg guide too large")

// countryFile maps an iptv-org country (group) name to its epgshare01 filename.
// Extend freely; unknown countries just have no guide.
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

// Supported reports whether a country has a known guide file.
func Supported(country string) bool { _, ok := countryFile[country]; return ok }

// Programme is a single scheduled show.
type Programme struct {
	Start, Stop time.Time
	Title       string
}

// Guide indexes programmes by channel id.
type Guide struct {
	byChannel map[string][]Programme
}

// NowNext returns the currently-airing and next programmes for channelID.
func (g *Guide) NowNext(channelID string, now time.Time) (current, next *Programme) {
	if g == nil {
		return nil, nil
	}
	ps := g.byChannel[channelID]
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

// Load fetches and parses the guide for country, keeping only programmes whose
// channel id is in `wanted` (bounding memory to the current country's channels).
func Load(ctx context.Context, country string, wanted map[string]bool) (*Guide, error) {
	file, ok := countryFile[country]
	if !ok {
		return nil, nil // no guide for this country; not an error
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+file, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("epg %s: status %d", file, resp.StatusCode)
	}
	if resp.ContentLength > maxBytes {
		return nil, ErrTooLarge
	}

	gz, err := gzip.NewReader(io.LimitReader(resp.Body, maxBytes))
	if err != nil {
		return nil, err
	}
	defer gz.Close()

	return parse(gz, wanted)
}

// parse streams an XMLTV document, collecting programmes for wanted channels.
func parse(r io.Reader, wanted map[string]bool) (*Guide, error) {
	type xmlProgramme struct {
		Start   string `xml:"start,attr"`
		Stop    string `xml:"stop,attr"`
		Channel string `xml:"channel,attr"`
		Title   string `xml:"title"`
	}

	g := &Guide{byChannel: map[string][]Programme{}}
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
		if !ok || se.Name.Local != "programme" {
			continue
		}
		var xp xmlProgramme
		if err := dec.DecodeElement(&xp, &se); err != nil {
			continue
		}
		if len(wanted) > 0 && !wanted[xp.Channel] {
			continue
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
	return g, nil
}

// parseTime parses XMLTV timestamps like "20240115120000 +0000".
func parseTime(s string) (time.Time, error) {
	if len(s) >= 20 {
		return time.Parse("20060102150405 -0700", s)
	}
	return time.Parse("20060102150405", s)
}
