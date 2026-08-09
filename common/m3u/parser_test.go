package m3u

import "testing"

const sample = `#EXTM3U
#EXTINF:-1 tvg-id="Globo.br" tvg-logo="http://x/l.png" group-title="Brazil",Globo HD
http://example.com/globo.m3u8
#EXTINF:-1 group-title="Brazil",SBT
#EXTVLCOPT:http-user-agent=Mozilla
http://example.com/sbt.m3u8
#EXTINF:-1,No Group Channel
http://example.com/ng.m3u8
`

func TestParse(t *testing.T) {
	chs := Parse([]byte(sample), "test")
	if len(chs) != 3 {
		t.Fatalf("want 3 channels, got %d", len(chs))
	}

	if chs[0].Name != "Globo HD" || chs[0].Group != "Brazil" || chs[0].TvgID != "Globo.br" {
		t.Errorf("channel 0 parsed wrong: %+v", chs[0])
	}
	if chs[0].URL != "http://example.com/globo.m3u8" {
		t.Errorf("channel 0 url wrong: %q", chs[0].URL)
	}
	if chs[0].Source != "test" {
		t.Errorf("source not stamped: %q", chs[0].Source)
	}

	// #EXTVLCOPT between EXTINF and URL must not break association.
	if chs[1].Name != "SBT" || chs[1].URL != "http://example.com/sbt.m3u8" {
		t.Errorf("channel 1 parsed wrong: %+v", chs[1])
	}

	if chs[2].Group != "" {
		t.Errorf("channel 2 should have empty group, got %q", chs[2].Group)
	}
}
