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

func TestReplicaSetScanner_Name(t *testing.T) {
	t.Parallel()
	s := NewReplicaSetScanner()
	if got := s.Name(); got != "ReplicaSet" {
		t.Errorf("Name() = %q, want %q", got, "ReplicaSet")
	}
}

func TestReplicaSetScanner_Scan(t *testing.T) {
	t.Parallel()

	deploymentOwner := metav1.OwnerReference{
		Kind:       "Deployment",
		Name:       "my-deployment",
		APIVersion: "apps/v1",
	}

	tests := []struct {
		name         string
		replicasets  []appsv1.ReplicaSet
		wantCount    int
		wantSeverity diagnosis.Severity
		wantRemedKey string
		wantIssueHas string
	}{
		{
			name: "orphaned ReplicaSet with active replicas detected",
			replicasets: []appsv1.ReplicaSet{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "orphaned-rs", Namespace: "default"},
					Status: appsv1.ReplicaSetStatus{
						Replicas:      5,
						ReadyReplicas: 5,
					},
				},
			},
			wantCount:    1,
			wantSeverity: diagnosis.Medium,
			wantRemedKey: "replicaset_orphaned",
			wantIssueHas: "Orphaned",
		},
		{
			name: "ReplicaSet with owner reference not reported as orphaned",
			replicasets: []appsv1.ReplicaSet{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:            "owned-rs",
						Namespace:       "default",
						OwnerReferences: []metav1.OwnerReference{deploymentOwner},
					},
					Status: appsv1.ReplicaSetStatus{
						Replicas:      3,
						ReadyReplicas: 3,
					},
				},
			},
			wantCount: 0,
		},
		{
			name: "orphaned ReplicaSet with zero replicas not reported",
			replicasets: []appsv1.ReplicaSet{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "empty-orphaned-rs", Namespace: "default"},
					Status: appsv1.ReplicaSetStatus{
						Replicas:      0,
						ReadyReplicas: 0,
					},
				},
			},
			wantCount: 0,
		},
		{
			name:        "no ReplicaSets produces zero problems",
			replicasets: nil,
			wantCount:   0,
		},
		{
			name: "multiple orphaned ReplicaSets all detected",
			replicasets: []appsv1.ReplicaSet{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "orphaned-rs-1", Namespace: "default"},
					Status: appsv1.ReplicaSetStatus{
						Replicas:      3,
						ReadyReplicas: 3,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "orphaned-rs-2", Namespace: "production"},
					Status: appsv1.ReplicaSetStatus{
						Replicas:      5,
						ReadyReplicas: 5,
					},
				},
			},
			wantCount:    2,
			wantSeverity: diagnosis.Medium,
		},
		{
			name: "mixed ReplicaSet statuses",
			replicasets: []appsv1.ReplicaSet{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:            "owned-rs",
						Namespace:       "default",
						OwnerReferences: []metav1.OwnerReference{deploymentOwner},
					},
					Status: appsv1.ReplicaSetStatus{
						Replicas:      1,
						ReadyReplicas: 1,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "orphaned-rs", Namespace: "default"},
					Status: appsv1.ReplicaSetStatus{
						Replicas:      3,
						ReadyReplicas: 3,
					},
				},
			},
			wantCount:    1,
			wantSeverity: diagnosis.Medium,
		},
		{
			name: "ReplicaSet with replica mismatch detected",
			replicasets: []appsv1.ReplicaSet{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:            "mismatch-rs",
						Namespace:       "default",
						OwnerReferences: []metav1.OwnerReference{deploymentOwner},
					},
					Status: appsv1.ReplicaSetStatus{
						Replicas:      5,
						ReadyReplicas: 2,
					},
				},
			},
			wantCount:    1,
			wantSeverity: diagnosis.High,
			wantRemedKey: "replicaset_mismatch",
			wantIssueHas: "mismatch",
		},
		{
			name: "healthy ReplicaSet with all replicas ready not reported",
			replicasets: []appsv1.ReplicaSet{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:            "healthy-rs",
						Namespace:       "default",
						OwnerReferences: []metav1.OwnerReference{deploymentOwner},
					},
					Status: appsv1.ReplicaSetStatus{
						Replicas:      3,
						ReadyReplicas: 3,
					},
				},
			},
			wantCount: 0,
		},
		{
			name: "orphaned ReplicaSet with single replica detected",
			replicasets: []appsv1.ReplicaSet{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "orphaned-single", Namespace: "default"},
					Status: appsv1.ReplicaSetStatus{
						Replicas:      1,
						ReadyReplicas: 1,
					},
				},
			},
			wantCount:    1,
			wantSeverity: diagnosis.Medium,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			client := fake.NewSimpleClientset(replicasetsToObjects(tt.replicasets)...)
			s := NewReplicaSetScanner()

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

func replicasetsToObjects(replicasets []appsv1.ReplicaSet) []runtime.Object {
	objs := make([]runtime.Object, len(replicasets))
	for i := range replicasets {
		objs[i] = &replicasets[i]
	}
	return objs
}
