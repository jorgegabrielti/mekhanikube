# Cluster Scanner Specification

## Objective

Perform global cluster-level health checks that are not tied to a specific namespace or resource type.

## Problems Detected

### 1. API Server Unhealthy

- **Condition**: GET `/healthz` on the API server returns an error
- **Severity**: CRITICAL
- **Base Score**: 100
- **Remediation Key**: `cluster-api-unhealthy`
- **Remediation**:
  - Check control plane node status
  - Review API server pod logs: `kubectl logs -n kube-system -l component=kube-apiserver`

### 2. CoreDNS Missing

- **Condition**: No pods found in `kube-system` matching `k8s-app=kube-dns`
- **Severity**: CRITICAL
- **Base Score**: 95
- **Remediation Key**: `cluster-coredns-missing`
- **Remediation**:
  - `kubectl get pods -n kube-system -l k8s-app=kube-dns`
  - Reinstall CoreDNS via the cluster addon manager or `kubeadm`

### 3. CoreDNS Unhealthy

- **Condition**: CoreDNS pods exist but none are in `Running` state with `Ready=True`
- **Severity**: CRITICAL
- **Base Score**: 95
- **Remediation Key**: `cluster-coredns-unhealthy`
- **Remediation**:
  - `kubectl describe pods -n kube-system -l k8s-app=kube-dns`
  - `kubectl logs -n kube-system -l k8s-app=kube-dns`

### 4. Node Version Skew

- **Condition**: Nodes in the cluster are running more than one distinct Kubernetes
  kubelet version (version skew detected)
- **Severity**: HIGH
- **Base Score**: 60
- **Remediation Key**: `cluster-version-skew`
- **Remediation**:
  - `kubectl get nodes -o wide`
  - Upgrade nodes to a consistent kubelet version
  - Refer to the Kubernetes version skew support policy

## Edge Cases

- Fake client (unit tests): REST client unavailable — API Server check is skipped gracefully
- Single-node cluster: version skew check should not trigger false positives
- Multiple CoreDNS pods where some are healthy: not reported (at least one healthy)
- Namespace filter specified by user: Cluster scanner still performs cluster-wide checks

## Acceptance Criteria

- [x] Detects API Server health failures with CRITICAL severity
- [x] Detects missing CoreDNS with CRITICAL severity
- [x] Detects unhealthy CoreDNS (none ready) with CRITICAL severity
- [x] Detects node version skew with HIGH severity
- [x] Healthy API Server produces an INFO finding (not a problem)
- [x] Healthy CoreDNS produces an INFO finding (not a problem)
- [x] Handles REST client unavailability gracefully (fake client)
- [x] Tests use fake K8s client where applicable
