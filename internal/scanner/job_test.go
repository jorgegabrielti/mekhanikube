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

func TestJobScanner_Name(t *testing.T) {
	t.Parallel()
	s := NewJobScanner()
	if got := s.Name(); got != "Job" {
		t.Errorf("Name() = %q, want %q", got, "Job")
	}
}

func TestJobScanner_Scan(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		jobs         []batchv1.Job
		wantCount    int
		wantSeverity diagnosis.Severity
		wantRemedKey string
		wantIssueHas string
	}{
		{
			name: "Job with failed executions detected",
			jobs: []batchv1.Job{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "failed-job", Namespace: "default"},
					Status: batchv1.JobStatus{
						Failed: 3,
					},
				},
			},
			wantCount:    1,
			wantSeverity: diagnosis.High,
			wantRemedKey: "job_failed",
			wantIssueHas: "failed",
		},
		{
			name: "Job with successful executions not reported",
			jobs: []batchv1.Job{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "success-job", Namespace: "default"},
					Status: batchv1.JobStatus{
						Succeeded: 5,
						Failed:    0,
					},
				},
			},
			wantCount: 0,
		},
		{
			name: "Job with zero failures not reported",
			jobs: []batchv1.Job{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "healthy-job", Namespace: "default"},
					Status: batchv1.JobStatus{
						Failed: 0,
					},
				},
			},
			wantCount: 0,
		},
		{
			name:      "no Jobs produces zero problems",
			jobs:      nil,
			wantCount: 0,
		},
		{
			name: "multiple failed Jobs all detected",
			jobs: []batchv1.Job{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "failed-job-1", Namespace: "default"},
					Status: batchv1.JobStatus{
						Failed: 2,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "failed-job-2", Namespace: "production"},
					Status: batchv1.JobStatus{
						Failed: 5,
					},
				},
			},
			wantCount:    2,
			wantSeverity: diagnosis.High,
		},
		{
			name: "mixed Job statuses",
			jobs: []batchv1.Job{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "success-job", Namespace: "default"},
					Status: batchv1.JobStatus{
						Succeeded: 10,
						Failed:    0,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "failed-job", Namespace: "default"},
					Status: batchv1.JobStatus{
						Failed: 4,
					},
				},
			},
			wantCount:    1,
			wantSeverity: diagnosis.High,
		},
		{
			name: "Job with single failure detected",
			jobs: []batchv1.Job{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "single-failure", Namespace: "batch"},
					Status: batchv1.JobStatus{
						Failed: 1,
					},
				},
			},
			wantCount:    1,
			wantSeverity: diagnosis.High,
		},
		{
			name: "Job with many failures detected",
			jobs: []batchv1.Job{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "many-failures", Namespace: "default"},
					Status: batchv1.JobStatus{
						Failed: 100,
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
			client := fake.NewSimpleClientset(jobsToObjects(tt.jobs)...)
			s := NewJobScanner()

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

func jobsToObjects(jobs []batchv1.Job) []runtime.Object {
	objs := make([]runtime.Object, len(jobs))
	for i := range jobs {
		objs[i] = &jobs[i]
	}
	return objs
}
