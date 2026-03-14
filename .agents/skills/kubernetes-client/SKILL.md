---
name: kubernetes-client
description: How to use client-go and fake clients in NautiKube
---

# Kubernetes Client Skill

Reference for working with the Kubernetes API in NautiKube using `client-go`.

## Client Factory

The K8s client is created in `internal/k8s/client.go` using functional options:

```go
client, err := k8s.New(
    k8s.WithKubeconfig("/path/to/config"),
    k8s.WithContext("staging"),
)
```

The client factory tries 4 connection strategies in order:
1. In-cluster config (`rest.InClusterConfig()`)
2. Explicit kubeconfig path (from `--kubeconfig` flag)
3. `KUBECONFIG` environment variable
4. Default `~/.kube/config`

It returns `kubernetes.Interface`, not the concrete `*kubernetes.Clientset`, for testability.

## Common API Calls

### Listing Pods

```go
pods, err := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
if err != nil {
    return nil, fmt.Errorf("failed to list pods: %w", err)
}
for _, pod := range pods.Items {
    // Inspect pod.Status.Phase, pod.Status.ContainerStatuses, etc.
}
```

### Listing Deployments

```go
deps, err := client.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
```

### Listing Services + Endpoints

```go
svcs, err := client.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
eps, err := client.CoreV1().Endpoints(namespace).List(ctx, metav1.ListOptions{})
```

### Listing Nodes

```go
nodes, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
for _, node := range nodes.Items {
    for _, condition := range node.Status.Conditions {
        // Check condition.Type and condition.Status
    }
}
```

### Listing Events (Warning only, last 1h)

```go
events, err := client.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{
    FieldSelector: "type=Warning",
})
```

## Fake Client for Tests

Use `k8s.io/client-go/kubernetes/fake` to create a mock Kubernetes API:

```go
import (
    "k8s.io/client-go/kubernetes/fake"
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestPodScanner(t *testing.T) {
    // Create fake pods
    crashingPod := &corev1.Pod{
        ObjectMeta: metav1.ObjectMeta{
            Name:      "web-app",
            Namespace: "default",
        },
        Status: corev1.PodStatus{
            Phase: corev1.PodRunning,
            ContainerStatuses: []corev1.ContainerStatus{
                {
                    Name: "nginx",
                    State: corev1.ContainerState{
                        Waiting: &corev1.ContainerStateWaiting{
                            Reason: "CrashLoopBackOff",
                        },
                    },
                    RestartCount: 15,
                },
            },
        },
    }

    // Create fake client with the pod
    client := fake.NewSimpleClientset(crashingPod)

    // Use the client in your scanner
    scanner := NewPodScanner()
    problems, err := scanner.Scan(context.Background(), client, "")
    // assert...
}
```

### Creating Fake Objects for Each Resource Type

**Pod states to test:**
- Waiting: CrashLoopBackOff, ImagePullBackOff, ErrImagePull
- Terminated: OOMKilled, Error
- Phase: Pending (with no container statuses)
- High RestartCount (>5, >20, >50)
- Not Ready (Ready condition False)

**Deployment states to test:**
- `Status.UnavailableReplicas > 0`
- `Status.ReadyReplicas < Spec.Replicas`
- `Spec.Replicas == 0`

**Node conditions to test:**
- `Ready: False`
- `MemoryPressure: True`
- `DiskPressure: True`
- `PIDPressure: True`
- `Spec.Unschedulable: true`

**Service states to test:**
- Service exists but no matching Endpoints
- LoadBalancer with empty `Status.LoadBalancer.Ingress`

## Important: Namespace Handling

When `namespace` is empty string `""`, the Kubernetes API returns resources from **all namespaces**. This is the default behavior in NautiKube (`nautikube scan` without `-n` flag).

## RBAC Considerations

Scanners should handle permission errors gracefully:

```go
pods, err := client.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{})
if err != nil {
    if apierrors.IsForbidden(err) {
        slog.Warn("insufficient permissions to list pods", "namespace", ns)
        return nil, nil // Skip this resource type, don't fail
    }
    return nil, fmt.Errorf("failed to list pods: %w", err)
}
```
