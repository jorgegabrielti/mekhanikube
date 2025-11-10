<div align="center">

# 🔧 Mekhanikube

**Seu mecânico de Kubernetes com IA**

[![Licença: MIT](https://img.shields.io/badge/Licen%C3%A7a-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Versão](https://img.shields.io/badge/vers%C3%A3o-2.0.0-blue.svg)](https://github.com/jorgegabrielti/mekhanikube/releases)
[![Go](https://img.shields.io/badge/Go-1.21-00ADD8?logo=go)](https://golang.org/)

Análise inteligente de clusters Kubernetes com **engine próprio** + **Ollama**  
Totalmente local • Privado • Leve • Rápido

[Começar](#-início-rápido) • [Documentação](docs/) • [Contribuir](CONTRIBUTING.md)

</div>

---

## 🎯 O que faz?

Escaneia seu cluster Kubernetes, identifica problemas e **explica em linguagem simples** usando IA local (llama3.1:8b).

```bash
# Execute uma análise completa em português
docker exec mekhanikube mekhanikube analyze --explain --language Portuguese
```

**Exemplo de saída:**
```
🔍 Encontrados 2 problema(s):

0: Pod default/nginx-deployment-xxx
- Error: Container nginx in CrashLoopBackOff
- IA: O container está reiniciando continuamente. Isso geralmente 
  acontece quando o comando de entrada falha ou o aplicativo trava 
  logo após iniciar. Verifique os logs com kubectl logs e corrija 
  o problema no código ou configuração.
  
1: ConfigMap kube-system/unused-config
- Error: ConfigMap unused-config is not used by any pods
- IA: Este ConfigMap existe mas não está sendo usado por nenhum pod.
  Você pode removê-lo com segurança se não for mais necessário.
```

---

## ⚡ Por que Mekhanikube v2?

### Engine Próprio vs K8sGPT

| Característica | K8sGPT (v1) | Mekhanikube (v2) |
|----------------|-------------|------------------|
| **Tamanho da imagem** | ~200MB | **~80MB** 🎯 |
| **Startup** | ~30s | **<10s** ⚡ |
| **Performance** | Boa | **Excelente** 🚀 |
| **Código** | Dependência externa | **100% próprio** 💪 |
| **Controle** | Limitado | **Total** ✅ |
| **Manutenção** | Depende de updates externos | **Independente** 🔧 |
| **Extensibilidade** | Moderada | **Total** 🎨 |

### ✨ Vantagens

- 🔥 **60% mais leve** que a solução anterior
- ⚡ **3x mais rápido** no startup
- 🛠️ **Código próprio** - controle total sobre features
- 🇧🇷 **Português nativo** - melhor suporte ao idioma
- 🎯 **Focado** - apenas o essencial, sem bloat
- 🔒 **Privado** - tudo roda local, zero cloud

---

## 🚀 Início Rápido

### Pré-requisitos
- Docker & Docker Compose
- Cluster Kubernetes ativo (local ou remoto)
- ~5GB de espaço livre (modelo IA: 4.7GB)

### Instalação

```bash
# 1. Clone o repositório
git clone https://github.com/jorgegabrielti/mekhanikube.git
cd mekhanikube

# 2. Inicie os serviços (Ollama + Mekhanikube)
docker-compose up -d

# 3. Baixe o modelo de IA (primeira vez - ~4.7GB)
docker exec mekhanikube-ollama ollama pull llama3.1:8b

# 4. Pronto! Analise seu cluster
docker exec mekhanikube mekhanikube analyze --explain --language Portuguese
```

---

## 📖 Comandos

### Análise Básica (sem IA)
```bash
# Análise rápida - detecta problemas sem explicações
docker exec mekhanikube mekhanikube analyze

# Namespace específico
docker exec mekhanikube mekhanikube analyze -n kube-system

# Filtrar por tipo de recurso
docker exec mekhanikube mekhanikube analyze --filter Pod
docker exec mekhanikube mekhanikube analyze --filter ConfigMap
```

### Análise com IA (recomendado)
```bash
# Análise completa em português
docker exec mekhanikube mekhanikube analyze --explain --language Portuguese

# Análise em inglês
docker exec mekhanikube mekhanikube analyze --explain --language English

# Combinando filtros
docker exec mekhanikube mekhanikube analyze \
  --explain \
  --language Portuguese \
  --namespace default \
  --filter Pod
```

### Flags Disponíveis

| Flag | Descrição | Padrão |
|------|-----------|--------|
| `-n, --namespace` | Namespace específico | todos |
| `-f, --filter` | Filtrar por tipo (Pod, ConfigMap) | todos |
| `-e, --explain` | Explicar com IA | false |
| `-l, --language` | Idioma (Portuguese, English) | Portuguese |
| `--no-cache` | Forçar análise sem cache | false |

### Outros Comandos
```bash
# Ver versão
docker exec mekhanikube mekhanikube version

# Verificar status dos containers
docker-compose ps

# Ver logs do Mekhanikube
docker logs mekhanikube

# Ver logs do Ollama
docker logs mekhanikube-ollama
```

---

## 🤖 Modelos de IA

| Modelo | Tamanho | Velocidade | Qualidade | Português | Recomendado para |
|--------|---------|------------|-----------|-----------|------------------|
| **llama3.1:8b** ⭐ | 4.7GB | Bom | Excelente | ⭐⭐⭐⭐⭐ | **Padrão (PT-BR)** |
| **gemma2:9b** | 5.4GB | Médio | Excelente | ⭐⭐⭐⭐⭐ | Melhor qualidade |
| **qwen2.5:7b** | 4.7GB | Rápido | Muito Boa | ⭐⭐⭐⭐ | Velocidade |
| **mistral** | 4.1GB | Médio | Boa | ⭐⭐⭐ | Uso geral |
| **tinyllama** | 1.1GB | Muito Rápido | Básica | ⭐⭐ | Scans rápidos |

### Trocar Modelo

```bash
# 1. Baixar novo modelo
docker exec mekhanikube-ollama ollama pull gemma2:9b

# 2. Usar nas análises
docker exec mekhanikube mekhanikube analyze --explain --language Portuguese
# (O Mekhanikube usa o modelo configurado no docker-compose.yml)

# 3. Para mudar permanentemente, edite docker-compose.yml:
# OLLAMA_MODEL=gemma2:9b
```

---

## 🔍 Tipos de Problemas Detectados

### ✅ Implementado

- **Pods**
  - CrashLoopBackOff
  - ImagePullBackOff / ErrImagePull
  - Container terminated with error
  - Pending state (scheduling issues)
  - Failed pods

- **ConfigMaps**
  - ConfigMaps não utilizados por pods
  - Recursos órfãos

### 🚧 Em Desenvolvimento (próximas versões)

- **Services** - endpoints não disponíveis
- **Deployments** - replicas inconsistentes
- **StatefulSets** - problemas de persistência
- **PersistentVolumeClaims** - storage issues
- **Ingress** - problemas de roteamento
- **Resource Limits** - uso excessivo de recursos

---

## 🛠️ Arquitetura

```
┌─────────────────────────────────────────────┐
│           Mekhanikube v2                    │
│                                             │
│  ┌─────────────┐      ┌─────────────┐     │
│  │   Scanner   │─────▶│  Analyzer   │     │
│  │  (K8s API)  │      │   (Logic)   │     │
│  └─────────────┘      └──────┬──────┘     │
│                              │             │
│                              ▼             │
│                      ┌─────────────┐       │
│                      │   Ollama    │       │
│                      │   Client    │       │
│                      └──────┬──────┘       │
└─────────────────────────────┼──────────────┘
                              │
                              ▼
                    ┌──────────────────┐
                    │  Ollama Server   │
                    │  (llama3.1:8b)   │
                    └──────────────────┘
```

### Componentes

- **Scanner** (`internal/scanner`) - Conecta ao cluster via client-go, coleta recursos
- **Analyzer** (`internal/analyzer`) - Lógica de detecção de problemas
- **Ollama Client** (`internal/ollama`) - Comunicação HTTP com Ollama para explicações
- **CLI** (`cmd/mekhanikube`) - Interface Cobra para linha de comando
- **Types** (`pkg/types`) - Estruturas de dados compartilhadas

---

## 🐳 Docker Compose

O projeto usa **profiles** para permitir escolher entre K8sGPT (v1) ou Mekhanikube (v2):

```yaml
# Usar Mekhanikube v2 (padrão - recomendado)
docker-compose up -d

# Usar K8sGPT v1 (legado)
docker-compose --profile k8sgpt up -d
```

---

## 🔧 Solução de Problemas

**Container não inicia?**
```bash
docker-compose logs mekhanikube
```

**Ollama não responde?**
```bash
docker logs mekhanikube-ollama
docker exec mekhanikube-ollama ollama list
```

**Mekhanikube não acessa o cluster?**
```bash
docker exec mekhanikube kubectl get nodes
docker exec mekhanikube cat /root/.kube/config_mod
```

**Timeout ao gerar explicações?**
```bash
# Modelos grandes podem demorar. Tente um modelo menor:
docker exec mekhanikube-ollama ollama pull tinyllama
```

---

## 📚 Documentação

- 📖 [Arquitetura](docs/ARCHITECTURE.md) - Como funciona internamente
- 🔧 [Solução de Problemas](docs/TROUBLESHOOTING.md) - Problemas comuns
- ❓ [FAQ](docs/FAQ.md) - Perguntas frequentes
- 🤝 [Contribuir](CONTRIBUTING.md) - Como contribuir com o projeto
- 👨‍💻 [Desenvolvimento](docs/DEVELOPMENT.md) - Guia para desenvolvedores

---

## 🗺️ Roadmap

### v2.1 (próximo)
- [ ] Scanner para Services
- [ ] Scanner para Deployments
- [ ] Scanner para StatefulSets
- [ ] Suporte a output JSON/YAML
- [ ] Cache de resultados

### v2.2
- [ ] Interface web simples
- [ ] Relatórios em HTML
- [ ] Integração com Slack/Discord
- [ ] Métricas e dashboard

### v3.0 (futuro)
- [ ] Análise preditiva
- [ ] Auto-remediation (correção automática)
- [ ] Multi-cluster support
- [ ] Plugin system

---

## 🤝 Contribuindo

Contribuições são bem-vindas! Veja [CONTRIBUTING.md](CONTRIBUTING.md) para detalhes.

### Desenvolvimento Local

```bash
# Clone o repo
git clone https://github.com/jorgegabrielti/mekhanikube.git
cd mekhanikube

# Build local
go build -o mekhanikube ./cmd/mekhanikube

# Rodar testes
go test ./...

# Build Docker
docker-compose build mekhanikube
```

---

## 📝 Licença

Licença MIT - consulte o arquivo [LICENSE](LICENSE) para mais detalhes.

---

## 🙏 Créditos

- [Ollama](https://ollama.ai/) - Plataforma de modelos de linguagem locais
- [Kubernetes client-go](https://github.com/kubernetes/client-go) - Cliente oficial do Kubernetes
- [Cobra](https://github.com/spf13/cobra) - Framework CLI
- Inspirado por [K8sGPT](https://github.com/k8sgpt-ai/k8sgpt)

---

<div align="center">

**Feito com ❤️ para a comunidade Kubernetes**

⭐ Se este projeto foi útil, considere dar uma estrela!

</div>
