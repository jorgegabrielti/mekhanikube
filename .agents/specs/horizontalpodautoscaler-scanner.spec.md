# HorizontalPodAutoscaler Scanner Specification

## Objective

Detect HorizontalPodAutoscaler objects that have reached their maximum replica count in a live Kubernetes cluster.

## Problems Detected

### 1. HPA at Maximum Replicas

- **Condition**: `Status.CurrentReplicas >= Spec.MaxReplicas`
- **Severity**: MEDIUM
- **Base Score**: 50
- **Remediation Key**: `hpa_at_max`
- **Remediation**:
  - `kubectl describe hpa {name} -n {namespace}`
  - `kubectl top pods -n {namespace}` (if metrics-server available)
  - Consider increasing `maxReplicas` if load is consistently high
  - Profile the application for performance bottlenecks

## Edge Cases

- HPA below maximum replicas: healthy, skip
- HPA at maximum for a very short duration: still reported for visibility
- HPA with `MaxReplicas == 0`: should not occur in practice (Kubernetes enforces min 1)
- Target deployment may be over-scaled due to a metrics spike: still reported

## Acceptance Criteria

- [x] Detects HPAs where `CurrentReplicas >= MaxReplicas`
- [x] HPAs below maximum replicas are not reported
- [x] Multiple HPAs at max are all reported
- [x] Mixed HPAs: only those at max are reported
- [x] Tests use fake K8s client
- [x] Zero false positives on HPAs with available headroom
