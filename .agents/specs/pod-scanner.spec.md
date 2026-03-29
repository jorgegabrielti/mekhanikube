# Pod Scanner Specification

## Objective

Detect problems in Pods of a live Kubernetes cluster.

## Problems Detected

### 1. CrashLoopBackOff

- **Condition**: Container in `Waiting` state with reason `CrashLoopBackOff`
- **Severity**: CRITICAL
- **Base Score**: 90
- **Score Adjustments**: +10 if namespace is `kube-system` or `default`
- **Remediation Key**: `crashloopbackoff`
- **Remediation**:
  - `kubectl logs {pod} -n {namespace} --previous`
  - `kubectl describe pod {pod} -n {namespace}`
  - `kubectl get events -n {namespace} --field-selector involvedObject.name={pod}`

### 2. ImagePullBackOff / ErrImagePull

- **Condition**: Container in `Waiting` state with reason `ImagePullBackOff` or `ErrImagePull`
- **Severity**: HIGH
- **Base Score**: 70
- **Remediation Key**: `imagepullbackoff`
- **Remediation**:
  - `kubectl describe pod {pod} -n {namespace}`
  - Verify image name and tag exist in registry
  - Check imagePullSecrets if using private registry

### 3. OOMKilled

- **Condition**: Container in `Terminated` state with reason `OOMKilled`
- **Severity**: CRITICAL
- **Base Score**: 90
- **Remediation Key**: `oomkilled`
- **Remediation**:
  - `kubectl describe pod {pod} -n {namespace}` (check resource limits)
  - `kubectl top pod {pod} -n {namespace}` (if metrics-server available)
  - Increase memory limits in deployment spec

### 4. High Restart Count

- **Condition**: `RestartCount > threshold`
- **Severity**:
  - `> 50`: CRITICAL (base score 90)
  - `> 20`: HIGH (base score 70)
  - `> 5`: MEDIUM (base score 50)
- **Remediation Key**: `high-restarts`
- **Remediation**:
  - `kubectl logs {pod} -n {namespace} --previous`
  - `kubectl describe pod {pod} -n {namespace}`

### 5. Pod Pending

- **Condition**: `Phase == Pending` and no container statuses (not yet scheduled)
- **Severity**: HIGH
- **Base Score**: 70
- **Remediation Key**: `pod-pending`
- **Remediation**:
  - `kubectl describe pod {pod} -n {namespace}` (check Events section)
  - `kubectl get events -n {namespace}`
  - Common causes: insufficient resources, node selector mismatch, PVC pending

### 6. Container Not Ready

- **Condition**: Container running but `Ready == false`
- **Severity**: MEDIUM
- **Base Score**: 50
- **Remediation Key**: `container-not-ready`
- **Remediation**:
  - `kubectl describe pod {pod} -n {namespace}` (check readiness probe)
  - `kubectl logs {pod} -n {namespace}`

## Edge Cases

- Pod with no `ContainerStatuses` (just created): skip, don't report
- Pod with `Phase == Succeeded` (completed Job): skip entirely
- Pod with `Phase == Failed` and no container info: report as HIGH
- Init containers failing: detect separately with `InitContainerStatuses`
- Multiple containers in same Pod with different issues: report each separately
- Healthy pods in `kube-system`: must NOT produce false positives

## Acceptance Criteria

- [x] Detects all 6 problem types above
- [x] Score calculated correctly with contextual adjustments
- [x] Each problem has remediation key linking to knowledge base
- [x] Tests cover all 6 scenarios using fake K8s client
- [x] Tests cover all edge cases listed above
- [x] Zero false positives on a healthy cluster
- [x] Handles RBAC permission errors gracefully (skip, don't crash)
