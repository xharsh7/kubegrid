package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type helpModel struct {
	width  int
	height int
}

func NewHelpView() helpModel {
	return helpModel{}
}

func (h helpModel) Init() tea.Cmd {
	return nil
}

func (h helpModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h.width = msg.Width
		h.height = msg.Height
	case tea.KeyMsg:
		// Help view is typically read-only, parent handles navigation
	}
	return h, nil
}

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("170")).
			MarginTop(1).
			MarginBottom(1)

	sectionStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("99")).
			MarginTop(1)

	keyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true)

	descStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("246"))
)

func (h helpModel) View() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("╔═══════════════════════════════════════════════════════════════╗"))
	s.WriteString("\n")
	s.WriteString(titleStyle.Render("║              KUBEGRID - HELP & KEYBINDINGS                  ║"))
	s.WriteString("\n")
	s.WriteString(titleStyle.Render("╚═══════════════════════════════════════════════════════════════╝"))
	s.WriteString("\n\n")

	// Navigation
	s.WriteString(sectionStyle.Render("📍 NAVIGATION"))
	s.WriteString("\n")
	s.WriteString(helpLine("↑/k", "Move cursor up"))
	s.WriteString(helpLine("↓/j", "Move cursor down"))
	s.WriteString(helpLine("g", "Jump to top"))
	s.WriteString(helpLine("G", "Jump to bottom"))
	s.WriteString(helpLine("Tab", "Cycle focus between panes"))
	s.WriteString(helpLine("Shift+Tab", "Cycle focus backwards"))
	s.WriteString("\n")

	// Resource Views
	s.WriteString(sectionStyle.Render("📦 RESOURCE VIEWS"))
	s.WriteString("\n")
	s.WriteString(helpLine("1", "View Pods"))
	s.WriteString(helpLine("2", "View Deployments"))
	s.WriteString(helpLine("3", "View Services"))
	s.WriteString(helpLine("4", "View Namespaces"))
	s.WriteString(helpLine("5", "View Events"))
	s.WriteString(helpLine("6", "View Custom Resources (CRDs)"))
	s.WriteString(helpLine("Enter", "Inspect selected cluster / CRD instances"))
	s.WriteString(helpLine("Esc", "Go back to list view"))
	s.WriteString("\n")

	// Actions
	s.WriteString(sectionStyle.Render("⚡ ACTIONS"))
	s.WriteString("\n")
	s.WriteString(helpLine("r", "Refresh current view"))
	s.WriteString(helpLine("l", "View pod logs (select container if multi-container)"))
	s.WriteString(helpLine("y", "View resource YAML / describe"))
	s.WriteString(helpLine("d", "Delete selected pod"))
	s.WriteString(helpLine("n", "Switch namespace (from namespace list)"))
	s.WriteString(helpLine("/", "Filter/search resources"))
	s.WriteString("\n")

	// Log Viewing
	s.WriteString(sectionStyle.Render("📜 LOG VIEWING"))
	s.WriteString("\n")
	s.WriteString(helpLine("f", "Start/stop log follow mode (streaming)"))
	s.WriteString(helpLine("t", "Toggle timestamps"))
	s.WriteString(helpLine("↑/k", "Scroll up"))
	s.WriteString(helpLine("↓/j", "Scroll down"))
	s.WriteString(helpLine("Esc", "Close log view"))
	s.WriteString("\n")

	// Window Management
	s.WriteString(sectionStyle.Render("🪟 WINDOW MANAGEMENT"))
	s.WriteString("\n")
	s.WriteString(helpLine("Ctrl+B", "Split pane vertically (side by side)"))
	s.WriteString(helpLine("Ctrl+G", "Split pane horizontally (top/bottom)"))
	s.WriteString(helpLine("Ctrl+X", "Close active pane"))
	s.WriteString(helpLine("Tab", "Cycle focus to next pane"))
	s.WriteString(helpLine("Shift+Tab", "Cycle focus to prev pane"))
	s.WriteString("\n")

	// Commands
	s.WriteString(sectionStyle.Render("⌨️  COMMAND MODE (Vim-style)"))
	s.WriteString("\n")
	s.WriteString(helpLine(":", "Enter command mode"))
	s.WriteString(helpLine(":q or :quit", "Quit application"))
	s.WriteString(helpLine(":vsplit", "Split vertically"))
	s.WriteString(helpLine(":hsplit", "Split horizontally"))
	s.WriteString(helpLine(":close", "Close active pane"))
	s.WriteString(helpLine(":help", "Show available commands"))
	s.WriteString("\n")

	// General
	s.WriteString(sectionStyle.Render("🔧 GENERAL"))
	s.WriteString("\n")
	s.WriteString(helpLine("?", "Toggle help"))
	s.WriteString(helpLine("q", "Quit (from most views)"))
	s.WriteString(helpLine("Ctrl+C", "Force quit"))
	s.WriteString("\n")

	// Tips
	s.WriteString(sectionStyle.Render("💡 TIPS"))
	s.WriteString("\n")
	s.WriteString(descStyle.Render("  • Use splits to monitor multiple clusters simultaneously"))
	s.WriteString("\n")
	s.WriteString(descStyle.Render("  • Filter with / to quickly find resources"))
	s.WriteString("\n")
	s.WriteString(descStyle.Render("  • Switch namespaces: press 4 → navigate → press n"))
	s.WriteString("\n")
	s.WriteString(descStyle.Render("  • Press l on a pod to view logs; press f to follow"))
	s.WriteString("\n")
	s.WriteString(descStyle.Render("  • Press y on a resource to view its YAML describe"))
	s.WriteString("\n")
	s.WriteString(descStyle.Render("  • Press 6 to browse Custom Resources (CRDs)"))
	s.WriteString("\n\n")

	s.WriteString(descStyle.Render("Press any key to close help..."))

	return s.String()
}

func helpLine(key, description string) string {
	return fmt.Sprintf("  %s  %s\n",
		keyStyle.Render(fmt.Sprintf("%-15s", key)),
		descStyle.Render(description))
}
