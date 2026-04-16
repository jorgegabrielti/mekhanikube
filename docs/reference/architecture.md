# Architectural Structure

NautiKube is structured idiomatically according to Go community standards. The codebase relies heavily on the `internal/` package convention to restrict public API surface.

## Package Layout

```text
cmd/nautikube/          # Entry point (Main)
└── main.go             # Simply calls cli.Execute()

internal/
├── cli/                 # Command routing
│   ├── root.go          # Cobra Root
│   ├── scan.go          # Cobra "scan" command
│   └── version.go       # Cobra "version" command
│
├── k8s/                 # Kubernetes API
│   └── client.go        # Client factory with 'Functional Options' pattern
│
├── scanner/             # Detection Logic
│   ├── scanner.go       # interface Scanner { Scan(...) }
│   ├── registry.go      # Orchestrator running scanners concurrently
│   ├── deployment.go    # Implementation
│   ├── event.go         # Implementation
│   ├── node.go          # Implementation
│   ├── pod.go           # Implementation
│   └── service.go       # Implementation
│
├── diagnosis/           # Business Logic Models
│   ├── problem.go       # Core types: Problem, Severity, Score math
│   ├── errors.go        # Sentinel errors (ErrNoKubeconfig, etc)
│   ├── knowledge.go     # Parse embedded troubleshooting knowledge
│   └── knowledge/       # YAML files embedded via //go:embed
│       ├── crashloopbackoff.yaml
│       └── imagepullbackoff.yaml
│
└── output/              # Formatting output
    ├── formatter.go     # interface Formatter { Format(...) }
    ├── json.go          # Implementation
    ├── table.go         # Implementation (colorized)
    └── yaml.go          # Implementation
```

## Key Tradeoffs

1. **No External Dependencies (Besides stdlib and k8s client)**
   We do not use `testify` for testing, choosing raw table-driven tests. The only external UI dependency is `fatih/color` to keep binary size tiny.

2. **No Dependency Injection Framework**
   Instead of using tools like `Wire` or `Dig`, we manually pass interfaces (like `kubernetes.Interface`) down the call stack. This simplifies debugging and test mocking.

3. **//go:embed vs Remote Data**
   The knowledge base of how to fix problems (`kubectl` commands) is compiled directly into the binary using `//go:embed`. This guarantees the tool can offer help even in air-gapped environments.
