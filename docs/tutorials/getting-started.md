# Getting Started

Welcome to NautiKube! This tutorial will guide you through your very first cluster scan. 

We assume you already have a running Kubernetes cluster (like Minikube, Kind, Docker Desktop, or an EKS/GKE cluster) and `kubectl` correctly configured to access it.

## 1. Verify Installation

If you haven't installed NautiKube yet, follow the instructions on the [home page](../index.md).

Open your terminal and check that NautiKube is ready:

```bash
nautikube version
```

## 2. Your First Scan

NautiKube automatically uses your default Kubernetes config (usually `~/.kube/config`). To run a complete scan across your entire cluster, simply run:

```bash
nautikube scan
```

NautiKube will now query the Kubernetes API (read-only) across all namespaces. It will output a colorized table.

### Understanding the Output

Look at the table output. You will see columns like:
- **SEVERITY**: From `INFO` (blue) to `CRITICAL` (red).
- **RESOURCE**: The type of resource (e.g., `Pod`, `Service`).
- **NAMESPACE & NAME**: Exactly where the issue is.
- **ISSUE**: What NautiKube detected (e.g., `CrashLoopBackOff`).
- **SCORE**: A number from `0` to `100`. The higher the score, the more urgent the fix.
- **REMEDIATION**: The exact `kubectl` command you can run to investigate or fix it.

## 3. Dealing with a Broken Deployment

Let's intentionally break something to see NautiKube in action.

Create a deployment with an image that doesn't exist:

```bash
kubectl create deployment broken-nginx --image=nginx:nonexistent-version
```

Wait a few seconds for Kubernetes to try pulling the image. Then run NautiKube again, but this time let's filter just the `default` namespace to reduce noise:

```bash
nautikube scan -n default
```

You should see NautiKube catch this! 
- It will flag an `ErrImagePull` or `ImagePullBackOff` on the Pod.
- It will score it highly.
- The remediation command will tell you to `kubectl describe pod ...` to see the exact error.

## 4. Cleaning Up

You can remove the broken deployment we just made:

```bash
kubectl delete deployment broken-nginx
```

## Next Steps

Congratulations! You've used NautiKube to confidently assess the health of your cluster. 

- To learn how to integrate this into automated pipelines with JSON, see [CI/CD Integration](../how-to/ci-cd-integration.md).
- To learn how scores are calculated, read about the [Scoring System](../reference/scoring-system.md).
