---
name: scanner-development
description: How to create a new resource scanner for NautiKube
---

# Scanner Development Skill

This skill guides the creation of new resource scanners for NautiKube.

## Scanner Interface

Every scanner implements this interface (defined in `internal/scanner/scanner.go`):

```go
type Scanner interface {
    Name() string
    Scan(ctx context.Context, client kubernetes.Interface, namespace string) ([]diagnosis.Problem, error)
}
```

## Step-by-Step: Adding a New Scanner

### 1. Write the Spec

Create `.agents/specs/<resource>-scanner.spec.md` with:
- List of problems detected (what condition triggers each)
- Severity and base score for each problem
- Remediation commands
- Edge cases to handle
- Acceptance criteria

### 2. Create the Scanner File

Create `internal/scanner/<resource>.go`:

```go
package scanner

import (
    "context"
    "fmt"

    "github.com/jorgegabrielti/nautikube/internal/diagnosis"
    "k8s.io/client-go/kubernetes"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// <Resource>Scanner scans <Resource> objects for problems.
type <Resource>Scanner struct{}

// New<Resource>Scanner creates a new <Resource>Scanner.
func New<Resource>Scanner() *<Resource>Scanner {
    return &<Resource>Scanner{}
}

// Name returns the scanner name.
func (s *<Resource>Scanner) Name() string {
    return "<Resource>"
}

// Scan inspects <Resource> objects and returns detected problems.
func (s *<Resource>Scanner) Scan(ctx context.Context, client kubernetes.Interface, namespace string) ([]diagnosis.Problem, error) {
    // 1. List resources from the K8s API
    // 2. Iterate and check for known problem patterns
    // 3. For each problem, create a diagnosis.Problem with:
    //    - Resource, Namespace, Name, Issue
    //    - Severity (from diagnosis constants)
    //    - Remediation key (to look up in knowledge base)
    // 4. Return the list of problems
    return nil, nil
}
```

### 3. Create Tests

Create `internal/scanner/<resource>_test.go` using the fake K8s client:

```go
package scanner

import (
    "context"
    "testing"

    "k8s.io/client-go/kubernetes/fake"
    // Import the relevant K8s API types
)

func TestResourceScanner_ProblemName(t *testing.T) {
    t.Parallel()

    tests := []struct {
        name       string
        objects    []runtime.Object  // Fake K8s objects
        namespace  string
        wantCount  int
        wantSev    diagnosis.Severity
    }{
        {
            name:      "healthy resource",
            objects:   []runtime.Object{healthyResource()},
            namespace: "",
            wantCount: 0,
        },
        {
            name:      "problematic resource",
            objects:   []runtime.Object{brokenResource()},
            namespace: "",
            wantCount: 1,
            wantSev:   diagnosis.Critical,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            client := fake.NewSimpleClientset(tt.objects...)
            scanner := NewResourceScanner()

            problems, err := scanner.Scan(context.Background(), client, tt.namespace)
            if err != nil {
                t.Fatalf("unexpected error: %v", err)
            }
            if got := len(problems); got != tt.wantCount {
                t.Errorf("got %d problems, want %d", got, tt.wantCount)
            }
            if tt.wantCount > 0 && problems[0].Severity != tt.wantSev {
                t.Errorf("got severity %q, want %q", problems[0].Severity, tt.wantSev)
            }
        })
    }
}

// Helper functions to create fake K8s objects
func healthyResource() *v1.Resource { ... }
func brokenResource() *v1.Resource { ... }
```

### 4. Add Knowledge Base Entries

Create `internal/diagnosis/knowledge/<problem-key>.yaml`:

```yaml
key: "crashloopbackoff"
title: "Container in CrashLoopBackOff"
explanation: |
  The container is crashing repeatedly and Kubernetes keeps restarting it
  with an exponential backoff delay. This usually indicates an application
  error, misconfigured command, or missing configuration.
commands:
  - "kubectl logs {pod} -n {namespace} --previous"
  - "kubectl describe pod {pod} -n {namespace}"
  - "kubectl get events -n {namespace} --field-selector involvedObject.name={pod}"
```

### 5. Register the Scanner

In `internal/scanner/scanner.go`, add the scanner to the default registry:

```go
func DefaultRegistry() *Registry {
    return NewRegistry(
        NewPodScanner(),
        NewDeploymentScanner(),
        NewServiceScanner(),
        // Add your new scanner here
    )
}
```

### 6. Run Quality Gate

```bash
make check  # Must pass: fmt → vet → lint → vuln → test → build
```

## Checklist

- [ ] Spec written and approved
- [ ] Scanner implements `Scanner` interface
- [ ] Tests use `fake.NewSimpleClientset()` — no real cluster needed
- [ ] Tests cover all problem types from spec + edge cases
- [ ] Knowledge base YAML files for all new problem types
- [ ] Scanner registered in `DefaultRegistry()`
- [ ] `make check` passes
- [ ] README updated with new resource type
