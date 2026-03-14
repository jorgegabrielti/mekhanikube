# How to Add a New Scanner

NautiKube uses a pluggable `Scanner` interface. If you want NautiKube to detect problems in `StatefulSets` or `Ingresses`, you just need to write a new Scanner.

We strongly follow **Spec-Driven Development (SDD)**.

## 1. Write the Specification

Before writing any code, create a specification file in `.agents/specs/`. For example, `.agents/specs/ingress-scanner.spec.md`:

```markdown
# Ingress Scanner Spec

## Scope
Scans `networking.k8s.io/v1/Ingress` resources.

## Scenarios
1. **No Backend Service**: The ingress points to a service that does not exist in the namespace. Severity: HIGH.
2. **Missing TLS Secret**: The ingress defines a TLS block but the referenced secret is missing. Severity: HIGH.
```

## 2. Write the Tests First

In Go, we use table-driven tests. Create `internal/scanner/ingress_test.go`:

```go
package scanner

import (
    "testing"
    "k8s.io/client-go/kubernetes/fake"
    // ... setup standard fake client testing
)

// Following the scenarios defined in the spec:
func TestIngressScanner(t *testing.T) {
    tests := []struct{
        name string
        // ... inputs and expected problem counts
    }{
        {
            name: "detects missing backend service",
            // ...
        },
    }
    // ...
}
```

## 3. Implement the Scanner

Create `internal/scanner/ingress.go` to implement the interface:

```go
package scanner

import (
    "context"
    "github.com/jorgegabrielti/nautikube/internal/diagnosis"
    "k8s.io/client-go/kubernetes"
)

type IngressScanner struct{}

func NewIngressScanner() *IngressScanner {
    return &IngressScanner{}
}

func (s *IngressScanner) Name() string {
    return "Ingress"
}

func (s *IngressScanner) Scan(ctx context.Context, client kubernetes.Interface, namespace string) ([]diagnosis.Problem, error) {
    var problems []diagnosis.Problem
    
    // ... fetch ingresses and apply logic
    
    return problems, nil
}
```

## 4. Register the Scanner

Finally, edit `internal/scanner/registry.go` so NautiKube actually runs your code:

```go
func DefaultRegistry() *Registry {
    return NewRegistry(
        NewPodScanner(),
        NewDeploymentScanner(),
        NewServiceScanner(),
        NewNodeScanner(),
        NewEventScanner(),
        NewIngressScanner(), // <--- Add your new scanner here
    )
}
```

Run `make test` and `make build` to verify!
