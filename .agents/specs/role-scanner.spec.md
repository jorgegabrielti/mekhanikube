# Role Scanner Specification

## Objective

Detect RBAC Role objects with overly permissive wildcard verbs in a live Kubernetes cluster.

## Problems Detected

### 1. Role with Wildcard Verbs

- **Condition**: Any `PolicyRule` in `Rules` contains `"*"` in its `Verbs` list
- **Severity**: MEDIUM
- **Base Score**: 50
- **Remediation Key**: `rbac_wildcard`
- **Remediation**:
  - `kubectl describe role {name} -n {namespace}`
  - Replace wildcard verbs with specific verbs (`get`, `list`, `watch`, `create`, `update`, `delete`)
  - Follow principle of least privilege (PoLP)

## Edge Cases

- Role with only specific verbs (e.g., `["get", "list"]`): safe, skip
- Role with mixed verbs including `"*"` (e.g., `["get", "*"]`): reported — wildcard is present
- Role with no rules: skip
- Wildcard on resources while verbs are specific: this check is only for verbs

## Acceptance Criteria

- [x] Detects Roles where any rule has `"*"` in `Verbs`
- [x] Roles with only specific verbs are not reported
- [x] Roles with mixed verbs including wildcard are reported
- [x] Multiple wildcard Roles are all reported
- [x] Mixed Roles: only wildcard ones are reported
- [x] Tests use fake K8s client
- [x] Zero false positives on least-privilege Roles
