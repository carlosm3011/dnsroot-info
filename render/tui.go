package render

import (
	"fmt"
	"math"
	"slices"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"rootinfo/query"
)

type resultMsg []query.Result
type tickMsg struct{}

type sortKey int

const (
	sortByLetter  sortKey = iota
	sortByIPv4RTT sortKey = iota
	sortByIPv6RTT sortKey = iota
)

// TUIConfig holds the parameters for a continuous TUI session.
type TUIConfig struct {
	Runner    *query.Runner
	Interval  time.Duration
	MaxCount  int
	Opts      Options
	Meta      Meta
	OnRefresh func(results []query.Result, n int) // called after each query batch; optional
}

type tuiModel struct {
	cfg      TUIConfig
	results  []query.Result
	refresh  int
	querying bool
	sort     sortKey
	sortAsc  bool
}

func newTUIModel(cfg TUIConfig) tuiModel {
	return tuiModel{cfg: cfg, querying: true, sort: sortByLetter, sortAsc: true}
}

func (m tuiModel) Init() tea.Cmd {
	return runQueryCmd(m.cfg.Runner)
}

func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case resultMsg:
		m.results = []query.Result(msg)
		m.querying = false
		m.refresh++
		if m.cfg.OnRefresh != nil {
			m.cfg.OnRefresh(m.results, m.refresh)
		}
		if m.cfg.MaxCount > 0 && m.refresh >= m.cfg.MaxCount {
			return m, tea.Quit
		}
		return m, tea.Tick(m.cfg.Interval, func(time.Time) tea.Msg { return tickMsg{} })

	case tickMsg:
		m.querying = true
		return m, runQueryCmd(m.cfg.Runner)

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "1":
			if m.sort == sortByLetter {
				m.sortAsc = !m.sortAsc
			} else {
				m.sort = sortByLetter
				m.sortAsc = true
			}
		case "2":
			if m.sort == sortByIPv4RTT {
				m.sortAsc = !m.sortAsc
			} else {
				m.sort = sortByIPv4RTT
				m.sortAsc = true
			}
		case "3":
			if m.sort == sortByIPv6RTT {
				m.sortAsc = !m.sortAsc
			} else {
				m.sort = sortByIPv6RTT
				m.sortAsc = true
			}
		}
	}
	return m, nil
}

func (m tuiModel) View() string {
	status := "querying..."
	if !m.querying && m.refresh > 0 {
		status = fmt.Sprintf("refresh #%d  next in %v  q to quit", m.refresh, m.cfg.Interval)
	}
	header := fmt.Sprintf("rootinfo  %s  %s\n\n",
		time.Now().UTC().Format("2006-01-02 15:04:05 UTC"),
		status,
	)
	if len(m.results) == 0 {
		return header
	}
	sorted := sortResults(m.results, m.sort, m.sortAsc)
	hint := sortHint(m.sort, m.sortAsc)
	return header + FormatTable(sorted, m.cfg.Opts, m.cfg.Meta) + "\n" + hint
}

func sortResults(results []query.Result, key sortKey, asc bool) []query.Result {
	out := slices.Clone(results)
	slices.SortStableFunc(out, func(a, b query.Result) int {
		var cmp int
		switch key {
		case sortByIPv4RTT:
			cmp = cmpRTT(a.IPv4RTT, b.IPv4RTT, a.IPv4Err, b.IPv4Err)
		case sortByIPv6RTT:
			cmp = cmpRTT(a.IPv6RTT, b.IPv6RTT, a.IPv6Err, b.IPv6Err)

		default:
			if a.Server.Letter < b.Server.Letter {
				cmp = -1
			} else if a.Server.Letter > b.Server.Letter {
				cmp = 1
			}
		}
		if !asc {
			cmp = -cmp
		}
		return cmp
	})
	return out
}

func cmpRTT(aRTT, bRTT time.Duration, aErr, bErr error) int {
	aVal := rttValue(aRTT, aErr)
	bVal := rttValue(bRTT, bErr)
	if aVal < bVal {
		return -1
	}
	if aVal > bVal {
		return 1
	}
	return 0
}

func rttValue(rtt time.Duration, err error) float64 {
	if err != nil {
		return math.MaxFloat64
	}
	return float64(rtt)
}

func sortHint(key sortKey, asc bool) string {
	dir := func(k sortKey) string {
		if key == k {
			if asc {
				return " ▲"
			}
			return " ▼"
		}
		return ""
	}
	return fmt.Sprintf("[1] sort by server%s  [2] sort by IPv4 RTT%s  [3] sort by IPv6 RTT%s  (same key toggles asc/desc)",
		dir(sortByLetter), dir(sortByIPv4RTT), dir(sortByIPv6RTT))
}

func runQueryCmd(runner *query.Runner) tea.Cmd {
	return func() tea.Msg {
		return resultMsg(runner.Run())
	}
}

// RunTUI starts the full-screen bubbletea TUI for continuous mode.
func RunTUI(cfg TUIConfig) error {
	p := tea.NewProgram(newTUIModel(cfg), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
