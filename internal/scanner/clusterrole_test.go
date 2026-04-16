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

func TestClusterRoleScanner_Name(t *testing.T) {
	t.Parallel()
	s := NewClusterRoleScanner()
	if got := s.Name(); got != "ClusterRole" {
		t.Errorf("Name() = %q, want %q", got, "ClusterRole")
	}
}

func TestClusterRoleScanner_Scan(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		clusterroles []rbacv1.ClusterRole
		wantCount    int
		wantSeverity diagnosis.Severity
		wantRemedKey string
		wantIssueHas string
	}{
		{
			name: "ClusterRole with wildcard verbs detected",
			clusterroles: []rbacv1.ClusterRole{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "wildcard-cr"},
					Rules: []rbacv1.PolicyRule{
						{
							Verbs:     []string{"*"},
							Resources: []string{"*"},
						},
					},
				},
			},
			wantCount:    1,
			wantSeverity: diagnosis.High,
			wantRemedKey: "rbac_cluster_wildcard",
			wantIssueHas: "wildcard",
		},
		{
			name: "ClusterRole without wildcards not reported",
			clusterroles: []rbacv1.ClusterRole{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "safe-cr"},
					Rules: []rbacv1.PolicyRule{
						{
							Verbs:     []string{"get", "list", "watch"},
							Resources: []string{"pods"},
						},
					},
				},
			},
			wantCount: 0,
		},
		{
			name: "system: prefixed ClusterRole skipped",
			clusterroles: []rbacv1.ClusterRole{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "system:nodes"},
					Rules: []rbacv1.PolicyRule{
						{
							Verbs:     []string{"*"},
							Resources: []string{"*"},
						},
					},
				},
			},
			wantCount: 0,
		},
		{
			name:         "no ClusterRoles produces zero problems",
			clusterroles: nil,
			wantCount:    0,
		},
		{
			name: "multiple non-system wildcard ClusterRoles all detected",
			clusterroles: []rbacv1.ClusterRole{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "wildcard-cr-1"},
					Rules: []rbacv1.PolicyRule{
						{
							Verbs:     []string{"*"},
							Resources: []string{"pods"},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "wildcard-cr-2"},
					Rules: []rbacv1.PolicyRule{
						{
							Verbs:     []string{"get", "*"},
							Resources: []string{"secrets"},
						},
					},
				},
			},
			wantCount:    2,
			wantSeverity: diagnosis.High,
		},
		{
			name: "mixed ClusterRole permissions",
			clusterroles: []rbacv1.ClusterRole{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "safe-cr"},
					Rules: []rbacv1.PolicyRule{
						{
							Verbs:     []string{"get", "list"},
							Resources: []string{"nodes"},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "wildcard-cr"},
					Rules: []rbacv1.PolicyRule{
						{
							Verbs:     []string{"*"},
							Resources: []string{"configmaps"},
						},
					},
				},
			},
			wantCount:    1,
			wantSeverity: diagnosis.High,
		},
		{
			name: "system: ClusterRole with wildcards not reported",
			clusterroles: []rbacv1.ClusterRole{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "system:kubelet-api-admin"},
					Rules: []rbacv1.PolicyRule{
						{
							Verbs:     []string{"*"},
							Resources: []string{"*"},
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
			client := fake.NewSimpleClientset(clusterrolesToObjects(tt.clusterroles)...)
			s := NewClusterRoleScanner()

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

func clusterrolesToObjects(clusterroles []rbacv1.ClusterRole) []runtime.Object {
	objs := make([]runtime.Object, len(clusterroles))
	for i := range clusterroles {
		objs[i] = &clusterroles[i]
	}
	return objs
}
