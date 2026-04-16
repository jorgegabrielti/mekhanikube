package scanner

import (
	"context"
	"testing"

	"github.com/jorgegabrielti/nautikube/internal/diagnosis"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

func TestDaemonSetScanner_Name(t *testing.T) {
	t.Parallel()
	s := NewDaemonSetScanner()
	if got := s.Name(); got != "DaemonSet" {
		t.Errorf("Name() = %q, want %q", got, "DaemonSet")
	}
}

func TestDaemonSetScanner_Scan(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		daemonsets   []appsv1.DaemonSet
		wantCount    int
		wantSeverity diagnosis.Severity
		wantRemedKey string
		wantIssueHas string
	}{
		{
			name: "DaemonSet not fully scheduled detected",
			daemonsets: []appsv1.DaemonSet{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "unscheduled-ds", Namespace: "default"},
					Status: appsv1.DaemonSetStatus{
						DesiredNumberScheduled: 5,
						NumberReady:            2,
					},
				},
			},
			wantCount:    1,
			wantSeverity: diagnosis.High,
			wantRemedKey: "daemonset_misscheduled",
			wantIssueHas: "not fully scheduled",
		},
		{
			name: "healthy DaemonSet with all nodes ready not reported",
			daemonsets: []appsv1.DaemonSet{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "healthy-ds", Namespace: "default"},
					Status: appsv1.DaemonSetStatus{
						DesiredNumberScheduled: 3,
						NumberReady:            3,
					},
				},
			},
			wantCount: 0,
		},
		{
			name: "DaemonSet with zero ready nodes detected",
			daemonsets: []appsv1.DaemonSet{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "zero-ready-ds", Namespace: "default"},
					Status: appsv1.DaemonSetStatus{
						DesiredNumberScheduled: 10,
						NumberReady:            0,
					},
				},
			},
			wantCount:    1,
			wantSeverity: diagnosis.High,
		},
		{
			name:       "no DaemonSets produces zero problems",
			daemonsets: nil,
			wantCount:  0,
		},
		{
			name: "multiple DaemonSets with scheduling issues all detected",
			daemonsets: []appsv1.DaemonSet{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "unscheduled-ds-1", Namespace: "default"},
					Status: appsv1.DaemonSetStatus{
						DesiredNumberScheduled: 5,
						NumberReady:            1,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "unscheduled-ds-2", Namespace: "kube-system"},
					Status: appsv1.DaemonSetStatus{
						DesiredNumberScheduled: 8,
						NumberReady:            4,
					},
				},
			},
			wantCount:    2,
			wantSeverity: diagnosis.High,
		},
		{
			name: "mixed DaemonSet statuses",
			daemonsets: []appsv1.DaemonSet{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "healthy-ds", Namespace: "default"},
					Status: appsv1.DaemonSetStatus{
						DesiredNumberScheduled: 2,
						NumberReady:            2,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "unhealthy-ds", Namespace: "default"},
					Status: appsv1.DaemonSetStatus{
						DesiredNumberScheduled: 5,
						NumberReady:            0,
					},
				},
			},
			wantCount:    1,
			wantSeverity: diagnosis.High,
		},
		{
			name: "DaemonSet with partial readiness detected",
			daemonsets: []appsv1.DaemonSet{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "partial-ds", Namespace: "monitoring"},
					Status: appsv1.DaemonSetStatus{
						DesiredNumberScheduled: 10,
						NumberReady:            7,
					},
				},
			},
			wantCount:    1,
			wantSeverity: diagnosis.High,
		},
		{
			name: "DaemonSet with single node mismatch",
			daemonsets: []appsv1.DaemonSet{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "single-mismatch-ds", Namespace: "default"},
					Status: appsv1.DaemonSetStatus{
						DesiredNumberScheduled: 3,
						NumberReady:            2,
					},
				},
			},
			wantCount:    1,
			wantSeverity: diagnosis.High,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			client := fake.NewSimpleClientset(daemonsetsToObjects(tt.daemonsets)...)
			s := NewDaemonSetScanner()

			problems, err := s.Scan(context.Background(), client, "")
			if err != nil {
				t.Fatalf("Scan() error = %v", err)
			}

			if len(problems) != tt.wantCount {
				t.Fatalf("Scan() returned %d problems, want %d", len(problems), tt.wantCount)
			}

			if tt.wantCount == 0 {
				return
			}

			p := problems[0]
			if tt.wantSeverity != "" && p.Severity != tt.wantSeverity {
				t.Errorf("Severity = %q, want %q", p.Severity, tt.wantSeverity)
			}
			if tt.wantRemedKey != "" && p.RemediationKey != tt.wantRemedKey {
				t.Errorf("RemediationKey = %q, want %q", p.RemediationKey, tt.wantRemedKey)
			}
			if tt.wantIssueHas != "" && !containsStr(p.Issue, tt.wantIssueHas) {
				t.Errorf("Issue = %q, want it to contain %q", p.Issue, tt.wantIssueHas)
			}
		})
	}
}

// --- test helpers ---

func daemonsetsToObjects(daemonsets []appsv1.DaemonSet) []runtime.Object {
	objs := make([]runtime.Object, len(daemonsets))
	for i := range daemonsets {
		objs[i] = &daemonsets[i]
	}
	return objs
}
