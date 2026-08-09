package iptvorg

import "testing"

func ptr(s string) *string { return &s }

func TestJoinEnrichesAndGroups(t *testing.T) {
	channels := []apiChannel{
		{ID: "Globo.br", Name: "Globo", Country: "BR", Categories: []string{"general", "news"}},
		{ID: "NHK.jp", Name: "NHK", Country: "JP", Categories: []string{"news"}},
	}
	streams := []apiStream{
		{Channel: ptr("Globo.br"), URL: "http://a/globo.m3u8", Quality: "1080p"},
		{Channel: ptr("NHK.jp"), URL: "http://a/nhk.m3u8"},
		{Channel: nil, Title: "Mystery Stream", URL: "http://a/mystery.m3u8"}, // no channel
		{Channel: ptr("Globo.br"), URL: ""},                                   // empty URL dropped
	}
	countries := []apiCountry{{Name: "Brazil", Code: "BR"}, {Name: "Japan", Code: "JP"}}

	got := join(channels, streams, countries)

	if len(got) != 3 {
		t.Fatalf("want 3 channels (empty-URL dropped), got %d", len(got))
	}

	// Enriched: stable id, country name group, category, quality in name.
	byURL := map[string]int{}
	for i, c := range got {
		byURL[c.URL] = i
	}
	g := got[byURL["http://a/globo.m3u8"]]
	if g.Name != "Globo (1080p)" || g.Group != "Brazil" || g.TvgID != "Globo.br" || g.Category != "general" {
		t.Errorf("globo enrichment wrong: %+v", g)
	}

	// Null-channel stream kept under Undefined with its title.
	m := got[byURL["http://a/mystery.m3u8"]]
	if m.Group != "Undefined" || m.Name != "Mystery Stream" {
		t.Errorf("null-channel stream wrong: %+v", m)
	}

	// Sorted by group then name: Brazil < Japan < Undefined.
	if got[0].Group != "Brazil" || got[len(got)-1].Group != "Undefined" {
		t.Errorf("not grouped/sorted: first=%q last=%q", got[0].Group, got[len(got)-1].Group)
	}
}

func TestJoinUnknownChannelFallsBack(t *testing.T) {
	// Stream references a channel id we don't have metadata for.
	got := join(nil, []apiStream{{Channel: ptr("missing.id"), Title: "Fallback", URL: "http://x"}}, nil)
	if len(got) != 1 || got[0].Name != "Fallback" || got[0].Group != "Undefined" {
		t.Errorf("unknown channel should fall back to title/Undefined: %+v", got)
	}
}
