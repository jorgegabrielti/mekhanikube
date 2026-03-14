# Service Scanner Specification

## Objective

Detect problems in Services of a live Kubernetes cluster.

## Problems Detected

### 1. No Endpoints

- **Condition**: Service exists but has no corresponding Endpoints (or Endpoints has zero addresses)
- **Severity**: HIGH
- **Base Score**: 70
- **Score Adjustments**: +10 if namespace is `kube-system` or `default`
- **Remediation Key**: `service-no-endpoints`
- **Remediation**:
  - `kubectl describe service {name} -n {namespace}`
  - `kubectl get endpoints {name} -n {namespace}`
  - `kubectl get pods -l <selector> -n {namespace}`
  - Check if selector matches any running pods

### 2. LoadBalancer Pending

- **Condition**: Service type is `LoadBalancer` and `Status.LoadBalancer.Ingress` is empty
- **Severity**: MEDIUM
- **Base Score**: 50
- **Remediation Key**: `service-lb-pending`
- **Remediation**:
  - `kubectl describe service {name} -n {namespace}`
  - Check cloud provider load balancer quota
  - On local clusters, install MetalLB or use NodePort instead

### 3. ExternalName Without Target

- **Condition**: Service type is `ExternalName` and `Spec.ExternalName` is empty
- **Severity**: LOW
- **Base Score**: 30
- **Remediation Key**: `service-externalname-empty`
- **Remediation**:
  - `kubectl describe service {name} -n {namespace}`
  - Set `spec.externalName` to the target DNS name

## Edge Cases

- Headless Services (ClusterIP: None): skip endpoint check (they work differently)
- Services without selectors (manual endpoints): skip endpoint auto-check
- Services in `kube-system` (kubernetes API service): skip, it's managed
- LoadBalancer on local clusters: always pending, report as LOW instead of MEDIUM

## Acceptance Criteria

- [ ] Detects all 3 problem types
- [ ] Correctly cross-references Services with Endpoints
- [ ] Skips headless services and services without selectors
- [ ] Tests use fake Services and Endpoints
- [ ] Zero false positives on healthy services
