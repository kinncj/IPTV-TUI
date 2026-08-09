// Package features holds acceptance tests that exercise packages together
// through realistic flows, without the network or a live terminal.
package features

import (
	"testing"

	"github.com/kinncj/iptv/common/catalog"
	"github.com/kinncj/iptv/common/m3u"
	"github.com/kinncj/iptv/common/probe"
)

// A miniature of the real upstream shape: two sources, overlapping countries,
// VLC options between EXTINF and URL, and a groupless entry.
const iptvOrgSample = `#EXTM3U
#EXTINF:-1 tvg-id="Globo.br" group-title="Brazil",Globo
http://a/globo.m3u8
#EXTINF:-1 tvg-id="RecordTV.br" group-title="Brazil",Record
http://a/record.m3u8
#EXTINF:-1 group-title="Argentina",TelefeAR
http://a/telefe.m3u8
`

const freeTVSample = `#EXTM3U
#EXTINF:-1 group-title="Brazil",SBT
#EXTVLCOPT:http-user-agent=Mozilla
http://b/sbt.m3u8
#EXTINF:-1 group-title="Japan",NHK World
http://b/nhk.m3u8
`

func TestEndToEndCatalogBuild(t *testing.T) {
	var channels []catalog.Channel
	channels = append(channels, m3u.Parse([]byte(iptvOrgSample), "iptv-org")...)
	channels = append(channels, m3u.Parse([]byte(freeTVSample), "free-tv")...)

	cat := catalog.Build(channels)

	if cat.Total != 5 {
		t.Fatalf("total channels = %d, want 5", cat.Total)
	}

	want := map[string]int{"Argentina": 1, "Brazil": 3, "Japan": 1}
	got := map[string]int{}
	for _, g := range cat.Groups {
		got[g.Name] = len(g.Channels)
	}
	for name, n := range want {
		if got[name] != n {
			t.Errorf("group %q = %d channels, want %d", name, got[name], n)
		}
	}

	// Groups must be alphabetically ordered for stable navigation.
	if cat.Groups[0].Name != "Argentina" {
		t.Errorf("first group = %q, want Argentina", cat.Groups[0].Name)
	}

	// Provenance survives the merge.
	brazil := findGroup(cat, "Brazil")
	if !hasSource(brazil, "free-tv") || !hasSource(brazil, "iptv-org") {
		t.Errorf("Brazil should contain channels from both sources: %+v", brazil.Channels)
	}
}

func TestProbeStatusIcons(t *testing.T) {
	cases := map[probe.Status]string{
		probe.OK:      "✓",
		probe.Blocked: "⚠",
		probe.Dead:    "✗",
	}
	for st, icon := range cases {
		if st.Icon() != icon {
			t.Errorf("status %d icon = %q, want %q", st, st.Icon(), icon)
		}
	}
}

func findGroup(c catalog.Catalog, name string) catalog.Group {
	for _, g := range c.Groups {
		if g.Name == name {
			return g
		}
	}
	return catalog.Group{}
}

func hasSource(g catalog.Group, src string) bool {
	for _, ch := range g.Channels {
		if ch.Source == src {
			return true
		}
	}
	return false
}
