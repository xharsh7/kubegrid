package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type nodeType int

const (
	nodePane nodeType = iota
	nodeSplitV
	nodeSplitH
)

type layoutNode struct {
	kind   nodeType
	pane   *pane
	first  *layoutNode
	second *layoutNode
}

func newPaneNode(p pane) *layoutNode {
	return &layoutNode{kind: nodePane, pane: &p}
}

func splitNode(target *layoutNode, vertical bool) *layoutNode {
	if target.kind != nodePane {
		return target
	}

	// Create a copy of the current pane
	old := *target
	
	// Transform this node into a split
	if vertical {
		target.kind = nodeSplitV
	} else {
		target.kind = nodeSplitH
	}
	
	target.first = &old
	target.second = newPaneNode(*old.pane)
	target.pane = nil
	
	return target
}

var (
	activeStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62"))
	
	inactiveStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240"))
)

func (n *layoutNode) render(width, height int, activePane *layoutNode) string {
	if n == nil {
		return ""
	}

	if n.kind == nodePane {
		// Render the pane content
		view := n.pane.m.View()
		lines := strings.Split(view, "\n")

		// Calculate inner dimensions (account for borders)
		innerWidth := width - 2
		innerHeight := height - 2
		if innerWidth < 1 {
			innerWidth = 1
		}
		if innerHeight < 1 {
			innerHeight = 1
		}

		var contentLines []string
		for i := 0; i < innerHeight; i++ {
			line := ""
			if i < len(lines) {
				line = lines[i]
			}
			if len(line) > innerWidth {
				line = line[:innerWidth]
			}
			contentLines = append(contentLines, padRight(line, innerWidth))
		}

		content := strings.Join(contentLines, "\n")

		// Apply border style based on active state
		style := inactiveStyle
		if n == activePane {
			style = activeStyle
		}

		return style.Width(innerWidth).Height(innerHeight).Render(content)
	}

	if n.kind == nodeSplitV {
		// Vertical split: side by side
		w1 := width / 2
		w2 := width - w1

		left := n.first.render(w1, height, activePane)
		right := n.second.render(w2, height, activePane)

		return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
	}

	// Horizontal split: top and bottom
	h1 := height / 2
	h2 := height - h1

	top := n.first.render(width, h1, activePane)
	bottom := n.second.render(width, h2, activePane)

	return lipgloss.JoinVertical(lipgloss.Left, top, bottom)
}

func padRight(s string, w int) string {
	if len(s) >= w {
		return s
	}
	return s + strings.Repeat(" ", w-len(s))
}
