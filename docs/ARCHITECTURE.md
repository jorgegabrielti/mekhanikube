# Arquitetura

## Visão Geral do Sistema

Mekhanikube v2.0 é uma solução containerizada **própria** desenvolvida em Go que analisa clusters Kubernetes e fornece explicações via IA local (Ollama). A solução substitui o K8sGPT por um engine customizado mais leve e rápido.

## Componentes Principais

### 🎯 Mekhanikube v2 (Padrão)

**Engine Próprio em Go**

#### CLI (`cmd/mekhanikube/main.go`)
- Framework: Cobra v1.8.0
- Comandos: `analyze`, `version`
- Flags: `--namespace`, `--filter`, `--explain`, `--language`
- Entry point da aplicação

#### Scanner (`internal/scanner/scanner.go`)
- Usa `client-go` para acessar API Kubernetes
- Detecta problemas em Pods:
  - CrashLoopBackOff
  - ImagePullBackOff
  - ContainerStatusUnknown
  - Containers terminados
- Detecta ConfigMaps não utilizados
- Extensível para novos tipos de recursos

#### Analyzer (`internal/analyzer/analyzer.go`)
- Coordena scanning e análise
- Aplica filtros por tipo de recurso
- Integra com Ollama Client para explicações
- Retorna lista estruturada de problemas

#### Ollama Client (`internal/ollama/client.go`)
- Cliente HTTP para API Ollama
- Prompts otimizados para português
- Timeout de 120s para geração
- Health check do serviço Ollama
- URL: `http://host.docker.internal:11434`

#### Types (`pkg/types/types.go`)
- Estruturas compartilhadas
- `Problem`: representa problema detectado
- `AnalyzeOptions`: opções de análise
- `OllamaRequest/Response`: comunicação com Ollama

**Recursos Principais**:
- 🚀 Startup <10s (3x mais rápido que K8sGPT)
- 💾 Imagem ~80MB (60% menor)
- 🔧 Configuração automática (zero setup)
- 🇧🇷 Suporte nativo a português
- ⚡ Performance otimizada

### 🔄 K8sGPT (Modo Legado via Profile)

Disponível com `docker-compose --profile k8sgpt up -d` para compatibilidade.

**Função**: Análise original via ferramenta externa

**Recursos**:
- Análise completa de recursos K8s
- Requer configuração manual de backend
- Imagem maior (~200MB)
- Mantido para retrocompatibilidade

### 🤖 Contêiner Ollama (Compartilhado)

**Função**: Executa modelos LLM localmente

**Recursos**:
- API REST na porta 11434
- Armazenamento persistente de modelos
- Modelo padrão: llama3.1:8b (4.7GB)
- Compartilhado entre Mekhanikube e K8sGPT

## Fluxo de Dados (v2.0)

1. **Solicitação**: `mekhanikube analyze --explain` → CLI Cobra
2. **Inicialização**: CLI → Analyzer → Scanner (client-go)
3. **Scanning**: Scanner → API Kubernetes → Lista de recursos
4. **Análise**: Scanner → Detecta problemas → Lista de Problems
5. **Filtragem**: Analyzer → Aplica filtros → Problems filtrados
6. **Explicação IA** (opcional): 
   - Analyzer → Ollama Client → HTTP POST
   - Ollama → llama3.1:8b → Explicação em PT-BR
   - Ollama Client → Analyzer → Problem com explicação
7. **Output**: CLI → Formata → Console → Usuário

## Estrutura do Código (v2.0)

```
mekhanikube/
├── cmd/
│   └── mekhanikube/
│       └── main.go              # Entry point, CLI Cobra
├── internal/
│   ├── scanner/
│   │   └── scanner.go           # K8s resource scanner
│   ├── analyzer/
│   │   └── analyzer.go          # Analysis coordinator
│   └── ollama/
│       └── client.go            # Ollama HTTP client
├── pkg/
│   └── types/
│       └── types.go             # Shared structures
├── configs/
│   ├── Dockerfile.mekhanikube   # Multi-stage build
│   └── entrypoint-mekhanikube.sh # Container init
├── go.mod / go.sum              # Dependencies
└── docker-compose.yml           # Orchestration
```

## Arquitetura de Rede

### Host Network Mode

```yaml
network_mode: host
```

**Usado por**: Mekhanikube, K8sGPT

**Vantagens**:
- Acesso direto ao cluster K8s local
- Acesso ao Ollama via `host.docker.internal:11434`
- Performance otimizada
- Configuração simplificada

**Considerações**:
- Necessário para acesso ao kubeconfig local
- Permite comunicação entre containers via host

## Gerenciamento de Volumes

### Volumes Persistentes

1. **mekhanikube-ollama-data**: 
   - Armazena modelos LLM (~4.7GB por modelo)
   - Compartilhado entre versões
   
2. **mekhanikube-k8sgpt-config** (legacy):
   - Configuração K8sGPT
   - Apenas quando usando profile k8sgpt

3. **~/.kube/config**: 
   - Montado read-only em ambos containers
   - Modificado automaticamente pelo entrypoint

## Considerações de Segurança

### Acesso ao Kubeconfig
- Montado como **read-only**
- Modificado temporariamente em `/root/.kube/config_mod`
- Nunca altera arquivo original
- Isolado dentro do container

### Segurança de Rede
- Network mode host necessário para acesso ao cluster
- Ollama acessível apenas via containers (não exposto)
- Sem exposição de portas externas
- Comunicação via Docker internal networking

### Privacidade de Dados
- ✅ 100% local - nenhum dado sai da máquina
- ✅ Sem telemetria ou analytics
- ✅ Modelos LLM rodando offline
- ✅ Logs apenas em stdout/stderr

## Performance e Escalabilidade

### Requisitos de Recursos (v2.0)

**Mínimo (Mekhanikube)**:
- 1 núcleo CPU
- 2GB RAM
- 5GB disco (modelo llama3.1:8b)

**Recomendado**:
- 2-4 núcleos CPU
- 4-8GB RAM (Ollama usa ~4GB)
- 10GB disco (múltiplos modelos)

**Comparação v1 vs v2**:
| Métrica | K8sGPT (v1) | Mekhanikube (v2) |
|---------|-------------|------------------|
| Imagem Docker | ~200MB | ~80MB |
| RAM em execução | ~150MB | ~50MB |
| Startup | ~30s | <10s |
| Scan 50 Pods | ~5s | ~2s |

**Recomendado**:
- 4+ núcleos de CPU
- 8GB+ RAM
- 20GB+ disco

### Performance do Modelo

| Modelo | Tamanho | Velocidade | Qualidade |
|--------|---------|------------|-----------|
| tinyllama | 1.1GB | Rápido | Básica |
| gemma:7b | 4.8GB | Médio | Boa |
| mistral | 4.1GB | Médio | Boa |
| llama2:13b | 7.4GB | Lento | Excelente |

### Dicas de Otimização

1. Use modelos menores para análises rápidas
2. Limite o escopo com filtros e namespaces
3. Aloque mais recursos para modelos maiores
4. Use aceleração GPU quando disponível

## Pontos de Integração

### Integração CI/CD

```yaml
# GitLab CI
k8s-analysis:
  script:
    - docker-compose up -d
    - docker exec mekhanikube-k8sgpt k8sgpt analyze --explain > report.txt
  artifacts:
    paths:
      - report.txt
```

### Integração de Monitoramento

Mekhanikube complementa ferramentas de monitoramento existentes:
- Prometheus/Grafana: métricas
- Mekhanikube: detecção e explicação de problemas

### Integração de Alertas

Use Mekhanikube em resposta a alertas para diagnóstico automatizado.

## Solução de Problemas de Arquitetura

Para problemas comuns, consulte [TROUBLESHOOTING.md](TROUBLESHOOTING.md).
