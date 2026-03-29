# Job Scanner Specification

## Objective

Detect Jobs with failed executions in a live Kubernetes cluster.

## Problems Detected

### 1. Job with Failed Executions

- **Condition**: `Status.Failed > 0`
- **Severity**: HIGH
- **Base Score**: 70
- **Remediation Key**: `job_failed`
- **Remediation**:
  - `kubectl describe job {name} -n {namespace}`
  - `kubectl get pods -n {namespace} -l job-name={name}`
  - `kubectl logs <failed-pod> -n {namespace} --previous`
  - Check backoffLimit and activeDeadlineSeconds settings

## Edge Cases

- Job with `Status.Failed == 0` (zero failures): healthy, skip
- Completed Job (`Status.Succeeded > 0`, `Status.Failed == 0`): healthy, skip
- Job with backoffLimit exhausted: still reported (Failed > 0)
- CronJob-managed Jobs: treated the same as standalone Jobs

## Acceptance Criteria

- [x] Detects Jobs with `Status.Failed > 0`
- [x] Jobs with zero failures are not reported
- [x] Jobs with only successes are not reported
- [x] Multiple failed Jobs are all reported
- [x] Mixed Jobs: only those with failures are reported
- [x] Jobs with a single failure are reported
- [x] Tests use fake K8s client
- [x] Zero false positives on successful Jobs
