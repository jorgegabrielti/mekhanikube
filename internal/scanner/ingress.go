package scanner

import (
	"context"

	"github.com/jorgegabrielti/nautikube/internal/diagnosis"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type IngressScanner struct{}

func NewIngressScanner() *IngressScanner {
	return &IngressScanner{}
}

func (s *IngressScanner) Name() string {
	return "Ingress"
}

func (s *IngressScanner) Scan(ctx context.Context, client kubernetes.Interface, namespace string) ([]diagnosis.Problem, error) {
	if namespace == "" {
		namespace = metav1.NamespaceAll
	}

	// This is a basic scaffold. It will be expanded with real checks.
	items, err := client.NetworkingV1().Ingresses(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var problems []diagnosis.Problem
	for _, item := range items.Items {
		// Check for missing default backend or empty rules
		if len(item.Spec.Rules) == 0 && item.Spec.DefaultBackend == nil {
			problems = append(problems, diagnosis.Problem{
				Resource:       "Ingress",
				Name:           item.Name,
				Namespace:      item.Namespace,
				Severity:       diagnosis.Medium,
				Issue:          "Ingress has no rules or default backend",
				Explanation:    "The Ingress resource is defined but does not route traffic to any backends. It lacks both rules and a default backend.",
				RemediationKey: "ingress_empty",
				Remediation: []string{
					"kubectl describe ingress " + item.Name + " -n " + item.Namespace,
				},
			})
		}
	}
	return problems, nil
}
