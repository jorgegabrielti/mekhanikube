# Feature Spec: Initial Configuration (`nautikube init`)

**Status:** Proposed
**Date:** 2026-04-14

## Context and Motivation

Users who run NautiKube regularly end up repeating the same flags (`--lang pt`, `--severity medium`, `--output yaml`). A one-time `nautikube init` command lets them persist preferences to a config file, reducing friction and improving the default experience. This follows the pattern established by tools like `aws configure`, `gh auth login`, and `kubectl config`.

## Functional Requirements

- **FR1:** A new `nautikube init` command launches an interactive prompt that asks for: language, minimum severity, and default output format.
- **FR2:** Answers are written to `~/.nautikube/config.yaml` (all platforms). The directory is created if it does not exist.
- **FR3:** If `~/.nautikube/config.yaml` already exists, `nautikube init` displays current values as defaults and allows the user to update them.
- **FR4:** Each config key can also be overridden by CLI flags at runtime. Priority order: CLI flag > config file > built-in default.
- **FR5:** A `nautikube config show` subcommand prints the resolved configuration (merged from defaults, config file, and any active CLI flags).
- **FR6:** If the config file is missing or malformed, NautiKube proceeds silently with built-in defaults — never errors on missing config.

## Config File Schema

```yaml
# ~/.nautikube/config.yaml
language: pt          # en | pt (default: en)
severity: medium      # critical | high | medium | low | info (default: low)
output: table         # table | yaml | csv | json (default: table)
```

## Non-Functional Requirements

- **NFR1:** Config file must be human-readable and hand-editable YAML.
- **NFR2:** No third-party prompt library — use standard I/O with `fmt.Scan` or `bufio.Scanner` for interactive input.
- **NFR3:** File permissions on created config: `0644` (file), `0755` (directory).
- **NFR4:** Config loading adds < 1ms to startup time.

## Acceptance Criteria

- **AC1:** Given no `~/.nautikube/` directory exists, when `nautikube init` runs and the user answers all prompts, then `~/.nautikube/config.yaml` is created with the correct values.
- **AC2:** Given `~/.nautikube/config.yaml` exists with `language: en`, when `nautikube init` runs, then the current language `en` is shown as the default and the user can press Enter to keep it or type a new value.
- **AC3:** Given `severity: medium` in config and no `--severity` CLI flag, when `nautikube scan` runs, then only problems with severity >= MEDIUM are displayed.
- **AC4:** Given `output: yaml` in config and `--output table` on CLI, when `nautikube scan` runs, then table format is used (CLI overrides config).
- **AC5:** Given a malformed `config.yaml` (invalid YAML), when `nautikube scan` runs, then built-in defaults are used and no error is shown to the user.
- **AC6:** Given a valid config file, when `nautikube config show` runs, then all resolved values are printed with their source (default / config / flag).
- **AC7:** Given `nautikube init` is run non-interactively (piped stdin), when EOF is reached before all prompts, then existing/default values are preserved for unanswered prompts.

## Known Risks and Limitations

- Config file location is not configurable via env var in v1 (always `~/.nautikube/`).
- No `nautikube config set <key> <value>` command in v1 — users edit YAML directly or re-run `init`.
- Kubeconfig path and namespace are intentionally excluded from v1 scope to avoid conflicting with `kubectl config` and `KUBECONFIG` env var.

## Impact on Tests

- New `internal/cli/init_test.go` — test interactive prompt, file creation, and update flow.
- New `internal/cli/config_test.go` — test config loading, merging priority, malformed file handling.
- Existing `internal/cli/scan_test.go` — add cases for config-driven defaults being overridden by flags.
