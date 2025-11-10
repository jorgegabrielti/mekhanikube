package analyzer

import (
	"context"
	"fmt"

	"github.com/jorgegabrielti/mekhanikube/internal/ollama"
	"github.com/jorgegabrielti/mekhanikube/internal/scanner"
	"github.com/jorgegabrielti/mekhanikube/pkg/types"
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

// contains verifica se uma string está em um slice
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
