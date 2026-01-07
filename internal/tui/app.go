package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type pane struct {
	m model
}

type appModel struct {
	panes  []pane
	active int
}

func NewApp(base model) appModel {
	return appModel{
		panes:  []pane{{m: base}},
		active: 0,
	}
}

func (a appModel) Init() tea.Cmd {
	return nil
}

func (a appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {

		case "ctrl+v":
			a.splitVertical()

		case "tab":
			a.active = (a.active + 1) % len(a.panes)
		}
	}

	// Forward input to active pane
	m, cmd := a.panes[a.active].m.Update(msg)
	a.panes[a.active].m = m.(model)

	return a, cmd
}

func (a appModel) View() string {
	if len(a.panes) == 1 {
		return a.panes[0].m.View()
	}

	left := a.panes[0].m.View()
	right := a.panes[1].m.View()

	return joinVertical(left, right)
}

func (a *appModel) splitVertical() {
	base := a.panes[a.active].m
	newPane := pane{m: base}
	a.panes = append(a.panes, newPane)
}

func joinVertical(a, b string) string {
	al := strings.Split(a, "\n")
	bl := strings.Split(b, "\n")

	max := len(al)
	if len(bl) > max {
		max = len(bl)
	}

	var out string
	for i := 0; i < max; i++ {
		var l, r string
		if i < len(al) {
			l = al[i]
		}
		if i < len(bl) {
			r = bl[i]
		}
		out += fmt.Sprintf("%-80s │ %s\n", l, r)
	}
	return out
}
