package k8s

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/yaml"
)

type Client struct {
	clientset     *kubernetes.Clientset
	dynamicClient dynamic.Interface
	namespace     string
}

func NewClient(kubeconfigPath string, namespace string) (*Client, error) {
	if namespace == "" {
		namespace = "default"
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to build config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}

	return &Client{
		clientset:     clientset,
		dynamicClient: dynamicClient,
		namespace:     namespace,
	}, nil
}

func (c *Client) SetNamespace(ns string) {
	c.namespace = ns
}

func (c *Client) GetNamespace() string {
	return c.namespace
}

// Pod represents a simplified pod with essential information
type Pod struct {
	Name      string
	Namespace string
	Status    string
	Ready     string
	Restarts  int32
	Age       time.Duration
	Node      string
}

// Deployment represents a simplified deployment
type Deployment struct {
	Name      string
	Namespace string
	Ready     string
	UpToDate  int32
	Available int32
	Age       time.Duration
}

// Service represents a simplified service
type Service struct {
	Name      string
	Namespace string
	Type      string
	ClusterIP string
	Ports     string
	Age       time.Duration
}

// Namespace represents a Kubernetes namespace
type Namespace struct {
	Name   string
	Status string
	Age    time.Duration
}

// Event represents a Kubernetes event
type Event struct {
	LastSeen time.Duration
	Type     string
	Reason   string
	Object   string
	Message  string
}

// CRDInfo represents a custom resource definition found via discovery
type CRDInfo struct {
	Name       string
	Kind       string
	Group      string
	Version    string
	Namespaced bool
}

// CRDInstance represents an instance of a custom resource
type CRDInstance struct {
	Name      string
	Namespace string
	Age       time.Duration
}

// ListPods retrieves all pods in the current namespace
func (c *Client) ListPods(ctx context.Context) ([]Pod, error) {
	pods, err := c.clientset.CoreV1().Pods(c.namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var result []Pod
	for _, p := range pods.Items {
		ready := 0
		total := len(p.Status.ContainerStatuses)
		restarts := int32(0)

		for _, cs := range p.Status.ContainerStatuses {
			if cs.Ready {
				ready++
			}
			restarts += cs.RestartCount
		}

		age := time.Since(p.CreationTimestamp.Time)

		result = append(result, Pod{
			Name:      p.Name,
			Namespace: p.Namespace,
			Status:    string(p.Status.Phase),
			Ready:     fmt.Sprintf("%d/%d", ready, total),
			Restarts:  restarts,
			Age:       age,
			Node:      p.Spec.NodeName,
		})
	}

	return result, nil
}

// ListDeployments retrieves all deployments in the current namespace
func (c *Client) ListDeployments(ctx context.Context) ([]Deployment, error) {
	deps, err := c.clientset.AppsV1().Deployments(c.namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var result []Deployment
	for _, d := range deps.Items {
		ready := fmt.Sprintf("%d/%d", d.Status.ReadyReplicas, *d.Spec.Replicas)
		age := time.Since(d.CreationTimestamp.Time)

		result = append(result, Deployment{
			Name:      d.Name,
			Namespace: d.Namespace,
			Ready:     ready,
			UpToDate:  d.Status.UpdatedReplicas,
			Available: d.Status.AvailableReplicas,
			Age:       age,
		})
	}

	return result, nil
}

// ListServices retrieves all services in the current namespace
func (c *Client) ListServices(ctx context.Context) ([]Service, error) {
	svcs, err := c.clientset.CoreV1().Services(c.namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var result []Service
	for _, s := range svcs.Items {
		ports := ""
		for i, p := range s.Spec.Ports {
			if i > 0 {
				ports += ","
			}
			ports += fmt.Sprintf("%d/%s", p.Port, p.Protocol)
		}

		age := time.Since(s.CreationTimestamp.Time)

		result = append(result, Service{
			Name:      s.Name,
			Namespace: s.Namespace,
			Type:      string(s.Spec.Type),
			ClusterIP: s.Spec.ClusterIP,
			Ports:     ports,
			Age:       age,
		})
	}

	return result, nil
}

// ListNamespaces retrieves all namespaces in the cluster
func (c *Client) ListNamespaces(ctx context.Context) ([]Namespace, error) {
	nsList, err := c.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var result []Namespace
	for _, ns := range nsList.Items {
		age := time.Since(ns.CreationTimestamp.Time)

		result = append(result, Namespace{
			Name:   ns.Name,
			Status: string(ns.Status.Phase),
			Age:    age,
		})
	}

	return result, nil
}

// ListEvents retrieves all events in the current namespace
func (c *Client) ListEvents(ctx context.Context) ([]Event, error) {
	events, err := c.clientset.CoreV1().Events(c.namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var result []Event
	for _, e := range events.Items {
		lastSeen := time.Since(e.LastTimestamp.Time)
		if e.LastTimestamp.IsZero() {
			lastSeen = time.Since(e.CreationTimestamp.Time)
		}
		result = append(result, Event{
			LastSeen: lastSeen,
			Type:     e.Type,
			Reason:   e.Reason,
			Object:   e.InvolvedObject.Kind + "/" + e.InvolvedObject.Name,
			Message:  e.Message,
		})
	}
	return result, nil
}

// ListCRDs discovers custom resources via the API discovery mechanism
func (c *Client) ListCRDs(ctx context.Context) ([]CRDInfo, error) {
	_, apiResources, err := c.clientset.Discovery().ServerGroupsAndResources()
	if err != nil {
		return nil, err
	}

	coreGroups := map[string]bool{
		"": true, "apps": true, "batch": true, "autoscaling": true,
		"policy": true, "rbac.authorization.k8s.io": true,
		"storage.k8s.io": true, "certificates.k8s.io": true,
		"networking.k8s.io": true, "node.k8s.io": true,
		"coordination.k8s.io": true, "events.k8s.io": true,
		"discovery.k8s.io": true, "scheduling.k8s.io": true,
		"apiextensions.k8s.io": true, "admissionregistration.k8s.io": true,
	}

	var crds []CRDInfo
	for _, apiRL := range apiResources {
		gv, err := schema.ParseGroupVersion(apiRL.GroupVersion)
		if err != nil {
			continue
		}
		if coreGroups[gv.Group] {
			continue
		}
		for _, r := range apiRL.APIResources {
			if !strings.Contains(r.Name, "/") {
				crds = append(crds, CRDInfo{
					Name:       r.Name,
					Kind:       r.Kind,
					Group:      gv.Group,
					Version:    gv.Version,
					Namespaced: r.Namespaced,
				})
			}
		}
	}
	return crds, nil
}

// ListCRDInstances lists all instances of a custom resource in the cluster
func (c *Client) ListCRDInstances(ctx context.Context, crd CRDInfo) ([]CRDInstance, error) {
	gvr := schema.GroupVersionResource{
		Group:    crd.Group,
		Version:  crd.Version,
		Resource: crd.Name,
	}

	var list *unstructured.UnstructuredList
	var err error

	if crd.Namespaced {
		list, err = c.dynamicClient.Resource(gvr).Namespace(c.namespace).List(ctx, metav1.ListOptions{})
	} else {
		list, err = c.dynamicClient.Resource(gvr).List(ctx, metav1.ListOptions{})
	}
	if err != nil {
		return nil, err
	}

	var result []CRDInstance
	for _, item := range list.Items {
		age := time.Since(item.GetCreationTimestamp().Time)
		result = append(result, CRDInstance{
			Name:      item.GetName(),
			Namespace: item.GetNamespace(),
			Age:       age,
		})
	}
	return result, nil
}

// GetResourceYAML returns YAML for any namespaced resource via the dynamic client
func (c *Client) GetResourceYAML(ctx context.Context, gvr schema.GroupVersionResource, name string) (string, error) {
	obj, err := c.dynamicClient.Resource(gvr).Namespace(c.namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	// Remove managed fields for readability
	delete(obj.Object, "managedFields")
	// Remove metadata managed fields
	if metadata, ok := obj.Object["metadata"].(map[string]interface{}); ok {
		delete(metadata, "managedFields")
	}

	yamlBytes, err := yaml.Marshal(obj)
	if err != nil {
		return "", err
	}
	return string(yamlBytes), nil
}

// GVRForType maps resource view types to GVR for describe operations
func GVRForResource(r string) (schema.GroupVersionResource, error) {
	switch r {
	case "pods":
		return schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}, nil
	case "deployments":
		return schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}, nil
	case "services":
		return schema.GroupVersionResource{Group: "", Version: "v1", Resource: "services"}, nil
	case "namespaces":
		return schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespaces"}, nil
	}
	return schema.GroupVersionResource{}, fmt.Errorf("unknown resource: %s", r)
}

// GetPodContainers returns container names for a pod
func (c *Client) GetPodContainers(ctx context.Context, podName string) ([]string, error) {
	pod, err := c.clientset.CoreV1().Pods(c.namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	var names []string
	for _, container := range pod.Spec.Containers {
		names = append(names, container.Name)
	}
	return names, nil
}

// GetPodLogs retrieves logs from a specific pod
func (c *Client) GetPodLogs(ctx context.Context, podName string, tailLines int64) (string, error) {
	opts := &corev1.PodLogOptions{
		TailLines: &tailLines,
	}

	req := c.clientset.CoreV1().Pods(c.namespace).GetLogs(podName, opts)
	stream, err := req.Stream(ctx)
	if err != nil {
		return "", err
	}
	defer stream.Close()

	logs, err := io.ReadAll(stream)
	if err != nil {
		return "", err
	}

	return string(logs), nil
}

// GetPodLogsStream streams logs from a pod with follow/container/timestamps options
func (c *Client) GetPodLogsStream(ctx context.Context, podName, container string, tailLines int64, timestamps bool) (io.ReadCloser, error) {
	opts := &corev1.PodLogOptions{
		TailLines:  &tailLines,
		Follow:     true,
		Timestamps: timestamps,
	}
	if container != "" {
		opts.Container = container
	}

	req := c.clientset.CoreV1().Pods(c.namespace).GetLogs(podName, opts)
	return req.Stream(ctx)
}

// DeletePod deletes a pod
func (c *Client) DeletePod(ctx context.Context, podName string) error {
	return c.clientset.CoreV1().Pods(c.namespace).Delete(ctx, podName, metav1.DeleteOptions{})
}

// FormatAge formats a duration into a human-readable age string
func FormatAge(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	return fmt.Sprintf("%dd", int(d.Hours()/24))
}
