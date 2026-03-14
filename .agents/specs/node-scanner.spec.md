# Node Scanner Specification

## Objective

Detect problems in Nodes of a live Kubernetes cluster.

## Problems Detected

### 1. Node Not Ready

- **Condition**: Node has condition `Ready` with status `False` or `Unknown`
- **Severity**: CRITICAL
- **Base Score**: 90
- **Remediation Key**: `node-not-ready`
- **Remediation**:
  - `kubectl describe node {name}`
  - `kubectl get events --field-selector involvedObject.name={name}`
  - Check kubelet logs on the node

### 2. Memory Pressure

- **Condition**: Node has condition `MemoryPressure` with status `True`
- **Severity**: HIGH
- **Base Score**: 70
- **Remediation Key**: `node-memory-pressure`
- **Remediation**:
  - `kubectl top node {name}`
  - `kubectl describe node {name}` (check Allocatable vs Capacity)
  - Identify high memory pods: `kubectl top pods -A --sort-by=memory`

### 3. Disk Pressure

- **Condition**: Node has condition `DiskPressure` with status `True`
- **Severity**: HIGH
- **Base Score**: 70
- **Remediation Key**: `node-disk-pressure`
- **Remediation**:
  - `kubectl describe node {name}`
  - Check disk usage on the node
  - Clean up unused images: `docker system prune` or `crictl rmi --prune`

### 4. PID Pressure

- **Condition**: Node has condition `PIDPressure` with status `True`
- **Severity**: HIGH
- **Base Score**: 70
- **Remediation Key**: `node-pid-pressure`
- **Remediation**:
  - `kubectl describe node {name}`
  - Identify pods with many processes

### 5. Unschedulable

- **Condition**: `Spec.Unschedulable == true` (node cordoned)
- **Severity**: LOW
- **Base Score**: 30
- **Remediation Key**: `node-unschedulable`
- **Remediation**:
  - `kubectl uncordon {name}`
  - Note: node may have been intentionally cordoned for maintenance

## Edge Cases

- Control plane nodes (may have taints but are healthy): don't flag taints as issues
- Single-node clusters: NotReady is CRITICAL (entire cluster down)
- Node with multiple pressure conditions: report each separately

## Acceptance Criteria

- [ ] Detects all 5 problem types
- [ ] Correctly reads Node conditions
- [ ] Tests use fake Nodes with various conditions
- [ ] Zero false positives on healthy nodes
