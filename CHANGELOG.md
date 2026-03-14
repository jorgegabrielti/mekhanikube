# Changelog

All notable changes to NautiKube are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2026-03-14

### 🚀 BREAKING CHANGE — Complete Architecture Rewrite

NautiKube v1.0.0 is a ground-up rewrite into a pure Go CLI following idiomatic Go practices. The application no longer depends on Docker, Ollama, or any external AI services.

### ✨ Added

- **Score-based prioritization** — every problem receives a 0–100 score based on severity and context
- **Embedded knowledge base** — remediation commands compiled into the binary via `//go:embed`
- **Scanner registry** — pluggable architecture with 5 built-in scanners:
  - `PodScanner` — CrashLoopBackOff, ImagePullBackOff, OOMKilled, Pending, high restarts, unready containers
  - `DeploymentScanner` — unavailable replicas, replica mismatch, stuck rollouts
  - `ServiceScanner` — no endpoints, pending LoadBalancer, ExternalName without target
  - `NodeScanner` — NotReady, memory/disk/PID pressure, cordoned nodes
  - `EventScanner` — high-frequency Warning event aggregation
- **Multiple output formats** — colorized table (default), JSON, YAML
- **CLI flags** — namespace filter (`-n`), resource filter (`-r`), minimum severity (`-s`), output format (`-o`), custom kubeconfig/context, `--no-color`
- **Quality gates** — `make check` runs fmt, vet, lint, vuln, test, and build in one command
- **CI/CD** — GitHub Actions pipeline (lint, test matrix, vuln scan, cross-platform build)
- **Release automation** — GoReleaser config for multi-platform binary releases
- **SDD infrastructure** — CONSTITUTION.md, scanner specs, skills, and workflow docs

### 🧪 Tested

- 49+ table-driven unit tests using stdlib `testing` + `client-go/kubernetes/fake`
- Race detector enabled on all tests
- Zero `go vet` warnings
- All files `gofmt`-formatted

### 🗑️ Removed

- `internal/ollama/` — AI client removed (replaced by embedded knowledge base)
- `docker-compose.yml` — Docker orchestration removed
- `configs/` — Dockerfile and entrypoint script removed
- `pkg/types/` — replaced by `internal/diagnosis/`
- `internal/analyzer/` — replaced by `scanner.Registry`

### 📦 Dependencies (minimal)

| Dependency | Purpose |
|---|---|
| `spf13/cobra` | CLI framework |
| `fatih/color` | Terminal colors |
| `k8s.io/client-go` | Kubernetes API access |
| `gopkg.in/yaml.v3` | YAML output |

All other functionality uses Go stdlib (`slog`, `encoding/json`, `testing`, `embed`, `errors`).

---

<details>
<summary>Legacy Changelog (v0.9.x and earlier)</summary>

## [0.9.1] - 2025-11-20

- Added severity enum (CRITICAL, HIGH, MEDIUM, LOW, INFO)
- Added score calculation with contextual adjustments
- 23 unit tests for scoring system

## [0.9.0] - 2025-11-20

- Version reset from v2.0.5 to v0.9.0 for honest versioning
- All features from v2.0.5 preserved

## [2.0.5] - 2025-11-20

- Advanced cluster provider detection
- Resilient connectivity with fallback strategies

## [2.0.4] - 2025-11-20

- Fixed kubeconfig handling with PyYAML

## [2.0.3] - 2025-11-19

- Optimized prompt engineering for LLM responses

## [2.0.2] - 2025-11-11

- Auto-detection for corporate environments (EKS + Proxy)

## [2.0.1] - 2025-01-10

- Renamed from Mekhanikube to NautiKube

## [2.0.0] - 2025-01

- Custom Go engine replacing K8sGPT
- Docker-first architecture with Ollama integration

## [1.0.0-legacy] - 2025-11-09

- Initial release with K8sGPT + Ollama via Docker Compose

</details>
