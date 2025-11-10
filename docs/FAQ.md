# Perguntas Frequentes (FAQ)

## Questões Gerais

### O que é o Mekhanikube?

Mekhanikube é uma solução containerizada **própria** desenvolvida em Go que analisa clusters Kubernetes e fornece explicações via IA local (Ollama). Versão 2.0 traz engine customizado que substitui o K8sGPT por uma solução 60% mais leve e 3x mais rápida.

### Qual a diferença entre v1 e v2?

**Mekhanikube v2.0 (atual)**:
- ✅ Engine próprio em Go (1.618 linhas)
- ✅ Imagem ~80MB (60% menor)
- ✅ Startup <10s (3x mais rápido)
- ✅ Configuração automática (zero setup)
- ✅ Suporte nativo ao português

**K8sGPT (v1 - legado)**:
- Ferramenta externa
- Imagem ~200MB
- Startup ~30s
- Requer configuração manual
- Disponível via `--profile k8sgpt`

### Por que "Mekhanikube"?

**Mekhani** (Grego: μηχανικός) = mecânico + **kube** (Kubernetes) = Seu mecânico Kubernetes!

### É gratuito?

Sim! Mekhanikube é código aberto sob a Licença MIT. Ollama também é gratuito e de código aberto.

### Ele envia meus dados para algum lugar?

Não! Tudo roda 100% localmente na sua máquina. Os dados do seu cluster nunca saem da sua infraestrutura. Sem telemetria, sem chamadas de API externas.

---

## Instalação & Configuração

### Quais são os requisitos do sistema?

**Mínimo (v2.0)**:
- Docker & Docker Compose
- 1 núcleo de CPU
- 2GB RAM
- 5GB de espaço em disco
- Cluster Kubernetes ativo

**Recomendado**:
- 2-4 núcleos de CPU
- 4-8GB RAM
- 10GB de espaço em disco (múltiplos modelos)

### Quais sistemas operacionais são suportados?

- ✅ Windows 10/11 (com Docker Desktop)
- ✅ macOS (Intel & Apple Silicon)
- ✅ Linux (qualquer distribuição com Docker)

### Posso usar com qualquer cluster Kubernetes?

Sim! Mekhanikube funciona com:
- Clusters locais (Docker Desktop, Minikube, Kind)
- Clusters na nuvem (EKS, GKE, AKS)
- Clusters on-premise
- Qualquer cluster acessível via kubeconfig

### Quanto tempo leva a configuração?

**Mekhanikube v2.0**:
- Primeira vez: ~10-15 minutos (incluindo download do modelo)
- Inicializações subsequentes: <10 segundos
- Mudanças de modelo: ~5-10 minutos por modelo

**K8sGPT (legado)**:
- Primeira vez: ~15-20 minutos
- Inicializações subsequentes: ~30 segundos
- Requer configuração manual do backend

---

## Questões de Uso

### Qual modelo de IA devo usar?

| Modelo | Melhor Para | Velocidade | Qualidade | Português | Tamanho |
|--------|-------------|------------|-----------|-----------|---------|
| **llama3.1:8b** ⭐ | **Recomendado (PT-BR)** | Boa | Excelente | ⭐⭐⭐⭐⭐ | 4.7GB |
| **gemma2:9b** | Melhor qualidade | Média | Excelente | ⭐⭐⭐⭐⭐ | 5.4GB |
| **qwen2.5:7b** | Velocidade | Rápida | Muito Boa | ⭐⭐⭐⭐ | 4.7GB |
| **mistral** | Uso geral | Média | Boa | ⭐⭐⭐ | 4.1GB |
| **tinyllama** | Varreduras rápidas | Muito Rápida | Básica | ⭐⭐ | 1.1GB |

Comece com `llama3.1:8b` - oferece excelente suporte ao português brasileiro.

### Posso usar múltiplos modelos?

Sim! Instale múltiplos modelos:

```bash
# Instalar modelos adicionais
docker exec mekhanikube-ollama ollama pull gemma2:9b
docker exec mekhanikube-ollama ollama pull mistral

# Mekhanikube v2 usa automaticamente o modelo disponível
# Para K8sGPT, reconfigure o backend:
docker exec mekhanikube-k8sgpt k8sgpt auth add --backend ollama --model gemma2:9b --baseurl http://localhost:11434
```

### Com que frequência devo executar a análise?

**Cronograma recomendado**:
- **Após deployments**: Verificar problemas imediatamente
- **Diariamente**: Verificações rotineiras de saúde
- **Antes de releases**: Validar estado do cluster
- **Quando alertas disparam**: Investigar causa raiz

### Posso analisar apenas recursos específicos?

Sim! Use filtros:

**Mekhanikube v2.0**:
```bash
# Analisar apenas Pods
docker exec mekhanikube mekhanikube analyze --filter Pod --explain --language Portuguese

# Analisar apenas ConfigMaps
docker exec mekhanikube mekhanikube analyze --filter ConfigMap --explain --language Portuguese

# Namespace específico
docker exec mekhanikube mekhanikube analyze -n production --explain --language Portuguese
```

**K8sGPT (legado)**:
```bash
# Com profile k8sgpt
docker exec mekhanikube-k8sgpt k8sgpt analyze --filter=Pod --explain --language Portuguese

# Listar filtros disponíveis
docker exec mekhanikube-k8sgpt k8sgpt filters list
```

> 💡 Mekhanikube v2 tem suporte nativo ao português, mas você pode especificar `--language Portuguese` ou `--language English`.

### Que tipos de problemas ele pode detectar?

**Mekhanikube v2.0 detecta**:
- **Pods**: 
  - CrashLoopBackOff
  - ImagePullBackOff
  - ContainerStatusUnknown
  - Containers terminados
- **ConfigMaps**: 
  - ConfigMaps não utilizados

**K8sGPT (legado) analisa**:
- **Pods**, **Services**, **Deployments**, **PVCs**, **Ingress**
- **StatefulSets**, **HPA**, **NetworkPolicies**
- E mais tipos de recursos

> 💡 Mekhanikube v2 é focado nos problemas mais comuns (Pods e ConfigMaps). Novos scanners podem ser adicionados facilmente.

---

## Questões Técnicas

### Como funciona?

**Mekhanikube v2.0**:
1. CLI recebe comando `mekhanikube analyze`
2. Scanner conecta à API Kubernetes via client-go
3. Detecta problemas em Pods e ConfigMaps
4. Analyzer aplica filtros (se especificados)
5. Se `--explain`, envia para Ollama via HTTP
6. Ollama (llama3.1:8b) gera explicação em português
7. CLI exibe resultados formatados

**K8sGPT (legado)**:
1. K8sGPT escaneia cluster via API Kubernetes
2. Analisadores identificam problemas
3. Envia contexto para Ollama
4. LLM gera explicação
5. Resultados exibidos

### Ele modifica meu cluster?

**Não!** Mekhanikube é somente leitura. Ele:
- ✅ Lê o estado do cluster
- ✅ Analisa configurações
- ✅ Gera relatórios
- ❌ Nunca faz mudanças
- ❌ Nunca deleta recursos
- ❌ Nunca aplica configurações

### Quais permissões ele precisa?

K8sGPT requer acesso **somente leitura** aos recursos do cluster. As mesmas permissões dos comandos `kubectl get`.

### Posso executar em CI/CD?

Sim! Exemplo:

```yaml
# GitLab CI - Mekhanikube v2
k8s-analysis:
  script:
    - docker-compose up -d
    - docker exec mekhanikube mekhanikube analyze --explain --language Portuguese > report.txt
  artifacts:
    paths:
      - report.txt

# GitLab CI - K8sGPT legado
k8s-analysis-legacy:
  script:
    - docker-compose --profile k8sgpt up -d
    - docker exec mekhanikube-k8sgpt k8sgpt analyze --explain --language Portuguese > report.txt
  artifacts:
    paths:
      - report.txt
```

### As análises são sempre em português?

**Mekhanikube v2.0**: Suporte nativo ao português! Basta usar `--language Portuguese` (ou omitir para inglês).

```bash
# Português (recomendado)
docker exec mekhanikube mekhanikube analyze --explain --language Portuguese

# Inglês
docker exec mekhanikube mekhanikube analyze --explain --language English
```

**K8sGPT (legado)**: Requer flag `--language Portuguese` explicitamente.

**Idiomas suportados**: English, Portuguese

> ⭐ O modelo **llama3.1:8b** oferece excelente qualidade em português brasileiro!

### Posso exportar resultados?

Sim! Redirecione a saída:

```bash
# Mekhanikube v2 - Salvar em arquivo
docker exec mekhanikube mekhanikube analyze --explain --language Portuguese > analysis.txt

# K8sGPT - JSON
docker exec mekhanikube-k8sgpt k8sgpt analyze --explain --output json --language Portuguese > analysis.json
```

---

## Solução de Problemas

### Por que está lento?

**Possíveis causas**:
1. **Modelo grande**: Tente `tinyllama` para respostas mais rápidas
2. **Muitos recursos**: Use filtros ou escopo de namespace
3. **RAM limitada**: Aloque mais memória para o Docker
4. **Gargalo de CPU**: Feche outras aplicações

**Otimização**:
```bash
# Usar modelo menor
docker exec mekhanikube-ollama ollama pull tinyllama

# Limitar escopo
docker exec mekhanikube-k8sgpt k8sgpt analyze --namespace default --explain
docker exec mekhanikube-k8sgpt k8sgpt analyze --filter=Pod --explain
```

### Diz "nenhum problema encontrado" mas sei que há problemas

1. **Verificar namespace**: Padrão é todos os namespaces
   ```bash
   docker exec mekhanikube-k8sgpt k8sgpt analyze --namespace seu-namespace --explain
   ```

2. **Tentar filtros diferentes**: Alguns problemas precisam de analisadores específicos
   ```bash
   docker exec mekhanikube-k8sgpt k8sgpt filters list
   docker exec mekhanikube-k8sgpt k8sgpt analyze --filter=Pod --explain
   ```

3. **Verificar acesso ao cluster**:
   ```bash
   docker exec mekhanikube-k8sgpt kubectl get pods --all-namespaces
   ```

### Ollama continua baixando modelos

Modelos são armazenados em volumes Docker. Se você executar `docker-compose down -v`, modelos são deletados.

**Preservar modelos**:
```bash
# Parar sem remover volumes
docker-compose down

# Ou apenas reiniciar
docker-compose restart
```

### Posso usar uma instância Ollama externa?

Sim! Modifique o `docker-compose.yml`:

```yaml
k8sgpt:
  environment:
    - OLLAMA_BASEURL=http://seu-servidor-ollama:11434
```

Então remova a definição do serviço Ollama.

---

## Uso Avançado

### Posso personalizar os analisadores do K8sGPT?

K8sGPT usa analisadores integrados. Para habilitar/desabilitar:

```bash
# Listar filtros disponíveis
docker exec mekhanikube-k8sgpt k8sgpt filters list

# Usar filtros específicos
docker exec mekhanikube-k8sgpt k8sgpt analyze --filter=Pod,Service --explain
```

### Posso usar um backend LLM diferente?

Sim! K8sGPT suporta:
- Ollama (local) - padrão
- OpenAI (nuvem)
- Azure OpenAI (nuvem)
- LocalAI (alternativa local)

Exemplo para OpenAI:
```bash
docker exec mekhanikube-k8sgpt k8sgpt auth add \
  --backend openai \
  --model gpt-4 \
  --password SUA_API_KEY
```

### Como faço backup da minha configuração?

```bash
# Backup dos modelos Ollama
docker run --rm \
  -v mekhanikube-ollama-data:/data \
  -v ${PWD}:/backup \
  alpine tar czf /backup/ollama-backup.tar.gz /data

# Backup da config K8sGPT
docker run --rm \
  -v mekhanikube-k8sgpt-config:/data \
  -v ${PWD}:/backup \
  alpine tar czf /backup/k8sgpt-backup.tar.gz /data
```

### Posso executar múltiplas instâncias?

Sim, mas altere os nomes dos contêineres para evitar conflitos:

```bash
# No arquivo .env
CONTAINER_NAME_OLLAMA=mekhanikube-ollama-2
CONTAINER_NAME_K8SGPT=mekhanikube-k8sgpt-2
OLLAMA_PORT=11435
```

### Como atualizo para a versão mais recente?

```bash
# Puxar código mais recente
git pull origin main

# Reconstruir contêineres
docker-compose build

# Reiniciar serviços
docker-compose restart
```

---

## Performance & Otimização

### Quanto espaço em disco preciso?

- **Instalação base**: ~500MB (contêineres)
- **Por modelo**: 1-10GB dependendo do modelo
- **Logs**: ~100MB (cresce com o tempo)
- **Recomendação**: 20GB de espaço livre

### Posso limitar o uso de recursos?

Sim! Edite o `docker-compose.yml`:

```yaml
services:
  ollama:
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 4G
```

### Qual modelo é mais rápido?

Ranking de velocidade (mais rápido para mais lento):
1. `tinyllama` - ~2-5 segundos por explicação
2. `gemma:7b` - ~5-10 segundos por explicação
3. `mistral` - ~8-15 segundos por explicação
4. `llama2:13b` - ~15-30 segundos por explicação

### Posso usar aceleração GPU?

Sim, se você tiver GPU NVIDIA:

1. Instale o [NVIDIA Container Toolkit](https://github.com/NVIDIA/nvidia-docker)
2. Modifique o `docker-compose.yml`:

```yaml
services:
  ollama:
    deploy:
      resources:
        reservations:
          devices:
            - driver: nvidia
              count: 1
              capabilities: [gpu]
```

---

## Segurança & Privacidade

### Meu kubeconfig está seguro?

Sim:
- Montado como **somente leitura** no contêiner
- Nunca modificado ou exposto
- Usado apenas para acesso à API
- Contêiner está isolado

### Quais dados são coletados?

**Nenhum!** Mekhanikube:
- ❌ Sem telemetria
- ❌ Sem analytics
- ❌ Sem conexões externas (exceto downloads de modelos)
- ❌ Sem compartilhamento de dados

### Posso usar em produção?

Sim, mas:
- ✅ É somente leitura (seguro)
- ✅ Sem modificações no cluster
- ⚠️ Garanta recursos adequados
- ⚠️ Teste em dev/staging primeiro
- ⚠️ Monitore o uso de recursos

### Devo fazer commit do meu arquivo .env?

**NÃO!** O arquivo `.env` pode conter informações sensíveis. Já está no `.gitignore`.

---

## Contribuindo & Suporte

### Como posso contribuir?

Veja [CONTRIBUTING.md](../CONTRIBUTING.md) para:
- Reportar bugs
- Sugerir funcionalidades
- Enviar pull requests
- Melhorar a documentação

### Onde reporto bugs?

Abra uma issue no [GitHub Issues](https://github.com/jorgegabrielti/mekhanikube/issues) com:
- SO e versão do Docker
- Saída de `docker-compose ps`
- Passos para reproduzir
- Mensagens de erro/logs

### Posso solicitar novas funcionalidades?

Sim! Abra uma GitHub Issue com:
- Descrição da funcionalidade
- Caso de uso
- Comportamento esperado
- Exemplo de uso

### Como obtenho ajuda?

1. Verifique este FAQ
2. Leia [TROUBLESHOOTING.md](TROUBLESHOOTING.md)
3. Pesquise [issues existentes](https://github.com/jorgegabrielti/mekhanikube/issues)
4. Abra uma nova issue
5. Participe das discussões

---

## Roadmap & Futuro

### O que está planejado para versões futuras?

- Dashboard Web UI
- Integrações Slack/Teams
- Plugins de analisador customizados
- Rastreamento de análise histórica
- Modo operador Kubernetes
- Suporte multi-cluster

### Posso patrocinar o projeto?

Ainda não, mas fique ligado! Enquanto isso, contribuições e estrelas no GitHub são apreciadas! ⭐

---

## Comparação com Outras Ferramentas

### Mekhanikube vs kubectl

- **kubectl**: Comandos de baixo nível, interpretação manual
- **Mekhanikube**: Análise automatizada com explicações de IA

### Mekhanikube vs K9s

- **K9s**: TUI interativa para gerenciamento de cluster
- **Mekhanikube**: Detecção automatizada de problemas com IA

### Mekhanikube vs Lens

- **Lens**: IDE desktop GUI para Kubernetes
- **Mekhanikube**: Ferramenta CLI com análise de IA

### Mekhanikube vs Prometheus/Grafana

- **Prometheus/Grafana**: Métricas e monitoramento
- **Mekhanikube**: Detecção e explicação de problemas

**Eles se complementam!** Use Mekhanikube para diagnósticos junto com suas ferramentas existentes.

---

## Recursos Adicionais

- 📖 [Documentação de Arquitetura](ARCHITECTURE.md)
- 🔧 [Guia de Solução de Problemas](TROUBLESHOOTING.md)
- 🤝 [Diretrizes de Contribuição](../CONTRIBUTING.md)
- 📝 [Histórico de Mudanças](../CHANGELOG.md)
- 🐙 [Repositório GitHub](https://github.com/jorgegabrielti/mekhanikube)
- 🔗 [Documentação K8sGPT](https://docs.k8sgpt.ai/)
- 🦙 [Documentação Ollama](https://github.com/ollama/ollama)

---

**Não encontrou sua resposta?** Abra uma issue no GitHub!
