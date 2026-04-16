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

func TestSecretScanner_Name(t *testing.T) {
	t.Parallel()
	s := NewSecretScanner()
	if got := s.Name(); got != "Secret" {
		t.Errorf("Name() = %q, want %q", got, "Secret")
	}
}

func TestSecretScanner_Scan(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		secrets      []corev1.Secret
		wantCount    int
		wantSeverity diagnosis.Severity
		wantRemedKey string
		wantIssueHas string
	}{
		{
			name: "empty Secret detected",
			secrets: []corev1.Secret{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "empty-secret", Namespace: "default"},
					Type:       corev1.SecretTypeOpaque,
					Data:       map[string][]byte{},
					StringData: map[string]string{},
				},
			},
			wantCount:    1,
			wantSeverity: diagnosis.Low,
			wantRemedKey: "secret_empty",
			wantIssueHas: "no data",
		},
		{
			name: "secret with only StringData empty",
			secrets: []corev1.Secret{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "empty-string", Namespace: "default"},
					Type:       corev1.SecretTypeOpaque,
					Data:       map[string][]byte{},
					StringData: map[string]string{},
				},
			},
			wantCount:    1,
			wantSeverity: diagnosis.Low,
		},
		{
			name: "healthy Secret with data not reported",
			secrets: []corev1.Secret{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "healthy-secret", Namespace: "default"},
					Type:       corev1.SecretTypeOpaque,
					Data: map[string][]byte{
						"username": []byte("admin"),
						"password": []byte("secret123"),
					},
				},
			},
			wantCount: 0,
		},
		{
			name: "service account token Secret skipped",
			secrets: []corev1.Secret{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "sa-token", Namespace: "default"},
					Type:       "kubernetes.io/service-account-token",
					Data: map[string][]byte{
						"token": []byte("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."),
					},
				},
			},
			wantCount: 0,
		},
		{
			name: "helm release Secret skipped",
			secrets: []corev1.Secret{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "helm-release", Namespace: "kube-system"},
					Type:       "helm.sh/release.v1",
					Data: map[string][]byte{
						"release": []byte("H4sIAAAAAAAC..."),
					},
				},
			},
			wantCount: 0,
		},
		{
			name:      "no Secrets produces zero problems",
			secrets:   nil,
			wantCount: 0,
		},
		{
			name: "multiple empty Secrets all detected",
			secrets: []corev1.Secret{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "empty1", Namespace: "default"},
					Type:       corev1.SecretTypeOpaque,
					Data:       map[string][]byte{},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "empty2", Namespace: "production"},
					Type:       corev1.SecretTypeOpaque,
					Data:       map[string][]byte{},
				},
			},
			wantCount:    2,
			wantSeverity: diagnosis.Low,
		},
		{
			name: "mixed empty and healthy Secrets",
			secrets: []corev1.Secret{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "healthy", Namespace: "default"},
					Type:       corev1.SecretTypeOpaque,
					Data: map[string][]byte{
						"key": []byte("value"),
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "empty", Namespace: "default"},
					Type:       corev1.SecretTypeOpaque,
					Data:       map[string][]byte{},
				},
			},
			wantCount:    1,
			wantSeverity: diagnosis.Low,
		},
		{
			name: "Secret with StringData but no Data not reported",
			secrets: []corev1.Secret{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "stringdata-secret", Namespace: "default"},
					Type:       corev1.SecretTypeOpaque,
					Data:       map[string][]byte{},
					StringData: map[string]string{
						"username": "admin",
						"password": "secret456",
					},
				},
			},
			wantCount: 0,
		},
		{
			name: "TLS Secret with data not reported",
			secrets: []corev1.Secret{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "tls-secret", Namespace: "ingress"},
					Type:       corev1.SecretTypeTLS,
					Data: map[string][]byte{
						"tls.crt": []byte("-----BEGIN CERTIFICATE-----\n..."),
						"tls.key": []byte("-----BEGIN PRIVATE KEY-----\n..."),
					},
				},
			},
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			client := fake.NewSimpleClientset(secretsToObjects(tt.secrets)...)
			s := NewSecretScanner()

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

func secretsToObjects(secrets []corev1.Secret) []runtime.Object {
	objs := make([]runtime.Object, len(secrets))
	for i := range secrets {
		objs[i] = &secrets[i]
	}
	return objs
}
