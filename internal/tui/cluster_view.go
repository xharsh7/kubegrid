package tui

import (
	"fmt"

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
	refresh  func() []cluster.ClusterStatus
	contexts []config.KubeContext

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
			}

		case "down", "j":
			if m.cursor < len(m.clusters)-1 {
				m.cursor++
			}

		case "g":
			m.cursor = 0

		case "G":
			m.cursor = len(m.clusters) - 1

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
	var s string

	header := "  KUBEGRID :: MULTI-CLUSTER OPERATIONS CONSOLE"
	s += "┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓\n"
	s += fmt.Sprintf("┃%-78s┃\n", header)
	s += "┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫\n"
	s += "┃   CLUSTER        CONTEXT                                   STATE    LATENCY  ┃\n"
	s += "┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫\n"

	for i, c := range m.clusters {
		cursor := " "
		if i == m.cursor {
			cursor = ">"
		}

		state := "DOWN"
		if c.Reachable {
			state = "UP"
		}

		lat := c.Latency.String()
		if len(lat) > 7 {
			lat = lat[:7]
		}

		cluster := c.Context.FriendlyName
		context := c.Context.Name

		if len(cluster) > 14 {
			cluster = cluster[:14]
		}
		if len(context) > 42 {
			context = context[:42]
		}

		s += fmt.Sprintf("┃ %s %-14s %-42s %-8s %-7s ┃\n",
			cursor,
			cluster,
			context,
			state,
			lat,
		)
	}

	s += "┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛\n"
	s += "  ↑↓/jk:Navigate g/G:Top/Bottom Enter:Inspect R:Refresh ?:Help Q:Quit\n"

	return s
}

func (m clusterViewModel) inspectView() string {
	c := m.clusters[m.cursor]

	const width = 46

	line := func(label, value string) string {
		content := fmt.Sprintf("%s %s", label, value)
		if len(content) > width {
			content = content[:width]
		}
		return fmt.Sprintf("┃ %-*s ┃\n", width, content)
	}

	var s string
	s += "┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓\n"
	s += fmt.Sprintf("┃ %-*s ┃\n", width, "Cluster: "+c.Context.FriendlyName)
	s += "┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫\n"

	status := "DOWN"
	if c.Reachable {
		status = "UP"
	}

	s += line("Status:", status)
	s += line("Context:", c.Context.Name)
	s += line("Latency:", c.Latency.String())

	if c.Error != nil {
		s += line("Error:", c.Error.Error())
	}

	s += "┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛\n"
	s += "  Esc/Q:Back   R:Refresh\n"

	return s
}
