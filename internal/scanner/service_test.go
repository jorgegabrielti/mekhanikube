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

func TestServiceScanner_Name(t *testing.T) {
	t.Parallel()
	s := NewServiceScanner()
	if got := s.Name(); got != "Service" {
		t.Errorf("Name() = %q, want %q", got, "Service")
	}
}

func TestServiceScanner_Scan(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		objects      []runtime.Object
		wantCount    int
		wantSeverity diagnosis.Severity
		wantRemedKey string
		wantIssueHas string
	}{
		{
			name: "no endpoints detected",
			objects: []runtime.Object{
				&corev1.Service{
					ObjectMeta: metav1.ObjectMeta{Name: "my-svc", Namespace: "default"},
					Spec: corev1.ServiceSpec{
						ClusterIP: "10.0.0.1",
						Selector:  map[string]string{"app": "my-app"},
					},
				},
			},
			wantCount:    1,
			wantSeverity: diagnosis.High,
			wantRemedKey: "service-no-endpoints",
			wantIssueHas: "no endpoints",
		},
		{
			name: "endpoints exist with zero addresses",
			objects: []runtime.Object{
				&corev1.Service{
					ObjectMeta: metav1.ObjectMeta{Name: "my-svc", Namespace: "default"},
					Spec: corev1.ServiceSpec{
						ClusterIP: "10.0.0.1",
						Selector:  map[string]string{"app": "my-app"},
					},
				},
				&corev1.Endpoints{
					ObjectMeta: metav1.ObjectMeta{Name: "my-svc", Namespace: "default"},
					Subsets:    []corev1.EndpointSubset{{Addresses: nil}},
				},
			},
			wantCount:    1,
			wantSeverity: diagnosis.High,
			wantRemedKey: "service-no-endpoints",
		},
		{
			name: "loadbalancer pending",
			objects: []runtime.Object{
				&corev1.Service{
					ObjectMeta: metav1.ObjectMeta{Name: "lb-svc", Namespace: "default"},
					Spec: corev1.ServiceSpec{
						Type:      corev1.ServiceTypeLoadBalancer,
						ClusterIP: "10.0.0.2",
						Selector:  map[string]string{"app": "web"},
					},
					Status: corev1.ServiceStatus{},
				},
				&corev1.Endpoints{
					ObjectMeta: metav1.ObjectMeta{Name: "lb-svc", Namespace: "default"},
					Subsets:    []corev1.EndpointSubset{{Addresses: []corev1.EndpointAddress{{IP: "10.1.0.1"}}}},
				},
			},
			wantCount:    1,
			wantSeverity: diagnosis.Medium,
			wantRemedKey: "service-lb-pending",
		},
		{
			name: "externalname without target",
			objects: []runtime.Object{
				&corev1.Service{
					ObjectMeta: metav1.ObjectMeta{Name: "ext-svc", Namespace: "default"},
					Spec: corev1.ServiceSpec{
						Type:         corev1.ServiceTypeExternalName,
						ExternalName: "",
					},
				},
			},
			wantCount:    1,
			wantSeverity: diagnosis.Low,
			wantRemedKey: "service-externalname-empty",
		},
		{
			name: "headless service skipped",
			objects: []runtime.Object{
				&corev1.Service{
					ObjectMeta: metav1.ObjectMeta{Name: "headless-svc", Namespace: "default"},
					Spec: corev1.ServiceSpec{
						ClusterIP: "None",
						Selector:  map[string]string{"app": "stateful"},
					},
				},
			},
			wantCount: 0,
		},
		{
			name: "service without selector skipped",
			objects: []runtime.Object{
				&corev1.Service{
					ObjectMeta: metav1.ObjectMeta{Name: "manual-svc", Namespace: "default"},
					Spec: corev1.ServiceSpec{
						ClusterIP: "10.0.0.3",
						Selector:  nil,
					},
				},
			},
			wantCount: 0,
		},
		{
			name: "kubernetes default service skipped",
			objects: []runtime.Object{
				&corev1.Service{
					ObjectMeta: metav1.ObjectMeta{Name: "kubernetes", Namespace: "default"},
					Spec: corev1.ServiceSpec{
						ClusterIP: "10.96.0.1",
						Selector:  map[string]string{"component": "apiserver"},
					},
				},
			},
			wantCount: 0,
		},
		{
			name: "healthy service with endpoints",
			objects: []runtime.Object{
				&corev1.Service{
					ObjectMeta: metav1.ObjectMeta{Name: "good-svc", Namespace: "default"},
					Spec: corev1.ServiceSpec{
						ClusterIP: "10.0.0.4",
						Selector:  map[string]string{"app": "good"},
					},
				},
				&corev1.Endpoints{
					ObjectMeta: metav1.ObjectMeta{Name: "good-svc", Namespace: "default"},
					Subsets: []corev1.EndpointSubset{
						{Addresses: []corev1.EndpointAddress{{IP: "10.1.0.5"}}},
					},
				},
			},
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			client := fake.NewSimpleClientset(tt.objects...)
			s := NewServiceScanner()

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
