package scanner

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/jorgegabrielti/nautikube/internal/diagnosis"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// EventScanner aggregates Warning events to surface recurring issues.
// It deduplicates events by (Kind, Namespace, Name, Reason) and skips
// resource types that already have dedicated scanners.
type EventScanner struct{}

// NewEventScanner creates a new EventScanner.
func NewEventScanner() *EventScanner {
	return &EventScanner{}
}

// Name returns the scanner name.
func (s *EventScanner) Name() string { return "Event" }

// resourcesWithDedicatedScanner lists resource kinds that have their own
// scanner. Events for these kinds are skipped to avoid duplicated diagnostics.
var resourcesWithDedicatedScanner = map[string]bool{
	"Pod":                     true,
	"Deployment":              true,
	"ReplicaSet":              true,
	"StatefulSet":             true,
	"DaemonSet":               true,
	"Service":                 true,
	"Node":                    true,
	"Job":                     true,
	"CronJob":                 true,
	"Ingress":                 true,
	"PersistentVolume":        true,
	"PersistentVolumeClaim":   true,
	"ConfigMap":               true,
	"Secret":                  true,
	"PodDisruptionBudget":     true,
	"ResourceQuota":           true,
	"HorizontalPodAutoscaler": true,
	"ServiceAccount":          true,
	"NetworkPolicy":           true,
	"Role":                    true,
	"RoleBinding":             true,
	"ClusterRole":             true,
	"ClusterRoleBinding":      true,
}

// eventGroup aggregates multiple events for the same object + reason.
type eventGroup struct {
	kind      string
	namespace string
	name      string
	reason    string
	message   string // message from the highest-count event
	total     int32  // sum of all event counts in this group
}

// Scan aggregates Warning events from the last hour, deduplicates by
// (Kind, Namespace, Name, Reason), and reports them with specific remediation.
func (s *EventScanner) Scan(ctx context.Context, client kubernetes.Interface, namespace string) ([]diagnosis.Problem, error) {
	events, err := client.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{
		FieldSelector: "type=Warning",
	})
	if err != nil {
		slog.Warn("failed to list events", "namespace", namespace, "error", err)
		return nil, fmt.Errorf("failed to list events: %w", err)
	}

	cutoff := time.Now().Add(-1 * time.Hour)

	// Deduplicate: group by (Kind, Namespace, Name, Reason)
	type groupKey struct {
		kind, namespace, name, reason string
	}
	groups := make(map[groupKey]*eventGroup)

	for _, event := range events.Items {
		eventTime := event.LastTimestamp.Time
		if eventTime.IsZero() {
			eventTime = event.CreationTimestamp.Time
		}
		if eventTime.Before(cutoff) {
			continue
		}

		if event.Count <= 5 {
			continue
		}

		kind := event.InvolvedObject.Kind
		if resourcesWithDedicatedScanner[kind] {
			continue
		}

		key := groupKey{
			kind:      kind,
			namespace: event.InvolvedObject.Namespace,
			name:      event.InvolvedObject.Name,
			reason:    event.Reason,
		}

		g, ok := groups[key]
		if !ok {
			g = &eventGroup{
				kind:      kind,
				namespace: event.InvolvedObject.Namespace,
				name:      event.InvolvedObject.Name,
				reason:    event.Reason,
				message:   event.Message,
			}
			groups[key] = g
		}

		g.total += event.Count
		// Keep the message from the event with the highest count
		if event.Count > 0 && len(event.Message) > len(g.message) {
			g.message = event.Message
		}
	}

	var problems []diagnosis.Problem
	for _, g := range groups {
		sev := severityForEventCount(g.total)
		remKey := remediationKeyForReason(g.reason, g.message)

		problems = append(problems, diagnosis.Problem{
			Resource:       g.kind,
			Namespace:      g.namespace,
			Name:           g.name,
			Issue:          fmt.Sprintf("%s: %s (seen %d times)", g.reason, g.message, g.total),
			Severity:       sev,
			RemediationKey: remKey,
		})
	}

	return problems, nil
}

// severityForEventCount returns severity based on total event count.
func severityForEventCount(count int32) diagnosis.Severity {
	switch {
	case count > 50:
		return diagnosis.High
	case count > 20:
		return diagnosis.Medium
	default:
		return diagnosis.Low
	}
}

// remediationKeyForReason maps Kubernetes event Reason (and sometimes Message)
// to a specific knowledge-base remediation key.
func remediationKeyForReason(reason, message string) string {
	lower := strings.ToLower(reason)
	lowerMsg := strings.ToLower(message)

	switch {
	// Secret sync failures (ExternalSecret operator, Vault, etc.)
	case lower == "updatefailed" && (strings.Contains(lowerMsg, "vault") || strings.Contains(lowerMsg, "secret does not exist") || strings.Contains(lowerMsg, "externalsecret")):
		return "event-secret-sync-failed"

	// Image pull issues
	case lower == "failed" && strings.Contains(lowerMsg, "invalidimagename"):
		return "event-invalid-image"
	case lower == "failed" && (strings.Contains(lowerMsg, "imagepullbackoff") || strings.Contains(lowerMsg, "errimagepull")):
		return "imagepullbackoff"
	case lower == "failedtoretrieveimagepullsecret":
		return "event-image-pull-secret-missing"
	case lower == "inspectfailed":
		return "event-invalid-image"

	// Mount failures
	case lower == "failedmount":
		return "event-failed-mount"

	// Network / CNI
	case lower == "networknotready":
		return "event-network-not-ready"

	// Probe failures
	case lower == "unhealthy":
		return "event-probe-failed"

	// HPA issues
	case lower == "failedgetscale" || lower == "selectorrequired":
		return "event-hpa-misconfigured"
	case lower == "scaledobjectcheckfailed":
		return "event-scaledobject-failed"

	// Endpoint slice issues
	case lower == "failedtoupdateendpointslices":
		return "event-endpoint-slice-failed"

	// BackOff (generic container restart)
	case lower == "backoff":
		return "crashloopbackoff"

	default:
		return "frequent-events"
	}
}
