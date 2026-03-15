package scanner

import (
	"context"
	"strings"

	"github.com/jorgegabrielti/nautikube/internal/diagnosis"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type ClusterRoleScanner struct{}

func NewClusterRoleScanner() *ClusterRoleScanner {
	return &ClusterRoleScanner{}
}

func (s *ClusterRoleScanner) Name() string {
	return "ClusterRole"
}

func (s *ClusterRoleScanner) Scan(ctx context.Context, client kubernetes.Interface, namespace string) ([]diagnosis.Problem, error) {
	// Cluster-scoped resource
	items, err := client.RbacV1().ClusterRoles().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var problems []diagnosis.Problem
	for _, item := range items.Items {
		// Skip system roles
		if strings.HasPrefix(item.Name, "system:") {
			continue
		}

		hasWildcard := false
		for _, rule := range item.Rules {
			for _, v := range rule.Verbs {
				if v == "*" {
					hasWildcard = true
					break
				}
			}
			if hasWildcard {
				break
			}
		}

		if hasWildcard {
			problems = append(problems, diagnosis.Problem{
				Resource:       "ClusterRole",
				Name:           item.Name,
				Namespace:      "",
				Severity:       diagnosis.High,
				Issue:          "ClusterRole has wildcard permissions",
				Explanation:    "This ClusterRole uses wildcard ('*') verbs or resources across the entire cluster. This is high risk and should be strictly audited.",
				RemediationKey: "rbac_cluster_wildcard",
				Remediation: []string{
					"kubectl describe clusterrole " + item.Name,
				},
			})
		}
	}
	return problems, nil
}
