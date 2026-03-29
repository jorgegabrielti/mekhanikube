# ClusterRoleBinding Scanner Specification

## Objective

Detect ClusterRoleBinding objects with no subjects in a live Kubernetes cluster.

## Problems Detected

### 1. ClusterRoleBinding with No Subjects

- **Condition**: `len(Subjects) == 0`
- **Severity**: LOW
- **Base Score**: 20
- **Remediation Key**: `rbac_cluster_empty_binding`
- **Remediation**:
  - `kubectl describe clusterrolebinding {name}`
  - Add the appropriate subjects (User, Group, or ServiceAccount) to the binding
  - Or delete the binding if no longer needed:
    `kubectl delete clusterrolebinding {name}`

## Edge Cases

- ClusterRoleBinding with at least one subject: valid, skip
- Empty binding on a high-privilege ClusterRole (e.g., `cluster-admin`): still LOW severity since no subjects means no actual grant
- ClusterRoleBinding is cluster-scoped (no namespace)

## Acceptance Criteria

- [x] Detects ClusterRoleBindings with no subjects
- [x] ClusterRoleBindings with at least one subject are not reported
- [x] Multiple empty ClusterRoleBindings are all reported
- [x] Mixed bindings: only empty ones are reported
- [x] Tests use fake K8s client
- [x] Zero false positives on properly bound ClusterRoleBindings
