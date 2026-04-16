# Welcome to NautiKube

**NautiKube** is a read-only Kubernetes Cluster Diagnostic CLI that detects, prioritizes, and helps remediate issues in your live cluster.

## Why NautiKube?

Unlike tools that validate static YAML manifests, NautiKube talks directly to the Kubernetes API to find problems in your **running** cluster. It doesn't just list issues — it calculates a **Score (0-100)** to tell you exactly what is most critical, and provides the exact `kubectl` command to investigate or fix it.

- **Zero dependencies**: No Docker, no Python, no heavy frameworks. Just a single Go binary.
- **No external calls**: 100% local. No AI providers, no telemetry, no data leaves your machine.
- **Read-only**: Safe to run anywhere. It never alters your cluster state.
- **Instant results**: Written in Go, scanning hundreds of resources in milliseconds.

## Documentation Structure

This documentation uses the [Diátaxis](https://diataxis.fr/) framework, dividing content into four distinct quadrants based on what you need:

1. 🎓 **[Tutorials](tutorials/getting-started.md)**: Learning-oriented. Start here if you are new to NautiKube.
2. 🛠️ **[How-To Guides](how-to/filter-results.md)**: Problem-oriented. Recipes to accomplish specific tasks.
3. 📖 **[Reference](reference/cli-commands.md)**: Information-oriented. Exhaustive lists of commands and mathematical formulas.
4. 🧠 **[Concepts](concepts/spec-driven-dev.md)**: Understanding-oriented. Deeper explanations of NautiKube's architecture and design choices.

---

## Quick Install

Download the latest binary from GitHub Releases:

```bash
# Linux (amd64)
curl -sL https://github.com/jorgegabrielti/nautikube/releases/latest/download/nautikube_linux_amd64.tar.gz | tar xz
sudo mv nautikube /usr/local/bin/

# Verify
nautikube version
```

Ready to scan your first cluster? Head over to the **[Getting Started Tutorial](tutorials/getting-started.md)**.
