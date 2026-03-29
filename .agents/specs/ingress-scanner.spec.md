# Ingress Scanner Specification

## Objective

Detect Ingress resources with no routing configuration in a live Kubernetes cluster.

## Problems Detected

### 1. Empty Ingress (No Rules and No Default Backend)

- **Condition**: `len(Spec.Rules) == 0 AND Spec.DefaultBackend == nil`
- **Severity**: MEDIUM
- **Base Score**: 50
- **Remediation Key**: `ingress_empty`
- **Remediation**:
  - `kubectl describe ingress {name} -n {namespace}`
  - Add routing rules or a default backend to the Ingress spec
  - Delete if no longer needed: `kubectl delete ingress {name} -n {namespace}`

## Edge Cases

- Ingress with rules defined: healthy, skip
- Ingress with only a `DefaultBackend` (no rules): valid, skip
- Ingress with both rules and default backend: healthy, skip
- Empty rules array combined with no default backend: reported

## Acceptance Criteria

- [x] Detects Ingresses with no rules and no default backend
- [x] Ingresses with rules are not reported
- [x] Ingresses with only a default backend are not reported
- [x] Multiple empty Ingresses are all reported
- [x] Mixed Ingresses: only empty ones reported
- [x] Tests use fake K8s client
- [x] Zero false positives on properly configured Ingresses
