$knowledgeDir = Join-Path $PSScriptRoot "..\internal\diagnosis\knowledge"

# --- Problemas de Pod ---

Set-Content -Path (Join-Path $knowledgeDir "crashloopbackoff.pt.yaml") -Encoding UTF8 -Value @'
key: "crashloopbackoff"
title: "Container em CrashLoopBackOff"
explanation: |
  O container está falhando imediatamente após iniciar e o Kubernetes reinicia com backoff
  exponencial (10s -> 20s -> 40s -> ... até 5 min entre tentativas).

  Causa raiz por código de saída:
    - Exit 0:   Processo terminou mas não deveria — entrypoint errado ou comando one-shot em pod de longa execução.
    - Exit 1/2: Erro da aplicação — config inválida, variável de ambiente ausente, exceção não tratada no startup.
    - Exit 127: Binário não encontrado na imagem — CMD/ENTRYPOINT errado ou dependência ausente.
    - Exit 137: OOMKilled — container excedeu o limite de memória (resources.limits.memory).

  Fluxo de resolução:
    1. Leia os logs do container ANTERIOR (a instância atual pode não ter saída ainda).
    2. Verifique o código de saída para classificar a falha.
    3. Inspecione variáveis de ambiente, montagens de volume e configurações de probe no spec do pod.
    4. Aplique a correção direcionada (imagem, config, recursos) e verifique se o pod estabiliza.
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

Set-Content -Path (Join-Path $knowledgeDir "oomkilled.pt.yaml") -Encoding UTF8 -Value @'
key: "oomkilled"
title: "Container OOMKilled"
explanation: |
  O container excedeu o limite de memória (resources.limits.memory) e foi terminado pelo
  OOM killer do kernel Linux (código de saída 137). O pod será reiniciado, mas continuará
  sendo terminado se o limite não for aumentado ou o vazamento de memória não for corrigido.

  Fluxo de resolução:
    1. Confirme o motivo OOMKill no último estado do pod.
    2. Verifique o uso atual de memória vs. limites configurados.
    3. Aumente o limite de memória para um valor seguro (comece com 2x o limite atual).
    4. Se o problema persistir, a aplicação tem vazamento de memória — faça profiling.
    5. Verifique se o pod roda de forma estável após a alteração.
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

Set-Content -Path (Join-Path $knowledgeDir "pod_pending.pt.yaml") -Encoding UTF8 -Value @'
key: "pod-pending"
title: "Pod Preso em Estado Pending"
explanation: |
  O scheduler não consegue colocar este pod em nenhum node. Ele permanecerá Pending até que
  a restrição seja resolvida.

  Causas comuns e correções:
    1. CPU/Memória insuficiente — nenhum node tem a capacidade solicitada. Escale o cluster ou
       reduza os resource requests do pod.
    2. NodeSelector/Affinity — nenhum node corresponde aos labels requeridos. Adicione o label
       a um node ou relaxe as regras de affinity.
    3. Taints — todos os nodes candidatos têm taints que o pod não tolera. Remova o taint
       ou adicione uma toleration ao spec do pod.
    4. PVC não vinculado — uma claim de volume não pode ser satisfeita. Verifique StorageClass e disponibilidade de PV.

  Fluxo de resolução:
    1. Leia os eventos do scheduler para identificar o motivo exato da rejeição.
    2. Aplique a correção direcionada baseada na causa.
    3. Verifique se o pod transiciona para Running.
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

Set-Content -Path (Join-Path $knowledgeDir "high_restarts.pt.yaml") -Encoding UTF8 -Value @'
key: "high-restarts"
title: "Alto Número de Reinicializações do Container"
explanation: |
  Um container neste pod reiniciou um número anormalmente alto de vezes, indicando uma
  falha recorrente que o Kubernetes continua recuperando.

  Causa raiz por código de saída:
    - Exit 137 (OOMKilled): Aumente resources.limits.memory.
    - Exit 1/2 (Crash da app): Corrija config, variáveis de ambiente ou lógica de startup.
    - Exit 0 (Saída inesperada): Corrija o entrypoint — o processo não deveria terminar.
    - Falha de liveness probe: A probe está muito agressiva ou o endpoint muito lento.

  Fluxo de resolução:
    1. Leia os logs da instância ANTERIOR que crashou.
    2. Verifique o código de saída e último estado para classificar o tipo de falha.
    3. Aplique a correção direcionada e reinicie o deployment.
    4. Monitore a contagem de restarts para verificar se estabiliza em zero.
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

Set-Content -Path (Join-Path $knowledgeDir "container_not_ready.pt.yaml") -Encoding UTF8 -Value @'
key: "container-not-ready"
title: "Container Não Pronto (Falha na Readiness Probe)"
explanation: |
  O container está rodando mas sua readiness probe continua falhando. O Kubernetes o remove
  de todos os endpoints do Service, então nenhum tráfego é roteado para este pod.

  Causas comuns:
    - A aplicação está em deadlock ou travada na inicialização.
    - O path/porta da readiness probe está mal configurado.
    - initialDelaySeconds é muito curto para a app iniciar.
    - O endpoint de health retorna 5xx sob carga.

  Fluxo de resolução:
    1. Verifique qual probe está falhando e sua configuração exata.
    2. Leia os logs da aplicação para ver se iniciou corretamente.
    3. Ajuste o timing da probe (aumente initialDelaySeconds, periodSeconds ou timeoutSeconds).
    4. Verifique se o pod fica Ready e aparece nos endpoints do Service.
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

Set-Content -Path (Join-Path $knowledgeDir "createcontainerconfigerror.pt.yaml") -Encoding UTF8 -Value @'
key: "createcontainerconfigerror"
title: "CreateContainerConfigError"
explanation: |
  O Kubernetes não consegue criar o container porque um ConfigMap, Secret ou ServiceAccount
  referenciado não existe no namespace, ou uma chave específica dentro do recurso não
  corresponde à referência no volume mount ou variável de ambiente.

  Fluxo de resolução:
    1. Leia os eventos do pod para identificar o nome e chave exatos do recurso ausente.
    2. Verifique se o ConfigMap/Secret existe no namespace.
    3. Crie o recurso ausente ou corrija a referência no spec do pod.
    4. O pod fará retry automático — verifique se transiciona para Running.
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

Set-Content -Path (Join-Path $knowledgeDir "createcontainererror.pt.yaml") -Encoding UTF8 -Value @'
key: "createcontainererror"
title: "CreateContainerError"
explanation: |
  O runtime do container (containerd/CRI-O) falhou ao criar o container. ConfigMaps e
  Secrets existem, mas algo no nível do runtime deu errado.

  Causas comuns:
    - Comando ou entrypoint inválido especificado no spec do container.
    - Permissão negada ao montar volumes (SELinux, AppArmor, filesystem read-only).
    - Montagem de PVC falhou no nível do node (disco não anexado, NFS inacessível).
    - Conflito de SecurityContext (runAsNonRoot mas imagem roda como root).

  Fluxo de resolução:
    1. Leia os eventos para obter a mensagem de erro do CRI.
    2. Corrija o spec do container (imagem, comando, securityContext, volumeMounts).
    3. Faça redeploy e verifique se o container inicia com sucesso.
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

Set-Content -Path (Join-Path $knowledgeDir "imagepullbackoff.pt.yaml") -Encoding UTF8 -Value @'
key: "imagepullbackoff"
title: "Falha no Pull da Imagem (ImagePullBackOff)"
explanation: |
  O kubelet não consegue puxar a imagem do container do registry.

  Causas comuns e correções:
    - Imagem não encontrada: O nome ou tag da imagem está errado. Corrija a referência.
    - Não autorizado: O registry requer autenticação. Crie um imagePullSecret.
    - Registry inacessível: O node não consegue alcançar o registry. Verifique rede/firewall.
    - Rate limited: Limite do Docker Hub atingido. Use cache pull-through ou autentique.

  Fluxo de resolução:
    1. Leia os eventos do pod para ver o erro exato do registry (unauthorized, not found, timeout).
    2. Verifique se o nome e tag da imagem estão corretos.
    3. Se problema de auth, crie um secret docker-registry e faça patch no ServiceAccount.
    4. Se imagem errada, corrija a referência da imagem no deployment.
    5. Verifique se o pod puxa a imagem e inicia.
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

# --- Problemas de Deployment/ReplicaSet ---

Set-Content -Path (Join-Path $knowledgeDir "deployment_stuck.pt.yaml") -Encoding UTF8 -Value @'
key: "deployment-stuck"
title: "Rollout do Deployment Travado"
explanation: |
  O rollout do Deployment não está progredindo — o novo ReplicaSet não consegue escalar seus pods.
  O Kubernetes marca como travado após o progressDeadlineSeconds (padrão 600s) ser excedido.

  Causas comuns:
    - Novos pods crasham ao iniciar (CrashLoopBackOff, ImagePullBackOff).
    - Recursos insuficientes no cluster para agendar novos pods.
    - Readiness probes falham nos novos pods.

  Fluxo de resolução:
    1. Verifique o status do rollout e identifique por que o novo ReplicaSet está travado.
    2. Inspecione os pods do novo ReplicaSet para erros específicos.
    3. Se a nova versão está quebrada, faça rollback para a última revisão funcional.
    4. Se corrigível, resolva o problema e deixe o rollout continuar.
    5. Verifique se o deployment está totalmente disponível.
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

Set-Content -Path (Join-Path $knowledgeDir "deployment_mismatch.pt.yaml") -Encoding UTF8 -Value @'
key: "deployment-mismatch"
title: "Incompatibilidade de Réplicas do Deployment"
explanation: |
  O número de réplicas prontas não corresponde à contagem desejada. Alguns pods estão falhando
  ao ficar prontos (CrashLoop, Pending, ImagePullBackOff ou readiness probes falhando).

  Fluxo de resolução:
    1. Verifique o status do rollout — se um rollout está em progresso, a incompatibilidade pode ser transitória.
    2. Liste os pods para identificar quais estão falhando e por quê.
    3. Corrija o problema subjacente (recursos, imagem, config) ou faça rollback.
    4. Se os pods estão saudáveis mas travados, force um rollout restart.
    5. Verifique se todas as réplicas ficam prontas.
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

Set-Content -Path (Join-Path $knowledgeDir "deployment_unavailable.pt.yaml") -Encoding UTF8 -Value @'
key: "deployment-unavailable"
title: "Deployment com Réplicas Indisponíveis"
explanation: |
  Uma ou mais réplicas não estão disponíveis (status.unavailableReplicas > 0). Pods estão
  falhando ao iniciar, crashando, falhando probes ou presos em Pending.

  Fluxo de resolução:
    1. Verifique quais pods não estão prontos e identifique o erro específico.
    2. Se rollout ruim, desfaça para a última revisão funcional conhecida.
    3. Se problema transitório, reinicie o deployment para recriar os pods.
    4. Verifique se todas as réplicas ficam disponíveis.
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

Set-Content -Path (Join-Path $knowledgeDir "deployment_zero_replicas.pt.yaml") -Encoding UTF8 -Value @'
key: "deployment-zero-replicas"
title: "Deployment Escalado para Zero"
explanation: |
  Este deployment tem zero réplicas — nenhum pod está rodando. Isso pode ser intencional
  (manutenção) ou causado por um autoscaler (HPA/KEDA) reduzindo, ou erro humano.

  Fluxo de resolução:
    1. Verifique se um HPA ou KEDA ScaledObject controla este deployment.
    2. Se intencional, nenhuma ação necessária.
    3. Se acidental, escale de volta para a contagem de réplicas desejada.
    4. Verifique se os pods iniciam com sucesso.
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

Set-Content -Path (Join-Path $knowledgeDir "replicaset_mismatch.pt.yaml") -Encoding UTF8 -Value @'
key: replicaset_mismatch
title: "Incompatibilidade de Réplicas do ReplicaSet"
explanation: |
  O ReplicaSet tem menos pods prontos que o desejado. Esta é a causa subjacente para a
  indisponibilidade do Deployment.

  Fluxo de resolução:
    1. Identifique quais pods pertencentes a este ReplicaSet estão falhando.
    2. Verifique esses pods para CrashLoopBackOff, erros de ImagePull ou estado Pending.
    3. Corrija a causa raiz no spec do Deployment e deixe propagar.
    4. Se urgente, reinicie o Deployment para forçar a recriação dos pods.
    5. Verifique se todas as réplicas ficam prontas.
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

Set-Content -Path (Join-Path $knowledgeDir "replicaset_orphaned.pt.yaml") -Encoding UTF8 -Value @'
key: replicaset_orphaned
title: "ReplicaSet Órfão Detectado"
explanation: |
  Este ReplicaSet não tem ownerReferences — não é gerenciado por nenhum Deployment. Foi
  criado manualmente ou deixado para trás após um Deployment ser deletado com --cascade=orphan.

  ReplicaSets órfãos consomem recursos (CPU, memória, IPs) sem gerenciamento de ciclo de vida —
  sem rollouts, sem auto-recuperação, sem capacidade de rollback.

  Árvore de decisão:
    - Servindo tráfego de produção? -> Envolva em um novo Deployment para gerenciamento adequado.
    - Sobra de Deployment deletado? -> Verifique se os pods não estão em uso, depois delete.
    - Workload de teste intencional? -> Aceite o risco ou migre para um Deployment.

  Fluxo de resolução:
    1. Confirme que não tem owner references.
    2. Verifique se seus pods estão tratando tráfego (verifique endpoints do Service).
    3. Delete se não usado, ou recrie como Deployment para gerenciamento adequado.
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

Set-Content -Path (Join-Path $knowledgeDir "statefulset_mismatch.pt.yaml") -Encoding UTF8 -Value @'
key: statefulset_mismatch
title: "Incompatibilidade de Réplicas do StatefulSet"
explanation: |
  O StatefulSet tem menos réplicas prontas que o desejado. StatefulSets criam pods sequencialmente
  (pod-0, pod-1, ...) e não prosseguem para o próximo pod se o atual não estiver Ready.

  Causas comuns:
    - PVCs não podem ser vinculados (StorageClass ausente ou sem PVs disponíveis).
    - Inicialização do pod está falhando (init containers, problemas de startup da app).
    - O headless Service necessário para DNS está mal configurado.

  Fluxo de resolução:
    1. Verifique os eventos do StatefulSet e identifique qual pod está travado.
    2. Inspecione esse pod específico para erros de PVC, init container ou aplicação.
    3. Corrija o problema (crie PV, corrija StorageClass, corrija config da app).
    4. Se um pod está travado e não consegue recuperar, delete-o para acionar recriação.
    5. Verifique se todas as réplicas ficam prontas.
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

# --- Problemas de Node ---

Set-Content -Path (Join-Path $knowledgeDir "node_not_ready.pt.yaml") -Encoding UTF8 -Value @'
key: "node-not-ready"
title: "Node Não Pronto"
explanation: |
  O node parou de reportar saúde ao control plane. Não pode agendar novos pods, e pods
  existentes serão despejados após o pod-eviction-timeout (padrão 5 minutos).

  Causas comuns:
    - O processo kubelet crashou ou não está respondendo.
    - Partição de rede entre o node e o API server.
    - A VM/máquina subjacente crashou ou ficou sem recursos.
    - Certificado TLS expirado no kubelet.

  Fluxo de resolução:
    1. Verifique as condições do node para identificar a falha específica (kubelet, rede, disco).
    2. Verifique os eventos deste node.
    3. Se o node é recuperável, reinicie o kubelet na máquina.
    4. Se irrecuperável, drene o node para mover workloads com segurança, depois remova-o.
    5. Faça uncordon após recuperação, ou substitua o node.
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

Set-Content -Path (Join-Path $knowledgeDir "node_disk_pressure.pt.yaml") -Encoding UTF8 -Value @'
key: "node-disk-pressure"
title: "Node com Pressão de Disco"
explanation: |
  O node está criticamente baixo em espaço em disco. O Kubernetes vai despejar pods
  agressivamente e recusar agendar novos até que espaço seja liberado.

  Causas comuns:
    - Imagens de container enchendo o filesystem de imagens.
    - Logs da aplicação gravando no filesystem do node (não stdout).
    - Volumes não utilizados ou camadas de container órfãs.
    - Volumes emptyDir grandes consumindo armazenamento efêmero.

  Fluxo de resolução:
    1. Verifique as condições do node para confirmar pressão de disco.
    2. Identifique os pods neste node que mais consomem armazenamento efêmero.
    3. Limpe imagens e containers não utilizados no node (via SSH ou pod de debug do node).
    4. Se crítico, drene o node para mover workloads enquanto limpa.
    5. Verifique se a condição é resolvida.
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

Set-Content -Path (Join-Path $knowledgeDir "node_memory_pressure.pt.yaml") -Encoding UTF8 -Value @'
key: "node-memory-pressure"
title: "Node com Pressão de Memória"
explanation: |
  O node está perigosamente baixo em memória disponível. O kubelet vai começar a despejar
  pods que não têm resource limits ou têm a menor classe de QoS (BestEffort primeiro, depois Burstable).

  Fluxo de resolução:
    1. Identifique quais pods neste node consomem mais memória.
    2. Defina ou reduza memory limits nos maiores consumidores.
    3. Se crítico, drene o node para redistribuir workloads pelo cluster.
    4. Considere adicionar mais nodes ou aumentar a memória do node.
    5. Verifique se a condição é resolvida após redistribuição dos workloads.
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

Set-Content -Path (Join-Path $knowledgeDir "node_pid_pressure.pt.yaml") -Encoding UTF8 -Value @'
key: "node-pid-pressure"
title: "Node com Pressão de PID"
explanation: |
  O node está ficando sem IDs de processo. Muitos processos/threads estão rodando, arriscando
  instabilidade do sistema. O kubelet vai começar a despejar pods para reduzir a contagem de PIDs.

  Causas comuns:
    - Aplicação gerando processos filhos ou threads excessivos sem recolhimento.
    - Fork bombs ou processos descontrolados.
    - Muitos pods agendados em um único node.

  Fluxo de resolução:
    1. Verifique as condições do node para confirmar pressão de PID.
    2. Identifique os pods com maior consumo de CPU neste node (alta CPU frequentemente correlaciona com muitos PIDs).
    3. Defina limites de PID nos containers ou reduza a densidade de pods no node.
    4. Se crítico, drene o node.
    5. Verifique se a condição é resolvida.
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

Set-Content -Path (Join-Path $knowledgeDir "node_unschedulable.pt.yaml") -Encoding UTF8 -Value @'
key: "node-unschedulable"
title: "Node em Cordon (Não Agendável)"
explanation: |
  Este node está marcado como não agendável (cordoned). Nenhum novo pod será agendado nele.
  Isso é tipicamente intencional durante manutenção ou operações de drain do node.

  Se inesperado, alguém ou um processo automatizado fez cordon por engano.

  Fluxo de resolução:
    1. Verifique se o node está em cordon e quem/quando fez o cordon.
    2. Se a manutenção está completa, faça uncordon para aceitar workloads novamente.
    3. Verifique se novos pods podem ser agendados no node.
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

Set-Content -Path (Join-Path $knowledgeDir "cluster_node_skew.pt.yaml") -Encoding UTF8 -Value @'
key: cluster-node-skew
title: "Divergência de Versão dos Nodes Detectada"
explanation: |
  Os nodes no cluster estão rodando versões diferentes do kubelet. O Kubernetes permite que
  kubelets estejam até 3 versões menores atrás do API server, mas versões mistas aumentam
  o risco de comportamento incompatível e complicam o troubleshooting.

  Causas comuns:
    - Um upgrade rolling foi iniciado mas não completado.
    - Grupos de auto-scaling usando AMI/imagem de máquina desatualizada.
    - Nodes gerenciados manualmente que não foram atualizados.

  Fluxo de resolução:
    1. Identifique quais nodes estão em versões mais antigas.
    2. Drene os nodes desatualizados um por vez para mover workloads com segurança.
    3. Atualize o node (atualize AMI, execute kubeadm upgrade ou substitua o node).
    4. Faça uncordon e verifique se o node reintegra com a versão correta.
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

# --- Problemas de Storage ---

Set-Content -Path (Join-Path $knowledgeDir "pvc_pending.pt.yaml") -Encoding UTF8 -Value @'
key: pvc_pending
title: "PVC Preso em Estado Pending"
explanation: |
  O PersistentVolumeClaim não consegue vincular a um volume.

  Causas comuns e correções:
    - StorageClass não existe -> Crie a StorageClass ou altere o PVC para usar uma existente.
    - Nenhum PV corresponde ao tamanho/modo de acesso solicitado -> Crie um PV ou reduza a solicitação.
    - Erro do provisionador cloud (AWS EBS, GCP PD) -> Verifique permissões IAM e cotas.
    - WaitForFirstConsumer -> O PVC só vinculará quando um pod que o usa for agendado.

  Fluxo de resolução:
    1. Verifique os eventos do PVC para o erro exato de provisionamento.
    2. Verifique se a StorageClass existe e está configurada corretamente.
    3. Corrija a referência da StorageClass ou crie a StorageClass ausente.
    4. Verifique se o PVC transiciona para Bound.
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

Set-Content -Path (Join-Path $knowledgeDir "pvc_lost.pt.yaml") -Encoding UTF8 -Value @'
key: pvc_lost
title: "PVC em Estado Lost"
explanation: |
  O PersistentVolume que estava vinculado a este PVC foi deletado ou está ausente. Este é um
  estado crítico que frequentemente implica perda de dados se o armazenamento subjacente também foi deletado.

  Fluxo de resolução:
    1. Verifique se o PV ainda existe ou foi deletado acidentalmente.
    2. Se o armazenamento subjacente (volume EBS, share NFS) ainda existe, recrie o PV apontando para ele.
    3. Se os dados foram perdidos, delete o PVC Lost e crie um novo.
    4. Atualize o pod/deployment para referenciar o novo PVC se necessário.
    5. Verifique se o novo PVC está Bound e pods podem montá-lo.
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

Set-Content -Path (Join-Path $knowledgeDir "pv_failed.pt.yaml") -Encoding UTF8 -Value @'
key: pv_failed
title: "PV em Estado Failed"
explanation: |
  O PersistentVolume falhou durante o processo de reclaim (recycle ou delete). O armazenamento
  subjacente não pôde ser limpo pelo provisionador.

  Causas comuns:
    - Armazenamento subjacente (EBS, NFS) foi deletado manualmente fora do Kubernetes.
    - O provisionador não tem permissões IAM/RBAC para deletar o volume.
    - Problema de conectividade de rede com o backend de armazenamento.

  Fluxo de resolução:
    1. Verifique os eventos do PV e a política de reclaim.
    2. Se os dados não são necessários, altere a política de reclaim para Retain e limpe manualmente.
    3. Delete o PV com falha do Kubernetes.
    4. Se o armazenamento subjacente ainda existe, recrie o PV com a referência correta.
    5. Verifique o estado do volume.
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

Set-Content -Path (Join-Path $knowledgeDir "pv_released.pt.yaml") -Encoding UTF8 -Value @'
key: pv_released
title: "PV Released mas Não Reciclado"
explanation: |
  O PVC que estava vinculado a este PV foi deletado, mas o PV tem política de reclaim Retain,
  então permanece em estado Released. Não pode ser automaticamente vinculado a um novo PVC até
  que o claimRef antigo seja limpo manualmente.

  Fluxo de resolução:
    1. Verifique se os dados no volume ainda são necessários.
    2. Se os dados são necessários, faça backup antes de fazer alterações.
    3. Limpe o claimRef para tornar o PV Available novamente para novos PVCs.
    4. Verifique se o PV transiciona para Available e pode vincular a um novo PVC.
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

Set-Content -Path (Join-Path $knowledgeDir "configmap_large.pt.yaml") -Encoding UTF8 -Value @'
key: configmap_large
title: "ConfigMap Excede 500KB"
explanation: |
  Este ConfigMap é incomumente grande (>500KB). ConfigMaps grandes degradam a performance do
  etcd, aumentam a latência do API server e tornam a inicialização dos pods mais lenta quando
  montados como volumes.

  O limite rígido é 1MB (limite de tamanho de valor do etcd). Aproximar-se deste limite arrisca falhas de escrita.

  Fluxo de resolução:
    1. Verifique o tamanho real dos dados do ConfigMap.
    2. Identifique quais chaves contêm os maiores dados.
    3. Divida em múltiplos ConfigMaps menores, ou mova dados grandes para armazenamento externo
       (S3, banco de dados, serviço de configuração externo).
    4. Atualize os specs dos pods para referenciar os novos nomes de ConfigMap.
    5. Delete o ConfigMap superdimensionado após a migração.
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

# --- Problemas de Service/Rede ---

Set-Content -Path (Join-Path $knowledgeDir "service_no_endpoints.pt.yaml") -Encoding UTF8 -Value @'
key: "service-no-endpoints"
title: "Service Sem Endpoints"
explanation: |
  Nenhum pod corresponde aos labels de selector deste Service. Todo tráfego enviado a este
  Service falhará com timeouts de conexão. Services só roteiam para pods que estão Running E Ready.

  Causas comuns:
    - Os labels do selector no Service não correspondem a nenhum label de pod (typo ou incompatibilidade).
    - Os pods existem mas não estão Ready (readiness probes falhando).
    - O Deployment está escalado para zero réplicas.

  Fluxo de resolução:
    1. Verifique os labels do selector do Service.
    2. Liste pods com labels correspondentes para ver se existem e estão Ready.
    3. Corrija a incompatibilidade de labels ou escale o Deployment.
    4. Verifique se os endpoints são populados.
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

Set-Content -Path (Join-Path $knowledgeDir "service_lb_pending.pt.yaml") -Encoding UTF8 -Value @'
key: "service-lb-pending"
title: "LoadBalancer com IP Externo Pendente"
explanation: |
  Este Service LoadBalancer não recebeu um IP externo do provedor cloud.

  Causas comuns e correções:
    - Cota cloud: Cota de LB excedida. Verifique cotas do provedor cloud.
    - Permissões IAM: Cloud Controller Manager não tem permissão para criar LBs.
    - Cluster bare-metal: LoadBalancer precisa de MetalLB ou similar. Mude para NodePort como alternativa.
    - Erro de annotation: Annotations específicas do cloud estão mal configuradas.

  Fluxo de resolução:
    1. Verifique os eventos do Service para o erro do provedor cloud.
    2. Corrija a causa raiz (cota, IAM, annotations).
    3. Se em bare-metal, mude para NodePort ou instale MetalLB.
    4. Verifique se o IP externo é atribuído.
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

Set-Content -Path (Join-Path $knowledgeDir "service_externalname_empty.pt.yaml") -Encoding UTF8 -Value @'
key: service-externalname-empty
title: "Service ExternalName Sem DNS Alvo"
explanation: |
  Este Service ExternalName não tem spec.externalName definido. Services ExternalName atuam como
  alias DNS CNAME — sem um alvo, a resolução DNS falha e qualquer pod conectando a este
  Service recebe erro de lookup DNS.

  Fluxo de resolução:
    1. Identifique o alvo DNS externo correto para o qual este Service deve apontar.
    2. Faça patch no Service com o externalName correto.
    3. Verifique se a resolução DNS funciona de dentro de um pod.
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

Set-Content -Path (Join-Path $knowledgeDir "ingress_empty.pt.yaml") -Encoding UTF8 -Value @'
key: ingress_empty
title: "Ingress Sem Regras ou Backend"
explanation: |
  Este recurso Ingress não tem regras de roteamento e nenhum backend padrão configurado. É
  completamente inefetivo — nenhum tráfego externo será roteado para nenhum Service.

  Fluxo de resolução:
    1. Liste os Services disponíveis no namespace para identificar o backend pretendido.
    2. Adicione regras de roteamento que mapeiam paths/hosts para os Services alvo.
    3. Verifique se o Ingress controller aplica as regras e atribui um endereço.
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

Set-Content -Path (Join-Path $knowledgeDir "netpol_allow_all.pt.yaml") -Encoding UTF8 -Value @'
key: netpol_allow_all
title: "NetworkPolicy Excessivamente Permissiva (Allow All)"
explanation: |
  Esta NetworkPolicy permite todo tráfego de ingress, ignorando segmentação de rede. Qualquer
  pod em qualquer namespace pode enviar tráfego para os pods selecionados, expondo-os a riscos desnecessários.

  Fluxo de resolução:
    1. Identifique quais pods esta NetworkPolicy seleciona.
    2. Determine quais origens realmente precisam alcançar esses pods (namespaces/pods/portas específicos).
    3. Substitua a regra allow-all por regras de ingress específicas.
    4. Verifique se o tráfego legítimo continua fluindo e o tráfego não autorizado é bloqueado.
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

Set-Content -Path (Join-Path $knowledgeDir "event_network_not_ready.pt.yaml") -Encoding UTF8 -Value @'
key: "event-network-not-ready"
title: "Plugin de Rede Não Pronto"
explanation: |
  O plugin CNI (Container Network Interface) não está pronto em um ou mais nodes, impedindo
  que a rede dos pods seja configurada. Pods nos nodes afetados não podem iniciar.

  Causas comuns:
    - Pods do DaemonSet CNI (calico-node, aws-node, cilium, kindnet) não estão rodando.
    - Binário ou config do CNI está ausente/corrompido no node.
    - Node recém-adicionado onde o CNI ainda não inicializou.

  Fluxo de resolução:
    1. Identifique qual CNI está instalado e verifique o status do DaemonSet.
    2. Verifique os logs do pod CNI no node afetado para erros de inicialização.
    3. Reinicie o DaemonSet do CNI para reinicializar em todos os nodes.
    4. Verifique se a condição NetworkReady é resolvida em todos os nodes.
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

# --- Problemas de Workload ---

Set-Content -Path (Join-Path $knowledgeDir "job_failed.pt.yaml") -Encoding UTF8 -Value @'
key: job_failed
title: "Job com Execuções Falhadas"
explanation: |
  O Job tem um ou mais pods falhados. Se o backoffLimit for atingido, o Job para de tentar
  novamente e permanece em estado permanentemente falhado.

  Fluxo de resolução:
    1. Verifique os eventos do Job e os logs dos pods para entender o motivo da falha.
    2. Corrija a causa raiz (imagem, comando, config, permissões).
    3. Delete o Job falhado e crie uma retentativa a partir do spec original.
    4. Verifique se o Job retentado completa com sucesso.
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

Set-Content -Path (Join-Path $knowledgeDir "cronjob_suspended.pt.yaml") -Encoding UTF8 -Value @'
key: cronjob_suspended
title: "CronJob Está Suspenso"
explanation: |
  Este CronJob está suspenso (spec.suspend=true). Não criará novos Jobs de acordo com seu
  agendamento. Isso pode ser intencional (manutenção) ou uma suspensão manual esquecida.

  Fluxo de resolução:
    1. Verifique se a suspensão foi intencional.
    2. Se pronto para retomar, faça patch no CronJob para dessuspendê-lo.
    3. Opcionalmente acione uma execução manual para verificar se funciona.
    4. Verifique se o CronJob cria Jobs no agendamento.
commands:
  - "kubectl describe cronjob {name} -n {namespace}"
  - "kubectl get cronjob {name} -n {namespace} -o jsonpath='{.spec.suspend}'  # true = suspenso"
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

Set-Content -Path (Join-Path $knowledgeDir "daemonset_misscheduled.pt.yaml") -Encoding UTF8 -Value @'
key: daemonset_misscheduled
title: "DaemonSet Não Totalmente Agendado"
explanation: |
  O DaemonSet não está rodando em todos os nodes desejados.

  Causas comuns e correções:
    - Taints: Nodes têm taints que o DaemonSet não tolera. Adicione tolerations ao DaemonSet.
    - NodeSelector: O selector do DaemonSet exclui alguns nodes. Ajuste o selector ou rotule os nodes.
    - Recursos insuficientes: Nodes não têm CPU/memória suficiente. Reduza os resource requests do DaemonSet.

  Fluxo de resolução:
    1. Verifique quais nodes estão sem o pod do DaemonSet e por quê.
    2. Verifique os taints dos nodes e as tolerations do DaemonSet.
    3. Adicione a toleration ausente ou label, ou reduza os resource requests.
    4. Verifique se o DaemonSet faz rollout para todos os nodes desejados.
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

# --- RBAC/Segurança ---

Set-Content -Path (Join-Path $knowledgeDir "rbac_wildcard.pt.yaml") -Encoding UTF8 -Value @'
key: rbac_wildcard
title: "Role com Permissões Wildcard"
explanation: |
  Esta Role usa '*' wildcard para verbs ou resources dentro do namespace, concedendo acesso
  mais amplo que o necessário. Isso viola o princípio do menor privilégio e pode expor Secrets,
  ConfigMaps ou specs de workload a acesso não intencional.

  Fluxo de resolução:
    1. Inspecione a Role para encontrar quais rules usam wildcards.
    2. Encontre todos os RoleBindings que referenciam esta Role para identificar subjects afetados.
    3. Determine os verbs mínimos que cada subject realmente precisa (get, list, watch, etc.).
    4. Substitua entradas wildcard por verbs e nomes de recursos explícitos.
    5. Verifique se as permissões restritas funcionam corretamente com auth can-i.
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

Set-Content -Path (Join-Path $knowledgeDir "rbac_cluster_wildcard.pt.yaml") -Encoding UTF8 -Value @'
key: rbac_cluster_wildcard
title: "ClusterRole com Permissões Wildcard"
explanation: |
  Este ClusterRole concede permissões cluster-wide usando '*' wildcard para verbs, resources
  ou apiGroups. Qualquer subject vinculado recebe acesso irrestrito a esse tipo de recurso em
  TODOS os namespaces — um risco significativo de segurança.

  Nota: ClusterRoles built-in (cluster-admin, admin, edit, view) são intencionalmente amplos.
  Só tome ação em ClusterRoles customizados.

  Fluxo de resolução:
    1. Verifique se é um ClusterRole built-in ou customizado.
    2. Encontre todos os ClusterRoleBindings que referenciam este role.
    3. Substitua wildcards por verbs, resources e apiGroups explícitos.
    4. Verifique se as permissões restritas funcionam corretamente.
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

Set-Content -Path (Join-Path $knowledgeDir "rbac_empty_binding.pt.yaml") -Encoding UTF8 -Value @'
key: rbac_empty_binding
title: "RoleBinding Sem Subjects"
explanation: |
  Este RoleBinding existe mas vincula a Role a nenhum subject (Users, Groups ou ServiceAccounts).
  Não fornece permissões e deve ser limpo ou configurado corretamente.

  Fluxo de resolução:
    1. Verifique se este binding deveria ter subjects.
    2. Adicione o subject pretendido ou delete o binding vazio.
    3. Verifique a alteração.
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

Set-Content -Path (Join-Path $knowledgeDir "rbac_cluster_empty_binding.pt.yaml") -Encoding UTF8 -Value @'
key: rbac_cluster_empty_binding
title: "ClusterRoleBinding Sem Subjects"
explanation: |
  Este ClusterRoleBinding existe mas não vincula a nenhum subject. Não fornece permissões e
  deve ser limpo ou configurado corretamente.

  Fluxo de resolução:
    1. Verifique se este binding deveria ter subjects.
    2. Adicione o subject pretendido ou delete o binding vazio.
    3. Verifique a alteração.
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

Set-Content -Path (Join-Path $knowledgeDir "secret_empty.pt.yaml") -Encoding UTF8 -Value @'
key: secret_empty
title: "Secret Sem Dados"
explanation: |
  Este Secret tem payload de dados vazio. Embora tecnicamente válido, frequentemente indica uma
  misconfiguration — o pipeline de deploy falhou ao populá-lo, ou um gerenciador de secrets
  externo (Vault, ExternalSecrets) falhou ao sincronizar.

  Fluxo de resolução:
    1. Verifique se este Secret é gerenciado por um operador externo (ExternalSecret, Vault agent).
    2. Se gerenciado externamente, verifique o status do operador para erros de sincronização.
    3. Se gerenciado manualmente, popule o Secret com os dados necessários.
    4. Reinicie os pods que montam este Secret para aplicar os novos dados.
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

Set-Content -Path (Join-Path $knowledgeDir "hpa_at_max.pt.yaml") -Encoding UTF8 -Value @'
key: hpa_at_max
title: "HPA no Máximo de Réplicas"
explanation: |
  O HorizontalPodAutoscaler atingiu sua contagem máxima de réplicas e não pode escalar mais.
  O workload está sob carga pesada e pode experimentar performance degradada.

  Fluxo de resolução:
    1. Verifique as métricas atuais do HPA e utilização alvo.
    2. Verifique se os pods estão realmente limitados por CPU/memória.
    3. Aumente maxReplicas se o cluster tem capacidade.
    4. Alternativamente, otimize a aplicação para lidar com mais carga por pod.
    5. Verifique o estado do HPA após a alteração.
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

Set-Content -Path (Join-Path $knowledgeDir "quota_reached.pt.yaml") -Encoding UTF8 -Value @'
key: quota_reached
title: "Limite do ResourceQuota Atingido"
explanation: |
  O limite do ResourceQuota foi atingido para um ou mais recursos (CPU, Memória, Pods, etc.)
  neste namespace. O Kubernetes rejeitará qualquer nova solicitação de recurso que exceda este limite.

  Fluxo de resolução:
    1. Verifique quais recursos específicos atingiram o limite da quota.
    2. Identifique pods/recursos não utilizados ou ociosos que podem ser limpos.
    3. Se a limpeza não for suficiente, aumente a quota para acomodar crescimento legítimo.
    4. Verifique se novos recursos podem ser criados.
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

Set-Content -Path (Join-Path $knowledgeDir "quota_near_limit.pt.yaml") -Encoding UTF8 -Value @'
key: quota_near_limit
title: "ResourceQuota Próximo do Limite (>90%)"
explanation: |
  O uso do ResourceQuota está acima de 90%. Embora recursos ainda possam ser criados, o namespace
  está próximo da exaustão e novos deployments ou scale-ups podem ser bloqueados em breve.

  Fluxo de resolução:
    1. Verifique o uso atual da quota vs. limites rígidos.
    2. Identifique os maiores consumidores de recursos.
    3. Limpe recursos não utilizados ou aumente a quota proativamente.
    4. Monitore o uso após as alterações.
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

Set-Content -Path (Join-Path $knowledgeDir "pdb_no_disruptions.pt.yaml") -Encoding UTF8 -Value @'
key: pdb_no_disruptions
title: "PDB Bloqueia Todas as Interrupções"
explanation: |
  O PodDisruptionBudget tem DisruptionsAllowed=0, significando que nenhuma interrupção voluntária
  é permitida. Isso bloqueia drains de node, upgrades do cluster e operações de scaling.

  Isso acontece quando o número de pods prontos é igual a minAvailable, ou maxUnavailable foi atingido.

  Fluxo de resolução:
    1. Verifique a configuração do PDB e contagem atual de pods.
    2. Escale o Deployment para ter mais réplicas que minAvailable.
    3. Ou ajuste temporariamente o PDB para permitir interrupções durante manutenção.
    4. Após a manutenção, restaure as configurações originais do PDB.
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

# --- Problemas de Cluster ---

Set-Content -Path (Join-Path $knowledgeDir "cluster_api_unhealthy.pt.yaml") -Encoding UTF8 -Value @'
key: cluster-api-unhealthy
title: "Verificação de Saúde do API Server Falhou"
explanation: |
  O endpoint /healthz do Kubernetes API Server retornou um erro. O control plane está
  degradado ou inacessível. Todas as operações do cluster dependem de um API Server saudável.

  Causas comuns:
    - etcd está parado ou inacessível.
    - Processo do API Server crashou ou está sobrecarregado.
    - Problemas de rede entre o cliente e o control plane.
    - Certificado TLS expirado no API Server.

  Fluxo de resolução:
    1. Execute verificação de saúde detalhada para ver qual componente está falhando.
    2. Para clusters gerenciados (EKS/GKE/AKS), verifique o dashboard de saúde do provedor cloud.
    3. Para clusters auto-gerenciados, verifique os logs dos pods do API Server e etcd.
    4. Reinicie o componente do control plane que está falhando.
    5. Verifique se a verificação de saúde passa.
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

Set-Content -Path (Join-Path $knowledgeDir "cluster_coredns_missing.pt.yaml") -Encoding UTF8 -Value @'
key: cluster-coredns-missing
title: "Pods do CoreDNS Não Encontrados"
explanation: |
  Nenhum pod do CoreDNS foi encontrado no kube-system. A resolução DNS está completamente
  quebrada em todo o cluster — Services não podem ser descobertos por nome DNS e a maioria
  das aplicações falhará.

  Causas comuns:
    - O Deployment do CoreDNS foi deletado acidentalmente.
    - Os pods do CoreDNS estão em crash-looping e não sendo recriados.
    - O label selector foi modificado.

  Fluxo de resolução:
    1. Verifique se o Deployment do CoreDNS ainda existe.
    2. Se ausente, re-aplique o manifesto do CoreDNS para a versão do seu cluster.
    3. Se presente mas falhando, verifique seus eventos e logs.
    4. Escale e verifique se a resolução DNS funciona.
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

Set-Content -Path (Join-Path $knowledgeDir "cluster_coredns_unhealthy.pt.yaml") -Encoding UTF8 -Value @'
key: cluster-coredns-unhealthy
title: "Pods do CoreDNS Não Saudáveis"
explanation: |
  Os pods do CoreDNS existem mas nenhum está em estado saudável Running+Ready. A resolução DNS está quebrada.

  Causas comuns:
    - CrashLooping devido a configuração ruim do Corefile.
    - Resource limits muito baixos causando OOMKill.
    - ConfigMap com Corefile contém erros de sintaxe.

  Fluxo de resolução:
    1. Verifique os status e logs dos pods do CoreDNS.
    2. Inspecione o ConfigMap do CoreDNS (Corefile) para erros de sintaxe.
    3. Corrija a configuração e reinicie o CoreDNS.
    4. Verifique se a resolução DNS está restaurada.
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

# --- Problemas Baseados em Eventos ---

Set-Content -Path (Join-Path $knowledgeDir "frequent_events.pt.yaml") -Encoding UTF8 -Value @'
key: "frequent-events"
title: "Eventos de Warning Frequentes"
explanation: |
  Este recurso está gerando eventos Warning repetidamente. A correção depende do tipo de evento:

    - BackOff/CrashLoopBackOff: Container está crashando. Verifique logs --previous e código de saída.
    - FailedMount: Um Secret, ConfigMap ou PVC não existe. Crie o recurso ausente.
    - FailedScheduling: Nenhum node tem recursos suficientes. Escale o cluster ou reduza requests.
    - Liveness/Readiness ProbeErr: Corrija config da probe (path, porta, delays).
    - ErrImagePull/ImagePullBackOff: Corrija nome da imagem ou crie imagePullSecret.
    - Evicted: Node sob pressão. Verifique condições do node e defina resource limits.

  Fluxo de resolução:
    1. Liste eventos Warning e identifique o Reason recorrente.
    2. Associe o Reason à categoria acima e aplique a correção direcionada.
    3. Verifique se os eventos param de se repetir após a correção.
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

Set-Content -Path (Join-Path $knowledgeDir "event_probe_failed.pt.yaml") -Encoding UTF8 -Value @'
key: "event-probe-failed"
title: "Falha na Health Probe"
explanation: |
  Uma liveness ou readiness probe está falhando repetidamente:
    - Falha de readiness: Pod removido dos endpoints do Service (sem tráfego roteado para ele).
    - Falha de liveness: Pod é terminado e reiniciado pelo kubelet.

  Causas comuns:
    - Aplicação lenta para iniciar — initialDelaySeconds muito curto.
    - Endpoint/porta da probe mal configurados.
    - Timeout da probe muito agressivo para o tempo real de resposta.

  Fluxo de resolução:
    1. Identifique qual probe está falhando e sua configuração.
    2. Verifique se a aplicação está realmente saudável (logs, curl manual).
    3. Ajuste o timing da probe (aumente initialDelaySeconds, periodSeconds, timeoutSeconds).
    4. Verifique se as probes passam consistentemente.
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

Set-Content -Path (Join-Path $knowledgeDir "event_failed_mount.pt.yaml") -Encoding UTF8 -Value @'
key: "event-failed-mount"
title: "Falha na Montagem de Volume"
explanation: |
  Um volume não pôde ser montado no container.

  Causas comuns e correções:
    - Secret/ConfigMap ausente: Crie o Secret ou ConfigMap referenciado.
    - PVC Pending: Corrija a StorageClass ou provisione um PV.
    - NFS/EBS inacessível: Verifique conectividade de rede e mount target.
    - Conflito ReadWriteOnce: Outro pod tem acesso exclusivo. Delete esse pod ou use ReadWriteMany.

  Fluxo de resolução:
    1. Leia o evento para identificar qual volume falhou e por quê.
    2. Verifique se o Secret/ConfigMap/PVC referenciado existe.
    3. Crie o recurso ausente ou corrija o problema de vinculação do PVC.
    4. O pod fará retry automático da montagem — verifique se inicia.
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

Set-Content -Path (Join-Path $knowledgeDir "event_hpa_misconfigured.pt.yaml") -Encoding UTF8 -Value @'
key: "event-hpa-misconfigured"
title: "Alvo do HPA Mal Configurado"
explanation: |
  O HorizontalPodAutoscaler não consegue encontrar ou escalar seu recurso alvo.

  Causas comuns:
    - FailedGetScale: O Deployment/StatefulSet alvo não existe ou foi deletado.
    - O HPA referencia um apiVersion ou kind errado para o alvo.
    - Metrics Server não está instalado ou inacessível.

  Fluxo de resolução:
    1. Verifique os eventos e status do HPA para o erro específico.
    2. Verifique se o workload alvo existe com o nome e kind corretos.
    3. Corrija o scaleTargetRef do HPA para corresponder ao recurso real.
    4. Verifique se o HPA começa a escalar corretamente.
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

Set-Content -Path (Join-Path $knowledgeDir "event_image_pull_secret_missing.pt.yaml") -Encoding UTF8 -Value @'
key: "event-image-pull-secret-missing"
title: "ImagePullSecret Não Encontrado"
explanation: |
  O imagePullSecret referenciado por um pod ou ServiceAccount não existe no namespace.
  Containers não conseguem autenticar no registry privado e pulls de imagem falharão.

  Fluxo de resolução:
    1. Identifique qual imagePullSecret é esperado.
    2. Crie o secret docker-registry com credenciais válidas.
    3. Faça patch no ServiceAccount para referenciar o novo secret.
    4. Delete o pod falhando para acionar retry com o novo secret.
    5. Verifique se a imagem é puxada com sucesso.
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

Set-Content -Path (Join-Path $knowledgeDir "event_invalid_image.pt.yaml") -Encoding UTF8 -Value @'
key: "event-invalid-image"
title: "Imagem de Container Inválida"
explanation: |
  A referência da imagem do container está malformada ou não pode ser resolvida.

  Causas comuns:
    - Typo no nome da imagem, tag ou hostname do registry.
    - A tag da imagem não existe no registry.
    - Usando tag "latest" mas ela não existe.

  Fluxo de resolução:
    1. Verifique a referência exata da imagem no spec do pod.
    2. Verifique se a imagem existe no registry (tente puxar localmente).
    3. Corrija a referência da imagem no Deployment.
    4. Verifique se o pod inicia com a imagem corrigida.
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

Set-Content -Path (Join-Path $knowledgeDir "event_scaledobject_failed.pt.yaml") -Encoding UTF8 -Value @'
key: "event-scaledobject-failed"
title: "Verificação do ScaledObject KEDA Falhou"
explanation: |
  O controller do ScaledObject do KEDA falhou ao reconciliar o autoscaler.

  Causas comuns:
    - Deployment/StatefulSet alvo não existe ou foi deletado.
    - A fonte do trigger (Kafka, SQS, Prometheus) está inacessível.
    - O próprio operador KEDA não está saudável.

  Fluxo de resolução:
    1. Verifique o status e condições do ScaledObject para o erro específico.
    2. Verifique se o workload alvo existe.
    3. Verifique os logs do pod do operador KEDA para erros detalhados.
    4. Corrija a configuração do trigger ou recrie o ScaledObject.
    5. Verifique se o KEDA reconcilia com sucesso.
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

Set-Content -Path (Join-Path $knowledgeDir "event_secret_sync_failed.pt.yaml") -Encoding UTF8 -Value @'
key: "event-secret-sync-failed"
title: "Falha na Sincronização do ExternalSecret"
explanation: |
  O operador ExternalSecret falhou ao sincronizar um secret do provedor externo (Vault,
  AWS Secrets Manager, GCP Secret Manager).

  Causas comuns:
    - O caminho do secret não existe no provedor externo.
    - Credenciais do SecretStore/ClusterSecretStore expiraram ou estão mal configuradas.
    - Provedor externo inacessível a partir do cluster.
    - Role ou policy IAM não tem as permissões necessárias.

  Fluxo de resolução:
    1. Verifique o status do ExternalSecret para a mensagem de erro específica.
    2. Verifique se o SecretStore está saudável e pode conectar ao provedor.
    3. Confirme que o caminho do secret existe no provedor externo.
    4. Corrija credenciais/IAM e acione uma ressincronização.
    5. Verifique se o Secret é populado.
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

Set-Content -Path (Join-Path $knowledgeDir "event_endpoint_slice_failed.pt.yaml") -Encoding UTF8 -Value @'
key: "event-endpoint-slice-failed"
title: "Falha na Atualização do EndpointSlice"
explanation: |
  O controller do EndpointSlice falhou ao atualizar informações de endpoint para um Service.
  O roteamento de tráfego para pods backend pode ser afetado.

  Isso geralmente é transitório (carga do API server, muitos endpoints). Se persistente, pode
  indicar problemas de RBAC ou do controller.

  Fluxo de resolução:
    1. Verifique o status do Service e EndpointSlice.
    2. Se transitório, reinicie o Deployment para regenerar os endpoints.
    3. Se persistente, delete EndpointSlices obsoletos para forçar recriação.
    4. Verifique se os endpoints estão corretamente populados.
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

Set-Content -Path (Join-Path $knowledgeDir "serviceaccount_no_secrets.pt.yaml") -Encoding UTF8 -Value @'
key: serviceaccount_no_secrets
title: "ServiceAccount Sem Secrets ou ImagePullSecrets"
explanation: |
  Este ServiceAccount não tem Secrets ou ImagePullSecrets associados.

  IMPORTANTE: Desde o Kubernetes v1.24, ServiceAccounts não criam mais automaticamente token
  Secrets de longa duração — isso é intencional e uma melhoria de segurança. Para ServiceAccounts
  do sistema no kube-system, isso é ESPERADO e benigno.

  Só aja se TODAS estas condições forem verdadeiras:
    - O ServiceAccount está em um namespace gerenciado pelo usuário.
    - Pods usando-o precisam puxar de um registry PRIVADO.
    - Pods estão falhando com ImagePullBackOff.

  Fluxo de resolução:
    1. Verifique se isso está realmente causando erros ImagePullBackOff.
    2. Crie um imagePullSecret com as credenciais do seu registry.
    3. Faça patch no ServiceAccount para referenciar o secret.
    4. Reinicie os pods afetados e verifique se as imagens são puxadas com sucesso.
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

Write-Host "Todos os 57 arquivos de conhecimento PT reescritos com sucesso!" -ForegroundColor Green
