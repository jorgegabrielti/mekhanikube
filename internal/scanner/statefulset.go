package scanner

import (
	"context"
	"fmt"

	"github.com/jorgegabrielti/nautikube/internal/diagnosis"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type StatefulSetScanner struct{}

func NewStatefulSetScanner() *StatefulSetScanner {
	return &StatefulSetScanner{}
}

func (s *StatefulSetScanner) Name() string {
	return "StatefulSet"
}

func (s *StatefulSetScanner) Scan(ctx context.Context, client kubernetes.Interface, namespace string) ([]diagnosis.Problem, error) {
	if namespace == "" {
		namespace = metav1.NamespaceAll
	}

	// This is a basic scaffold. It will be expanded with real checks.
	items, err := client.AppsV1().StatefulSets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var problems []diagnosis.Problem
	for _, item := range items.Items {
		// Detect replication mismatch or unhealthy pods
		if item.Status.Replicas != item.Status.ReadyReplicas {
			problems = append(problems, diagnosis.Problem{
				Resource:       "StatefulSet",
				Name:           item.Name,
				Namespace:      item.Namespace,
				Severity:       diagnosis.High,
				Issue:          "StatefulSet replicas mismatch",
				Explanation:    fmt.Sprintf("StatefulSet has %d desired replicas but only %d are ready. StatefulSets often fail due to PersistentVolumeClaim binding issues or pod initialization failures.", item.Status.Replicas, item.Status.ReadyReplicas),
				RemediationKey: "statefulset_mismatch",
				Remediation: []string{
					"kubectl describe sts " + item.Name + " -n " + item.Namespace,
				},
			})
		}
	}
	return problems, nil
}
