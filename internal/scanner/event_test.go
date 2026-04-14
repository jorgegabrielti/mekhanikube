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
		wantRemKey   string
	}{
		{
			name: "high frequency ExternalSecret event",
			events: []corev1.Event{
				warningEventKind("es-1", "ExternalSecret", "UpdateFailed", "cannot read secret data from Vault", 55, now),
			},
			wantCount:    1,
			wantSeverity: diagnosis.High,
			wantRemKey:   "event-secret-sync-failed",
		},
		{
			name: "medium frequency ScaledObject event",
			events: []corev1.Event{
				warningEventKind("so-1", "ScaledObject", "ScaledObjectCheckFailed", "failed to ensure HPA", 25, now),
			},
			wantCount:    1,
			wantSeverity: diagnosis.Medium,
			wantRemKey:   "event-scaledobject-failed",
		},
		{
			name: "low frequency generic event",
			events: []corev1.Event{
				warningEventKind("cr-1", "CustomResource", "SomeReason", "some message", 10, now),
			},
			wantCount:    1,
			wantSeverity: diagnosis.Low,
			wantRemKey:   "frequent-events",
		},
		{
			name: "event with count 5 or below not reported",
			events: []corev1.Event{
				warningEventKind("obj-1", "ExternalSecret", "UpdateFailed", "transient", 5, now),
			},
			wantCount: 0,
		},
		{
			name: "old event filtered out",
			events: []corev1.Event{
				warningEventKind("obj-2", "ExternalSecret", "UpdateFailed", "old event", 100, twoHoursAgo),
			},
			wantCount: 0,
		},
		{
			name:      "no events produces zero problems",
			events:    nil,
			wantCount: 0,
		},
		{
			name: "Pod events are skipped (dedicated scanner)",
			events: []corev1.Event{
				warningEvent("pod-1", "FailedMount", "Unable to mount volume", 55, now),
			},
			wantCount: 0,
		},
		{
			name: "Deployment events are skipped (dedicated scanner)",
			events: []corev1.Event{
				warningEventKind("deploy-1", "Deployment", "FailedCreate", "quota exceeded", 55, now),
			},
			wantCount: 0,
		},
		{
			name: "deduplication: same object + reason merged",
			events: []corev1.Event{
				warningEventKind("es-dup", "ExternalSecret", "UpdateFailed", "Secret does not exist", 30, now),
				{
					ObjectMeta:     metav1.ObjectMeta{Name: "es-dup-event-2", Namespace: "default"},
					Type:           "Warning",
					Reason:         "UpdateFailed",
					Message:        "Secret does not exist (another key)",
					Count:          40,
					LastTimestamp:  now,
					InvolvedObject: corev1.ObjectReference{Kind: "ExternalSecret", Name: "es-dup", Namespace: "default"},
				},
			},
			wantCount:    1,
			wantSeverity: diagnosis.High, // 30+40=70 > 50
			wantRemKey:   "event-secret-sync-failed",
		},
		{
			name: "HPA events are skipped (dedicated scanner)",
			events: []corev1.Event{
				warningEventKind("hpa-1", "HorizontalPodAutoscaler", "FailedGetScale", "not found", 55, now),
			},
			wantCount: 0,
		},
		{
			name: "ServiceAccount events are skipped (dedicated scanner)",
			events: []corev1.Event{
				warningEventKind("sa-ips", "ServiceAccount", "FailedToRetrieveImagePullSecret", "Unable to retrieve secret", 55, now),
			},
			wantCount: 0,
		},
		{
			name: "NetworkNotReady gets specific key",
			events: []corev1.Event{
				warningEventKind("node-1", "CustomNode", "NetworkNotReady", "network is not ready", 55, now),
			},
			wantCount:  1,
			wantRemKey: "event-network-not-ready",
		},
		{
			name: "Unhealthy probe gets specific key",
			events: []corev1.Event{
				warningEventKind("ep-1", "EndpointSlice", "Unhealthy", "Readiness probe failed", 55, now),
			},
			wantCount:  1,
			wantRemKey: "event-probe-failed",
		},
		{
			name: "FailedMount gets specific key",
			events: []corev1.Event{
				warningEventKind("vol-1", "ExternalSecret", "FailedMount", "MountVolume.SetUp failed", 10, now),
			},
			wantCount:  1,
			wantRemKey: "event-failed-mount",
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

			if tt.wantCount > 0 {
				if tt.wantSeverity != "" && problems[0].Severity != tt.wantSeverity {
					t.Errorf("Severity = %q, want %q", problems[0].Severity, tt.wantSeverity)
				}
				if tt.wantRemKey != "" && problems[0].RemediationKey != tt.wantRemKey {
					t.Errorf("RemediationKey = %q, want %q", problems[0].RemediationKey, tt.wantRemKey)
				}
			}
		})
	}
}

func TestRemediationKeyForReason(t *testing.T) {
	t.Parallel()

	tests := []struct {
		reason  string
		message string
		want    string
	}{
		{"UpdateFailed", "cannot read secret data from Vault: error", "event-secret-sync-failed"},
		{"UpdateFailed", "Secret does not exist", "event-secret-sync-failed"},
		{"UpdateFailed", "something else entirely", "frequent-events"},
		{"FailedToRetrieveImagePullSecret", "Unable to retrieve secret", "event-image-pull-secret-missing"},
		{"InspectFailed", "Failed to apply default image tag", "event-invalid-image"},
		{"FailedMount", "MountVolume.SetUp failed", "event-failed-mount"},
		{"NetworkNotReady", "network is not ready", "event-network-not-ready"},
		{"Unhealthy", "Readiness probe failed", "event-probe-failed"},
		{"FailedGetScale", "deployments.apps not found", "event-hpa-misconfigured"},
		{"SelectorRequired", "selector is required", "event-hpa-misconfigured"},
		{"ScaledObjectCheckFailed", "failed to ensure HPA", "event-scaledobject-failed"},
		{"FailedToUpdateEndpointSlices", "combined from similar events", "event-endpoint-slice-failed"},
		{"BackOff", "Back-off restarting failed container", "crashloopbackoff"},
		{"Failed", "Error: ImagePullBackOff", "imagepullbackoff"},
		{"Failed", "Error: InvalidImageName", "event-invalid-image"},
		{"SomeUnknown", "random message", "frequent-events"},
	}

	for _, tt := range tests {
		t.Run(tt.reason+"_"+tt.message, func(t *testing.T) {
			t.Parallel()
			got := remediationKeyForReason(tt.reason, tt.message)
			if got != tt.want {
				t.Errorf("remediationKeyForReason(%q, %q) = %q, want %q", tt.reason, tt.message, got, tt.want)
			}
		})
	}
}

func warningEvent(objName, reason, message string, count int32, lastTime metav1.Time) corev1.Event {
	return warningEventKind(objName, "Pod", reason, message, count, lastTime)
}

func warningEventKind(objName, kind, reason, message string, count int32, lastTime metav1.Time) corev1.Event {
	return corev1.Event{
		ObjectMeta:     metav1.ObjectMeta{Name: objName + "-event", Namespace: "default"},
		Type:           "Warning",
		Reason:         reason,
		Message:        message,
		Count:          count,
		LastTimestamp:  lastTime,
		InvolvedObject: corev1.ObjectReference{Kind: kind, Name: objName, Namespace: "default"},
	}
}
