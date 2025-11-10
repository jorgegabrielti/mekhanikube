# Integração MCP com GitHub Copilot

> 🚀 **Feature Experimental**: Conecte o GitHub Copilot diretamente ao seu cluster Kubernetes via Model Context Protocol (MCP)

## O que é MCP?

O **Model Context Protocol (MCP)** permite que o GitHub Copilot acesse dados em tempo real do seu cluster Kubernetes através do K8sGPT. Com isso, você pode conversar naturalmente com o Copilot sobre o estado do seu cluster.

## Benefícios

- 💬 **Conversação Natural**: Pergunte sobre seu cluster em linguagem natural
- 🔄 **Dados em Tempo Real**: Copilot acessa informações atualizadas do cluster
- 🇧🇷 **Respostas em Português**: Use llama3.1:8b para respostas em português
- 🎯 **Análise Inteligente**: Copilot combina conhecimento Kubernetes + estado do cluster
- 🔒 **100% Local**: Tudo roda na sua máquina, sem enviar dados externos

## Diferença entre Modo Normal e MCP

### Modo Normal (Tradicional)
```bash
# Você executa comandos manualmente
docker exec mekhanikube-k8sgpt k8sgpt analyze --explain --language Portuguese
```

### Modo MCP (Integrado ao Copilot)
```
Você no Copilot Chat: "Quais problemas existem no meu cluster?"

Copilot: "Encontrei 3 problemas no seu cluster:
1. Pod nginx-deploy-xyz está em CrashLoopBackOff..."
```

---

## Requisitos

- ✅ VS Code com extensão GitHub Copilot instalada
- ✅ Docker e Docker Compose funcionando
- ✅ Mekhanikube já configurado (modo normal)
- ✅ Cluster Kubernetes acessível

---

## Instalação

### 1. Ativar Serviço MCP

O serviço MCP é **opcional** e usa um perfil separado:

```powershell
# Iniciar APENAS o serviço MCP (além dos serviços normais)
docker-compose --profile mcp up -d
```

Isso iniciará:
- ✅ `mekhanikube-ollama` (se ainda não estiver rodando)
- ✅ `mekhanikube-k8sgpt` (modo tradicional - continua funcionando)
- ✅ `mekhanikube-k8sgpt-mcp` (novo serviço MCP na porta 3000)

### 2. Configurar GitHub Copilot

Execute o script de configuração automática:

```powershell
.\scripts\setup-mcp.ps1
```

Este script:
- Cria arquivo de configuração MCP para o VS Code
- Faz backup de configurações existentes
- Mostra próximos passos

**Caminho da configuração**: `%APPDATA%\Code\User\globalStorage\github.copilot-chat\mcpServers.json`

### 3. Reiniciar VS Code

Feche e abra o VS Code completamente para carregar a nova configuração MCP.

### 4. Verificar Conexão

Abra o **GitHub Copilot Chat** e digite:

```
@mekhanikube Você está conectado?
```

Se estiver funcionando, o Copilot responderá confirmando a conexão com o cluster.

---

## Como Usar

### Exemplos de Perguntas

**Análise Geral:**
```
Analise meu cluster Kubernetes
Existe algum problema no cluster?
Qual a saúde do cluster agora?
```

**Namespace Específico:**
```
Quais problemas existem no namespace kube-system?
Mostre o status dos pods no namespace default
```

**Recursos Específicos:**
```
Analise apenas os Pods
Tem algum Service com problema?
Verifique os Deployments
```

**Solução de Problemas:**
```
Como resolver o erro CrashLoopBackOff do pod nginx?
Por que meu Ingress não está funcionando?
Explique o problema do PersistentVolumeClaim
```

**Em Português:**
```
Explique os problemas em português brasileiro
Quero um relatório completo em PT-BR
```

---

## Configuração Avançada

### Alterar Porta MCP

Edite o arquivo `.env` (crie se não existir):

```env
MCP_PORT=3000
```

Depois reinicie:

```powershell
docker-compose --profile mcp down
docker-compose --profile mcp up -d
```

### Usar Outro Modelo

O serviço MCP usa o mesmo modelo configurado no Mekhanikube. Para trocar:

```bash
# Baixar novo modelo
docker exec mekhanikube-ollama ollama pull gemma2:9b

# Reconfigurar (afeta ambos os serviços)
docker exec mekhanikube-k8sgpt-mcp k8sgpt auth remove -b ollama
docker exec mekhanikube-k8sgpt-mcp k8sgpt auth add --backend ollama --model gemma2:9b --baseurl http://host.docker.internal:11434
docker exec mekhanikube-k8sgpt-mcp k8sgpt auth default -p ollama
```

### Logs do MCP

```powershell
# Ver logs em tempo real
docker logs -f mekhanikube-k8sgpt-mcp

# Ver últimas 50 linhas
docker logs --tail 50 mekhanikube-k8sgpt-mcp
```

---

## Troubleshooting

### Copilot não responde sobre o cluster

**Verificar se serviço MCP está rodando:**
```powershell
docker ps --filter "name=mekhanikube-k8sgpt-mcp"
```

Se não aparecer nada:
```powershell
docker-compose --profile mcp up -d
```

### Erro "MCP server not available"

1. **Reinicie o serviço MCP:**
```powershell
docker-compose --profile mcp restart
```

2. **Verifique os logs:**
```powershell
docker logs mekhanikube-k8sgpt-mcp
```

3. **Teste a porta:**
```powershell
curl http://localhost:3000/health
```

### Configuração não carrega no VS Code

1. **Verifique o arquivo de configuração:**
```powershell
cat "$env:APPDATA\Code\User\globalStorage\github.copilot-chat\mcpServers.json"
```

2. **Execute setup novamente:**
```powershell
.\scripts\setup-mcp.ps1
```

3. **Reinicie VS Code completamente** (feche todas as janelas)

### Respostas em inglês ao invés de português

O modelo precisa ser instruído. Tente:

```
@mekhanikube Responda sempre em português brasileiro. Analise meu cluster.
```

Ou configure o Copilot:
```
Configurações VS Code → GitHub Copilot → Chat: Locale → pt-BR
```

### Erro "cannot access kubeconfig"

Verifique se o caminho do kubeconfig está correto no docker-compose.yml:

```yaml
volumes:
  - C:/Users/SEU_USUARIO/.kube/config:/root/.kube/config:ro
```

Substitua `SEU_USUARIO` pelo seu nome de usuário Windows.

---

## Desativar MCP

Se quiser voltar ao modo tradicional:

```powershell
# Parar apenas o serviço MCP
docker-compose --profile mcp stop k8sgpt-mcp

# Ou remover completamente
docker-compose --profile mcp down
```

O serviço tradicional (`mekhanikube-k8sgpt`) continua funcionando normalmente.

---

## Segurança e Privacidade

- ✅ **Tudo local**: Nenhum dado sai da sua máquina
- ✅ **Sem API externa**: Não usa APIs pagas da OpenAI/Azure
- ✅ **Controle total**: Você gerencia o que o Copilot acessa
- ✅ **Cluster read-only**: K8sGPT apenas lê, nunca modifica

---

## Limitações Atuais

- ⚠️ Funciona apenas com VS Code (não funciona com Visual Studio)
- ⚠️ Requer GitHub Copilot (extensão paga)
- ⚠️ Feature experimental do K8sGPT (pode ter bugs)
- ⚠️ Não suporta múltiplos clusters simultaneamente

---

## Próximos Passos

- 🔄 Suporte a Claude Desktop (além do Copilot)
- 🌐 Multi-cluster support
- 📊 Dashboard web integrado
- 🤖 Ações automatizadas via Copilot

---

## Recursos Adicionais

- 📖 [K8sGPT MCP Documentation](https://docs.k8sgpt.ai/reference/mcp/)
- 📖 [Model Context Protocol Spec](https://modelcontextprotocol.io/)
- 💬 [Comunidade K8sGPT no Slack](https://join.slack.com/t/k8sgpt/shared_invite/zt-332vhyaxv-bfjJwHZLXWVCB3QaXafEYQ)

---

## FAQ

**P: O modo MCP substitui o modo tradicional?**  
R: Não! São complementares. Você pode usar ambos ao mesmo tempo.

**P: Preciso pagar pelo GitHub Copilot?**  
R: Sim, a feature MCP requer assinatura ativa do GitHub Copilot.

**P: Funciona com outros editores?**  
R: Atualmente apenas VS Code. Claude Desktop também suporta MCP.

**P: Posso usar MCP sem Docker?**  
R: Sim, mas precisará instalar K8sGPT localmente e ajustar a configuração manualmente.

**P: O Copilot vai modificar meu cluster?**  
R: Não! K8sGPT é read-only. Ele apenas analisa, nunca faz mudanças.

---

**🎉 Pronto!** Agora você pode conversar com o GitHub Copilot sobre seu cluster Kubernetes em tempo real!
