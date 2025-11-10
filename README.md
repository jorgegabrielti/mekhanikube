<div align="center">

#  Mekhanikube

**Seu mecânico de Kubernetes com IA**

[![Licença: MIT](https://img.shields.io/badge/Licen%C3%A7a-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Versão](https://img.shields.io/badge/vers%C3%A3o-1.0.0-blue.svg)](https://github.com/jorgegabrielti/mekhanikube/releases)

Análise inteligente de clusters Kubernetes usando **K8sGPT** + **Ollama**  
Totalmente local • Privado • Fácil de usar

[Começar](#-início-rápido) • [Documentação](docs/) • [Contribuir](CONTRIBUTING.md)

</div>

---

##  O que faz?

Escaneia seu cluster Kubernetes, identifica problemas e **explica em linguagem simples** usando IA local.

```bash
# Execute uma análise
docker exec mekhanikube-k8sgpt k8sgpt analyze --explain
```

**Exemplo de saída:**
```
0: Pod default/nginx-5d5d5d5d-xxx
- Error: CrashLoopBackOff
- IA: Este pod está reiniciando continuamente porque o comando de entrada 
  está falhando. Verifique os logs com kubectl logs e corrija o comando.
```

---

##  Início Rápido

### Pré-requisitos
- Docker & Docker Compose
- Cluster Kubernetes ativo
- ~10GB de espaço livre

### Instalação

```bash
# 1. Clone o repositório
git clone https://github.com/jorgegabrielti/mekhanikube.git
cd mekhanikube

# 2. Inicie os serviços
docker-compose up -d

# 3. Baixe o modelo de IA (primeira vez - ~4.7GB)
docker exec mekhanikube-ollama ollama pull llama3.1:8b

# 4. Configure o backend
docker exec mekhanikube-k8sgpt k8sgpt auth add --backend ollama --model llama3.1:8b --baseurl http://localhost:11434
docker exec mekhanikube-k8sgpt k8sgpt auth default -p ollama

# 5. Analisar cluster
docker exec mekhanikube-k8sgpt k8sgpt analyze --explain --language Portuguese
```

---

##  Comandos Úteis

```bash
# Análise completa em português
docker exec mekhanikube-k8sgpt k8sgpt analyze --explain --language Portuguese

# Análise completa (inglês)
docker exec mekhanikube-k8sgpt k8sgpt analyze --explain

# Analisar namespace específico
docker exec mekhanikube-k8sgpt k8sgpt analyze -n kube-system --explain --language Portuguese

# Filtrar por tipo (Pod, Service, Deployment, etc)
docker exec mekhanikube-k8sgpt k8sgpt analyze --filter=Pod --explain --language Portuguese

# Listar modelos instalados
docker exec mekhanikube-ollama ollama list

# Ver status dos containers
docker-compose ps
```

---

##  Modelos Disponíveis

| Modelo | Tamanho | Velocidade | Qualidade | Português | Recomendado para |
|--------|---------|------------|-----------|-----------|------------------|
| **llama3.1:8b** ⭐ | 4.7GB | Bom | Excelente | ⭐⭐⭐⭐⭐ | **Recomendado (PT-BR)** |
| **gemma2:9b** | 5.4GB | Médio | Excelente | ⭐⭐⭐⭐⭐ | Melhor qualidade |
| **qwen2.5:7b** | 4.7GB | Rápido | Muito Boa | ⭐⭐⭐⭐ | Velocidade |
| **mistral** | 4.1GB | Médio | Boa | ⭐⭐⭐ | Uso geral |
| **tinyllama** | 1.1GB | Muito Rápido | Básica | ⭐⭐ | Scans rápidos |

> 💡 **llama3.1:8b** é o modelo padrão por oferecer excelente suporte ao português brasileiro

**Trocar modelo:**
```bash
# Instalar outro modelo
docker exec mekhanikube-ollama ollama pull gemma2:9b

# Reconfigurar K8sGPT
docker exec mekhanikube-k8sgpt k8sgpt auth remove --backend ollama
docker exec mekhanikube-k8sgpt k8sgpt auth add --backend ollama --model gemma2:9b --baseurl http://localhost:11434
docker exec mekhanikube-k8sgpt k8sgpt auth default -p ollama
```

---

## 🚀 Feature Experimental: Integração MCP

> **Novidade!** Conecte o **GitHub Copilot** diretamente ao seu cluster Kubernetes

### O que é?

Converse com o GitHub Copilot sobre seu cluster em **tempo real**:

```
Você: "Quais problemas existem no meu cluster?"
Copilot: "Encontrei 3 problemas: Pod nginx em CrashLoopBackOff..."
```

### Como ativar?

```powershell
# 1. Ativar serviço MCP
docker-compose --profile mcp up -d

# 2. Configurar GitHub Copilot
.\scripts\setup-mcp.ps1

# 3. Reiniciar VS Code

# 4. Conversar com Copilot!
```

### Benefícios

- 💬 Conversação natural sobre o cluster
- 🔄 Dados em tempo real
- 🇧🇷 Respostas em português
- 🔒 100% local e privado

📖 **[Documentação completa da integração MCP →](docs/MCP.md)**

---

##  Solução de Problemas

**Container não inicia?**
```bash
docker-compose logs
```

**Ollama não responde?**
```bash
docker logs mekhanikube-ollama
docker exec mekhanikube-ollama ollama list
```

**K8sGPT não acessa o cluster?**
```bash
docker exec mekhanikube-k8sgpt kubectl get nodes
docker exec mekhanikube-k8sgpt cat /root/.kube/config_mod
```

---

##  Documentação

-  [🚀 Integração MCP (GitHub Copilot)](docs/MCP.md) - Conecte Copilot ao cluster
-  [Arquitetura](docs/ARCHITECTURE.md) - Como funciona internamente
-  [Solução de Problemas](docs/TROUBLESHOOTING.md) - Problemas comuns e soluções
-  [Perguntas Frequentes](docs/FAQ.md) - Dúvidas mais comuns
-  [Como Contribuir](CONTRIBUTING.md) - Guia para contribuições

---

##  Licença

Licença MIT - consulte o arquivo [LICENSE](LICENSE) para mais detalhes.

---

##  Créditos

- [K8sGPT](https://github.com/k8sgpt-ai/k8sgpt) - Ferramenta de análise de clusters Kubernetes
- [Ollama](https://ollama.ai/) - Plataforma de modelos de linguagem locais

---

<div align="center">

**Feito com  para a comunidade Kubernetes**

</div>
