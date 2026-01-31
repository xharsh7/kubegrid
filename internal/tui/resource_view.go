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
	case tea.KeyMsg:
		if m.filtering {
			return m.handleFilterInput(msg)
		}

		if m.viewingLogs {
			return m.handleLogInput(msg)
		}

		switch msg.String() {
		case "q", "ctrl+c":
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
	case "up", "k":
		if m.logScroll > 0 {
			m.logScroll--
		}
	case "down", "j":
		m.logScroll++
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

	var s strings.Builder

	// Header
	s.WriteString("┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓\n")
	title := fmt.Sprintf("  KUBEGRID :: %s [%s] :: Namespace: %s",
		m.context.FriendlyName, m.resource.String(), m.namespace)
	s.WriteString(fmt.Sprintf("┃%-78s┃\n", title))
	s.WriteString("┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫\n")

	if m.loading {
		s.WriteString("┃  Loading...                                                                  ┃\n")
		s.WriteString("┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛\n")
		return s.String()
	}

	if m.error != nil {
		s.WriteString(fmt.Sprintf("┃  Error: %-68s┃\n", m.error.Error()))
		s.WriteString("┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛\n")
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
	s.WriteString("┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛\n")
	
	if m.filtering {
		s.WriteString(fmt.Sprintf("  Filter: %s_\n", m.filterInput))
	} else {
		s.WriteString("  1:Pods 2:Deploy 3:Svc 4:NS | /:Filter L:Logs D:Delete R:Refresh N:SwitchNS Q:Quit\n")
	}

	return s.String()
}

func (m resourceViewModel) renderPods() string {
	var s strings.Builder
	s.WriteString("┃   NAME                                      STATUS    READY  RESTARTS  AGE     ┃\n")
	s.WriteString("┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫\n")

	pods := m.getFilteredPods()
	for i, pod := range pods {
		cursor := " "
		if i == m.cursor {
			cursor = ">"
		}

		name := truncate(pod.Name, 40)
		status := truncate(pod.Status, 9)
		age := k8s.FormatAge(pod.Age)

		s.WriteString(fmt.Sprintf("┃ %s %-40s %-9s %-6s %-9d %-7s ┃\n",
			cursor, name, status, pod.Ready, pod.Restarts, age))
	}

	return s.String()
}

func (m resourceViewModel) renderDeployments() string {
	var s strings.Builder
	s.WriteString("┃   NAME                                      READY     UP-TO-DATE  AVAILABLE  ┃\n")
	s.WriteString("┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫\n")

	deployments := m.getFilteredDeployments()
	for i, dep := range deployments {
		cursor := " "
		if i == m.cursor {
			cursor = ">"
		}

		name := truncate(dep.Name, 40)

		s.WriteString(fmt.Sprintf("┃ %s %-40s %-9s %-11d %-10d ┃\n",
			cursor, name, dep.Ready, dep.UpToDate, dep.Available))
	}

	return s.String()
}

func (m resourceViewModel) renderServices() string {
	var s strings.Builder
	s.WriteString("┃   NAME                          TYPE            CLUSTER-IP        PORTS      ┃\n")
	s.WriteString("┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫\n")

	services := m.getFilteredServices()
	for i, svc := range services {
		cursor := " "
		if i == m.cursor {
			cursor = ">"
		}

		name := truncate(svc.Name, 30)
		svcType := truncate(svc.Type, 14)
		ip := truncate(svc.ClusterIP, 15)
		ports := truncate(svc.Ports, 11)

		s.WriteString(fmt.Sprintf("┃ %s %-30s %-14s %-15s %-11s ┃\n",
			cursor, name, svcType, ip, ports))
	}

	return s.String()
}

func (m resourceViewModel) renderNamespaces() string {
	var s strings.Builder
	s.WriteString("┃   NAME                                                STATUS      AGE        ┃\n")
	s.WriteString("┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫\n")

	namespaces := m.getFilteredNamespaces()
	for i, ns := range namespaces {
		cursor := " "
		if i == m.cursor {
			cursor = ">"
		}

		name := truncate(ns.Name, 50)
		status := truncate(ns.Status, 10)
		age := k8s.FormatAge(ns.Age)

		s.WriteString(fmt.Sprintf("┃ %s %-50s %-10s %-10s ┃\n",
			cursor, name, status, age))
	}

	return s.String()
}

func (m resourceViewModel) renderLogs() string {
	var s strings.Builder

	s.WriteString("┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓\n")
	
	if m.cursor < len(m.pods) {
		title := fmt.Sprintf("  LOGS :: %s", m.pods[m.cursor].Name)
		s.WriteString(fmt.Sprintf("┃%-78s┃\n", title))
	}
	
	s.WriteString("┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫\n")

	lines := strings.Split(m.logs, "\n")
	start := m.logScroll
	maxLines := 20

	for i := start; i < start+maxLines && i < len(lines); i++ {
		line := truncate(lines[i], 76)
		s.WriteString(fmt.Sprintf("┃ %-76s ┃\n", line))
	}

	s.WriteString("┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛\n")
	s.WriteString("  ↑↓:Scroll  Q/Esc:Back\n")

	return s.String()
}
