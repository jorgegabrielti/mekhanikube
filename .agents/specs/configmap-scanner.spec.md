# ConfigMap Scanner Specification

## Objective

Detect oversized ConfigMap objects in a live Kubernetes cluster.

## Problems Detected

### 1. Oversized ConfigMap

- **Condition**: Total size of `data` + `binaryData` fields >= 500KB
- **Severity**: MEDIUM
- **Base Score**: 50
- **Remediation Key**: `configmap_large`
- **Remediation**:
  - `kubectl describe configmap {name} -n {namespace}`
  - Split large configs into multiple smaller ConfigMaps
  - Consider storing large blobs in a dedicated secret store or object storage

## Edge Cases

- ConfigMap with no data or empty maps: skip, don't report
- ConfigMap with only `binaryData`: size is counted correctly
- ConfigMap with both `data` and `binaryData`: total size is the sum of both
- Threshold is inclusive (500KB triggers the problem)

## Acceptance Criteria

- [x] Detects ConfigMaps >= 500KB
- [x] Counts both `data` (string values) and `binaryData` (byte values) toward total size
- [x] Empty ConfigMaps do not trigger any problem
- [x] ConfigMaps just under 500KB are not reported
- [x] Tests use fake K8s client
- [x] Zero false positives on healthy small ConfigMaps
