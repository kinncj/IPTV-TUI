package epg

import (
	"strings"
	"testing"
	"time"
)

const sample = `<?xml version="1.0"?>
<tv>
  <channel id="X"><display-name>Ex</display-name></channel>
  <programme start="20200101120000 +0000" stop="20200101130000 +0000" channel="X"><title>Now Show</title></programme>
  <programme start="20200101130000 +0000" stop="20200101140000 +0000" channel="X"><title>Next Show</title></programme>
  <programme start="20200101120000 +0000" stop="20200101130000 +0000" channel="Y"><title>Other</title></programme>
</tv>`

func TestParseAndNowNext(t *testing.T) {
	g, err := parse(strings.NewReader(sample), map[string]bool{"X": true})
	if err != nil {
		t.Fatal(err)
	}
	// Y filtered out by wanted set.
	if len(g.byChannel["Y"]) != 0 {
		t.Errorf("channel Y should be filtered out")
	}

	now := time.Date(2020, 1, 1, 12, 30, 0, 0, time.UTC)
	cur, next := g.NowNext("X", now)
	if cur == nil || cur.Title != "Now Show" {
		t.Errorf("current = %+v, want Now Show", cur)
	}
	if next == nil || next.Title != "Next Show" {
		t.Errorf("next = %+v, want Next Show", next)
	}
}

func TestNowNextNilGuide(t *testing.T) {
	var g *Guide
	if c, n := g.NowNext("X", time.Now()); c != nil || n != nil {
		t.Error("nil guide should return nil/nil")
	}
}

func TestSupported(t *testing.T) {
	if !Supported("Brazil") {
		t.Error("Brazil should be supported")
	}
	if Supported("Nowhereland") {
		t.Error("unknown country should not be supported")
	}
}

func TestParseTime(t *testing.T) {
	got, err := parseTime("20200101120000 +0000")
	if err != nil {
		t.Fatal(err)
	}
	if !got.Equal(time.Date(2020, 1, 1, 12, 0, 0, 0, time.UTC)) {
		t.Errorf("parseTime = %v", got)
	}
}
