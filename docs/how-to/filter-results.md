# How to Filter Scan Results

By default, `nautikube scan` checks *all* compatible resources across *all* namespaces. In large clusters, this can result in too much output. NautiKube offers powerful filters to narrow down your investigation.

## 1. Filter by Namespace

If you only care about a specific product namespace, use the `-n` or `--namespace` flag:

```bash
nautikube scan -n production-api
```

You can also scan system components to check cluster health:

```bash
nautikube scan -n kube-system
```

## 2. Filter by Resource Type

If you are just debugging networking and don't care about Pods, use the `-r` or `--resource` flag to scan only specific endpoints:

```bash
nautikube scan -r Service
```

You can pass multiple resources separated by commas (no spaces):

```bash
nautikube scan -r Pod,Deployment,Service
```

*Note: Resource names are case-sensitive and should match the official Kubernetes Kind (e.g., `Pod`, not `pod`).*

## 3. Filter by Minimum Severity

When you only have time to fix the "house on fire," filter by severity using `-s` or `--min-severity`:

```bash
nautikube scan -s high
```

This will hide `MEDIUM`, `LOW`, and `INFO` problems. The severity levels you can provide are:
- `critical`
- `high`
- `medium`
- `low`
- `info`

## 4. Combining Filters

You can mix any of the flags to create a laser-focused scan. For example, to find Critical and High severity issues in Pods within the `default` namespace:

```bash
nautikube scan -n default -r Pod -s high
```
