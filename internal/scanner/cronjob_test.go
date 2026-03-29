package scanner

import (
	"context"
	"testing"

	"github.com/jorgegabrielti/nautikube/internal/diagnosis"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

func TestCronJobScanner_Name(t *testing.T) {
	t.Parallel()
	s := NewCronJobScanner()
	if got := s.Name(); got != "CronJob" {
		t.Errorf("Name() = %q, want %q", got, "CronJob")
	}
}

func TestCronJobScanner_Scan(t *testing.T) {
	t.Parallel()

	trueVal := true
	falseVal := false

	tests := []struct {
		name         string
		cronjobs     []batchv1.CronJob
		wantCount    int
		wantSeverity diagnosis.Severity
		wantRemedKey string
		wantIssueHas string
	}{
		{
			name: "suspended CronJob detected",
			cronjobs: []batchv1.CronJob{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "suspended-cron", Namespace: "default"},
					Spec: batchv1.CronJobSpec{
						Suspend: &trueVal,
					},
				},
			},
			wantCount:    1,
			wantSeverity: diagnosis.Low,
			wantRemedKey: "cronjob_suspended",
			wantIssueHas: "suspended",
		},
		{
			name: "active CronJob not reported",
			cronjobs: []batchv1.CronJob{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "active-cron", Namespace: "default"},
					Spec: batchv1.CronJobSpec{
						Suspend: &falseVal,
					},
				},
			},
			wantCount: 0,
		},
		{
			name: "CronJob with no Suspend field not reported",
			cronjobs: []batchv1.CronJob{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "default-cron", Namespace: "default"},
					Spec:       batchv1.CronJobSpec{
						// Suspend not set (defaults to false)
					},
				},
			},
			wantCount: 0,
		},
		{
			name:      "no CronJobs produces zero problems",
			cronjobs:  nil,
			wantCount: 0,
		},
		{
			name: "multiple suspended CronJobs all detected",
			cronjobs: []batchv1.CronJob{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "suspended-cron-1", Namespace: "default"},
					Spec: batchv1.CronJobSpec{
						Suspend: &trueVal,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "suspended-cron-2", Namespace: "kube-system"},
					Spec: batchv1.CronJobSpec{
						Suspend: &trueVal,
					},
				},
			},
			wantCount:    2,
			wantSeverity: diagnosis.Low,
		},
		{
			name: "mixed CronJob statuses",
			cronjobs: []batchv1.CronJob{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "active-cron", Namespace: "default"},
					Spec: batchv1.CronJobSpec{
						Suspend: &falseVal,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "suspended-cron", Namespace: "default"},
					Spec: batchv1.CronJobSpec{
						Suspend: &trueVal,
					},
				},
			},
			wantCount:    1,
			wantSeverity: diagnosis.Low,
		},
		{
			name: "suspended CronJob has Low severity",
			cronjobs: []batchv1.CronJob{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "suspended-cron", Namespace: "production"},
					Spec: batchv1.CronJobSpec{
						Suspend: &trueVal,
					},
				},
			},
			wantCount:    1,
			wantSeverity: diagnosis.Low,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			client := fake.NewSimpleClientset(cronjobsToObjects(tt.cronjobs)...)
			s := NewCronJobScanner()

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

func cronjobsToObjects(cronjobs []batchv1.CronJob) []runtime.Object {
	objs := make([]runtime.Object, len(cronjobs))
	for i := range cronjobs {
		objs[i] = &cronjobs[i]
	}
	return objs
}
