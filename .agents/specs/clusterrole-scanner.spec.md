# ClusterRole Scanner Specification

## Objective

Detect ClusterRole objects with overly permissive wildcard verbs in a live Kubernetes cluster.

## Problems Detected

### 1. ClusterRole with Wildcard Verbs

- **Condition**: Any `PolicyRule` in `Rules` contains `"*"` in its `Verbs` list,
  and the ClusterRole name does NOT start with `"system:"`
- **Severity**: HIGH
- **Base Score**: 70
- **Score Adjustment**: Higher severity than Role because ClusterRoles apply cluster-wide
- **Remediation Key**: `rbac_cluster_wildcard`
- **Remediation**:
  - `kubectl describe clusterrole {name}`
  - Replace wildcard verbs with specific verbs following the principle of least privilege
  - Consider downscoping to a namespaced Role if cluster-wide access is not required

## Edge Cases

- ClusterRoles with `system:` prefix (managed by Kubernetes): skip always
- ClusterRole with specific verbs: safe, skip
- Wildcard in verbs combined with wildcard in resources: still reported (verb wildcard is the trigger)
- Built-in roles like `cluster-admin`: prefix is NOT `system:` but typically managed — still reported for security visibility

## Acceptance Criteria

- [x] Detects ClusterRoles with `"*"` in any rule's `Verbs`
- [x] Skips ClusterRoles with `system:` name prefix
- [x] ClusterRoles with specific verbs are not reported
- [x] Multiple wildcard ClusterRoles are all reported
- [x] Mixed ClusterRoles: only wildcard non-system ones are reported
- [x] Tests use fake K8s client
- [x] Zero false positives on system or least-privilege ClusterRoles
