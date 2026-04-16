package scanner

import (
	"context"
	"fmt"
	"sort"
	"strings"

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

func (r *Registry) ScanAll(ctx context.Context, client kubernetes.Interface, namespace string, kb *diagnosis.KnowledgeBase, resourceFilter []string) ([]diagnosis.Problem, error) {
	// Validate resource filter
	if len(resourceFilter) > 0 {
		found := false
		for _, rf := range resourceFilter {
			for _, s := range r.scanners {
				if s.Name() == rf {
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("no scanners match resource filter: %v", resourceFilter)
		}
	}

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

	allProblems = deduplicateProblems(allProblems)

	// Sort by score descending (highest priority first)
	sort.Slice(allProblems, func(i, j int) bool {
		return allProblems[i].Score > allProblems[j].Score
	})

	return allProblems, nil
}

// deduplicateProblems removes redundant problems for the same resource.
// Rules:
//   - If a Pod has CrashLoopBackOff, suppress "high-restarts" for the same pod.
//   - If a Pod has CrashLoopBackOff (OOMKilled), suppress standalone "OOMKilled" for the same pod.
func deduplicateProblems(problems []diagnosis.Problem) []diagnosis.Problem {
	// Build a set of (namespace/name) → issue-type for Pod problems
	type podKey struct {
		namespace, name string
	}
	podIssues := make(map[podKey]map[string]bool)
	for _, p := range problems {
		if p.Resource != "Pod" {
			continue
		}
		key := podKey{p.Namespace, p.Name}
		if podIssues[key] == nil {
			podIssues[key] = make(map[string]bool)
		}
		lower := strings.ToLower(p.Issue)
		if strings.Contains(lower, "crashloopbackoff") {
			podIssues[key]["crashloop"] = true
			if strings.Contains(lower, "oomkilled") {
				podIssues[key]["oomkilled"] = true
			}
		}
	}

	var result []diagnosis.Problem
	for _, p := range problems {
		if p.Resource != "Pod" {
			result = append(result, p)
			continue
		}

		key := podKey{p.Namespace, p.Name}
		issues := podIssues[key]
		lower := strings.ToLower(p.Issue)

		// Suppress high-restarts if CrashLoopBackOff already reported for this pod
		if issues["crashloop"] && strings.Contains(lower, "restarted") {
			continue
		}

		// Suppress standalone OOMKilled if CrashLoopBackOff (OOMKilled) already covers it
		if issues["oomkilled"] && strings.Contains(lower, "oomkilled") && !strings.Contains(lower, "crashloopbackoff") {
			continue
		}

		result = append(result, p)
	}

	return result
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
