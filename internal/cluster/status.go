package cluster

import (
	"time"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/kubernetes"

	"github.com/xharsh7/kubegrid/internal/config"
)

type ClusterStatus struct {
	Context   config.KubeContext
	Reachable bool
	Latency   time.Duration
	Error     error
}

func CheckStatus(ctx config.KubeContext) ClusterStatus {
	start := time.Now()

	restCfg, err := clientcmd.BuildConfigFromFlags("", ctx.Source)
	if err != nil {
		return ClusterStatus{Context: ctx, Error: err}
	}

	// Set timeout so unreachable or auth-broken clusters don't hang
	restCfg.Timeout = 5 * time.Second

	clientset, err := kubernetes.NewForConfig(restCfg)
	if err != nil {
		return ClusterStatus{Context: ctx, Error: err}
	}

	_, err = clientset.Discovery().ServerVersion()
	latency := time.Since(start)

	if err != nil {
		return ClusterStatus{
			Context:   ctx,
			Reachable: false,
			Latency:   latency,
			Error:     err,
		}
	}

	return ClusterStatus{
		Context:   ctx,
		Reachable: true,
		Latency:   latency,
	}
}
