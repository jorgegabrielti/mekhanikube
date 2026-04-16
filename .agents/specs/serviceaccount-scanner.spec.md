# ServiceAccount Scanner Specification

## Objective

Detect ServiceAccounts with no associated secrets or image pull secrets in a live Kubernetes cluster.

## Problems Detected

### 1. ServiceAccount with No Secrets

- **Condition**: ServiceAccount has no entries in `Secrets` and no entries in `ImagePullSecrets`
- **Severity**: LOW
- **Base Score**: 20
- **Remediation Key**: `serviceaccount_no_secrets`
- **Remediation**:
  - `kubectl describe serviceaccount {name} -n {namespace}`
  - Verify whether this ServiceAccount requires authentication tokens
  - Associate the correct image pull secret if pulling from a private registry
  - `kubectl create secret docker-registry regcred --docker-server=... -n {namespace}`

## Edge Cases

- `default` ServiceAccount in most namespaces has no secrets in newer Kubernetes versions (1.24+):
  this is intentional behavior (ServiceAccount tokens are now short-lived projected tokens).
  **However**, the scanner still reports it for visibility — operators can suppress as needed.
- ServiceAccount with only `ImagePullSecrets` is considered non-empty (not reported)
- ServiceAccount with only `Secrets` is considered non-empty (not reported)

## Acceptance Criteria

- [x] Detects ServiceAccounts with no `Secrets` and no `ImagePullSecrets`
- [x] ServiceAccounts with only `Secrets` are not reported
- [x] ServiceAccounts with only `ImagePullSecrets` are not reported
- [x] ServiceAccounts with both lists are not reported
- [x] Multiple empty ServiceAccounts are all reported
- [x] Tests use fake K8s client
- [x] Zero false positives on properly configured ServiceAccounts
