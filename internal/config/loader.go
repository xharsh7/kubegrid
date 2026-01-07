package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"k8s.io/client-go/tools/clientcmd"
)

type KubeContext struct {
	Name         string
	Cluster      string
	User         string
	Source       string
	FriendlyName string
}

func DiscoverKubeconfigs() ([]string, error) {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".kube")

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var configs []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.HasSuffix(e.Name(), ".config") || e.Name() == "config" {
			configs = append(configs, filepath.Join(dir, e.Name()))
		}
	}

	if len(configs) == 0 {
		return nil, fmt.Errorf("no kubeconfigs found")
	}

	return configs, nil
}

func LoadContexts(paths []string) ([]KubeContext, error) {
	var contexts []KubeContext

	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		cfg, err := clientcmd.Load(data)
		if err != nil {
			continue
		}

		file := filepath.Base(path)
		friendly := strings.TrimSuffix(file, ".config")

		for name, ctx := range cfg.Contexts {
			contexts = append(contexts, KubeContext{
				Name:         name,
				Cluster:      ctx.Cluster,
				User:         ctx.AuthInfo,
				Source:       path,
				FriendlyName: friendly,
			})
		}
	}

	if len(contexts) == 0 {
		return nil, fmt.Errorf("no contexts found in configs")
	}

	return contexts, nil
}
