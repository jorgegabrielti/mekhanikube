package analyzer

import (
	"context"
	"fmt"

	"github.com/jorgegabrielti/nautikube/internal/ollama"
	"github.com/jorgegabrielti/nautikube/internal/scanner"
	"github.com/jorgegabrielti/nautikube/pkg/types"
)

// Analyzer coordena o scanning e análise de problemas
type Analyzer struct {
	scanner      *scanner.Scanner
	ollamaClient *ollama.Client
}

// New cria um novo Analyzer
func New(scanner *scanner.Scanner, ollamaClient *ollama.Client) *Analyzer {
	return &Analyzer{
		scanner:      scanner,
		ollamaClient: ollamaClient,
	}
}

// Analyze executa a análise completa do cluster
func (a *Analyzer) Analyze(ctx context.Context, opts types.AnalyzeOptions) ([]types.Problem, error) {
	var allProblems []types.Problem

	// Define quais recursos escanear baseado nos filtros
	shouldScanPods := len(opts.Filter) == 0 || contains(opts.Filter, "Pod")
	shouldScanConfigMaps := len(opts.Filter) == 0 || contains(opts.Filter, "ConfigMap")

	// Escaneia Pods
	if shouldScanPods {
		problems, err := a.scanner.ScanPods(ctx, opts.Namespace)
		if err != nil {
			return nil, fmt.Errorf("erro ao escanear pods: %w", err)
		}
		allProblems = append(allProblems, problems...)
	}

	// Escaneia ConfigMaps
	if shouldScanConfigMaps {
		problems, err := a.scanner.ScanConfigMaps(ctx, opts.Namespace)
		if err != nil {
			return nil, fmt.Errorf("erro ao escanear configmaps: %w", err)
		}
		allProblems = append(allProblems, problems...)
	}

	// Define severidade e calcula score para cada problema
	for i := range allProblems {
		a.assignSeverity(&allProblems[i])
		allProblems[i].CalculateScore()
	}

	// Se deve explicar com IA, processa cada problema
	if opts.Explain && a.ollamaClient != nil {
		for i := range allProblems {
			explanation, err := a.ollamaClient.Explain(ctx, &allProblems[i], opts.Language)
			if err != nil {
				// Continua mesmo se falhar em um problema
				allProblems[i].Explanation = fmt.Sprintf("Erro ao obter explicação: %v", err)
			} else {
				allProblems[i].Explanation = explanation
			}
		}
	}

	return allProblems, nil
}

// assignSeverity define a severidade baseada no tipo de problema
func (a *Analyzer) assignSeverity(p *types.Problem) {
	errorLower := toLower(p.Error)

	// Critical: Problemas que afetam diretamente a disponibilidade
	if containsAny(errorLower, []string{"crashloopbackoff", "oomkilled", "error", "failed"}) {
		p.Severity = types.Critical
		return
	}

	// High: Problemas que podem afetar funcionalidade
	if containsAny(errorLower, []string{"imagepullbackoff", "pending", "no endpoints"}) {
		p.Severity = types.High
		return
	}

	// Medium: Avisos importantes
	if containsAny(errorLower, []string{"warning", "restart"}) {
		p.Severity = types.Medium
		return
	}

	// Low: Outros problemas menores
	p.Severity = types.Low
}

// containsAny verifica se a string contém alguma das substrings
func containsAny(s string, substrs []string) bool {
	for _, substr := range substrs {
		if indexCaseInsensitive(s, substr) >= 0 {
			return true
		}
	}
	return false
}

// indexCaseInsensitive encontra substr em s ignorando case
func indexCaseInsensitive(s, substr string) int {
	s = toLower(s)
	substr = toLower(substr)
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// toLower converte string para lowercase
func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c = c + ('a' - 'A')
		}
		result[i] = c
	}
	return string(result)
}

// contains verifica se uma string está em um slice
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
