package tui

import (
	"sort"

	"github.com/charmbracelet/lipgloss"
	"github.com/kinncj/IPTV-TUI/common/probe"
)

// Palette is a named color scheme. lipgloss/termenv downgrades these hex values
// to the terminal's actual color depth, so the same palette looks right on a
// truecolor Ghostty and on a 16-color tty.
type Palette struct {
	Dark                        bool
	Bg, BgElev, Fg              lipgloss.Color
	Primary, Secondary, Accent  lipgloss.Color
	Ok, Warn, Err               lipgloss.Color
	Muted, Border               lipgloss.Color
	SheenMid, SheenPeak, Shadow lipgloss.Color
}

// themes registry. Distro packs (arch, cachyos, omarchy) are included alongside
// the classics; unknown names fall back to the auto default.
var themes = map[string]Palette{
	"tokyonight": {
		Dark: true, Bg: "#1a1b26", BgElev: "#24283b", Fg: "#c0caf5",
		Primary: "#bb9af7", Secondary: "#7dcfff", Accent: "#ff9e64",
		Ok: "#9ece6a", Warn: "#e0af68", Err: "#f7768e",
		Muted: "#565f89", Border: "#3b4261",
		SheenMid: "#c4b5fd", SheenPeak: "#f7f7ff", Shadow: "#33305a",
	},
	"tokyonight-light": {
		Dark: false, Bg: "#e1e2e7", BgElev: "#d0d1d9", Fg: "#343b58",
		Primary: "#7847bd", Secondary: "#007197", Accent: "#b15c00",
		Ok: "#587539", Warn: "#8c6c3e", Err: "#c64343",
		Muted: "#8990b3", Border: "#a8aecb",
		SheenMid: "#7847bd", SheenPeak: "#ffffff", Shadow: "#bcbfd0",
	},
	"catppuccin": {
		Dark: true, Bg: "#1e1e2e", BgElev: "#313244", Fg: "#cdd6f4",
		Primary: "#cba6f7", Secondary: "#89dceb", Accent: "#fab387",
		Ok: "#a6e3a1", Warn: "#f9e2af", Err: "#f38ba8",
		Muted: "#6c7086", Border: "#45475a",
		SheenMid: "#b4befe", SheenPeak: "#f5f5ff", Shadow: "#302d41",
	},
	"catppuccin-latte": {
		Dark: false, Bg: "#eff1f5", BgElev: "#dce0e8", Fg: "#4c4f69",
		Primary: "#8839ef", Secondary: "#209fb5", Accent: "#fe640b",
		Ok: "#40a02b", Warn: "#df8e1d", Err: "#d20f39",
		Muted: "#8c8fa1", Border: "#bcc0cc",
		SheenMid: "#7287fd", SheenPeak: "#ffffff", Shadow: "#ccd0da",
	},
	"gruvbox": {
		Dark: true, Bg: "#282828", BgElev: "#3c3836", Fg: "#ebdbb2",
		Primary: "#d3869b", Secondary: "#83a598", Accent: "#fe8019",
		Ok: "#b8bb26", Warn: "#fabd2f", Err: "#fb4934",
		Muted: "#928374", Border: "#504945",
		SheenMid: "#d3869b", SheenPeak: "#fbf1c7", Shadow: "#1d2021",
	},
	"nord": {
		Dark: true, Bg: "#2e3440", BgElev: "#3b4252", Fg: "#d8dee9",
		Primary: "#b48ead", Secondary: "#88c0d0", Accent: "#d08770",
		Ok: "#a3be8c", Warn: "#ebcb8b", Err: "#bf616a",
		Muted: "#616e88", Border: "#434c5e",
		SheenMid: "#81a1c1", SheenPeak: "#eceff4", Shadow: "#272c36",
	},
	"everforest": {
		Dark: true, Bg: "#2d353b", BgElev: "#343f44", Fg: "#d3c6aa",
		Primary: "#d699b6", Secondary: "#7fbbb3", Accent: "#e69875",
		Ok: "#a7c080", Warn: "#dbbc7f", Err: "#e67e80",
		Muted: "#859289", Border: "#3d484d",
		SheenMid: "#a7c080", SheenPeak: "#fdf6e3", Shadow: "#232a2e",
	},
	"arch": {
		Dark: true, Bg: "#0f1419", BgElev: "#1b2733", Fg: "#c5d4e3",
		Primary: "#1793d1", Secondary: "#33a1de", Accent: "#ffb454",
		Ok: "#7fd962", Warn: "#e6b450", Err: "#f07178",
		Muted: "#4d6273", Border: "#243b4d",
		SheenMid: "#33a1de", SheenPeak: "#eafcff", Shadow: "#0a2233",
	},
	"cachyos": {
		Dark: true, Bg: "#12181b", BgElev: "#1c262a", Fg: "#c8e0d8",
		Primary: "#1abc9c", Secondary: "#3fe0c0", Accent: "#f4a259",
		Ok: "#8ce99a", Warn: "#f6c177", Err: "#eb6f92",
		Muted: "#4a6a63", Border: "#233631",
		SheenMid: "#3fe0c0", SheenPeak: "#eafff8", Shadow: "#082820",
	},
}

// aliases map friendly names onto registry keys.
var aliases = map[string]string{
	"default":     "tokyonight",
	"tokyo-night": "tokyonight",
	"tokyo":       "tokyonight",
	"omarchy":     "tokyonight", // Omarchy ships Tokyo Night as its default
	"mocha":       "catppuccin",
	"latte":       "catppuccin-latte",
	"light":       "tokyonight-light",
}

// --- active theme state (assigned by ApplyTheme) ----------------------------

var (
	cBg, cBgElev, cFg          lipgloss.Color
	cPurple, cCyan, cOrange    lipgloss.Color
	cGreen, cYellow, cRed      lipgloss.Color
	cMuted, cBorder            lipgloss.Color
	cLavender, cWhite, cShadow lipgloss.Color
	logoGradient               []lipgloss.Color
	unicodeOK                  = true
)

var (
	canvasStyle                                        lipgloss.Style
	logoStyle, tagStyle, brandStyle                    lipgloss.Style
	headerChevron, headerTitle, headerCount, ruleStyle lipgloss.Style
	sectionStyle, mutedStyle, labelStyle, valueStyle   lipgloss.Style
	cardStyle, overlayStyle                            lipgloss.Style
	scrollTrack, scrollThumb                           lipgloss.Style
	pillStyle, pillFlashStyle                          lipgloss.Style
)

// ThemeNames returns the selectable theme names (registry + aliases), sorted.
func ThemeNames() []string {
	seen := map[string]bool{}
	var out []string
	for k := range themes {
		if !seen[k] {
			out = append(out, k)
			seen[k] = true
		}
	}
	for k := range aliases {
		if !seen[k] {
			out = append(out, k)
			seen[k] = true
		}
	}
	sort.Strings(out)
	return out
}

// ApplyTheme resolves a theme name (with "auto" following the terminal's
// dark/light background) and rebuilds all styles. Returns the resolved name.
func ApplyTheme(name string, darkBG, unicode bool) string {
	unicodeOK = unicode

	resolved := name
	if resolved == "" || resolved == "auto" {
		if darkBG {
			resolved = "tokyonight"
		} else {
			resolved = "tokyonight-light"
		}
	}
	if target, ok := aliases[resolved]; ok {
		resolved = target
	}
	p, ok := themes[resolved]
	if !ok {
		resolved = "tokyonight"
		p = themes[resolved]
	}

	cBg, cBgElev, cFg = p.Bg, p.BgElev, p.Fg
	cPurple, cCyan, cOrange = p.Primary, p.Secondary, p.Accent
	cGreen, cYellow, cRed = p.Ok, p.Warn, p.Err
	cMuted, cBorder = p.Muted, p.Border
	cLavender, cWhite, cShadow = p.SheenMid, p.SheenPeak, p.Shadow
	logoGradient = []lipgloss.Color{
		cPurple, cPurple, cLavender, cWhite, cCyan, cLavender, cPurple, cPurple,
	}

	buildStyles()
	return resolved
}

func buildStyles() {
	border := lipgloss.RoundedBorder()
	if !unicodeOK {
		border = lipgloss.NormalBorder()
	}

	canvasStyle = lipgloss.NewStyle().Background(cBg).Foreground(cFg)

	logoStyle = lipgloss.NewStyle().Foreground(cPurple).Background(cBg).Bold(true)
	tagStyle = lipgloss.NewStyle().Foreground(cMuted).Background(cBg)
	brandStyle = lipgloss.NewStyle().Foreground(cCyan).Background(cBg).Bold(true)

	headerChevron = lipgloss.NewStyle().Foreground(cPurple).Background(cBg).Bold(true)
	headerTitle = lipgloss.NewStyle().Foreground(cFg).Background(cBg).Bold(true)
	headerCount = lipgloss.NewStyle().Foreground(cCyan).Background(cBg).Bold(true)
	ruleStyle = lipgloss.NewStyle().Foreground(cBorder).Background(cBg)

	sectionStyle = lipgloss.NewStyle().Foreground(cPurple).Background(cBg).Bold(true)
	mutedStyle = lipgloss.NewStyle().Foreground(cMuted).Background(cBg)
	labelStyle = lipgloss.NewStyle().Foreground(cMuted).Background(cBgElev)
	valueStyle = lipgloss.NewStyle().Foreground(cFg).Background(cBgElev)

	cardStyle = lipgloss.NewStyle().Background(cBgElev).
		Border(border).BorderForeground(cBorder).BorderBackground(cBg).Padding(1, 2)
	overlayStyle = lipgloss.NewStyle().Background(cBg).
		Border(border).BorderForeground(cPurple).BorderBackground(cBg).Padding(1, 3)

	scrollTrack = lipgloss.NewStyle().Foreground(cBorder).Background(cBg)
	scrollThumb = lipgloss.NewStyle().Foreground(cPurple).Background(cBg)

	pillStyle = lipgloss.NewStyle().Foreground(cBg).Background(cCyan).Bold(true).Padding(0, 1)
	pillFlashStyle = lipgloss.NewStyle().Foreground(cBg).Background(cOrange).Bold(true).Padding(0, 1)
}

func statusColor(s probe.Status) lipgloss.Color {
	switch s {
	case probe.OK:
		return cGreen
	case probe.Blocked:
		return cYellow
	case probe.Dead:
		return cRed
	case probe.Checking:
		return cCyan
	default:
		return cMuted
	}
}

func statusLabel(s probe.Status) string {
	switch s {
	case probe.OK:
		return "reachable"
	case probe.Blocked:
		return "blocked"
	case probe.Dead:
		return "dead"
	case probe.Checking:
		return "checking…"
	default:
		return "unprobed"
	}
}
