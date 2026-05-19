package tui

import (
	"bufio"
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
	resourceEvents
	resourceCRDs
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
	case resourceEvents:
		return "Events"
	case resourceCRDs:
		return "CRDs"
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
	events      []k8s.Event
	crds        []k8s.CRDInfo
	crdCursor   int
	crdScroll   int
	selectedCRD *k8s.CRDInfo
	crdInstances []k8s.CRDInstance

	// Log viewing
	viewingLogs bool
	logs        string
	logScroll   int
	logFollow   bool
	logFollowChan chan logLineMsg
	logFollowCancel context.CancelFunc
	logTimestamps bool
	logContainer  string
	containers    []string
	showContainerPicker bool
	containerCursor int
	tailLines     int64

	// Describe view
	viewingDescribe bool
	describeContent string

	wantBack bool
	width    int
	height   int

	listScroll int
}

type resourceLoadedMsg struct {
	resourceType resourceType
	pods         []k8s.Pod
	deployments  []k8s.Deployment
	services     []k8s.Service
	namespaces   []k8s.Namespace
	err          error
}

type logLineMsg struct {
	line string
	done bool
	err  error
}

type logsLoadedMsg struct {
	logs string
	err  error
}

type eventsLoadedMsg struct {
	events []k8s.Event
	err    error
}

type crdsLoadedMsg struct {
	crds []k8s.CRDInfo
	err  error
}

type crdInstancesLoadedMsg struct {
	instances []k8s.CRDInstance
	err       error
}

type describeLoadedMsg struct {
	content string
	err     error
}

type containersLoadedMsg struct {
	containers []string
	err        error
}

type podDeletedMsg struct {
	err error
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
		tailLines: 100,
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
		(&m).ensureListScrollVisible()
		return m, nil

	case tea.KeyMsg:
		if m.filtering {
			return m.handleFilterInput(msg)
		}

		if m.viewingDescribe {
			return m.handleDescribeInput(msg)
		}

		if m.showContainerPicker {
			return m.handleContainerPickerInput(msg)
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
			if m.selectedCRD != nil {
				if m.crdCursor > 0 {
					m.crdCursor--
				}
			} else if m.resource == resourceCRDs {
				if m.crdCursor > 0 {
					m.crdCursor--
				}
			} else {
				if m.cursor > 0 {
					m.cursor--
				}
			}

		case "down", "j":
			if m.selectedCRD != nil {
				max := len(m.crdInstances) - 1
				if m.crdCursor < max {
					m.crdCursor++
				}
			} else if m.resource == resourceCRDs {
				max := len(m.crds) - 1
				if m.crdCursor < max {
					m.crdCursor++
				}
			} else {
				max := m.getItemCount() - 1
				if m.cursor < max {
					m.cursor++
				}
			}

		case "g":
			if m.selectedCRD != nil {
				m.crdCursor = 0
			} else if m.resource == resourceCRDs {
				m.crdCursor = 0
			} else {
				m.cursor = 0
			}

		case "G":
			if m.selectedCRD != nil {
				n := len(m.crdInstances)
				if n == 0 {
					m.crdCursor = 0
				} else {
					m.crdCursor = n - 1
				}
			} else if m.resource == resourceCRDs {
				n := len(m.crds)
				if n == 0 {
					m.crdCursor = 0
				} else {
					m.crdCursor = n - 1
				}
			} else {
				if n := m.getItemCount(); n == 0 {
					m.cursor = 0
				} else {
					m.cursor = n - 1
				}
			}

		case "r":
			m.loading = true
			return m, m.loadResources()

		case "1":
			m.selectedCRD = nil
			m.resource = resourcePods
			m.cursor = 0
			m.loading = true
			return m, m.loadResources()

		case "2":
			m.selectedCRD = nil
			m.resource = resourceDeployments
			m.cursor = 0
			m.loading = true
			return m, m.loadResources()

		case "3":
			m.selectedCRD = nil
			m.resource = resourceServices
			m.cursor = 0
			m.loading = true
			return m, m.loadResources()

		case "4":
			m.selectedCRD = nil
			m.resource = resourceNamespaces
			m.cursor = 0
			m.loading = true
			return m, m.loadResources()

		case "5":
			m.selectedCRD = nil
			m.resource = resourceEvents
			m.cursor = 0
			m.loading = true
			return m, m.loadResources()

		case "6":
			m.selectedCRD = nil
			m.resource = resourceCRDs
			m.cursor = 0
			m.crdCursor = 0
			m.loading = true
			return m, m.loadResources()

		case "n":
			nsList := m.getFilteredNamespaces()
			if m.resource == resourceNamespaces && m.cursor < len(nsList) {
				newNs := nsList[m.cursor].Name
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

		case "enter":
			if m.resource == resourceCRDs && m.crdCursor < len(m.crds) {
				crd := m.crds[m.crdCursor]
				m.selectedCRD = &crd
				m.crdCursor = 0
				m.loading = true
				return m, m.loadCRDInstances(crd)
			}

		case "l":
			pods := m.getFilteredPods()
			if m.resource == resourcePods && m.cursor < len(pods) {
				pod := pods[m.cursor]
				m.loading = true
				return m, m.loadContainers(pod.Name)
			}

		case "y":
			if m.resource == resourcePods {
				pods := m.getFilteredPods()
				if m.cursor < len(pods) {
					m.loading = true
					return m, m.loadDescribe("pods", pods[m.cursor].Name)
				}
			} else if m.resource == resourceDeployments {
				deps := m.getFilteredDeployments()
				if m.cursor < len(deps) {
					m.loading = true
					return m, m.loadDescribe("deployments", deps[m.cursor].Name)
				}
			} else if m.resource == resourceServices {
				svcs := m.getFilteredServices()
				if m.cursor < len(svcs) {
					m.loading = true
					return m, m.loadDescribe("services", svcs[m.cursor].Name)
				}
			}

		case "d":
			pods := m.getFilteredPods()
			if m.resource == resourcePods && m.cursor < len(pods) {
				pod := pods[m.cursor]
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
		m.adjustCursor()

	case eventsLoadedMsg:
		m.loading = false
		m.error = msg.err
		if msg.err == nil {
			m.events = msg.events
		}
		m.adjustCursor()

	case crdsLoadedMsg:
		m.loading = false
		m.error = msg.err
		if msg.err == nil {
			m.crds = msg.crds
		}
		if len(m.crds) == 0 {
			m.crdCursor = 0
		} else if m.crdCursor >= len(m.crds) {
			m.crdCursor = len(m.crds) - 1
		}

	case crdInstancesLoadedMsg:
		m.loading = false
		m.error = msg.err
		if msg.err == nil {
			m.crdInstances = msg.instances
		}
		if len(m.crdInstances) == 0 {
			m.crdCursor = 0
		} else if m.crdCursor >= len(m.crdInstances) {
			m.crdCursor = len(m.crdInstances) - 1
		}

	case describeLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.error = msg.err
			m.viewingDescribe = false
		} else {
			m.error = nil
			m.viewingDescribe = true
			m.describeContent = msg.content
		}

	case containersLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.error = msg.err
		} else {
			m.containers = msg.containers
			if len(msg.containers) == 1 {
				m.logContainer = msg.containers[0]
				return m, m.loadPodLogs()
			}
			m.showContainerPicker = true
			m.containerCursor = 0
		}

	case logsLoadedMsg:
		if msg.err != nil {
			m.error = msg.err
			m.viewingLogs = false
			m.logs = ""
			m.logScroll = 0
		} else {
			m.error = nil
			m.viewingLogs = true
			m.logs = msg.logs
			m.logScroll = 0
		}

	case logLineMsg:
		if msg.err != nil {
			m.logFollow = false
			m.error = msg.err
			return m, nil
		}
		if msg.done {
			m.logFollow = false
			return m, nil
		}
		m.logs += msg.line + "\n"
		// Auto-scroll to bottom
		lines := strings.Split(m.logs, "\n")
		maxVisible := m.height - 5
		if maxVisible < 3 {
			maxVisible = 3
		}
		if len(lines) > maxVisible {
			m.logScroll = len(lines) - maxVisible
		}
		return m, m.waitForLogLine()

	case podDeletedMsg:
		if msg.err != nil {
			m.error = msg.err
		} else {
			m.loading = true
			return m, m.loadResources()
		}
	}

	(&m).ensureListScrollVisible()
	return m, nil
}

func (m resourceViewModel) adjustCursor() {
	if c := m.getItemCount(); c == 0 {
		m.cursor = 0
	} else if m.cursor >= c {
		m.cursor = c - 1
	}
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
		if len(msg.Runes) > 0 {
			m.filterInput += string(msg.Runes)
		}
	}
	if c := m.getItemCount(); c == 0 {
		m.cursor = 0
	} else if m.cursor >= c {
		m.cursor = c - 1
	}
	(&m).ensureListScrollVisible()
	return m, nil
}

func (m resourceViewModel) handleLogInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		if m.logFollow {
			if m.logFollowCancel != nil {
				m.logFollowCancel()
			}
			m.logFollow = false
		}
		m.viewingLogs = false
		m.logs = ""
		m.logScroll = 0
		(&m).ensureListScrollVisible()

	case "f":
		if !m.logFollow {
			pods := m.getFilteredPods()
			if m.cursor < len(pods) {
				m.logFollow = true
				m.logs = ""
				m.loading = true
				return m, m.startLogFollow(pods[m.cursor].Name)
			}
		}

	case "t":
		m.logTimestamps = !m.logTimestamps

	case "up", "k":
		if !m.logFollow {
			if m.logScroll > 0 {
				m.logScroll--
			}
		}

	case "down", "j":
		if !m.logFollow {
			lines := strings.Split(m.logs, "\n")
			maxVisible := m.height - 5
			if maxVisible < 3 {
				maxVisible = 3
			}
			maxScroll := len(lines) - maxVisible
			if maxScroll < 0 {
				maxScroll = 0
			}
			if m.logScroll < maxScroll {
				m.logScroll++
			}
		}
	}
	return m, nil
}

func (m resourceViewModel) handleContainerPickerInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.showContainerPicker = false
		m.containers = nil
	case "enter":
		if m.containerCursor >= 0 && m.containerCursor < len(m.containers) {
			m.logContainer = m.containers[m.containerCursor]
			m.showContainerPicker = false
			m.containers = nil
			return m, m.loadPodLogs()
		}
	case "up", "k":
		if m.containerCursor > 0 {
			m.containerCursor--
		}
	case "down", "j":
		if m.containerCursor < len(m.containers)-1 {
			m.containerCursor++
		}
	}
	return m, nil
}

func (m resourceViewModel) handleDescribeInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		m.viewingDescribe = false
		m.describeContent = ""
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
	case resourceEvents:
		return len(m.events)
	case resourceCRDs:
		if m.selectedCRD != nil {
			return len(m.crdInstances)
		}
		return len(m.crds)
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

func (m resourceViewModel) listVisibleRows() int {
	if m.height <= 0 {
		return 1
	}
	v := m.height - 7
	if v < 1 {
		v = 1
	}
	return v
}

func (m *resourceViewModel) ensureListScrollVisible() {
	n := m.getItemCount()
	vr := m.listVisibleRows()
	if n == 0 {
		m.listScroll = 0
		return
	}
	maxScroll := n - vr
	if maxScroll < 0 {
		maxScroll = 0
	}
	if m.cursor < m.listScroll {
		m.listScroll = m.cursor
	}
	if m.cursor >= m.listScroll+vr {
		m.listScroll = m.cursor - vr + 1
	}
	if m.listScroll < 0 {
		m.listScroll = 0
	}
	if m.listScroll > maxScroll {
		m.listScroll = maxScroll
	}
}

func (m resourceViewModel) loadResources() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		switch m.resource {
		case resourcePods:
			pods, err := m.client.ListPods(ctx)
			return resourceLoadedMsg{resourceType: m.resource, pods: pods, err: err}
		case resourceDeployments:
			deps, err := m.client.ListDeployments(ctx)
			return resourceLoadedMsg{resourceType: m.resource, deployments: deps, err: err}
		case resourceServices:
			svcs, err := m.client.ListServices(ctx)
			return resourceLoadedMsg{resourceType: m.resource, services: svcs, err: err}
		case resourceNamespaces:
			ns, err := m.client.ListNamespaces(ctx)
			return resourceLoadedMsg{resourceType: m.resource, namespaces: ns, err: err}
		case resourceEvents:
			events, err := m.client.ListEvents(ctx)
			return eventsLoadedMsg{events: events, err: err}
		case resourceCRDs:
			crds, err := m.client.ListCRDs(ctx)
			return crdsLoadedMsg{crds: crds, err: err}
		}
		return nil
	}
}

func (m resourceViewModel) loadCRDInstances(crd k8s.CRDInfo) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		instances, err := m.client.ListCRDInstances(ctx, crd)
		return crdInstancesLoadedMsg{instances: instances, err: err}
	}
}

func (m resourceViewModel) loadDescribe(resource, name string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		gvr, err := k8s.GVRForResource(resource)
		if err != nil {
			return describeLoadedMsg{err: err}
		}

		yaml, err := m.client.GetResourceYAML(ctx, gvr, name)
		return describeLoadedMsg{content: yaml, err: err}
	}
}

func (m resourceViewModel) loadContainers(podName string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		containers, err := m.client.GetPodContainers(ctx, podName)
		return containersLoadedMsg{containers: containers, err: err}
	}
}

func (m resourceViewModel) loadPodLogs() tea.Cmd {
	pods := m.getFilteredPods()
	if m.cursor >= len(pods) {
		return nil
	}

	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		logs, err := m.client.GetPodLogs(ctx, pods[m.cursor].Name, m.tailLines)
		return logsLoadedMsg{logs: logs, err: err}
	}
}

func (m resourceViewModel) startLogFollow(podName string) tea.Cmd {
	ctx, cancel := context.WithCancel(context.Background())
	m.logFollowCancel = cancel
	ch := make(chan logLineMsg, 100)
	m.logFollowChan = ch

	go func() {
		defer close(ch)

		stream, err := m.client.GetPodLogsStream(ctx, podName, m.logContainer, m.tailLines, m.logTimestamps)
		if err != nil {
			ch <- logLineMsg{err: err}
			return
		}
		defer stream.Close()

		scanner := bufio.NewScanner(stream)
		for scanner.Scan() {
			select {
			case ch <- logLineMsg{line: scanner.Text()}:
			case <-ctx.Done():
				return
			}
		}
	}()

	return m.waitForLogLine()
}

func (m resourceViewModel) waitForLogLine() tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-m.logFollowChan
		if !ok {
			return logLineMsg{done: true}
		}
		return msg
	}
}

func (m resourceViewModel) deletePod(podName string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err := m.client.DeletePod(ctx, podName)
		return podDeletedMsg{err: err}
	}
}

// View renders the current view
func (m resourceViewModel) View() string {
	if m.viewingDescribe {
		return m.renderDescribe()
	}

	if m.viewingLogs {
		if m.logFollow {
			return m.renderLogFollow()
		}
		return m.renderLogs()
	}

	if m.showContainerPicker {
		return m.renderContainerPicker()
	}

	w := m.width
	if w < 20 {
		w = 80
	}
	inner := w
	hline := strings.Repeat("━", inner)

	var s strings.Builder

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

	switch m.resource {
	case resourcePods:
		s.WriteString(m.renderPods())
	case resourceDeployments:
		s.WriteString(m.renderDeployments())
	case resourceServices:
		s.WriteString(m.renderServices())
	case resourceNamespaces:
		s.WriteString(m.renderNamespaces())
	case resourceEvents:
		s.WriteString(m.renderEvents())
	case resourceCRDs:
		if m.selectedCRD != nil {
			s.WriteString(m.renderCRDInstances())
		} else {
			s.WriteString(m.renderCRDs())
		}
	}

	s.WriteString("┗" + hline + "┛\n")

	if m.filtering {
		s.WriteString(fmt.Sprintf("  Filter: %s_\n", m.filterInput))
	} else if inner >= 60 {
		s.WriteString("  1:Pods 2:Dply 3:Svc 4:NS 5:Evnt 6:CRD | /:Flt L:Logs D:Del Y:Desc R:Ref N:NS Esc:Back\n")
	} else {
		s.WriteString("  1-6:Res /:Flt L:Log D:Del Y:Desc R:Ref Esc:Back\n")
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

	showRestarts := inner >= 70
	showAge := inner >= 50

	statusW := 9
	readyW := 5
	restartsW := 4
	ageW := 5

	fixed := 4 + statusW + readyW
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
	n := len(pods)
	vr := m.listVisibleRows()
	start := m.listScroll
	if start < 0 {
		start = 0
	}
	if n > 0 && start > n-1 {
		start = n - 1
	}
	end := start + vr
	if end > n {
		end = n
	}

	linesOut := 0
	for rowIdx := start; rowIdx < end; rowIdx++ {
		pod := pods[rowIdx]
		cursor := " "
		if rowIdx == m.cursor {
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
		linesOut++
	}
	for linesOut < vr {
		s.WriteString(fmt.Sprintf("┃%-*s┃\n", inner, ""))
		linesOut++
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
	n := len(deployments)
	vr := m.listVisibleRows()
	start := m.listScroll
	if start < 0 {
		start = 0
	}
	if n > 0 && start > n-1 {
		start = n - 1
	}
	end := start + vr
	if end > n {
		end = n
	}

	linesOut := 0
	for rowIdx := start; rowIdx < end; rowIdx++ {
		dep := deployments[rowIdx]
		cursor := " "
		if rowIdx == m.cursor {
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
		linesOut++
	}
	for linesOut < vr {
		s.WriteString(fmt.Sprintf("┃%-*s┃\n", inner, ""))
		linesOut++
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
	n := len(services)
	vr := m.listVisibleRows()
	start := m.listScroll
	if start < 0 {
		start = 0
	}
	if n > 0 && start > n-1 {
		start = n - 1
	}
	end := start + vr
	if end > n {
		end = n
	}

	linesOut := 0
	for rowIdx := start; rowIdx < end; rowIdx++ {
		svc := services[rowIdx]
		cursor := " "
		if rowIdx == m.cursor {
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
		linesOut++
	}
	for linesOut < vr {
		s.WriteString(fmt.Sprintf("┃%-*s┃\n", inner, ""))
		linesOut++
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
	n := len(namespaces)
	vr := m.listVisibleRows()
	start := m.listScroll
	if start < 0 {
		start = 0
	}
	if n > 0 && start > n-1 {
		start = n - 1
	}
	end := start + vr
	if end > n {
		end = n
	}

	linesOut := 0
	for rowIdx := start; rowIdx < end; rowIdx++ {
		ns := namespaces[rowIdx]
		cursor := " "
		if rowIdx == m.cursor {
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
		linesOut++
	}
	for linesOut < vr {
		s.WriteString(fmt.Sprintf("┃%-*s┃\n", inner, ""))
		linesOut++
	}

	return s.String()
}

func (m resourceViewModel) renderEvents() string {
	w := m.width
	if w < 20 {
		w = 80
	}
	inner := w
	hline := strings.Repeat("━", inner)

	showMsg := inner >= 70

	typeW := 8
	reasonW := 12
	objectW := 20
	ageW := 8

	fixed := 4 + typeW + reasonW + objectW + ageW
	nameW := inner - fixed
	if nameW < 8 {
		nameW = 8
	}

	var s strings.Builder
	colHeader := fmt.Sprintf("   %-*s %-*s %-*s %-*s %-*s",
		ageW, "LAST SEEN", typeW, "TYPE", reasonW, "REASON", objectW, "OBJECT", inner-fixed+nameW, "MESSAGE")
	s.WriteString(fmt.Sprintf("┃%-*s┃\n", inner, colHeader))
	s.WriteString("┣" + hline + "┫\n")

	events := m.events
	n := len(events)
	vr := m.listVisibleRows()
	start := m.listScroll
	if start < 0 {
		start = 0
	}
	if n > 0 && start > n-1 {
		start = n - 1
	}
	end := start + vr
	if end > n {
		end = n
	}

	linesOut := 0
	for rowIdx := start; rowIdx < end; rowIdx++ {
		e := events[rowIdx]
		cursor := " "
		if rowIdx == m.cursor {
			cursor = ">"
		}

		age := k8s.FormatAge(e.LastSeen)
		etype := truncate(e.Type, typeW)
		reason := truncate(e.Reason, reasonW)
		obj := truncate(e.Object, objectW)
		msg := e.Message
		if !showMsg && len(msg) > inner-fixed+nameW {
			msg = msg[:inner-fixed+nameW]
		} else if showMsg && len(msg) > inner-fixed+nameW {
			msg = msg[:inner-fixed+nameW]
		}

		row := fmt.Sprintf(" %s %-*s %-*s %-*s %-*s %-*s",
			cursor, ageW, age, typeW, etype, reasonW, reason, objectW, obj, inner-fixed+nameW, msg)
		s.WriteString(fmt.Sprintf("┃%-*s┃\n", inner, row))
		linesOut++
	}
	for linesOut < vr {
		s.WriteString(fmt.Sprintf("┃%-*s┃\n", inner, ""))
		linesOut++
	}

	return s.String()
}

func (m resourceViewModel) renderCRDs() string {
	w := m.width
	if w < 20 {
		w = 80
	}
	inner := w
	hline := strings.Repeat("━", inner)

	kindW := 16
	groupW := 20
	versionW := 8
	scopeW := 5

	fixed := 4 + kindW + groupW + versionW + scopeW
	nameW := inner - fixed
	if nameW < 8 {
		nameW = 8
	}

	var s strings.Builder
	colHeader := fmt.Sprintf("   %-*s %-*s %-*s %-*s %-*s",
		nameW, "RESOURCE", kindW, "KIND", groupW, "GROUP", versionW, "VERSION", scopeW, "SCOPE")
	s.WriteString(fmt.Sprintf("┃%-*s┃\n", inner, colHeader))
	s.WriteString("┣" + hline + "┫\n")

	crds := m.crds
	n := len(crds)
	vr := m.listVisibleRows()
	start := m.listScroll
	if start < 0 {
		start = 0
	}
	if n > 0 && start > n-1 {
		start = n - 1
	}
	end := start + vr
	if end > n {
		end = n
	}

	linesOut := 0
	for rowIdx := start; rowIdx < end; rowIdx++ {
		crd := crds[rowIdx]
		cursor := " "
		if rowIdx == m.crdCursor {
			cursor = ">"
		}

		rname := truncate(crd.Name, nameW)
		kind := truncate(crd.Kind, kindW)
		group := truncate(crd.Group, groupW)
		version := truncate(crd.Version, versionW)
		scope := "NS"
		if !crd.Namespaced {
			scope = "CL"
		}

		row := fmt.Sprintf(" %s %-*s %-*s %-*s %-*s %-*s",
			cursor, nameW, rname, kindW, kind, groupW, group, versionW, version, scopeW, scope)
		s.WriteString(fmt.Sprintf("┃%-*s┃\n", inner, row))
		linesOut++
	}
	for linesOut < vr {
		s.WriteString(fmt.Sprintf("┃%-*s┃\n", inner, ""))
		linesOut++
	}

	return s.String()
}

func (m resourceViewModel) renderCRDInstances() string {
	w := m.width
	if w < 20 {
		w = 80
	}
	inner := w
	hline := strings.Repeat("━", inner)

	var s strings.Builder
	title := fmt.Sprintf("  %s :: %s instances", m.selectedCRD.Kind, m.selectedCRD.Name)
	s.WriteString(fmt.Sprintf("┃%-*s┃\n", inner, title))
	s.WriteString("┣" + hline + "┫\n")

	showNamespace := m.selectedCRD.Namespaced
	showAge := inner >= 40

	ageW := 5
	fixed := 4
	nsW := 0
	if showNamespace {
		nsW = 16
		fixed += 1 + nsW
	}
	if showAge {
		fixed += 1 + ageW
	}
	nameW := inner - fixed
	if nameW < 8 {
		nameW = 8
	}

	var colHeader string
	if showNamespace && showAge {
		colHeader = fmt.Sprintf("   %-*s %-*s %-*s", nameW, "NAME", nsW, "NAMESPACE", ageW, "AGE")
	} else if showNamespace {
		colHeader = fmt.Sprintf("   %-*s %-*s", nameW, "NAME", nsW, "NAMESPACE")
	} else if showAge {
		colHeader = fmt.Sprintf("   %-*s %-*s", nameW, "NAME", ageW, "AGE")
	} else {
		colHeader = fmt.Sprintf("   %-*s", nameW, "NAME")
	}
	s.WriteString(fmt.Sprintf("┃%-*s┃\n", inner, colHeader))
	s.WriteString("┣" + hline + "┫\n")

	instances := m.crdInstances
	n := len(instances)
	vr := m.listVisibleRows()
	start := m.listScroll
	if start < 0 {
		start = 0
	}
	if n > 0 && start > n-1 {
		start = n - 1
	}
	end := start + vr
	if end > n {
		end = n
	}

	linesOut := 0
	for rowIdx := start; rowIdx < end; rowIdx++ {
		inst := instances[rowIdx]
		cursor := " "
		if rowIdx == m.crdCursor {
			cursor = ">"
		}

		name := truncate(inst.Name, nameW)
		ns := truncate(inst.Namespace, nsW)
		age := truncate(k8s.FormatAge(inst.Age), ageW)

		var row string
		if showNamespace && showAge {
			row = fmt.Sprintf(" %s %-*s %-*s %-*s", cursor, nameW, name, nsW, ns, ageW, age)
		} else if showNamespace {
			row = fmt.Sprintf(" %s %-*s %-*s", cursor, nameW, name, nsW, ns)
		} else if showAge {
			row = fmt.Sprintf(" %s %-*s %-*s", cursor, nameW, name, ageW, age)
		} else {
			row = fmt.Sprintf(" %s %-*s", cursor, nameW, name)
		}
		s.WriteString(fmt.Sprintf("┃%-*s┃\n", inner, row))
		linesOut++
	}
	for linesOut < vr {
		s.WriteString(fmt.Sprintf("┃%-*s┃\n", inner, ""))
		linesOut++
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
	contentW := inner - 2
	if contentW < 1 {
		contentW = 1
	}

	var s strings.Builder
	s.WriteString("┏" + hline + "┓\n")

	pods := m.getFilteredPods()
	if m.cursor < len(pods) {
		pod := pods[m.cursor]
		title := fmt.Sprintf("  LOGS :: %s", pod.Name)
		if m.logContainer != "" {
			title += fmt.Sprintf(" [%s]", m.logContainer)
		}
		if m.logTimestamps {
			title += " [timestamps]"
		}
		if len(title) > inner {
			title = title[:inner]
		}
		s.WriteString(fmt.Sprintf("┃%-*s┃\n", inner, title))
	}
	s.WriteString("┣" + hline + "┫\n")

	lines := strings.Split(m.logs, "\n")
	maxLines := m.height - 5
	if maxLines < 3 {
		maxLines = 3
	}

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

	shown := len(lines) - start
	if shown > maxLines {
		shown = maxLines
	}
	for i := shown; i < maxLines; i++ {
		s.WriteString(fmt.Sprintf("┃ %-*s ┃\n", contentW, ""))
	}

	s.WriteString("┗" + hline + "┛\n")
	s.WriteString("  ↑↓:Scroll F:Follow T:Timestamps Esc:Back\n")

	return s.String()
}

func (m resourceViewModel) renderLogFollow() string {
	w := m.width
	if w < 20 {
		w = 80
	}
	inner := w
	hline := strings.Repeat("━", inner)
	contentW := inner - 2
	if contentW < 1 {
		contentW = 1
	}

	var s strings.Builder
	s.WriteString("┏" + hline + "┓\n")

	title := "  LOGS :: FOLLOWING"
	if m.logContainer != "" {
		title += fmt.Sprintf(" [%s]", m.logContainer)
	}
	if m.logTimestamps {
		title += " [timestamps]"
	}
	if len(title) > inner {
		title = title[:inner]
	}
	s.WriteString(fmt.Sprintf("┃%-*s┃\n", inner, title))
	s.WriteString("┣" + hline + "┫\n")

	lines := strings.Split(m.logs, "\n")
	maxLines := m.height - 5
	if maxLines < 3 {
		maxLines = 3
	}

	start := len(lines) - maxLines
	if start < 0 {
		start = 0
	}

	for i := start; i < start+maxLines && i < len(lines); i++ {
		line := truncate(lines[i], contentW)
		s.WriteString(fmt.Sprintf("┃ %-*s ┃\n", contentW, line))
	}

	s.WriteString("┗" + hline + "┛\n")
	s.WriteString("  Esc:Stop Follow\n")

	return s.String()
}

func (m resourceViewModel) renderDescribe() string {
	w := m.width
	if w < 20 {
		w = 80
	}
	inner := w
	hline := strings.Repeat("━", inner)
	contentW := inner - 2
	if contentW < 1 {
		contentW = 1
	}

	var s strings.Builder
	s.WriteString("┏" + hline + "┓\n")
	s.WriteString(fmt.Sprintf("┃%-*s┃\n", inner, "  RESOURCE YAML"))
	s.WriteString("┣" + hline + "┫\n")

	lines := strings.Split(m.describeContent, "\n")
	maxLines := m.height - 4
	if maxLines < 3 {
		maxLines = 3
	}

	for i := 0; i < maxLines && i < len(lines); i++ {
		line := truncate(lines[i], contentW)
		s.WriteString(fmt.Sprintf("┃ %-*s ┃\n", contentW, line))
	}

	s.WriteString("┗" + hline + "┛\n")
	s.WriteString("  Esc:Back\n")

	return s.String()
}

func (m resourceViewModel) renderContainerPicker() string {
	w := m.width
	if w < 20 {
		w = 80
	}
	inner := w
	hline := strings.Repeat("━", inner)

	var s strings.Builder
	s.WriteString("┏" + hline + "┓\n")
	s.WriteString(fmt.Sprintf("┃%-*s┃\n", inner, "  SELECT CONTAINER"))
	s.WriteString("┣" + hline + "┫\n")

	nameW := inner - 4
	if nameW < 1 {
		nameW = 1
	}

	for i, container := range m.containers {
		cursor := " "
		if i == m.containerCursor {
			cursor = ">"
		}
		name := truncate(container, nameW)
		row := fmt.Sprintf(" %s %-*s", cursor, nameW, name)
		s.WriteString(fmt.Sprintf("┃%-*s┃\n", inner, row))
	}

	s.WriteString("┗" + hline + "┛\n")
	s.WriteString("  ↑↓:Nav Enter:Select Esc:Cancel\n")

	return s.String()
}
