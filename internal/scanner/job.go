package scanner

import (
	"context"
	"fmt"

	"github.com/jorgegabrielti/nautikube/internal/diagnosis"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type JobScanner struct{}

func NewJobScanner() *JobScanner {
	return &JobScanner{}
}

func (s *JobScanner) Name() string {
	return "Job"
}

func (s *JobScanner) Scan(ctx context.Context, client kubernetes.Interface, namespace string) ([]diagnosis.Problem, error) {
	if namespace == "" {
		namespace = metav1.NamespaceAll
	}

	// This is a basic scaffold. It will be expanded with real checks.
	items, err := client.BatchV1().Jobs(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var problems []diagnosis.Problem
	for _, item := range items.Items {
		// Focus on failed jobs
		if item.Status.Failed > 0 {
			problems = append(problems, diagnosis.Problem{
				Resource:       "Job",
				Name:           item.Name,
				Namespace:      item.Namespace,
				Severity:       diagnosis.High,
				Issue:          "Job has failed executions",
				Explanation:    fmt.Sprintf("Job has detected %d failed pods. This indicates the task performed by the job encountered errors and exhausted its backoff limit.", item.Status.Failed),
				RemediationKey: "job_failed",
				Remediation: []string{
					"kubectl describe job " + item.Name + " -n " + item.Namespace,
					"kubectl logs -l job-name=" + item.Name + " -n " + item.Namespace,
				},
			})
		}
	}
	return problems, nil
}
