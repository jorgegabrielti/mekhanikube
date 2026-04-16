# Event Aggregator Specification

## Objective

Aggregate Warning events from the Kubernetes cluster to surface recurring issues.

## Behavior

Unlike other scanners that inspect resource state, the Event Aggregator reads Kubernetes
Events and groups them by resource + reason to identify patterns.

## Problems Detected

### 1. Frequent Warning Events

- **Condition**: A Warning event with `Count > threshold` for the same resource
- **Severity Tiers**:
  - `Count > 50`: HIGH (base score 70)
  - `Count > 20`: MEDIUM (base score 50)
  - `Count > 5`: LOW (base score 30)
- **Remediation Key**: `frequent-events`
- **Reported Fields**:
  - Resource: Event's `InvolvedObject.Kind`
  - Name: Event's `InvolvedObject.Name`
  - Namespace: Event's `InvolvedObject.Namespace`
  - Issue: `"{Reason}: {Message} (seen {Count} times)"`
- **Remediation**:
  - `kubectl describe {kind} {name} -n {namespace}`
  - `kubectl get events -n {namespace} --field-selector involvedObject.name={name}`

## Filtering

- Only process events of type `Warning` (skip `Normal`)
- Only process events from the last 1 hour (compare `LastTimestamp` with `time.Now()`)
- Skip events for resources already reported by other scanners (deduplication in orchestrator)

## Edge Cases

- Events with `Count == 0` or nil: skip
- Events with `LastTimestamp` zero: use `EventTime` or `CreationTimestamp` as fallback
- Large clusters with thousands of events: process only first 500 (use `Limit` in ListOptions)
- Events without `InvolvedObject`: skip

## Acceptance Criteria

- [x] Aggregates Warning events by resource and reason
- [x] Correctly assigns severity based on event count
- [x] Filters out events older than 1 hour
- [x] Tests use fake Events with various counts and timestamps
- [x] Handles missing timestamps gracefully
