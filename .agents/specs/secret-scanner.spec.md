# Secret Scanner Specification

## Objective

Detect empty or unused Secret objects in a live Kubernetes cluster.

## Problems Detected

### 1. Empty Secret

- **Condition**: Secret has no `data` entries and no `stringData` entries
- **Severity**: LOW
- **Base Score**: 20
- **Remediation Key**: `secret_empty`
- **Remediation**:
  - `kubectl describe secret {name} -n {namespace}`
  - Populate the secret with the required values or delete it if unused

## Edge Cases

- Secrets of type `kubernetes.io/service-account-token`: skip (managed by Kubernetes)
- Secrets of type `helm.sh/release.v1`: skip (managed by Helm)
- Secrets with `stringData` but no `data`: `stringData` counts as populated (not empty)
- TLS secrets, Docker registry secrets with valid data: not reported
- `data` map with entries (even if values are empty byte slices): considered populated

## Acceptance Criteria

- [x] Detects Opaque Secrets with no `data` and no `stringData`
- [x] Skips service account token secrets (`kubernetes.io/service-account-token`)
- [x] Skips Helm release secrets (`helm.sh/release.v1`)
- [x] Secrets with `stringData` populated are not reported
- [x] Secrets with `data` populated are not reported
- [x] Tests use fake K8s client
- [x] Zero false positives on healthy secrets
