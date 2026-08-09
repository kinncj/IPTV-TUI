// Package tui is the Bubble Tea presentation layer. It depends inward on the
// domain (catalog) and on infrastructure interfaces (player, probe); it owns no
// business rules beyond navigation and rendering.
package tui

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kinncj/IPTV-TUI/common/catalog"
	"github.com/kinncj/IPTV-TUI/common/config"
	"github.com/kinncj/IPTV-TUI/common/epg"
	"github.com/kinncj/IPTV-TUI/common/inlinevid"
	"github.com/kinncj/IPTV-TUI/common/player"
	"github.com/kinncj/IPTV-TUI/common/probe"
	"github.com/kinncj/IPTV-TUI/common/state"
)

type viewState int

const (
	viewHome viewState = iota
	viewGroups
	viewChannels
	viewSources
)

// SourceInfo describes a playlist source for the in-TUI manager.
type SourceInfo struct {
	Name    string
	URL     string
	Builtin bool
}

// Model is the root Bubble Tea model.
type Model struct {
	cat     catalog.Catalog
	store   *statusStore
	prober  *probe.Prober
	players []player.Player
	playerI int

	groups   list.Model
	channels list.Model
	state    viewState
	showHelp bool

	width, height int
	status        string
	curGroup      string

	// persistence + guide + reload (v1.1)
	st         *state.State
	termVO     string // "" disables inline terminal playback
	reload     func() (catalog.Catalog, error)
	guide      *epg.Guide
	guideGroup string

	// in-TUI source manager
	builtins  []SourceInfo
	localPath string
	srcRows   []SourceInfo
	srcIdx    int
	srcInput  textinput.Model
	srcAdding bool

	// grouping mode (v1.3): country vs category
	byCountry    catalog.Catalog
	byCategory   catalog.Catalog
	byCategoryOn bool

	// terminal-player chooser + built-in inline video
	picking    bool
	pickIdx    int
	pickOpts   []playerOption
	vid        *inlinevid.Stream
	vidRows    []string
	vidPrelude string // kitty image-transmit escape emitted before the grid
	vidTitle   string
	vidURL     string
	vidFull    bool
	vidGen     int    // invalidates frames from a superseded stream (resize/close)
	inlineMode string // inlinevid.ModeHalfBlock or ModeKitty

	// animation + transient notifications
	frame     int
	animating bool
	toast     string
	toastTTL  int

	// background probe sweep (v1.2)
	probeConc int
	probeAll  bool
	sweeping  bool
	sweepURLs []string
}

type probeResultMsg struct {
	url    string
	status probe.Status
}

type launchResultMsg struct {
	channel string
	err     error
}

type termDoneMsg struct {
	channel string
	err     error
}

type reloadMsg struct {
	cat catalog.Catalog
	err error
}

type epgLoadedMsg struct {
	group string
	guide *epg.Guide
	err   error
}

type vidFrameMsg struct {
	prelude string
	rows    []string
	gen     int
}
type vidEndMsg struct {
	err error
	gen int
}

type tickMsg time.Time

type sweepProgressMsg struct{ done int }

// sweepBatch probes the next chunk of URLs concurrently (bounded by the prober's
// own semaphore) and reports progress. One batch = one re-render, so a full
// sweep stays responsive without spawning a goroutine per channel up front.
func (m Model) sweepBatch(start int) tea.Cmd {
	const batch = 64
	end := start + batch
	if end > len(m.sweepURLs) {
		end = len(m.sweepURLs)
	}
	urls := m.sweepURLs[start:end]
	store := m.store
	prober := m.prober
	return func() tea.Msg {
		var wg sync.WaitGroup
		for _, u := range urls {
			if s := store.get(u); s == probe.OK || s == probe.Dead || s == probe.Blocked {
				continue // already have a verdict
			}
			wg.Add(1)
			go func(u string) {
				defer wg.Done()
				store.set(u, probe.Checking)
				store.set(u, prober.Check(context.Background(), u))
			}(u)
		}
		wg.Wait()
		return sweepProgressMsg{done: end}
	}
}

// tickInterval drives the logo animation and toast timing (~11 fps).
const tickInterval = 90 * time.Millisecond

func tick() tea.Cmd {
	return tea.Tick(tickInterval, func(t time.Time) tea.Msg { return tickMsg(t) })
}

// animate (re)starts the tick loop if it isn't already running.
func (m *Model) animate() tea.Cmd {
	if m.animating {
		return nil
	}
	m.animating = true
	return tick()
}

// New builds the model from a catalog and the detected players.
func New(cat catalog.Catalog, players []player.Player) Model {
	in := textinput.New()
	in.Placeholder = "https://…/playlist.m3u8"
	in.Prompt = "url "
	in.CharLimit = 2048

	m := Model{
		cat:       cat,
		store:     newStatusStore(),
		prober:    probe.New(12, 6e9), // 12 concurrent, 6s timeout
		players:   players,
		groups:    newList(nil),
		channels:  newList(nil),
		state:     viewHome,
		localPath: config.UserPath(), // writes go to the OS config dir
		srcInput:  in,
	}
	if len(players) == 0 {
		m.status = "no player found — install mpv or vlc"
	} else {
		m.status = "player: " + players[0].Name()
	}
	m.setCatalog(cat)
	return m
}

// setCatalog stores the country-grouped catalog, derives the category grouping,
// points m.cat at the active mode, and rebuilds the group list.
func (m *Model) setCatalog(countryCat catalog.Catalog) {
	m.byCountry = countryCat
	m.byCategory = catalog.BuildByCategory(countryCat.Flatten())
	if m.byCategoryOn {
		m.cat = m.byCategory
	} else {
		m.cat = m.byCountry
	}
	m.refreshGroups()
}

// WithState attaches persisted favorites/last-played and rebuilds the group list
// so Favorites/Recent appear on top. Returns the model (builder style).
func (m Model) WithState(s *state.State) Model {
	m.st = s
	m.refreshGroups()
	return m
}

// WithTerminalPlayback enables inline playback via mpv with the given video
// output (empty string leaves it disabled).
func (m Model) WithTerminalPlayback(vo string) Model {
	m.termVO = vo
	return m
}

// WithReload injects a function that re-fetches the catalog from source.
func (m Model) WithReload(fn func() (catalog.Catalog, error)) Model {
	m.reload = fn
	return m
}

// WithSources records the built-in sources so the in-TUI manager can show them
// alongside user-added ones.
func (m Model) WithSources(builtins []SourceInfo) Model {
	m.builtins = builtins
	return m
}

// WithProbe configures the reachability prober's concurrency and per-probe timeout.
func (m Model) WithProbe(concurrency int, timeout time.Duration) Model {
	m.prober = probe.New(concurrency, timeout)
	m.probeConc = concurrency
	return m
}

// WithProbeAll enables a background sweep that probes every channel at startup.
func (m Model) WithProbeAll(on bool) Model {
	m.probeAll = on
	return m
}

// WithInlineMode sets the built-in inline video renderer (inlinevid.ModeKitty
// for sharp graphics, or ModeHalfBlock).
func (m Model) WithInlineMode(mode string) Model {
	m.inlineMode = mode
	return m
}

// refreshGroups rebuilds the group list: synthetic Favorites/Recent first, then
// the catalog's countries.
func (m *Model) refreshGroups() {
	var items []list.Item
	if m.st != nil {
		fav, recent := m.st.Snapshot()
		if len(fav) > 0 {
			items = append(items, groupItem{
				group: entriesGroup("Favorites", fav), icon: glStar(), iconColor: cYellow})
		}
		if len(recent) > 0 {
			items = append(items, groupItem{
				group: entriesGroup("Recent", recent), icon: gl("◷", "@"), iconColor: cCyan})
		}
	}
	for _, g := range m.cat.Groups {
		items = append(items, groupItem{group: g})
	}
	m.groups.SetItems(items)
}

// newList returns a bubbles list stripped of its built-in chrome, so we can draw
// our own header, scrollbar, and footer around the bare rows.
func newList(items []list.Item) list.Model {
	l := list.New(items, rowDelegate{}, 0, 0)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetShowPagination(false)
	l.SetFilteringEnabled(true)
	l.InfiniteScrolling = true
	return l
}

func (m Model) Init() tea.Cmd {
	m.animating = true
	return tick() // home splash animates from the first frame
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		lw, lh := m.listDims()
		m.groups.SetSize(lw, lh)
		m.channels.SetSize(m.channelListWidth(), lh)
		// Kick off the background probe sweep once, on the first size event.
		if m.probeAll && !m.sweeping && len(m.sweepURLs) == 0 {
			for _, g := range m.cat.Groups {
				for _, ch := range g.Channels {
					m.sweepURLs = append(m.sweepURLs, ch.URL)
				}
			}
			if len(m.sweepURLs) > 0 {
				m.sweeping = true
				return m, m.sweepBatch(0)
			}
		}
		return m, nil

	case sweepProgressMsg:
		if msg.done >= len(m.sweepURLs) {
			m.sweeping = false
			m.toast = "probe sweep complete"
			m.toastTTL = 20
			return m, m.animate()
		}
		return m, m.sweepBatch(msg.done)

	case tickMsg:
		m.frame++
		if m.toastTTL > 0 {
			m.toastTTL--
			if m.toastTTL == 0 {
				m.toast = ""
			}
		}
		// Keep ticking only while something needs to animate.
		if m.state == viewHome || m.toast != "" {
			return m, tick()
		}
		m.animating = false
		return m, nil

	case probeResultMsg:
		m.store.set(msg.url, msg.status)
		return m, nil

	case launchResultMsg:
		if msg.err != nil {
			m.status = "launch failed: " + msg.err.Error()
		} else {
			m.status = "▶ " + msg.channel + "  (" + m.playerName() + ")"
		}
		return m, nil

	case termDoneMsg:
		if msg.err != nil {
			m.toast = "playback error"
			m.toastTTL = 18
			return m, m.animate()
		}
		return m, nil

	case reloadMsg:
		if msg.err != nil {
			m.toast = "reload failed"
		} else {
			m.setCatalog(msg.cat)
			m.toast = "reloaded " + itoa(len(m.cat.Groups)) + " groups"
		}
		m.toastTTL = 20
		return m, m.animate()

	case epgLoadedMsg:
		if msg.err == nil && msg.group == m.curGroup {
			m.guide = msg.guide
			m.guideGroup = msg.group
		}
		return m, nil

	case vidFrameMsg:
		if m.vid == nil || msg.gen != m.vidGen {
			return m, nil // stale frame from a superseded stream
		}
		m.vidRows = msg.rows
		m.vidPrelude = msg.prelude
		return m, m.nextVidFrame()

	case vidEndMsg:
		if msg.gen != m.vidGen {
			return m, nil // the stream that ended was already replaced
		}
		m.closeVid()
		m.toast = "playback ended"
		m.toastTTL = 16
		return m, m.animate()

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m.delegateToList(msg)
}

func (m Model) delegateToList(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch m.state {
	case viewGroups:
		m.groups, cmd = m.groups.Update(msg)
	case viewChannels:
		m.channels, cmd = m.channels.Update(msg)
	}
	return m, cmd
}

func (m Model) playerName() string {
	if len(m.players) == 0 {
		return "none"
	}
	return m.players[m.playerI].Name()
}

// probeGroup returns a batch of commands probing every un-probed channel in the group.
func (m Model) probeGroup(g catalog.Group) tea.Cmd {
	cmds := make([]tea.Cmd, 0, len(g.Channels))
	for _, ch := range g.Channels {
		url := ch.URL
		if s := m.store.get(url); s == probe.OK || s == probe.Dead || s == probe.Blocked {
			continue
		}
		cmds = append(cmds, func() tea.Msg {
			m.store.set(url, probe.Checking)
			return probeResultMsg{url: url, status: m.prober.Check(context.Background(), url)}
		})
	}
	return tea.Batch(cmds...)
}

// loadGuide fetches the EPG for a group (if supported), bounded to its channels.
func (m Model) loadGuide(g catalog.Group) tea.Cmd {
	if !epg.Supported(g.Name) {
		return nil
	}
	wanted := make(map[string]bool, len(g.Channels))
	for _, ch := range g.Channels {
		if ch.TvgID != "" {
			wanted[ch.TvgID] = true
		}
	}
	name := g.Name
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
		defer cancel()
		guide, err := epg.Load(ctx, name, wanted)
		return epgLoadedMsg{group: name, guide: guide, err: err}
	}
}

// reloadCmd re-fetches the catalog from source off the UI goroutine.
func (m Model) reloadCmd() tea.Cmd {
	if m.reload == nil {
		return nil
	}
	fn := m.reload
	return func() tea.Msg {
		cat, err := fn()
		return reloadMsg{cat: cat, err: err}
	}
}

func (m Model) groupCountLabel() string {
	return fmt.Sprintf("%d groups · %s channels", len(m.cat.Groups), commas(m.cat.Total))
}

func itoa(n int) string { return strconv.Itoa(n) }

// commas formats an int with thousands separators (e.g. 16927 -> "16,927").
func commas(n int) string {
	s := strconv.Itoa(n)
	if n < 1000 {
		return s
	}
	var out []byte
	for i, c := range []byte(s) {
		if i > 0 && (len(s)-i)%3 == 0 {
			out = append(out, ',')
		}
		out = append(out, c)
	}
	return string(out)
}
