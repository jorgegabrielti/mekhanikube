# Política de Segurança

## Versões Suportadas

As seguintes versões do Mekhanikube estão atualmente recebendo atualizações de segurança:

| Versão  | Suportada          | Observações |
| ------- | ------------------ | ----------- |
| 2.0.x   | :white_check_mark: | Engine próprio Go (recomendado) |
| 1.0.x   | :white_check_mark: | K8sGPT legacy (manutenção) |
| < 1.0   | :x:                | Não suportado |

## Considerações de Segurança

### Implantação Apenas Local

O Mekhanikube foi projetado para ser executado **localmente** na sua infraestrutura. Não deve ser exposto à internet pública.

**Recursos de Segurança Principais**:
- ✅ Nenhuma chamada de API externa (exceto downloads de modelos do Ollama)
- ✅ Todos os dados permanecem locais
- ✅ Sem telemetria ou rastreamento
- ✅ Acesso somente leitura ao cluster Kubernetes
- ✅ Kubeconfig montado como somente leitura

### O que o Mekhanikube NÃO Faz

- ❌ Não modifica seu cluster Kubernetes
- ❌ Não envia dados para serviços externos
- ❌ Não armazena credenciais sensíveis externamente
- ❌ Não expõe APIs publicamente

## Reportar uma Vulnerabilidade

Se você descobrir uma vulnerabilidade de segurança no Mekhanikube, por favor reporte-a de forma responsável:

### Como Reportar

**Email**: [jorgegabrielti@gmail.com](mailto:jorgegabrielti@gmail.com)

**Assunto**: `[SEGURANÇA] Breve descrição do problema`

**Por favor, inclua**:
1. Descrição da vulnerabilidade
2. Passos para reproduzir
3. Impacto potencial
4. Correção sugerida (se disponível)
5. Suas informações de contato (opcional, para acompanhamento)

### O que Esperar

- **Resposta Inicial**: Dentro de 48 horas
- **Atualização de Status**: Dentro de 7 dias
- **Prazo para Correção**: Depende da severidade (veja abaixo)
- **Divulgação Pública**: Após a correção ser lançada ou 90 dias (o que vier primeiro)

### Níveis de Severidade

#### Crítico (Correção em 7 dias)
- Execução remota de código
- Escalação de privilégios
- Exposição de kubeconfig ou credenciais
- Modificação de cluster sem autorização

#### Alto (Correção em 14 dias)
- Divulgação de informações (dados do cluster)
- Negação de serviço afetando o cluster
- Bypass de controles de acesso

#### Médio (Correção em 30 dias)
- Negação de serviço (apenas local)
- Vazamento de informações (não sensíveis)
- Problemas de configuração

#### Baixo (Correção quando possível)
- Problemas menores com impacto limitado
- Erros de documentação
- Violações de boas práticas

## Melhores Práticas de Segurança

### Para Usuários

1. **Isolamento de Rede**
   ```yaml
   # Mantenha contêineres na rede do host (padrão)
   # Ou use rede Docker privada
   network_mode: host
   ```

2. **Proteção do Kubeconfig**
   ```yaml
   # Sempre monte como somente leitura
   volumes:
     - ~/.kube/config:/root/.kube/config:ro
   ```

3. **Atualizações Regulares**
   ```bash
   # Mantenha o Mekhanikube atualizado
   git pull origin main
   make build
   make restart
   ```

4. **Limitar Acesso ao Cluster**
   ```bash
   # Use service account com permissões somente leitura
   # Crie kubeconfig dedicado para o Mekhanikube
   ```

5. **Monitorar Logs**
   ```bash
   # Verifique os logs regularmente para anomalias
   make logs
   ```

### Para Desenvolvedores

1. **Gerenciamento de Dependências**
   - Mantenha as imagens base atualizadas
   - Faça varredura de vulnerabilidades regularmente
   - Fixe versões de dependências

2. **Revisão de Código**
   - Todos os PRs requerem revisão
   - Mudanças sensíveis à segurança precisam de escrutínio extra
   - Execute verificações de segurança no CI/CD

3. **Gerenciamento de Segredos**
   - Nunca faça commit de segredos no git
   - Use arquivos `.env` (já no `.gitignore`)
   - Rotacione credenciais regularmente

4. **Segurança de Contêineres**
   - Execute contêineres como não-root quando possível
   - Minimize o tamanho da imagem
   - Use imagens base oficiais
   - Habilite varredura de segurança

## Testes de Segurança

### Verificações Automáticas de Segurança

O Mekhanikube inclui:

- **Varredura de vulnerabilidades Trivy** (no CI/CD)
- **ShellCheck** para lint de scripts
- **Validação do Docker Compose**

Execute localmente:
```bash
# Lint de configuração
make lint

# Executar testes
make test

# Verificar saúde
make health
```

### Revisão Manual de Segurança

Antes de cada lançamento:
- [ ] Revisar todas as dependências
- [ ] Verificar vulnerabilidades conhecidas
- [ ] Testar com permissões mínimas
- [ ] Verificar que não há exposição de dados sensíveis
- [ ] Confirmar acesso somente leitura ao cluster

## Limitações Conhecidas

### 1. Exposição do Kubeconfig no Contêiner

**Problema**: Kubeconfig é montado no sistema de arquivos do contêiner.

**Mitigação**:
- Montado como somente leitura
- Sistema de arquivos do contêiner é efêmero
- Não exposto externamente

**Recomendação**: Use kubeconfig dedicado com permissões mínimas.

### 2. Acesso ao Socket do Docker (Não Necessário)

**Status**: O Mekhanikube NÃO requer acesso ao socket do Docker.

**Se você ver solicitações para `/var/run/docker.sock`**: Isso não é necessário e não deve ser concedido.

### 3. Modo de Rede Host

**Trade-off**: O modo de rede host simplifica a conectividade, mas compartilha a pilha de rede do host.

**Alternativa**: Use modo bridge com mapeamentos explícitos de porta:
```yaml
network_mode: bridge
ports:
  - "11434:11434"
```

## Política de Divulgação de Segurança

### Divulgação Pública

Vulnerabilidades de segurança serão divulgadas:

1. **Após uma correção ser lançada**
2. **Após 90 dias** (se nenhuma correção estiver disponível)
3. **Com crédito ao relator** (se desejado)

### Mural da Fama

Reconhecemos pesquisadores de segurança que divulgam vulnerabilidades de forma responsável:

- *Seja o primeiro!*

## Conformidade

### Privacidade de Dados

O Mekhanikube foi projetado para privacidade:
- Sem coleta de dados
- Sem comunicações externas
- Sem telemetria
- Compatível com GDPR (nenhum dado pessoal processado)

### Trilha de Auditoria

Para fins de conformidade:
```bash
# Todas as ações são registradas
docker logs mekhanikube-k8sgpt

# Logs de auditoria da API Kubernetes (no seu cluster)
kubectl logs -n kube-system kube-apiserver-*
```

## Recursos de Segurança

- [Melhores Práticas de Segurança do Docker](https://docs.docker.com/engine/security/)
- [Segurança do Kubernetes](https://kubernetes.io/docs/concepts/security/)
- [Segurança do K8sGPT](https://docs.k8sgpt.ai/)
- [Segurança do Ollama](https://github.com/ollama/ollama/blob/main/docs/security.md)

## Contato

Para questões de segurança:
- **Email**: [jorgegabrielti@gmail.com](mailto:jorgegabrielti@gmail.com)
- **GitHub Issues**: Para problemas não sensíveis
- **GitHub Security Advisory**: Para divulgação responsável

---

**Última Atualização**: 2025-11-09
