package scanner

import (
	"context"

	"github.com/jorgegabrielti/nautikube/internal/diagnosis"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type SecretScanner struct{}

func NewSecretScanner() *SecretScanner {
	return &SecretScanner{}
}

func (s *SecretScanner) Name() string {
	return "Secret"
}

func (s *SecretScanner) Scan(ctx context.Context, client kubernetes.Interface, namespace string) ([]diagnosis.Problem, error) {
	if namespace == "" {
		namespace = metav1.NamespaceAll
	}

	// This is a basic scaffold. It will be expanded with real checks.
	items, err := client.CoreV1().Secrets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var problems []diagnosis.Problem
	for _, item := range items.Items {
		// Skip default service account tokens and helm secrets which are managed by system
		if item.Type == "kubernetes.io/service-account-token" || item.Type == "helm.sh/release.v1" {
			continue
		}

		if len(item.Data) == 0 && len(item.StringData) == 0 {
			problems = append(problems, diagnosis.Problem{
				Resource:       "Secret",
				Name:           item.Name,
				Namespace:      item.Namespace,
				Severity:       diagnosis.Low,
				Issue:          "Secret contains no data",
				Explanation:    "The Secret has an empty data payload. Empty secrets are generally useless and might indicate a misconfiguration during deployment or external secret syncing failures.",
				RemediationKey: "secret_empty",
				Remediation: []string{
					"kubectl describe secret " + item.Name + " -n " + item.Namespace,
				},
			})
		}
	}
	return problems, nil
}
