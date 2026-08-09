package inlinevid

import (
	"encoding/base64"
	"fmt"
	"strings"
)

// kittyTransmit builds the escape sequence that sends one RGB frame (w x h
// pixels) to the terminal as image `id`, to be shown across cols x rows cells.
// It is chunked to kitty's 4096-byte payload limit and marked U=1 so it displays
// through unicode placeholders. c=cols,r=rows make the image span the whole
// placeholder grid instead of a default cell size. Re-sending with the same id
// replaces the image, which is how video updates.
func kittyTransmit(rgb []byte, w, h, cols, rows, id int) string {
	b64 := base64.StdEncoding.EncodeToString(rgb)
	const chunk = 4096
	var sb strings.Builder
	first := true
	for {
		n := chunk
		if n > len(b64) {
			n = len(b64)
		}
		part := b64[:n]
		b64 = b64[n:]
		more := 0
		if len(b64) > 0 {
			more = 1
		}
		sb.WriteString("\x1b_G")
		if first {
			// f=24: RGB. a=T: transmit and display. U=1: unicode placement.
			// c/r: span this many columns/rows of cells. q=2: no acknowledgements.
			fmt.Fprintf(&sb, "f=24,s=%d,v=%d,c=%d,r=%d,i=%d,a=T,U=1,q=2,m=%d", w, h, cols, rows, id, more)
			first = false
		} else {
			fmt.Fprintf(&sb, "m=%d", more)
		}
		sb.WriteByte(';')
		sb.WriteString(part)
		sb.WriteString("\x1b\\")
		if more == 0 {
			break
		}
	}
	return sb.String()
}

// kittyPlaceholders renders the cols x rows grid of placeholder cells that show
// image `id`. Each cell is U+10EEEE with row and column diacritics, and the
// foreground color carries the image id. Cell counts are clamped to the number
// of diacritics kitty defines.
func kittyPlaceholders(cols, rows, id int) []string {
	if cols > len(rowColumnDiacritics) {
		cols = len(rowColumnDiacritics)
	}
	if rows > len(rowColumnDiacritics) {
		rows = len(rowColumnDiacritics)
	}
	fg := fmt.Sprintf("\x1b[38;5;%dm", id) // id <= 255, read as the image id
	out := make([]string, rows)
	for r := 0; r < rows; r++ {
		var b strings.Builder
		b.WriteString(fg)
		for c := 0; c < cols; c++ {
			b.WriteString(placeholder)
			b.WriteRune(rowColumnDiacritics[r])
			b.WriteRune(rowColumnDiacritics[c])
		}
		b.WriteString("\x1b[0m")
		out[r] = b.String()
	}
	return out
}
