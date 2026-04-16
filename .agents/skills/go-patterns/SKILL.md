---
name: go-patterns
description: Go idioms and patterns enforced in this project
---

# Go Patterns Skill

Reference for Go idioms used across NautiKube. Follow these patterns in all code.

## Error Handling

### Sentinel Errors

Define in `internal/diagnosis/errors.go` or relevant package:

```go
var ErrNoKubeconfig = errors.New("nautikube: no kubeconfig found")
```

### Wrapping

Always add context when returning errors:

```go
pods, err := client.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{})
if err != nil {
    return nil, fmt.Errorf("failed to list pods in namespace %q: %w", ns, err)
}
```

### Checking

```go
if errors.Is(err, k8s.ErrNoKubeconfig) {
    // Handle specifically
}
```

## Functional Options

Use for any struct with optional configuration:

```go
type Option func(*Client)

func WithKubeconfig(path string) Option {
    return func(c *Client) {
        c.kubeconfigPath = path
    }
}

func New(opts ...Option) (*Client, error) {
    c := &Client{} // defaults
    for _, opt := range opts {
        opt(c)
    }
    // initialize...
    return c, nil
}
```

## Table-Driven Tests

Standard pattern for all test functions:

```go
func TestCalculateScore(t *testing.T) {
    t.Parallel()

    tests := []struct {
        name string
        // inputs
        want int
    }{
        {name: "case 1", ..., want: 90},
        {name: "case 2", ..., want: 70},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            got := calculateScore(tt.input)
            if got != tt.want {
                t.Errorf("calculateScore() = %d, want %d", got, tt.want)
            }
        })
    }
}
```

### Test Helpers

```go
func assertEqual(t *testing.T, got, want interface{}) {
    t.Helper()
    if got != want {
        t.Errorf("got %v, want %v", got, want)
    }
}
```

## Structured Logging

```go
import "log/slog"

// In functions:
slog.Info("scanning namespace", "namespace", ns, "scanner", s.Name())
slog.Error("failed to list pods", "error", err, "namespace", ns)
slog.Debug("found problem", "resource", pod.Name, "issue", reason)
```

## Interface Design

- Define interfaces in the **consumer** package, not the provider.
- Keep interfaces small (1-3 methods).
- Name single-method interfaces with `-er` suffix.

```go
// Good: defined where it's consumed
type Scanner interface {
    Scan(ctx context.Context, client kubernetes.Interface, ns string) ([]Problem, error)
}

// Bad: large interface covering everything
type EverythingDoer interface {
    Scan(...)
    Format(...)
    Report(...)
}
```

## Embedding Files

Use `//go:embed` for static data compiled into the binary:

```go
import "embed"

//go:embed knowledge/*.yaml
var knowledgeFS embed.FS
```

Access at runtime:

```go
data, err := knowledgeFS.ReadFile("knowledge/crashloopbackoff.yaml")
```
