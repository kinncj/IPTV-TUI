package epg

import (
	"strings"
	"testing"
	"time"
)

const sample = `<?xml version="1.0"?>
<tv>
  <channel id="X"><display-name>Example TV (HD)</display-name></channel>
  <channel id="Y"><display-name>Other</display-name></channel>
  <programme start="20200101120000 +0000" stop="20200101130000 +0000" channel="X"><title>Now Show</title></programme>
  <programme start="20200101130000 +0000" stop="20200101140000 +0000" channel="X"><title>Next Show</title></programme>
  <programme start="20200101120000 +0000" stop="20200101130000 +0000" channel="Y"><title>Other</title></programme>
</tv>`

func TestParseAndNowNextByID(t *testing.T) {
	g, err := parse(strings.NewReader(sample), map[string]bool{"X": true}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(g.byChannel["Y"]) != 0 {
		t.Errorf("channel Y should be filtered out")
	}
	now := time.Date(2020, 1, 1, 12, 30, 0, 0, time.UTC)
	cur, next := g.NowNext("X", "", now)
	if cur == nil || cur.Title != "Now Show" {
		t.Errorf("current = %+v, want Now Show", cur)
	}
	if next == nil || next.Title != "Next Show" {
		t.Errorf("next = %+v, want Next Show", next)
	}
}

func TestNowNextByName(t *testing.T) {
	// Want by name; the guide id is "X" but we only know the channel name.
	g, err := parse(strings.NewReader(sample), nil, map[string]bool{normalize("Example TV"): true})
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2020, 1, 1, 12, 30, 0, 0, time.UTC)
	// Look up by a differently-formatted name; normalize should reconcile them.
	cur, _ := g.NowNext("no-such-id", "Example TV (1080p)", now)
	if cur == nil || cur.Title != "Now Show" {
		t.Errorf("name match failed: %+v", cur)
	}
}

func TestNormalize(t *testing.T) {
	cases := map[string]string{
		"SBT HD":            "sbt",
		"Globo (1080p)":     "globo",
		"Band News FHD":     "bandnews",
		"A&E Latin America": "aelatinamerica",
	}
	for in, want := range cases {
		if got := normalize(in); got != want {
			t.Errorf("normalize(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestSupported(t *testing.T) {
	if !Supported("Brazil", "") {
		t.Error("Brazil should be supported")
	}
	if Supported("Nowhereland", "") {
		t.Error("unknown country without custom URL should not be supported")
	}
	if !Supported("Nowhereland", "https://example.com/guide.xml") {
		t.Error("a custom URL makes any country supported")
	}
}

func TestNowNextNilGuide(t *testing.T) {
	var g *Guide
	if c, n := g.NowNext("X", "", time.Now()); c != nil || n != nil {
		t.Error("nil guide should return nil/nil")
	}
}
