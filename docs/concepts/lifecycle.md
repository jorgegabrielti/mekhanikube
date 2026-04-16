# Project Lifecycle & Quality

Because NautiKube executes inside sensitive production environments and accesses Kubernetes APIs, it enforces strict quality gates on its codebase.

## The Quality Gate Sequence

You cannot merge code into `main` without passing the sequential quality pipeline. This is driven by `make check` locally, and GitHub Actions remotely.

1. **`go fmt`**: Strictly formats all code.
2. **`go vet`**: Performs lexical analysis to catch shadow variables and common bugs.
3. **`golangci-lint`**: Runs 50+ aggressive linters. We strictly enforce cyclomatic complexity and unhandled errors.
4. **`govulncheck`**: Scans the Go Abstract Syntax Tree explicitly against known CVE vulnerabilities in dependencies.
5. **`go test -race`**: Runs tests while injecting race-condition detection.

If a pull request fails any of these 5 steps, it is automatically blocked.

## Multi-Platform Release

We use [GoReleaser](https://goreleaser.com/) to handle the binary lifecycle.

When a Git tag (`v1.0.0`) is pushed, a GitHub action triggers GoReleaser which:
1. Compiles statically linked binaries for `linux`, `darwin` (macOS), and `windows` across `amd64` and `arm64` architectures.
2. Strips debugging symbols (`-s -w`) to reduce binary size.
3. Injects the Git tag into the binary version variable via `-ldflags`.
4. Calculates SHA256 checksums to guarantee integrity.
5. Parses `CHANGELOG.md` and generates a GitHub release page.

This guarantees that local development environments generate exactly the same reproducible binaries as our CI/CD systems.
