package scanner

import (
	"context"

	"github.com/jorgegabrielti/nautikube/internal/diagnosis"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type NetworkPolicyScanner struct{}

func NewNetworkPolicyScanner() *NetworkPolicyScanner {
	return &NetworkPolicyScanner{}
}

func (s *NetworkPolicyScanner) Name() string {
	return "NetworkPolicy"
}

func (s *NetworkPolicyScanner) Scan(ctx context.Context, client kubernetes.Interface, namespace string) ([]diagnosis.Problem, error) {
	if namespace == "" {
		namespace = metav1.NamespaceAll
	}

	items, err := client.NetworkingV1().NetworkPolicies(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var problems []diagnosis.Problem
	for _, item := range items.Items {
		if isAllowAllIngress(item) {
			problems = append(problems, diagnosis.Problem{
				Resource:       "NetworkPolicy",
				Name:           item.Name,
				Namespace:      item.Namespace,
				Severity:       diagnosis.High,
				Issue:          "NetworkPolicy allows all ingress traffic (allow-all)",
				Explanation:    "This NetworkPolicy contains an empty ingress rule (no From selectors and no Port restrictions), which permits all inbound traffic. This is a security risk that may expose workloads to unnecessary access.",
				RemediationKey: "netpol_allow_all",
				Remediation: []string{
					"kubectl describe netpol " + item.Name + " -n " + item.Namespace,
				},
			})
		}
	}
	return problems, nil
}

// isAllowAllIngress returns true when the NetworkPolicy has at least one
// ingress rule with no From selectors and no Port restrictions, which
// effectively allows all ingress traffic.
func isAllowAllIngress(np networkingv1.NetworkPolicy) bool {
	for _, rule := range np.Spec.Ingress {
		if len(rule.From) == 0 && len(rule.Ports) == 0 {
			return true
		}
	}
	return false
}
