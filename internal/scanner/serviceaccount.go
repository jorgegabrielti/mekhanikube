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

func (s *ServiceAccountScanner) Scan(ctx context.Context, client kubernetes.Interface, namespace string) ([]diagnosis.Problem, error) {
	if namespace == "" {
		namespace = metav1.NamespaceAll
	}

	// This is a basic scaffold. It will be expanded with real checks.
	items, err := client.CoreV1().ServiceAccounts(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var problems []diagnosis.Problem
	for _, item := range items.Items {
		// Detect ServiceAccounts without any secrets (might be missing image pull secrets)
		// Note: Default token auto-generation is common, so we check for zero secrets as a low-severity hint
		if len(item.Secrets) == 0 && len(item.ImagePullSecrets) == 0 {
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
	}
	return problems, nil
}
