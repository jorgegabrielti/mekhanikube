package scanner

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jorgegabrielti/nautikube/internal/diagnosis"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// NodeScanner scans Node resources for common problems.
type NodeScanner struct{}

// NewNodeScanner creates a new NodeScanner.
func NewNodeScanner() *NodeScanner {
	return &NodeScanner{}
}

// Name returns the scanner name.
func (s *NodeScanner) Name() string { return "Node" }

// Scan inspects Nodes and returns detected problems.
func (s *NodeScanner) Scan(ctx context.Context, client kubernetes.Interface, namespace string) ([]diagnosis.Problem, error) {
	// Nodes are cluster-scoped; namespace is ignored
	nodes, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		slog.Warn("failed to list nodes", "error", err)
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}

	var problems []diagnosis.Problem
	for _, node := range nodes.Items {
		problems = append(problems, s.checkNode(node)...)
	}

	return problems, nil
}

func (s *NodeScanner) checkNode(node corev1.Node) []diagnosis.Problem {
	var problems []diagnosis.Problem

	for _, condition := range node.Status.Conditions {
		switch condition.Type {
		case corev1.NodeReady:
			if condition.Status != corev1.ConditionTrue {
				problems = append(problems, diagnosis.Problem{
					Resource:       "Node",
					Namespace:      "",
					Name:           node.Name,
					Issue:          fmt.Sprintf("Node is not Ready (status: %s)", condition.Status),
					Severity:       diagnosis.Critical,
					RemediationKey: "node-not-ready",
				})
			}
		case corev1.NodeMemoryPressure:
			if condition.Status == corev1.ConditionTrue {
				problems = append(problems, diagnosis.Problem{
					Resource:       "Node",
					Namespace:      "",
					Name:           node.Name,
					Issue:          "Node is under memory pressure",
					Severity:       diagnosis.High,
					RemediationKey: "node-memory-pressure",
				})
			}
		case corev1.NodeDiskPressure:
			if condition.Status == corev1.ConditionTrue {
				problems = append(problems, diagnosis.Problem{
					Resource:       "Node",
					Namespace:      "",
					Name:           node.Name,
					Issue:          "Node is under disk pressure",
					Severity:       diagnosis.High,
					RemediationKey: "node-disk-pressure",
				})
			}
		case corev1.NodePIDPressure:
			if condition.Status == corev1.ConditionTrue {
				problems = append(problems, diagnosis.Problem{
					Resource:       "Node",
					Namespace:      "",
					Name:           node.Name,
					Issue:          "Node is under PID pressure",
					Severity:       diagnosis.High,
					RemediationKey: "node-pid-pressure",
				})
			}
		}
	}

	// Unschedulable (cordoned)
	if node.Spec.Unschedulable {
		problems = append(problems, diagnosis.Problem{
			Resource:       "Node",
			Namespace:      "",
			Name:           node.Name,
			Issue:          "Node is cordoned (unschedulable)",
			Severity:       diagnosis.Low,
			RemediationKey: "node-unschedulable",
		})
	}

	return problems
}
