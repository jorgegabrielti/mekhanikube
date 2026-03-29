# PersistentVolume Scanner Specification

## Objective

Detect PersistentVolumes in problematic states in a live Kubernetes cluster.

## Problems Detected

### 1. PersistentVolume in Failed State

- **Condition**: PV `Status.Phase == "Failed"`
- **Severity**: HIGH
- **Base Score**: 70
- **Remediation Key**: `pv_failed`
- **Remediation**:
  - `kubectl describe pv {name}`
  - Check cloud provider volume status
  - Consider deleting and recreating the PV if the underlying volume is healthy

### 2. PersistentVolume in Released State

- **Condition**: PV `Status.Phase == "Released"` (previously bound PVC was deleted)
- **Severity**: LOW
- **Base Score**: 20
- **Remediation Key**: `pv_released`
- **Remediation**:
  - `kubectl describe pv {name}`
  - Remove the `claimRef` from the PV spec to make it `Available` again
  - `kubectl patch pv {name} -p '{"spec":{"claimRef": null}}'`
  - Or delete and recreate the PV to reclaim storage

## Edge Cases

- PV in `Bound` state: healthy, skip
- PV in `Available` state: not yet claimed but not problematic, skip
- PV in `Pending` state: transitional, skip (not yet relevant)
- PersistentVolume is cluster-scoped (no namespace)

## Acceptance Criteria

- [x] Detects PVs in `Failed` state with HIGH severity
- [x] Detects PVs in `Released` state with LOW severity
- [x] PVs in `Bound` state are not reported
- [x] PVs in `Available` state are not reported
- [x] Mixed states: only Failed and Released are reported
- [x] Tests use fake K8s client
- [x] Zero false positives on healthy PVs
