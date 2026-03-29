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

func TestStatefulSetScanner_Name(t *testing.T) {
	t.Parallel()
	s := NewStatefulSetScanner()
	if got := s.Name(); got != "StatefulSet" {
		t.Errorf("Name() = %q, want %q", got, "StatefulSet")
	}
}

func TestStatefulSetScanner_Scan(t *testing.T) {
	t.Parallel()

	var (
		replicas1 int32 = 1
		replicas3 int32 = 3
		replicas5 int32 = 5
	)

	tests := []struct {
		name         string
		statefulsets []appsv1.StatefulSet
		wantCount    int
		wantSeverity diagnosis.Severity
		wantRemedKey string
		wantIssueHas string
	}{
		{
			name: "StatefulSet with replica mismatch detected",
			statefulsets: []appsv1.StatefulSet{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "mismatch-sts", Namespace: "default"},
					Spec: appsv1.StatefulSetSpec{
						Replicas: &replicas3,
					},
					Status: appsv1.StatefulSetStatus{
						Replicas:      3,
						ReadyReplicas: 1, // Mismatch
					},
				},
			},
			wantCount:    1,
			wantSeverity: diagnosis.High,
			wantRemedKey: "statefulset_mismatch",
			wantIssueHas: "mismatch",
		},
		{
			name: "healthy StatefulSet with all replicas ready not reported",
			statefulsets: []appsv1.StatefulSet{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "healthy-sts", Namespace: "default"},
					Spec: appsv1.StatefulSetSpec{
						Replicas: &replicas3,
					},
					Status: appsv1.StatefulSetStatus{
						Replicas:      3,
						ReadyReplicas: 3, // All ready
					},
				},
			},
			wantCount: 0,
		},
		{
			name: "StatefulSet with zero ready replicas detected",
			statefulsets: []appsv1.StatefulSet{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "zero-ready-sts", Namespace: "default"},
					Spec: appsv1.StatefulSetSpec{
						Replicas: &replicas5,
					},
					Status: appsv1.StatefulSetStatus{
						Replicas:      5,
						ReadyReplicas: 0, // None ready
					},
				},
			},
			wantCount:    1,
			wantSeverity: diagnosis.High,
		},
		{
			name:         "no StatefulSets produces zero problems",
			statefulsets: nil,
			wantCount:    0,
		},
		{
			name: "multiple StatefulSets with mismatches all detected",
			statefulsets: []appsv1.StatefulSet{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "mismatch-sts-1", Namespace: "default"},
					Spec: appsv1.StatefulSetSpec{
						Replicas: &replicas3,
					},
					Status: appsv1.StatefulSetStatus{
						Replicas:      3,
						ReadyReplicas: 1,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "mismatch-sts-2", Namespace: "production"},
					Spec: appsv1.StatefulSetSpec{
						Replicas: &replicas5,
					},
					Status: appsv1.StatefulSetStatus{
						Replicas:      5,
						ReadyReplicas: 2,
					},
				},
			},
			wantCount:    2,
			wantSeverity: diagnosis.High,
		},
		{
			name: "mixed StatefulSet statuses",
			statefulsets: []appsv1.StatefulSet{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "healthy-sts", Namespace: "default"},
					Spec: appsv1.StatefulSetSpec{
						Replicas: &replicas1,
					},
					Status: appsv1.StatefulSetStatus{
						Replicas:      1,
						ReadyReplicas: 1,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "unhealthy-sts", Namespace: "default"},
					Spec: appsv1.StatefulSetSpec{
						Replicas: &replicas3,
					},
					Status: appsv1.StatefulSetStatus{
						Replicas:      3,
						ReadyReplicas: 0,
					},
				},
			},
			wantCount:    1,
			wantSeverity: diagnosis.High,
		},
		{
			name: "StatefulSet with partial readiness detected",
			statefulsets: []appsv1.StatefulSet{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "partial-sts", Namespace: "staging"},
					Spec: appsv1.StatefulSetSpec{
						Replicas: &replicas5,
					},
					Status: appsv1.StatefulSetStatus{
						Replicas:      5,
						ReadyReplicas: 3,
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
			client := fake.NewSimpleClientset(statefulsetsToObjects(tt.statefulsets)...)
			s := NewStatefulSetScanner()

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

func statefulsetsToObjects(statefulsets []appsv1.StatefulSet) []runtime.Object {
	objs := make([]runtime.Object, len(statefulsets))
	for i := range statefulsets {
		objs[i] = &statefulsets[i]
	}
	return objs
}
