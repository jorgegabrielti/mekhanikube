package scanner

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jorgegabrielti/nautikube/internal/diagnosis"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// DeploymentScanner scans Deployment resources for common problems.
type DeploymentScanner struct{}

// NewDeploymentScanner creates a new DeploymentScanner.
func NewDeploymentScanner() *DeploymentScanner {
	return &DeploymentScanner{}
}

// Name returns the scanner name.
func (s *DeploymentScanner) Name() string { return "Deployment" }

// Scan inspects Deployments and returns detected problems.
func (s *DeploymentScanner) Scan(ctx context.Context, client kubernetes.Interface, namespace string) ([]diagnosis.Problem, error) {
	deployments, err := client.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		slog.Warn("failed to list deployments", "namespace", namespace, "error", err)
		return nil, fmt.Errorf("failed to list deployments: %w", err)
	}

	var problems []diagnosis.Problem

	for _, dep := range deployments.Items {
		problems = append(problems, s.checkDeployment(dep)...)
	}

	return problems, nil
}

func (s *DeploymentScanner) checkDeployment(dep appsv1.Deployment) []diagnosis.Problem {
	var problems []diagnosis.Problem
	desired := int32(1) // Default if nil
	if dep.Spec.Replicas != nil {
		desired = *dep.Spec.Replicas
	}

	// Zero replicas
	if desired == 0 {
		problems = append(problems, diagnosis.Problem{
			Resource:       "Deployment",
			Namespace:      dep.Namespace,
			Name:           dep.Name,
			Issue:          "Deployment is scaled to zero replicas",
			Severity:       diagnosis.Low,
			RemediationKey: "deployment-zero-replicas",
		})
		return problems
	}

	// Unavailable replicas
	if dep.Status.UnavailableReplicas > 0 {
		problems = append(problems, diagnosis.Problem{
			Resource:       "Deployment",
			Namespace:      dep.Namespace,
			Name:           dep.Name,
			Issue:          fmt.Sprintf("%d unavailable replicas", dep.Status.UnavailableReplicas),
			Severity:       diagnosis.High,
			RemediationKey: "deployment-unavailable",
		})
	}

	// Rollout stuck (Progressing condition is False)
	for _, cond := range dep.Status.Conditions {
		if cond.Type == appsv1.DeploymentProgressing && cond.Status == "False" {
			problems = append(problems, diagnosis.Problem{
				Resource:       "Deployment",
				Namespace:      dep.Namespace,
				Name:           dep.Name,
				Issue:          fmt.Sprintf("Rollout stuck: %s", cond.Reason),
				Severity:       diagnosis.High,
				RemediationKey: "deployment-stuck",
			})
		}
	}

	// Replicas mismatch (ready < desired, but no unavailable reported)
	if dep.Status.ReadyReplicas < desired && dep.Status.UnavailableReplicas == 0 {
		problems = append(problems, diagnosis.Problem{
			Resource:       "Deployment",
			Namespace:      dep.Namespace,
			Name:           dep.Name,
			Issue:          fmt.Sprintf("Ready replicas (%d) less than desired (%d)", dep.Status.ReadyReplicas, desired),
			Severity:       diagnosis.Medium,
			RemediationKey: "deployment-mismatch",
		})
	}

	return problems
}
