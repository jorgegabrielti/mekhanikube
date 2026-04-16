package scanner

import (
	"testing"

	"github.com/jorgegabrielti/nautikube/internal/diagnosis"
)

func TestDeduplicateProblems(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		problems  []diagnosis.Problem
		wantCount int
		wantIssue string // if set, verify first remaining problem contains this
	}{
		{
			name: "CrashLoopBackOff suppresses high-restarts for same pod",
			problems: []diagnosis.Problem{
				{Resource: "Pod", Namespace: "default", Name: "app-abc-123", Issue: "Container main in CrashLoopBackOff", Severity: diagnosis.Critical},
				{Resource: "Pod", Namespace: "default", Name: "app-abc-123", Issue: "Container main has restarted 42 times", Severity: diagnosis.High},
			},
			wantCount: 1,
			wantIssue: "CrashLoopBackOff",
		},
		{
			name: "CrashLoopBackOff OOMKilled suppresses standalone OOMKilled for same pod",
			problems: []diagnosis.Problem{
				{Resource: "Pod", Namespace: "default", Name: "app-abc-123", Issue: "Container main in CrashLoopBackOff (OOMKilled)", Severity: diagnosis.Critical},
				{Resource: "Pod", Namespace: "default", Name: "app-abc-123", Issue: "Container main was OOMKilled", Severity: diagnosis.Critical},
				{Resource: "Pod", Namespace: "default", Name: "app-abc-123", Issue: "Container main has restarted 55 times", Severity: diagnosis.Critical},
			},
			wantCount: 1,
			wantIssue: "CrashLoopBackOff",
		},
		{
			name: "different pods are not deduplicated",
			problems: []diagnosis.Problem{
				{Resource: "Pod", Namespace: "default", Name: "app-1", Issue: "Container main in CrashLoopBackOff", Severity: diagnosis.Critical},
				{Resource: "Pod", Namespace: "default", Name: "app-2", Issue: "Container main has restarted 42 times", Severity: diagnosis.High},
			},
			wantCount: 2,
		},
		{
			name: "non-Pod resources are never deduplicated",
			problems: []diagnosis.Problem{
				{Resource: "Deployment", Namespace: "default", Name: "app", Issue: "unavailable replicas", Severity: diagnosis.High},
				{Resource: "Service", Namespace: "default", Name: "app-svc", Issue: "no endpoints", Severity: diagnosis.High},
			},
			wantCount: 2,
		},
		{
			name: "high-restarts without CrashLoopBackOff is kept",
			problems: []diagnosis.Problem{
				{Resource: "Pod", Namespace: "default", Name: "app-abc-123", Issue: "Container main has restarted 42 times", Severity: diagnosis.High},
			},
			wantCount: 1,
		},
		{
			name:      "empty input returns empty",
			problems:  nil,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := deduplicateProblems(tt.problems)
			if len(got) != tt.wantCount {
				issues := make([]string, len(got))
				for i, p := range got {
					issues[i] = p.Issue
				}
				t.Fatalf("deduplicateProblems() returned %d problems %v, want %d", len(got), issues, tt.wantCount)
			}
			if tt.wantIssue != "" && len(got) > 0 {
				found := false
				for _, p := range got {
					if containsCI(p.Issue, tt.wantIssue) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected remaining problem to contain %q", tt.wantIssue)
				}
			}
		})
	}
}

func containsCI(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}
