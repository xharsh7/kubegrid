package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempKubeconfig(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestDiscoverKubeconfigs(t *testing.T) {
	// Save original home and restore after
	origHome, _ := os.UserHomeDir()
	t.Cleanup(func() {
		if origHome != "" {
			os.Setenv("HOME", origHome)
		}
	})

	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)

	// No .kube dir yet
	_, err := DiscoverKubeconfigs()
	if err == nil {
		t.Error("expected error with no .kube directory")
	}

	// Create .kube dir with a config file
	kubeDir := filepath.Join(tmpDir, ".kube")
	if err := os.MkdirAll(kubeDir, 0755); err != nil {
		t.Fatal(err)
	}

	// .kube/config should be discovered
	writeTempKubeconfig(t, kubeDir, "config", "")
	paths, err := DiscoverKubeconfigs()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(paths) != 1 {
		t.Fatalf("expected 1 config, got %d", len(paths))
	}

	// .kube/something.config should also be discovered
	writeTempKubeconfig(t, kubeDir, "prod.config", "")
	paths, err = DiscoverKubeconfigs()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(paths) != 2 {
		t.Fatalf("expected 2 configs, got %d", len(paths))
	}

	// .kube/random.yaml should NOT be discovered
	writeTempKubeconfig(t, kubeDir, "other.yaml", "")
	paths, _ = DiscoverKubeconfigs()
	if len(paths) != 2 {
		t.Fatalf("expected 2 configs (ignoring .yaml), got %d", len(paths))
	}
}

func TestLoadContexts(t *testing.T) {
	tmpDir := t.TempDir()

	validConfig := `apiVersion: v1
kind: Config
current-context: dev
contexts:
- context:
    cluster: minikube
    user: minikube
  name: dev
clusters:
- cluster:
    server: https://127.0.0.1:6443
  name: minikube
users:
- name: minikube
  user:
    token: fake
`

	invalidConfig := `garbage: yaml: [[[invalid`

	// Valid config
	validPath := writeTempKubeconfig(t, tmpDir, "dev.config", validConfig)
	contexts, err := LoadContexts([]string{validPath})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(contexts) != 1 {
		t.Fatalf("expected 1 context, got %d", len(contexts))
	}
	if contexts[0].Name != "dev" {
		t.Errorf("expected context name 'dev', got %q", contexts[0].Name)
	}
	if contexts[0].Cluster != "minikube" {
		t.Errorf("expected cluster 'minikube', got %q", contexts[0].Cluster)
	}

	// Invalid config should be skipped
	invalidPath := writeTempKubeconfig(t, tmpDir, "bad.config", invalidConfig)
	contexts, err = LoadContexts([]string{invalidPath})
	if err == nil {
		t.Error("expected error for invalid config")
	}

	// Mix of valid and invalid
	contexts, err = LoadContexts([]string{invalidPath, validPath})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(contexts) != 1 {
		t.Fatalf("expected 1 context, got %d", len(contexts))
	}

	// Empty paths
	_, err = LoadContexts([]string{})
	if err == nil {
		t.Error("expected error for empty paths")
	}
}
