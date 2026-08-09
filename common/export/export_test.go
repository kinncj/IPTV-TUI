package export

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kinncj/IPTV-TUI/common/catalog"
)

func TestWriteRegeneratesPlaylists(t *testing.T) {
	cat := catalog.Build([]catalog.Channel{
		{Name: "Globo", URL: "http://a/g.m3u8", Group: "Brazil", TvgID: "Globo.br"},
		{Name: "SBT", URL: "http://a/s.m3u8", Group: "Brazil"},
		{Name: "Telefe", URL: "http://a/t.m3u8", Group: "Argentina"},
	})

	dir := t.TempDir()
	n, err := Write(cat, dir)
	if err != nil {
		t.Fatal(err)
	}
	// all.m3u + 2 country files.
	if n != 3 {
		t.Fatalf("wrote %d files, want 3", n)
	}

	all, err := os.ReadFile(filepath.Join(dir, "all.m3u"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(all), "#EXTINF") != 3 {
		t.Errorf("all.m3u should have 3 channels:\n%s", all)
	}
	if !strings.HasPrefix(string(all), "#EXTM3U") {
		t.Errorf("all.m3u missing header")
	}

	br, err := os.ReadFile(filepath.Join(dir, "countries", "Brazil.m3u"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(br), "#EXTINF") != 2 {
		t.Errorf("Brazil.m3u should have 2 channels:\n%s", br)
	}
	if !strings.Contains(string(br), `group-title="Brazil"`) {
		t.Errorf("Brazil.m3u missing group-title")
	}
}

func TestSanitize(t *testing.T) {
	cases := map[string]string{
		"United States":   "United States",
		"Trinidad/Tobago": "Trinidad-Tobago",
		"日本 / Japan":      "日本 - Japan",
		"":                "Uncategorized",
	}
	for in, want := range cases {
		if got := sanitize(in); got != want {
			t.Errorf("sanitize(%q) = %q, want %q", in, got, want)
		}
	}
}
