// Package iptvorg ingests channels from the iptv-org JSON API (channels +
// streams + countries), joining them into enriched catalog channels with stable
// IDs, proper country names, and categories.
//
// This is an OPT-IN alternative to the M3U playlist for the built-in iptv-org
// source only. It never touches user-added playlists — those always flow through
// the M3U parser and are merged separately — so enabling it cannot break a
// user's own lists.
package iptvorg

import (
	"encoding/json"
	"sort"
	"strings"

	"github.com/kinncj/IPTV-TUI/common/catalog"
	"github.com/kinncj/IPTV-TUI/common/source"
)

// Categories fetches just the channel-to-category mapping from the API's
// channels.json (channel id to its primary category). It enriches M3U channels
// with categories without changing their country grouping.
func Categories(loader *source.Loader, refresh bool) (map[string]string, error) {
	b, err := loader.Load(source.Source{Name: "iptvorg-channels", URL: channelsURL}, refresh)
	if err != nil {
		return nil, err
	}
	var channels []apiChannel
	if err := json.Unmarshal(b, &channels); err != nil {
		return nil, err
	}
	m := make(map[string]string, len(channels))
	for _, c := range channels {
		if len(c.Categories) > 0 {
			m[c.ID] = c.Categories[0]
		}
	}
	return m, nil
}

// CategoryFor returns the category for an M3U tvg-id by stripping the feed
// suffix ("Globo.br@SD" -> "Globo.br") and looking it up.
func CategoryFor(catmap map[string]string, tvgID string) string {
	if tvgID == "" || catmap == nil {
		return ""
	}
	id := tvgID
	if i := strings.IndexByte(id, '@'); i >= 0 {
		id = id[:i]
	}
	return catmap[id]
}

const (
	channelsURL  = "https://iptv-org.github.io/api/channels.json"
	streamsURL   = "https://iptv-org.github.io/api/streams.json"
	countriesURL = "https://iptv-org.github.io/api/countries.json"
)

type apiChannel struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Country    string   `json:"country"`
	Categories []string `json:"categories"`
}

type apiStream struct {
	Channel *string `json:"channel"`
	Title   string  `json:"title"`
	URL     string  `json:"url"`
	Quality string  `json:"quality"`
}

type apiCountry struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

// Fetch pulls the API JSON (via the caching loader) and returns enriched
// channels grouped by country. Streams without a known channel are kept under
// "Undefined" so nothing is silently dropped.
func Fetch(loader *source.Loader, refresh bool) ([]catalog.Channel, error) {
	chBytes, err := loader.Load(source.Source{Name: "iptvorg-channels", URL: channelsURL}, refresh)
	if err != nil {
		return nil, err
	}
	stBytes, err := loader.Load(source.Source{Name: "iptvorg-streams", URL: streamsURL}, refresh)
	if err != nil {
		return nil, err
	}
	coBytes, _ := loader.Load(source.Source{Name: "iptvorg-countries", URL: countriesURL}, refresh)

	var channels []apiChannel
	if err := json.Unmarshal(chBytes, &channels); err != nil {
		return nil, err
	}
	var streams []apiStream
	if err := json.Unmarshal(stBytes, &streams); err != nil {
		return nil, err
	}
	var countries []apiCountry
	_ = json.Unmarshal(coBytes, &countries)

	return join(channels, streams, countries), nil
}

// join merges channel metadata, streams, and country names into catalog
// channels. Pure — no I/O — so it is unit-testable.
func join(channels []apiChannel, streams []apiStream, countries []apiCountry) []catalog.Channel {
	countryName := make(map[string]string, len(countries))
	for _, c := range countries {
		countryName[c.Code] = c.Name
	}
	meta := make(map[string]apiChannel, len(channels))
	for _, c := range channels {
		meta[c.ID] = c
	}

	out := make([]catalog.Channel, 0, len(streams))
	for _, s := range streams {
		if s.URL == "" {
			continue
		}
		ch := catalog.Channel{URL: s.URL, Source: "iptv-org", Group: "Undefined"}
		if s.Channel != nil {
			if m, ok := meta[*s.Channel]; ok {
				ch.Name = m.Name
				ch.TvgID = m.ID
				if n, ok := countryName[m.Country]; ok && n != "" {
					ch.Group = n
				}
				if len(m.Categories) > 0 {
					ch.Category = m.Categories[0]
				}
			}
		}
		if ch.Name == "" {
			ch.Name = firstNonEmpty(s.Title, s.URL)
		}
		if s.Quality != "" {
			ch.Name += " (" + s.Quality + ")"
		}
		out = append(out, ch)
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].Group != out[j].Group {
			return out[i].Group < out[j].Group
		}
		return out[i].Name < out[j].Name
	})
	return out
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
