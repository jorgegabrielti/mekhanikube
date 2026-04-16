package scanner

import (
	"context"
	"testing"

	"github.com/jorgegabrielti/nautikube/internal/diagnosis"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

func TestIngressScanner_Name(t *testing.T) {
	t.Parallel()
	s := NewIngressScanner()
	if got := s.Name(); got != "Ingress" {
		t.Errorf("Name() = %q, want %q", got, "Ingress")
	}
}

func TestIngressScanner_Scan(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		ingresses    []networkingv1.Ingress
		wantCount    int
		wantSeverity diagnosis.Severity
		wantRemedKey string
		wantIssueHas string
	}{
		{
			name: "Ingress with no rules and no default backend detected",
			ingresses: []networkingv1.Ingress{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "empty-ingress", Namespace: "default"},
					Spec:       networkingv1.IngressSpec{},
				},
			},
			wantCount:    1,
			wantSeverity: diagnosis.Medium,
			wantRemedKey: "ingress_empty",
			wantIssueHas: "no rules",
		},
		{
			name: "Ingress with rules not reported",
			ingresses: []networkingv1.Ingress{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "healthy-ingress", Namespace: "default"},
					Spec: networkingv1.IngressSpec{
						Rules: []networkingv1.IngressRule{
							{
								Host: "example.com",
							},
						},
					},
				},
			},
			wantCount: 0,
		},
		{
			name: "Ingress with default backend not reported",
			ingresses: []networkingv1.Ingress{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "default-backend-ingress", Namespace: "default"},
					Spec: networkingv1.IngressSpec{
						DefaultBackend: &networkingv1.IngressBackend{
							Service: &networkingv1.IngressServiceBackend{
								Name: "backend-svc",
								Port: networkingv1.ServiceBackendPort{Number: 80},
							},
						},
					},
				},
			},
			wantCount: 0,
		},
		{
			name:      "no Ingresses produces zero problems",
			ingresses: nil,
			wantCount: 0,
		},
		{
			name: "multiple empty Ingresses all detected",
			ingresses: []networkingv1.Ingress{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "empty-ingress-1", Namespace: "default"},
					Spec:       networkingv1.IngressSpec{},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "empty-ingress-2", Namespace: "production"},
					Spec:       networkingv1.IngressSpec{},
				},
			},
			wantCount:    2,
			wantSeverity: diagnosis.Medium,
		},
		{
			name: "mixed Ingress statuses",
			ingresses: []networkingv1.Ingress{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "healthy-ingress", Namespace: "default"},
					Spec: networkingv1.IngressSpec{
						Rules: []networkingv1.IngressRule{
							{
								Host: "api.example.com",
							},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "empty-ingress", Namespace: "default"},
					Spec:       networkingv1.IngressSpec{},
				},
			},
			wantCount:    1,
			wantSeverity: diagnosis.Medium,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			client := fake.NewSimpleClientset(ingressesToObjects(tt.ingresses)...)
			s := NewIngressScanner()

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

func ingressesToObjects(ingresses []networkingv1.Ingress) []runtime.Object {
	objs := make([]runtime.Object, len(ingresses))
	for i := range ingresses {
		objs[i] = &ingresses[i]
	}
	return objs
}
