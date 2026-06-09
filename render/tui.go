package render

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"rootinfo/query"
)

type resultMsg []query.Result
type tickMsg struct{}

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
}

func newTUIModel(cfg TUIConfig) tuiModel {
	return tuiModel{cfg: cfg, querying: true}
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
	return header + FormatTable(m.results, m.cfg.Opts, m.cfg.Meta)
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
