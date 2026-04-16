# DaemonSet Scanner Specification

## Objective

Detect DaemonSets that are not fully scheduled across all nodes in a live Kubernetes cluster.

## Problems Detected

### 1. DaemonSet Not Fully Scheduled

- **Condition**: `Status.NumberReady < Status.DesiredNumberScheduled`
- **Severity**: HIGH
- **Base Score**: 70
- **Remediation Key**: `daemonset_misscheduled`
- **Remediation**:
  - `kubectl describe daemonset {name} -n {namespace}`
  - `kubectl get pods -n {namespace} -l <daemonset-selector> -o wide`
  - Check for node taints, resource pressure, or image pull failures on affected nodes
  - `kubectl get events -n {namespace} --sort-by='.lastTimestamp'`

## Edge Cases

- DaemonSet with `DesiredNumberScheduled == 0`: no nodes match the selector, skip
- DaemonSet with `NumberReady == DesiredNumberScheduled`: healthy, skip
- Partial scheduling (some nodes ready, not all) is reported
- DaemonSets in `kube-system` (e.g., `kube-proxy`, `fluentd`) are treated the same as application DaemonSets

## Acceptance Criteria

- [x] Detects DaemonSets where `NumberReady < DesiredNumberScheduled`
- [x] DaemonSets with all nodes ready are not reported
- [x] DaemonSets with zero ready nodes are reported
- [x] Multiple misscheduled DaemonSets are all reported
- [x] Mixed healthy and unhealthy DaemonSets: only unhealthy are reported
- [x] Tests use fake K8s client
- [x] Zero false positives on fully scheduled DaemonSets
