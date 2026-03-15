package scanner

import (
	"context"

	"github.com/jorgegabrielti/nautikube/internal/diagnosis"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type NetworkPolicyScanner struct{}

func NewNetworkPolicyScanner() *NetworkPolicyScanner {
	return &NetworkPolicyScanner{}
}

func (s *NetworkPolicyScanner) Name() string {
	return "NetworkPolicy"
}

func (s *NetworkPolicyScanner) Scan(ctx context.Context, client kubernetes.Interface, namespace string) ([]diagnosis.Problem, error) {
	if namespace == "" {
		namespace = metav1.NamespaceAll
	}

	// This is a basic scaffold. It will be expanded with real checks.
	items, err := client.NetworkingV1().NetworkPolicies(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var problems []diagnosis.Problem
	for _, item := range items.Items {
		_ = item // implement checks here
	}
	return problems, nil
}
