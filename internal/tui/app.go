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
		root:        root,
		activePane:  root,
		currentMode: modeNormal,
	}
}

func (a appModel) Init() tea.Cmd {
	if a.activePane != nil && a.activePane.pane != nil {
		return a.activePane.pane.m.Init()
	}
	return nil
}

// paneCount returns the total number of leaf panes in the tree
func (a *appModel) paneCount() int {
	var panes []*layoutNode
	a.collectPanes(a.root, &panes)
	return len(panes)
}

func (a appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		// Broadcast size to ALL leaf panes so each knows the terminal dimensions
		a.broadcastSize(msg)
		return a, nil

	case tea.KeyMsg:
		// Handle command mode
		if a.currentMode == modeCommand {
			return a.handleCommandMode(msg)
		}

		// Handle normal mode — tmux-like splits and navigation
		switch msg.String() {
		case "ctrl+c":
			return a, tea.Quit

		case ":":
			a.currentMode = modeCommand
			a.commandInput = ""
			return a, nil

		// Vertical split (side by side) — ctrl+b works in all terminals
		case "ctrl+b", "ctrl+v":
			if a.activePane != nil && a.activePane.kind == nodePane {
				newFirst := splitNode(a.activePane, true)
				a.activePane = newFirst
				a.broadcastSize(tea.WindowSizeMsg{Width: a.width, Height: a.height})
				a.message = "Split vertical"
			}
			return a, nil

		// Horizontal split (top/bottom) — ctrl+g works in all terminals
		case "ctrl+g", "ctrl+h":
			if a.activePane != nil && a.activePane.kind == nodePane {
				newFirst := splitNode(a.activePane, false)
				a.activePane = newFirst
				a.broadcastSize(tea.WindowSizeMsg{Width: a.width, Height: a.height})
				a.message = "Split horizontal"
			}
			return a, nil

		// Close active pane
		case "ctrl+x":
			if a.activePane != nil && a.paneCount() > 1 {
				replacement := closePane(a.activePane)
				// Find the first leaf of the replacement
				a.activePane = firstLeaf(replacement)
				a.broadcastSize(tea.WindowSizeMsg{Width: a.width, Height: a.height})
				a.message = "Pane closed"
			}
			return a, nil

		// Tab / Shift+Tab: cycle focus
		case "tab":
			a.cycleFocus()
			return a, nil

		case "shift+tab":
			a.cycleFocusReverse()
			return a, nil
		}
	}

	// Forward all other input to active pane
	if a.activePane != nil && a.activePane.pane != nil {
		m, cmd := a.activePane.pane.m.Update(msg)
		a.activePane.pane.m = m
		return a, cmd
	}

	return a, nil
}

// broadcastSize sends a WindowSizeMsg to every leaf pane with its approximate
// inner dimensions so each can adapt its rendering (scroll viewport etc.)
func (a *appModel) broadcastSize(msg tea.WindowSizeMsg) {
	var panes []*layoutNode
	a.collectPanes(a.root, &panes)
	count := len(panes)
	if count == 0 {
		return
	}

	// Estimate inner size per pane. For simplicity, divide evenly.
	// The layout.render will handle exact clipping, but this gives
	// each model a reasonable width/height to work with.
	cols, rows := a.estimatePaneDimensions()

	for _, p := range panes {
		if p.pane != nil {
			m, _ := p.pane.m.Update(tea.WindowSizeMsg{Width: cols, Height: rows})
			p.pane.m = m
		}
	}
	_ = cols
}

// estimatePaneDimensions returns approximate (width, height) for a single pane
func (a *appModel) estimatePaneDimensions() (int, int) {
	contentH := a.height - 2 // status + command line
	if contentH < 4 {
		contentH = 4
	}
	// Walk the tree to count horizontal and vertical splits on the
	// deepest path to get a rough divisor
	hSplits, vSplits := a.countSplitDepth(a.root)

	w := a.width
	if vSplits > 0 {
		w = a.width / (vSplits + 1)
	}
	w -= 2 // border
	if w < 10 {
		w = 10
	}

	h := contentH
	if hSplits > 0 {
		h = contentH / (hSplits + 1)
	}
	h -= 2 // border
	if h < 4 {
		h = 4
	}

	return w, h
}

func (a *appModel) countSplitDepth(n *layoutNode) (hSplits, vSplits int) {
	if n == nil || n.kind == nodePane {
		return 0, 0
	}
	h1, v1 := a.countSplitDepth(n.first)
	h2, v2 := a.countSplitDepth(n.second)
	hMax := h1
	if h2 > hMax {
		hMax = h2
	}
	vMax := v1
	if v2 > vMax {
		vMax = v2
	}
	if n.kind == nodeSplitH {
		hMax++
	} else {
		vMax++
	}
	return hMax, vMax
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
		if a.activePane != nil && a.activePane.kind == nodePane {
			newFirst := splitNode(a.activePane, true)
			a.activePane = newFirst
			a.broadcastSize(tea.WindowSizeMsg{Width: a.width, Height: a.height})
			a.message = "Split vertical"
		}
	case "hsplit":
		if a.activePane != nil && a.activePane.kind == nodePane {
			newFirst := splitNode(a.activePane, false)
			a.activePane = newFirst
			a.broadcastSize(tea.WindowSizeMsg{Width: a.width, Height: a.height})
			a.message = "Split horizontal"
		}
	case "close":
		if a.activePane != nil && a.paneCount() > 1 {
			replacement := closePane(a.activePane)
			a.activePane = firstLeaf(replacement)
			a.broadcastSize(tea.WindowSizeMsg{Width: a.width, Height: a.height})
			a.message = "Pane closed"
		}
	case "help":
		a.message = "Commands: :q :vsplit :hsplit :close"
	default:
		a.message = fmt.Sprintf("Unknown command: %s", cmd)
	}
	return nil
}

func (a *appModel) cycleFocus() {
	next := a.findNextPane(a.root, a.activePane, false)
	if next != nil {
		a.activePane = next
		a.message = fmt.Sprintf("Pane %d", a.activePaneIndex()+1)
	}
}

func (a *appModel) cycleFocusReverse() {
	next := a.findNextPane(a.root, a.activePane, true)
	if next != nil {
		a.activePane = next
		a.message = fmt.Sprintf("Pane %d", a.activePaneIndex()+1)
	}
}

func (a *appModel) activePaneIndex() int {
	var panes []*layoutNode
	a.collectPanes(a.root, &panes)
	for i, p := range panes {
		if p == a.activePane {
			return i
		}
	}
	return 0
}

func (a *appModel) findNextPane(node *layoutNode, current *layoutNode, reverse bool) *layoutNode {
	var panes []*layoutNode
	a.collectPanes(node, &panes)

	if len(panes) <= 1 {
		return current
	}

	currentIdx := 0
	for i, p := range panes {
		if p == current {
			currentIdx = i
			break
		}
	}

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
	// Don't render until we have dimensions
	if a.width == 0 || a.height == 0 {
		return "  KUBEGRID :: Loading...\n"
	}

	var s strings.Builder

	// Calculate content height (reserve space for status and command line)
	contentHeight := a.height - 2
	if contentHeight < 4 {
		contentHeight = 4
	}

	// Render the layout tree
	if a.root != nil {
		content := a.root.render(a.width, contentHeight, a.activePane)
		s.WriteString(content)
		s.WriteString("\n")
	}

	// Status line
	paneIdx := a.activePaneIndex() + 1
	paneTotal := a.paneCount()
	statusLine := fmt.Sprintf(" KUBEGRID [%d/%d]", paneIdx, paneTotal)
	if a.message != "" {
		statusLine += fmt.Sprintf(" | %s", a.message)
	}
	s.WriteString(fmt.Sprintf("%-*s\n", a.width, statusLine))

	// Command line
	if a.currentMode == modeCommand {
		s.WriteString(fmt.Sprintf(":%s_", a.commandInput))
	} else {
		s.WriteString(" :cmd  ^B:vsplit  ^G:hsplit  ^X:close  Tab:focus  ^C:quit")
	}

	return s.String()
}
