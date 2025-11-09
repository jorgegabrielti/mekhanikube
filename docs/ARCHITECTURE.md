# Arquitetura

## Visão Geral do Sistema

Mekhanikube é uma solução containerizada que combina K8sGPT e Ollama para fornecer análise de clusters Kubernetes alimentada por IA.

## Componentes Principais

### Contêiner K8sGPT

**Função**: Analisa o cluster Kubernetes e identifica problemas

**Recursos Principais**:
- Lê kubeconfig do volume montado
- Ajusta automaticamente configuração para rede do contêiner  
- Suporta múltiplos tipos de recursos (Pods, Services, Deployments, etc.)
- Análise com filtros e escopo de namespace
- Comunica com Ollama para explicações de IA

### Contêiner Ollama

**Função**: Executa modelos LLM localmente para gerar explicações

**Recursos Principais**:
- API REST na porta 11434
- Armazenamento persistente de modelos
- Suporta múltiplos modelos (Gemma, Mistral, Llama2, etc.)
- Compatível com formato de API OpenAI

## Fluxo de Dados

1. **Solicitação de Análise**: Usuário → K8sGPT
2. **Varredura do Cluster**: K8sGPT → API Kubernetes → Dados de Recursos
3. **Detecção de Problemas**: K8sGPT → Analisadores → Identificação de Problemas
4. **Explicação IA**: K8sGPT → API Ollama → Modelo LLM → Explicação
5. **Exibição de Resultados**: K8sGPT → Console → Usuário

## Arquitetura de Rede

### Modo Rede Host (Padrão)

```yaml
network_mode: host
```

**Vantagens**:
- Acesso direto ao host (ex: `localhost:11434`)
- Mais simples de configurar
- Melhor performance

**Desvantagens**:
- Compartilha stack de rede do host
- Não funciona em todos os ambientes

### Modo Rede Bridge (Alternativa)

```yaml
network_mode: bridge
ports:
  - "11434:11434"
```

**Vantagens**:
- Isolamento de rede melhor
- Funciona em qualquer ambiente
- Mapeamento explícito de portas

## Gerenciamento de Volumes

### Volumes Persistentes

1. **mekhanikube-ollama-data**: Armazena modelos LLM
2. **mekhanikube-k8sgpt-config**: Armazena configuração K8sGPT
3. **~/.kube/config**: Kubeconfig montado como somente leitura

## Considerações de Segurança

### Acesso ao Kubeconfig
- Montado como **somente leitura**
- Nunca modificado
- Isolado dentro do contêiner

### Segurança de Rede
- Recomendado: usar rede host para simplicidade
- Alternativa: rede bridge com portas explícitas
- Nunca expor publicamente

### Privacidade de Dados
- Todos os dados permanecem locais
- Nenhuma chamada de API externa (exceto downloads de modelos)
- Sem telemetria

## Escalabilidade e Performance

### Requisitos de Recursos

**Mínimo**:
- 2 núcleos de CPU
- 4GB RAM
- 10GB disco

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
