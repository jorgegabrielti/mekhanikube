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

func TestServiceAccountScanner_Name(t *testing.T) {
	t.Parallel()
	s := NewServiceAccountScanner()
	if got := s.Name(); got != "ServiceAccount" {
		t.Errorf("Name() = %q, want %q", got, "ServiceAccount")
	}
}

func TestServiceAccountScanner_Scan(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		sas          []corev1.ServiceAccount
		wantCount    int
		wantSeverity diagnosis.Severity
		wantRemedKey string
		wantIssueHas string
	}{
		{
			name: "ServiceAccount with no secrets detected",
			sas: []corev1.ServiceAccount{
				{
					ObjectMeta:       metav1.ObjectMeta{Name: "empty-sa", Namespace: "default"},
					Secrets:          []corev1.ObjectReference{},
					ImagePullSecrets: []corev1.LocalObjectReference{},
				},
			},
			wantCount:    1,
			wantSeverity: diagnosis.Low,
			wantRemedKey: "serviceaccount_no_secrets",
			wantIssueHas: "no secrets",
		},
		{
			name: "ServiceAccount with Secrets not reported",
			sas: []corev1.ServiceAccount{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "healthy-sa", Namespace: "default"},
					Secrets: []corev1.ObjectReference{
						{
							Name:      "sa-token-abc123",
							Namespace: "default",
						},
					},
					ImagePullSecrets: []corev1.LocalObjectReference{},
				},
			},
			wantCount: 0,
		},
		{
			name: "ServiceAccount with ImagePullSecrets not reported",
			sas: []corev1.ServiceAccount{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "pull-sa", Namespace: "default"},
					Secrets:    []corev1.ObjectReference{},
					ImagePullSecrets: []corev1.LocalObjectReference{
						{
							Name: "docker-registry-secret",
						},
					},
				},
			},
			wantCount: 0,
		},
		{
			name: "ServiceAccount with both Secrets and ImagePullSecrets not reported",
			sas: []corev1.ServiceAccount{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "full-sa", Namespace: "default"},
					Secrets: []corev1.ObjectReference{
						{
							Name:      "sa-token-abc123",
							Namespace: "default",
						},
					},
					ImagePullSecrets: []corev1.LocalObjectReference{
						{
							Name: "docker-registry-secret",
						},
					},
				},
			},
			wantCount: 0,
		},
		{
			name:      "no ServiceAccounts produces zero problems",
			sas:       nil,
			wantCount: 0,
		},
		{
			name: "multiple empty ServiceAccounts all detected",
			sas: []corev1.ServiceAccount{
				{
					ObjectMeta:       metav1.ObjectMeta{Name: "empty-sa-1", Namespace: "default"},
					Secrets:          []corev1.ObjectReference{},
					ImagePullSecrets: []corev1.LocalObjectReference{},
				},
				{
					ObjectMeta:       metav1.ObjectMeta{Name: "empty-sa-2", Namespace: "kube-system"},
					Secrets:          []corev1.ObjectReference{},
					ImagePullSecrets: []corev1.LocalObjectReference{},
				},
			},
			wantCount:    2,
			wantSeverity: diagnosis.Low,
		},
		{
			name: "mixed ServiceAccounts",
			sas: []corev1.ServiceAccount{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "default", Namespace: "default"},
					Secrets: []corev1.ObjectReference{
						{
							Name:      "sa-token-xyz789",
							Namespace: "default",
						},
					},
					ImagePullSecrets: []corev1.LocalObjectReference{},
				},
				{
					ObjectMeta:       metav1.ObjectMeta{Name: "empty-sa", Namespace: "default"},
					Secrets:          []corev1.ObjectReference{},
					ImagePullSecrets: []corev1.LocalObjectReference{},
				},
			},
			wantCount:    1, // Only the empty one
			wantSeverity: diagnosis.Low,
		},
		{
			name: "ServiceAccount with only Secrets (no ImagePullSecrets)",
			sas: []corev1.ServiceAccount{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "token-only-sa", Namespace: "production"},
					Secrets: []corev1.ObjectReference{
						{
							Name:      "token-abc",
							Namespace: "production",
						},
					},
					ImagePullSecrets: []corev1.LocalObjectReference{},
				},
			},
			wantCount: 0,
		},
		{
			name: "ServiceAccount with multiple ImagePullSecrets",
			sas: []corev1.ServiceAccount{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "multi-pull-sa", Namespace: "default"},
					Secrets:    []corev1.ObjectReference{},
					ImagePullSecrets: []corev1.LocalObjectReference{
						{
							Name: "docker-secret-1",
						},
						{
							Name: "docker-secret-2",
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
			client := fake.NewSimpleClientset(sasToObjects(tt.sas)...)
			s := NewServiceAccountScanner()

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

func sasToObjects(sas []corev1.ServiceAccount) []runtime.Object {
	objs := make([]runtime.Object, len(sas))
	for i := range sas {
		objs[i] = &sas[i]
	}
	return objs
}
