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

func TestPersistentVolumeClaimScanner_Name(t *testing.T) {
	t.Parallel()
	s := NewPersistentVolumeClaimScanner()
	if got := s.Name(); got != "PersistentVolumeClaim" {
		t.Errorf("Name() = %q, want %q", got, "PersistentVolumeClaim")
	}
}

func TestPersistentVolumeClaimScanner_Scan(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		pvcs         []corev1.PersistentVolumeClaim
		wantCount    int
		wantSeverity diagnosis.Severity
		wantRemedKey string
		wantIssueHas string
	}{
		{
			name: "PVC in Pending state detected",
			pvcs: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "pending-pvc", Namespace: "default"},
					Status: corev1.PersistentVolumeClaimStatus{
						Phase: corev1.ClaimPending,
					},
				},
			},
			wantCount:    1,
			wantSeverity: diagnosis.High,
			wantRemedKey: "pvc_pending",
			wantIssueHas: "Pending",
		},
		{
			name: "PVC in Lost state detected",
			pvcs: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "lost-pvc", Namespace: "default"},
					Status: corev1.PersistentVolumeClaimStatus{
						Phase: corev1.ClaimLost,
					},
				},
			},
			wantCount:    1,
			wantSeverity: diagnosis.Critical,
			wantRemedKey: "pvc_lost",
			wantIssueHas: "Lost",
		},
		{
			name: "healthy PVC in Bound state not reported",
			pvcs: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "healthy-pvc", Namespace: "default"},
					Status: corev1.PersistentVolumeClaimStatus{
						Phase: corev1.ClaimBound,
					},
				},
			},
			wantCount: 0,
		},
		{
			name:      "no PVCs produces zero problems",
			pvcs:      nil,
			wantCount: 0,
		},
		{
			name: "multiple Pending PVCs all detected",
			pvcs: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "pending-pvc-1", Namespace: "default"},
					Status: corev1.PersistentVolumeClaimStatus{
						Phase: corev1.ClaimPending,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "pending-pvc-2", Namespace: "production"},
					Status: corev1.PersistentVolumeClaimStatus{
						Phase: corev1.ClaimPending,
					},
				},
			},
			wantCount:    2,
			wantSeverity: diagnosis.High,
		},
		{
			name: "multiple Lost PVCs all detected",
			pvcs: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "lost-pvc-1", Namespace: "default"},
					Status: corev1.PersistentVolumeClaimStatus{
						Phase: corev1.ClaimLost,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "lost-pvc-2", Namespace: "staging"},
					Status: corev1.PersistentVolumeClaimStatus{
						Phase: corev1.ClaimLost,
					},
				},
			},
			wantCount:    2,
			wantSeverity: diagnosis.Critical,
		},
		{
			name: "mixed PVC states",
			pvcs: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "pvc-bound", Namespace: "default"},
					Status: corev1.PersistentVolumeClaimStatus{
						Phase: corev1.ClaimBound,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "pvc-pending", Namespace: "default"},
					Status: corev1.PersistentVolumeClaimStatus{
						Phase: corev1.ClaimPending,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "pvc-lost", Namespace: "default"},
					Status: corev1.PersistentVolumeClaimStatus{
						Phase: corev1.ClaimLost,
					},
				},
			},
			wantCount: 2, // Pending and Lost only
		},
		{
			name: "Lost PVC has Critical severity",
			pvcs: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "critical-pvc", Namespace: "production"},
					Status: corev1.PersistentVolumeClaimStatus{
						Phase: corev1.ClaimLost,
					},
				},
			},
			wantCount:    1,
			wantSeverity: diagnosis.Critical,
		},
		{
			name: "Pending PVC has High severity",
			pvcs: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "stuck-pvc", Namespace: "default"},
					Status: corev1.PersistentVolumeClaimStatus{
						Phase: corev1.ClaimPending,
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
			client := fake.NewSimpleClientset(pvcsToObjects(tt.pvcs)...)
			s := NewPersistentVolumeClaimScanner()

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

func pvcsToObjects(pvcs []corev1.PersistentVolumeClaim) []runtime.Object {
	objs := make([]runtime.Object, len(pvcs))
	for i := range pvcs {
		objs[i] = &pvcs[i]
	}
	return objs
}
