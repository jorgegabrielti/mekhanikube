package types

import "testing"

// TestProblemIntegration testa a integração completa do Problem com Severity e Score
func TestProblemIntegration(t *testing.T) {
	// Simula criação de problema como o scanner faz
	problem := Problem{
		Kind:      "Pod",
		Namespace: "kube-system",
		Name:      "coredns-abc123",
		Error:     "Container in CrashLoopBackOff",
		Details:   []string{"Exit code: 1", "Last restart: 5m ago"},
	}

	// Define severidade
	problem.Severity = Critical

	// Calcula score
	problem.CalculateScore()

	// Validações
	if problem.Severity != Critical {
		t.Errorf("Expected severity %s, got %s", Critical, problem.Severity)
	}

	// Score esperado: 90 (Critical) + 10 (kube-system) + 10 (CrashLoopBackOff) = 100 (capped)
	expectedScore := 100
	if problem.Score != expectedScore {
		t.Errorf("Expected score %d, got %d", expectedScore, problem.Score)
	}

	// Testa serialização JSON
	str := problem.String()
	if str == "" {
		t.Error("Problem string should not be empty")
	}
}

// TestProblemWorkflow testa o workflow completo
func TestProblemWorkflow(t *testing.T) {
	scenarios := []struct {
		name          string
		problem       Problem
		severity      Severity
		expectedScore int
	}{
		{
			name: "Pod normal em namespace comum",
			problem: Problem{
				Kind:      "Pod",
				Namespace: "production",
				Name:      "api-server",
				Error:     "Normal warning",
			},
			severity:      Low,
			expectedScore: 30,
		},
		{
			name: "Service crítico sem endpoints",
			problem: Problem{
				Kind:      "Service",
				Namespace: "default",
				Name:      "api-gateway",
				Error:     "Service has no endpoints available",
			},
			severity:      High,
			expectedScore: 90, // 70 + 10 (default) + 10 (no endpoints)
		},
		{
			name: "Pod com OOMKilled",
			problem: Problem{
				Kind:      "Pod",
				Namespace: "production",
				Name:      "memory-app",
				Error:     "Container was OOMKilled",
			},
			severity:      Critical,
			expectedScore: 100, // 90 + 10 (OOMKilled) = 100
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			// Simula workflow: criar problema → definir severidade → calcular score
			p := scenario.problem
			p.Severity = scenario.severity
			p.CalculateScore()

			if p.Score != scenario.expectedScore {
				t.Errorf("Expected score %d, got %d for scenario '%s'",
					scenario.expectedScore, p.Score, scenario.name)
			}

			// Valida que todos os campos estão populados
			if p.Kind == "" || p.Name == "" || p.Error == "" {
				t.Error("Problem should have all required fields populated")
			}

			if p.Severity == "" {
				t.Error("Severity should be set")
			}

			if p.Score < 0 || p.Score > 100 {
				t.Errorf("Score should be in range 0-100, got %d", p.Score)
			}
		})
	}
}
