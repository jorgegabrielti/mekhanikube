package scanner

import (
	"context"
	"strings"

	"github.com/jorgegabrielti/nautikube/internal/diagnosis"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type ClusterRoleBindingScanner struct{}

func NewClusterRoleBindingScanner() *ClusterRoleBindingScanner {
	return &ClusterRoleBindingScanner{}
}

func (s *ClusterRoleBindingScanner) Name() string {
	return "ClusterRoleBinding"
}

func (s *ClusterRoleBindingScanner) Scan(ctx context.Context, client kubernetes.Interface, namespace string) ([]diagnosis.Problem, error) {
	// Cluster-scoped resource
	items, err := client.RbacV1().ClusterRoleBindings().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var problems []diagnosis.Problem
	for _, item := range items.Items {
		// Skip system bindings
		if strings.HasPrefix(item.Name, "system:") {
			continue
		}

		if len(item.Subjects) == 0 {
			problems = append(problems, diagnosis.Problem{
				Resource:       "ClusterRoleBinding",
				Name:           item.Name,
				Namespace:      "",
				Severity:       diagnosis.Low,
				Issue:          "ClusterRoleBinding has no subjects",
				Explanation:    "This ClusterRoleBinding is defined but does not bind the ClusterRole to any subjects. It is currently ineffective.",
				RemediationKey: "rbac_cluster_empty_binding",
				Remediation: []string{
					"kubectl describe clusterrolebinding " + item.Name,
				},
			})
		}
	}
	return problems, nil
}
