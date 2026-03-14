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

// PodScanner scans Pod resources for common problems.
type PodScanner struct{}

// NewPodScanner creates a new PodScanner.
func NewPodScanner() *PodScanner {
	return &PodScanner{}
}

// Name returns the scanner name.
func (s *PodScanner) Name() string { return "Pod" }

// Scan inspects Pods and returns detected problems.
func (s *PodScanner) Scan(ctx context.Context, client kubernetes.Interface, namespace string) ([]diagnosis.Problem, error) {
	pods, err := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		slog.Warn("failed to list pods", "namespace", namespace, "error", err)
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	var problems []diagnosis.Problem

	for _, pod := range pods.Items {
		// Skip completed pods (Jobs)
		if pod.Status.Phase == corev1.PodSucceeded {
			continue
		}

		problems = append(problems, s.checkContainerStatuses(pod)...)
		problems = append(problems, s.checkPendingPod(pod)...)
	}

	return problems, nil
}

func (s *PodScanner) checkContainerStatuses(pod corev1.Pod) []diagnosis.Problem {
	var problems []diagnosis.Problem

	allStatuses := append(pod.Status.ContainerStatuses, pod.Status.InitContainerStatuses...)

	for _, cs := range allStatuses {
		// Waiting state: CrashLoopBackOff, ImagePullBackOff, ErrImagePull
		if cs.State.Waiting != nil {
			reason := cs.State.Waiting.Reason
			switch reason {
			case "CrashLoopBackOff":
				problems = append(problems, diagnosis.Problem{
					Resource:       "Pod",
					Namespace:      pod.Namespace,
					Name:           pod.Name,
					Issue:          fmt.Sprintf("Container %s in CrashLoopBackOff", cs.Name),
					Severity:       diagnosis.Critical,
					RemediationKey: "crashloopbackoff",
				})
			case "ImagePullBackOff", "ErrImagePull":
				problems = append(problems, diagnosis.Problem{
					Resource:       "Pod",
					Namespace:      pod.Namespace,
					Name:           pod.Name,
					Issue:          fmt.Sprintf("Container %s: %s", cs.Name, reason),
					Severity:       diagnosis.High,
					RemediationKey: "imagepullbackoff",
				})
			}
		}

		// Terminated state: OOMKilled, Error
		if cs.State.Terminated != nil {
			reason := cs.State.Terminated.Reason
			if reason == "OOMKilled" {
				problems = append(problems, diagnosis.Problem{
					Resource:       "Pod",
					Namespace:      pod.Namespace,
					Name:           pod.Name,
					Issue:          fmt.Sprintf("Container %s was OOMKilled", cs.Name),
					Severity:       diagnosis.Critical,
					RemediationKey: "oomkilled",
				})
			}
		}

		// High restart count
		if cs.RestartCount > 5 {
			sev := diagnosis.Medium
			if cs.RestartCount > 50 {
				sev = diagnosis.Critical
			} else if cs.RestartCount > 20 {
				sev = diagnosis.High
			}
			problems = append(problems, diagnosis.Problem{
				Resource:       "Pod",
				Namespace:      pod.Namespace,
				Name:           pod.Name,
				Issue:          fmt.Sprintf("Container %s has restarted %d times", cs.Name, cs.RestartCount),
				Severity:       sev,
				RemediationKey: "high-restarts",
			})
		}

		// Container not ready (running but not passing readiness probe)
		if cs.State.Running != nil && !cs.Ready {
			problems = append(problems, diagnosis.Problem{
				Resource:       "Pod",
				Namespace:      pod.Namespace,
				Name:           pod.Name,
				Issue:          fmt.Sprintf("Container %s is running but not ready", cs.Name),
				Severity:       diagnosis.Medium,
				RemediationKey: "container-not-ready",
			})
		}
	}

	return problems
}

func (s *PodScanner) checkPendingPod(pod corev1.Pod) []diagnosis.Problem {
	if pod.Status.Phase != corev1.PodPending {
		return nil
	}

	// Only report Pending if there are no container statuses yet (truly unscheduled)
	if len(pod.Status.ContainerStatuses) > 0 {
		return nil
	}

	return []diagnosis.Problem{
		{
			Resource:       "Pod",
			Namespace:      pod.Namespace,
			Name:           pod.Name,
			Issue:          "Pod is stuck in Pending state",
			Severity:       diagnosis.High,
			RemediationKey: "pod-pending",
		},
	}
}
