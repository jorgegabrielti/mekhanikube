# Contributing to NautiKube

Thank you for your interest in contributing! This guide will help you get started.

## Prerequisites

- **Go 1.22+** — [Download](https://go.dev/dl/)
- **kubectl** configured with a valid kubeconfig (for integration tests)
- **golangci-lint** — [Install](https://golangci-lint.run/welcome/install/)
- **govulncheck** — `go install golang.org/x/vuln/cmd/govulncheck@latest`

## Getting Started

```bash
git clone https://github.com/jorgegabrielti/nautikube.git
cd nautikube
make check     # Run all quality gates
```

## Development Workflow

1. **Fork** the repository and create a feature branch from `main`.
2. Write your code following [Effective Go](https://go.dev/doc/effective_go) idioms.
3. Write **table-driven tests** using the stdlib `testing` package (no testify).
4. Run `make check` to verify all quality gates pass.
5. Commit using [Conventional Commits](https://www.conventionalcommits.org/) format.
6. Open a Pull Request against `main`.

### Commit Message Format

```
<type>(<scope>): <description>

feat(scanner): add StatefulSet scanner
fix(output): handle empty problem list in JSON formatter
docs(readme): update installation instructions
refactor(k8s): simplify client option handling
test(scanner): add edge cases for node pressure conditions
```

## Quality Gates

All PRs must pass the following before merge:

```bash
make fmt          # Code formatting (gofmt)
make vet          # Bug detection (go vet)
make lint         # 50+ linters (golangci-lint)
make vuln         # CVE scanning (govulncheck)
make test         # Unit tests with race detector
make build        # Binary build with ldflags
```

Or run them all at once:

```bash
make check
```

## Project Structure

```
cmd/nautikube/               → Entry point
internal/
├── cli/                     → Cobra CLI commands
├── k8s/                     → Kubernetes client factory
├── scanner/                 → Scanner interface + implementations
├── diagnosis/               → Problem types, scoring, knowledge base
└── output/                  → Output formatters (table, JSON, YAML)
.agents/
├── CONSTITUTION.md          → Project laws and constraints
├── specs/                   → Feature specifications
├── skills/                  → Development skill guides
└── workflows/               → Development workflows
```

## Adding a New Scanner

Refer to `.agents/skills/scanner-development/SKILL.md` for the complete guide. In summary:

1. **Write a spec** in `.agents/specs/`
2. **Write tests first** (SDD: spec → test → implement)
3. **Implement** the `scanner.Scanner` interface
4. **Register** the scanner in `internal/scanner/registry.go`
5. **Add knowledge entries** in `internal/diagnosis/knowledge/`

## Code Style

- Follow the [NautiKube Constitution](.agents/CONSTITUTION.md)
- Accept interfaces, return structs
- `context.Context` as first parameter on all I/O functions
- Sentinel errors with `errors.New()` and `errors.Is()`
- No global mutable state
- All comments and error messages in English

## Reporting Issues

- Use GitHub Issues with clear reproduction steps
- Include `nautikube version` output
- Include relevant `kubectl` output if applicable
