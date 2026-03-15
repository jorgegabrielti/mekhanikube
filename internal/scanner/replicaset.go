package scanner

import (
	"context"
	"fmt"

	"github.com/jorgegabrielti/nautikube/internal/diagnosis"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type ReplicaSetScanner struct{}

func NewReplicaSetScanner() *ReplicaSetScanner {
	return &ReplicaSetScanner{}
}

func (s *ReplicaSetScanner) Name() string {
	return "ReplicaSet"
}

func (s *ReplicaSetScanner) Scan(ctx context.Context, client kubernetes.Interface, namespace string) ([]diagnosis.Problem, error) {
	if namespace == "" {
		namespace = metav1.NamespaceAll
	}

	// This is a basic scaffold. It will be expanded with real checks.
	items, err := client.AppsV1().ReplicaSets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var problems []diagnosis.Problem
	for _, item := range items.Items {
		// Check for orphaned ReplicaSets (no owner)
		if len(item.OwnerReferences) == 0 && item.Status.Replicas > 0 {
			problems = append(problems, diagnosis.Problem{
				Resource:       "ReplicaSet",
				Name:           item.Name,
				Namespace:      item.Namespace,
				Severity:       diagnosis.Medium,
				Issue:          "Orphaned ReplicaSet detected",
				Explanation:    "This ReplicaSet has no owner references but still has active replicas. This usually happens when a Deployment is deleted but its ReplicaSets are orphaned.",
				RemediationKey: "replicaset_orphaned",
				Remediation: []string{
					"kubectl describe rs " + item.Name + " -n " + item.Namespace,
				},
			})
		} else if item.Status.Replicas != item.Status.ReadyReplicas {
			problems = append(problems, diagnosis.Problem{
				Resource:       "ReplicaSet",
				Name:           item.Name,
				Namespace:      item.Namespace,
				Severity:       diagnosis.High,
				Issue:          "ReplicaSet replicas mismatch",
				Explanation:    fmt.Sprintf("ReplicaSet has %d replicas but only %d are ready.", item.Status.Replicas, item.Status.ReadyReplicas),
				RemediationKey: "replicaset_mismatch",
				Remediation: []string{
					"kubectl describe rs " + item.Name + " -n " + item.Namespace,
				},
			})
		}
	}
	return problems, nil
}
