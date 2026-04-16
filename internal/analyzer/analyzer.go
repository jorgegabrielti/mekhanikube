package analyzer

import (
	"context"
	"fmt"

	"github.com/jorgegabrielti/nautikube/pkg/types"
)

// ClusterScanner define a interface para escanear recursos do cluster
type ClusterScanner interface {
	ScanPods(ctx context.Context, namespace string) ([]types.Problem, error)
	ScanConfigMaps(ctx context.Context, namespace string) ([]types.Problem, error)
}

// Explainer define a interface para explicar problemas com IA
type Explainer interface {
	Explain(ctx context.Context, problem *types.Problem, language string) (string, error)
}

// Analyzer coordena o scanning e análise de problemas
type Analyzer struct {
	scanner   ClusterScanner
	explainer Explainer
}

// New cria um novo Analyzer
func New(scanner ClusterScanner, explainer Explainer) *Analyzer {
	return &Analyzer{
		scanner:   scanner,
		explainer: explainer,
	}
}

// Analyze executa a análise completa do cluster
func (a *Analyzer) Analyze(ctx context.Context, opts types.AnalyzeOptions) ([]types.Problem, error) {
	var allProblems []types.Problem

	// Define quais recursos escanear baseado nos filtros
	shouldScanPods := len(opts.Filter) == 0 || types.ContainsString(opts.Filter, "Pod")
	shouldScanConfigMaps := len(opts.Filter) == 0 || types.ContainsString(opts.Filter, "ConfigMap")

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
	if opts.Explain && a.explainer != nil {
		for i := range allProblems {
			explanation, err := a.explainer.Explain(ctx, &allProblems[i], opts.Language)
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
	errorLower := types.ToLower(p.Error)

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
		if types.IndexCaseInsensitive(s, substr) >= 0 {
			return true
		}
	}
	return false
}
