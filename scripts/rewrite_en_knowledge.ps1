$knowledgeDir = Join-Path $PSScriptRoot "..\internal\diagnosis\knowledge"

# --- Pod Issues ---

Set-Content -Path (Join-Path $knowledgeDir "crashloopbackoff.en.yaml") -Encoding UTF8 -Value @'
key: "crashloopbackoff"
title: "Container in CrashLoopBackOff"
explanation: |
  The container crashes immediately after starting and Kubernetes restarts it with exponential
  backoff (10s -> 20s -> 40s -> ... up to 5 min between attempts).

  Root-cause by exit code:
    - Exit 0:   Process finished but was not meant to — wrong entrypoint or one-shot command in a long-running pod.
    - Exit 1/2: Application error — bad config, missing env var, unhandled exception on startup.
    - Exit 127: Binary not found inside the image — wrong CMD/ENTRYPOINT or missing dependency.
    - Exit 137: OOMKilled — the container exceeded its memory limit (resources.limits.memory).

  Resolution flow:
    1. Read the PREVIOUS container logs (current instance may not have output yet).
    2. Check the exit code to classify the failure.
    3. Inspect env vars, volume mounts, and probe settings in the pod spec.
    4. Apply the targeted fix (image, config, resources) and verify the pod stabilizes.
commands:
  - "kubectl logs {pod} -n {namespace} --previous --tail=100"
  - "kubectl get pod {pod} -n {namespace} -o jsonpath='{.status.containerStatuses[0].lastState.terminated.exitCode}'"
  - "kubectl describe pod {pod} -n {namespace} | grep -A5 'Environment\\|Mounts\\|Events'"
  - "kubectl set resources deployment/{name} -n {namespace} --limits=memory=512Mi --requests=memory=256Mi"
  - "kubectl rollout restart deployment/{name} -n {namespace}"
  - "kubectl rollout status deployment/{name} -n {namespace} --timeout=120s"
windows_commands:
  - "kubectl logs {pod} -n {namespace} --previous --tail=100"
  - "kubectl get pod {pod} -n {namespace} -o jsonpath='{.status.containerStatuses[0].lastState.terminated.exitCode}'"
  - "kubectl describe pod {pod} -n {namespace}"
  - "kubectl set resources deployment/{name} -n {namespace} --limits=memory=512Mi --requests=memory=256Mi"
  - "kubectl rollout restart deployment/{name} -n {namespace}"
  - "kubectl rollout status deployment/{name} -n {namespace} --timeout=120s"
'@

Set-Content -Path (Join-Path $knowledgeDir "oomkilled.en.yaml") -Encoding UTF8 -Value @'
key: "oomkilled"
title: "Container OOMKilled"
explanation: |
  The container exceeded its memory limit (resources.limits.memory) and was killed by the Linux
  kernel OOM killer (exit code 137). The pod will restart, but will keep being killed if the
  limit is not increased or the memory leak is not fixed.

  Resolution flow:
    1. Confirm the OOMKill reason in the pod's last state.
    2. Check current memory usage vs. configured limits.
    3. Increase the memory limit to a safe value (start with 2x current limit).
    4. If the problem recurs, the application has a memory leak — profile the application.
    5. Verify the pod runs stably after the change.
commands:
  - "kubectl get pod {pod} -n {namespace} -o jsonpath='{.status.containerStatuses[0].lastState.terminated}'"
  - "kubectl top pod {pod} -n {namespace}"
  - "kubectl get pod {pod} -n {namespace} -o jsonpath='{.spec.containers[0].resources}'"
  - "kubectl set resources deployment/{name} -n {namespace} --limits=memory=1Gi --requests=memory=512Mi"
  - "kubectl rollout status deployment/{name} -n {namespace} --timeout=120s"
  - "kubectl top pod -n {namespace} -l app={name} --sort-by=memory"
windows_commands:
  - "kubectl get pod {pod} -n {namespace} -o jsonpath='{.status.containerStatuses[0].lastState.terminated}'"
  - "kubectl top pod {pod} -n {namespace}"
  - "kubectl get pod {pod} -n {namespace} -o jsonpath='{.spec.containers[0].resources}'"
  - "kubectl set resources deployment/{name} -n {namespace} --limits=memory=1Gi --requests=memory=512Mi"
  - "kubectl rollout status deployment/{name} -n {namespace} --timeout=120s"
  - "kubectl top pod -n {namespace} -l app={name} --sort-by=memory"
'@

Set-Content -Path (Join-Path $knowledgeDir "pod_pending.en.yaml") -Encoding UTF8 -Value @'
key: "pod-pending"
title: "Pod Stuck in Pending State"
explanation: |
  The scheduler cannot place this pod on any node. It will remain Pending until the constraint
  is resolved.

  Common causes and fixes:
    1. Insufficient CPU/Memory — no node has the requested capacity. Scale up the cluster or
       reduce the pod's resource requests.
    2. NodeSelector/Affinity — no node matches the required labels. Add the label to a node
       or relax the affinity rules.
    3. Taints — all candidate nodes have taints the pod does not tolerate. Remove the taint
       or add a toleration to the pod spec.
    4. PVC not bound — a volume claim cannot be satisfied. Check StorageClass and PV availability.

  Resolution flow:
    1. Read the scheduler events to identify the exact rejection reason.
    2. Apply the targeted fix based on the cause.
    3. Verify the pod transitions to Running.
commands:
  - "kubectl describe pod {pod} -n {namespace} | grep -A10 'Events:'"
  - "kubectl get nodes -o custom-columns=NAME:.metadata.name,CPU:.status.allocatable.cpu,MEM:.status.allocatable.memory,TAINTS:.spec.taints[*].key"
  - "kubectl get pvc -n {namespace} -o wide"
  - "kubectl set resources deployment/{name} -n {namespace} --requests=cpu=100m,memory=128Mi"
  - "kubectl label node <node-name> <key>=<value>"
  - "kubectl taint nodes <node-name> <key>:NoSchedule-"
  - "kubectl get pod {pod} -n {namespace} -o wide"
windows_commands:
  - "kubectl describe pod {pod} -n {namespace}"
  - "kubectl get nodes -o custom-columns=NAME:.metadata.name,CPU:.status.allocatable.cpu,MEM:.status.allocatable.memory,TAINTS:.spec.taints[*].key"
  - "kubectl get pvc -n {namespace} -o wide"
  - "kubectl set resources deployment/{name} -n {namespace} --requests=cpu=100m,memory=128Mi"
  - "kubectl label node <node-name> <key>=<value>"
  - "kubectl taint nodes <node-name> <key>:NoSchedule-"
  - "kubectl get pod {pod} -n {namespace} -o wide"
'@

Set-Content -Path (Join-Path $knowledgeDir "high_restarts.en.yaml") -Encoding UTF8 -Value @'
key: "high-restarts"
title: "High Container Restart Count"
explanation: |
  A container in this pod has restarted an abnormally high number of times, indicating a
  recurring failure that Kubernetes keeps recovering from.

  Root-cause by exit code:
    - Exit 137 (OOMKilled): Increase resources.limits.memory.
    - Exit 1/2 (App crash): Fix application config, env vars, or startup logic.
    - Exit 0 (Unexpected exit): Fix entrypoint — the process should not terminate.
    - Liveness probe failure: The probe is too aggressive or the endpoint is too slow.

  Resolution flow:
    1. Read logs from the PREVIOUS crashed instance.
    2. Check the exit code and last state to classify the failure type.
    3. Apply the targeted fix and restart the deployment.
    4. Monitor restart count to verify it stabilizes at zero.
commands:
  - "kubectl logs {pod} -n {namespace} --previous --tail=100"
  - "kubectl get pod {pod} -n {namespace} -o jsonpath='{.status.containerStatuses[0].lastState.terminated}'"
  - "kubectl describe pod {pod} -n {namespace} | grep -A5 'Liveness\\|Readiness'"
  - "kubectl set resources deployment/{name} -n {namespace} --limits=memory=512Mi --requests=memory=256Mi"
  - "kubectl rollout restart deployment/{name} -n {namespace}"
  - "kubectl get pod -n {namespace} -l app={name} -w"
windows_commands:
  - "kubectl logs {pod} -n {namespace} --previous --tail=100"
  - "kubectl get pod {pod} -n {namespace} -o jsonpath='{.status.containerStatuses[0].lastState.terminated}'"
  - "kubectl describe pod {pod} -n {namespace}"
  - "kubectl set resources deployment/{name} -n {namespace} --limits=memory=512Mi --requests=memory=256Mi"
  - "kubectl rollout restart deployment/{name} -n {namespace}"
  - "kubectl get pod -n {namespace} -l app={name} -w"
'@

Set-Content -Path (Join-Path $knowledgeDir "container_not_ready.en.yaml") -Encoding UTF8 -Value @'
key: "container-not-ready"
title: "Container Not Ready (Readiness Probe Failed)"
explanation: |
  The container is running but its readiness probe keeps failing. Kubernetes removes it from
  all Service endpoints, so no traffic is routed to this pod.

  Common causes:
    - The application is deadlocked or stuck in initialization.
    - The readiness probe path/port is misconfigured.
    - initialDelaySeconds is too short for the app to start.
    - The health endpoint returns 5xx under load.

  Resolution flow:
    1. Check which probe is failing and its exact configuration.
    2. Read the application logs to see if it started correctly.
    3. Adjust the probe timing (increase initialDelaySeconds, periodSeconds, or timeoutSeconds).
    4. Verify the pod becomes Ready and appears in Service endpoints.
commands:
  - "kubectl describe pod {pod} -n {namespace} | grep -A3 'Readiness:'"
  - "kubectl logs {pod} -n {namespace} --tail=100"
  - "kubectl get events -n {namespace} --field-selector involvedObject.name={pod},reason=Unhealthy --sort-by='.lastTimestamp'"
  - "kubectl patch deployment/{name} -n {namespace} --type=strategic -p '{\"spec\":{\"template\":{\"spec\":{\"containers\":[{\"name\":\"{name}\",\"readinessProbe\":{\"initialDelaySeconds\":30,\"periodSeconds\":10,\"timeoutSeconds\":5,\"failureThreshold\":5}}]}}}}'"
  - "kubectl rollout status deployment/{name} -n {namespace} --timeout=120s"
  - "kubectl get endpoints {name} -n {namespace}"
windows_commands:
  - "kubectl describe pod {pod} -n {namespace}"
  - "kubectl logs {pod} -n {namespace} --tail=100"
  - "kubectl get events -n {namespace} --field-selector involvedObject.name={pod},reason=Unhealthy --sort-by='.lastTimestamp'"
  - "kubectl patch deployment/{name} -n {namespace} --type=strategic -p '{\"spec\":{\"template\":{\"spec\":{\"containers\":[{\"name\":\"{name}\",\"readinessProbe\":{\"initialDelaySeconds\":30,\"periodSeconds\":10,\"timeoutSeconds\":5,\"failureThreshold\":5}}]}}}}'"
  - "kubectl rollout status deployment/{name} -n {namespace} --timeout=120s"
  - "kubectl get endpoints {name} -n {namespace}"
'@

Set-Content -Path (Join-Path $knowledgeDir "createcontainerconfigerror.en.yaml") -Encoding UTF8 -Value @'
key: "createcontainerconfigerror"
title: "CreateContainerConfigError"
explanation: |
  Kubernetes cannot create the container because a referenced ConfigMap, Secret, or
  ServiceAccount does not exist in the namespace, or a specific key within the resource
  does not match the volume mount or env var reference.

  Resolution flow:
    1. Read the pod events to identify the exact missing resource name and key.
    2. Check if the ConfigMap/Secret exists in the namespace.
    3. Create the missing resource or fix the reference in the pod spec.
    4. The pod will auto-retry — verify it transitions to Running.
commands:
  - "kubectl describe pod {pod} -n {namespace} | grep -A10 'Events:'"
  - "kubectl get configmap -n {namespace} -o name"
  - "kubectl get secret -n {namespace} -o name"
  - "kubectl create configmap <cm-name> -n {namespace} --from-literal=<key>=<value>"
  - "kubectl create secret generic <secret-name> -n {namespace} --from-literal=<key>=<value>"
  - "kubectl get pod {pod} -n {namespace} -w"
windows_commands:
  - "kubectl describe pod {pod} -n {namespace}"
  - "kubectl get configmap -n {namespace} -o name"
  - "kubectl get secret -n {namespace} -o name"
  - "kubectl create configmap <cm-name> -n {namespace} --from-literal=<key>=<value>"
  - "kubectl create secret generic <secret-name> -n {namespace} --from-literal=<key>=<value>"
  - "kubectl get pod {pod} -n {namespace} -w"
'@

Set-Content -Path (Join-Path $knowledgeDir "createcontainererror.en.yaml") -Encoding UTF8 -Value @'
key: "createcontainererror"
title: "CreateContainerError"
explanation: |
  The container runtime (containerd/CRI-O) failed to create the container. ConfigMaps and
  Secrets exist, but something at the runtime level went wrong.

  Common causes:
    - Invalid command or entrypoint specified in the container spec.
    - Permission denied when mounting volumes (SELinux, AppArmor, read-only filesystem).
    - PVC mount failed at the node level (disk not attached, NFS unreachable).
    - SecurityContext conflicts (runAsNonRoot but image runs as root).

  Resolution flow:
    1. Read the events to get the CRI error message.
    2. Fix the container spec (image, command, securityContext, volumeMounts).
    3. Redeploy and verify the container starts successfully.
commands:
  - "kubectl describe pod {pod} -n {namespace} | grep -A10 'Events:'"
  - "kubectl get pod {pod} -n {namespace} -o jsonpath='{.spec.containers[0].command}'"
  - "kubectl get pod {pod} -n {namespace} -o jsonpath='{.spec.containers[0].securityContext}'"
  - "kubectl patch deployment/{name} -n {namespace} --type=strategic -p '{\"spec\":{\"template\":{\"spec\":{\"containers\":[{\"name\":\"{name}\",\"securityContext\":{\"runAsNonRoot\":false}}]}}}}'"
  - "kubectl rollout restart deployment/{name} -n {namespace}"
  - "kubectl rollout status deployment/{name} -n {namespace} --timeout=120s"
windows_commands:
  - "kubectl describe pod {pod} -n {namespace}"
  - "kubectl get pod {pod} -n {namespace} -o jsonpath='{.spec.containers[0].command}'"
  - "kubectl get pod {pod} -n {namespace} -o jsonpath='{.spec.containers[0].securityContext}'"
  - "kubectl patch deployment/{name} -n {namespace} --type=strategic -p '{\"spec\":{\"template\":{\"spec\":{\"containers\":[{\"name\":\"{name}\",\"securityContext\":{\"runAsNonRoot\":false}}]}}}}'"
  - "kubectl rollout restart deployment/{name} -n {namespace}"
  - "kubectl rollout status deployment/{name} -n {namespace} --timeout=120s"
'@

Set-Content -Path (Join-Path $knowledgeDir "imagepullbackoff.en.yaml") -Encoding UTF8 -Value @'
key: "imagepullbackoff"
title: "Image Pull Failure (ImagePullBackOff)"
explanation: |
  The kubelet cannot pull the container image from the registry.

  Common causes and fixes:
    - Image not found: The image name or tag is misspelled. Fix the image reference.
    - Unauthorized: The registry requires authentication. Create an imagePullSecret.
    - Registry unreachable: The node cannot reach the registry. Check network/firewall.
    - Rate limited: Docker Hub rate limit hit. Use a pull-through cache or authenticate.

  Resolution flow:
    1. Read the pod events to see the exact registry error (unauthorized, not found, timeout).
    2. Verify the image name and tag are correct.
    3. If auth issue, create a docker-registry secret and patch the ServiceAccount.
    4. If wrong image, fix the deployment image reference.
    5. Verify the pod pulls the image and starts.
commands:
  - "kubectl describe pod {pod} -n {namespace} | grep -A5 'Events:'"
  - "kubectl get pod {pod} -n {namespace} -o jsonpath='{.spec.containers[0].image}'"
  - "kubectl create secret docker-registry regcred -n {namespace} --docker-server=<registry> --docker-username=<user> --docker-password=<pass>"
  - "kubectl patch serviceaccount default -n {namespace} -p '{\"imagePullSecrets\":[{\"name\":\"regcred\"}]}'"
  - "kubectl set image deployment/{name} -n {namespace} {name}=<correct-image:tag>"
  - "kubectl rollout status deployment/{name} -n {namespace} --timeout=120s"
windows_commands:
  - "kubectl describe pod {pod} -n {namespace}"
  - "kubectl get pod {pod} -n {namespace} -o jsonpath='{.spec.containers[0].image}'"
  - "kubectl create secret docker-registry regcred -n {namespace} --docker-server=<registry> --docker-username=<user> --docker-password=<pass>"
  - "kubectl patch serviceaccount default -n {namespace} -p '{\"imagePullSecrets\":[{\"name\":\"regcred\"}]}'"
  - "kubectl set image deployment/{name} -n {namespace} {name}=<correct-image:tag>"
  - "kubectl rollout status deployment/{name} -n {namespace} --timeout=120s"
'@

# --- Deployment/ReplicaSet Issues ---

Set-Content -Path (Join-Path $knowledgeDir "deployment_stuck.en.yaml") -Encoding UTF8 -Value @'
key: "deployment-stuck"
title: "Deployment Rollout Stuck"
explanation: |
  The Deployment rollout is not progressing — the new ReplicaSet cannot scale up its pods.
  Kubernetes marks it as stuck after the progressDeadlineSeconds (default 600s) is exceeded.

  Common causes:
    - New pods crash on start (CrashLoopBackOff, ImagePullBackOff).
    - Insufficient cluster resources to schedule new pods.
    - Readiness probes fail on the new pods.

  Resolution flow:
    1. Check the rollout status and identify why the new ReplicaSet is stuck.
    2. Inspect pods from the new ReplicaSet for specific errors.
    3. If the new version is broken, rollback to the last working revision.
    4. If fixable, correct the issue and let the rollout resume.
    5. Verify the deployment is fully available.
commands:
  - "kubectl rollout status deployment/{name} -n {namespace} --watch=false"
  - "kubectl get rs -n {namespace} -l app={name} --sort-by='.metadata.creationTimestamp'"
  - "kubectl describe deployment {name} -n {namespace} | grep -A10 'Conditions:'"
  - "kubectl rollout undo deployment/{name} -n {namespace}"
  - "kubectl rollout status deployment/{name} -n {namespace} --timeout=120s"
  - "kubectl get pods -n {namespace} -l app={name} -o wide"
windows_commands:
  - "kubectl rollout status deployment/{name} -n {namespace} --watch=false"
  - "kubectl get rs -n {namespace} -l app={name} --sort-by='.metadata.creationTimestamp'"
  - "kubectl describe deployment {name} -n {namespace}"
  - "kubectl rollout undo deployment/{name} -n {namespace}"
  - "kubectl rollout status deployment/{name} -n {namespace} --timeout=120s"
  - "kubectl get pods -n {namespace} -l app={name} -o wide"
'@

Set-Content -Path (Join-Path $knowledgeDir "deployment_mismatch.en.yaml") -Encoding UTF8 -Value @'
key: "deployment-mismatch"
title: "Deployment Replicas Mismatch"
explanation: |
  The number of ready replicas does not match the desired count. Some pods are failing to
  become ready (CrashLoop, Pending, ImagePullBackOff, or failing readiness probes).

  Resolution flow:
    1. Check the rollout status — if a rollout is in progress, the mismatch may be transient.
    2. List pods to identify which ones are failing and why.
    3. Fix the underlying issue (resources, image, config) or rollback.
    4. If pods are healthy but stuck, force a rollout restart.
    5. Verify all replicas become ready.
commands:
  - "kubectl rollout status deployment/{name} -n {namespace} --watch=false"
  - "kubectl get pods -n {namespace} -l app={name} -o wide"
  - "kubectl describe deployment {name} -n {namespace} | grep -A10 'Conditions:'"
  - "kubectl rollout restart deployment/{name} -n {namespace}"
  - "kubectl rollout status deployment/{name} -n {namespace} --timeout=120s"
  - "kubectl get deployment {name} -n {namespace}"
windows_commands:
  - "kubectl rollout status deployment/{name} -n {namespace} --watch=false"
  - "kubectl get pods -n {namespace} -l app={name} -o wide"
  - "kubectl describe deployment {name} -n {namespace}"
  - "kubectl rollout restart deployment/{name} -n {namespace}"
  - "kubectl rollout status deployment/{name} -n {namespace} --timeout=120s"
  - "kubectl get deployment {name} -n {namespace}"
'@

Set-Content -Path (Join-Path $knowledgeDir "deployment_unavailable.en.yaml") -Encoding UTF8 -Value @'
key: "deployment-unavailable"
title: "Deployment Has Unavailable Replicas"
explanation: |
  One or more replicas are not available (status.unavailableReplicas > 0). Pods are failing
  to start, crashing, failing probes, or stuck Pending.

  Resolution flow:
    1. Check which pods are not ready and identify their specific error.
    2. If a bad rollout, undo to the last known good revision.
    3. If a transient issue, restart the deployment to recreate pods.
    4. Verify all replicas become available.
commands:
  - "kubectl get pods -n {namespace} -l app={name} -o wide"
  - "kubectl describe deployment {name} -n {namespace} | grep -A10 'Conditions:'"
  - "kubectl get events -n {namespace} --field-selector involvedObject.name={name} --sort-by='.lastTimestamp'"
  - "kubectl rollout undo deployment/{name} -n {namespace}"
  - "kubectl rollout status deployment/{name} -n {namespace} --timeout=120s"
  - "kubectl get deployment {name} -n {namespace}"
windows_commands:
  - "kubectl get pods -n {namespace} -l app={name} -o wide"
  - "kubectl describe deployment {name} -n {namespace}"
  - "kubectl get events -n {namespace} --field-selector involvedObject.name={name} --sort-by='.lastTimestamp'"
  - "kubectl rollout undo deployment/{name} -n {namespace}"
  - "kubectl rollout status deployment/{name} -n {namespace} --timeout=120s"
  - "kubectl get deployment {name} -n {namespace}"
'@

Set-Content -Path (Join-Path $knowledgeDir "deployment_zero_replicas.en.yaml") -Encoding UTF8 -Value @'
key: "deployment-zero-replicas"
title: "Deployment Scaled to Zero"
explanation: |
  This deployment has zero replicas — no pods are running. This could be intentional
  (maintenance) or caused by an autoscaler (HPA/KEDA) scaling down, or human error.

  Resolution flow:
    1. Check if an HPA or KEDA ScaledObject is controlling this deployment.
    2. If intentional, no action needed.
    3. If accidental, scale back up to the desired replica count.
    4. Verify pods start successfully.
commands:
  - "kubectl get hpa -n {namespace} | grep {name}"
  - "kubectl get scaledobject -n {namespace} 2>/dev/null | grep {name}"
  - "kubectl describe deployment {name} -n {namespace} | grep -E 'Replicas|Annotations'"
  - "kubectl scale deployment/{name} -n {namespace} --replicas=2"
  - "kubectl rollout status deployment/{name} -n {namespace} --timeout=120s"
  - "kubectl get pods -n {namespace} -l app={name}"
windows_commands:
  - "kubectl get hpa -n {namespace}"
  - "kubectl get scaledobject -n {namespace} 2>$null"
  - "kubectl describe deployment {name} -n {namespace}"
  - "kubectl scale deployment/{name} -n {namespace} --replicas=2"
  - "kubectl rollout status deployment/{name} -n {namespace} --timeout=120s"
  - "kubectl get pods -n {namespace} -l app={name}"
'@

Set-Content -Path (Join-Path $knowledgeDir "replicaset_mismatch.en.yaml") -Encoding UTF8 -Value @'
key: replicaset_mismatch
title: "ReplicaSet Replicas Mismatch"
explanation: |
  The ReplicaSet has fewer ready pods than desired. This is the underlying cause for
  Deployment unavailability.

  Resolution flow:
    1. Identify which pods owned by this ReplicaSet are failing.
    2. Check those pods for CrashLoopBackOff, ImagePull errors, or Pending state.
    3. Fix the root cause in the Deployment spec and let it propagate.
    4. If urgent, restart the Deployment to force pod recreation.
    5. Verify all replicas are ready.
commands:
  - "kubectl describe rs {name} -n {namespace}"
  - "kubectl get pods -n {namespace} --selector=app={name} -o wide"
  - "kubectl logs -l app={name} -n {namespace} --tail=50"
  - "kubectl rollout restart deployment/{name} -n {namespace}"
  - "kubectl rollout status deployment/{name} -n {namespace} --timeout=120s"
  - "kubectl get rs {name} -n {namespace}"
windows_commands:
  - "kubectl describe rs {name} -n {namespace}"
  - "kubectl get pods -n {namespace} --selector=app={name} -o wide"
  - "kubectl logs -l app={name} -n {namespace} --tail=50"
  - "kubectl rollout restart deployment/{name} -n {namespace}"
  - "kubectl rollout status deployment/{name} -n {namespace} --timeout=120s"
  - "kubectl get rs {name} -n {namespace}"
'@

Set-Content -Path (Join-Path $knowledgeDir "replicaset_orphaned.en.yaml") -Encoding UTF8 -Value @'
key: replicaset_orphaned
title: "Orphaned ReplicaSet Detected"
explanation: |
  This ReplicaSet has no ownerReferences — it is not managed by any Deployment. It was either
  created manually or left behind after a Deployment was deleted with --cascade=orphan.

  Orphaned ReplicaSets consume resources (CPU, memory, IPs) with no lifecycle management —
  no rollouts, no self-healing, no rollback capability.

  Decision tree:
    - Serving production traffic? -> Wrap it in a new Deployment for proper management.
    - Leftover from deleted Deployment? -> Verify pods are not in use, then delete it.
    - Intentional test workload? -> Accept the risk or migrate to a Deployment.

  Resolution flow:
    1. Confirm it has no owner references.
    2. Check if its pods are handling traffic (check Service endpoints).
    3. Delete if unused, or recreate as a Deployment for proper management.
commands:
  - "kubectl get rs {name} -n {namespace} -o jsonpath='{.metadata.ownerReferences}'"
  - "kubectl get pods -n {namespace} --selector=app={name} -o wide"
  - "kubectl get endpoints -n {namespace} | grep {name}"
  - "kubectl delete replicaset {name} -n {namespace}"
  - "kubectl get rs -n {namespace}"
windows_commands:
  - "kubectl get rs {name} -n {namespace} -o jsonpath='{.metadata.ownerReferences}'"
  - "kubectl get pods -n {namespace} --selector=app={name} -o wide"
  - "kubectl get endpoints -n {namespace}"
  - "kubectl delete replicaset {name} -n {namespace}"
  - "kubectl get rs -n {namespace}"
'@

Set-Content -Path (Join-Path $knowledgeDir "statefulset_mismatch.en.yaml") -Encoding UTF8 -Value @'
key: statefulset_mismatch
title: "StatefulSet Replicas Mismatch"
explanation: |
  The StatefulSet has fewer ready replicas than desired. StatefulSets create pods sequentially
  (pod-0, pod-1, ...) and will not proceed to the next pod if the current one is not Ready.

  Common causes:
    - PVCs cannot be bound (missing StorageClass or no available PVs).
    - Pod initialization is failing (init containers, app startup issues).
    - The headless Service required for DNS is misconfigured.

  Resolution flow:
    1. Check the StatefulSet events and identify which pod is stuck.
    2. Inspect that specific pod for PVC, init container, or application errors.
    3. Fix the issue (create PV, fix StorageClass, fix app config).
    4. If a pod is stuck and cannot recover, delete it to trigger recreation.
    5. Verify all replicas become ready.
commands:
  - "kubectl describe sts {name} -n {namespace}"
  - "kubectl get pods -n {namespace} -l app={name} -o wide"
  - "kubectl get pvc -n {namespace} -l app={name}"
  - "kubectl logs {name}-0 -n {namespace} --tail=100"
  - "kubectl delete pod {name}-0 -n {namespace}"
  - "kubectl rollout status statefulset/{name} -n {namespace} --timeout=180s"
windows_commands:
  - "kubectl describe sts {name} -n {namespace}"
  - "kubectl get pods -n {namespace} -l app={name} -o wide"
  - "kubectl get pvc -n {namespace} -l app={name}"
  - "kubectl logs {name}-0 -n {namespace} --tail=100"
  - "kubectl delete pod {name}-0 -n {namespace}"
  - "kubectl rollout status statefulset/{name} -n {namespace} --timeout=180s"
'@

# --- Node Issues ---

Set-Content -Path (Join-Path $knowledgeDir "node_not_ready.en.yaml") -Encoding UTF8 -Value @'
key: "node-not-ready"
title: "Node Not Ready"
explanation: |
  The node has stopped reporting health to the control plane. It cannot schedule new pods,
  and existing pods will be evicted after the pod-eviction-timeout (default 5 minutes).

  Common causes:
    - The kubelet process crashed or is unresponsive.
    - Network partition between the node and the API server.
    - The underlying VM/machine crashed or ran out of resources.
    - TLS certificate expired on the kubelet.

  Resolution flow:
    1. Check node conditions to identify the specific failure (kubelet, network, disk).
    2. Check events for this node.
    3. If the node is recoverable, restart kubelet on the machine.
    4. If unrecoverable, drain the node to safely move workloads, then remove it.
    5. Uncordon after recovery, or replace the node.
commands:
  - "kubectl describe node {name} | grep -A15 'Conditions:'"
  - "kubectl get events --field-selector involvedObject.name={name} --sort-by='.lastTimestamp'"
  - "kubectl get pods --field-selector spec.nodeName={name} -A -o wide"
  - "kubectl drain {name} --ignore-daemonsets --delete-emptydir-data --timeout=120s"
  - "kubectl uncordon {name}"
  - "kubectl get node {name}"
windows_commands:
  - "kubectl describe node {name}"
  - "kubectl get events --field-selector involvedObject.name={name} --sort-by='.lastTimestamp'"
  - "kubectl get pods --field-selector spec.nodeName={name} -A -o wide"
  - "kubectl drain {name} --ignore-daemonsets --delete-emptydir-data --timeout=120s"
  - "kubectl uncordon {name}"
  - "kubectl get node {name}"
'@

Set-Content -Path (Join-Path $knowledgeDir "node_disk_pressure.en.yaml") -Encoding UTF8 -Value @'
key: "node-disk-pressure"
title: "Node Under Disk Pressure"
explanation: |
  The node is critically low on disk space. Kubernetes will aggressively evict pods and
  refuse to schedule new ones until space is freed.

  Common causes:
    - Container images filling up the image filesystem.
    - Application logs writing to the node filesystem (not stdout).
    - Unused volumes or orphan container layers.
    - Large emptyDir volumes consuming ephemeral storage.

  Resolution flow:
    1. Check node conditions to confirm disk pressure.
    2. Identify pods on this node that consume the most ephemeral storage.
    3. Clean up unused images and containers on the node (via SSH or node debug pod).
    4. If critical, drain the node to move workloads while you clean up.
    5. Verify the condition clears.
commands:
  - "kubectl describe node {name} | grep -A15 'Conditions:'"
  - "kubectl get pods --field-selector spec.nodeName={name} -A --sort-by='.metadata.creationTimestamp'"
  - "kubectl debug node/{name} -it --image=busybox -- sh -c 'df -h /host; du -sh /host/var/lib/containerd/io.containerd.snapshotter/*/snapshots/* 2>/dev/null | sort -rh | head -20'"
  - "kubectl drain {name} --ignore-daemonsets --delete-emptydir-data --timeout=120s"
  - "kubectl uncordon {name}"
  - "kubectl get node {name} -o jsonpath='{.status.conditions[?(@.type==\"DiskPressure\")].status}'"
windows_commands:
  - "kubectl describe node {name}"
  - "kubectl get pods --field-selector spec.nodeName={name} -A --sort-by='.metadata.creationTimestamp'"
  - "kubectl debug node/{name} -it --image=busybox -- sh -c 'df -h /host'"
  - "kubectl drain {name} --ignore-daemonsets --delete-emptydir-data --timeout=120s"
  - "kubectl uncordon {name}"
  - "kubectl get node {name} -o jsonpath='{.status.conditions[?(@.type==\"DiskPressure\")].status}'"
'@

Set-Content -Path (Join-Path $knowledgeDir "node_memory_pressure.en.yaml") -Encoding UTF8 -Value @'
key: "node-memory-pressure"
title: "Node Under Memory Pressure"
explanation: |
  The node is dangerously low on available memory. The kubelet will start evicting pods
  that lack resource limits or have the lowest QoS class (BestEffort first, then Burstable).

  Resolution flow:
    1. Identify which pods on this node consume the most memory.
    2. Set or reduce memory limits on the heaviest consumers.
    3. If critical, drain the node to redistribute workloads across the cluster.
    4. Consider adding more nodes or increasing node memory.
    5. Verify the condition clears after workloads are redistributed.
commands:
  - "kubectl top node {name}"
  - "kubectl top pods --field-selector spec.nodeName={name} -A --sort-by=memory"
  - "kubectl get pods --field-selector spec.nodeName={name} -A -o jsonpath='{range .items[*]}{.metadata.namespace}/{.metadata.name} limits={.spec.containers[0].resources.limits.memory}{\"\\n\"}{end}'"
  - "kubectl drain {name} --ignore-daemonsets --delete-emptydir-data --timeout=120s"
  - "kubectl uncordon {name}"
  - "kubectl get node {name} -o jsonpath='{.status.conditions[?(@.type==\"MemoryPressure\")].status}'"
windows_commands:
  - "kubectl top node {name}"
  - "kubectl top pods -A --sort-by=memory"
  - "kubectl get pods --field-selector spec.nodeName={name} -A"
  - "kubectl drain {name} --ignore-daemonsets --delete-emptydir-data --timeout=120s"
  - "kubectl uncordon {name}"
  - "kubectl get node {name} -o jsonpath='{.status.conditions[?(@.type==\"MemoryPressure\")].status}'"
'@

Set-Content -Path (Join-Path $knowledgeDir "node_pid_pressure.en.yaml") -Encoding UTF8 -Value @'
key: "node-pid-pressure"
title: "Node Under PID Pressure"
explanation: |
  The node is running out of process IDs. Too many processes/threads are running, risking
  system instability. The kubelet will start evicting pods to reduce the PID count.

  Common causes:
    - Application spawning excessive child processes or threads without reaping.
    - Fork bombs or runaway processes.
    - Too many pods scheduled on a single node.

  Resolution flow:
    1. Check node conditions to confirm PID pressure.
    2. Identify the top CPU-consuming pods on this node (high CPU often correlates with many PIDs).
    3. Set PID limits on containers or reduce pod density on the node.
    4. If critical, drain the node.
    5. Verify the condition clears.
commands:
  - "kubectl describe node {name} | grep -A15 'Conditions:'"
  - "kubectl top pods --field-selector spec.nodeName={name} -A --sort-by=cpu"
  - "kubectl get pods --field-selector spec.nodeName={name} -A --no-headers | wc -l"
  - "kubectl drain {name} --ignore-daemonsets --delete-emptydir-data --timeout=120s"
  - "kubectl uncordon {name}"
  - "kubectl get node {name} -o jsonpath='{.status.conditions[?(@.type==\"PIDPressure\")].status}'"
windows_commands:
  - "kubectl describe node {name}"
  - "kubectl top pods -A --sort-by=cpu"
  - "kubectl get pods --field-selector spec.nodeName={name} -A --no-headers"
  - "kubectl drain {name} --ignore-daemonsets --delete-emptydir-data --timeout=120s"
  - "kubectl uncordon {name}"
  - "kubectl get node {name} -o jsonpath='{.status.conditions[?(@.type==\"PIDPressure\")].status}'"
'@

Set-Content -Path (Join-Path $knowledgeDir "node_unschedulable.en.yaml") -Encoding UTF8 -Value @'
key: "node-unschedulable"
title: "Node Cordoned (Unschedulable)"
explanation: |
  This node is marked as unschedulable (cordoned). No new pods will be scheduled on it.
  This is typically intentional during maintenance or node drain operations.

  If unexpected, someone or an automated process cordoned it by mistake.

  Resolution flow:
    1. Verify the node is cordoned and check who/when it was cordoned.
    2. If maintenance is complete, uncordon to accept workloads again.
    3. Verify new pods can be scheduled on the node.
commands:
  - "kubectl get node {name} -o jsonpath='{.spec.unschedulable}'"
  - "kubectl get events --field-selector involvedObject.name={name} --sort-by='.lastTimestamp'"
  - "kubectl uncordon {name}"
  - "kubectl get node {name}"
  - "kubectl get pods --field-selector spec.nodeName={name} -A"
windows_commands:
  - "kubectl get node {name} -o jsonpath='{.spec.unschedulable}'"
  - "kubectl get events --field-selector involvedObject.name={name} --sort-by='.lastTimestamp'"
  - "kubectl uncordon {name}"
  - "kubectl get node {name}"
  - "kubectl get pods --field-selector spec.nodeName={name} -A"
'@

Set-Content -Path (Join-Path $knowledgeDir "cluster_node_skew.en.yaml") -Encoding UTF8 -Value @'
key: cluster-node-skew
title: "Node Version Skew Detected"
explanation: |
  Nodes in the cluster are running different kubelet versions. Kubernetes allows kubelets
  to be up to 3 minor versions behind the API server, but mixed versions increase the risk
  of incompatible behavior and complicate troubleshooting.

  Common causes:
    - A rolling upgrade was started but not completed.
    - Auto-scaling groups using an outdated AMI/machine image.
    - Manually managed nodes that were not upgraded.

  Resolution flow:
    1. Identify which nodes are on older versions.
    2. Drain outdated nodes one at a time to safely move workloads.
    3. Upgrade the node (update AMI, run kubeadm upgrade, or replace the node).
    4. Uncordon and verify the node rejoins with the correct version.
commands:
  - "kubectl get nodes -o custom-columns=NAME:.metadata.name,VERSION:.status.nodeInfo.kubeletVersion,OS:.status.nodeInfo.osImage --sort-by='.status.nodeInfo.kubeletVersion'"
  - "kubectl drain {name} --ignore-daemonsets --delete-emptydir-data --timeout=120s"
  - "kubectl uncordon {name}"
  - "kubectl get nodes -o custom-columns=NAME:.metadata.name,VERSION:.status.nodeInfo.kubeletVersion"
windows_commands:
  - "kubectl get nodes -o custom-columns=NAME:.metadata.name,VERSION:.status.nodeInfo.kubeletVersion,OS:.status.nodeInfo.osImage --sort-by='.status.nodeInfo.kubeletVersion'"
  - "kubectl drain {name} --ignore-daemonsets --delete-emptydir-data --timeout=120s"
  - "kubectl uncordon {name}"
  - "kubectl get nodes -o custom-columns=NAME:.metadata.name,VERSION:.status.nodeInfo.kubeletVersion"
'@

# --- Storage Issues ---

Set-Content -Path (Join-Path $knowledgeDir "pvc_pending.en.yaml") -Encoding UTF8 -Value @'
key: pvc_pending
title: "PVC Stuck in Pending State"
explanation: |
  The PersistentVolumeClaim cannot bind to a volume.

  Common causes and fixes:
    - StorageClass does not exist -> Create the StorageClass or change the PVC to use an existing one.
    - No PV matches the requested size/access mode -> Create a PV or reduce the request.
    - Cloud provisioner error (AWS EBS, GCP PD) -> Check IAM permissions and quotas.
    - WaitForFirstConsumer -> The PVC will bind only when a pod using it is scheduled.

  Resolution flow:
    1. Check PVC events for the exact provisioning error.
    2. Verify the StorageClass exists and is configured correctly.
    3. Fix the StorageClass reference or create the missing StorageClass.
    4. Verify the PVC transitions to Bound.
commands:
  - "kubectl describe pvc {name} -n {namespace}"
  - "kubectl get storageclass"
  - "kubectl get pv --sort-by='.spec.capacity.storage'"
  - "kubectl patch pvc {name} -n {namespace} -p '{\"spec\":{\"storageClassName\":\"standard\"}}'"
  - "kubectl get pvc {name} -n {namespace} -w"
windows_commands:
  - "kubectl describe pvc {name} -n {namespace}"
  - "kubectl get storageclass"
  - "kubectl get pv --sort-by='.spec.capacity.storage'"
  - "kubectl patch pvc {name} -n {namespace} -p '{\"spec\":{\"storageClassName\":\"standard\"}}'"
  - "kubectl get pvc {name} -n {namespace} -w"
'@

Set-Content -Path (Join-Path $knowledgeDir "pvc_lost.en.yaml") -Encoding UTF8 -Value @'
key: pvc_lost
title: "PVC in Lost State"
explanation: |
  The PersistentVolume that was bound to this PVC has been deleted or is missing. This is a
  critical state that often implies data loss if the backing storage was also deleted.

  Resolution flow:
    1. Check if the PV still exists or was accidentally deleted.
    2. If the backing storage (EBS volume, NFS share) still exists, recreate the PV pointing to it.
    3. If data is lost, delete the Lost PVC and create a fresh one.
    4. Update the pod/deployment to reference the new PVC if needed.
    5. Verify the new PVC is Bound and pods can mount it.
commands:
  - "kubectl describe pvc {name} -n {namespace}"
  - "kubectl get pv | grep {name}"
  - "kubectl delete pvc {name} -n {namespace}"
  - "kubectl apply -f <pvc-manifest.yaml>"
  - "kubectl get pvc {name} -n {namespace} -w"
windows_commands:
  - "kubectl describe pvc {name} -n {namespace}"
  - "kubectl get pv"
  - "kubectl delete pvc {name} -n {namespace}"
  - "kubectl apply -f <pvc-manifest.yaml>"
  - "kubectl get pvc {name} -n {namespace} -w"
'@

Set-Content -Path (Join-Path $knowledgeDir "pv_failed.en.yaml") -Encoding UTF8 -Value @'
key: pv_failed
title: "PV in Failed State"
explanation: |
  The PersistentVolume failed during the reclaim process (recycle or delete). The backing
  storage could not be cleaned up by the provisioner.

  Common causes:
    - Backing storage (EBS, NFS) was manually deleted outside Kubernetes.
    - The provisioner lacks IAM/RBAC permissions to delete the volume.
    - Network connectivity issue to the storage backend.

  Resolution flow:
    1. Check PV events and the reclaim policy.
    2. If data is not needed, change the reclaim policy to Retain and manually clean up.
    3. Delete the failed PV from Kubernetes.
    4. If the backing storage still exists, recreate the PV with the correct reference.
    5. Verify the volume state.
commands:
  - "kubectl describe pv {name}"
  - "kubectl get pv {name} -o jsonpath='{.spec.persistentVolumeReclaimPolicy}'"
  - "kubectl patch pv {name} -p '{\"spec\":{\"persistentVolumeReclaimPolicy\":\"Retain\"}}'"
  - "kubectl patch pv {name} -p '{\"spec\":{\"claimRef\":null}}'"
  - "kubectl delete pv {name}"
  - "kubectl get pv"
windows_commands:
  - "kubectl describe pv {name}"
  - "kubectl get pv {name} -o jsonpath='{.spec.persistentVolumeReclaimPolicy}'"
  - "kubectl patch pv {name} -p '{\"spec\":{\"persistentVolumeReclaimPolicy\":\"Retain\"}}'"
  - "kubectl patch pv {name} -p '{\"spec\":{\"claimRef\":null}}'"
  - "kubectl delete pv {name}"
  - "kubectl get pv"
'@

Set-Content -Path (Join-Path $knowledgeDir "pv_released.en.yaml") -Encoding UTF8 -Value @'
key: pv_released
title: "PV Released but Not Recycled"
explanation: |
  The PVC that was bound to this PV was deleted, but the PV has a Retain reclaim policy,
  so it stays in Released state. It cannot be automatically bound to a new PVC until the
  old claimRef is manually cleared.

  Resolution flow:
    1. Check if the data on the volume is still needed.
    2. If data is needed, back it up before making changes.
    3. Clear the claimRef to make the PV Available again for new PVCs.
    4. Verify the PV transitions to Available and can bind to a new PVC.
commands:
  - "kubectl describe pv {name}"
  - "kubectl get pv {name} -o jsonpath='{.spec.claimRef}'"
  - "kubectl patch pv {name} --type=json -p '[{\"op\":\"remove\",\"path\":\"/spec/claimRef\"}]'"
  - "kubectl get pv {name}"
windows_commands:
  - "kubectl describe pv {name}"
  - "kubectl get pv {name} -o jsonpath='{.spec.claimRef}'"
  - "kubectl patch pv {name} --type=json -p '[{\"op\":\"remove\",\"path\":\"/spec/claimRef\"}]'"
  - "kubectl get pv {name}"
'@

Set-Content -Path (Join-Path $knowledgeDir "configmap_large.en.yaml") -Encoding UTF8 -Value @'
key: configmap_large
title: "ConfigMap Exceeds 500KB"
explanation: |
  This ConfigMap is unusually large (>500KB). Large ConfigMaps degrade etcd performance,
  increase API server latency, and slow pod initialization when mounted as volumes.

  The hard limit is 1MB (etcd value size limit). Approaching this limit risks write failures.

  Resolution flow:
    1. Check the actual data size of the ConfigMap.
    2. Identify which keys contain the largest data.
    3. Split into multiple smaller ConfigMaps, or move large data to an external store
       (S3, database, external config service).
    4. Update pod specs to reference the new ConfigMap names.
    5. Delete the oversized ConfigMap after migration.
commands:
  - "kubectl get configmap {name} -n {namespace} -o jsonpath='{.data}' | wc -c"
  - "kubectl get configmap {name} -n {namespace} -o json | jq '.data | to_entries[] | {key: .key, size: (.value | length)}' | sort -t: -k2 -rn"
  - "kubectl get configmap {name} -n {namespace} -o yaml > configmap-{name}-backup.yaml"
  - "kubectl create configmap {name}-part1 -n {namespace} --from-literal=<key1>=<value1>"
  - "kubectl create configmap {name}-part2 -n {namespace} --from-literal=<key2>=<value2>"
  - "kubectl get configmap -n {namespace} | grep {name}"
windows_commands:
  - "kubectl get configmap {name} -n {namespace} -o jsonpath='{.data}' | Measure-Object -Character"
  - "kubectl get configmap {name} -n {namespace} -o json"
  - "kubectl get configmap {name} -n {namespace} -o yaml > configmap-{name}-backup.yaml"
  - "kubectl create configmap {name}-part1 -n {namespace} --from-literal=<key1>=<value1>"
  - "kubectl create configmap {name}-part2 -n {namespace} --from-literal=<key2>=<value2>"
  - "kubectl get configmap -n {namespace}"
'@

# --- Service/Networking Issues ---

Set-Content -Path (Join-Path $knowledgeDir "service_no_endpoints.en.yaml") -Encoding UTF8 -Value @'
key: "service-no-endpoints"
title: "Service Has No Endpoints"
explanation: |
  No pods match this Service's selector labels. All traffic sent to this Service will fail
  with connection timeouts. Services only route to pods that are Running AND Ready.

  Common causes:
    - The selector labels in the Service don't match any pod labels (typo or mismatch).
    - The pods exist but are not Ready (failing readiness probes).
    - The Deployment is scaled to zero replicas.

  Resolution flow:
    1. Check the Service selector labels.
    2. List pods with matching labels to see if any exist and are Ready.
    3. Fix the label mismatch or scale up the Deployment.
    4. Verify endpoints populate.
commands:
  - "kubectl get svc {name} -n {namespace} -o jsonpath='{.spec.selector}'"
  - "kubectl get pods -n {namespace} --show-labels"
  - "kubectl get endpoints {name} -n {namespace}"
  - "kubectl label pod <pod-name> -n {namespace} <key>=<value> --overwrite"
  - "kubectl scale deployment/{name} -n {namespace} --replicas=2"
  - "kubectl get endpoints {name} -n {namespace}"
windows_commands:
  - "kubectl get svc {name} -n {namespace} -o jsonpath='{.spec.selector}'"
  - "kubectl get pods -n {namespace} --show-labels"
  - "kubectl get endpoints {name} -n {namespace}"
  - "kubectl label pod <pod-name> -n {namespace} <key>=<value> --overwrite"
  - "kubectl scale deployment/{name} -n {namespace} --replicas=2"
  - "kubectl get endpoints {name} -n {namespace}"
'@

Set-Content -Path (Join-Path $knowledgeDir "service_lb_pending.en.yaml") -Encoding UTF8 -Value @'
key: "service-lb-pending"
title: "LoadBalancer Pending External IP"
explanation: |
  This LoadBalancer Service has not received an external IP from the cloud provider.

  Common causes and fixes:
    - Cloud quota: LB quota exceeded. Check cloud provider quotas.
    - IAM permissions: Cloud Controller Manager lacks permission to create LBs.
    - Bare-metal cluster: LoadBalancer needs MetalLB or similar. Switch to NodePort as alternative.
    - Annotation error: Cloud-specific annotations are misconfigured.

  Resolution flow:
    1. Check Service events for the cloud provider error.
    2. Fix the root cause (quota, IAM, annotations).
    3. If on bare-metal, switch to NodePort or install MetalLB.
    4. Verify the external IP is assigned.
commands:
  - "kubectl describe svc {name} -n {namespace}"
  - "kubectl get events -n {namespace} --field-selector involvedObject.name={name} --sort-by='.lastTimestamp'"
  - "kubectl patch svc {name} -n {namespace} -p '{\"spec\":{\"type\":\"NodePort\"}}'"
  - "kubectl annotate svc {name} -n {namespace} service.beta.kubernetes.io/aws-load-balancer-type=nlb --overwrite"
  - "kubectl get svc {name} -n {namespace} -w"
windows_commands:
  - "kubectl describe svc {name} -n {namespace}"
  - "kubectl get events -n {namespace} --field-selector involvedObject.name={name} --sort-by='.lastTimestamp'"
  - "kubectl patch svc {name} -n {namespace} -p '{\"spec\":{\"type\":\"NodePort\"}}'"
  - "kubectl annotate svc {name} -n {namespace} service.beta.kubernetes.io/aws-load-balancer-type=nlb --overwrite"
  - "kubectl get svc {name} -n {namespace} -w"
'@

Set-Content -Path (Join-Path $knowledgeDir "service_externalname_empty.en.yaml") -Encoding UTF8 -Value @'
key: service-externalname-empty
title: "ExternalName Service Has No Target DNS"
explanation: |
  This ExternalName Service has no spec.externalName set. ExternalName Services act as a
  DNS CNAME alias — without a target, DNS resolution fails and any pod connecting to this
  Service gets a DNS lookup error.

  Resolution flow:
    1. Identify the correct external DNS target this Service should point to.
    2. Patch the Service with the correct externalName.
    3. Verify DNS resolution works from within a pod.
commands:
  - "kubectl get svc {name} -n {namespace} -o yaml"
  - "kubectl patch svc {name} -n {namespace} -p '{\"spec\":{\"externalName\":\"target.example.com\"}}'"
  - "kubectl run dns-test --rm -it --image=busybox:1.36 --restart=Never -- nslookup {name}.{namespace}.svc.cluster.local"
  - "kubectl get svc {name} -n {namespace}"
windows_commands:
  - "kubectl get svc {name} -n {namespace} -o yaml"
  - "kubectl patch svc {name} -n {namespace} -p '{\"spec\":{\"externalName\":\"target.example.com\"}}'"
  - "kubectl run dns-test --rm -it --image=busybox:1.36 --restart=Never -- nslookup {name}.{namespace}.svc.cluster.local"
  - "kubectl get svc {name} -n {namespace}"
'@

Set-Content -Path (Join-Path $knowledgeDir "ingress_empty.en.yaml") -Encoding UTF8 -Value @'
key: ingress_empty
title: "Ingress Has No Rules or Backend"
explanation: |
  This Ingress resource has no routing rules and no default backend configured. It is
  completely ineffective — no external traffic will be routed to any Service.

  Resolution flow:
    1. List available Services in the namespace to identify the intended backend.
    2. Add routing rules that map paths/hosts to the target Services.
    3. Verify the Ingress controller picks up the rules and assigns an address.
commands:
  - "kubectl get ingress {name} -n {namespace} -o yaml"
  - "kubectl get svc -n {namespace}"
  - "kubectl patch ingress {name} -n {namespace} --type=json -p '[{\"op\":\"add\",\"path\":\"/spec/rules\",\"value\":[{\"host\":\"app.example.com\",\"http\":{\"paths\":[{\"path\":\"/\",\"pathType\":\"Prefix\",\"backend\":{\"service\":{\"name\":\"<svc-name>\",\"port\":{\"number\":80}}}}]}}]}]'"
  - "kubectl patch ingress {name} -n {namespace} --type=json -p '[{\"op\":\"add\",\"path\":\"/spec/defaultBackend\",\"value\":{\"service\":{\"name\":\"<svc-name>\",\"port\":{\"number\":80}}}}]'"
  - "kubectl get ingress {name} -n {namespace}"
windows_commands:
  - "kubectl get ingress {name} -n {namespace} -o yaml"
  - "kubectl get svc -n {namespace}"
  - "kubectl patch ingress {name} -n {namespace} --type=json -p '[{\"op\":\"add\",\"path\":\"/spec/rules\",\"value\":[{\"host\":\"app.example.com\",\"http\":{\"paths\":[{\"path\":\"/\",\"pathType\":\"Prefix\",\"backend\":{\"service\":{\"name\":\"<svc-name>\",\"port\":{\"number\":80}}}}]}}]}]'"
  - "kubectl patch ingress {name} -n {namespace} --type=json -p '[{\"op\":\"add\",\"path\":\"/spec/defaultBackend\",\"value\":{\"service\":{\"name\":\"<svc-name>\",\"port\":{\"number\":80}}}}]'"
  - "kubectl get ingress {name} -n {namespace}"
'@

Set-Content -Path (Join-Path $knowledgeDir "netpol_allow_all.en.yaml") -Encoding UTF8 -Value @'
key: netpol_allow_all
title: "Overly Permissive NetworkPolicy (Allow All)"
explanation: |
  This NetworkPolicy allows all ingress traffic, bypassing network segmentation. Any pod in
  any namespace can send traffic to the selected pods, exposing them to unnecessary risk.

  Resolution flow:
    1. Identify which pods this NetworkPolicy selects.
    2. Determine which sources actually need to reach those pods (specific namespaces/pods/ports).
    3. Replace the allow-all rule with specific ingress rules.
    4. Verify legitimate traffic still flows and unauthorized traffic is blocked.
commands:
  - "kubectl get networkpolicy {name} -n {namespace} -o yaml"
  - "kubectl get pods -n {namespace} --show-labels"
  - "kubectl patch networkpolicy {name} -n {namespace} --type=merge -p '{\"spec\":{\"ingress\":[{\"from\":[{\"namespaceSelector\":{\"matchLabels\":{\"name\":\"<allowed-ns>\"}}}],\"ports\":[{\"port\":8080,\"protocol\":\"TCP\"}]}]}}'"
  - "kubectl run nettest --rm -it --image=busybox:1.36 --restart=Never -n {namespace} -- wget -qO- --timeout=3 http://<pod-ip>:8080"
  - "kubectl get networkpolicy {name} -n {namespace} -o yaml"
windows_commands:
  - "kubectl get networkpolicy {name} -n {namespace} -o yaml"
  - "kubectl get pods -n {namespace} --show-labels"
  - "kubectl patch networkpolicy {name} -n {namespace} --type=merge -p '{\"spec\":{\"ingress\":[{\"from\":[{\"namespaceSelector\":{\"matchLabels\":{\"name\":\"<allowed-ns>\"}}}],\"ports\":[{\"port\":8080,\"protocol\":\"TCP\"}]}]}}'"
  - "kubectl run nettest --rm -it --image=busybox:1.36 --restart=Never -n {namespace} -- wget -qO- --timeout=3 http://<pod-ip>:8080"
  - "kubectl get networkpolicy {name} -n {namespace} -o yaml"
'@

Set-Content -Path (Join-Path $knowledgeDir "event_network_not_ready.en.yaml") -Encoding UTF8 -Value @'
key: "event-network-not-ready"
title: "Network Plugin Not Ready"
explanation: |
  The CNI (Container Network Interface) plugin is not ready on one or more nodes, preventing
  pod networking from being configured. Pods on affected nodes cannot start.

  Common causes:
    - CNI DaemonSet pods (calico-node, aws-node, cilium, kindnet) are not running.
    - CNI binary or config is missing/corrupted on the node.
    - Newly added node where CNI hasn't initialized yet.

  Resolution flow:
    1. Identify which CNI is installed and check its DaemonSet status.
    2. Check CNI pod logs on the affected node for initialization errors.
    3. Restart the CNI DaemonSet to reinitialize on all nodes.
    4. Verify the NetworkReady condition clears on all nodes.
commands:
  - "kubectl get pods -n kube-system -o wide | grep -E 'calico|cilium|aws-node|kindnet|flannel'"
  - "kubectl logs -n kube-system -l k8s-app=calico-node --tail=50"
  - "kubectl rollout restart daemonset/calico-node -n kube-system"
  - "kubectl rollout status daemonset/calico-node -n kube-system --timeout=120s"
  - "kubectl get nodes -o custom-columns=NAME:.metadata.name,READY:.status.conditions[?(@.type==\"Ready\")].status"
windows_commands:
  - "kubectl get pods -n kube-system -o wide"
  - "kubectl logs -n kube-system -l k8s-app=calico-node --tail=50"
  - "kubectl rollout restart daemonset/calico-node -n kube-system"
  - "kubectl rollout status daemonset/calico-node -n kube-system --timeout=120s"
  - "kubectl get nodes -o custom-columns=NAME:.metadata.name,READY:.status.conditions[?(@.type==\"Ready\")].status"
'@

# --- Workload Issues ---

Set-Content -Path (Join-Path $knowledgeDir "job_failed.en.yaml") -Encoding UTF8 -Value @'
key: job_failed
title: "Job Has Failed Executions"
explanation: |
  The Job has one or more failed pods. If the backoffLimit is reached, the Job stops retrying
  and remains in a permanently failed state.

  Resolution flow:
    1. Check the Job events and its pod logs to understand the failure reason.
    2. Fix the root cause (image, command, config, permissions).
    3. Delete the failed Job and create a retry from the original spec.
    4. Verify the retried Job completes successfully.
commands:
  - "kubectl describe job {name} -n {namespace}"
  - "kubectl logs -l job-name={name} -n {namespace} --tail=100"
  - "kubectl get pods -l job-name={name} -n {namespace} -o wide"
  - "kubectl delete job {name} -n {namespace}"
  - "kubectl create job {name}-retry --from=job/{name} -n {namespace}"
  - "kubectl get job -n {namespace} | grep {name}"
windows_commands:
  - "kubectl describe job {name} -n {namespace}"
  - "kubectl logs -l job-name={name} -n {namespace} --tail=100"
  - "kubectl get pods -l job-name={name} -n {namespace} -o wide"
  - "kubectl delete job {name} -n {namespace}"
  - "kubectl create job {name}-retry --from=job/{name} -n {namespace}"
  - "kubectl get job -n {namespace}"
'@

Set-Content -Path (Join-Path $knowledgeDir "cronjob_suspended.en.yaml") -Encoding UTF8 -Value @'
key: cronjob_suspended
title: "CronJob Is Suspended"
explanation: |
  This CronJob is suspended (spec.suspend=true). It will not create any new Jobs according
  to its schedule. This might be intentional (maintenance) or a forgotten manual suspension.

  Resolution flow:
    1. Verify if the suspension was intentional.
    2. If ready to resume, patch the CronJob to unsuspend it.
    3. Optionally trigger a manual run to verify it works.
    4. Verify the CronJob creates Jobs on schedule.
commands:
  - "kubectl describe cronjob {name} -n {namespace}"
  - "kubectl get cronjob {name} -n {namespace} -o jsonpath='{.spec.suspend}'  # true = suspended"
  - "kubectl patch cronjob {name} -n {namespace} -p '{\"spec\":{\"suspend\":false}}'"
  - "kubectl create job {name}-manual --from=cronjob/{name} -n {namespace}"
  - "kubectl get jobs -n {namespace} | grep {name}"
windows_commands:
  - "kubectl describe cronjob {name} -n {namespace}"
  - "kubectl get cronjob {name} -n {namespace} -o jsonpath='{.spec.suspend}'"
  - "kubectl patch cronjob {name} -n {namespace} -p '{\"spec\":{\"suspend\":false}}'"
  - "kubectl create job {name}-manual --from=cronjob/{name} -n {namespace}"
  - "kubectl get jobs -n {namespace}"
'@

Set-Content -Path (Join-Path $knowledgeDir "daemonset_misscheduled.en.yaml") -Encoding UTF8 -Value @'
key: daemonset_misscheduled
title: "DaemonSet Not Fully Scheduled"
explanation: |
  The DaemonSet is not running on all desired nodes.

  Common causes and fixes:
    - Taints: Nodes have taints the DaemonSet doesn't tolerate. Add tolerations to the DaemonSet.
    - NodeSelector: The DaemonSet selector excludes some nodes. Adjust the selector or label nodes.
    - Insufficient resources: Nodes don't have enough CPU/memory. Reduce DaemonSet resource requests.

  Resolution flow:
    1. Check which nodes are missing the DaemonSet pod and why.
    2. Check node taints and DaemonSet tolerations.
    3. Add the missing toleration or label, or reduce resource requests.
    4. Verify the DaemonSet rolls out to all desired nodes.
commands:
  - "kubectl describe ds {name} -n {namespace}"
  - "kubectl get nodes -o custom-columns=NAME:.metadata.name,TAINTS:.spec.taints[*].key,LABELS:.metadata.labels"
  - "kubectl get pods -n {namespace} -l app={name} -o wide"
  - "kubectl patch daemonset {name} -n {namespace} --type=strategic -p '{\"spec\":{\"template\":{\"spec\":{\"tolerations\":[{\"key\":\"<taint-key>\",\"operator\":\"Exists\",\"effect\":\"NoSchedule\"}]}}}}'"
  - "kubectl rollout status daemonset/{name} -n {namespace} --timeout=120s"
  - "kubectl get ds {name} -n {namespace}"
windows_commands:
  - "kubectl describe ds {name} -n {namespace}"
  - "kubectl get nodes -o custom-columns=NAME:.metadata.name,TAINTS:.spec.taints[*].key"
  - "kubectl get pods -n {namespace} -l app={name} -o wide"
  - "kubectl patch daemonset {name} -n {namespace} --type=strategic -p '{\"spec\":{\"template\":{\"spec\":{\"tolerations\":[{\"key\":\"<taint-key>\",\"operator\":\"Exists\",\"effect\":\"NoSchedule\"}]}}}}'"
  - "kubectl rollout status daemonset/{name} -n {namespace} --timeout=120s"
  - "kubectl get ds {name} -n {namespace}"
'@

# --- RBAC/Security ---

Set-Content -Path (Join-Path $knowledgeDir "rbac_wildcard.en.yaml") -Encoding UTF8 -Value @'
key: rbac_wildcard
title: "Role Has Wildcard Permissions"
explanation: |
  This Role uses '*' wildcard for verbs or resources within the namespace, granting broader
  access than needed. This violates the principle of least privilege and may expose Secrets,
  ConfigMaps, or workload specs to unintended access.

  Resolution flow:
    1. Inspect the Role to find which rules use wildcards.
    2. Find all RoleBindings that reference this Role to identify affected subjects.
    3. Determine the minimum verbs each subject actually needs (get, list, watch, etc.).
    4. Replace wildcard entries with explicit verbs and resource names.
    5. Verify the restricted permissions work correctly with auth can-i.
commands:
  - "kubectl get role {name} -n {namespace} -o yaml"
  - "kubectl get rolebindings -n {namespace} -o wide | grep {name}"
  - "kubectl auth can-i --list --as=system:serviceaccount:{namespace}:<sa-name> -n {namespace}"
  - "kubectl edit role {name} -n {namespace}"
  - "kubectl auth can-i get pods --as=system:serviceaccount:{namespace}:<sa-name> -n {namespace}"
windows_commands:
  - "kubectl get role {name} -n {namespace} -o yaml"
  - "kubectl get rolebindings -n {namespace} -o wide"
  - "kubectl auth can-i --list --as=system:serviceaccount:{namespace}:<sa-name> -n {namespace}"
  - "kubectl edit role {name} -n {namespace}"
  - "kubectl auth can-i get pods --as=system:serviceaccount:{namespace}:<sa-name> -n {namespace}"
'@

Set-Content -Path (Join-Path $knowledgeDir "rbac_cluster_wildcard.en.yaml") -Encoding UTF8 -Value @'
key: rbac_cluster_wildcard
title: "ClusterRole Has Wildcard Permissions"
explanation: |
  This ClusterRole grants cluster-wide permissions using '*' wildcard for verbs, resources,
  or apiGroups. Any bound subject gets unrestricted access to that resource type across ALL
  namespaces — a significant security risk.

  Note: Built-in ClusterRoles (cluster-admin, admin, edit, view) are intentionally broad.
  Only take action on custom ClusterRoles.

  Resolution flow:
    1. Check if this is a built-in or custom ClusterRole.
    2. Find all ClusterRoleBindings referencing this role.
    3. Replace wildcards with explicit verbs, resources, and apiGroups.
    4. Verify restricted permissions work correctly.
commands:
  - "kubectl get clusterrole {name} -o yaml"
  - "kubectl get clusterrolebindings -o wide | grep {name}"
  - "kubectl auth can-i --list --as=system:serviceaccount:<namespace>:<sa-name>"
  - "kubectl edit clusterrole {name}"
  - "kubectl auth can-i get secrets --as=system:serviceaccount:<namespace>:<sa-name> -A"
windows_commands:
  - "kubectl get clusterrole {name} -o yaml"
  - "kubectl get clusterrolebindings -o wide"
  - "kubectl auth can-i --list --as=system:serviceaccount:<namespace>:<sa-name>"
  - "kubectl edit clusterrole {name}"
  - "kubectl auth can-i get secrets --as=system:serviceaccount:<namespace>:<sa-name> -A"
'@

Set-Content -Path (Join-Path $knowledgeDir "rbac_empty_binding.en.yaml") -Encoding UTF8 -Value @'
key: rbac_empty_binding
title: "RoleBinding Has No Subjects"
explanation: |
  This RoleBinding exists but binds the Role to no subjects (Users, Groups, or ServiceAccounts).
  It provides no permissions and should be cleaned up or properly configured.

  Resolution flow:
    1. Check if this binding was supposed to have subjects.
    2. Either add the intended subject or delete the empty binding.
    3. Verify the change.
commands:
  - "kubectl get rolebinding {name} -n {namespace} -o yaml"
  - "kubectl patch rolebinding {name} -n {namespace} -p '{\"subjects\":[{\"kind\":\"ServiceAccount\",\"name\":\"<sa-name>\",\"namespace\":\"{namespace}\"}]}'"
  - "kubectl delete rolebinding {name} -n {namespace}"
  - "kubectl get rolebindings -n {namespace}"
windows_commands:
  - "kubectl get rolebinding {name} -n {namespace} -o yaml"
  - "kubectl patch rolebinding {name} -n {namespace} -p '{\"subjects\":[{\"kind\":\"ServiceAccount\",\"name\":\"<sa-name>\",\"namespace\":\"{namespace}\"}]}'"
  - "kubectl delete rolebinding {name} -n {namespace}"
  - "kubectl get rolebindings -n {namespace}"
'@

Set-Content -Path (Join-Path $knowledgeDir "rbac_cluster_empty_binding.en.yaml") -Encoding UTF8 -Value @'
key: rbac_cluster_empty_binding
title: "ClusterRoleBinding Has No Subjects"
explanation: |
  This ClusterRoleBinding exists but binds to no subjects. It provides no permissions and
  should be cleaned up or properly configured.

  Resolution flow:
    1. Check if this binding was supposed to have subjects.
    2. Either add the intended subject or delete the empty binding.
    3. Verify the change.
commands:
  - "kubectl get clusterrolebinding {name} -o yaml"
  - "kubectl patch clusterrolebinding {name} -p '{\"subjects\":[{\"kind\":\"ServiceAccount\",\"name\":\"<sa-name>\",\"namespace\":\"<namespace>\"}]}'"
  - "kubectl delete clusterrolebinding {name}"
  - "kubectl get clusterrolebindings | grep {name}"
windows_commands:
  - "kubectl get clusterrolebinding {name} -o yaml"
  - "kubectl patch clusterrolebinding {name} -p '{\"subjects\":[{\"kind\":\"ServiceAccount\",\"name\":\"<sa-name>\",\"namespace\":\"<namespace>\"}]}'"
  - "kubectl delete clusterrolebinding {name}"
  - "kubectl get clusterrolebindings"
'@

Set-Content -Path (Join-Path $knowledgeDir "secret_empty.en.yaml") -Encoding UTF8 -Value @'
key: secret_empty
title: "Secret Contains No Data"
explanation: |
  This Secret has an empty data payload. While technically valid, it often indicates a
  misconfiguration — the deployment pipeline failed to populate it, or an external secret
  manager (Vault, ExternalSecrets) failed to sync.

  Resolution flow:
    1. Check if this Secret is managed by an external operator (ExternalSecret, Vault agent).
    2. If externally managed, check the operator status for sync errors.
    3. If manually managed, populate the Secret with the required data.
    4. Restart pods that mount this Secret to pick up the new data.
commands:
  - "kubectl describe secret {name} -n {namespace}"
  - "kubectl get externalsecret -n {namespace} 2>/dev/null | grep {name}"
  - "kubectl patch secret {name} -n {namespace} -p '{\"data\":{\"<key>\":\"'$(echo -n '<value>' | base64)'\"}}'"
  - "kubectl rollout restart deployment/<deployment-name> -n {namespace}"
  - "kubectl get secret {name} -n {namespace} -o jsonpath='{.data}'"
windows_commands:
  - "kubectl describe secret {name} -n {namespace}"
  - "kubectl get externalsecret -n {namespace} 2>$null"
  - "kubectl edit secret {name} -n {namespace}"
  - "kubectl rollout restart deployment/<deployment-name> -n {namespace}"
  - "kubectl get secret {name} -n {namespace} -o jsonpath='{.data}'"
'@

# --- HPA/Quota/PDB ---

Set-Content -Path (Join-Path $knowledgeDir "hpa_at_max.en.yaml") -Encoding UTF8 -Value @'
key: hpa_at_max
title: "HPA at Maximum Replicas"
explanation: |
  The HorizontalPodAutoscaler has reached its maximum replica count and cannot scale further.
  The workload is under heavy load and may experience degraded performance.

  Resolution flow:
    1. Check current HPA metrics and target utilization.
    2. Check if pods are actually CPU/memory constrained.
    3. Increase maxReplicas if the cluster has capacity.
    4. Alternatively, optimize the application to handle more load per pod.
    5. Verify the HPA state after the change.
commands:
  - "kubectl describe hpa {name} -n {namespace}"
  - "kubectl top pods -n {namespace} --sort-by=cpu"
  - "kubectl get nodes -o custom-columns=NAME:.metadata.name,CPU:.status.allocatable.cpu,MEM:.status.allocatable.memory"
  - "kubectl patch hpa {name} -n {namespace} -p '{\"spec\":{\"maxReplicas\":20}}'"
  - "kubectl set resources deployment/{name} -n {namespace} --requests=cpu=200m --limits=cpu=500m"
  - "kubectl get hpa {name} -n {namespace}"
windows_commands:
  - "kubectl describe hpa {name} -n {namespace}"
  - "kubectl top pods -n {namespace} --sort-by=cpu"
  - "kubectl get nodes -o custom-columns=NAME:.metadata.name,CPU:.status.allocatable.cpu,MEM:.status.allocatable.memory"
  - "kubectl patch hpa {name} -n {namespace} -p '{\"spec\":{\"maxReplicas\":20}}'"
  - "kubectl set resources deployment/{name} -n {namespace} --requests=cpu=200m --limits=cpu=500m"
  - "kubectl get hpa {name} -n {namespace}"
'@

Set-Content -Path (Join-Path $knowledgeDir "quota_reached.en.yaml") -Encoding UTF8 -Value @'
key: quota_reached
title: "ResourceQuota Limit Reached"
explanation: |
  The ResourceQuota limit has been reached for one or more resources (CPU, Memory, Pods, etc.)
  in this namespace. Kubernetes will reject any new resource requests that exceed this limit.

  Resolution flow:
    1. Check which specific resources hit the quota limit.
    2. Identify unused or idle pods/resources that can be cleaned up.
    3. If cleanup is not enough, increase the quota to accommodate legitimate growth.
    4. Verify new resources can be created.
commands:
  - "kubectl describe quota {name} -n {namespace}"
  - "kubectl get pods -n {namespace} --sort-by='.status.startTime'"
  - "kubectl top pods -n {namespace} --sort-by=cpu"
  - "kubectl delete pod <idle-pod> -n {namespace}"
  - "kubectl patch resourcequota {name} -n {namespace} -p '{\"spec\":{\"hard\":{\"pods\":\"50\",\"requests.cpu\":\"10\",\"requests.memory\":\"20Gi\"}}}'"
  - "kubectl describe quota {name} -n {namespace}"
windows_commands:
  - "kubectl describe quota {name} -n {namespace}"
  - "kubectl get pods -n {namespace} --sort-by='.status.startTime'"
  - "kubectl top pods -n {namespace} --sort-by=cpu"
  - "kubectl delete pod <idle-pod> -n {namespace}"
  - "kubectl patch resourcequota {name} -n {namespace} -p '{\"spec\":{\"hard\":{\"pods\":\"50\",\"requests.cpu\":\"10\",\"requests.memory\":\"20Gi\"}}}'"
  - "kubectl describe quota {name} -n {namespace}"
'@

Set-Content -Path (Join-Path $knowledgeDir "quota_near_limit.en.yaml") -Encoding UTF8 -Value @'
key: quota_near_limit
title: "ResourceQuota Near Limit (>90%)"
explanation: |
  The ResourceQuota usage is above 90%. While resources can still be created, the namespace
  is close to exhaustion and new deployments or scale-ups may soon be blocked.

  Resolution flow:
    1. Check current quota usage vs. hard limits.
    2. Identify the top resource consumers.
    3. Clean up unused resources or increase the quota proactively.
    4. Monitor usage after changes.
commands:
  - "kubectl describe quota {name} -n {namespace}"
  - "kubectl top pods -n {namespace} --sort-by=memory"
  - "kubectl get pods -n {namespace} --field-selector=status.phase!=Running"
  - "kubectl delete pod -n {namespace} --field-selector=status.phase=Succeeded"
  - "kubectl patch resourcequota {name} -n {namespace} -p '{\"spec\":{\"hard\":{\"pods\":\"50\",\"requests.cpu\":\"10\",\"requests.memory\":\"20Gi\"}}}'"
  - "kubectl describe quota {name} -n {namespace}"
windows_commands:
  - "kubectl describe quota {name} -n {namespace}"
  - "kubectl top pods -n {namespace} --sort-by=memory"
  - "kubectl get pods -n {namespace} --field-selector=status.phase!=Running"
  - "kubectl delete pod -n {namespace} --field-selector=status.phase=Succeeded"
  - "kubectl patch resourcequota {name} -n {namespace} -p '{\"spec\":{\"hard\":{\"pods\":\"50\",\"requests.cpu\":\"10\",\"requests.memory\":\"20Gi\"}}}'"
  - "kubectl describe quota {name} -n {namespace}"
'@

Set-Content -Path (Join-Path $knowledgeDir "pdb_no_disruptions.en.yaml") -Encoding UTF8 -Value @'
key: pdb_no_disruptions
title: "PDB Blocks All Disruptions"
explanation: |
  The PodDisruptionBudget has DisruptionsAllowed=0, meaning no voluntary disruptions are
  permitted. This blocks node drains, cluster upgrades, and scaling operations.

  This happens when the number of ready pods equals minAvailable, or maxUnavailable is reached.

  Resolution flow:
    1. Check the PDB configuration and current pod count.
    2. Scale up the Deployment to have more replicas than minAvailable.
    3. Or temporarily adjust the PDB to allow disruptions during maintenance.
    4. After maintenance, restore the original PDB settings.
commands:
  - "kubectl describe pdb {name} -n {namespace}"
  - "kubectl get pods -n {namespace} -l app={name} -o wide"
  - "kubectl scale deployment/{name} -n {namespace} --replicas=3"
  - "kubectl patch pdb {name} -n {namespace} -p '{\"spec\":{\"minAvailable\":1}}'"
  - "kubectl get pdb {name} -n {namespace}"
windows_commands:
  - "kubectl describe pdb {name} -n {namespace}"
  - "kubectl get pods -n {namespace} -l app={name} -o wide"
  - "kubectl scale deployment/{name} -n {namespace} --replicas=3"
  - "kubectl patch pdb {name} -n {namespace} -p '{\"spec\":{\"minAvailable\":1}}'"
  - "kubectl get pdb {name} -n {namespace}"
'@

# --- Cluster Issues ---

Set-Content -Path (Join-Path $knowledgeDir "cluster_api_unhealthy.en.yaml") -Encoding UTF8 -Value @'
key: cluster-api-unhealthy
title: "API Server Health Check Failed"
explanation: |
  The Kubernetes API Server /healthz endpoint returned an error. The control plane is
  degraded or unreachable. All cluster operations depend on a healthy API Server.

  Common causes:
    - etcd is down or unreachable.
    - API Server process crashed or is overloaded.
    - Network issues between client and control plane.
    - TLS certificate expired on the API Server.

  Resolution flow:
    1. Run verbose health check to see which component is failing.
    2. For managed clusters (EKS/GKE/AKS), check the cloud provider's health dashboard.
    3. For self-managed clusters, check API Server and etcd pod logs.
    4. Restart the failing control plane component.
    5. Verify the health check passes.
commands:
  - "kubectl get --raw /healthz?verbose"
  - "kubectl get componentstatuses"
  - "kubectl get pods -n kube-system -l component=kube-apiserver -o wide"
  - "kubectl logs -n kube-system -l component=kube-apiserver --tail=50"
  - "kubectl logs -n kube-system -l component=etcd --tail=50"
  - "kubectl get --raw /healthz"
windows_commands:
  - "kubectl get --raw /healthz?verbose"
  - "kubectl get componentstatuses"
  - "kubectl get pods -n kube-system -l component=kube-apiserver -o wide"
  - "kubectl logs -n kube-system -l component=kube-apiserver --tail=50"
  - "kubectl logs -n kube-system -l component=etcd --tail=50"
  - "kubectl get --raw /healthz"
'@

Set-Content -Path (Join-Path $knowledgeDir "cluster_coredns_missing.en.yaml") -Encoding UTF8 -Value @'
key: cluster-coredns-missing
title: "CoreDNS Pods Not Found"
explanation: |
  No CoreDNS pods were found in kube-system. DNS resolution is completely broken across the
  cluster — Services cannot be discovered by DNS name and most applications will fail.

  Common causes:
    - CoreDNS Deployment was accidentally deleted.
    - CoreDNS pods are crash-looping and not being recreated.
    - The label selector was modified.

  Resolution flow:
    1. Check if the CoreDNS Deployment still exists.
    2. If missing, re-apply the CoreDNS manifest for your cluster version.
    3. If present but failing, check its events and logs.
    4. Scale up and verify DNS resolution works.
commands:
  - "kubectl get deployment coredns -n kube-system"
  - "kubectl get pods -n kube-system -l k8s-app=kube-dns -o wide"
  - "kubectl describe deployment coredns -n kube-system"
  - "kubectl scale deployment coredns -n kube-system --replicas=2"
  - "kubectl rollout status deployment/coredns -n kube-system --timeout=120s"
  - "kubectl run dns-test --rm -it --image=busybox:1.36 --restart=Never -- nslookup kubernetes.default"
windows_commands:
  - "kubectl get deployment coredns -n kube-system"
  - "kubectl get pods -n kube-system -l k8s-app=kube-dns -o wide"
  - "kubectl describe deployment coredns -n kube-system"
  - "kubectl scale deployment coredns -n kube-system --replicas=2"
  - "kubectl rollout status deployment/coredns -n kube-system --timeout=120s"
  - "kubectl run dns-test --rm -it --image=busybox:1.36 --restart=Never -- nslookup kubernetes.default"
'@

Set-Content -Path (Join-Path $knowledgeDir "cluster_coredns_unhealthy.en.yaml") -Encoding UTF8 -Value @'
key: cluster-coredns-unhealthy
title: "CoreDNS Pods Unhealthy"
explanation: |
  CoreDNS pods exist but none are in a healthy Running+Ready state. DNS resolution is broken.

  Common causes:
    - CrashLooping due to bad Corefile configuration.
    - Resource limits too low causing OOMKill.
    - ConfigMap with Corefile contains syntax errors.

  Resolution flow:
    1. Check CoreDNS pod statuses and logs.
    2. Inspect the CoreDNS ConfigMap (Corefile) for syntax errors.
    3. Fix the configuration and restart CoreDNS.
    4. Verify DNS resolution is restored.
commands:
  - "kubectl get pods -n kube-system -l k8s-app=kube-dns -o wide"
  - "kubectl logs -n kube-system -l k8s-app=kube-dns --tail=50"
  - "kubectl get configmap coredns -n kube-system -o yaml"
  - "kubectl rollout restart deployment/coredns -n kube-system"
  - "kubectl rollout status deployment/coredns -n kube-system --timeout=120s"
  - "kubectl run dns-test --rm -it --image=busybox:1.36 --restart=Never -- nslookup kubernetes.default"
windows_commands:
  - "kubectl get pods -n kube-system -l k8s-app=kube-dns -o wide"
  - "kubectl logs -n kube-system -l k8s-app=kube-dns --tail=50"
  - "kubectl get configmap coredns -n kube-system -o yaml"
  - "kubectl rollout restart deployment/coredns -n kube-system"
  - "kubectl rollout status deployment/coredns -n kube-system --timeout=120s"
  - "kubectl run dns-test --rm -it --image=busybox:1.36 --restart=Never -- nslookup kubernetes.default"
'@

# --- Event-based Issues ---

Set-Content -Path (Join-Path $knowledgeDir "frequent_events.en.yaml") -Encoding UTF8 -Value @'
key: "frequent-events"
title: "Frequent Warning Events"
explanation: |
  This resource is repeatedly generating Warning events. The fix depends on the event type:

    - BackOff/CrashLoopBackOff: Container is crashing. Check --previous logs and exit code.
    - FailedMount: A Secret, ConfigMap, or PVC doesn't exist. Create the missing resource.
    - FailedScheduling: No node has enough resources. Scale cluster or reduce requests.
    - Liveness/Readiness ProbeErr: Fix probe config (path, port, delays).
    - ErrImagePull/ImagePullBackOff: Fix image name or create imagePullSecret.
    - Evicted: Node under pressure. Check node conditions and set resource limits.

  Resolution flow:
    1. List Warning events and identify the recurring Reason.
    2. Match the Reason to the category above and apply the targeted fix.
    3. Verify the events stop repeating after the fix.
commands:
  - "kubectl get events -n {namespace} --field-selector involvedObject.name={name},type=Warning --sort-by='.lastTimestamp'"
  - "kubectl describe pod {pod} -n {namespace}"
  - "kubectl logs {pod} -n {namespace} --previous --tail=50 2>/dev/null"
  - "kubectl rollout restart deployment/{name} -n {namespace}"
  - "kubectl get events -n {namespace} --field-selector involvedObject.name={name},type=Warning --sort-by='.lastTimestamp'"
windows_commands:
  - "kubectl get events -n {namespace} --field-selector involvedObject.name={name},type=Warning --sort-by='.lastTimestamp'"
  - "kubectl describe pod {pod} -n {namespace}"
  - "kubectl logs {pod} -n {namespace} --previous --tail=50 2>$null"
  - "kubectl rollout restart deployment/{name} -n {namespace}"
  - "kubectl get events -n {namespace} --field-selector involvedObject.name={name},type=Warning --sort-by='.lastTimestamp'"
'@

Set-Content -Path (Join-Path $knowledgeDir "event_probe_failed.en.yaml") -Encoding UTF8 -Value @'
key: "event-probe-failed"
title: "Health Probe Failure"
explanation: |
  A liveness or readiness probe is repeatedly failing:
    - Readiness failure: Pod removed from Service endpoints (no traffic routed to it).
    - Liveness failure: Pod is killed and restarted by kubelet.

  Common causes:
    - Application slow to start — initialDelaySeconds too short.
    - Probe endpoint/port misconfigured.
    - Probe timeout too aggressive for actual response time.

  Resolution flow:
    1. Identify which probe is failing and its configuration.
    2. Check if the application is actually healthy (logs, manual curl).
    3. Adjust probe timing (increase initialDelaySeconds, periodSeconds, timeoutSeconds).
    4. Verify probes pass consistently.
commands:
  - "kubectl describe pod {pod} -n {namespace} | grep -A5 'Liveness:\\|Readiness:'"
  - "kubectl get events -n {namespace} --field-selector involvedObject.name={name},reason=Unhealthy --sort-by='.lastTimestamp'"
  - "kubectl logs {pod} -n {namespace} --tail=100"
  - "kubectl patch deployment/{name} -n {namespace} --type=strategic -p '{\"spec\":{\"template\":{\"spec\":{\"containers\":[{\"name\":\"{name}\",\"livenessProbe\":{\"initialDelaySeconds\":60,\"periodSeconds\":15,\"timeoutSeconds\":5,\"failureThreshold\":5},\"readinessProbe\":{\"initialDelaySeconds\":30,\"periodSeconds\":10,\"timeoutSeconds\":5,\"failureThreshold\":3}}]}}}}'"
  - "kubectl rollout status deployment/{name} -n {namespace} --timeout=120s"
  - "kubectl get pods -n {namespace} -l app={name}"
windows_commands:
  - "kubectl describe pod {pod} -n {namespace}"
  - "kubectl get events -n {namespace} --field-selector involvedObject.name={name},reason=Unhealthy --sort-by='.lastTimestamp'"
  - "kubectl logs {pod} -n {namespace} --tail=100"
  - "kubectl patch deployment/{name} -n {namespace} --type=strategic -p '{\"spec\":{\"template\":{\"spec\":{\"containers\":[{\"name\":\"{name}\",\"livenessProbe\":{\"initialDelaySeconds\":60,\"periodSeconds\":15,\"timeoutSeconds\":5,\"failureThreshold\":5},\"readinessProbe\":{\"initialDelaySeconds\":30,\"periodSeconds\":10,\"timeoutSeconds\":5,\"failureThreshold\":3}}]}}}}'"
  - "kubectl rollout status deployment/{name} -n {namespace} --timeout=120s"
  - "kubectl get pods -n {namespace} -l app={name}"
'@

Set-Content -Path (Join-Path $knowledgeDir "event_failed_mount.en.yaml") -Encoding UTF8 -Value @'
key: "event-failed-mount"
title: "Volume Mount Failure"
explanation: |
  A volume could not be mounted to the container.

  Common causes and fixes:
    - Secret/ConfigMap missing: Create the referenced Secret or ConfigMap.
    - PVC Pending: Fix the StorageClass or provision a PV.
    - NFS/EBS unreachable: Check network connectivity and mount target.
    - ReadWriteOnce conflict: Another pod has exclusive access. Delete that pod or use ReadWriteMany.

  Resolution flow:
    1. Read the event to identify which volume failed and why.
    2. Verify the referenced Secret/ConfigMap/PVC exists.
    3. Create the missing resource or fix the PVC binding issue.
    4. The pod will auto-retry mounting — verify it starts.
commands:
  - "kubectl describe pod {pod} -n {namespace} | grep -A10 'Events:\\|Volumes:'"
  - "kubectl get configmap -n {namespace} -o name"
  - "kubectl get secret -n {namespace} -o name"
  - "kubectl get pvc -n {namespace} -o wide"
  - "kubectl create configmap <cm-name> -n {namespace} --from-literal=<key>=<value>"
  - "kubectl get pod {pod} -n {namespace} -w"
windows_commands:
  - "kubectl describe pod {pod} -n {namespace}"
  - "kubectl get configmap -n {namespace} -o name"
  - "kubectl get secret -n {namespace} -o name"
  - "kubectl get pvc -n {namespace} -o wide"
  - "kubectl create configmap <cm-name> -n {namespace} --from-literal=<key>=<value>"
  - "kubectl get pod {pod} -n {namespace} -w"
'@

Set-Content -Path (Join-Path $knowledgeDir "event_hpa_misconfigured.en.yaml") -Encoding UTF8 -Value @'
key: "event-hpa-misconfigured"
title: "HPA Target Misconfigured"
explanation: |
  The HorizontalPodAutoscaler cannot find or scale its target resource.

  Common causes:
    - FailedGetScale: The target Deployment/StatefulSet does not exist or was deleted.
    - The HPA references a wrong apiVersion or kind for the target.
    - Metrics Server is not installed or unreachable.

  Resolution flow:
    1. Check the HPA events and status for the specific error.
    2. Verify the target workload exists with the correct name and kind.
    3. Fix the HPA scaleTargetRef to match the actual resource.
    4. Verify the HPA starts scaling correctly.
commands:
  - "kubectl describe hpa {name} -n {namespace}"
  - "kubectl get hpa {name} -n {namespace} -o jsonpath='{.spec.scaleTargetRef}'"
  - "kubectl get deployment -n {namespace}"
  - "kubectl patch hpa {name} -n {namespace} -p '{\"spec\":{\"scaleTargetRef\":{\"apiVersion\":\"apps/v1\",\"kind\":\"Deployment\",\"name\":\"<deployment-name>\"}}}'"
  - "kubectl top pods -n {namespace}"
  - "kubectl get hpa {name} -n {namespace}"
windows_commands:
  - "kubectl describe hpa {name} -n {namespace}"
  - "kubectl get hpa {name} -n {namespace} -o jsonpath='{.spec.scaleTargetRef}'"
  - "kubectl get deployment -n {namespace}"
  - "kubectl patch hpa {name} -n {namespace} -p '{\"spec\":{\"scaleTargetRef\":{\"apiVersion\":\"apps/v1\",\"kind\":\"Deployment\",\"name\":\"<deployment-name>\"}}}'"
  - "kubectl top pods -n {namespace}"
  - "kubectl get hpa {name} -n {namespace}"
'@

Set-Content -Path (Join-Path $knowledgeDir "event_image_pull_secret_missing.en.yaml") -Encoding UTF8 -Value @'
key: "event-image-pull-secret-missing"
title: "ImagePullSecret Not Found"
explanation: |
  The imagePullSecret referenced by a pod or ServiceAccount does not exist in the namespace.
  Containers cannot authenticate to the private registry and image pulls will fail.

  Resolution flow:
    1. Identify which imagePullSecret is expected.
    2. Create the docker-registry secret with valid credentials.
    3. Patch the ServiceAccount to reference the new secret.
    4. Delete the failing pod to trigger a retry with the new secret.
    5. Verify the image is pulled successfully.
commands:
  - "kubectl get events -n {namespace} --field-selector involvedObject.name={name} --sort-by='.lastTimestamp'"
  - "kubectl get sa default -n {namespace} -o jsonpath='{.imagePullSecrets}'"
  - "kubectl create secret docker-registry regcred -n {namespace} --docker-server=<registry> --docker-username=<user> --docker-password=<pass>"
  - "kubectl patch serviceaccount default -n {namespace} -p '{\"imagePullSecrets\":[{\"name\":\"regcred\"}]}'"
  - "kubectl delete pod {pod} -n {namespace}"
  - "kubectl get pod -n {namespace} -l app={name} -w"
windows_commands:
  - "kubectl get events -n {namespace} --field-selector involvedObject.name={name} --sort-by='.lastTimestamp'"
  - "kubectl get sa default -n {namespace} -o jsonpath='{.imagePullSecrets}'"
  - "kubectl create secret docker-registry regcred -n {namespace} --docker-server=<registry> --docker-username=<user> --docker-password=<pass>"
  - "kubectl patch serviceaccount default -n {namespace} -p '{\"imagePullSecrets\":[{\"name\":\"regcred\"}]}'"
  - "kubectl delete pod {pod} -n {namespace}"
  - "kubectl get pod -n {namespace} -l app={name} -w"
'@

Set-Content -Path (Join-Path $knowledgeDir "event_invalid_image.en.yaml") -Encoding UTF8 -Value @'
key: "event-invalid-image"
title: "Invalid Container Image"
explanation: |
  The container image reference is malformed or cannot be resolved.

  Common causes:
    - Typo in image name, tag, or registry hostname.
    - The image tag does not exist in the registry.
    - Using "latest" tag but it doesn't exist.

  Resolution flow:
    1. Check the exact image reference in the pod spec.
    2. Verify the image exists in the registry (try pulling locally).
    3. Fix the image reference in the Deployment.
    4. Verify the pod starts with the corrected image.
commands:
  - "kubectl get pod {pod} -n {namespace} -o jsonpath='{.spec.containers[0].image}'"
  - "kubectl describe pod {pod} -n {namespace} | grep -A5 'Events:'"
  - "kubectl set image deployment/{name} -n {namespace} {name}=<correct-image:tag>"
  - "kubectl rollout status deployment/{name} -n {namespace} --timeout=120s"
  - "kubectl get pods -n {namespace} -l app={name}"
windows_commands:
  - "kubectl get pod {pod} -n {namespace} -o jsonpath='{.spec.containers[0].image}'"
  - "kubectl describe pod {pod} -n {namespace}"
  - "kubectl set image deployment/{name} -n {namespace} {name}=<correct-image:tag>"
  - "kubectl rollout status deployment/{name} -n {namespace} --timeout=120s"
  - "kubectl get pods -n {namespace} -l app={name}"
'@

Set-Content -Path (Join-Path $knowledgeDir "event_scaledobject_failed.en.yaml") -Encoding UTF8 -Value @'
key: "event-scaledobject-failed"
title: "KEDA ScaledObject Check Failed"
explanation: |
  KEDA's ScaledObject controller failed to reconcile the autoscaler.

  Common causes:
    - Target Deployment/StatefulSet does not exist or was deleted.
    - The trigger source (Kafka, SQS, Prometheus) is unreachable.
    - KEDA operator itself is unhealthy.

  Resolution flow:
    1. Check ScaledObject status and conditions for the specific error.
    2. Verify the target workload exists.
    3. Check KEDA operator pod logs for detailed errors.
    4. Fix the trigger configuration or recreate the ScaledObject.
    5. Verify KEDA reconciles successfully.
commands:
  - "kubectl describe scaledobject {name} -n {namespace}"
  - "kubectl get deployment -n {namespace} | grep {name}"
  - "kubectl logs -n keda -l app=keda-operator --tail=50"
  - "kubectl delete scaledobject {name} -n {namespace}"
  - "kubectl apply -f <scaledobject-manifest.yaml>"
  - "kubectl get scaledobject {name} -n {namespace}"
windows_commands:
  - "kubectl describe scaledobject {name} -n {namespace}"
  - "kubectl get deployment -n {namespace}"
  - "kubectl logs -n keda -l app=keda-operator --tail=50"
  - "kubectl delete scaledobject {name} -n {namespace}"
  - "kubectl apply -f <scaledobject-manifest.yaml>"
  - "kubectl get scaledobject {name} -n {namespace}"
'@

Set-Content -Path (Join-Path $knowledgeDir "event_secret_sync_failed.en.yaml") -Encoding UTF8 -Value @'
key: "event-secret-sync-failed"
title: "ExternalSecret Sync Failure"
explanation: |
  The ExternalSecret operator failed to sync a secret from the external provider (Vault,
  AWS Secrets Manager, GCP Secret Manager).

  Common causes:
    - Secret path does not exist in the external provider.
    - SecretStore/ClusterSecretStore credentials are expired or misconfigured.
    - External provider is unreachable from the cluster.
    - IAM role or policy lacks required permissions.

  Resolution flow:
    1. Check the ExternalSecret status for the specific error message.
    2. Verify the SecretStore is healthy and can connect to the provider.
    3. Confirm the secret path exists in the external provider.
    4. Fix credentials/IAM and trigger a resync.
    5. Verify the Secret is populated.
commands:
  - "kubectl describe externalsecret {name} -n {namespace}"
  - "kubectl get secretstore -n {namespace} -o yaml"
  - "kubectl get events -n {namespace} --field-selector involvedObject.name={name} --sort-by='.lastTimestamp'"
  - "kubectl annotate externalsecret {name} -n {namespace} force-sync=$(date +%s) --overwrite"
  - "kubectl get secret {name} -n {namespace} -o jsonpath='{.data}'"
windows_commands:
  - "kubectl describe externalsecret {name} -n {namespace}"
  - "kubectl get secretstore -n {namespace} -o yaml"
  - "kubectl get events -n {namespace} --field-selector involvedObject.name={name} --sort-by='.lastTimestamp'"
  - "kubectl annotate externalsecret {name} -n {namespace} force-sync=$([DateTimeOffset]::UtcNow.ToUnixTimeSeconds()) --overwrite"
  - "kubectl get secret {name} -n {namespace} -o jsonpath='{.data}'"
'@

Set-Content -Path (Join-Path $knowledgeDir "event_endpoint_slice_failed.en.yaml") -Encoding UTF8 -Value @'
key: "event-endpoint-slice-failed"
title: "EndpointSlice Update Failure"
explanation: |
  The EndpointSlice controller failed to update endpoint information for a Service. Traffic
  routing to backend pods may be affected.

  This is usually transient (API server load, too many endpoints). If persistent, it may
  indicate RBAC issues or controller problems.

  Resolution flow:
    1. Check Service and EndpointSlice status.
    2. If transient, restart the Deployment to regenerate endpoints.
    3. If persistent, delete stale EndpointSlices to force recreation.
    4. Verify endpoints are correctly populated.
commands:
  - "kubectl get endpointslice -n {namespace} -l kubernetes.io/service-name={name}"
  - "kubectl describe service {name} -n {namespace}"
  - "kubectl rollout restart deployment/{name} -n {namespace}"
  - "kubectl delete endpointslice -n {namespace} -l kubernetes.io/service-name={name}"
  - "kubectl get endpoints {name} -n {namespace}"
windows_commands:
  - "kubectl get endpointslice -n {namespace} -l kubernetes.io/service-name={name}"
  - "kubectl describe service {name} -n {namespace}"
  - "kubectl rollout restart deployment/{name} -n {namespace}"
  - "kubectl delete endpointslice -n {namespace} -l kubernetes.io/service-name={name}"
  - "kubectl get endpoints {name} -n {namespace}"
'@

# --- ServiceAccount ---

Set-Content -Path (Join-Path $knowledgeDir "serviceaccount_no_secrets.en.yaml") -Encoding UTF8 -Value @'
key: serviceaccount_no_secrets
title: "ServiceAccount Has No Secrets or ImagePullSecrets"
explanation: |
  This ServiceAccount has no Secrets or ImagePullSecrets associated.

  IMPORTANT: Since Kubernetes v1.24, ServiceAccounts no longer auto-create long-lived token
  Secrets — this is intentional and a security improvement. For system ServiceAccounts in
  kube-system, this is EXPECTED and benign.

  Only act if ALL of these are true:
    - The ServiceAccount is in a user-managed namespace.
    - Pods using it need to pull from a PRIVATE registry.
    - Pods are failing with ImagePullBackOff.

  Resolution flow:
    1. Verify this is actually causing ImagePullBackOff errors.
    2. Create an imagePullSecret with your registry credentials.
    3. Patch the ServiceAccount to reference the secret.
    4. Restart affected pods and verify images pull successfully.
commands:
  - "kubectl get sa {name} -n {namespace} -o yaml"
  - "kubectl get pods -n {namespace} --field-selector=status.phase!=Running -o wide"
  - "kubectl create secret docker-registry regcred -n {namespace} --docker-server=<registry> --docker-username=<user> --docker-password=<pass>"
  - "kubectl patch serviceaccount {name} -n {namespace} -p '{\"imagePullSecrets\":[{\"name\":\"regcred\"}]}'"
  - "kubectl delete pods -n {namespace} -l serviceaccount={name}"
  - "kubectl get pods -n {namespace} -l serviceaccount={name} -w"
windows_commands:
  - "kubectl get sa {name} -n {namespace} -o yaml"
  - "kubectl get pods -n {namespace} --field-selector=status.phase!=Running -o wide"
  - "kubectl create secret docker-registry regcred -n {namespace} --docker-server=<registry> --docker-username=<user> --docker-password=<pass>"
  - "kubectl patch serviceaccount {name} -n {namespace} -p '{\"imagePullSecrets\":[{\"name\":\"regcred\"}]}'"
  - "kubectl delete pods -n {namespace} -l serviceaccount={name}"
  - "kubectl get pods -n {namespace} -w"
'@

Write-Host "All 57 EN knowledge files rewritten successfully!" -ForegroundColor Green
