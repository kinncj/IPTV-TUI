package inlinevid

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestKittyTransmitRoundTrip(t *testing.T) {
	rgb := make([]byte, 2*2*3) // 2x2 image
	for i := range rgb {
		rgb[i] = byte(i * 7)
	}
	esc := kittyTransmit(rgb, 2, 2, 42)

	if !strings.HasPrefix(esc, "\x1b_Gf=24,s=2,v=2,i=42,a=T,U=1,q=2,m=0;") {
		t.Fatalf("control header wrong: %q", esc[:40])
	}
	if !strings.HasSuffix(esc, "\x1b\\") {
		t.Fatalf("escape not terminated with ST")
	}
	// Extract the base64 payload and confirm it decodes to the input.
	payload := esc[strings.Index(esc, ";")+1 : len(esc)-2]
	got, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		t.Fatalf("payload not valid base64: %v", err)
	}
	if string(got) != string(rgb) {
		t.Errorf("payload does not round-trip the RGB data")
	}
}

func TestKittyTransmitChunks(t *testing.T) {
	// A frame large enough to need multiple 4096-byte base64 chunks.
	rgb := make([]byte, 8000)
	esc := kittyTransmit(rgb, 40, 66, 42)
	if strings.Count(esc, "\x1b_G") < 2 {
		t.Errorf("large frame should be split into multiple chunks")
	}
	if !strings.Contains(esc, "m=1") || !strings.Contains(esc, "m=0") {
		t.Errorf("chunked transmit should have m=1 continuation and m=0 final")
	}
}

func TestKittyPlaceholders(t *testing.T) {
	rows := kittyPlaceholders(3, 2, 42)
	if len(rows) != 2 {
		t.Fatalf("want 2 placeholder rows, got %d", len(rows))
	}
	if !strings.Contains(rows[0], "\x1b[38;5;42m") {
		t.Errorf("row should set the image id in the foreground: %q", rows[0])
	}
	if strings.Count(rows[0], placeholder) != 3 {
		t.Errorf("row should have 3 placeholder cells, got %d", strings.Count(rows[0], placeholder))
	}
	// First cell of row 0 uses row diacritic index 0 and column diacritic 0.
	first := string(placeholder) + string(rowColumnDiacritics[0]) + string(rowColumnDiacritics[0])
	if !strings.Contains(rows[0], first) {
		t.Errorf("row 0 col 0 diacritics missing")
	}
	// Row 1 uses row diacritic index 1.
	if !strings.Contains(rows[1], string(placeholder)+string(rowColumnDiacritics[1])) {
		t.Errorf("row 1 should use row diacritic index 1")
	}
}
