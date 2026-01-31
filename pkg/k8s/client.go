package k8s

import (
	"context"
	"fmt"
	"io"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// Client wraps a Kubernetes clientset with helper methods
type Client struct {
	clientset *kubernetes.Clientset
	namespace string
}

// NewClient creates a new Kubernetes client from a kubeconfig path
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

	return &Client{
		clientset: clientset,
		namespace: namespace,
	}, nil
}

// SetNamespace changes the active namespace
func (c *Client) SetNamespace(ns string) {
	c.namespace = ns
}

// GetNamespace returns the current namespace
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

// GetPodLogsStream streams logs from a pod
func (c *Client) GetPodLogsStream(ctx context.Context, podName string, follow bool) (io.ReadCloser, error) {
	opts := &corev1.PodLogOptions{
		Follow: follow,
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
