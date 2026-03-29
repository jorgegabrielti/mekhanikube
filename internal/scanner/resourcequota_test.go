package scanner

import (
	"context"
	"testing"

	"github.com/jorgegabrielti/nautikube/internal/diagnosis"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

func TestResourceQuotaScanner_Name(t *testing.T) {
	t.Parallel()
	s := NewResourceQuotaScanner()
	if got := s.Name(); got != "ResourceQuota" {
		t.Errorf("Name() = %q, want %q", got, "ResourceQuota")
	}
}

func TestResourceQuotaScanner_Scan(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		quotas       []corev1.ResourceQuota
		wantCount    int
		wantSeverity diagnosis.Severity
		wantRemedKey string
		wantIssueHas string
	}{
		{
			name: "ResourceQuota exhausted for pods",
			quotas: []corev1.ResourceQuota{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "exhausted-quota", Namespace: "default"},
					Spec: corev1.ResourceQuotaSpec{
						Hard: corev1.ResourceList{
							corev1.ResourcePods: resource.MustParse("10"),
						},
					},
					Status: corev1.ResourceQuotaStatus{
						Hard: corev1.ResourceList{
							corev1.ResourcePods: resource.MustParse("10"),
						},
						Used: corev1.ResourceList{
							corev1.ResourcePods: resource.MustParse("10"),
						},
					},
				},
			},
			wantCount:    1,
			wantSeverity: diagnosis.High,
			wantRemedKey: "quota_reached",
			wantIssueHas: "reached",
		},
		{
			name: "ResourceQuota with available capacity not reported",
			quotas: []corev1.ResourceQuota{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "healthy-quota", Namespace: "default"},
					Spec: corev1.ResourceQuotaSpec{
						Hard: corev1.ResourceList{
							corev1.ResourcePods: resource.MustParse("10"),
						},
					},
					Status: corev1.ResourceQuotaStatus{
						Hard: corev1.ResourceList{
							corev1.ResourcePods: resource.MustParse("10"),
						},
						Used: corev1.ResourceList{
							corev1.ResourcePods: resource.MustParse("2"),
						},
					},
				},
			},
			wantCount: 0,
		},
		{
			name:      "no ResourceQuotas produces zero problems",
			quotas:    nil,
			wantCount: 0,
		},
		{
			name: "multiple exhausted ResourceQuotas all detected",
			quotas: []corev1.ResourceQuota{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "exhausted-quota-1", Namespace: "default"},
					Spec: corev1.ResourceQuotaSpec{
						Hard: corev1.ResourceList{
							corev1.ResourcePods: resource.MustParse("5"),
						},
					},
					Status: corev1.ResourceQuotaStatus{
						Hard: corev1.ResourceList{
							corev1.ResourcePods: resource.MustParse("5"),
						},
						Used: corev1.ResourceList{
							corev1.ResourcePods: resource.MustParse("5"),
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "exhausted-quota-2", Namespace: "production"},
					Spec: corev1.ResourceQuotaSpec{
						Hard: corev1.ResourceList{
							corev1.ResourcePods: resource.MustParse("20"),
						},
					},
					Status: corev1.ResourceQuotaStatus{
						Hard: corev1.ResourceList{
							corev1.ResourcePods: resource.MustParse("20"),
						},
						Used: corev1.ResourceList{
							corev1.ResourcePods: resource.MustParse("20"),
						},
					},
				},
			},
			wantCount:    2,
			wantSeverity: diagnosis.High,
		},
		{
			name: "mixed ResourceQuota states",
			quotas: []corev1.ResourceQuota{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "healthy-quota", Namespace: "default"},
					Spec: corev1.ResourceQuotaSpec{
						Hard: corev1.ResourceList{
							corev1.ResourcePods: resource.MustParse("10"),
						},
					},
					Status: corev1.ResourceQuotaStatus{
						Hard: corev1.ResourceList{
							corev1.ResourcePods: resource.MustParse("10"),
						},
						Used: corev1.ResourceList{
							corev1.ResourcePods: resource.MustParse("5"),
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "exhausted-quota", Namespace: "default"},
					Spec: corev1.ResourceQuotaSpec{
						Hard: corev1.ResourceList{
							corev1.ResourcePods: resource.MustParse("10"),
						},
					},
					Status: corev1.ResourceQuotaStatus{
						Hard: corev1.ResourceList{
							corev1.ResourcePods: resource.MustParse("10"),
						},
						Used: corev1.ResourceList{
							corev1.ResourcePods: resource.MustParse("10"),
						},
					},
				},
			},
			wantCount:    1,
			wantSeverity: diagnosis.High,
		},
		{
			name: "ResourceQuota with low capacity still available",
			quotas: []corev1.ResourceQuota{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "low-capacity-quota", Namespace: "default"},
					Spec: corev1.ResourceQuotaSpec{
						Hard: corev1.ResourceList{
							corev1.ResourcePods: resource.MustParse("10"),
						},
					},
					Status: corev1.ResourceQuotaStatus{
						Hard: corev1.ResourceList{
							corev1.ResourcePods: resource.MustParse("10"),
						},
						Used: corev1.ResourceList{
							corev1.ResourcePods: resource.MustParse("9"),
						},
					},
				},
			},
			wantCount: 0,
		},
		{
			name: "ResourceQuota with CPU exhaustion detected",
			quotas: []corev1.ResourceQuota{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "cpu-quota", Namespace: "default"},
					Spec: corev1.ResourceQuotaSpec{
						Hard: corev1.ResourceList{
							corev1.ResourceCPU: resource.MustParse("100m"),
						},
					},
					Status: corev1.ResourceQuotaStatus{
						Hard: corev1.ResourceList{
							corev1.ResourceCPU: resource.MustParse("100m"),
						},
						Used: corev1.ResourceList{
							corev1.ResourceCPU: resource.MustParse("100m"),
						},
					},
				},
			},
			wantCount:    1,
			wantSeverity: diagnosis.High,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			client := fake.NewSimpleClientset(quotasToObjects(tt.quotas)...)
			s := NewResourceQuotaScanner()

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

func quotasToObjects(quotas []corev1.ResourceQuota) []runtime.Object {
	objs := make([]runtime.Object, len(quotas))
	for i := range quotas {
		objs[i] = &quotas[i]
	}
	return objs
}
