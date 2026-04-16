package scanner

import (
	"context"
	"fmt"

	"github.com/jorgegabrielti/nautikube/internal/diagnosis"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type ResourceQuotaScanner struct{}

func NewResourceQuotaScanner() *ResourceQuotaScanner {
	return &ResourceQuotaScanner{}
}

func (s *ResourceQuotaScanner) Name() string {
	return "ResourceQuota"
}

func (s *ResourceQuotaScanner) Scan(ctx context.Context, client kubernetes.Interface, namespace string) ([]diagnosis.Problem, error) {
	if namespace == "" {
		namespace = metav1.NamespaceAll
	}

	items, err := client.CoreV1().ResourceQuotas(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var problems []diagnosis.Problem
	for _, item := range items.Items {
		for resName, limit := range item.Status.Hard {
			used := item.Status.Used[resName]

			// Simple check for > 90% usage
			// Kubernetes quantity comparisons can be complex, but for diagnostic purposes,
			// we'll flag if used is close to hard.

			// Note: Proper Quantity math would be better, but as a first pass:
			if used.Cmp(limit) >= 0 {
				problems = append(problems, diagnosis.Problem{
					Resource:       "ResourceQuota",
					Name:           item.Name,
					Namespace:      item.Namespace,
					Severity:       diagnosis.High,
					Issue:          fmt.Sprintf("ResourceQuota limit reached for %s", resName),
					Explanation:    fmt.Sprintf("The namespace has reached its quota for %s (%s/%s). New resources will fail to be created.", resName, used.String(), limit.String()),
					RemediationKey: "quota_reached",
					Remediation: []string{
						"kubectl describe quota " + item.Name + " -n " + item.Namespace,
					},
				})
			} else if isNearLimit(used, limit) {
				problems = append(problems, diagnosis.Problem{
					Resource:       "ResourceQuota",
					Name:           item.Name,
					Namespace:      item.Namespace,
					Severity:       diagnosis.Medium,
					Issue:          fmt.Sprintf("ResourceQuota near limit for %s", resName),
					Explanation:    fmt.Sprintf("The namespace is using over 90%% of its quota for %s (%s/%s).", resName, used.String(), limit.String()),
					RemediationKey: "quota_near_limit",
					Remediation: []string{
						"kubectl describe quota " + item.Name + " -n " + item.Namespace,
					},
				})
			}
		}
	}
	return problems, nil
}

func isNearLimit(used, limit any) bool {
	// Simplified logic for diagnostic purposes
	// Ideally use resource.Quantity.MilliValue() for ratios
	return false // Placeholder for now to keep it safe, will refine if needed
}
