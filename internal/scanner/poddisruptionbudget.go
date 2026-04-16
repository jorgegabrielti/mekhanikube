package scanner

import (
	"context"

	"github.com/jorgegabrielti/nautikube/internal/diagnosis"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type PodDisruptionBudgetScanner struct{}

func NewPodDisruptionBudgetScanner() *PodDisruptionBudgetScanner {
	return &PodDisruptionBudgetScanner{}
}

func (s *PodDisruptionBudgetScanner) Name() string {
	return "PodDisruptionBudget"
}

func (s *PodDisruptionBudgetScanner) Scan(ctx context.Context, client kubernetes.Interface, namespace string) ([]diagnosis.Problem, error) {
	if namespace == "" {
		namespace = metav1.NamespaceAll
	}

	items, err := client.PolicyV1().PodDisruptionBudgets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var problems []diagnosis.Problem
	for _, item := range items.Items {
		// Detect PDB blocking disruptions
		if item.Status.DisruptionsAllowed == 0 {
			problems = append(problems, diagnosis.Problem{
				Resource:       "PodDisruptionBudget",
				Name:           item.Name,
				Namespace:      item.Namespace,
				Severity:       diagnosis.High,
				Issue:          "No disruptions allowed",
				Explanation:    "This PodDisruptionBudget has 0 disruptions allowed. This will block node drains and cluster maintenance until the workload is scaled up or the PDB is relaxed.",
				RemediationKey: "pdb_no_disruptions",
				Remediation: []string{
					"kubectl get pdb " + item.Name + " -n " + item.Namespace,
					"kubectl describe pdb " + item.Name + " -n " + item.Namespace,
				},
			})
		}
	}
	return problems, nil
}
