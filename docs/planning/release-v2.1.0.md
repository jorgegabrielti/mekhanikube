# Plano de Release: v2.1.0 - Simulação de Cloud

**Data Alvo:** 25-11-2025
**Status:** RASCUNHO

## 🎯 Objetivos
Habilitar a simulação local de clusters Kubernetes de Provedores de Cloud (especificamente AWS EKS) para validar as funcionalidades de "Conexão Agnóstica" sem incorrer em custos de nuvem. Isso permite testes de ponta a ponta das cadeias de autenticação (AWS CLI, etc.) dentro do container NautiKube.

## ✨ Novas Funcionalidades

### Funcionalidade 1: Integração com LocalStack EKS
- **Descrição:** Adicionar uma configuração de ambiente de desenvolvimento que inicia o LocalStack com EKS habilitado.
- **História de Usuário:** Como desenvolvedor, quero rodar `make dev-eks` para iniciar um cluster compatível com EKS localmente para que eu possa testar a lógica de autenticação AWS do NautiKube.
- **Implementação Técnica:**
    - Adicionar serviço `localstack` ao `docker-compose.dev.yml` (ou similar).
    - Criar scripts para:
        1. Inicializar o LocalStack EKS.
        2. Gerar um kubeconfig que usa `aws` CLI para autenticação.
        3. Montar este kubeconfig no NautiKube.
- **Critérios de Aceite:**
    - NautiKube inicia.
    - Detecta o cluster como "AWS EKS" (baseado na URL/Auth).
    - Conecta com sucesso e lista os nós usando o método de autenticação AWS CLI dentro do container.

## 🔧 Melhorias
- **[Docs]**: Adicionar guia sobre "Testando Provedores de Cloud Localmente".

## 🏗 Mudanças Técnicas
- **[Dependência]**: Adicionar `localstack` às dependências de desenvolvimento do projeto (docker-compose).
- **[Script]**: Novo `scripts/setup-local-eks.sh`.

## 🧪 Plano de Verificação
- [ ] Rodar `scripts/setup-local-eks.sh`.
- [ ] Verificar se `kubectl` no host consegue conectar.
- [ ] Iniciar NautiKube.
- [ ] Checar logs para confirmar detecção "Tipo: AWS EKS".
- [ ] Verificar se `nautikube analyze` funciona contra o cluster localstack.

## 📝 Documentação
- [ ] CONTRIBUTING.md (Atualizar com novo fluxo de dev)
- [ ] docs/CLOUD-SIMULATION.md (Novo)
