package scanner

import (
	"context"
	"fmt"

	"github.com/jorgegabrielti/nautikube/internal/diagnosis"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type HorizontalPodAutoscalerScanner struct{}

func NewHorizontalPodAutoscalerScanner() *HorizontalPodAutoscalerScanner {
	return &HorizontalPodAutoscalerScanner{}
}

func (s *HorizontalPodAutoscalerScanner) Name() string {
	return "HorizontalPodAutoscaler"
}

func (s *HorizontalPodAutoscalerScanner) Scan(ctx context.Context, client kubernetes.Interface, namespace string) ([]diagnosis.Problem, error) {
	if namespace == "" {
		namespace = metav1.NamespaceAll
	}

	// This is a basic scaffold. It will be expanded with real checks.
	items, err := client.AutoscalingV1().HorizontalPodAutoscalers(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var problems []diagnosis.Problem
	for _, item := range items.Items {
		// Detect HPA at max replicas
		if item.Status.CurrentReplicas == item.Spec.MaxReplicas {
			problems = append(problems, diagnosis.Problem{
				Resource:       "HorizontalPodAutoscaler",
				Name:           item.Name,
				Namespace:      item.Namespace,
				Severity:       diagnosis.Medium,
				Issue:          "HPA at maximum replicas",
				Explanation:    fmt.Sprintf("HPA has reached its maximum replica limit (%d). This indicates the workload is under sustained heavy load and may be unable to scale further to meet demand.", item.Spec.MaxReplicas),
				RemediationKey: "hpa_at_max",
				Remediation: []string{
					"kubectl get hpa " + item.Name + " -n " + item.Namespace,
					"kubectl describe hpa " + item.Name + " -n " + item.Namespace,
				},
			})
		}
	}
	return problems, nil
}
