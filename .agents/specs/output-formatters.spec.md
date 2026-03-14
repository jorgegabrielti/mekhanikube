# Output Formatters Specification

## Objective

Format diagnostic results for display and automation. Three formats: table (default), JSON, YAML.

## Common Behavior

All formatters receive:
- `[]diagnosis.Problem` — sorted by Score (highest first)
- `Summary` — aggregate stats (total problems, count per severity)
- `io.Writer` — output destination (usually os.Stdout)

## Table Formatter (default)

### Output Structure

```
NautiKube Scan Results
======================

🔴 CRITICAL  Pod  default/web-app
   Score: 100/100
   Issue: Container nginx in CrashLoopBackOff
   Fix:   kubectl logs web-app -n default --previous
          kubectl describe pod web-app -n default

🟠 HIGH      Deployment  production/api-server
   Score: 70/100
   Issue: 2 unavailable replicas
   Fix:   kubectl describe deployment api-server -n production
          kubectl rollout status deployment/api-server -n production

─────────────────────────────────────
Summary: 5 problems found
  🔴 Critical: 1  🟠 High: 2  🟡 Medium: 1  🔵 Low: 1

```

### Colors (fatih/color)

| Severity | Color | Icon |
|---|---|---|
| CRITICAL | Red Bold | 🔴 |
| HIGH | Yellow | 🟠 |
| MEDIUM | Yellow | 🟡 |
| LOW | Blue | 🔵 |
| INFO | White | ⚪ |

### No-Color Mode

When `--no-color` flag is set or `NO_COLOR` env var exists, disable all ANSI colors.
Icons (emoji) remain — they work in all terminals.

### No Problems

When no problems are detected, output:

```
NautiKube Scan Results
======================

✅ No problems found. Your cluster looks healthy!
```

## JSON Formatter

```json
{
  "timestamp": "2026-03-13T22:00:00Z",
  "problems": [
    {
      "resource": "Pod",
      "namespace": "default",
      "name": "web-app",
      "issue": "Container nginx in CrashLoopBackOff",
      "severity": "CRITICAL",
      "score": 100,
      "remediation": [
        "kubectl logs web-app -n default --previous",
        "kubectl describe pod web-app -n default"
      ]
    }
  ],
  "summary": {
    "total": 5,
    "critical": 1,
    "high": 2,
    "medium": 1,
    "low": 1,
    "info": 0
  }
}
```

## YAML Formatter

```yaml
timestamp: "2026-03-13T22:00:00Z"
problems:
  - resource: Pod
    namespace: default
    name: web-app
    issue: "Container nginx in CrashLoopBackOff"
    severity: CRITICAL
    score: 100
    remediation:
      - "kubectl logs web-app -n default --previous"
      - "kubectl describe pod web-app -n default"
summary:
  total: 5
  critical: 1
  high: 2
  medium: 1
  low: 1
  info: 0
```

## Acceptance Criteria

- [ ] Table formatter produces colored, readable output
- [ ] Table respects `--no-color` flag and `NO_COLOR` env var
- [ ] JSON output is valid JSON (verify with `jq`)
- [ ] YAML output is valid YAML
- [ ] All formatters sort problems by Score descending
- [ ] All formatters include summary statistics
- [ ] Empty results produce a clean "no problems" message
- [ ] Tests validate output against expected strings
