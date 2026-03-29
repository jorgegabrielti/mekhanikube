# ResourceQuota Scanner Specification

## Objective

Detect ResourceQuotas that have been fully exhausted in a live Kubernetes cluster.

## Problems Detected

### 1. ResourceQuota Exhausted

- **Condition**: For any resource in `Status.Hard`, the corresponding value in `Status.Used`
  is greater than or equal to the hard limit
- **Severity**: HIGH
- **Base Score**: 70
- **Remediation Key**: `quota_reached`
- **Remediation**:
  - `kubectl describe resourcequota {name} -n {namespace}`
  - Identify which resource is exhausted (pods, CPU, memory, etc.)
  - Delete unused resources in the namespace or increase the quota:
    `kubectl edit resourcequota {name} -n {namespace}`

## Edge Cases

- ResourceQuota with available capacity: healthy, skip
- ResourceQuota with usage below hard limit by 1 unit: skip (not exhausted)
- Multiple resources within one ResourceQuota: reports the quota once if **any** resource is exhausted
- ResourceQuota with only limits (no `Status.Used` yet): skip if Used is not populated

## Acceptance Criteria

- [x] Detects ResourceQuotas where `Used >= Hard` for any resource key
- [x] ResourceQuotas with usage below hard limits are not reported
- [x] Exhausted CPU quota triggers a problem
- [x] Exhausted pod count quota triggers a problem
- [x] Multiple exhausted ResourceQuotas are all reported
- [x] Mixed ResourceQuotas: only exhausted ones are reported
- [x] Tests use fake K8s client
- [x] Zero false positives on ResourceQuotas with available capacity
