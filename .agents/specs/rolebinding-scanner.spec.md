# RoleBinding Scanner Specification

## Objective

Detect RoleBinding objects with no subjects in a live Kubernetes cluster.

## Problems Detected

### 1. RoleBinding with No Subjects

- **Condition**: `len(Subjects) == 0`
- **Severity**: LOW
- **Base Score**: 20
- **Remediation Key**: `rbac_empty_binding`
- **Remediation**:
  - `kubectl describe rolebinding {name} -n {namespace}`
  - Add the appropriate subjects (User, Group, or ServiceAccount) to the binding
  - Or delete the binding if it is no longer needed:
    `kubectl delete rolebinding {name} -n {namespace}`

## Edge Cases

- RoleBinding with at least one subject: valid, skip
- Empty RoleBinding with multiple roles: still reported (subjects is the check)
- RoleBinding grants permissions to nobody — no immediate security risk, but is dead configuration

## Acceptance Criteria

- [x] Detects RoleBindings with no subjects
- [x] RoleBindings with at least one subject are not reported
- [x] Multiple empty RoleBindings are all reported
- [x] Mixed RoleBindings: only empty ones are reported
- [x] Tests use fake K8s client
- [x] Zero false positives on properly bound RoleBindings
