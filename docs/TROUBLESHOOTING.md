# Guia de Solução de Problemas

## 🆕 Mekhanikube v2.0

### 1. Mekhanikube Não Consegue Conectar ao Cluster

**Sintomas**:
```
Error: unable to connect to Kubernetes cluster
```

**Diagnóstico**:
```bash
# Verificar conexão do container
docker exec mekhanikube kubectl get nodes

# Ver logs do container
docker logs mekhanikube

# Verificar kubeconfig
docker exec mekhanikube cat /root/.kube/config_mod
```

**Soluções**:

#### A. Cluster Kubernetes não está rodando
```bash
# Verificar cluster local
kubectl cluster-info

# Docker Desktop - verificar se Kubernetes está habilitado
# Settings → Kubernetes → Enable Kubernetes
```

#### B. Kubeconfig inválido
```bash
# Verificar no host
kubectl get nodes

# Verificar montagem no container
docker inspect mekhanikube | grep -A 10 Mounts
```

**Correção**: Garantir que `~/.kube/config` existe e está montado corretamente em `docker-compose.yml`.

---

### 2. Mekhanikube Não Consegue Conectar ao Ollama

**Sintomas**:
```
Error: failed to connect to Ollama API
dial tcp: lookup host.docker.internal: no such host
```

**Diagnóstico**:
```bash
# Verificar Ollama
docker ps | grep ollama
docker logs mekhanikube-ollama

# Testar conectividade
docker exec mekhanikube curl -s http://host.docker.internal:11434/api/tags
```

**Soluções**:

#### A. Ollama não está rodando
```bash
# Verificar status
docker-compose ps

# Reiniciar serviço
docker-compose restart ollama
```

#### B. Modelo não instalado
```bash
# Listar modelos
docker exec mekhanikube-ollama ollama list

# Instalar llama3.1:8b (padrão)
docker exec mekhanikube-ollama ollama pull llama3.1:8b
```

---

### 3. Análise Sem Explicação da IA

**Sintomas**:
Problemas encontrados mas sem explicações da IA.

**Causas**:
- Flag `--explain` não foi usada
- Ollama não está respondendo

**Correção**:
```bash
# Usar flag --explain
docker exec mekhanikube mekhanikube analyze --explain --language Portuguese

# Verificar saúde do Ollama
docker exec mekhanikube-ollama ollama list
```

---

### 4. Performance Lenta

**Sintomas**:
Análise demorando muito.

**Soluções**:

#### A. Primeira execução (normal)
- Ollama carrega modelo em memória (~5-10s)
- Execuções subsequentes são mais rápidas

#### B. Modelo muito grande
```bash
# Usar modelo mais leve
docker exec mekhanikube-ollama ollama pull tinyllama

# Ou desativar explicações
docker exec mekhanikube mekhanikube analyze
```

---

##  Problemas Gerais do Sistema

### 1. Mekhanikube Não Consegue Conectar à API Kubernetes

**Sintomas**:
```
Error: unable to connect to Kubernetes cluster
```

**Causas e Soluções**:

#### A. Kubeconfig Não Encontrado ou Inválido

```bash
# Verificar se o kubeconfig existe
ls ~/.kube/config

# Verificar se o kubeconfig é válido
kubectl cluster-info

# Verificar se o contêiner pode ver o arquivo
docker exec mekhanikube-k8sgpt ls -l /root/.kube/config
```

**Correção**: Garantir que o caminho do kubeconfig no `docker-compose.yml` está correto:
```yaml
volumes:
  - C:/Users/${USERNAME}/.kube/config:/root/.kube/config:ro
```

#### B. Endereço Incorreto do Servidor API

**Problema**: Kubeconfig usa `127.0.0.1` que não funciona em contêineres.

**Correção**: O entrypoint.sh corrige isso automaticamente, mas verifique:
```bash
docker exec mekhanikube-k8sgpt cat /root/.kube/config_mod
```

Deve mostrar `host.docker.internal` em vez de `127.0.0.1`.

#### C. API Kubernetes Não Acessível

```bash
# Testar do contêiner K8sGPT
docker exec mekhanikube-k8sgpt kubectl cluster-info

# Testar resolução DNS
docker exec mekhanikube-k8sgpt nslookup host.docker.internal
```

**Correção**: Garantir que seu cluster Kubernetes está rodando:
```bash
# Para Kubernetes do Docker Desktop
docker info | grep -i kubernetes

# Para Minikube
minikube status
```

---

### 2. API Ollama Não Responde

**Sintomas**:
```
Error: failed to connect to Ollama API
Connection refused on localhost:11434
```

**Diagnóstico**:
```bash
# Verificar status do contêiner Ollama
docker ps | grep ollama

# Verificar logs do Ollama
docker logs mekhanikube-ollama

# Testar API Ollama
curl http://localhost:11434/api/tags
```

**Soluções**:

#### A. Contêiner Não Está Rodando
```bash
# Reiniciar Ollama
docker-compose restart ollama

# Verificar saúde
docker inspect mekhanikube-ollama | grep Health -A 10
```

#### B. Modelo Não Carregado
```bash
# Listar modelos instalados
docker exec mekhanikube-ollama ollama list

# Instalar modelo se estiver faltando
docker exec mekhanikube-ollama ollama pull gemma:7b
```

#### C. Conflito de Porta
```bash
# Verificar se a porta 11434 está em uso
netstat -ano | findstr :11434  # Windows
lsof -i :11434                  # Linux/Mac

# Mudar porta no .env:
OLLAMA_PORT=11435
```

---

### 3. Backend K8sGPT Não Configurado

**Sintomas**:
```
Error: no backend configured
```

**Diagnóstico**:
```bash
# Verificar status de autenticação do K8sGPT
docker exec mekhanikube-k8sgpt k8sgpt auth list
```

**Correção**:
```bash
# Reconfigurar backend
docker exec mekhanikube-k8sgpt k8sgpt auth add --backend ollama --model gemma:7b --baseurl http://localhost:11434
docker exec mekhanikube-k8sgpt k8sgpt auth default -p ollama
```

---

### 4. Falha no Download do Modelo

**Sintomas**:
```
Error: failed to pull model
```

**Causas**:

#### A. Sem Conexão com Internet
```bash
# Testar conectividade
docker exec mekhanikube-ollama ping -c 3 ollama.ai
```

#### B. Espaço em Disco Insuficiente
```bash
# Verificar uso de disco do Docker
docker system df

# Verificar tamanho do volume
docker volume inspect mekhanikube-ollama-data
```

**Correção**:
```bash
# Limpar recursos não utilizados
docker system prune

# Ou mais agressivo:
docker system prune -a --volumes
```

#### C. Erro de Digitação no Nome do Modelo
```bash
# Listar modelos disponíveis em: https://ollama.ai/library

# Exemplos corretos:
docker exec mekhanikube-ollama ollama pull gemma:7b
docker exec mekhanikube-ollama ollama pull mistral
docker exec mekhanikube-ollama ollama pull llama2
```

---

### 5. Contêiner Não Inicia

**Sintomas**:
```
Error response from daemon: container exited immediately
```

**Diagnóstico**:
```bash
# Verificar logs
docker-compose logs

# Logs de contêiner específico
docker logs mekhanikube-k8sgpt
docker logs mekhanikube-ollama

# Verificar configuração do Docker Compose
docker-compose config
```

**Correções Comuns**:

#### A. Erro de Sintaxe no docker-compose.yml
```bash
# Validar configuração
docker-compose config -q
```

#### B. Falha na Montagem de Volume
```bash
# Windows: Verificar formato do caminho
# Correto: C:/Users/username/.kube/config
# Errado: C:\Users\username\.kube\config

# Linux/Mac: Verificar permissões
chmod 644 ~/.kube/config
```

#### C. Porta Já em Uso
```bash
# Encontrar o que está usando a porta
netstat -ano | findstr :11434  # Windows
sudo lsof -i :11434            # Linux/Mac

# Matar o processo ou mudar a porta
```

---

### 6. Análise Não Retorna Problemas

**Sintomas**:
```
No problems detected
```

**Isso pode ser normal!** Mas verifique:

#### A. Verificar Namespaces Específicos
```bash
# Listar todos os namespaces
docker exec mekhanikube-k8sgpt kubectl get namespaces

# Analisar namespace específico
docker exec mekhanikube-k8sgpt k8sgpt analyze --namespace kube-system --explain
```

#### B. Verificar Filtros Disponíveis
```bash
# Listar todos os analisadores
docker exec mekhanikube-k8sgpt k8sgpt filters list

# Tentar recursos específicos
docker exec mekhanikube-k8sgpt k8sgpt analyze --filter=Pod --explain
docker exec mekhanikube-k8sgpt k8sgpt analyze --filter=Service --explain
```

#### C. Verificar se o Cluster Tem Recursos
```bash
# Verificar se o cluster tem cargas de trabalho
docker exec mekhanikube-k8sgpt kubectl get all --all-namespaces
```

---

### 7. Explicações de IA Não Funcionam

**Sintomas**:
```
Issues detected but no AI explanations
```

**Verificações**:

#### A. Verificar Flag --explain
```bash
# Deve incluir --explain
docker exec mekhanikube-k8sgpt k8sgpt analyze --explain

# Ou manualmente:
docker exec mekhanikube-k8sgpt k8sgpt analyze --explain
```

#### B. Verificar Conexão do Backend
```bash
# Verificar se o backend está ativo
docker exec mekhanikube-k8sgpt k8sgpt auth list

# Deve mostrar:
# Active: true
# Provider: ollama
```

#### C. Compatibilidade do Modelo
```bash
# Alguns modelos podem não funcionar bem
# Modelos recomendados:
docker exec mekhanikube-ollama ollama pull gemma:7b
docker exec mekhanikube-ollama ollama pull mistral
```

---

### 8. Performance Lenta

**Sintomas**:
- Análise leva >5 minutos
- Sistema fica sem resposta

**Otimizações**:

#### A. Usar Modelo Menor
```bash
# Mais rápido mas menos preciso
docker exec mekhanikube-ollama ollama pull tinyllama

# Balanceado
docker exec mekhanikube-ollama ollama pull gemma:7b
```

#### B. Limitar Escopo
```bash
# Analisar um namespace
docker exec mekhanikube-k8sgpt k8sgpt analyze --namespace default --explain

# Analisar tipo de recurso específico
docker exec mekhanikube-k8sgpt k8sgpt analyze --filter=Pod --explain
```

#### C. Alocar Mais Recursos
```yaml
# No docker-compose.yml, adicione:
services:
  ollama:
    deploy:
      resources:
        limits:
          memory: 8G
          cpus: '4'
```

---

### 9. Problemas de Permissão de Volume

**Sintomas** (Linux/Mac):
```
Error: permission denied
```

**Correção**:
```bash
# Verificar propriedade do volume
docker volume inspect mekhanikube-ollama-data

# Resetar permissões
docker-compose down -v
docker-compose up -d
```

---

### 10. Problemas de Rede no Windows

**Sintomas**:
```
Error: cannot resolve host.docker.internal
```

**Correção**:
```bash
# Garantir que o Docker Desktop está usando WSL 2
wsl --set-default-version 2

# Ou mudar para modo de rede bridge no docker-compose.yml:
network_mode: bridge

# E atualizar portas:
ports:
  - "11434:11434"
```

---

## Comandos de Depuração

### Verificar Saúde Geral
```bash
docker-compose ps
```

### Ver Todos os Logs
```bash
docker-compose logs
```

### Acesso Shell Interativo
```bash
# Contêiner K8sGPT
docker exec -it mekhanikube-k8sgpt /bin/sh

# Contêiner Ollama
docker exec -it mekhanikube-ollama /bin/sh
```

### Testar Conectividade
```bash
# De K8sGPT para Ollama
docker exec mekhanikube-k8sgpt curl -f http://localhost:11434/api/tags

# De K8sGPT para Kubernetes
docker exec mekhanikube-k8sgpt kubectl get nodes
```

### Resetar Tudo
```bash
# Opção nuclear - remove todos os dados
docker-compose down -v

# Então reconstrua
docker-compose up -d
```

---

## Obtendo Ajuda

Se ainda estiver travado:

1. **Verificar Logs**: `docker-compose logs`
2. **Executar Verificação de Saúde**: `docker-compose ps`
3. **Executar Testes**: Verificar conectividade básica
4. **Pesquisar Issues**: [GitHub Issues](https://github.com/jorgegabrielti/mekhanikube/issues)
5. **Abrir Nova Issue**: Incluir:
   - SO e versão do Docker
   - Saída de `docker-compose ps`
   - Logs relevantes
   - Passos para reproduzir

---

## Dicas de Prevenção

### Antes de Iniciar

- [ ] Garantir que o Docker Desktop está rodando
- [ ] Verificar que o cluster Kubernetes está acessível
- [ ] Verificar espaço em disco disponível (mín 10GB)
- [ ] Confirmar que o caminho do kubeconfig está correto

### Melhores Práticas

- Use `docker-compose up -d` para instalação inicial
- Execute `docker-compose ps` regularmente
- Mantenha o Docker Desktop atualizado
- Não modifique volumes manualmente
- Faça backup de configurações importantes

### Manutenção Regular

```bash
# Semanalmente
docker system prune              # Limpar recursos não utilizados

# Mensalmente
docker system prune -a           # Limpeza profunda

# Antes de atualizações
docker-compose down -v           # Reset completo
```
