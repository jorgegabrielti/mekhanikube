package k8s

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// Option configures the Client.
type Option func(*config)

type config struct {
	kubeconfigPath string
	context        string
}

// WithKubeconfig sets the path to a kubeconfig file.
func WithKubeconfig(path string) Option {
	return func(c *config) {
		c.kubeconfigPath = path
	}
}

// WithContext sets the Kubernetes context to use.
func WithContext(name string) Option {
	return func(c *config) {
		c.context = name
	}
}

// NewClient creates a Kubernetes client with agnostic kubeconfig detection.
// It tries multiple strategies in order: in-cluster, explicit path, KUBECONFIG env, default path.
func NewClient(opts ...Option) (kubernetes.Interface, error) {
	cfg := &config{}
	for _, opt := range opts {
		opt(cfg)
	}

	// Strategy 1: In-cluster config
	restConfig, err := rest.InClusterConfig()
	if err == nil {
		slog.Debug("using in-cluster configuration")
		restConfig.WarningHandler = rest.NoWarnings{}
		return kubernetes.NewForConfig(restConfig)
	}

	// Strategy 2: Explicit kubeconfig path (from --kubeconfig flag)
	if cfg.kubeconfigPath != "" {
		restConfig, err := buildConfig(cfg.kubeconfigPath, cfg.context)
		if err == nil {
			slog.Debug("using explicit kubeconfig", "path", cfg.kubeconfigPath)
			return kubernetes.NewForConfig(restConfig)
		}
	}

	// Strategy 3: KUBECONFIG environment variable
	if envPath := os.Getenv("KUBECONFIG"); envPath != "" {
		restConfig, err := buildConfig(envPath, cfg.context)
		if err == nil {
			slog.Debug("using KUBECONFIG env", "path", envPath)
			return kubernetes.NewForConfig(restConfig)
		}
	}

	// Strategy 4: Default ~/.kube/config
	if home := homedir.HomeDir(); home != "" {
		defaultPath := filepath.Join(home, ".kube", "config")
		if _, err := os.Stat(defaultPath); err == nil {
			restConfig, err := buildConfig(defaultPath, cfg.context)
			if err == nil {
				slog.Debug("using default kubeconfig", "path", defaultPath)
				return kubernetes.NewForConfig(restConfig)
			}
		}
	}

	return nil, fmt.Errorf("no kubeconfig found (tried: in-cluster, explicit path, KUBECONFIG env, ~/.kube/config)")
}

func buildConfig(kubeconfigPath, context string) (*rest.Config, error) {
	rules := &clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath}
	overrides := &clientcmd.ConfigOverrides{}
	if context != "" {
		overrides.CurrentContext = context
	}

	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, overrides).ClientConfig()
	if err != nil {
		return nil, err
	}

	config.WarningHandler = rest.NoWarnings{}
	return config, nil
}
