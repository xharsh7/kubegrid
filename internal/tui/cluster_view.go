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
	screenResource
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
	resourceView *resourceViewModel
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
	// When in resource screen, delegate all messages to resource view
	if m.activeScreen == screenResource && m.resourceView != nil {
		if wsm, ok := msg.(tea.WindowSizeMsg); ok {
			m.width = wsm.Width
			m.height = wsm.Height
		}
		updated, cmd := m.resourceView.Update(msg)
		rv := updated.(resourceViewModel)
		if rv.wantBack {
			m.activeScreen = screenList
			m.resourceView = nil
			return m, nil
		}
		m.resourceView = &rv
		return m, cmd
	}

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
			if m.cursor < len(m.clusters) {
				c := m.clusters[m.cursor]
				if !c.Reachable {
					m.activeScreen = screenInspect
					return m, nil
				}
				rv, err := NewResourceView(c.Context, "default")
				if err != nil {
					m.activeScreen = screenInspect
					return m, nil
				}
				m.resourceView = &rv
				m.activeScreen = screenResource
				return m, rv.Init()
			}

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
	case screenResource:
		if m.resourceView != nil {
			return m.resourceView.View()
		}
		return m.listView()
	default:
		return m.listView()
	}
}

// --------------------------------------------Retro Mainframe View---------------------------------------------
func (m clusterViewModel) listView() string {
	w := m.width
	if w < 20 {
		w = 80
	}

	inner := w
	hline := strings.Repeat("━", inner)

	var s string

	header := "  KUBEGRID :: MULTI-CLUSTER OPS"
	if inner >= 50 {
		header = "  KUBEGRID :: MULTI-CLUSTER OPERATIONS CONSOLE"
	}
	s += "┏" + hline + "┓\n"
	s += fmt.Sprintf("┃%-*s┃\n", inner, header)
	s += "┣" + hline + "┫\n"

	// Adaptive column widths based on available space
	stateW := 4 // "UP" / "DOWN"
	latencyW := 7
	if inner < 50 {
		latencyW = 5
	}

	// Give cluster and context proportional space from what remains
	remaining := inner - 3 - 1 - stateW - 1 - latencyW // " > " + gaps
	if remaining < 10 {
		remaining = 10
	}
	clusterW := remaining * 35 / 100
	contextW := remaining - clusterW
	if clusterW < 6 {
		clusterW = 6
	}
	if contextW < 6 {
		contextW = 6
	}

	colHeader := fmt.Sprintf("   %-*s %-*s %-*s %-*s",
		clusterW, "CLUSTER", contextW, "CONTEXT", stateW, "ST", latencyW, "LATENCY")
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

		clusterName := truncate(c.Context.FriendlyName, clusterW)
		contextName := truncate(c.Context.Name, contextW)

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
	if inner >= 60 {
		s += "  jk:Nav g/G:Top/Bot Enter:Resources R:Refresh ?:Help Q:Quit\n"
	} else {
		s += "  jk:Nav Enter:Open R:Ref ?:Help Q:Quit\n"
	}

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

	inner := w
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
