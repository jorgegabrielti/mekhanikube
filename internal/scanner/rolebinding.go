package scanner

import (
	"context"

	"github.com/jorgegabrielti/nautikube/internal/diagnosis"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type RoleBindingScanner struct{}

func NewRoleBindingScanner() *RoleBindingScanner {
	return &RoleBindingScanner{}
}

func (s *RoleBindingScanner) Name() string {
	return "RoleBinding"
}

func (s *RoleBindingScanner) Scan(ctx context.Context, client kubernetes.Interface, namespace string) ([]diagnosis.Problem, error) {
	if namespace == "" {
		namespace = metav1.NamespaceAll
	}

	// This is a basic scaffold. It will be expanded with real checks.
	items, err := client.RbacV1().RoleBindings(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var problems []diagnosis.Problem
	for _, item := range items.Items {
		if len(item.Subjects) == 0 {
			problems = append(problems, diagnosis.Problem{
				Resource:       "RoleBinding",
				Name:           item.Name,
				Namespace:      item.Namespace,
				Severity:       diagnosis.Low,
				Issue:          "RoleBinding has no subjects",
				Explanation:    "This RoleBinding is defined but does not bind the Role to any Users, Groups, or ServiceAccounts. It is currently ineffective.",
				RemediationKey: "rbac_empty_binding",
				Remediation: []string{
					"kubectl describe rolebinding " + item.Name + " -n " + item.Namespace,
				},
			})
		}
	}
	return problems, nil
}
