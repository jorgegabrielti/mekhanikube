# CronJob Scanner Specification

## Objective

Detect CronJobs that have been suspended in a live Kubernetes cluster.

## Problems Detected

### 1. Suspended CronJob

- **Condition**: `Spec.Suspend != nil && *Spec.Suspend == true`
- **Severity**: LOW
- **Base Score**: 20
- **Remediation Key**: `cronjob_suspended`
- **Remediation**:
  - `kubectl describe cronjob {name} -n {namespace}`
  - Resume if the suspension was unintentional:
    `kubectl patch cronjob {name} -n {namespace} -p '{"spec": {"suspend": false}}'`

## Edge Cases

- CronJob with `Spec.Suspend == nil` (default): not suspended, skip
- CronJob with `Spec.Suspend == false`: active, skip
- CronJob suspended intentionally (e.g., maintenance): still reported for visibility

## Acceptance Criteria

- [x] Detects CronJobs with `Spec.Suspend == true`
- [x] CronJobs with `Spec.Suspend == false` are not reported
- [x] CronJobs with `Spec.Suspend == nil` are not reported
- [x] Multiple suspended CronJobs are all reported
- [x] Mixed active and suspended CronJobs: only suspended are reported
- [x] Tests use fake K8s client
- [x] Zero false positives on active CronJobs
