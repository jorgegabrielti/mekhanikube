# StatefulSet Scanner Specification

## Objective

Detect StatefulSets with replica mismatches in a live Kubernetes cluster.

## Problems Detected

### 1. Replica Mismatch

- **Condition**: `Status.ReadyReplicas < Status.Replicas`
- **Severity**: HIGH
- **Base Score**: 70
- **Remediation Key**: `statefulset_mismatch`
- **Remediation**:
  - `kubectl describe statefulset {name} -n {namespace}`
  - `kubectl get pods -n {namespace} -l <statefulset-selector>`
  - Check for PVC binding issues, node resource constraints, or Pod failures
  - `kubectl get events -n {namespace} --sort-by='.lastTimestamp'`

## Edge Cases

- StatefulSet with `ReadyReplicas == Replicas`: healthy, skip
- StatefulSet with `Replicas == 0`: no active pods expected, skip
- StatefulSet during rolling update: transient mismatch is expected but still reported for visibility
- Namespace filtering must be applied correctly

## Acceptance Criteria

- [x] Detects StatefulSets where `ReadyReplicas < Replicas`
- [x] StatefulSets with all replicas ready are not reported
- [x] StatefulSets with zero replicas are not reported
- [x] Multiple mismatched StatefulSets are all reported
- [x] Partial readiness (some ready, not all) is reported
- [x] Tests use fake K8s client
- [x] Zero false positives on healthy StatefulSets
