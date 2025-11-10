# Guia de Desenvolvimento - Mekhanikube

## Arquitetura

Mekhanikube é uma ferramenta escrita em Go que analisa clusters Kubernetes e fornece explicações usando IA local.

### Estrutura do Projeto

```
mekhanikube/
├── cmd/
│   └── mekhanikube/          # CLI principal
│       └── main.go           # Ponto de entrada, comandos Cobra
├── internal/                 # Código interno (não exportado)
│   ├── scanner/              # Scanner Kubernetes
│   │   └── scanner.go        # Coleta recursos do cluster
│   ├── analyzer/             # Analisador de problemas
│   │   └── analyzer.go       # Coordena scan + análise
│   └── ollama/               # Cliente Ollama
│       └── client.go         # Comunicação HTTP com Ollama
├── pkg/
│   └── types/                # Tipos públicos
│       └── types.go          # Structs compartilhadas
├── configs/
│   ├── Dockerfile.mekhanikube      # Build otimizado
│   └── entrypoint-mekhanikube.sh   # Script de inicialização
├── go.mod                    # Dependências Go
└── docker-compose.yml        # Orquestração containers
```

## Componentes

### 1. Scanner (`internal/scanner/`)

Responsável por conectar ao cluster e coletar recursos.

**Recursos suportados:**
- ✅ Pods (CrashLoopBackOff, ImagePullBackOff, erros)
- ✅ ConfigMaps (não utilizados)
- 🔜 Services
- 🔜 Deployments
- 🔜 StatefulSets

**Como funciona:**
```go
scanner, _ := scanner.New()
problems, _ := scanner.ScanPods(ctx, "default")
```

### 2. Analyzer (`internal/analyzer/`)

Coordena o scanning e análise, aplicando filtros e enviando para IA.

**Opções:**
```go
opts := types.AnalyzeOptions{
    Namespace: "kube-system",     // Namespace específico
    Filter:    []string{"Pod"},   // Filtrar recursos
    Explain:   true,              // Explicar com IA
    Language:  "Portuguese",      // Idioma
}
```

### 3. Cliente Ollama (`internal/ollama/`)

Cliente HTTP para comunicação com Ollama.

**Features:**
- Timeout configurável (120s para LLMs)
- Prompts otimizados para português
- Health check do serviço
- Tratamento de erros

## Desenvolvimento

### Pré-requisitos

- Go 1.21+
- Docker & Docker Compose
- Cluster Kubernetes local (kind, minikube, k3d, Docker Desktop)

### Setup Local

```bash
# 1. Clone o repositório
git clone https://github.com/jorgegabrielti/mekhanikube.git
cd mekhanikube

# 2. Baixe dependências
go mod download

# 3. Build local
go build -o mekhanikube ./cmd/mekhanikube

# 4. Teste local (requer cluster + ollama rodando)
./mekhanikube analyze --explain --language Portuguese
```

### Build Container

```bash
# Build imagem
docker-compose build mekhanikube

# Iniciar serviços
docker-compose up -d

# Ver logs
docker-compose logs -f mekhanikube
```

### Testes

```bash
# Teste básico (sem IA)
docker exec mekhanikube mekhanikube analyze

# Teste com IA
docker exec mekhanikube mekhanikube analyze --explain --language Portuguese

# Teste filtros
docker exec mekhanikube mekhanikube analyze --filter Pod
docker exec mekhanikube mekhanikube analyze --filter ConfigMap -n kube-system
```

## Adicionando Novos Recursos

### 1. Adicionar Scanner

Edite `internal/scanner/scanner.go`:

```go
func (s *Scanner) ScanServices(ctx context.Context, namespace string) ([]types.Problem, error) {
    var problems []types.Problem
    
    services, err := s.clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
    if err != nil {
        return nil, err
    }
    
    for _, svc := range services.Items {
        // Lógica de detecção de problemas
        if problem := s.checkService(&svc); problem != nil {
            problems = append(problems, *problem)
        }
    }
    
    return problems, nil
}
```

### 2. Adicionar ao Analyzer

Edite `internal/analyzer/analyzer.go`:

```go
// No método Analyze()
shouldScanServices := len(opts.Filter) == 0 || contains(opts.Filter, "Service")

if shouldScanServices {
    problems, err := a.scanner.ScanServices(ctx, opts.Namespace)
    if err != nil {
        return nil, fmt.Errorf("erro ao escanear services: %w", err)
    }
    allProblems = append(allProblems, problems...)
}
```

### 3. Rebuild e Teste

```bash
docker-compose build mekhanikube
docker-compose up -d mekhanikube
docker exec mekhanikube mekhanikube analyze --filter Service --explain
```

## Performance

### Otimizações Implementadas

1. **Build multi-stage**: golang:alpine → alpine (~80MB)
2. **Binário estático**: CGO_ENABLED=0, flags -w -s
3. **Startup rápido**: <10s vs 30s do K8sGPT
4. **Cache Go modules**: Layers Docker otimizados
5. **Detecção automática**: Sem configuração manual

### Benchmarks

| Operação | K8sGPT | Mekhanikube | Melhoria |
|----------|---------|-------------|----------|
| Build | ~60s | ~30s | 50% |
| Startup | 30s | <10s | 67% |
| Scan (10 recursos) | ~2s | ~1s | 50% |
| Com IA (10 recursos) | ~60s | ~40s | 33% |

## Troubleshooting

### Build falha

```bash
# Limpar cache Docker
docker builder prune -a

# Rebuild sem cache
docker-compose build --no-cache mekhanikube
```

### Erro de conexão ao cluster

Verifique `entrypoint-mekhanikube.sh` e ajuste substituições:

```bash
sed 's|https://127.0.0.1|https://host.docker.internal|g'
```

### Ollama timeout

Aumente timeout em `internal/ollama/client.go`:

```go
httpClient: &http.Client{
    Timeout: 180 * time.Second, // 3 minutos
}
```

## Roadmap

### v2.1 (Próximo)
- [ ] Scanner para Services
- [ ] Scanner para Deployments
- [ ] Scanner para StatefulSets
- [ ] Cache de análises
- [ ] Output JSON

### v2.2
- [ ] Análise de resource limits
- [ ] Detecção de problemas de networking
- [ ] Análise de RBAC
- [ ] Métricas de performance

### v3.0
- [ ] Interface web (opcional)
- [ ] API REST
- [ ] Webhooks
- [ ] Integração CI/CD

## Contribuindo

1. Fork o repositório
2. Crie uma branch (`git checkout -b feature/nova-feature`)
3. Commit suas mudanças (`git commit -am 'feat: adiciona nova feature'`)
4. Push para a branch (`git push origin feature/nova-feature`)
5. Abra um Pull Request

## Licença

MIT - Veja [LICENSE](../LICENSE)
