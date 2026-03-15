package scanner

import (
	"context"

	"github.com/jorgegabrielti/nautikube/internal/diagnosis"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type ConfigMapScanner struct{}

func NewConfigMapScanner() *ConfigMapScanner {
	return &ConfigMapScanner{}
}

func (s *ConfigMapScanner) Name() string {
	return "ConfigMap"
}

func (s *ConfigMapScanner) Scan(ctx context.Context, client kubernetes.Interface, namespace string) ([]diagnosis.Problem, error) {
	if namespace == "" {
		namespace = metav1.NamespaceAll
	}

	// This is a basic scaffold. It will be expanded with real checks.
	items, err := client.CoreV1().ConfigMaps(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var problems []diagnosis.Problem
	for _, item := range items.Items {
		// Calculate estimated size
		size := 0
		for k, v := range item.Data {
			size += len(k) + len(v)
		}
		for k, v := range item.BinaryData {
			size += len(k) + len(v)
		}

		if size > 500*1024 { // 500KB warning threshold
			problems = append(problems, diagnosis.Problem{
				Resource:       "ConfigMap",
				Name:           item.Name,
				Namespace:      item.Namespace,
				Severity:       diagnosis.Medium,
				Issue:          "ConfigMap is dangerously large",
				Explanation:    "The ConfigMap size exceeds 500KB. Unusually large ConfigMaps can negatively impact etcd and control plane performance.",
				RemediationKey: "configmap_large",
				Remediation: []string{
					"kubectl describe configmap " + item.Name + " -n " + item.Namespace,
				},
			})
		}
	}
	return problems, nil
}
