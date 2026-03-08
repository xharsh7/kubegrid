package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/xharsh7/kubegrid/internal/config"
	"github.com/xharsh7/kubegrid/pkg/k8s"
)

type resourceType int

const (
	resourcePods resourceType = iota
	resourceDeployments
	resourceServices
	resourceNamespaces
)

func (r resourceType) String() string {
	switch r {
	case resourcePods:
		return "Pods"
	case resourceDeployments:
		return "Deployments"
	case resourceServices:
		return "Services"
	case resourceNamespaces:
		return "Namespaces"
	default:
		return "Unknown"
	}
}

type resourceViewModel struct {
	context     config.KubeContext
	client      *k8s.Client
	resource    resourceType
	namespace   string
	cursor      int
	filterInput string
	filtering   bool
	loading     bool
	error       error

	// Data
	pods        []k8s.Pod
	deployments []k8s.Deployment
	services    []k8s.Service
	namespaces  []k8s.Namespace

	// Log viewing
	viewingLogs bool
	logs        string
	logScroll   int

	wantBack bool
	width    int
	height   int
}

type resourceLoadedMsg struct {
	resourceType resourceType
	pods         []k8s.Pod
	deployments  []k8s.Deployment
	services     []k8s.Service
	namespaces   []k8s.Namespace
	err          error
}

func NewResourceView(ctx config.KubeContext, namespace string) (resourceViewModel, error) {
	client, err := k8s.NewClient(ctx.Source, namespace)
	if err != nil {
		return resourceViewModel{}, err
	}

	m := resourceViewModel{
		context:   ctx,
		client:    client,
		resource:  resourcePods,
		namespace: namespace,
	}

	return m, nil
}

func (m resourceViewModel) Init() tea.Cmd {
	return m.loadResources()
}

func (m resourceViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		if m.filtering {
			return m.handleFilterInput(msg)
		}

		if m.viewingLogs {
			return m.handleLogInput(msg)
		}

		switch msg.String() {
		case "esc":
			m.wantBack = true
			return m, nil
		case "ctrl+c":
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			max := m.getItemCount() - 1
			if m.cursor < max {
				m.cursor++
			}

		case "g":
			m.cursor = 0

		case "G":
			m.cursor = m.getItemCount() - 1

		case "r":
			m.loading = true
			return m, m.loadResources()

		case "1":
			m.resource = resourcePods
			m.cursor = 0
			m.loading = true
			return m, m.loadResources()

		case "2":
			m.resource = resourceDeployments
			m.cursor = 0
			m.loading = true
			return m, m.loadResources()

		case "3":
			m.resource = resourceServices
			m.cursor = 0
			m.loading = true
			return m, m.loadResources()

		case "4":
			m.resource = resourceNamespaces
			m.cursor = 0
			m.loading = true
			return m, m.loadResources()

		case "n":
			// Switch namespace (cycle through available namespaces)
			if m.resource == resourceNamespaces && m.cursor < len(m.namespaces) {
				newNs := m.namespaces[m.cursor].Name
				m.client.SetNamespace(newNs)
				m.namespace = newNs
				m.resource = resourcePods
				m.cursor = 0
				m.loading = true
				return m, m.loadResources()
			}

		case "/":
			m.filtering = true
			m.filterInput = ""

		case "l":
			// View logs for selected pod
			if m.resource == resourcePods && m.cursor < len(m.pods) {
				pod := m.pods[m.cursor]
				return m, m.loadPodLogs(pod.Name)
			}

		case "d":
			// Delete selected pod
			if m.resource == resourcePods && m.cursor < len(m.pods) {
				pod := m.pods[m.cursor]
				return m, m.deletePod(pod.Name)
			}
		}

	case resourceLoadedMsg:
		m.loading = false
		m.error = msg.err
		if msg.err == nil {
			m.pods = msg.pods
			m.deployments = msg.deployments
			m.services = msg.services
			m.namespaces = msg.namespaces
		}

	case logsLoadedMsg:
		m.viewingLogs = true
		m.logs = msg.logs
		m.logScroll = 0

	case podDeletedMsg:
		if msg.err != nil {
			m.error = msg.err
		} else {
			m.loading = true
			return m, m.loadResources()
		}
	}

	return m, nil
}

func (m resourceViewModel) handleFilterInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.filtering = false
		m.filterInput = ""
	case "enter":
		m.filtering = false
	case "backspace":
		if len(m.filterInput) > 0 {
			m.filterInput = m.filterInput[:len(m.filterInput)-1]
		}
	default:
		m.filterInput += msg.String()
	}
	return m, nil
}

func (m resourceViewModel) handleLogInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		m.viewingLogs = false
		m.logs = ""
		m.logScroll = 0
	case "up", "k":
		if m.logScroll > 0 {
			m.logScroll--
		}
	case "down", "j":
		// Clamp scroll to max
		lines := strings.Split(m.logs, "\n")
		maxVisible := m.height - 6
		if maxVisible < 5 {
			maxVisible = 5
		}
		maxScroll := len(lines) - maxVisible
		if maxScroll < 0 {
			maxScroll = 0
		}
		if m.logScroll < maxScroll {
			m.logScroll++
		}
	}
	return m, nil
}

func (m resourceViewModel) getItemCount() int {
	switch m.resource {
	case resourcePods:
		return len(m.getFilteredPods())
	case resourceDeployments:
		return len(m.getFilteredDeployments())
	case resourceServices:
		return len(m.getFilteredServices())
	case resourceNamespaces:
		return len(m.getFilteredNamespaces())
	default:
		return 0
	}
}

func (m resourceViewModel) getFilteredPods() []k8s.Pod {
	if m.filterInput == "" {
		return m.pods
	}
	var filtered []k8s.Pod
	for _, p := range m.pods {
		if strings.Contains(strings.ToLower(p.Name), strings.ToLower(m.filterInput)) {
			filtered = append(filtered, p)
		}
	}
	return filtered
}

func (m resourceViewModel) getFilteredDeployments() []k8s.Deployment {
	if m.filterInput == "" {
		return m.deployments
	}
	var filtered []k8s.Deployment
	for _, d := range m.deployments {
		if strings.Contains(strings.ToLower(d.Name), strings.ToLower(m.filterInput)) {
			filtered = append(filtered, d)
		}
	}
	return filtered
}

func (m resourceViewModel) getFilteredServices() []k8s.Service {
	if m.filterInput == "" {
		return m.services
	}
	var filtered []k8s.Service
	for _, s := range m.services {
		if strings.Contains(strings.ToLower(s.Name), strings.ToLower(m.filterInput)) {
			filtered = append(filtered, s)
		}
	}
	return filtered
}

func (m resourceViewModel) getFilteredNamespaces() []k8s.Namespace {
	if m.filterInput == "" {
		return m.namespaces
	}
	var filtered []k8s.Namespace
	for _, ns := range m.namespaces {
		if strings.Contains(strings.ToLower(ns.Name), strings.ToLower(m.filterInput)) {
			filtered = append(filtered, ns)
		}
	}
	return filtered
}

func (m resourceViewModel) loadResources() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var pods []k8s.Pod
		var deployments []k8s.Deployment
		var services []k8s.Service
		var namespaces []k8s.Namespace
		var err error

		switch m.resource {
		case resourcePods:
			pods, err = m.client.ListPods(ctx)
		case resourceDeployments:
			deployments, err = m.client.ListDeployments(ctx)
		case resourceServices:
			services, err = m.client.ListServices(ctx)
		case resourceNamespaces:
			namespaces, err = m.client.ListNamespaces(ctx)
		}

		return resourceLoadedMsg{
			resourceType: m.resource,
			pods:         pods,
			deployments:  deployments,
			services:     services,
			namespaces:   namespaces,
			err:          err,
		}
	}
}

type logsLoadedMsg struct {
	logs string
	err  error
}

func (m resourceViewModel) loadPodLogs(podName string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		logs, err := m.client.GetPodLogs(ctx, podName, 100)
		return logsLoadedMsg{logs: logs, err: err}
	}
}

type podDeletedMsg struct {
	err error
}

func (m resourceViewModel) deletePod(podName string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err := m.client.DeletePod(ctx, podName)
		return podDeletedMsg{err: err}
	}
}

func (m resourceViewModel) View() string {
	if m.viewingLogs {
		return m.renderLogs()
	}

	w := m.width
	if w < 20 {
		w = 80
	}
	inner := w
	hline := strings.Repeat("━", inner)

	var s strings.Builder

	// Header
	s.WriteString("┏" + hline + "┓\n")
	title := fmt.Sprintf("  %s [%s] ns:%s",
		m.context.FriendlyName, m.resource.String(), m.namespace)
	if inner >= 60 {
		title = fmt.Sprintf("  KUBEGRID :: %s [%s] :: Namespace: %s",
			m.context.FriendlyName, m.resource.String(), m.namespace)
	}
	if len(title) > inner {
		title = title[:inner]
	}
	s.WriteString(fmt.Sprintf("┃%-*s┃\n", inner, title))
	s.WriteString("┣" + hline + "┫\n")

	if m.loading {
		s.WriteString(fmt.Sprintf("┃  Loading...%-*s┃\n", inner-12, ""))
		s.WriteString("┗" + hline + "┛\n")
		return s.String()
	}

	if m.error != nil {
		errStr := fmt.Sprintf("  Error: %s", m.error.Error())
		if len(errStr) > inner {
			errStr = errStr[:inner]
		}
		s.WriteString(fmt.Sprintf("┃%-*s┃\n", inner, errStr))
		s.WriteString("┗" + hline + "┛\n")
		return s.String()
	}

	// Resource-specific rendering
	switch m.resource {
	case resourcePods:
		s.WriteString(m.renderPods())
	case resourceDeployments:
		s.WriteString(m.renderDeployments())
	case resourceServices:
		s.WriteString(m.renderServices())
	case resourceNamespaces:
		s.WriteString(m.renderNamespaces())
	}

	// Footer
	s.WriteString("┗" + hline + "┛\n")

	if m.filtering {
		s.WriteString(fmt.Sprintf("  Filter: %s_\n", m.filterInput))
	} else if inner >= 60 {
		s.WriteString("  1:Pods 2:Deploy 3:Svc 4:NS | /:Filter L:Logs D:Del R:Ref N:NS Esc:Back\n")
	} else {
		s.WriteString("  1-4:Res /:Flt L:Log D:Del R:Ref Esc:Back\n")
	}

	return s.String()
}

func (m resourceViewModel) renderPods() string {
	w := m.width
	if w < 20 {
		w = 80
	}
	inner := w
	hline := strings.Repeat("━", inner)

	// Adaptive columns: hide less important ones in narrow panels
	showRestarts := inner >= 70
	showAge := inner >= 50

	statusW := 9
	readyW := 5
	restartsW := 4
	ageW := 5

	fixed := 4 + statusW + readyW // " > " + status + ready
	if showRestarts {
		fixed += 1 + restartsW
	}
	if showAge {
		fixed += 1 + ageW
	}
	nameW := inner - fixed
	if nameW < 8 {
		nameW = 8
	}

	var s strings.Builder
	var colHeader string
	if showRestarts && showAge {
		colHeader = fmt.Sprintf("   %-*s %-*s %-*s %-*s %-*s",
			nameW, "NAME", statusW, "STATUS", readyW, "READY", restartsW, "RST", ageW, "AGE")
	} else if showAge {
		colHeader = fmt.Sprintf("   %-*s %-*s %-*s %-*s",
			nameW, "NAME", statusW, "STATUS", readyW, "READY", ageW, "AGE")
	} else {
		colHeader = fmt.Sprintf("   %-*s %-*s %-*s",
			nameW, "NAME", statusW, "STATUS", readyW, "READY")
	}
	s.WriteString(fmt.Sprintf("┃%-*s┃\n", inner, colHeader))
	s.WriteString("┣" + hline + "┫\n")

	pods := m.getFilteredPods()
	for i, pod := range pods {
		cursor := " "
		if i == m.cursor {
			cursor = ">"
		}

		name := truncate(pod.Name, nameW)
		status := truncate(pod.Status, statusW)
		age := truncate(k8s.FormatAge(pod.Age), ageW)

		var row string
		if showRestarts && showAge {
			row = fmt.Sprintf(" %s %-*s %-*s %-*s %-*d %-*s",
				cursor, nameW, name, statusW, status, readyW, pod.Ready, restartsW, pod.Restarts, ageW, age)
		} else if showAge {
			row = fmt.Sprintf(" %s %-*s %-*s %-*s %-*s",
				cursor, nameW, name, statusW, status, readyW, pod.Ready, ageW, age)
		} else {
			row = fmt.Sprintf(" %s %-*s %-*s %-*s",
				cursor, nameW, name, statusW, status, readyW, pod.Ready)
		}
		s.WriteString(fmt.Sprintf("┃%-*s┃\n", inner, row))
	}

	return s.String()
}

func (m resourceViewModel) renderDeployments() string {
	w := m.width
	if w < 20 {
		w = 80
	}
	inner := w
	hline := strings.Repeat("━", inner)

	showAvail := inner >= 60
	showUpToDate := inner >= 50

	readyW := 7
	upToDateW := 4
	availableW := 5

	fixed := 4 + readyW
	if showUpToDate {
		fixed += 1 + upToDateW
	}
	if showAvail {
		fixed += 1 + availableW
	}
	nameW := inner - fixed
	if nameW < 8 {
		nameW = 8
	}

	var s strings.Builder
	var colHeader string
	if showUpToDate && showAvail {
		colHeader = fmt.Sprintf("   %-*s %-*s %-*s %-*s",
			nameW, "NAME", readyW, "READY", upToDateW, "UTD", availableW, "AVAIL")
	} else if showUpToDate {
		colHeader = fmt.Sprintf("   %-*s %-*s %-*s",
			nameW, "NAME", readyW, "READY", upToDateW, "UTD")
	} else {
		colHeader = fmt.Sprintf("   %-*s %-*s",
			nameW, "NAME", readyW, "READY")
	}
	s.WriteString(fmt.Sprintf("┃%-*s┃\n", inner, colHeader))
	s.WriteString("┣" + hline + "┫\n")

	deployments := m.getFilteredDeployments()
	for i, dep := range deployments {
		cursor := " "
		if i == m.cursor {
			cursor = ">"
		}

		name := truncate(dep.Name, nameW)

		var row string
		if showUpToDate && showAvail {
			row = fmt.Sprintf(" %s %-*s %-*s %-*d %-*d",
				cursor, nameW, name, readyW, dep.Ready, upToDateW, dep.UpToDate, availableW, dep.Available)
		} else if showUpToDate {
			row = fmt.Sprintf(" %s %-*s %-*s %-*d",
				cursor, nameW, name, readyW, dep.Ready, upToDateW, dep.UpToDate)
		} else {
			row = fmt.Sprintf(" %s %-*s %-*s",
				cursor, nameW, name, readyW, dep.Ready)
		}
		s.WriteString(fmt.Sprintf("┃%-*s┃\n", inner, row))
	}

	return s.String()
}

func (m resourceViewModel) renderServices() string {
	w := m.width
	if w < 20 {
		w = 80
	}
	inner := w
	hline := strings.Repeat("━", inner)

	showIP := inner >= 60
	showPorts := inner >= 45

	typeW := 10
	ipW := 15
	portsW := 10

	fixed := 4 + typeW
	if showIP {
		fixed += 1 + ipW
	}
	if showPorts {
		fixed += 1 + portsW
	}
	nameW := inner - fixed
	if nameW < 8 {
		nameW = 8
	}

	var s strings.Builder
	var colHeader string
	if showIP && showPorts {
		colHeader = fmt.Sprintf("   %-*s %-*s %-*s %-*s",
			nameW, "NAME", typeW, "TYPE", ipW, "CLUSTER-IP", portsW, "PORTS")
	} else if showPorts {
		colHeader = fmt.Sprintf("   %-*s %-*s %-*s",
			nameW, "NAME", typeW, "TYPE", portsW, "PORTS")
	} else {
		colHeader = fmt.Sprintf("   %-*s %-*s",
			nameW, "NAME", typeW, "TYPE")
	}
	s.WriteString(fmt.Sprintf("┃%-*s┃\n", inner, colHeader))
	s.WriteString("┣" + hline + "┫\n")

	services := m.getFilteredServices()
	for i, svc := range services {
		cursor := " "
		if i == m.cursor {
			cursor = ">"
		}

		name := truncate(svc.Name, nameW)
		svcType := truncate(svc.Type, typeW)

		var row string
		if showIP && showPorts {
			ip := truncate(svc.ClusterIP, ipW)
			ports := truncate(svc.Ports, portsW)
			row = fmt.Sprintf(" %s %-*s %-*s %-*s %-*s",
				cursor, nameW, name, typeW, svcType, ipW, ip, portsW, ports)
		} else if showPorts {
			ports := truncate(svc.Ports, portsW)
			row = fmt.Sprintf(" %s %-*s %-*s %-*s",
				cursor, nameW, name, typeW, svcType, portsW, ports)
		} else {
			row = fmt.Sprintf(" %s %-*s %-*s",
				cursor, nameW, name, typeW, svcType)
		}
		s.WriteString(fmt.Sprintf("┃%-*s┃\n", inner, row))
	}

	return s.String()
}

func (m resourceViewModel) renderNamespaces() string {
	w := m.width
	if w < 20 {
		w = 80
	}
	inner := w
	hline := strings.Repeat("━", inner)

	showAge := inner >= 40

	statusW := 8
	ageW := 5

	fixed := 4 + statusW
	if showAge {
		fixed += 1 + ageW
	}
	nameW := inner - fixed
	if nameW < 8 {
		nameW = 8
	}

	var s strings.Builder
	var colHeader string
	if showAge {
		colHeader = fmt.Sprintf("   %-*s %-*s %-*s",
			nameW, "NAME", statusW, "STATUS", ageW, "AGE")
	} else {
		colHeader = fmt.Sprintf("   %-*s %-*s",
			nameW, "NAME", statusW, "STATUS")
	}
	s.WriteString(fmt.Sprintf("┃%-*s┃\n", inner, colHeader))
	s.WriteString("┣" + hline + "┫\n")

	namespaces := m.getFilteredNamespaces()
	for i, ns := range namespaces {
		cursor := " "
		if i == m.cursor {
			cursor = ">"
		}

		name := truncate(ns.Name, nameW)
		status := truncate(ns.Status, statusW)

		var row string
		if showAge {
			age := truncate(k8s.FormatAge(ns.Age), ageW)
			row = fmt.Sprintf(" %s %-*s %-*s %-*s",
				cursor, nameW, name, statusW, status, ageW, age)
		} else {
			row = fmt.Sprintf(" %s %-*s %-*s",
				cursor, nameW, name, statusW, status)
		}
		s.WriteString(fmt.Sprintf("┃%-*s┃\n", inner, row))
	}

	return s.String()
}

func (m resourceViewModel) renderLogs() string {
	w := m.width
	if w < 20 {
		w = 80
	}
	inner := w
	hline := strings.Repeat("━", inner)
	contentW := inner - 2 // space for " " padding on each side
	if contentW < 1 {
		contentW = 1
	}

	var s strings.Builder

	s.WriteString("┏" + hline + "┓\n")

	if m.cursor < len(m.pods) {
		title := fmt.Sprintf("  LOGS :: %s", m.pods[m.cursor].Name)
		if len(title) > inner {
			title = title[:inner]
		}
		s.WriteString(fmt.Sprintf("┃%-*s┃\n", inner, title))
	}

	s.WriteString("┣" + hline + "┫\n")

	lines := strings.Split(m.logs, "\n")
	// chrome: top border + title + separator + bottom border + footer = 5 lines
	maxLines := m.height - 5
	if maxLines < 3 {
		maxLines = 3
	}

	// Clamp scroll
	maxScroll := len(lines) - maxLines
	if maxScroll < 0 {
		maxScroll = 0
	}
	start := m.logScroll
	if start > maxScroll {
		start = maxScroll
	}

	for i := start; i < start+maxLines && i < len(lines); i++ {
		line := truncate(lines[i], contentW)
		s.WriteString(fmt.Sprintf("┃ %-*s ┃\n", contentW, line))
	}

	// Pad if fewer lines than maxLines
	shown := len(lines) - start
	if shown > maxLines {
		shown = maxLines
	}
	for i := shown; i < maxLines; i++ {
		s.WriteString(fmt.Sprintf("┃ %-*s ┃\n", contentW, ""))
	}

	s.WriteString("┗" + hline + "┛\n")
	s.WriteString("  ↑↓:Scroll  Esc:Back\n")

	return s.String()
}
