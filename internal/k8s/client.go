package k8s

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

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
	// Set a reasonable timeout so credential-provider plugins (e.g. AWS STS)
	// don't hang indefinitely when tokens are expired or SSO login is missing.
	if config.Timeout == 0 {
		config.Timeout = connectionTimeout
	}
	return config, nil
}

// connectionTimeout is the maximum time to wait for cluster connectivity check.
const connectionTimeout = 5 * time.Second

// CheckConnection performs a lightweight health check against the API server
// to verify that the cluster is reachable and credentials are valid.
// It returns a user-friendly error when authentication or connectivity fails.
func CheckConnection(ctx context.Context, client kubernetes.Interface) error {
	ctx, cancel := context.WithTimeout(ctx, connectionTimeout)
	defer cancel()

	type result struct {
		err error
	}
	ch := make(chan result, 1)
	go func() {
		_, err := client.Discovery().ServerVersion()
		ch <- result{err: err}
	}()

	select {
	case r := <-ch:
		if r.err == nil {
			return nil
		}
		return classifyConnectionError(r.err)
	case <-ctx.Done():
		return fmt.Errorf("cluster unreachable: connection timed out after %s\n\n"+
			"The Kubernetes API server did not respond in time.\n"+
			"This often happens when:\n"+
			"  1. AWS SSO session has expired — run: aws sso login --profile <profile>\n"+
			"  2. VPN is disconnected or the cluster endpoint is unreachable\n"+
			"  3. The cluster is down or the API server is overloaded", connectionTimeout)
	}
}

// classifyConnectionError inspects the raw error and returns a user-friendly message.
func classifyConnectionError(err error) error {
	msg := err.Error()
	lower := strings.ToLower(msg)

	// Network / connectivity errors — check first so "dial tcp" doesn't fall through
	var netErr net.Error
	if isNetError(err, &netErr) || strings.Contains(lower, "connection refused") ||
		strings.Contains(lower, "no such host") || strings.Contains(lower, "dial tcp") ||
		strings.Contains(lower, "i/o timeout") {
		return fmt.Errorf("cluster unreachable: %w\n\n"+
			"Could not connect to the Kubernetes API server.\n"+
			"Check that:\n"+
			"  1. The cluster is running and the API server is healthy\n"+
			"  2. Your network/VPN allows access to the cluster endpoint\n"+
			"  3. The server address in your kubeconfig is correct", err)
	}

	// Certificate errors — check before auth so "certificate expired" doesn't match "expired" auth keyword
	if strings.Contains(lower, "x509") || strings.Contains(lower, "tls") ||
		(strings.Contains(lower, "certificate") && !strings.Contains(lower, "credentials")) {
		return fmt.Errorf("TLS/certificate error: %w\n\n"+
			"The cluster's TLS certificate could not be verified.\n"+
			"Check that your kubeconfig has the correct certificate-authority-data", err)
	}

	// Auth / credentials errors (AWS SSO, GCP, Azure, expired tokens, etc.)
	authKeywords := []string{
		"unauthorized",
		"forbidden",
		"provide credentials",
		"token has expired",
		"token is expired",
		"security token",
		"no credentials",
		"could not get token",
		"unable to authenticate",
		"access denied",
		"invalididentitytoken",
		"expired",
		"exec plugin",
		"getting credentials",
	}
	for _, kw := range authKeywords {
		if strings.Contains(lower, kw) {
			return fmt.Errorf("authentication failed: %w\n\n"+
				"Your cluster credentials appear to be invalid or expired.\n"+
				"Common fixes:\n"+
				"  AWS EKS:  aws sso login --profile <profile>  then  aws eks update-kubeconfig ...\n"+
				"  GKE:     gcloud auth login  then  gcloud container clusters get-credentials ...\n"+
				"  AKS:     az login  then  az aks get-credentials ...\n"+
				"  Generic: check that your kubeconfig token/cert is still valid", err)
		}
	}

	// Fallback
	return fmt.Errorf("failed to connect to cluster: %w", err)
}

// isNetError checks if err (or any wrapped error) is a net.Error.
func isNetError(err error, target *net.Error) bool {
	for e := err; e != nil; {
		if ne, ok := e.(net.Error); ok {
			*target = ne
			return true
		}
		if uw, ok := e.(interface{ Unwrap() error }); ok {
			e = uw.Unwrap()
		} else {
			break
		}
	}
	return false
}
