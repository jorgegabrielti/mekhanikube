package scanner

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/jorgegabrielti/nautikube/internal/diagnosis"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// ClusterScanner performs global cluster-level health checks.
type ClusterScanner struct{}

// NewClusterScanner creates a new ClusterScanner.
func NewClusterScanner() *ClusterScanner {
	return &ClusterScanner{}
}

// Name returns the scanner name.
func (s *ClusterScanner) Name() string { return "Cluster" }

// Scan performs high-level cluster health checks.
func (s *ClusterScanner) Scan(ctx context.Context, client kubernetes.Interface, namespace string) ([]diagnosis.Problem, error) {
	// If a namespace is specified, Cluster scanner still performs cluster-wide checks
	// but we might want to skip them if the user ONLY wants to see namespace-specific resources.
	// However, cluster health affects all namespaces, so we include it.

	var problems []diagnosis.Problem

	// 1. API Server Health Check
	apiProblems := s.checkAPIHealth(ctx, client)
	problems = append(problems, apiProblems...)

	// 2. Core Components Health (CoreDNS)
	compProblems, err := s.checkCoreComponents(ctx, client)
	if err != nil {
		slog.Warn("failed to check core components", "error", err)
	} else {
		problems = append(problems, compProblems...)
	}

	// 3. Node Version Skew
	skewProblems, err := s.checkNodeVersionSkew(ctx, client)
	if err != nil {
		slog.Warn("failed to check node version skew", "error", err)
	} else {
		problems = append(problems, skewProblems...)
	}

	return problems, nil
}

func (s *ClusterScanner) checkAPIHealth(ctx context.Context, client kubernetes.Interface) []diagnosis.Problem {
	// Use Discovery client to hit /healthz
	if client.Discovery() == nil || client.Discovery().RESTClient() == nil {
		return nil // Skip if client doesn't support REST calls (e.g. fake client)
	}
	res := client.Discovery().RESTClient().Get().AbsPath("/healthz").Do(ctx)
	if err := res.Error(); err != nil {
		return []diagnosis.Problem{{
			Resource:       "Cluster",
			Name:           "API Server",
			Issue:          fmt.Sprintf("API Server health check failed: %v", err),
			Severity:       diagnosis.Critical,
			RemediationKey: "cluster-api-unhealthy",
		}}
	}

	return []diagnosis.Problem{{
		Resource: "Cluster",
		Name:     "API Server",
		Issue:    "API Server is healthy and responding",
		Severity: diagnosis.Info,
	}}
}

func (s *ClusterScanner) checkCoreComponents(ctx context.Context, client kubernetes.Interface) ([]diagnosis.Problem, error) {
	var problems []diagnosis.Problem

	// Check CoreDNS pods in kube-system
	pods, err := client.CoreV1().Pods("kube-system").List(ctx, metav1.ListOptions{
		LabelSelector: "k8s-app=kube-dns",
	})
	if err != nil {
		return nil, err
	}

	if len(pods.Items) == 0 {
		problems = append(problems, diagnosis.Problem{
			Resource:       "Cluster",
			Namespace:      "kube-system",
			Name:           "CoreDNS",
			Issue:          "CoreDNS pods not found in kube-system",
			Severity:       diagnosis.Critical,
			RemediationKey: "cluster-coredns-missing",
		})
	} else {
		healthy := 0
		for _, pod := range pods.Items {
			if pod.Status.Phase == "Running" {
				isReady := false
				for _, cond := range pod.Status.Conditions {
					if cond.Type == "Ready" && cond.Status == "True" {
						isReady = true
						break
					}
				}
				if isReady {
					healthy++
				}
			}
		}
		if healthy == 0 {
			problems = append(problems, diagnosis.Problem{
				Resource:       "Cluster",
				Namespace:      "kube-system",
				Name:           "CoreDNS",
				Issue:          "No healthy CoreDNS pods available",
				Severity:       diagnosis.Critical,
				RemediationKey: "cluster-coredns-unhealthy",
			})
		} else {
			problems = append(problems, diagnosis.Problem{
				Resource:  "Cluster",
				Namespace: "kube-system",
				Name:      "CoreDNS",
				Issue:     fmt.Sprintf("CoreDNS is healthy (%d pods ready)", healthy),
				Severity:  diagnosis.Info,
			})
		}
	}

	return problems, nil
}

func (s *ClusterScanner) checkNodeVersionSkew(ctx context.Context, client kubernetes.Interface) ([]diagnosis.Problem, error) {
	nodes, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	if len(nodes.Items) == 0 {
		return nil, nil
	}

	versions := make(map[string]int)
	for _, node := range nodes.Items {
		v := node.Status.NodeInfo.KubeletVersion
		versions[v]++
	}

	if len(versions) > 1 {
		var vStrings []string
		for v, count := range versions {
			vStrings = append(vStrings, fmt.Sprintf("%s (%d nodes)", v, count))
		}

		return []diagnosis.Problem{{
			Resource:       "Cluster",
			Name:           "Nodes",
			Issue:          fmt.Sprintf("Version skew detected between nodes: %s", strings.Join(vStrings, ", ")),
			Severity:       diagnosis.Medium,
			RemediationKey: "cluster-node-skew",
		}}, nil
	}

	var version string
	for v := range versions {
		version = v
		break
	}

	return []diagnosis.Problem{{
		Resource: "Cluster",
		Name:     "Nodes",
		Issue:    fmt.Sprintf("All nodes are running consistent version: %s", version),
		Severity: diagnosis.Info,
	}}, nil
}
