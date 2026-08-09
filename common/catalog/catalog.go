// Package catalog holds the domain model for IPTV content. It has no I/O and no
// dependency on parsing, networking, or the UI — those live in other packages
// and depend inward on this one.
package catalog

import "sort"

// Channel is a single playable stream.
type Channel struct {
	Name     string
	URL      string
	Group    string // country or category, from group-title
	TvgID    string
	Logo     string
	Category string // set by the iptv-org API ingestion; empty for plain M3U
	Source   string // which upstream playlist it came from
}

// Group is a named set of channels (a country, in practice).
type Group struct {
	Name     string
	Channels []Channel
}

// Catalog is the full browsable set, organized into groups.
type Catalog struct {
	Groups []Group
	Total  int
}

// Build groups channels by their Group field (country) and returns a Catalog
// with groups and channels sorted alphabetically. Channels with an empty group
// are bucketed under "Uncategorized".
func Build(channels []Channel) Catalog {
	return buildBy(channels, func(c Channel) string { return c.Group })
}

// BuildByCategory groups the same channels by their Category instead of country.
// Channels with no category (e.g. plain M3U) fall under "Uncategorized".
func BuildByCategory(channels []Channel) Catalog {
	return buildBy(channels, func(c Channel) string { return c.Category })
}

// Flatten returns every channel across the catalog's groups, so it can be
// re-grouped a different way.
func (c Catalog) Flatten() []Channel {
	out := make([]Channel, 0, c.Total)
	for _, g := range c.Groups {
		out = append(out, g.Channels...)
	}
	return out
}

func buildBy(channels []Channel, key func(Channel) string) Catalog {
	byGroup := make(map[string][]Channel)
	for _, ch := range channels {
		g := key(ch)
		if g == "" {
			g = "Uncategorized"
		}
		byGroup[g] = append(byGroup[g], ch)
	}

	groups := make([]Group, 0, len(byGroup))
	for name, chs := range byGroup {
		sort.Slice(chs, func(i, j int) bool { return chs[i].Name < chs[j].Name })
		groups = append(groups, Group{Name: name, Channels: chs})
	}
	sort.Slice(groups, func(i, j int) bool { return groups[i].Name < groups[j].Name })

	return Catalog{Groups: groups, Total: len(channels)}
}
