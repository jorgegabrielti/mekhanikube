# PodDisruptionBudget Scanner Specification

## Objective

Detect PodDisruptionBudgets that currently allow zero disruptions, blocking voluntary disruptions like node drains and rolling updates in a live Kubernetes cluster.

## Problems Detected

### 1. PDB Blocking All Disruptions

- **Condition**: `Status.DisruptionsAllowed == 0`
- **Severity**: HIGH
- **Base Score**: 70
- **Remediation Key**: `pdb_no_disruptions`
- **Remediation**:
  - `kubectl describe pdb {name} -n {namespace}`
  - Check if the target Pods are all healthy and running
  - If the cluster is resource-constrained, add capacity or reschedule Pods
  - Review `minAvailable` / `maxUnavailable` settings relative to current replica count:
    `kubectl get pdb {name} -n {namespace} -o yaml`

## Edge Cases

- PDB with `DisruptionsAllowed > 0`: healthy, skip
- PDB with `DisruptionsAllowed == 0` due to insufficient healthy replicas: still reported
- PDB targeting a StatefulSet or Deployment with no running Pods: `DisruptionsAllowed` will be 0, reported

## Acceptance Criteria

- [x] Detects PDBs with `Status.DisruptionsAllowed == 0`
- [x] PDBs with `DisruptionsAllowed > 0` are not reported
- [x] PDBs with `DisruptionsAllowed == 1` are not reported
- [x] Multiple blocking PDBs are all reported
- [x] Mixed PDBs: only blocking ones are reported
- [x] Tests use fake K8s client
- [x] Zero false positives on PDBs allowing disruptions
