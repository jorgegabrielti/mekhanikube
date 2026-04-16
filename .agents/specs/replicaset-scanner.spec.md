# ReplicaSet Scanner Specification

## Objective

Detect orphaned ReplicaSets and ReplicaSets with replica mismatches in a live Kubernetes cluster.

## Problems Detected

### 1. Orphaned ReplicaSet

- **Condition**: ReplicaSet has no `OwnerReferences` AND `Status.Replicas > 0`
- **Severity**: MEDIUM
- **Base Score**: 50
- **Remediation Key**: `replicaset_orphaned`
- **Remediation**:
  - `kubectl describe replicaset {name} -n {namespace}`
  - Determine if this ReplicaSet was unintentionally orphaned (e.g., Deployment deleted)
  - Delete if no longer needed: `kubectl delete replicaset {name} -n {namespace}`

### 2. Replica Mismatch

- **Condition**: ReplicaSet has `OwnerReferences` AND `Status.ReadyReplicas < Status.Replicas`
- **Severity**: HIGH
- **Base Score**: 70
- **Remediation Key**: `replicaset_mismatch`
- **Remediation**:
  - `kubectl describe replicaset {name} -n {namespace}`
  - `kubectl get pods -n {namespace} -l <selector>`
  - Check owning Deployment for rollout issues

## Edge Cases

- Orphaned ReplicaSet with `Replicas == 0`: not active, skip (not an orphan problem)
- ReplicaSet owned by Deployment with all replicas ready: healthy, skip
- ReplicaSet with `ReadyReplicas == Replicas` and an owner: healthy, skip

## Acceptance Criteria

- [x] Detects orphaned ReplicaSets (no owner, active replicas) with MEDIUM severity
- [x] Detects replica mismatches in owned ReplicaSets with HIGH severity
- [x] Orphaned ReplicaSets with zero replicas are not reported
- [x] Healthy owned ReplicaSets are not reported
- [x] Multiple problems across different ReplicaSets are all reported
- [x] Tests use fake K8s client
- [x] Zero false positives on healthy ReplicaSets
