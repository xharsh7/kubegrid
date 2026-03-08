package main

import (
	"fmt"
	"io"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"k8s.io/klog/v2"

	"github.com/xharsh7/kubegrid/internal/cluster"
	"github.com/xharsh7/kubegrid/internal/config"
	"github.com/xharsh7/kubegrid/internal/tui"
)

func init() {
	// Suppress k8s client-go logging to prevent TUI corruption
	klog.SetOutput(io.Discard)

	// Redirect stderr to /dev/null — k8s auth plugins (SSO, exec-based)
	// write errors directly to stderr which corrupts the TUI alt-screen
	devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err == nil {
		os.Stderr = devNull
	}
}

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