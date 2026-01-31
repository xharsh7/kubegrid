package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type pane struct {
	m tea.Model
}

type mode int

const (
	modeNormal mode = iota
	modeCommand
)

type appModel struct {
	root         *layoutNode
	activePane   *layoutNode
	currentMode  mode
	commandInput string
	message      string
	width        int
	height       int
}

func NewApp(base tea.Model) appModel {
	p := pane{m: base}
	root := newPaneNode(p)
	
	return appModel{
		root:       root,
		activePane: root,
		currentMode: modeNormal,
	}
}

func (a appModel) Init() tea.Cmd {
	if a.activePane != nil && a.activePane.pane != nil {
		return a.activePane.pane.m.Init()
	}
	return nil
}

func (a appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height

	case tea.KeyMsg:
		// Handle command mode
		if a.currentMode == modeCommand {
			return a.handleCommandMode(msg)
		}

		// Handle normal mode
		switch msg.String() {
		case "ctrl+c":
			return a, tea.Quit

		case ":":
			a.currentMode = modeCommand
			a.commandInput = ""
			return a, nil

		case "ctrl+v":
			if a.activePane != nil {
				splitNode(a.activePane, true)
				a.message = "Split vertical"
			}
			return a, nil

		case "ctrl+h":
			if a.activePane != nil {
				splitNode(a.activePane, false)
				a.message = "Split horizontal"
			}
			return a, nil

		case "tab":
			a.cycleFocus()
			return a, nil

		case "shift+tab":
			a.cycleFocusReverse()
			return a, nil
		}
	}

	// Forward input to active pane
	if a.activePane != nil && a.activePane.pane != nil {
		m, cmd := a.activePane.pane.m.Update(msg)
		a.activePane.pane.m = m
		return a, cmd
	}

	return a, nil
}

func (a *appModel) handleCommandMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		a.currentMode = modeNormal
		a.commandInput = ""
		return a, nil

	case "enter":
		a.currentMode = modeNormal
		cmd := a.executeCommand(a.commandInput)
		a.commandInput = ""
		return a, cmd

	case "backspace":
		if len(a.commandInput) > 0 {
			a.commandInput = a.commandInput[:len(a.commandInput)-1]
		}

	default:
		a.commandInput += msg.String()
	}

	return a, nil
}

func (a *appModel) executeCommand(cmd string) tea.Cmd {
	switch cmd {
	case "q", "quit":
		return tea.Quit
	case "vsplit":
		if a.activePane != nil {
			splitNode(a.activePane, true)
			a.message = "Split vertical"
		}
	case "hsplit":
		if a.activePane != nil {
			splitNode(a.activePane, false)
			a.message = "Split horizontal"
		}
	case "help":
		a.message = "Commands: :q :vsplit :hsplit"
	default:
		a.message = fmt.Sprintf("Unknown command: %s", cmd)
	}
	return nil
}

func (a *appModel) cycleFocus() {
	// Simple focus cycling - traverse the tree
	next := a.findNextPane(a.root, a.activePane, false)
	if next != nil {
		a.activePane = next
		a.message = "Focus changed"
	}
}

func (a *appModel) cycleFocusReverse() {
	next := a.findNextPane(a.root, a.activePane, true)
	if next != nil {
		a.activePane = next
		a.message = "Focus changed"
	}
}

func (a *appModel) findNextPane(node *layoutNode, current *layoutNode, reverse bool) *layoutNode {
	// Collect all panes in order
	var panes []*layoutNode
	a.collectPanes(node, &panes)

	if len(panes) <= 1 {
		return current
	}

	// Find current index
	currentIdx := 0
	for i, p := range panes {
		if p == current {
			currentIdx = i
			break
		}
	}

	// Calculate next index
	nextIdx := currentIdx + 1
	if reverse {
		nextIdx = currentIdx - 1
		if nextIdx < 0 {
			nextIdx = len(panes) - 1
		}
	} else if nextIdx >= len(panes) {
		nextIdx = 0
	}

	return panes[nextIdx]
}

func (a *appModel) collectPanes(node *layoutNode, panes *[]*layoutNode) {
	if node == nil {
		return
	}

	if node.kind == nodePane {
		*panes = append(*panes, node)
		return
	}

	a.collectPanes(node.first, panes)
	a.collectPanes(node.second, panes)
}

func (a appModel) View() string {
	var s strings.Builder

	// Calculate content height (reserve space for status and command line)
	contentHeight := a.height - 2
	if contentHeight < 10 {
		contentHeight = 10
	}

	// Render the layout tree
	if a.root != nil {
		content := a.root.render(a.width, contentHeight, a.activePane)
		s.WriteString(content)
		s.WriteString("\n")
	}

	// Status line
	statusLine := " KUBEGRID"
	if a.message != "" {
		statusLine += fmt.Sprintf(" | %s", a.message)
	}
	s.WriteString(fmt.Sprintf("%-*s\n", a.width, statusLine))

	// Command line
	if a.currentMode == modeCommand {
		s.WriteString(fmt.Sprintf(":%s_", a.commandInput))
	} else {
		s.WriteString(" :command  ^V:vsplit  ^H:hsplit  Tab:cycle  ^C:quit")
	}

	return s.String()
}
