package cluster

import (
	"sync"

	"github.com/xharsh7/kubegrid/internal/config"
)

func CollectStatuses(contexts []config.KubeContext) []ClusterStatus {
	var wg sync.WaitGroup
	results := make(chan ClusterStatus, len(contexts))

	for _, ctx := range contexts {
		wg.Add(1)
		go func(c config.KubeContext) {
			defer wg.Done()
			results <- CheckStatus(c)
		}(ctx)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var all []ClusterStatus
	for res := range results {
		all = append(all, res)
	}

	return all
}
