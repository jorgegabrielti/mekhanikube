# Deployment Scanner Specification

## Objective

Detect problems in Deployments of a live Kubernetes cluster.

## Problems Detected

### 1. Unavailable Replicas

- **Condition**: `Status.UnavailableReplicas > 0`
- **Severity**: HIGH
- **Base Score**: 70
- **Score Adjustments**: +10 if namespace is `kube-system` or `default`
- **Remediation Key**: `deployment-unavailable`
- **Remediation**:
  - `kubectl describe deployment {name} -n {namespace}`
  - `kubectl get pods -l app={name} -n {namespace}`
  - `kubectl rollout status deployment/{name} -n {namespace}`

### 2. Replicas Mismatch

- **Condition**: `Status.ReadyReplicas < *Spec.Replicas` (and UnavailableReplicas == 0)
- **Severity**: MEDIUM
- **Base Score**: 50
- **Remediation Key**: `deployment-mismatch`
- **Remediation**:
  - `kubectl describe deployment {name} -n {namespace}`
  - `kubectl get replicaset -l app={name} -n {namespace}`

### 3. Rollout Stuck

- **Condition**: Deployment has `Progressing` condition with status `False`
- **Severity**: HIGH
- **Base Score**: 70
- **Remediation Key**: `deployment-stuck`
- **Remediation**:
  - `kubectl rollout status deployment/{name} -n {namespace}`
  - `kubectl rollout undo deployment/{name} -n {namespace}`
  - `kubectl get events -n {namespace} --field-selector involvedObject.name={name}`

### 4. Zero Replicas

- **Condition**: `Spec.Replicas != nil && *Spec.Replicas == 0`
- **Severity**: LOW
- **Base Score**: 30
- **Remediation Key**: `deployment-zero-replicas`
- **Remediation**:
  - `kubectl scale deployment/{name} -n {namespace} --replicas=1`

## Edge Cases

- Deployments with nil `Spec.Replicas` (defaults to 1): use 1 as default
- Deployments being actively rolled out (UnavailableReplicas temporary): still report
- Deployments in `kube-system` with 0 ready (e.g., coredns down): CRITICAL override

## Acceptance Criteria

- [x] Detects all 4 problem types
- [x] Score calculated with contextual adjustments
- [x] Tests use fake Deployments with various status conditions
- [x] Handles nil Spec.Replicas gracefully
- [x] Zero false positives on healthy deployments
