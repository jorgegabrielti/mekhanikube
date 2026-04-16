package scanner

import (
	"context"
	"testing"

	"github.com/jorgegabrielti/nautikube/internal/diagnosis"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

func TestClusterRoleBindingScanner_Name(t *testing.T) {
	t.Parallel()
	s := NewClusterRoleBindingScanner()
	if got := s.Name(); got != "ClusterRoleBinding" {
		t.Errorf("Name() = %q, want %q", got, "ClusterRoleBinding")
	}
}

func TestClusterRoleBindingScanner_Scan(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                string
		clusterrolebindings []rbacv1.ClusterRoleBinding
		wantCount           int
		wantSeverity        diagnosis.Severity
		wantRemedKey        string
		wantIssueHas        string
	}{
		{
			name: "ClusterRoleBinding with no subjects detected",
			clusterrolebindings: []rbacv1.ClusterRoleBinding{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "empty-crb"},
					Subjects:   []rbacv1.Subject{},
				},
			},
			wantCount:    1,
			wantSeverity: diagnosis.Low,
			wantRemedKey: "rbac_cluster_empty_binding",
			wantIssueHas: "no subjects",
		},
		{
			name: "ClusterRoleBinding with subjects not reported",
			clusterrolebindings: []rbacv1.ClusterRoleBinding{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "healthy-crb"},
					Subjects: []rbacv1.Subject{
						{
							Kind: "Group",
							Name: "cluster-admins",
						},
					},
				},
			},
			wantCount: 0,
		},
		{
			name:                "no ClusterRoleBindings produces zero problems",
			clusterrolebindings: nil,
			wantCount:           0,
		},
		{
			name: "multiple empty ClusterRoleBindings all detected",
			clusterrolebindings: []rbacv1.ClusterRoleBinding{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "empty-crb-1"},
					Subjects:   []rbacv1.Subject{},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "empty-crb-2"},
					Subjects:   []rbacv1.Subject{},
				},
			},
			wantCount:    2,
			wantSeverity: diagnosis.Low,
		},
		{
			name: "mixed ClusterRoleBinding statuses",
			clusterrolebindings: []rbacv1.ClusterRoleBinding{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "healthy-crb"},
					Subjects: []rbacv1.Subject{
						{
							Kind: "ServiceAccount",
							Name: "admin-sa",
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "empty-crb"},
					Subjects:   []rbacv1.Subject{},
				},
			},
			wantCount:    1,
			wantSeverity: diagnosis.Low,
		},
		{
			name: "ClusterRoleBinding with multiple subjects not reported",
			clusterrolebindings: []rbacv1.ClusterRoleBinding{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "multi-subject-crb"},
					Subjects: []rbacv1.Subject{
						{
							Kind: "User",
							Name: "user1@example.com",
						},
						{
							Kind: "User",
							Name: "user2@example.com",
						},
					},
				},
			},
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			client := fake.NewSimpleClientset(clusterrolebindingsToObjects(tt.clusterrolebindings)...)
			s := NewClusterRoleBindingScanner()

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

func clusterrolebindingsToObjects(crbs []rbacv1.ClusterRoleBinding) []runtime.Object {
	objs := make([]runtime.Object, len(crbs))
	for i := range crbs {
		objs[i] = &crbs[i]
	}
	return objs
}
