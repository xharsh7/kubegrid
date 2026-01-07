package tui

import (
	"strings"
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
	old := *target
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

func (n *layoutNode) render(width, height int) string {
	if n.kind == nodePane {
		view := n.pane.m.View()
		lines := strings.Split(view, "\n")

		var out []string
		for i := 0; i < height; i++ {
			line := ""
			if i < len(lines) {
				line = lines[i]
			}
			if len(line) > width {
				line = line[:width]
			}
			out = append(out, padRight(line, width))
		}
		return strings.Join(out, "\n")
	}

	if n.kind == nodeSplitV {
		w1 := width / 2
		w2 := width - w1 - 1

		left := strings.Split(n.first.render(w1, height), "\n")
		right := strings.Split(n.second.render(w2, height), "\n")

		var out []string
		for i := 0; i < height; i++ {
			out = append(out, left[i]+"│"+right[i])
		}
		return strings.Join(out, "\n")
	}

	// Horizontal split
	h1 := height / 2
	h2 := height - h1 - 1

	top := strings.Split(n.first.render(width, h1), "\n")
	bottom := strings.Split(n.second.render(width, h2), "\n")

	bar := strings.Repeat("─", width)
	return strings.Join(append(append(top, bar), bottom...), "\n")
}

func padRight(s string, w int) string {
	if len(s) >= w {
		return s
	}
	return s + strings.Repeat(" ", w-len(s))
}
