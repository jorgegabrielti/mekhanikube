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

func TestRoleScanner_Name(t *testing.T) {
	t.Parallel()
	s := NewRoleScanner()
	if got := s.Name(); got != "Role" {
		t.Errorf("Name() = %q, want %q", got, "Role")
	}
}

func TestRoleScanner_Scan(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		roles        []rbacv1.Role
		wantCount    int
		wantSeverity diagnosis.Severity
		wantRemedKey string
		wantIssueHas string
	}{
		{
			name: "Role with wildcard verbs detected",
			roles: []rbacv1.Role{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "wildcard-role", Namespace: "default"},
					Rules: []rbacv1.PolicyRule{
						{
							Verbs:     []string{"*"},
							Resources: []string{"pods"},
						},
					},
				},
			},
			wantCount:    1,
			wantSeverity: diagnosis.Medium,
			wantRemedKey: "rbac_wildcard",
			wantIssueHas: "wildcard",
		},
		{
			name: "Role with specific verbs not reported",
			roles: []rbacv1.Role{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "safe-role", Namespace: "default"},
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
			name:      "no Roles produces zero problems",
			roles:     nil,
			wantCount: 0,
		},
		{
			name: "multiple wildcard Roles all detected",
			roles: []rbacv1.Role{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "wildcard-role-1", Namespace: "default"},
					Rules: []rbacv1.PolicyRule{
						{
							Verbs:     []string{"*"},
							Resources: []string{"*"},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "wildcard-role-2", Namespace: "production"},
					Rules: []rbacv1.PolicyRule{
						{
							Verbs:     []string{"get", "*"},
							Resources: []string{"pods"},
						},
					},
				},
			},
			wantCount:    2,
			wantSeverity: diagnosis.Medium,
		},
		{
			name: "mixed Role permissions",
			roles: []rbacv1.Role{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "safe-role", Namespace: "default"},
					Rules: []rbacv1.PolicyRule{
						{
							Verbs:     []string{"get", "list"},
							Resources: []string{"pods"},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "wildcard-role", Namespace: "default"},
					Rules: []rbacv1.PolicyRule{
						{
							Verbs:     []string{"*"},
							Resources: []string{"configmaps"},
						},
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
			client := fake.NewSimpleClientset(rolesToObjects(tt.roles)...)
			s := NewRoleScanner()

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

func rolesToObjects(roles []rbacv1.Role) []runtime.Object {
	objs := make([]runtime.Object, len(roles))
	for i := range roles {
		objs[i] = &roles[i]
	}
	return objs
}
