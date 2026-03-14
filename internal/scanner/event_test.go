package scanner

import (
	"context"
	"testing"
	"time"

	"github.com/jorgegabrielti/nautikube/internal/diagnosis"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

func TestEventScanner_Name(t *testing.T) {
	t.Parallel()
	s := NewEventScanner()
	if got := s.Name(); got != "Event" {
		t.Errorf("Name() = %q, want %q", got, "Event")
	}
}

func TestEventScanner_Scan(t *testing.T) {
	t.Parallel()

	now := metav1.Now()
	twoHoursAgo := metav1.NewTime(time.Now().Add(-2 * time.Hour))

	tests := []struct {
		name         string
		events       []corev1.Event
		wantCount    int
		wantSeverity diagnosis.Severity
	}{
		{
			name: "high frequency warning event over 50",
			events: []corev1.Event{
				warningEvent("pod-1", "default", "FailedMount", "Unable to mount volume", 55, now),
			},
			wantCount:    1,
			wantSeverity: diagnosis.High,
		},
		{
			name: "medium frequency warning event over 20",
			events: []corev1.Event{
				warningEvent("pod-2", "default", "Unhealthy", "Readiness probe failed", 25, now),
			},
			wantCount:    1,
			wantSeverity: diagnosis.Medium,
		},
		{
			name: "low frequency warning event over 5",
			events: []corev1.Event{
				warningEvent("pod-3", "default", "BackOff", "Back-off pulling image", 10, now),
			},
			wantCount:    1,
			wantSeverity: diagnosis.Low,
		},
		{
			name: "event with count 5 or below not reported",
			events: []corev1.Event{
				warningEvent("pod-4", "default", "Pulling", "Pulling image", 5, now),
			},
			wantCount: 0,
		},
		{
			name: "old event filtered out",
			events: []corev1.Event{
				warningEvent("pod-5", "default", "FailedMount", "Old event", 100, twoHoursAgo),
			},
			wantCount: 0,
		},
		{
			name:      "no events produces zero problems",
			events:    nil,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			objs := make([]runtime.Object, len(tt.events))
			for i := range tt.events {
				objs[i] = &tt.events[i]
			}
			client := fake.NewSimpleClientset(objs...)
			s := NewEventScanner()

			problems, err := s.Scan(context.Background(), client, "")
			if err != nil {
				t.Fatalf("Scan() error = %v", err)
			}

			if len(problems) != tt.wantCount {
				t.Fatalf("Scan() returned %d problems, want %d", len(problems), tt.wantCount)
			}

			if tt.wantCount > 0 && tt.wantSeverity != "" {
				if problems[0].Severity != tt.wantSeverity {
					t.Errorf("Severity = %q, want %q", problems[0].Severity, tt.wantSeverity)
				}
			}
		})
	}
}

func warningEvent(objName, ns, reason, message string, count int32, lastTime metav1.Time) corev1.Event {
	return corev1.Event{
		ObjectMeta:     metav1.ObjectMeta{Name: objName + "-event", Namespace: ns},
		Type:           "Warning",
		Reason:         reason,
		Message:        message,
		Count:          count,
		LastTimestamp:  lastTime,
		InvolvedObject: corev1.ObjectReference{Kind: "Pod", Name: objName, Namespace: ns},
	}
}
