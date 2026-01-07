package main

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/xharsh7/kubegrid/internal/cluster"
	"github.com/xharsh7/kubegrid/internal/config"
	"github.com/xharsh7/kubegrid/internal/tui"
)

func main() {
	paths, err := config.DiscoverKubeconfigs()
	if err != nil {
		panic(err)
	}

	contexts, err := config.LoadContexts(paths)
	if err != nil {
		panic(err)
	}

	refresh := func() []cluster.ClusterStatus {
		return cluster.CollectStatuses(contexts)
	}

	base := tui.NewClusterView(refresh(), refresh)
	app := tui.NewApp(base)

	if err := tea.NewProgram(app).Start(); err != nil {
		panic(err)
	}
}