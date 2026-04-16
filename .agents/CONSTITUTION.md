# NautiKube — Project Constitution

## Identity

NautiKube is a portable, single-binary CLI written in Go for Kubernetes cluster diagnostics.
It scans live clusters, detects problems, prioritizes them with a 0-100 score, and provides
actionable remediation with kubectl commands. No external dependencies at runtime.

## Architecture Rules

### Package Layout

```
cmd/nautikube/main.go     → Entry point only (~15 lines). Calls cli.Execute().
internal/cli/              → Cobra commands. No business logic here.
internal/k8s/              → Kubernetes client factory. Returns kubernetes.Interface.
internal/scanner/           → Scanner interface + implementations (one file per resource).
internal/diagnosis/         → Problem types, severity, score, knowledge base.
internal/output/            → Formatter interface + implementations (table, json, yaml).
```

### Design Principles

1. **Accept interfaces, return structs.** All cross-package dependencies use interfaces.
2. **Functional options** for configuration: `k8s.New(k8s.WithKubeconfig(path))`.
3. **Dependency injection** via constructor functions. No global state.
4. **`context.Context` as first parameter** on all functions that do I/O.
5. **`io.Writer` for output.** Formatters write to a writer, never directly to os.Stdout.
6. **Scanner interface is sacred.** New scanners implement it; never modify the interface.

## Go Idioms

### Language Version

- **Go 1.22+** minimum. Use modern features (range over int, etc.).

### Errors

- Wrap errors with context: `fmt.Errorf("failed to scan pods: %w", err)`.
- Define **sentinel errors** for known failure modes:
  ```go
  var ErrNoKubeconfig = errors.New("nautikube: no kubeconfig found")
  ```
- Check errors with `errors.Is()` and `errors.As()`. Never compare error strings.
- Never ignore errors silently. If intentionally ignoring, add a comment explaining why.

### Logging

- Use **`log/slog`** (stdlib). Never `fmt.Println` for diagnostic output.
- User-facing output goes through `output.Formatter`. Logs go through `slog`.
- Log levels: `slog.Debug` for development, `slog.Info` for operational, `slog.Error` for failures.

### Testing

- Use **stdlib `testing` package only**. No testify, no gomock.
- **Table-driven tests** for all logic functions.
- Use `t.Helper()` in test helper functions.
- Use `t.Parallel()` where safe.
- Unit tests: no build tag (run by default).
- Integration tests: `//go:build integration` build tag.
- Use `k8s.io/client-go/kubernetes/fake` for mocking K8s API.

### Naming

- Follow [Effective Go naming conventions](https://go.dev/doc/effective_go#names).
- Exported names: `PascalCase`. Unexported: `camelCase`.
- Interfaces: single-method interfaces use `-er` suffix (e.g., `Scanner`, `Formatter`).
- Files: `snake_case.go`. Test files: `snake_case_test.go`.

### Documentation

- All exported types and functions have **godoc comments in English**.
- Comments start with the name of the thing being documented.
- Example: `// Scanner defines the interface for resource scanners.`

### Dependencies

- **Minimize external dependencies.** "A little copying is better than a little dependency."
- Approved dependencies:
  - `github.com/spf13/cobra` — CLI framework
  - `github.com/fatih/color` — Terminal colors
  - `k8s.io/client-go` — Kubernetes API client
  - `gopkg.in/yaml.v3` — YAML encoding
- Any new dependency requires explicit justification.

## Code Prohibitions

- ❌ No global variables (except sentinel errors and `//go:embed` vars).
- ❌ No `os.Exit()` outside `main()`.
- ❌ No `log.Fatal()` or `log.Panic()`. Return errors instead.
- ❌ No `fmt.Println()` for logging. Use `slog`.
- ❌ No hardcoded version strings. Inject via `-ldflags` at build time.
- ❌ No `init()` functions. Use explicit initialization.
- ❌ No `interface{}` or `any` without strong justification.

## Git Conventions

### Branching

- `main` — stable, always passes `make check`.
- `feat/<name>` — new features.
- `fix/<name>` — bug fixes.
- `refactor/<name>` — code improvements without behavior change.
- `docs/<name>` — documentation only.
- `chore/<name>` — tooling, CI, dependencies.

### Commits

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
feat(scanner): add deployment scanner
fix(pod): handle nil containerStatuses
test(scanner): add node scanner edge cases
docs: update README with new resource types
refactor(output): extract severity icon helper
chore: update Go to 1.22
```

### Quality Gate

Every commit must pass `make check` (fmt → vet → lint → vuln → test → build).
