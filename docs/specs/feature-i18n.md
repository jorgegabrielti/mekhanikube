# Feature Spec: Internationalization (i18n)

**Status:** Proposed
**Date:** 2026-04-14

## Context and Motivation

NautiKube currently displays all diagnostic output (explanations, commands, TUI labels, CLI messages) exclusively in English. Portuguese-speaking users — the primary initial audience — would benefit from native-language remediation guidance. Adding i18n support makes the tool more accessible and reduces the cognitive load when debugging under pressure.

## Functional Requirements

- **FR1:** Knowledge YAML files are organized by language using a suffix convention: `<key>.en.yaml` (English, default) and `<key>.pt.yaml` (Portuguese).
- **FR2:** A `--lang` flag is available on the `scan` command (values: `en`, `pt`; default: `en`). When set, it selects the corresponding language files for explanation and command output.
- **FR3:** If a translation file is missing for the requested language, the system falls back to the English (`en`) file silently.
- **FR4:** The `language` key in the user config file (`~/.nautikube/config.yaml`) sets the default language. The `--lang` CLI flag overrides it.
- **FR5:** TUI static strings (banner subtitle, column headers like "Score", "Issue", "Cause", "Fix", severity labels) are translatable via a `locales/<lang>.yaml` bundle embedded at build time.
- **FR6:** All existing knowledge YAML files are renamed to `<key>.en.yaml` and corresponding `<key>.pt.yaml` translations are provided.

## Non-Functional Requirements

- **NFR1:** No external i18n library dependency — use `embed.FS` and simple map lookups.
- **NFR2:** Adding a new language requires only creating new `<key>.<lang>.yaml` files and a `locales/<lang>.yaml` — no Go code changes.
- **NFR3:** Build size increase must stay under 200 KB for the Portuguese language pack.

## Acceptance Criteria

- **AC1:** Given `--lang pt` is passed, when a CrashLoopBackOff problem is found, then the explanation and commands are read from `crashloopbackoff.pt.yaml`.
- **AC2:** Given `--lang pt` is passed but `oomkilled.pt.yaml` does not exist, when an OOMKilled problem is found, then the system falls back to `oomkilled.en.yaml` without error.
- **AC3:** Given no `--lang` flag and no config file, when a scan runs, then English (`en`) is used by default.
- **AC4:** Given `language: pt` in `~/.nautikube/config.yaml` and no `--lang` flag, when a scan runs, then Portuguese is used.
- **AC5:** Given `language: pt` in config and `--lang en` on CLI, when a scan runs, then English is used (CLI overrides config).
- **AC6:** Given `--lang pt`, when the TUI renders, then column headers and labels are displayed in Portuguese (e.g., "Problema", "Causa", "Correção").
- **AC7:** Given a new language code `fr` with no files present, when `--lang fr` is passed, then all output falls back to English.

## File Layout After Migration

```
internal/diagnosis/knowledge/
  crashloopbackoff.en.yaml
  crashloopbackoff.pt.yaml
  high_restarts.en.yaml
  high_restarts.pt.yaml
  ...
internal/diagnosis/locales/
  en.yaml      # TUI/CLI static strings
  pt.yaml
```

## Known Risks and Limitations

- Translation quality must be manually reviewed; no machine translation without review.
- `windows_commands` section is language-independent (kubectl commands are the same) but is duplicated per file for consistency.
- Languages beyond `en` and `pt` are out of scope for v1.

## Impact on Tests

- `internal/diagnosis/knowledge_test.go` — test language selection, fallback, and placeholder replacement per language.
- New `internal/diagnosis/locales_test.go` — test TUI string loading per language.
- Existing scanner tests remain unchanged (they don't test remediation text content).
