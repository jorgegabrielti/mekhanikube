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

func TestRoleBindingScanner_Name(t *testing.T) {
	t.Parallel()
	s := NewRoleBindingScanner()
	if got := s.Name(); got != "RoleBinding" {
		t.Errorf("Name() = %q, want %q", got, "RoleBinding")
	}
}

func TestRoleBindingScanner_Scan(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		rolebindings []rbacv1.RoleBinding
		wantCount    int
		wantSeverity diagnosis.Severity
		wantRemedKey string
		wantIssueHas string
	}{
		{
			name: "RoleBinding with no subjects detected",
			rolebindings: []rbacv1.RoleBinding{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "empty-binding", Namespace: "default"},
					Subjects:   []rbacv1.Subject{},
				},
			},
			wantCount:    1,
			wantSeverity: diagnosis.Low,
			wantRemedKey: "rbac_empty_binding",
			wantIssueHas: "no subjects",
		},
		{
			name: "RoleBinding with subjects not reported",
			rolebindings: []rbacv1.RoleBinding{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "healthy-binding", Namespace: "default"},
					Subjects: []rbacv1.Subject{
						{
							Kind: "ServiceAccount",
							Name: "app-sa",
						},
					},
				},
			},
			wantCount: 0,
		},
		{
			name:         "no RoleBindings produces zero problems",
			rolebindings: nil,
			wantCount:    0,
		},
		{
			name: "multiple empty RoleBindings all detected",
			rolebindings: []rbacv1.RoleBinding{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "empty-binding-1", Namespace: "default"},
					Subjects:   []rbacv1.Subject{},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "empty-binding-2", Namespace: "kube-system"},
					Subjects:   []rbacv1.Subject{},
				},
			},
			wantCount:    2,
			wantSeverity: diagnosis.Low,
		},
		{
			name: "mixed RoleBinding statuses",
			rolebindings: []rbacv1.RoleBinding{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "healthy-binding", Namespace: "default"},
					Subjects: []rbacv1.Subject{
						{
							Kind: "User",
							Name: "admin@example.com",
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "empty-binding", Namespace: "default"},
					Subjects:   []rbacv1.Subject{},
				},
			},
			wantCount:    1,
			wantSeverity: diagnosis.Low,
		},
		{
			name: "RoleBinding with single subject not reported",
			rolebindings: []rbacv1.RoleBinding{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "single-subject-binding", Namespace: "default"},
					Subjects: []rbacv1.Subject{
						{
							Kind: "Group",
							Name: "editors",
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
			client := fake.NewSimpleClientset(rolebindingsToObjects(tt.rolebindings)...)
			s := NewRoleBindingScanner()

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

func rolebindingsToObjects(rolebindings []rbacv1.RoleBinding) []runtime.Object {
	objs := make([]runtime.Object, len(rolebindings))
	for i := range rolebindings {
		objs[i] = &rolebindings[i]
	}
	return objs
}
