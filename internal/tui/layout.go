package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
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
	parent *layoutNode
}

func newPaneNode(p pane) *layoutNode {
	return &layoutNode{kind: nodePane, pane: &p}
}

func splitNode(target *layoutNode, vertical bool) *layoutNode {
	if target.kind != nodePane {
		return target
	}

	// Clone the existing pane's model for the new sibling
	oldPane := pane{m: target.pane.m}
	newPane := pane{m: target.pane.m}

	first := &layoutNode{kind: nodePane, pane: &oldPane, parent: target}
	second := &layoutNode{kind: nodePane, pane: &newPane, parent: target}

	if vertical {
		target.kind = nodeSplitV
	} else {
		target.kind = nodeSplitH
	}

	target.first = first
	target.second = second
	target.pane = nil

	return first // return the first child so caller can update activePane
}

// closePane removes a pane from the tree: its sibling replaces the parent split.
// Returns the sibling node that took the parent's place.
func closePane(target *layoutNode) *layoutNode {
	if target.parent == nil {
		// root pane, can't close
		return target
	}

	parent := target.parent
	var sibling *layoutNode
	if parent.first == target {
		sibling = parent.second
	} else {
		sibling = parent.first
	}

	// Replace parent with sibling
	parent.kind = sibling.kind
	parent.pane = sibling.pane
	parent.first = sibling.first
	parent.second = sibling.second
	if parent.first != nil {
		parent.first.parent = parent
	}
	if parent.second != nil {
		parent.second.parent = parent
	}

	return parent
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
		// Calculate inner dimensions (account for borders: 1 char each side)
		innerWidth := width - 2
		innerHeight := height - 2
		if innerWidth < 1 {
			innerWidth = 1
		}
		if innerHeight < 1 {
			innerHeight = 1
		}

		// Update pane model with actual dimensions before rendering
		m, _ := n.pane.m.Update(tea.WindowSizeMsg{Width: innerWidth, Height: innerHeight})
		n.pane.m = m

		// Render the pane content
		view := n.pane.m.View()
		lines := strings.Split(view, "\n")

		var contentLines []string
		for i := 0; i < innerHeight; i++ {
			line := ""
			if i < len(lines) {
				line = lines[i]
			}
			// Use display-width-aware truncation (handles multi-byte box-drawing chars)
			if runewidth.StringWidth(line) > innerWidth {
				line = runewidth.Truncate(line, innerWidth, "")
			}
			contentLines = append(contentLines, padRight(line, innerWidth))
		}

		content := strings.Join(contentLines, "\n")

		style := inactiveStyle
		if n == activePane {
			style = activeStyle
		}

		return style.Width(innerWidth).Height(innerHeight).Render(content)
	}

	if n.kind == nodeSplitV {
		// Vertical split: side by side (like tmux Ctrl-b %)
		w1 := width / 2
		w2 := width - w1

		left := n.first.render(w1, height, activePane)
		right := n.second.render(w2, height, activePane)

		return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
	}

	// Horizontal split: top and bottom (like tmux Ctrl-b ")
	h1 := height / 2
	h2 := height - h1

	top := n.first.render(width, h1, activePane)
	bottom := n.second.render(width, h2, activePane)

	return lipgloss.JoinVertical(lipgloss.Left, top, bottom)
}

// firstLeaf returns the deepest first leaf pane in this subtree
func firstLeaf(n *layoutNode) *layoutNode {
	if n == nil {
		return nil
	}
	if n.kind == nodePane {
		return n
	}
	return firstLeaf(n.first)
}

func padRight(s string, w int) string {
	sw := runewidth.StringWidth(s)
	if sw >= w {
		return s
	}
	return s + strings.Repeat(" ", w-sw)
}
