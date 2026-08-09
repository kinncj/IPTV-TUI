package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/kinncj/iptv/common/catalog"
	"github.com/kinncj/iptv/common/probe"
)

// header renders `❯ title …… right` followed by a full-width rule.
func header(title, right string, width int) string {
	left := headerChevron.Render(glChevron()+" ") + headerTitle.Render(title)
	r := headerCount.Render(right)
	gap := width - lipgloss.Width(left) - lipgloss.Width(r)
	if gap < 1 {
		gap = 1
	}
	bar := left + canvasStyle.Render(strings.Repeat(" ", gap)) + r
	rule := ruleStyle.Render(strings.Repeat("─", max(1, width)))
	return bar + "\n" + rule
}

// scrollbar builds a vertical scrollbar column of the given height, with a thumb
// sized/positioned from the selection index within the total item count.
func scrollbar(height, total, index int) string {
	if height <= 0 {
		return ""
	}
	if total <= 1 || height >= total {
		// Nothing to scroll — render a faint full-height track.
		col := make([]string, height)
		for i := range col {
			col[i] = scrollTrack.Render(glSbTrack())
		}
		return lipgloss.JoinVertical(lipgloss.Left, col...)
	}

	thumb := max(1, height*height/total)
	pos := 0
	if total-1 > 0 {
		pos = index * (height - thumb) / (total - 1)
	}

	col := make([]string, height)
	for i := 0; i < height; i++ {
		switch {
		case i == 0:
			col[i] = scrollThumb.Render(glSbUp())
		case i == height-1:
			col[i] = scrollThumb.Render(glSbDown())
		case i >= pos && i < pos+thumb:
			col[i] = scrollThumb.Render(glSbThumb())
		default:
			col[i] = scrollTrack.Render(glSbTrack())
		}
	}
	return lipgloss.JoinVertical(lipgloss.Left, col...)
}

// detailInfo carries the optional live extras rendered in the detail card.
type detailInfo struct {
	favorite bool
	epgNow   string
	epgNext  string
}

// detailCard renders the right-hand info panel for a channel.
func detailCard(ch catalog.Channel, st probe.Status, info detailInfo, width, height int) string {
	inner := width - 6 // padding + border
	if inner < 8 {
		inner = 8
	}

	title := truncate(ch.Name, inner-2)
	nameStyle := lipgloss.NewStyle().Foreground(cPurple).Background(cBgElev).Bold(true)
	name := nameStyle.Render(glStar() + " " + title)
	if info.favorite {
		name = nameStyle.Render(title) +
			lipgloss.NewStyle().Foreground(cYellow).Background(cBgElev).Render("  "+glStar()+" fav")
	}

	statusRow := labelStyle.Render("status  ") +
		lipgloss.NewStyle().Foreground(statusColor(st)).Background(cBgElev).
			Render(statusIcon(st)+" "+statusLabel(st))

	rows := []string{
		name,
		"",
		statusRow,
		kv("group", ch.Group, inner),
		kv("source", ch.Source, inner),
		kv("tvg-id", orDash(ch.TvgID), inner),
	}

	if info.epgNow != "" || info.epgNext != "" {
		rows = append(rows, "",
			lipgloss.NewStyle().Foreground(cPurple).Background(cBgElev).Bold(true).Render("Guide"))
		if info.epgNow != "" {
			rows = append(rows, labelStyle.Render("now   ")+valueStyle.Render(truncate(info.epgNow, inner-6)))
		}
		if info.epgNext != "" {
			rows = append(rows, labelStyle.Render("next  ")+valueStyle.Render(truncate(info.epgNext, inner-6)))
		}
	}

	rows = append(rows, "",
		labelStyle.Render("url"),
		valueStyle.Render(wrap(ch.URL, inner)))

	// Pad every line to the card's inner width with the elevated background so
	// there are no darker gaps between content and the card edge.
	rowbg := lipgloss.NewStyle().Background(cBgElev).Width(inner)
	for i, r := range rows {
		rows[i] = rowbg.Render(r)
	}

	body := lipgloss.JoinVertical(lipgloss.Left, rows...)
	return cardStyle.Width(width - 2).Height(height - 2).Render(body)
}

func kv(k, v string, inner int) string {
	key := labelStyle.Width(8).Render(k)
	val := valueStyle.Render(truncate(v, inner-8))
	return key + val
}

// footer renders a three-zone bar: a player pill on the left (flashing when just
// switched or showing a toast), the key legend centered, and brand on the right.
func footer(width int, pairs [][2]string, playerLabel, toast string, flash bool) string {
	pill := ""
	if playerLabel != "" {
		text := glPlay() + " " + playerLabel
		st := pillStyle
		if flash {
			st = pillFlashStyle
			if toast != "" {
				text = toast
			}
		}
		pill = st.Render(text)
	}

	var parts []string
	for _, p := range pairs {
		key := lipgloss.NewStyle().Foreground(cOrange).Background(cBg).Render("[" + p[0] + "]")
		parts = append(parts, key+mutedStyle.Render(" "+p[1]))
	}
	keys := strings.Join(parts, mutedStyle.Render("   "))

	pillW := lipgloss.Width(pill)
	center := lipgloss.NewStyle().Background(cBg).
		Width(max(1, width-pillW)).Align(lipgloss.Center).Render(keys)
	return lipgloss.JoinHorizontal(lipgloss.Top, pill, center)
}

func truncate(s string, n int) string {
	if n <= 1 {
		return ""
	}
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n-1]) + "…"
}

func wrap(s string, n int) string {
	if n <= 0 {
		return s
	}
	r := []rune(s)
	var b strings.Builder
	for i := 0; i < len(r); i += n {
		end := i + n
		if end > len(r) {
			end = len(r)
		}
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString(string(r[i:end]))
	}
	return b.String()
}

func orDash(s string) string {
	if s == "" {
		return "—"
	}
	return s
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
