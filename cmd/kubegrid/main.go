package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/xharsh7/kubegrid/internal/cluster"
	"github.com/xharsh7/kubegrid/internal/config"
	"github.com/xharsh7/kubegrid/internal/tui"
)

func main() {
	paths, err := config.DiscoverKubeconfigs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error discovering kubeconfigs: %v\n", err)
		os.Exit(1)
	}

	contexts, err := config.LoadContexts(paths)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading contexts: %v\n", err)
		os.Exit(1)
	}

	refresh := func() []cluster.ClusterStatus {
		return cluster.CollectStatuses(contexts)
	}

	base := tui.NewClusterView(refresh(), refresh)
	app := tui.NewApp(base)

	p := tea.NewProgram(
		app,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}