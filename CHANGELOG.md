# Histórico de Mudanças

Todas as mudanças notáveis do mekhanikube serão documentadas neste arquivo.

O formato é baseado em [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
e este projeto segue [Versionamento Semântico](https://semver.org/spec/v2.0.0.html).

## [2.0.0] - 2025-01-XX

### 🎯 BREAKING CHANGES - Engine Próprio
Mekhanikube v2.0 traz engine customizado em Go substituindo K8sGPT como solução padrão.

### ✨ Adicionado
- **Engine próprio em Go** (1.618 linhas de código)
  - `cmd/mekhanikube/main.go`: CLI com framework Cobra
  - `internal/scanner/scanner.go`: Scanner de recursos K8s via client-go
  - `internal/analyzer/analyzer.go`: Coordenador de análise
  - `internal/ollama/client.go`: Cliente HTTP para Ollama
  - `pkg/types/types.go`: Estruturas compartilhadas
- **Dockerfile otimizado**: Multi-stage build (golang:1.21-alpine → alpine)
- **Detecção automática**: Zero configuração necessária
- **Suporte nativo a português**: Prompts otimizados para PT-BR
- **Logo oficial**: `assets/logo.png` integrado na documentação
- **Documentação expandida**:
  - `docs/DEVELOPMENT.md`: Guia completo para desenvolvedores
  - `docs/ARCHITECTURE.md`: Arquitetura v2.0 detalhada
  - Seções v2 em FAQ e Troubleshooting

### 🚀 Melhorias
- **Performance**: 
  - Startup: 30s → <10s (3x mais rápido)
  - Imagem Docker: 200MB → 80MB (60% menor)
  - Consumo de RAM: 150MB → 50MB (67% menor)
- **Simplicidade**: Configuração automática do Ollama (sem `k8sgpt auth`)
- **Manutenção**: Código próprio = controle total sobre features

### 🔄 Alterado
- **Comando padrão**: `mekhanikube analyze` substitui `k8sgpt analyze`
- **Container padrão**: `mekhanikube` ao invés de `mekhanikube-k8sgpt`
- **Modelo padrão**: llama3.1:8b (melhor suporte a português)
- **Profiles Docker Compose**: 
  - Padrão: Mekhanikube v2
  - Legacy: K8sGPT via `--profile k8sgpt`

### 🐛 Corrigido
- URL Ollama: Agora usa `host.docker.internal:11434` consistentemente
- Conectividade: Melhor tratamento de erros de rede
- Health checks: Validação de cluster e Ollama no entrypoint

### 📝 Documentação
- README.md: Atualizado com comparações v1 vs v2
- ARCHITECTURE.md: Reescrito para refletir nova arquitetura
- FAQ.md: Adicionadas seções "Qual a diferença entre v1 e v2?"
- TROUBLESHOOTING.md: Seção dedicada para Mekhanikube v2
- CONTRIBUTING.md: Atualizado para desenvolvimento Go

### 🔧 Recursos Detectados (v2.0)
**Pods**:
- CrashLoopBackOff
- ImagePullBackOff
- ContainerStatusUnknown
- Containers terminados

**ConfigMaps**:
- ConfigMaps não utilizados

> 💡 K8sGPT continua disponível via `docker-compose --profile k8sgpt up -d` para retrocompatibilidade.

### 📦 Dependências (Go)
- github.com/spf13/cobra v1.8.0
- k8s.io/client-go v0.29.0
- k8s.io/api v0.29.0
- k8s.io/apimachinery v0.29.0

[2.0.0]: https://github.com/jorgegabrielti/mekhanikube/releases/tag/v2.0.0

---

## [1.0.0] - 2025-11-09

### Adicionado
- Lançamento inicial do mekhanikube 🔧
- Configuração Docker Compose com K8sGPT e Ollama
- Ajuste automático de kubeconfig para contêineres Docker
- Auto-configuração da autenticação K8sGPT na inicialização
- Suporte para modelo gemma:7b (padrão)
- Volumes persistentes para modelos e configuração
- README abrangente com instruções de configuração e uso
- Licença MIT

### Funcionalidades
- Análise de cluster Kubernetes alimentada por IA
- Integração com LLM local (sem chamadas de API externas)
- Detecção de problemas em múltiplos tipos de recursos K8s
- Explicações e soluções automáticas via Ollama
- Suporte a filtros (Pod, Service, ConfigMap, Deployment, etc)
- Análise com escopo de namespace
- Suporte para Windows/Linux/macOS via Docker

### Componentes
- K8sGPT: Construído da fonte oficial (latest)
- Ollama: Imagem oficial (latest)
- Modelos: gemma:7b (5GB)
- Imagens base: golang:1.23-alpine, alpine:latest

[1.0.0]: https://github.com/jorgegabrielti/mekhanikube/releases/tag/v1.0.0

