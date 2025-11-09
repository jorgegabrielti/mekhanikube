# Estrutura do Projeto

## Visão Geral

Este documento descreve a organização e propósito dos arquivos e diretórios no projeto Mekhanikube.

## Estrutura de Diretórios

```
mekhanikube/
├── docs/                    # Documentação
│   ├── ARCHITECTURE.md
│   ├── FAQ.md
│   ├── TROUBLESHOOTING.md
│   ├── PROJECT_STRUCTURE.md
│   └── PROJECT_IMPROVEMENTS.md
│
├── scripts/                 # Scripts utilitários
│   ├── analyze.sh
│   ├── change-model.sh
│   ├── healthcheck.sh
│   ├── release.sh
│   └── test.sh
│
├── configs/                 # Configurações
│   └── entrypoint.sh
│
├── tests/                   # Testes
│   └── integration/
│
├── .github/                 # GitHub workflows
│   └── workflows/
│       └── docker-build.yml
│
├── .devcontainer/           # Configuração Dev Container
│   └── devcontainer.json
│
├── docker-compose.yml       # Configuração principal
├── Dockerfile              # Build K8sGPT
├── .env.example            # Template de variáveis de ambiente
├── Makefile                # Comandos automatizados
│
├── README.md               # Documentação principal
├── LICENSE                 # Licença MIT
├── CHANGELOG.md            # Histórico de mudanças
├── CONTRIBUTING.md         # Guia de contribuição
├── CODE_OF_CONDUCT.md      # Código de conduta
├── SECURITY.md             # Política de segurança
└── VERSION                 # Número da versão

```

## Propósito dos Diretórios

### Arquivos Raiz

- **docker-compose.yml**: Orquestração dos serviços Ollama e K8sGPT
- **Dockerfile**: Build multi-estágio do K8sGPT
- **.env.example**: Template para configuração personalizada
- **Makefile**: Automação de comandos comuns

### `docs/`

Documentação completa do projeto:
- **ARCHITECTURE.md**: Arquitetura do sistema
- **FAQ.md**: Perguntas frequentes
- **TROUBLESHOOTING.md**: Guia de solução de problemas
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
