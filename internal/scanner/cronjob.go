package scanner

import (
	"context"

	"github.com/jorgegabrielti/nautikube/internal/diagnosis"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type CronJobScanner struct{}

func NewCronJobScanner() *CronJobScanner {
	return &CronJobScanner{}
}

func (s *CronJobScanner) Name() string {
	return "CronJob"
}

func (s *CronJobScanner) Scan(ctx context.Context, client kubernetes.Interface, namespace string) ([]diagnosis.Problem, error) {
	if namespace == "" {
		namespace = metav1.NamespaceAll
	}

	// This is a basic scaffold. It will be expanded with real checks.
	items, err := client.BatchV1().CronJobs(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var problems []diagnosis.Problem
	for _, item := range items.Items {
		// Detect suspended cronjobs
		if item.Spec.Suspend != nil && *item.Spec.Suspend {
			problems = append(problems, diagnosis.Problem{
				Resource:       "CronJob",
				Name:           item.Name,
				Namespace:      item.Namespace,
				Severity:       diagnosis.Low,
				Issue:          "CronJob is suspended",
				Explanation:    "The CronJob is currently suspended and will not create any new jobs. Ensure this is intentional.",
				RemediationKey: "cronjob_suspended",
				Remediation: []string{
					"kubectl describe cronjob " + item.Name + " -n " + item.Namespace,
				},
			})
		}
	}
	return problems, nil
}
