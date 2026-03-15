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
		// Find the matching container spec to extract surgical details (limits, requests, envs)
		var containerSpec *corev1.Container
		for i := range pod.Spec.Containers {
			if pod.Spec.Containers[i].Name == cs.Name {
				containerSpec = &pod.Spec.Containers[i]
				break
			}
		}
		if containerSpec == nil {
			for i := range pod.Spec.InitContainers {
				if pod.Spec.InitContainers[i].Name == cs.Name {
					containerSpec = &pod.Spec.InitContainers[i]
					break
				}
			}
		}

		// Waiting state: CrashLoopBackOff, ImagePullBackOff, ErrImagePull
		if cs.State.Waiting != nil {
			reason := cs.State.Waiting.Reason
			switch reason {
			case "CrashLoopBackOff":
				if cs.LastTerminationState.Terminated != nil && cs.LastTerminationState.Terminated.Reason == "OOMKilled" {
					var details []string
					var offendingProp string
					var mutativeFix string
					if containerSpec != nil && containerSpec.Resources.Limits.Memory().Value() > 0 {
						details = append(details, fmt.Sprintf("Exact Cause: Container exceeded its strict memory limit of %s.", containerSpec.Resources.Limits.Memory().String()))
						offendingProp = fmt.Sprintf(".spec.containers[%s].resources.limits.memory = %s", cs.Name, containerSpec.Resources.Limits.Memory().String())
						mutativeFix = fmt.Sprintf("kubectl edit pod %s -n %s", pod.Name, pod.Namespace)
					} else {
						details = append(details, "Exact Cause: Container consumed all available node memory (No limits were set).")
					}

					problems = append(problems, diagnosis.Problem{
						Resource:          "Pod",
						Namespace:         pod.Namespace,
						Name:              pod.Name,
						Issue:             fmt.Sprintf("Container %s in CrashLoopBackOff (OOMKilled)", cs.Name),
						Severity:          diagnosis.Critical,
						RemediationKey:    "oomkilled",
						Details:           details,
						OffendingProperty: offendingProp,
						MutativeFix:       mutativeFix,
					})
				} else {
					problems = append(problems, diagnosis.Problem{
						Resource:       "Pod",
						Namespace:      pod.Namespace,
						Name:           pod.Name,
						Issue:          fmt.Sprintf("Container %s in CrashLoopBackOff", cs.Name),
						Severity:       diagnosis.Critical,
						RemediationKey: "crashloopbackoff",
					})
				}
			case "ImagePullBackOff", "ErrImagePull":
				var offendingProp string
				var mutativeFix string
				if containerSpec != nil {
					offendingProp = fmt.Sprintf(".spec.containers[%s].image = %s", cs.Name, containerSpec.Image)
					mutativeFix = fmt.Sprintf("kubectl set image pod %s %s=NEW_IMAGE_NAME -n %s", pod.Name, cs.Name, pod.Namespace)
				}

				problems = append(problems, diagnosis.Problem{
					Resource:          "Pod",
					Namespace:         pod.Namespace,
					Name:              pod.Name,
					Issue:             fmt.Sprintf("Container %s: %s", cs.Name, reason),
					Severity:          diagnosis.High,
					RemediationKey:    "imagepullbackoff",
					OffendingProperty: offendingProp,
					MutativeFix:       mutativeFix,
				})
			case "CreateContainerConfigError":
				var details []string
				if cs.State.Waiting.Message != "" {
					details = append(details, fmt.Sprintf("Exact Cause: %s", cs.State.Waiting.Message))
				}
				problems = append(problems, diagnosis.Problem{
					Resource:       "Pod",
					Namespace:      pod.Namespace,
					Name:           pod.Name,
					Issue:          fmt.Sprintf("Container %s failed to configure: %s", cs.Name, reason),
					Severity:       diagnosis.Critical,
					RemediationKey: "createcontainerconfigerror",
					Details:        details,
				})
			case "CreateContainerError":
				var details []string
				if cs.State.Waiting.Message != "" {
					details = append(details, fmt.Sprintf("Exact Cause: %s", cs.State.Waiting.Message))
				}
				problems = append(problems, diagnosis.Problem{
					Resource:       "Pod",
					Namespace:      pod.Namespace,
					Name:           pod.Name,
					Issue:          fmt.Sprintf("Container %s failed to start: %s", cs.Name, reason),
					Severity:       diagnosis.High,
					RemediationKey: "createcontainererror",
					Details:        details,
				})
			}
		}

		// Terminated state: OOMKilled, Error
		if cs.State.Terminated != nil {
			reason := cs.State.Terminated.Reason
			if reason == "OOMKilled" {
				var details []string
				var offendingProp string
				var mutativeFix string
				if containerSpec != nil && containerSpec.Resources.Limits.Memory().Value() > 0 {
					details = append(details, fmt.Sprintf("Exact Cause: Container exceeded its strict memory limit of %s.", containerSpec.Resources.Limits.Memory().String()))
					offendingProp = fmt.Sprintf(".spec.containers[%s].resources.limits.memory = %s", cs.Name, containerSpec.Resources.Limits.Memory().String())
					mutativeFix = fmt.Sprintf("kubectl edit pod %s -n %s", pod.Name, pod.Namespace)
				} else {
					details = append(details, "Exact Cause: Container consumed all available node memory (No limits were set).")
				}

				problems = append(problems, diagnosis.Problem{
					Resource:          "Pod",
					Namespace:         pod.Namespace,
					Name:              pod.Name,
					Issue:             fmt.Sprintf("Container %s was OOMKilled", cs.Name),
					Severity:          diagnosis.Critical,
					RemediationKey:    "oomkilled",
					Details:           details,
					OffendingProperty: offendingProp,
					MutativeFix:       mutativeFix,
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

	// Extract the specific reason it is pending (e.g. Insufficient CPU)
	var details []string
	for _, cond := range pod.Status.Conditions {
		if cond.Type == corev1.PodScheduled && cond.Status == corev1.ConditionFalse && cond.Message != "" {
			details = append(details, fmt.Sprintf("Exact Cause: %s", cond.Message))
			break
		}
	}

	return []diagnosis.Problem{
		{
			Resource:       "Pod",
			Namespace:      pod.Namespace,
			Name:           pod.Name,
			Issue:          "Pod is stuck in Pending state",
			Severity:       diagnosis.High,
			RemediationKey: "pod-pending",
			Details:        details,
		},
	}
}
