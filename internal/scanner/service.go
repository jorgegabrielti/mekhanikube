package scanner

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jorgegabrielti/nautikube/internal/diagnosis"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// ServiceScanner scans Service resources for common problems.
type ServiceScanner struct{}

// NewServiceScanner creates a new ServiceScanner.
func NewServiceScanner() *ServiceScanner {
	return &ServiceScanner{}
}

// Name returns the scanner name.
func (s *ServiceScanner) Name() string { return "Service" }

// Scan inspects Services and returns detected problems.
func (s *ServiceScanner) Scan(ctx context.Context, client kubernetes.Interface, namespace string) ([]diagnosis.Problem, error) {
	services, err := client.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		slog.Warn("failed to list services", "namespace", namespace, "error", err)
		return nil, fmt.Errorf("failed to list services: %w", err)
	}

	endpoints, err := client.CoreV1().Endpoints(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		slog.Warn("failed to list endpoints", "namespace", namespace, "error", err)
		return nil, fmt.Errorf("failed to list endpoints: %w", err)
	}

	// Build endpoint lookup map
	epMap := make(map[string]bool)
	for _, ep := range endpoints.Items {
		hasAddresses := false
		for _, subset := range ep.Subsets {
			if len(subset.Addresses) > 0 {
				hasAddresses = true
				break
			}
		}
		key := ep.Namespace + "/" + ep.Name
		epMap[key] = hasAddresses
	}

	var problems []diagnosis.Problem
	for _, svc := range services.Items {
		problems = append(problems, s.checkService(svc, epMap)...)
	}

	return problems, nil
}

func (s *ServiceScanner) checkService(svc corev1.Service, epMap map[string]bool) []diagnosis.Problem {
	var problems []diagnosis.Problem

	// ExternalName services don't use selectors or endpoints in the typical way
	if svc.Spec.Type == corev1.ServiceTypeExternalName {
		if svc.Spec.ExternalName == "" {
			problems = append(problems, diagnosis.Problem{
				Resource:       "Service",
				Namespace:      svc.Namespace,
				Name:           svc.Name,
				Issue:          "ExternalName service has no target DNS name",
				Severity:       diagnosis.Low,
				RemediationKey: "service-externalname-empty",
			})
		}
		return problems
	}

	// Skip headless services and services without selectors
	if svc.Spec.ClusterIP == "None" || len(svc.Spec.Selector) == 0 {
		return nil
	}

	// Skip the default kubernetes service in default namespace
	if svc.Name == "kubernetes" && svc.Namespace == "default" {
		return nil
	}

	// No endpoints
	key := svc.Namespace + "/" + svc.Name
	hasEndpoints, exists := epMap[key]
	if !exists || !hasEndpoints {
		problems = append(problems, diagnosis.Problem{
			Resource:       "Service",
			Namespace:      svc.Namespace,
			Name:           svc.Name,
			Issue:          "Service has no endpoints",
			Severity:       diagnosis.High,
			RemediationKey: "service-no-endpoints",
		})
	}

	// LoadBalancer pending
	if svc.Spec.Type == corev1.ServiceTypeLoadBalancer && len(svc.Status.LoadBalancer.Ingress) == 0 {
		problems = append(problems, diagnosis.Problem{
			Resource:       "Service",
			Namespace:      svc.Namespace,
			Name:           svc.Name,
			Issue:          "LoadBalancer has no external IP assigned",
			Severity:       diagnosis.Medium,
			RemediationKey: "service-lb-pending",
		})
	}

	return problems
}
