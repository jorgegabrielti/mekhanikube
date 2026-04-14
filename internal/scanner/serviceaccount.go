package scanner

import (
	"context"

	"github.com/jorgegabrielti/nautikube/internal/diagnosis"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type ServiceAccountScanner struct{}

func NewServiceAccountScanner() *ServiceAccountScanner {
	return &ServiceAccountScanner{}
}

func (s *ServiceAccountScanner) Name() string {
	return "ServiceAccount"
}

// systemNamespaces are namespaces where empty SAs are expected and not actionable.
var systemNamespaces = map[string]bool{
	"kube-system":     true,
	"kube-public":     true,
	"kube-node-lease": true,
}

func (s *ServiceAccountScanner) Scan(ctx context.Context, client kubernetes.Interface, namespace string) ([]diagnosis.Problem, error) {
	if namespace == "" {
		namespace = metav1.NamespaceAll
	}

	items, err := client.CoreV1().ServiceAccounts(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var problems []diagnosis.Problem
	for _, item := range items.Items {
		if len(item.Secrets) > 0 || len(item.ImagePullSecrets) > 0 {
			continue
		}

		// Skip SAs in system namespaces — empty SAs are expected there.
		if systemNamespaces[item.Namespace] {
			continue
		}

		// Skip the auto-created "default" SA — every namespace has one and it's always empty since K8s 1.24.
		if item.Name == "default" {
			continue
		}

		problems = append(problems, diagnosis.Problem{
			Resource:       "ServiceAccount",
			Name:           item.Name,
			Namespace:      item.Namespace,
			Severity:       diagnosis.Low,
			Issue:          "ServiceAccount has no secrets or imagePullSecrets",
			Explanation:    "This ServiceAccount does not have any Secrets or ImagePullSecrets associated with it. While this might be intentional, it can lead to authentication failures or image pull errors if the pod needs to access a private registry or authenticated API.",
			RemediationKey: "serviceaccount_no_secrets",
			Remediation: []string{
				"kubectl describe sa " + item.Name + " -n " + item.Namespace,
			},
		})
	}
	return problems, nil
}
