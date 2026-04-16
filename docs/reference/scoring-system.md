# Scoring System Math

Unlike tools that yield a flat list of pass/fail checks, NautiKube processes every finding through a heuristic scoring engine that outputs a `0-100` numeric score.

## 1. Base Severity Weight

When a scanner detects a problem, it assigns an initial `Severity` enum. This translates to a base score:

| Severity | Base Score | Description |
|----------|------------|-------------|
| `CRITICAL` | 90 | Direct impact to availability (e.g., CrashLoopBackOff, OOMKilled) |
| `HIGH` | 70 | Immediate attention required (e.g., ImagePullBackOff, No Endpoints) |
| `MEDIUM` | 50 | Moderate problem, affects reliability (e.g., Unavailable Replicas) |
| `LOW` | 30 | Minor issue or configuration smell (e.g., Cordoned Node) |
| `INFO` | 10 | Informational finding, no action strictly required |

## 2. Contextual Modifiers

After the base score is assigned, the engine applies contextual modifiers. This is what makes NautiKube intelligent.

* **Namespace Modifier:** `+10 points`
  * If the issue occurs in `kube-system` or `default`. Small issues in core system components are elevated automatically.
  
* **Specific Pod Errors:** `+10 points`
  * If the resource is a `Pod` AND the issue is `CrashLoopBackOff`, `ImagePullBackOff`, or `OOMKilled`. These are the most notoriously noisy and impactful errors.

* **Service Endpoints:** `+10 points`
  * If the resource is a `Service` AND it has "no endpoints" in the issue string. A service without routing is essentially dead traffic.

## 3. Clamping

After all additions and subtractions, the final score is mathematically clamped between `0` and `100`:

```go
if score > 100 {
    score = 100
}
if score < 0 {
    score = 0
}
```

This guarantees uniform output metrics no matter how many negative modifiers stack on a single issue.
