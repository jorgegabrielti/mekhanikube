# 🔄 Nautikube v0.9.0-beta - Reset Honesto

## ⚠️ BREAKING CHANGE: Reset de Versionamento

**Esta é uma mudança importante:** O Nautikube está fazendo um **reset brutal de versionamento** de **v2.0.5 → v0.9.0-beta** para refletir corretamente a maturidade do projeto.

### 🤔 Por que este reset?

- **Honestidade primeiro:** Nunca tivemos uma v1.0.0 estável - pulamos direto para v2.0.0
- **Números inflacionados:** v2.x sugeria maturidade que ainda não atingimos
- **Recomeço correto:** v0.9.0 sinaliza que estamos a **90% de uma v1.0.0 real**
- **Respeito ao trabalho:** Reconhecemos o progresso significativo já feito

### ✅ O que muda?

**Resposta curta:** Apenas os números de versão. Todo o código funciona exatamente igual!

- ✅ **Todas as funcionalidades de v2.0.5 estão presentes**
- ✅ **Código 100% funcional**
- ✅ **Mesma performance e estabilidade**
- ✅ **Apenas refletindo status honesto: beta funcional**

---

## 🚀 Funcionalidades (Mantidas de v2.0.5)

### 🔍 Análise Completa de Kubernetes
- Escaneamento completo de todos os recursos do cluster
- Identificação automática de problemas e configurações incorretas
- Relatórios detalhados por namespace e tipo de recurso

### 🤖 Integração com IA Local (Ollama)
- Explicações em linguagem simples usando LLM local
- 100% privado - nenhum dado enviado para nuvem
- Timeout otimizado de 300s para primeira requisição

### 🌍 Conexão Agnóstica - Qualquer Cluster
Detecta e conecta automaticamente com:
- **Local:** Kind, Minikube, Docker Desktop, k3d
- **Cloud:** AWS EKS, Azure AKS, Google GKE
- **Custom:** Qualquer cluster Kubernetes padrão

### 🔄 Estratégia Multi-Nível de Fallback (4 níveis)
1. In-cluster config (quando rodando dentro do cluster)
2. `/root/.kube/config_mod` (modificado pelo entrypoint)
3. `~/.kube/config` (configuração padrão do usuário)
4. `KUBECONFIG` env var (variável de ambiente)

### 🎯 Filtros e Modos
- Filtro por **namespace** (`-n`)
- Filtro por **tipo de recurso** (`--filter`)
- Modo **explicado** com IA (`--explain`)
- Modo **verbose** para debugging

---

## 📦 Instalação e Uso

### Pré-requisitos
- Docker e Docker Compose instalados
- Cluster Kubernetes em execução
- Ollama rodando localmente (porta 11434)

### Quick Start

```bash
# Clone o repositório
git clone https://github.com/jorgegabrielti/nautikube.git
cd nautikube

# Inicie o container
docker-compose up -d

# Execute análise básica
docker exec nautikube nautikube analyze

# Análise com explicações IA
docker exec nautikube nautikube analyze --explain

# Filtrar por namespace
docker exec nautikube nautikube analyze -n kube-system

# Filtrar por tipo de recurso
docker exec nautikube nautikube analyze --filter Pod --filter Service
```

---

## 🛣️ Roadmap para v1.0.0

Estamos comprometidos com transparência total sobre nosso roadmap:

### v0.9.x (Novembro - Dezembro 2025)
- Refinamentos baseados em feedback
- Correções de bugs descobertos em uso real
- Melhorias de performance
- Documentação adicional

### v0.10.0 (Dezembro 2025)
- **Release Candidate (RC)**
- Feature freeze - sem novas funcionalidades
- Testes intensivos de integração
- Validação com usuários beta

### v1.0.0 (Janeiro 2026) - Primeira Versão Estável
- **Arquitetura CLI-First** (sem Docker obrigatório)
- **Suporte multi-provider IA:** Ollama, OpenAI, Anthropic, Gemini
- **Sistema de configuração:** `config.yaml` completo
- **Documentação profissional:** Guias completos e tutoriais
- **Garantia de backward compatibility** a partir deste ponto

---

## 📚 Documentação Completa

### Novos Documentos (v0.9.0)
- **[VERSION-RESET-BRUTAL.md](docs/VERSION-RESET-BRUTAL.md)** - Explicação completa da decisão de reset
- **[AGNOSTIC-CONNECTION.md](docs/AGNOSTIC-CONNECTION.md)** - Como funciona a conexão universal

### Documentação Geral
- **[README.md](README.md)** - Visão geral e início rápido
- **[ARCHITECTURE.md](docs/ARCHITECTURE.md)** - Arquitetura técnica
- **[DEVELOPMENT.md](docs/DEVELOPMENT.md)** - Guia de desenvolvimento
- **[TROUBLESHOOTING.md](docs/TROUBLESHOOTING.md)** - Solução de problemas

---

## 🤝 Contribuindo

Contribuições são bem-vindas! Este é um projeto open source e estamos construindo algo sólido.

1. Fork o projeto
2. Crie uma branch para sua feature (`git checkout -b feature/MinhaFeature`)
3. Commit suas mudanças (`git commit -m 'Add: Minha nova feature'`)
4. Push para a branch (`git push origin feature/MinhaFeature`)
5. Abra um Pull Request

Veja [CONTRIBUTING.md](CONTRIBUTING.md) para detalhes.

---

## 📝 Versões Anteriores

As versões v2.0.0 a v2.0.5 permanecem no histórico Git para referência. Todas as funcionalidades implementadas nessas versões estão presentes em v0.9.0.

### Histórico Preservado
- **v2.0.0** - Primeira versão Docker-First
- **v2.0.1** - Melhorias de interface
- **v2.0.2** - Correções de bugs
- **v2.0.3** - Conexão agnóstica implementada
- **v2.0.4** - Otimizações de timeout
- **v2.0.5** - Ajustes finais antes do reset

---

## 🎯 Compromisso com a Comunidade

A partir de v0.9.0, nos comprometemos a:

1. **Seguir SemVer rigorosamente** - Sem atalhos, sem pulos
2. **v1.0.0 será real** - Só lançaremos quando estivermos prontos de verdade
3. **Transparência sempre** - Comunicar claramente o estado do projeto
4. **Aprender com erros** - Usar isso como exemplo de como fazer certo

---

## 📄 Licença

MIT License - veja [LICENSE](LICENSE) para detalhes.

---

## 🙏 Agradecimentos

Este reset não é um fracasso, é uma **demonstração de maturidade e honestidade**. Estamos construindo algo sólido, e isso começa com ter coragem de fazer o que é certo, mesmo quando é difícil.

Obrigado pela compreensão e apoio! 🚀

---

**"A honestidade é a melhor política, especialmente em versionamento de software."**

[⬇️ Download v0.9.0](https://github.com/jorgegabrielti/nautikube/archive/refs/tags/v0.9.0.zip) | [📖 Documentação](https://github.com/jorgegabrielti/nautikube/tree/develop/docs) | [🐛 Reportar Bug](https://github.com/jorgegabrielti/nautikube/issues)
