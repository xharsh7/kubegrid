package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/xharsh7/kubegrid/internal/cluster"
)

type screen int

const (
	screenList screen = iota
	screenInspect
)

type model struct {
	clusters []cluster.ClusterStatus
	cursor   int
	refresh  func() []cluster.ClusterStatus

	activeScreen screen
}

func NewClusterView(data []cluster.ClusterStatus, refreshFn func() []cluster.ClusterStatus) model {
	return model{
		clusters:     data,
		refresh:      refreshFn,
		activeScreen: screenList,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.clusters)-1 {
				m.cursor++
			}

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
func (m model) View() string {
	switch m.activeScreen {
	case screenInspect:
		return m.inspectView()
	default:
		return m.listView()
	}
}

// --------------------------------------------Retro Mainframe View---------------------------------------------
func (m model) listView() string {
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
	s += "  ↑↓ Navigate   R Refresh   Q Quit\n"

	return s
}

func (m model) inspectView() string {
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
	s += "  Esc Back   R Refresh\n"

	return s
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
