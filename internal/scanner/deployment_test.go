package scanner

import (
	"context"
	"testing"

	"github.com/jorgegabrielti/nautikube/internal/diagnosis"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestDeploymentScanner_Name(t *testing.T) {
	t.Parallel()
	s := NewDeploymentScanner()
	if got := s.Name(); got != "Deployment" {
		t.Errorf("Name() = %q, want %q", got, "Deployment")
	}
}

func TestDeploymentScanner_Scan(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		dep          appsv1.Deployment
		wantCount    int
		wantSeverity diagnosis.Severity
		wantRemedKey string
		wantIssueHas string
	}{
		{
			name: "unavailable replicas",
			dep: appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{Name: "bad-dep", Namespace: "default"},
				Spec:       appsv1.DeploymentSpec{Replicas: int32Ptr(3)},
				Status: appsv1.DeploymentStatus{
					ReadyReplicas:       1,
					UnavailableReplicas: 2,
				},
			},
			wantCount:    1,
			wantSeverity: diagnosis.High,
			wantRemedKey: "deployment-unavailable",
			wantIssueHas: "unavailable",
		},
		{
			name: "replicas mismatch without unavailable",
			dep: appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{Name: "mismatch-dep", Namespace: "default"},
				Spec:       appsv1.DeploymentSpec{Replicas: int32Ptr(3)},
				Status: appsv1.DeploymentStatus{
					ReadyReplicas:       2,
					UnavailableReplicas: 0,
				},
			},
			wantCount:    1,
			wantSeverity: diagnosis.Medium,
			wantRemedKey: "deployment-mismatch",
		},
		{
			name: "rollout stuck",
			dep: appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{Name: "stuck-dep", Namespace: "default"},
				Spec:       appsv1.DeploymentSpec{Replicas: int32Ptr(3)},
				Status: appsv1.DeploymentStatus{
					ReadyReplicas: 3,
					Conditions: []appsv1.DeploymentCondition{
						{
							Type:   appsv1.DeploymentProgressing,
							Status: corev1.ConditionFalse,
							Reason: "ProgressDeadlineExceeded",
						},
					},
				},
			},
			wantCount:    1,
			wantSeverity: diagnosis.High,
			wantRemedKey: "deployment-stuck",
			wantIssueHas: "Rollout stuck",
		},
		{
			name: "zero replicas",
			dep: appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{Name: "zero-dep", Namespace: "default"},
				Spec:       appsv1.DeploymentSpec{Replicas: int32Ptr(0)},
			},
			wantCount:    1,
			wantSeverity: diagnosis.Low,
			wantRemedKey: "deployment-zero-replicas",
		},
		{
			name: "nil spec replicas defaults to 1",
			dep: appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{Name: "nil-dep", Namespace: "default"},
				Spec:       appsv1.DeploymentSpec{Replicas: nil},
				Status: appsv1.DeploymentStatus{
					ReadyReplicas: 1,
				},
			},
			wantCount: 0,
		},
		{
			name: "healthy deployment",
			dep: appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{Name: "healthy-dep", Namespace: "default"},
				Spec:       appsv1.DeploymentSpec{Replicas: int32Ptr(3)},
				Status: appsv1.DeploymentStatus{
					ReadyReplicas:       3,
					UnavailableReplicas: 0,
				},
			},
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			client := fake.NewSimpleClientset(&tt.dep)
			s := NewDeploymentScanner()

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

func int32Ptr(i int32) *int32 { return &i }
