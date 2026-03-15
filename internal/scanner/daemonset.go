package scanner

import (
	"context"
	"fmt"

	"github.com/jorgegabrielti/nautikube/internal/diagnosis"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type DaemonSetScanner struct{}

func NewDaemonSetScanner() *DaemonSetScanner {
	return &DaemonSetScanner{}
}

func (s *DaemonSetScanner) Name() string {
	return "DaemonSet"
}

func (s *DaemonSetScanner) Scan(ctx context.Context, client kubernetes.Interface, namespace string) ([]diagnosis.Problem, error) {
	if namespace == "" {
		namespace = metav1.NamespaceAll
	}

	// This is a basic scaffold. It will be expanded with real checks.
	items, err := client.AppsV1().DaemonSets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var problems []diagnosis.Problem
	for _, item := range items.Items {
		if item.Status.DesiredNumberScheduled != item.Status.NumberReady {
			problems = append(problems, diagnosis.Problem{
				Resource:       "DaemonSet",
				Name:           item.Name,
				Namespace:      item.Namespace,
				Severity:       diagnosis.High,
				Issue:          "DaemonSet not fully scheduled",
				Explanation:    fmt.Sprintf("DaemonSet has %d desired nodes but only %d are ready. This might be due to node taints, insufficient resources on specific nodes, or pod scheduling restrictions.", item.Status.DesiredNumberScheduled, item.Status.NumberReady),
				RemediationKey: "daemonset_misscheduled",
				Remediation: []string{
					"kubectl describe ds " + item.Name + " -n " + item.Namespace,
				},
			})
		}
	}
	return problems, nil
}
