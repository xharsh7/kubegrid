package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/xharsh7/kubegrid/internal/cluster"
	"github.com/xharsh7/kubegrid/internal/config"
)

type screen int

const (
	screenList screen = iota
	screenInspect
	screenHelp
)

type clusterViewModel struct {
	clusters []cluster.ClusterStatus
	cursor   int
	scroll   int // viewport scroll offset for list view
	refresh  func() []cluster.ClusterStatus
	contexts []config.KubeContext
	width    int
	height   int

	activeScreen screen
}

func NewClusterView(data []cluster.ClusterStatus, refreshFn func() []cluster.ClusterStatus) clusterViewModel {
	return clusterViewModel{
		clusters:     data,
		refresh:      refreshFn,
		activeScreen: screenList,
	}
}

func (m clusterViewModel) Init() tea.Cmd {
	return nil
}

func (m clusterViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		// Help screen handling
		if m.activeScreen == screenHelp {
			m.activeScreen = screenList
			return m, nil
		}

		switch msg.String() {
		case "q":
			if m.activeScreen == screenInspect {
				m.activeScreen = screenList
				return m, nil
			}
			return m, tea.Quit

		case "ctrl+c":
			return m, tea.Quit

		case "?":
			m.activeScreen = screenHelp
			return m, nil

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				m.ensureCursorVisible()
			}

		case "down", "j":
			if m.cursor < len(m.clusters)-1 {
				m.cursor++
				m.ensureCursorVisible()
			}

		case "g":
			m.cursor = 0
			m.scroll = 0

		case "G":
			m.cursor = len(m.clusters) - 1
			m.ensureCursorVisible()

		case "r":
			m.clusters = m.refresh()

		case "enter":
			m.activeScreen = screenInspect

		case "esc":
			m.activeScreen = screenList
		}
	}
	return m, nil
}

// --------------------------------------------Retro Mainframe View---------------------------------------------
func (m clusterViewModel) View() string {
	switch m.activeScreen {
	case screenHelp:
		return NewHelpView().View()
	case screenInspect:
		return m.inspectView()
	default:
		return m.listView()
	}
}

// --------------------------------------------Retro Mainframe View---------------------------------------------
func (m clusterViewModel) listView() string {
	w := m.width
	if w < 40 {
		w = 80
	}

	inner := w - 2 // space between ┃ and ┃
	hline := strings.Repeat("━", inner)

	var s string

	header := "  KUBEGRID :: MULTI-CLUSTER OPERATIONS CONSOLE"
	s += "┏" + hline + "┓\n"
	s += fmt.Sprintf("┃%-*s┃\n", inner, header)
	s += "┣" + hline + "┫\n"

	// Column widths
	clusterW := 14
	stateW := 8
	latencyW := 7
	fixedW := 3 + 1 + clusterW + 1 + stateW + 1 + latencyW + 1
	contextW := inner - fixedW
	if contextW < 10 {
		contextW = 10
	}

	colHeader := fmt.Sprintf("   CLUSTER        %-*s STATE    LATENCY", contextW, "CONTEXT")
	s += fmt.Sprintf("┃%-*s┃\n", inner, colHeader)
	s += "┣" + hline + "┫\n"

	// Calculate visible window for scrolling
	// header(3) + colheader(2) + footer(2) = 7 lines of chrome
	visibleRows := m.height - 7
	if visibleRows < 1 {
		visibleRows = 1
	}

	// Clamp scroll so cursor is always visible
	scroll := m.scroll
	if m.cursor < scroll {
		scroll = m.cursor
	}
	if m.cursor >= scroll+visibleRows {
		scroll = m.cursor - visibleRows + 1
	}
	if scroll < 0 {
		scroll = 0
	}
	maxScroll := len(m.clusters) - visibleRows
	if maxScroll < 0 {
		maxScroll = 0
	}
	if scroll > maxScroll {
		scroll = maxScroll
	}

	end := scroll + visibleRows
	if end > len(m.clusters) {
		end = len(m.clusters)
	}

	for i := scroll; i < end; i++ {
		c := m.clusters[i]
		cursor := " "
		if i == m.cursor {
			cursor = ">"
		}

		state := "DOWN"
		if c.Reachable {
			state = "UP"
		}

		lat := c.Latency.String()
		if len(lat) > latencyW {
			lat = lat[:latencyW]
		}

		clusterName := c.Context.FriendlyName
		contextName := c.Context.Name

		if len(clusterName) > clusterW {
			clusterName = clusterName[:clusterW]
		}
		if len(contextName) > contextW {
			contextName = contextName[:contextW]
		}

		row := fmt.Sprintf(" %s %-*s %-*s %-*s %-*s",
			cursor,
			clusterW, clusterName,
			contextW, contextName,
			stateW, state,
			latencyW, lat,
		)
		s += fmt.Sprintf("┃%-*s┃\n", inner, row)
	}

	// Pad remaining visible rows if list is shorter than viewport
	for i := end - scroll; i < visibleRows; i++ {
		s += fmt.Sprintf("┃%-*s┃\n", inner, "")
	}

	s += "┗" + hline + "┛\n"
	s += "  ↑↓/jk:Navigate g/G:Top/Bottom Enter:Inspect R:Refresh ?:Help Q:Quit\n"

	return s
}

func (m clusterViewModel) inspectView() string {
	if m.cursor >= len(m.clusters) {
		return ""
	}
	c := m.clusters[m.cursor]

	w := m.width
	if w < 40 {
		w = 50
	}

	inner := w - 2
	if inner < 20 {
		inner = 20
	}
	hline := strings.Repeat("━", inner)

	line := func(label, value string) string {
		content := fmt.Sprintf("%s %s", label, value)
		if len(content) > inner {
			content = content[:inner]
		}
		return fmt.Sprintf("┃ %-*s ┃\n", inner-2, content)
	}

	var s string
	s += "┏" + hline + "┓\n"
	title := "Cluster: " + c.Context.FriendlyName
	if len(title) > inner-2 {
		title = title[:inner-2]
	}
	s += fmt.Sprintf("┃ %-*s ┃\n", inner-2, title)
	s += "┣" + hline + "┫\n"

	status := "DOWN"
	if c.Reachable {
		status = "UP"
	}

	s += line("Status:", status)
	s += line("Context:", c.Context.Name)
	s += line("Latency:", c.Latency.String())

	if c.Error != nil {
		errMsg := c.Error.Error()
		if len(errMsg) > inner-12 {
			errMsg = errMsg[:inner-12]
		}
		s += line("Error:", errMsg)
	}

	s += "┗" + hline + "┛\n"
	s += "  Esc/Q:Back   R:Refresh\n"

	return s
}

// ensureCursorVisible adjusts scroll so the cursor is always in the viewport
func (m *clusterViewModel) ensureCursorVisible() {
	visibleRows := m.height - 7
	if visibleRows < 1 {
		visibleRows = 1
	}
	if m.cursor < m.scroll {
		m.scroll = m.cursor
	}
	if m.cursor >= m.scroll+visibleRows {
		m.scroll = m.cursor - visibleRows + 1
	}
	if m.scroll < 0 {
		m.scroll = 0
	}
}
