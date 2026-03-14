package diagnosis

import "testing"

func TestSeverityWeight(t *testing.T) {
	t.Parallel()

	tests := []struct {
		severity Severity
		want     int
	}{
		{Critical, 5},
		{High, 4},
		{Medium, 3},
		{Low, 2},
		{Info, 1},
		{Severity("UNKNOWN"), 0},
	}

	for _, tt := range tests {
		t.Run(string(tt.severity), func(t *testing.T) {
			t.Parallel()
			if got := tt.severity.Weight(); got != tt.want {
				t.Errorf("Weight() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestProblemCalculateScore(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		problem   Problem
		wantScore int
	}{
		{
			name: "critical base score",
			problem: Problem{
				Resource:  "Pod",
				Namespace: "production",
				Name:      "app",
				Issue:     "some error",
				Severity:  Critical,
			},
			wantScore: 90,
		},
		{
			name: "critical in kube-system gets +10",
			problem: Problem{
				Resource:  "Pod",
				Namespace: "kube-system",
				Name:      "coredns",
				Issue:     "some error",
				Severity:  Critical,
			},
			wantScore: 100,
		},
		{
			name: "pod CrashLoopBackOff gets +10",
			problem: Problem{
				Resource:  "Pod",
				Namespace: "production",
				Name:      "app",
				Issue:     "Container in CrashLoopBackOff",
				Severity:  Critical,
			},
			wantScore: 100,
		},
		{
			name: "pod in default with CrashLoopBackOff caps at 100",
			problem: Problem{
				Resource:  "Pod",
				Namespace: "default",
				Name:      "app",
				Issue:     "Container in CrashLoopBackOff",
				Severity:  Critical,
			},
			wantScore: 100, // 90 + 10 (default) + 10 (CrashLoopBackOff) = 110, capped
		},
		{
			name: "service no endpoints gets +10",
			problem: Problem{
				Resource:  "Service",
				Namespace: "production",
				Name:      "api",
				Issue:     "Service has no endpoints",
				Severity:  High,
			},
			wantScore: 80, // 70 + 10 (no endpoints)
		},
		{
			name: "high base score",
			problem: Problem{
				Resource:  "Deployment",
				Namespace: "production",
				Name:      "web",
				Issue:     "2 unavailable replicas",
				Severity:  High,
			},
			wantScore: 70,
		},
		{
			name: "medium base score",
			problem: Problem{
				Resource:  "Deployment",
				Namespace: "staging",
				Name:      "web",
				Issue:     "replicas mismatch",
				Severity:  Medium,
			},
			wantScore: 50,
		},
		{
			name: "low base score",
			problem: Problem{
				Resource:  "Node",
				Namespace: "",
				Name:      "node-1",
				Issue:     "unschedulable",
				Severity:  Low,
			},
			wantScore: 30,
		},
		{
			name: "info base score",
			problem: Problem{
				Resource:  "Pod",
				Namespace: "test",
				Name:      "debug",
				Issue:     "informational",
				Severity:  Info,
			},
			wantScore: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := tt.problem
			p.CalculateScore()
			if p.Score != tt.wantScore {
				t.Errorf("CalculateScore() = %d, want %d", p.Score, tt.wantScore)
			}
		})
	}
}

func TestCalculateScoreRange(t *testing.T) {
	t.Parallel()

	severities := []Severity{Critical, High, Medium, Low, Info}
	namespaces := []string{"default", "kube-system", "production", "test"}
	issues := []string{"CrashLoopBackOff", "ImagePullBackOff", "OOMKilled", "no endpoints", "Normal error"}

	for _, sev := range severities {
		for _, ns := range namespaces {
			for _, issue := range issues {
				p := Problem{
					Resource:  "Pod",
					Name:      "test",
					Namespace: ns,
					Issue:     issue,
					Severity:  sev,
				}
				p.CalculateScore()
				if p.Score < 0 || p.Score > 100 {
					t.Errorf("Score out of range [0-100]: %d (severity=%s, ns=%s, issue=%s)",
						p.Score, sev, ns, issue)
				}
			}
		}
	}
}

func TestProblemString(t *testing.T) {
	t.Parallel()
	p := Problem{
		Resource:  "Pod",
		Namespace: "default",
		Name:      "test-pod",
		Issue:     "CrashLoopBackOff",
		Severity:  Critical,
		Score:     90,
	}
	s := p.String()
	if s == "" {
		t.Error("String() should not be empty")
	}
}

func TestNewSummary(t *testing.T) {
	t.Parallel()
	problems := []Problem{
		{Severity: Critical},
		{Severity: Critical},
		{Severity: High},
		{Severity: Medium},
		{Severity: Low},
		{Severity: Info},
	}
	s := NewSummary(problems)
	if s.Total != 6 {
		t.Errorf("Total = %d, want 6", s.Total)
	}
	if s.Critical != 2 {
		t.Errorf("Critical = %d, want 2", s.Critical)
	}
	if s.High != 1 {
		t.Errorf("High = %d, want 1", s.High)
	}
	if s.Medium != 1 {
		t.Errorf("Medium = %d, want 1", s.Medium)
	}
	if s.Low != 1 {
		t.Errorf("Low = %d, want 1", s.Low)
	}
	if s.Info != 1 {
		t.Errorf("Info = %d, want 1", s.Info)
	}
}

func TestNewSummaryEmpty(t *testing.T) {
	t.Parallel()
	s := NewSummary(nil)
	if s.Total != 0 {
		t.Errorf("Total = %d, want 0", s.Total)
	}
}
