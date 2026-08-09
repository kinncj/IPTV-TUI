package inlinevid

import (
	"strings"
	"testing"
)

func TestRenderHalfBlocksDimensions(t *testing.T) {
	// 2x2 pixel frame -> 2 cols, 1 row (two vertical pixels per cell).
	// pixels: (0,0)=red (1,0)=green ; (0,1)=blue (1,1)=white
	buf := []byte{
		255, 0, 0, 0, 255, 0, // row 0: red, green
		0, 0, 255, 255, 255, 255, // row 1: blue, white
	}
	rows := renderHalfBlocks(buf, 2, 2, 2, 1)
	if len(rows) != 1 {
		t.Fatalf("want 1 row, got %d", len(rows))
	}
	line := rows[0]
	if strings.Count(line, "▀") != 2 {
		t.Errorf("want 2 half-block cells, got %d in %q", strings.Count(line, "▀"), line)
	}
	// First cell: fg red (top pixel), bg blue (bottom pixel).
	if !strings.Contains(line, "38;2;255;0;0;48;2;0;0;255m▀") {
		t.Errorf("first cell colors wrong: %q", line)
	}
	// Second cell: fg green, bg white.
	if !strings.Contains(line, "38;2;0;255;0;48;2;255;255;255m▀") {
		t.Errorf("second cell colors wrong: %q", line)
	}
	if !strings.HasSuffix(line, "\x1b[0m") {
		t.Errorf("row should reset SGR at the end: %q", line)
	}
}

func TestRenderHalfBlocksOddHeight(t *testing.T) {
	// h=1 but rows=1 expects a bottom pixel row that doesn't exist; must not panic.
	buf := []byte{10, 20, 30} // single pixel
	rows := renderHalfBlocks(buf, 1, 1, 1, 1)
	if len(rows) != 1 || strings.Count(rows[0], "▀") != 1 {
		t.Fatalf("expected one cell, got %q", rows[0])
	}
	// Bottom defaults to black (0,0,0) when absent.
	if !strings.Contains(rows[0], "38;2;10;20;30;48;2;0;0;0m") {
		t.Errorf("missing bottom pixel should default to black: %q", rows[0])
	}
}
