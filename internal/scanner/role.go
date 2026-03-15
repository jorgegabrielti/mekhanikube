package scanner

import (
	"context"

	"github.com/jorgegabrielti/nautikube/internal/diagnosis"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type RoleScanner struct{}

func NewRoleScanner() *RoleScanner {
	return &RoleScanner{}
}

func (s *RoleScanner) Name() string {
	return "Role"
}

func (s *RoleScanner) Scan(ctx context.Context, client kubernetes.Interface, namespace string) ([]diagnosis.Problem, error) {
	if namespace == "" {
		namespace = metav1.NamespaceAll
	}

	// This is a basic scaffold. It will be expanded with real checks.
	items, err := client.RbacV1().Roles(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var problems []diagnosis.Problem
	for _, item := range items.Items {
		// Detect overly permissive roles (wildcard verbs or resources)
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
				Resource:       "Role",
				Name:           item.Name,
				Namespace:      item.Namespace,
				Severity:       diagnosis.Medium,
				Issue:          "Role has wildcard permissions",
				Explanation:    "This Role uses wildcard ('*') verbs or resources, giving it broad permissions within the namespace. Follow the principle of least privilege and restrict access to specific verbs and resources.",
				RemediationKey: "rbac_wildcard",
				Remediation: []string{
					"kubectl describe role " + item.Name + " -n " + item.Namespace,
				},
			})
		}
	}
	return problems, nil
}
