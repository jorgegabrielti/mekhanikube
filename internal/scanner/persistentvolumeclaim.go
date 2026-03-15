package scanner

import (
	"context"

	"github.com/jorgegabrielti/nautikube/internal/diagnosis"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type PersistentVolumeClaimScanner struct{}

func NewPersistentVolumeClaimScanner() *PersistentVolumeClaimScanner {
	return &PersistentVolumeClaimScanner{}
}

func (s *PersistentVolumeClaimScanner) Name() string {
	return "PersistentVolumeClaim"
}

func (s *PersistentVolumeClaimScanner) Scan(ctx context.Context, client kubernetes.Interface, namespace string) ([]diagnosis.Problem, error) {
	if namespace == "" {
		namespace = metav1.NamespaceAll
	}

	// This is a basic scaffold. It will be expanded with real checks.
	items, err := client.CoreV1().PersistentVolumeClaims(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var problems []diagnosis.Problem
	for _, item := range items.Items {
		// Check for unbound or lost PVCs
		if item.Status.Phase == "Pending" {
			problems = append(problems, diagnosis.Problem{
				Resource:       "PersistentVolumeClaim",
				Name:           item.Name,
				Namespace:      item.Namespace,
				Severity:       diagnosis.High,
				Issue:          "PVC is stuck in Pending state",
				Explanation:    "The PVC is unable to bind to a PersistentVolume. This usually happens if the requested StorageClass doesn't exist, lacks capacity, or there is no provisioner available.",
				RemediationKey: "pvc_pending",
				Remediation: []string{
					"kubectl describe pvc " + item.Name + " -n " + item.Namespace,
				},
			})
		} else if item.Status.Phase == "Lost" {
			problems = append(problems, diagnosis.Problem{
				Resource:       "PersistentVolumeClaim",
				Name:           item.Name,
				Namespace:      item.Namespace,
				Severity:       diagnosis.Critical,
				Issue:          "PVC is in Lost state",
				Explanation:    "The underlying PersistentVolume bound to this claim has been deleted or is otherwise missing from the cluster, compromising data integrity.",
				RemediationKey: "pvc_lost",
				Remediation: []string{
					"kubectl describe pvc " + item.Name + " -n " + item.Namespace,
				},
			})
		}
	}
	return problems, nil
}
