package scanner

import (
	"context"
	"testing"

	"github.com/jorgegabrielti/nautikube/internal/diagnosis"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

func TestPersistentVolumeScanner_Name(t *testing.T) {
	t.Parallel()
	s := NewPersistentVolumeScanner()
	if got := s.Name(); got != "PersistentVolume" {
		t.Errorf("Name() = %q, want %q", got, "PersistentVolume")
	}
}

func TestPersistentVolumeScanner_Scan(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		pvs          []corev1.PersistentVolume
		wantCount    int
		wantSeverity diagnosis.Severity
		wantRemedKey string
		wantIssueHas string
	}{
		{
			name: "PV in Failed state detected",
			pvs: []corev1.PersistentVolume{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "failed-pv"},
					Status: corev1.PersistentVolumeStatus{
						Phase: "Failed",
					},
				},
			},
			wantCount:    1,
			wantSeverity: diagnosis.High,
			wantRemedKey: "pv_failed",
			wantIssueHas: "Failed state",
		},
		{
			name: "PV in Released state detected",
			pvs: []corev1.PersistentVolume{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "released-pv"},
					Status: corev1.PersistentVolumeStatus{
						Phase: "Released",
					},
				},
			},
			wantCount:    1,
			wantSeverity: diagnosis.Low,
			wantRemedKey: "pv_released",
			wantIssueHas: "Released",
		},
		{
			name: "healthy PV in Bound state not reported",
			pvs: []corev1.PersistentVolume{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "healthy-pv"},
					Status: corev1.PersistentVolumeStatus{
						Phase: "Bound",
					},
				},
			},
			wantCount: 0,
		},
		{
			name: "PV in Available state not reported",
			pvs: []corev1.PersistentVolume{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "available-pv"},
					Status: corev1.PersistentVolumeStatus{
						Phase: "Available",
					},
				},
			},
			wantCount: 0,
		},
		{
			name:      "no PVs produces zero problems",
			pvs:       nil,
			wantCount: 0,
		},
		{
			name: "multiple Failed PVs all detected",
			pvs: []corev1.PersistentVolume{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "failed-pv-1"},
					Status: corev1.PersistentVolumeStatus{
						Phase: "Failed",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "failed-pv-2"},
					Status: corev1.PersistentVolumeStatus{
						Phase: "Failed",
					},
				},
			},
			wantCount:    2,
			wantSeverity: diagnosis.High,
		},
		{
			name: "multiple Released PVs all detected",
			pvs: []corev1.PersistentVolume{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "released-pv-1"},
					Status: corev1.PersistentVolumeStatus{
						Phase: "Released",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "released-pv-2"},
					Status: corev1.PersistentVolumeStatus{
						Phase: "Released",
					},
				},
			},
			wantCount:    2,
			wantSeverity: diagnosis.Low,
		},
		{
			name: "mixed PV states",
			pvs: []corev1.PersistentVolume{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "pv-bound"},
					Status: corev1.PersistentVolumeStatus{
						Phase: "Bound",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "pv-failed"},
					Status: corev1.PersistentVolumeStatus{
						Phase: "Failed",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "pv-released"},
					Status: corev1.PersistentVolumeStatus{
						Phase: "Released",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "pv-available"},
					Status: corev1.PersistentVolumeStatus{
						Phase: "Available",
					},
				},
			},
			wantCount: 2, // Failed and Released only
		},
		{
			name: "Failed PV has Error in remediation",
			pvs: []corev1.PersistentVolume{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "error-pv"},
					Status: corev1.PersistentVolumeStatus{
						Phase: "Failed",
					},
				},
			},
			wantCount:    1,
			wantRemedKey: "pv_failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			client := fake.NewSimpleClientset(pvsToObjects(tt.pvs)...)
			s := NewPersistentVolumeScanner()

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

func pvsToObjects(pvs []corev1.PersistentVolume) []runtime.Object {
	objs := make([]runtime.Object, len(pvs))
	for i := range pvs {
		objs[i] = &pvs[i]
	}
	return objs
}
