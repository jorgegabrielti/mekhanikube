package scanner

import (
	"context"
	"testing"

	"github.com/jorgegabrielti/nautikube/internal/diagnosis"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

func TestHorizontalPodAutoscalerScanner_Name(t *testing.T) {
	t.Parallel()
	s := NewHorizontalPodAutoscalerScanner()
	if got := s.Name(); got != "HorizontalPodAutoscaler" {
		t.Errorf("Name() = %q, want %q", got, "HorizontalPodAutoscaler")
	}
}

func TestHorizontalPodAutoscalerScanner_Scan(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		hpas         []autoscalingv1.HorizontalPodAutoscaler
		wantCount    int
		wantSeverity diagnosis.Severity
		wantRemedKey string
		wantIssueHas string
	}{
		{
			name: "HPA at maximum replicas detected",
			hpas: []autoscalingv1.HorizontalPodAutoscaler{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "at-max-hpa", Namespace: "default"},
					Spec: autoscalingv1.HorizontalPodAutoscalerSpec{
						MaxReplicas: 10,
					},
					Status: autoscalingv1.HorizontalPodAutoscalerStatus{
						CurrentReplicas: 10,
					},
				},
			},
			wantCount:    1,
			wantSeverity: diagnosis.Medium,
			wantRemedKey: "hpa_at_max",
			wantIssueHas: "maximum",
		},
		{
			name: "HPA below maximum not reported",
			hpas: []autoscalingv1.HorizontalPodAutoscaler{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "healthy-hpa", Namespace: "default"},
					Spec: autoscalingv1.HorizontalPodAutoscalerSpec{
						MaxReplicas: 10,
					},
					Status: autoscalingv1.HorizontalPodAutoscalerStatus{
						CurrentReplicas: 5,
					},
				},
			},
			wantCount: 0,
		},
		{
			name:      "no HPAs produces zero problems",
			hpas:      nil,
			wantCount: 0,
		},
		{
			name: "multiple HPAs at max all detected",
			hpas: []autoscalingv1.HorizontalPodAutoscaler{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "at-max-hpa-1", Namespace: "default"},
					Spec: autoscalingv1.HorizontalPodAutoscalerSpec{
						MaxReplicas: 20,
					},
					Status: autoscalingv1.HorizontalPodAutoscalerStatus{
						CurrentReplicas: 20,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "at-max-hpa-2", Namespace: "production"},
					Spec: autoscalingv1.HorizontalPodAutoscalerSpec{
						MaxReplicas: 50,
					},
					Status: autoscalingv1.HorizontalPodAutoscalerStatus{
						CurrentReplicas: 50,
					},
				},
			},
			wantCount:    2,
			wantSeverity: diagnosis.Medium,
		},
		{
			name: "mixed HPA statuses",
			hpas: []autoscalingv1.HorizontalPodAutoscaler{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "healthy-hpa", Namespace: "default"},
					Spec: autoscalingv1.HorizontalPodAutoscalerSpec{
						MaxReplicas: 10,
					},
					Status: autoscalingv1.HorizontalPodAutoscalerStatus{
						CurrentReplicas: 3,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "at-max-hpa", Namespace: "default"},
					Spec: autoscalingv1.HorizontalPodAutoscalerSpec{
						MaxReplicas: 15,
					},
					Status: autoscalingv1.HorizontalPodAutoscalerStatus{
						CurrentReplicas: 15,
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
			client := fake.NewSimpleClientset(hpasToObjects(tt.hpas)...)
			s := NewHorizontalPodAutoscalerScanner()

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

func hpasToObjects(hpas []autoscalingv1.HorizontalPodAutoscaler) []runtime.Object {
	objs := make([]runtime.Object, len(hpas))
	for i := range hpas {
		objs[i] = &hpas[i]
	}
	return objs
}
