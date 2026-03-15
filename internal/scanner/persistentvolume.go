package scanner

import (
	"context"

	"github.com/jorgegabrielti/nautikube/internal/diagnosis"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type PersistentVolumeScanner struct{}

func NewPersistentVolumeScanner() *PersistentVolumeScanner {
	return &PersistentVolumeScanner{}
}

func (s *PersistentVolumeScanner) Name() string {
	return "PersistentVolume"
}

func (s *PersistentVolumeScanner) Scan(ctx context.Context, client kubernetes.Interface, namespace string) ([]diagnosis.Problem, error) {
	// Cluster-scoped resource
	items, err := client.CoreV1().PersistentVolumes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var problems []diagnosis.Problem
	for _, item := range items.Items {
		if item.Status.Phase == "Failed" {
			problems = append(problems, diagnosis.Problem{
				Resource:       "PersistentVolume",
				Name:           item.Name,
				Namespace:      "",
				Severity:       diagnosis.High,
				Issue:          "PV is in Failed state",
				Explanation:    "The volume failed to be recycled or deleted by its provisioner. Manual intervention might be required to clear the backing storage.",
				RemediationKey: "pv_failed",
				Remediation: []string{
					"kubectl describe pv " + item.Name,
				},
			})
		} else if item.Status.Phase == "Released" {
			// Released can be normal if ReclaimPolicy is Retain. We flag it as Low severity just for visibility.
			problems = append(problems, diagnosis.Problem{
				Resource:       "PersistentVolume",
				Name:           item.Name,
				Namespace:      "",
				Severity:       diagnosis.Low,
				Issue:          "PV is Released but not recycled",
				Explanation:    "The claim previously bound to this volume was deleted, but the volume is retained. It cannot be bound again automatically.",
				RemediationKey: "pv_released",
				Remediation: []string{
					"kubectl describe pv " + item.Name,
				},
			})
		}
	}
	return problems, nil
}
