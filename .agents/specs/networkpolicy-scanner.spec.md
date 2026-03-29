# NetworkPolicy Scanner Specification

## Objective

Detect overly permissive NetworkPolicy objects that allow all ingress traffic in a live Kubernetes cluster.

## Problems Detected

### 1. Allow-All Ingress Policy

- **Condition**: A NetworkPolicy has at least one ingress rule with no `From` selectors
  and no `Ports` restrictions (empty `IngressRule` object `{}`)
- **Severity**: HIGH
- **Base Score**: 70
- **Remediation Key**: `netpol_allow_all`
- **Remediation**:
  - `kubectl describe netpol {name} -n {namespace}`
  - Replace the empty ingress rule with specific `from` selectors (namespace, pod labels)
  - Restrict to known traffic sources to reduce the attack surface

## Edge Cases

- NetworkPolicy with no ingress rules (`ingress: []`): this is a **deny-all** policy — healthy, skip
- NetworkPolicy with ingress rules that have `From` selectors: restricted, skip
- NetworkPolicy with ingress rules that have `Ports` restrictions only: `Ports`-only rule with empty `From` still allows from all sources but on specific port — **reported** (From is empty)
- NetworkPolicy with egress rules only: not relevant for this check, skip
- Multiple ingress rules: if any single rule is empty, the policy is reported

## Acceptance Criteria

- [x] Detects NetworkPolicies with at least one empty ingress rule (`{}`)
- [x] NetworkPolicies with specific `From` selectors are not reported
- [x] NetworkPolicies with `Ports`-only restrictions (empty `From`) are reported
- [x] NetworkPolicies with no ingress rules (deny-all) are not reported
- [x] Multiple allow-all policies are all reported
- [x] Mixed policies: only allow-all ones are reported
- [x] Tests use fake K8s client
- [x] Zero false positives on correctly restricted NetworkPolicies
