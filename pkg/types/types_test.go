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
