package k8s

import (
	"os"
	"path/filepath"
	"sort"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// ContextInfo holds metadata about a single kubeconfig context.
type ContextInfo struct {
	Name      string
	Cluster   string
	User      string
	IsCurrent bool
}

// ListContexts returns all contexts found in the active kubeconfig.
// It respects KUBECONFIG env, then falls back to ~/.kube/config.
// Returns a nil slice (no error) when no kubeconfig file is found.
func ListContexts() ([]ContextInfo, error) {
	kubeconfigPath := os.Getenv("KUBECONFIG")
	if kubeconfigPath == "" {
		if home := homedir.HomeDir(); home != "" {
			kubeconfigPath = filepath.Join(home, ".kube", "config")
		}
	}
	if kubeconfigPath == "" {
		return nil, nil
	}

	rules := &clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath}
	cfg, err := rules.Load()
	if err != nil {
		return nil, err
	}

	current := cfg.CurrentContext
	var contexts []ContextInfo
	for name, ctx := range cfg.Contexts {
		contexts = append(contexts, ContextInfo{
			Name:      name,
			Cluster:   ctx.Cluster,
			User:      ctx.AuthInfo,
			IsCurrent: name == current,
		})
	}

	sort.Slice(contexts, func(i, j int) bool {
		return contexts[i].Name < contexts[j].Name
	})
	return contexts, nil
}
