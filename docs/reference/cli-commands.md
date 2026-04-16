# CLI Commands Reference

NautiKube is a single binary CLI application. It provides self-documenting help via `--help`.

## Global Commands

### `nautikube scan`

The core command that analyzes the Kubernetes cluster.

**Usage:**
```bash
nautikube scan [flags]
```

**Flags:**

| Flag | Shorthand | Default | Description |
|------|-----------|---------|-------------|
| `--namespace` | `-n` | `""` (all namespaces) | Scan a specific namespace |
| `--resource` | `-r` | `""` (all resources) | Filter by resource type (e.g., `Pod,Deployment,Service`) |
| `--min-severity` | `-s` | `""` (all severities) | Minimum severity to display (`critical`, `high`, `medium`, `low`, `info`) |
| `--output` | `-o` | `"table"` | Output format: `table`, `json`, `yaml` |
| `--no-color` | | `false` | Disable colored text output |
| `--kubeconfig` | | `~/.kube/config` | Explicit path to a kubeconfig file |
| `--context` | | current context | Explicitly define which Kubernetes context to use |

### `nautikube version`

Shows the current compiled version of NautiKube.

**Usage:**
```bash
nautikube version
```

## Environment Variables

While NautiKube is primarily driven by flags, the Kubernetes client respects standard environment variables:

- `KUBECONFIG`: Overrides the default kubeconfig path search order.
