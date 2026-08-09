package tui

import "testing"

func TestApplyThemeResolution(t *testing.T) {
	cases := []struct {
		name   string
		darkBG bool
		want   string
	}{
		{"auto", true, "tokyonight"},
		{"auto", false, "tokyonight-light"},
		{"", true, "tokyonight"},
		{"omarchy", true, "tokyonight"},     // alias
		{"latte", true, "catppuccin-latte"}, // alias
		{"catppuccin", true, "catppuccin"},  // direct
		{"arch", true, "arch"},
		{"cachyos", true, "cachyos"},
		{"does-not-exist", true, "tokyonight"}, // fallback
	}
	for _, c := range cases {
		if got := ApplyTheme(c.name, c.darkBG, true); got != c.want {
			t.Errorf("ApplyTheme(%q, dark=%v) = %q, want %q", c.name, c.darkBG, got, c.want)
		}
	}
}

func TestThemeNamesNonEmpty(t *testing.T) {
	names := ThemeNames()
	if len(names) < 8 {
		t.Fatalf("expected several themes, got %d", len(names))
	}
	// registry keys must all be resolvable palettes
	for _, n := range names {
		if _, isAlias := aliases[n]; isAlias {
			continue
		}
		if _, ok := themes[n]; !ok {
			t.Errorf("theme %q listed but not in registry", n)
		}
	}
}

func TestApplyThemeBuildsStyles(t *testing.T) {
	ApplyTheme("gruvbox", true, true)
	// A style built from the palette must carry that palette's background.
	if canvasStyle.GetBackground() != cBg {
		t.Errorf("canvasStyle background not rebuilt for theme")
	}
	if cBg != themes["gruvbox"].Bg {
		t.Errorf("active bg = %v, want gruvbox bg %v", cBg, themes["gruvbox"].Bg)
	}
	ApplyTheme("tokyonight", true, true)
}
