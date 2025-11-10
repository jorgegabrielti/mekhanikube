# Estrutura do Projeto

## Visão Geral

Este documento descreve a organização e propósito dos arquivos e diretórios no projeto Mekhanikube.

## Estrutura de Diretórios

```
mekhanikube/
├── cmd/                     # 🆕 Aplicações Go
│   └── mekhanikube/
│       └── main.go          # Entry point CLI (Cobra)
│
├── internal/                # 🆕 Código interno Go
│   ├── scanner/
│   │   └── scanner.go       # Scanner de recursos K8s
│   ├── analyzer/
│   │   └── analyzer.go      # Coordenador de análise
│   └── ollama/
│       └── client.go        # Cliente HTTP Ollama
│
├── pkg/                     # 🆕 Bibliotecas públicas Go
│   └── types/
│       └── types.go         # Estruturas compartilhadas
│
├── configs/                 # Configurações e Dockerfiles
│   ├── Dockerfile.mekhanikube
│   ├── entrypoint-mekhanikube.sh
│   ├── Dockerfile.k8sgpt    # Legacy
│   └── entrypoint.sh        # Legacy
│
├── assets/                  # 🆕 Recursos estáticos
│   └── logo.png             # Logo oficial Mekhanikube
│
├── docs/                    # Documentação
│   ├── ARCHITECTURE.md      # ✅ Atualizado para v2.0
│   ├── DEVELOPMENT.md       # 🆕 Guia de desenvolvimento Go
│   ├── FAQ.md               # ✅ Atualizado para v2.0
│   ├── TROUBLESHOOTING.md   # ✅ Atualizado para v2.0
│   ├── PROJECT_STRUCTURE.md # Este arquivo
│   └── PROJECT_IMPROVEMENTS.md
│
├── scripts/                 # Scripts utilitários (legacy)
│   ├── analyze.sh
│   ├── change-model.sh
│   ├── healthcheck.sh
│   └── release.sh
│
├── .github/                 # GitHub workflows
│   └── workflows/
│       └── docker-build.yml
│
├── go.mod                   # 🆕 Dependências Go
├── go.sum                   # 🆕 Checksums de módulos Go
├── docker-compose.yml       # Configuração principal (profiles)
│
├── README.md                # Documentação principal (v2.0)
├── LICENSE                  # Licença MIT
├── CHANGELOG.md             # ✅ Histórico de mudanças (v2.0.0)
├── CONTRIBUTING.md          # ✅ Guia de contribuição (Go)
├── CODE_OF_CONDUCT.md       # Código de conduta
└── SECURITY.md              # Política de segurança

```

## Propósito dos Diretórios (v2.0)

### 🆕 Código Go

#### `cmd/mekhanikube/`
**Entry point da aplicação**
- `main.go`: CLI usando Cobra framework
- Define comandos: `analyze`, `version`
- Configura flags e parâmetros

#### `internal/scanner/`
**Scanner de recursos Kubernetes**
- Conecta à API K8s via client-go
- Detecta problemas em Pods (CrashLoopBackOff, ImagePullBackOff, etc.)
- Detecta ConfigMaps não utilizados
- Retorna lista de `Problem`

#### `internal/analyzer/`
**Coordenador de análise**
- Orquestra scanning de recursos
- Aplica filtros por tipo de recurso
- Integra com Ollama para explicações
- Retorna resultados formatados

#### `internal/ollama/`
**Cliente HTTP para Ollama**
- Comunica com API Ollama (port 11434)
- Envia prompts otimizados para português
- Processa respostas da IA
- Health check do serviço

#### `pkg/types/`
**Estruturas compartilhadas**
- `Problem`: Representa problema detectado
- `AnalyzeOptions`: Opções de análise (namespace, filter, explain, language)
- `OllamaRequest/Response`: Estruturas de comunicação

### Arquivos Raiz

- **go.mod / go.sum**: Gerenciamento de dependências Go (Cobra, client-go, etc.)
- **docker-compose.yml**: Orquestração com profiles (default: v2, legacy: k8sgpt)
- **README.md**: Documentação principal com v2.0

### `configs/`

**Dockerfiles e entrypoints**:
- **Dockerfile.mekhanikube**: Multi-stage build Go (~80MB)
- **entrypoint-mekhanikube.sh**: Init script com health checks
- **Dockerfile.k8sgpt**: Build K8sGPT legacy (~200MB)
- **entrypoint.sh**: Init script K8sGPT legacy

### `assets/`

**Recursos estáticos**:
- **logo.png**: Logo oficial Mekhanikube (954KB, 800px width)

### `docs/`

Documentação completa do projeto:
- **ARCHITECTURE.md**: Arquitetura v2.0 com Go components
- **DEVELOPMENT.md**: Guia para desenvolvedores (Go + Docker)
- **FAQ.md**: Perguntas frequentes (v1 vs v2)
- **TROUBLESHOOTING.md**: Soluções para Mekhanikube v2 e K8sGPT legacy
- **PROJECT_STRUCTURE.md**: Este arquivo
- **PROJECT_IMPROVEMENTS.md**: Histórico de melhorias

### `scripts/`

Scripts utilitários para automação:
- **analyze.sh**: Script de análise
- **change-model.sh**: Trocar modelos Ollama
- **healthcheck.sh**: Verificação de saúde
- **release.sh**: Automação de releases
- **test.sh**: Testes automatizados

### `configs/`

Arquivos de configuração:
- **entrypoint.sh**: Script de inicialização do K8sGPT
  - Ajusta kubeconfig para Docker
  - Configura backend Ollama
  - Aguarda Ollama estar pronto

### `.github/`

Workflows GitHub Actions:
- **docker-build.yml**: Build e teste automatizados
- CI/CD para imagens Docker
- Validação de PRs

## Descrições de Arquivos

### Arquivos de Configuração

- **.env.example**: Template de variáveis de ambiente (copiar para `.env`)
- **docker-compose.yml**: Define serviços, volumes, redes
- **Dockerfile**: Build multi-estágio otimizado
- **Makefile**: Interface simplificada para comandos Docker

### Arquivos de Documentação

- **README.md**: Início rápido e visão geral
- **CHANGELOG.md**: Histórico de versões
- **CONTRIBUTING.md**: Como contribuir
- **CODE_OF_CONDUCT.md**: Padrões da comunidade
- **SECURITY.md**: Política de segurança

### Arquivos de Container

- **Dockerfile**: Build do K8sGPT da fonte oficial
- **configs/entrypoint.sh**: Configuração inicial do contêiner
- **docker-compose.yml**: Orquestração de serviços

## Decisões de Design Principais

### 1. Separação de Responsabilidades

- Configurações em `configs/`
- Scripts em `scripts/`
- Documentação em `docs/`
- Testes em `tests/`

### 2. Makefile como Interface Principal

Makefile fornece interface uniforme em todas as plataformas.

### 3. Flexibilidade de Ambiente

Arquivo `.env` permite personalização sem modificar código.

### 4. Documentação Abrangente

Documentação extensa em `docs/` para diferentes níveis de usuários.

### 5. Experiência do Desenvolvedor

- Dev containers para ambiente consistente
- Scripts automatizados
- CI/CD para garantir qualidade

## Adicionando Novos Componentes

### Novo Script

1. Criar em `scripts/`
2. Tornar executável: `chmod +x scripts/seu-script.sh`
3. Documentar no README.md

### Nova Documentação

1. Criar em `docs/`
2. Adicionar link no README.md

### Novo Teste

1. Criar em `tests/`
2. Integrar no CI/CD

### Nova Configuração

1. Adicionar em `configs/`
2. Documentar uso no README.md

## Convenções de Nomenclatura de Arquivos

- Scripts: `kebab-case.sh`
- Documentação: `UPPERCASE.md`
- Configuração: `lowercase` ou `kebab-case.yml`

## Fluxo Git

1. `main`: Branch principal (protegida)
2. `feature/*`: Novas funcionalidades
3. `fix/*`: Correções de bugs
4. `docs/*`: Atualizações de documentação

## Controle de Versão

- **VERSION**: Versionamento semântico (MAJOR.MINOR.PATCH)
- **CHANGELOG.md**: Histórico detalhado de mudanças
- **Git tags**: Tags de release (v1.0.0, v1.1.0, etc.)
