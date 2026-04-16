# PersistentVolumeClaim Scanner Specification

## Objective

Detect PersistentVolumeClaims in problematic states in a live Kubernetes cluster.

## Problems Detected

### 1. PVC in Pending State

- **Condition**: PVC `Status.Phase == ClaimPending`
- **Severity**: HIGH
- **Base Score**: 70
- **Remediation Key**: `pvc_pending`
- **Remediation**:
  - `kubectl describe pvc {name} -n {namespace}`
  - Check if a matching PV exists with sufficient capacity and compatible access modes
  - Verify the StorageClass exists and is configured correctly
  - `kubectl get pv`

### 2. PVC in Lost State

- **Condition**: PVC `Status.Phase == ClaimLost` (backing PV was deleted)
- **Severity**: CRITICAL
- **Base Score**: 95
- **Remediation Key**: `pvc_lost`
- **Remediation**:
  - `kubectl describe pvc {name} -n {namespace}`
  - The backing PV was deleted. Data may be unrecoverable.
  - Check cloud provider snapshots or backups
  - Consider recreating the PVC and restoring from backup

## Edge Cases

- PVC in `Bound` state: healthy, skip
- Namespace filtering must be applied correctly
- `ClaimPending` PVCs in newly created namespaces with no StorageClass: detected

## Acceptance Criteria

- [x] Detects PVCs in `Pending` state with HIGH severity
- [x] Detects PVCs in `Lost` state with CRITICAL severity
- [x] PVCs in `Bound` state are not reported
- [x] Multiple problematic PVCs are all reported
- [x] Mixed states: only Pending and Lost are reported
- [x] Tests use fake K8s client
- [x] Zero false positives on healthy PVCs
