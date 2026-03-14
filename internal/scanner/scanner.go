package scanner

import (
	"context"
	"sort"

	"github.com/jorgegabrielti/nautikube/internal/diagnosis"
	"k8s.io/client-go/kubernetes"
)

// Scanner defines the interface for resource scanners.
type Scanner interface {
	// Name returns the human-readable name of the resource this scanner inspects.
	Name() string

	// Scan inspects resources in the given namespace and returns detected problems.
	// An empty namespace scans all namespaces.
	Scan(ctx context.Context, client kubernetes.Interface, namespace string) ([]diagnosis.Problem, error)
}

// Registry holds a collection of scanners and runs them in sequence.
type Registry struct {
	scanners []Scanner
}

// NewRegistry creates a Registry with the given scanners.
func NewRegistry(scanners ...Scanner) *Registry {
	return &Registry{scanners: scanners}
}

// ScanAll runs all registered scanners and returns all detected problems sorted by score.
func (r *Registry) ScanAll(ctx context.Context, client kubernetes.Interface, namespace string, kb *diagnosis.KnowledgeBase, resourceFilter []string) ([]diagnosis.Problem, error) {
	var allProblems []diagnosis.Problem

	for _, s := range r.scanners {
		if len(resourceFilter) > 0 && !contains(resourceFilter, s.Name()) {
			continue
		}

		problems, err := s.Scan(ctx, client, namespace)
		if err != nil {
			// Log and continue — don't fail the whole scan for one resource type
			continue
		}

		// Calculate score and enrich with knowledge base
		for i := range problems {
			problems[i].CalculateScore()
			if kb != nil {
				kb.Enrich(&problems[i])
			}
		}

		allProblems = append(allProblems, problems...)
	}

	// Sort by score descending (highest priority first)
	sort.Slice(allProblems, func(i, j int) bool {
		return allProblems[i].Score > allProblems[j].Score
	})

	return allProblems, nil
}

// Names returns the names of all registered scanners.
func (r *Registry) Names() []string {
	names := make([]string, len(r.scanners))
	for i, s := range r.scanners {
		names[i] = s.Name()
	}
	return names
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
