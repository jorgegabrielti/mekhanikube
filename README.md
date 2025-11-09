<div align="center"><div align="center">



# Mekhanikube 🔧# Mekhanikube 🔧



**Your Kubernetes AI Mechanic****Your Kubernetes AI Mechanic**



[![Docker Build](https://github.com/jorgegabrielti/mekhanikube/actions/workflows/docker-build.yml/badge.svg)](https://github.com/jorgegabrielti/mekhanikube/actions)[![Docker Build](https://github.com/jorgegabrielti/mekhanikube/actions/workflows/docker-build.yml/badge.svg)](https://github.com/jorgegabrielti/mekhanikube/actions)

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

[![Version](https://img.shields.io/badge/version-1.0.0-blue.svg)](https://github.com/jorgegabrielti/mekhanikube/releases)[![Version](https://img.shields.io/badge/version-1.0.0-blue.svg)](https://github.com/jorgegabrielti/mekhanikube/releases)

[![K8sGPT](https://img.shields.io/badge/K8sGPT-latest-brightgreen.svg)](https://github.com/k8sgpt-ai/k8sgpt)[![K8sGPT](https://img.shields.io/badge/K8sGPT-latest-brightgreen.svg)](https://github.com/k8sgpt-ai/k8sgpt)

[![Ollama](https://img.shields.io/badge/Ollama-latest-orange.svg)](https://ollama.ai/)[![Ollama](https://img.shields.io/badge/Ollama-latest-orange.svg)](https://ollama.ai/)



AI-powered Kubernetes cluster analysis using K8sGPT with local LLM (Ollama). Automatically diagnoses problems, explains causes, and suggests solutions.Análise inteligente de clusters Kubernetes usando K8sGPT com IA local (Ollama). Diagnostica problemas, explica causas e sugere soluções automaticamente.



[Quick Start](#-quick-start) •[Quick Start](#-quick-start) •

[Documentation](docs/) •[Documentation](docs/) •

[Contributing](CONTRIBUTING.md) •[Contributing](CONTRIBUTING.md) •

[Changelog](CHANGELOG.md)[Changelog](CHANGELOG.md)



</div></div>



------



## ✨ Features## 🚀 Quick Start



- 🤖 **AI-Powered Analysis** - Local LLM explains Kubernetes issues in plain language### Prerequisites

- 🔒 **Privacy First** - All data stays local, no external API calls

- 🐳 **Easy Setup** - Single command installation with Docker Compose- [Docker](https://docs.docker.com/get-docker/) and [Docker Compose](https://docs.docker.com/compose/install/)

- ⚡ **Fast Diagnostics** - Quickly identify and understand cluster problems- Active Kubernetes cluster with configured kubeconfig

- 🎯 **Actionable Solutions** - Get concrete steps to fix issues- At least 8GB of free disk space for AI models

- 📦 **No Kubernetes Modification** - Read-only cluster access

- 🔄 **Multiple Models** - Support for various LLM models (Gemma, Mistral, Llama2)### Installation



## 🚀 Quick Start```bash

# Clone the repository

### Prerequisitesgit clone https://github.com/jorgegabrielti/mekhanikube.git

cd mekhanikube

- [Docker](https://docs.docker.com/get-docker/) and [Docker Compose](https://docs.docker.com/compose/install/)

- Active Kubernetes cluster with configured kubeconfig# (Optional) Copy and customize environment variables

- At least 8GB of free disk space for AI modelscp .env.example .env



### Installation# Complete setup: build, start services, and install AI model

make setup

```bash

# Clone the repository# Or step by step:

git clone https://github.com/jorgegabrielti/mekhanikube.gitmake build          # Build Docker images

cd mekhanikubemake up             # Start services

make install-model  # Download AI model (gemma:7b ~5GB)

# (Optional) Copy and customize environment variables```

cp .env.example .env

### Quick Analysis

# Complete setup: build, start services, and install AI model

make setup```bash

```# Analyze your cluster with AI explanations

make analyze

### Quick Analysis

# Or using docker-compose directly:

```bashdocker exec mekhanikube-k8sgpt k8sgpt analyze --explain

# Analyze your cluster with AI explanations```

make analyze

```### Makefile Commands



### Available Commands```bash

make help           # Show all available commands

```bashmake status         # Check service status

make help           # Show all available commandsmake logs           # View logs

make status         # Check service statusmake health         # Run health checks

make logs           # View logsmake analyze-pods   # Analyze only Pods

make health         # Run health checksmake test           # Run integration tests

make test           # Run integration tests```

```

## 📋 Comandos K8sGPT

## 📋 Usage Examples

```powershell

### Basic Analysis# Analisar cluster (sem IA)

docker exec mekhanikube-k8sgpt k8sgpt analyze

```bash

# Full cluster analysis# Analisar com explicações da IA

make analyzedocker exec mekhanikube-k8sgpt k8sgpt analyze --explain



# Analyze specific namespace# Analisar namespace específico

make analyze-ns NAMESPACE=kube-systemdocker exec mekhanikube-k8sgpt k8sgpt analyze -n kube-system --explain



# Analyze only Pods# Filtrar por tipo de recurso

make analyze-podsdocker exec mekhanikube-k8sgpt k8sgpt analyze --filter=Pod --explain

docker exec mekhanikube-k8sgpt k8sgpt analyze --filter=Service --explain

# Analyze only Services

make analyze-services# Listar filtros disponíveis

```docker exec mekhanikube-k8sgpt k8sgpt filters list



### Model Management# Verificar configuração

docker exec mekhanikube-k8sgpt k8sgpt auth list

```bash```

# List installed models

make list-models## 🛠️ Configuração



# Install a different model### Modelos Ollama Recomendados

make install-model MODEL=mistral

```powershell

# Switch active model# Gemma 7B (recomendado - boa qualidade)

make change-model MODEL=mistraldocker exec mekhanikube-ollama ollama pull gemma:7b

```

# Mistral (alternativa)

### Troubleshootingdocker exec mekhanikube-ollama ollama pull mistral



```bash# TinyLlama (mais rápido, qualidade inferior)

# Check system healthdocker exec mekhanikube-ollama ollama pull tinyllama

make health```



# View logs### Trocar modelo

make logs

```powershell

# Restart services# Remover backend atual

make restartdocker exec mekhanikube-k8sgpt k8sgpt auth remove --backend localai

```

# Adicionar com novo modelo

## 🛠️ Configurationdocker exec mekhanikube-k8sgpt k8sgpt auth add --backend localai --model mistral --baseurl http://localhost:11434/v1

docker exec mekhanikube-k8sgpt k8sgpt auth default -p localai

Mekhanikube can be configured via environment variables. Copy `.env.example` to `.env` and customize:```



```bash## 📊 Arquitetura

# AI Model Configuration

OLLAMA_MODEL=gemma:7b```

OLLAMA_PORT=11434┌─────────────────┐

│   Kubernetes    │

# Kubernetes Configuration│     Cluster     │

KUBECONFIG_PATH=C:/Users/${USERNAME}/.kube/config│   (em VM/Host)  │

└────────┬────────┘

# Container Configuration         │ kubeconfig (montado em /root/.kube/)

CONTAINER_NAME_OLLAMA=mekhanikube-ollama         │

CONTAINER_NAME_K8SGPT=mekhanikube-k8sgpt    ┌────▼──────────────┐

```    │  k8sgpt container │

    │  - Ajusta config  │

### Recommended Models    │    automaticamente│

    │  - Roda análises  │

| Model | Size | Speed | Quality | Best For |    └────────┬──────────┘

|-------|------|-------|---------|----------|             │ API calls (http://localhost:11434/v1)

| **gemma:7b** | 4.8GB | Medium | Good | General use (recommended) |             │

| **mistral** | 4.1GB | Medium | Good | Detailed explanations |    ┌────────▼──────────┐

| **tinyllama** | 1.1GB | Fast | Basic | Quick scans |    │ ollama container  │

| **llama2:13b** | 7.4GB | Slow | Excellent | Best quality |    │  - Gemma:7b model │

    │  - Gera explicações│

## 📖 Documentation    └───────────────────┘

```

- 📖 **[Architecture](docs/ARCHITECTURE.md)** - System design and components

- 🔧 **[Troubleshooting](docs/TROUBLESHOOTING.md)** - Common issues and solutions## 🔧 Troubleshooting

- ❓ **[FAQ](docs/FAQ.md)** - Frequently asked questions

- 📂 **[Project Structure](docs/PROJECT_STRUCTURE.md)** - File organization### K8sGPT não consegue acessar cluster

- 🤝 **[Contributing](CONTRIBUTING.md)** - How to contribute

- 🔒 **[Security](SECURITY.md)** - Security policy```powershell

- 📝 **[Changelog](CHANGELOG.md)** - Version history# Verificar se kubeconfig está montado

docker exec mekhanikube-k8sgpt ls -la /root/.kube/

## 🔍 How It Works

# Verificar se config_mod foi criado pelo entrypoint

1. **Automatic Configuration**: Container startup script adjusts kubeconfig and configures K8sGPT backenddocker exec mekhanikube-k8sgpt cat /root/.kube/config_mod

2. **Cluster Analysis**: K8sGPT scans your Kubernetes cluster and detects issues

3. **AI Explanation**: For each issue, K8sGPT sends context to Ollama for analysis# Testar conexão manual

4. **Results**: Clear, actionable output with explanations and solutionsdocker exec mekhanikube-k8sgpt kubectl get nodes

```

## 🏗️ Architecture

### Ollama não responde

```

┌─────────────────────────────────────────────────────────────┐```powershell

│                      Host Machine                            │# Ver logs

│                                                              │docker logs mekhanikube-ollama

│  ┌────────────────────┐         ┌────────────────────┐     │

│  │  Kubernetes Cluster│         │   Docker Host      │     │# Verificar modelos instalados

│  │  - Pods            │◄────────┤                    │     │docker exec mekhanikube-ollama ollama list

│  │  - Services        │ K8s API │  ┌──────────────┐ │     │

│  │  - Deployments     │         │  │   Ollama     │ │     │# Testar API

│  └────────────────────┘         │  │   Container  │ │     │Invoke-RestMethod -Uri http://localhost:11434/v1/models | ConvertTo-Json

│           ▲                      │  │  - Gemma:7b  │ │     │

│           │ kubeconfig           │  └──────┬───────┘ │     │# Baixar modelo novamente

│           │                      │         │ HTTP    │     │docker exec mekhanikube-ollama ollama pull gemma:7b

│  ┌────────┴────────┐             │  ┌──────▼───────┐ │     │```

│  │  ~/.kube/config │             │  │   K8sGPT     │ │     │

│  └─────────────────┘             │  │   Container  │ │     │### Container k8sgpt não inicia

│                                  │  │  - Analysis  │ │     │

│                                  │  └──────────────┘ │     │```powershell

│                                  └────────────────────┘     │# Ver logs

└─────────────────────────────────────────────────────────────┘docker logs mekhanikube-k8sgpt

```

# Reconstruir imagem

## 🤝 Contributingdocker-compose build k8sgpt

docker-compose up -d k8sgpt

Contributions are welcome! Please read our [Contributing Guidelines](CONTRIBUTING.md) and [Code of Conduct](CODE_OF_CONDUCT.md).```



1. Fork the repository## 📚 Recursos

2. Create a feature branch: `git checkout -b feature/amazing-feature`

3. Make your changes and test: `make test`- [K8sGPT Docs](https://docs.k8sgpt.ai/)

4. Commit: `git commit -m 'Add amazing feature'`- [Ollama Models](https://ollama.com/library)

5. Push: `git push origin feature/amazing-feature`- [K8sGPT GitHub](https://github.com/k8sgpt-ai/k8sgpt)

6. Open a Pull Request

## � Documentation

## 📝 License

- 📖 **[Architecture](docs/ARCHITECTURE.md)** - System design and components

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.- 🔧 **[Troubleshooting](docs/TROUBLESHOOTING.md)** - Common issues and solutions

- ❓ **[FAQ](docs/FAQ.md)** - Frequently asked questions

## 🙏 Acknowledgments- 📂 **[Project Structure](docs/PROJECT_STRUCTURE.md)** - File organization

- 🤝 **[Contributing](CONTRIBUTING.md)** - How to contribute

- [K8sGPT](https://github.com/k8sgpt-ai/k8sgpt) - AI-powered Kubernetes diagnostics- 🔒 **[Security](SECURITY.md)** - Security policy

- [Ollama](https://ollama.ai/) - Local LLM inference- 📝 **[Changelog](CHANGELOG.md)** - Version history

- All contributors who help improve this project

## 🔍 How It Works

## 📬 Contact & Support

1. **Automatic Configuration**: The k8sgpt container runs `/entrypoint.sh` at startup:

- 🐛 **Bug Reports**: [GitHub Issues](https://github.com/jorgegabrielti/mekhanikube/issues)   - Copies the mounted kubeconfig from `/root/.kube/config`

- 💡 **Feature Requests**: [GitHub Issues](https://github.com/jorgegabrielti/mekhanikube/issues)   - Replaces `127.0.0.1` with `host.docker.internal` for container networking

- 💬 **Discussions**: [GitHub Discussions](https://github.com/jorgegabrielti/mekhanikube/discussions)   - Saves modified config to `/root/.kube/config_mod`

   - Sets `KUBECONFIG=/root/.kube/config_mod`

---   - Configures K8sGPT backend with Ollama



<div align="center">2. **Cluster Analysis**: K8sGPT scans your Kubernetes cluster:

   - Detects issues across Pods, Services, Deployments, etc.

Made with ❤️ for the Kubernetes community   - Identifies misconfigurations and errors

   - Collects relevant context

**[⬆ Back to Top](#mekhanikube-)**

3. **AI Explanation**: For each issue found:

</div>   - K8sGPT sends problem context to Ollama

   - LLM generates human-readable explanation
   - Suggests potential solutions

4. **Results**: Clear, actionable output with:
   - Problem description
   - AI-generated explanation
   - Suggested remediation steps

2. **Análise**: K8sGPT escaneia o cluster e identifica problemas (ConfigMaps não usados, Pods com erro, etc)

3. **Explicação**: Quando usa `--explain`, K8sGPT envia o problema para Ollama via API REST

4. **Resposta**: Ollama processa com o modelo gemma:7b e retorna explicação + solução



Se você já tem Ollama rodando:export OLLAMA_MODEL=mistral



```powershell```2. Inicie o Ollama:

# O programa detecta automaticamente

.\kube-ai.exe```bash

```

## Usodocker-compose up -d

**Nota:** Ollama é significativamente mais lento (1-2 minutos por scan).

```

---

```bash

## 🔧 Configuração Avançada

# Iniciar chat interativo3. Instale o modelo Mistral:

### Variáveis de Ambiente

./kube-ai```bash

```powershell

# Forçar uso de LocalAIdocker exec -it ollama ollama pull mistral

$env:LLM_PROVIDER="localai"

$env:LOCALAI_URL="http://localhost:8080"# Comandos disponíveis:```

$env:LOCALAI_MODEL="phi-2"

# scan    - Escanear cluster em busca de problemas

# Forçar uso de Ollama

$env:LLM_PROVIDER="ollama"# exit    - Sair do chat4. Compile e instale a CLI:

$env:OLLAMA_URL="http://localhost:11434"

$env:OLLAMA_MODEL="mistral"# qualquer texto - Fazer perguntas sobre Kubernetes```bash

```

```go install ./cmd/kube-ai

---

```

## 📊 Comparação de Performance

## Exemplos

| Provider | Modelo    | Tempo/Scan | Qualidade | RAM   |

|----------|-----------|------------|-----------|-------|## Uso

| LocalAI  | phi-2     | ~5-10s     | ⭐⭐⭐⭐    | 2GB   |

| Ollama   | mistral   | ~60-120s   | ⭐⭐⭐⭐⭐  | 4GB   |```

| Ollama   | tinyllama | ~30-60s    | ⭐⭐⭐     | 2GB   |

> scanSimplesmente execute:

**Recomendação:** Use LocalAI com phi-2 para melhor balance entre velocidade e qualidade.

🔍 Escaneando cluster...```bash

---

🤖 Analisando 2 problemas encontrados...kube-ai

## 🛠️ Troubleshooting

```

### LocalAI não inicia

> O que é um CrashLoopBackOff?

```powershell

# Verifique se o modelo foi baixado🤖 CrashLoopBackOff indica que um container está falhando...A ferramenta irá:

dir .\models\

1. Conectar ao seu cluster Kubernetes

# Verifique logs do container

docker-compose logs localai> Como debugar um pod?2. Procurar por pods com problemas



# Reinicie o serviço🤖 Use kubectl describe pod <name> para ver eventos...3. Coletar informações detalhadas

docker-compose restart

``````4. Usar IA local para analisar e sugerir soluções



### Scan muito lento

Se nenhum problema for encontrado, você verá:

- ✅ **Solução:** Use LocalAI em vez de Ollama```

- Execute: `.\download-model.ps1` e `docker-compose up -d`✅ Cluster saudável

```

### Erro de conexão com Kubernetes

Se problemas forem encontrados, você receberá uma análise detalhada com:

```powershell- Causa provável do problema

# Verifique se o cluster está acessível- Como resolver o problema

kubectl cluster-info- Como prevenir que aconteça novamente

go mod init kube-ai

# Verifique o contexto atualgo get k8s.io/client-go

kubectl config current-contextgo build -o kube-ai ./cmd/kube-ai

``````



---## Uso



## 📦 Requisitos```bash

./kube-ai

- **Go:** 1.21 ou superior```

- **Docker Desktop:** Com Kubernetes habilitado

- **RAM:** 4GB disponível## Estrutura do Projeto

- **Disco:** 2GB para modelo Phi-2

```

---kube-ai/

 ├── cmd/

## 🏗️ Arquitetura │    └── kube-ai/        # main.go, parsing de comandos CLI

 ├── internal/

``` │    ├── k8s/            # conexão + scanner

kube-ai/ │    │    ├── connect.go

├── cmd/kube-ai/          # CLI principal │    │    └── scan.go

├── internal/ │    ├── llm/            # integração com ollama

│   ├── k8s/             # Cliente Kubernetes │    │    └── ollama.go

│   └── llm/             # Cliente LLM (LocalAI/Ollama) │    └── explain/        # heurísticas e montagem de prompts

├── models/              # Modelos de IA │         └── explain.go

├── docker-compose.yml   # LocalAI setup ├── go.mod

└── download-model.ps1   # Script para baixar Phi-2 └── README.md

``````

---

## 📝 Licença

MIT


