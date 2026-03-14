package scanner

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jorgegabrielti/nautikube/internal/diagnosis"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// EventScanner aggregates Warning events to surface recurring issues.
type EventScanner struct{}

// NewEventScanner creates a new EventScanner.
func NewEventScanner() *EventScanner {
	return &EventScanner{}
}

// Name returns the scanner name.
func (s *EventScanner) Name() string { return "Event" }

// Scan aggregates Warning events from the last hour and reports high-frequency ones.
func (s *EventScanner) Scan(ctx context.Context, client kubernetes.Interface, namespace string) ([]diagnosis.Problem, error) {
	events, err := client.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{
		FieldSelector: "type=Warning",
	})
	if err != nil {
		slog.Warn("failed to list events", "namespace", namespace, "error", err)
		return nil, fmt.Errorf("failed to list events: %w", err)
	}

	cutoff := time.Now().Add(-1 * time.Hour)
	var problems []diagnosis.Problem

	for _, event := range events.Items {
		// Filter to events from the last hour
		eventTime := event.LastTimestamp.Time
		if eventTime.IsZero() {
			eventTime = event.CreationTimestamp.Time
		}
		if eventTime.Before(cutoff) {
			continue
		}

		count := event.Count
		if count <= 5 {
			continue
		}

		sev := diagnosis.Low
		if count > 50 {
			sev = diagnosis.High
		} else if count > 20 {
			sev = diagnosis.Medium
		}

		problems = append(problems, diagnosis.Problem{
			Resource:       event.InvolvedObject.Kind,
			Namespace:      event.InvolvedObject.Namespace,
			Name:           event.InvolvedObject.Name,
			Issue:          fmt.Sprintf("%s: %s (seen %d times)", event.Reason, event.Message, count),
			Severity:       sev,
			RemediationKey: "frequent-events",
		})
	}

	return problems, nil
}
