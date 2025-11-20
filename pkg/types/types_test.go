package types

import (
	"testing"
)

// TestProblemString testa a representação string de um problema
func TestProblemString(t *testing.T) {
	problem := Problem{
		Kind:      "Pod",
		Name:      "test-pod",
		Namespace: "default",
		Error:     "CrashLoopBackOff",
	}

	str := problem.String()
	if str == "" {
		t.Error("Problem string should not be empty")
	}

	// Verifica se contém informações essenciais
	if !contains(str, "Pod") {
		t.Error("Problem string should contain kind")
	}
	if !contains(str, "test-pod") {
		t.Error("Problem string should contain name")
	}
	if !contains(str, "default") {
		t.Error("Problem string should contain namespace")
	}
}

// TestAnalyzeOptions valida opções de análise
func TestAnalyzeOptions(t *testing.T) {
	tests := []struct {
		name    string
		options AnalyzeOptions
		valid   bool
	}{
		{
			name: "Opções válidas com explain",
			options: AnalyzeOptions{
				Namespace: "default",
				Filter:    []string{"Pod"},
				Explain:   true,
				Language:  "Portuguese",
			},
			valid: true,
		},
		{
			name: "Opções válidas sem explain",
			options: AnalyzeOptions{
				Namespace: "",
				Filter:    []string{},
				Explain:   false,
				Language:  "English",
			},
			valid: true,
		},
		{
			name: "Múltiplos filtros",
			options: AnalyzeOptions{
				Filter: []string{"Pod", "ConfigMap", "Service"},
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validações básicas
			if tt.options.Explain && tt.options.Language == "" {
				t.Error("Language should be set when Explain is true")
			}
		})
	}
}

// TestOllamaRequest valida estrutura de requisição Ollama
func TestOllamaRequest(t *testing.T) {
	req := OllamaRequest{
		Model:  "llama3.1:8b",
		Prompt: "Test prompt",
		Stream: false,
	}

	if req.Model == "" {
		t.Error("Model should not be empty")
	}
	if req.Prompt == "" {
		t.Error("Prompt should not be empty")
	}
	if req.Stream {
		t.Error("Stream should be false by default")
	}
}

// TestOllamaResponse valida estrutura de resposta Ollama
func TestOllamaResponse(t *testing.T) {
	resp := OllamaResponse{
		Response: "Test explanation",
		Model:    "llama3.1:8b",
		Done:     true,
	}

	if resp.Response == "" {
		t.Error("Response should not be empty")
	}
	if resp.Model == "" {
		t.Error("Model should not be empty")
	}
	if !resp.Done {
		t.Error("Done should be true")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr ||
		len(s) > len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// BenchmarkProblemString benchmarks Problem.String()
func BenchmarkProblemString(b *testing.B) {
	problem := Problem{
		Kind:      "Pod",
		Name:      "test-pod",
		Namespace: "default",
		Error:     "CrashLoopBackOff",
		Details:   []string{"Container failed", "Exit code 1"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = problem.String()
	}
}

// TestSeverityEnum testa os valores do enum Severity
func TestSeverityEnum(t *testing.T) {
	tests := []struct {
		severity Severity
		expected string
	}{
		{Critical, "CRITICAL"},
		{High, "HIGH"},
		{Medium, "MEDIUM"},
		{Low, "LOW"},
		{Info, "INFO"},
	}

	for _, tt := range tests {
		t.Run(string(tt.severity), func(t *testing.T) {
			if string(tt.severity) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, string(tt.severity))
			}
		})
	}
}

// TestCalculateScore testa o cálculo de score
func TestCalculateScore(t *testing.T) {
	tests := []struct {
		name          string
		problem       Problem
		expectedScore int
	}{
		{
			name: "Critical sem contexto",
			problem: Problem{
				Kind:      "Pod",
				Name:      "test-pod",
				Namespace: "production",
				Error:     "Error",
				Severity:  Critical,
			},
			expectedScore: 90,
		},
		{
			name: "Critical em kube-system",
			problem: Problem{
				Kind:      "Pod",
				Name:      "kube-dns",
				Namespace: "kube-system",
				Error:     "Error",
				Severity:  Critical,
			},
			expectedScore: 100,
		},
		{
			name: "High com CrashLoopBackOff",
			problem: Problem{
				Kind:      "Pod",
				Name:      "app-pod",
				Namespace: "default",
				Error:     "CrashLoopBackOff",
				Severity:  High,
			},
			expectedScore: 90, // 70 + 10 (default) + 10 (CrashLoopBackOff)
		},
		{
			name: "Medium sem contexto",
			problem: Problem{
				Kind:      "ConfigMap",
				Name:      "config",
				Namespace: "app",
				Error:     "Not found",
				Severity:  Medium,
			},
			expectedScore: 50,
		},
		{
			name: "Low sem contexto",
			problem: Problem{
				Kind:      "Pod",
				Name:      "test",
				Namespace: "test",
				Error:     "Warning",
				Severity:  Low,
			},
			expectedScore: 30,
		},
		{
			name: "Info sem contexto",
			problem: Problem{
				Kind:      "Service",
				Name:      "svc",
				Namespace: "app",
				Error:     "Info message",
				Severity:  Info,
			},
			expectedScore: 10,
		},
		{
			name: "Service sem endpoints",
			problem: Problem{
				Kind:      "Service",
				Name:      "api-svc",
				Namespace: "default",
				Error:     "Service has no endpoints available",
				Severity:  High,
			},
			expectedScore: 90, // 70 + 10 (default) + 10 (no endpoints)
		},
		{
			name: "Pod OOMKilled em kube-system",
			problem: Problem{
				Kind:      "Pod",
				Name:      "metrics-server",
				Namespace: "kube-system",
				Error:     "Container OOMKilled",
				Severity:  Critical,
			},
			expectedScore: 100, // 90 + 10 (kube-system) + 10 (OOMKilled) = 110, capped at 100
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.problem.CalculateScore()
			if tt.problem.Score != tt.expectedScore {
				t.Errorf("Expected score %d, got %d", tt.expectedScore, tt.problem.Score)
			}
		})
	}
}

// TestCalculateScoreRange testa que o score está sempre entre 0-100
func TestCalculateScoreRange(t *testing.T) {
	severities := []Severity{Critical, High, Medium, Low, Info}
	namespaces := []string{"default", "kube-system", "production", "test"}
	errors := []string{"CrashLoopBackOff", "ImagePullBackOff", "OOMKilled", "no endpoints", "Normal error"}

	for _, sev := range severities {
		for _, ns := range namespaces {
			for _, err := range errors {
				problem := Problem{
					Kind:      "Pod",
					Name:      "test",
					Namespace: ns,
					Error:     err,
					Severity:  sev,
				}
				problem.CalculateScore()

				if problem.Score < 0 || problem.Score > 100 {
					t.Errorf("Score out of range [0-100]: %d for severity=%s, namespace=%s, error=%s",
						problem.Score, sev, ns, err)
				}
			}
		}
	}
}
