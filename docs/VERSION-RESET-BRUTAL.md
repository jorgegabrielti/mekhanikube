# Reset Brutal de Versionamento - v2.0.5 → v0.9.0

**Data:** 20 de Novembro de 2025  
**Decisão:** Reset completo do versionamento do projeto  
**Autor:** Jorge Gabriel

---

## 🎯 Resumo Executivo

Este documento registra a decisão de **resetar brutalmente** o versionamento do Nautikube de **v2.0.5** para **v0.9.0-beta**, reconhecendo que o projeto nunca teve uma versão estável v1.0.0 e que os números de versão foram inflacionados prematuramente.

## 🤔 Contexto da Decisão

### Situação Anterior
- **Versão atual:** v2.0.5
- **Problema identificado:** Pulamos direto para v2.0.0 sem nunca ter lançado uma v1.0.0 estável
- **Realidade:** O projeto está funcional mas ainda em desenvolvimento ativo, sem a maturidade que v2.x sugere
- **Impacto:** Números de versão não refletem o estado real do projeto

### Opções Consideradas

1. **Reset para v1.0.0** - Começar com versão "estável" imediatamente
   - ❌ Ainda seria desonesto, pois não atingimos maturidade de v1.0

2. **Continuar para v3.0.0** - Prosseguir com a numeração atual
   - ❌ Perpetua o problema, torna ainda mais difícil corrigir depois

3. **Reset Brutal para v0.9.0** - Recomeçar com honestidade ✅
   - ✅ Reconhece o trabalho já feito (90% do caminho)
   - ✅ É honesto sobre o estado atual (beta)
   - ✅ Segue convenções da comunidade open source
   - ✅ Permite crescimento estruturado até v1.0.0

## ✅ Decisão Final

**Escolhemos a opção #3: Reset Brutal para v0.9.0-beta**

### Por que v0.9.0 especificamente?

1. **Sinaliza Progresso:** O "9" indica que estamos a 90% do caminho para v1.0.0
2. **Respeita o Trabalho:** Não voltamos para v0.1.0, reconhecemos o que já foi construído
3. **Convenção da Comunidade:** v0.9.x é usado tradicionalmente como "quase pronto"
4. **Permite Refinamentos:** v0.9.x → v0.10.0 (RC) → v1.0.0 (estável)

## 📊 Histórico de Versões (Antes do Reset)

### Versões que Existiram
- **v2.0.0** (Outubro 2025) - Primeira versão com Docker-First
- **v2.0.1** (Outubro 2025) - Melhorias na interface
- **v2.0.2** (Outubro 2025) - Correções de bugs
- **v2.0.3** (Novembro 2025) - Conexão agnóstica com clusters
- **v2.0.4** (Novembro 2025) - Otimizações de timeout
- **v2.0.5** (Novembro 2025) - Ajustes finais antes do reset

### Funcionalidades Implementadas (Mantidas em v0.9.0)
- ✅ Análise completa de recursos Kubernetes
- ✅ Integração com Ollama para explicações IA
- ✅ Detecção agnóstica de 7 tipos de cluster (Kind, Minikube, Docker Desktop, k3d, EKS, AKS, GKE)
- ✅ Estratégia de fallback multi-nível (4 níveis)
- ✅ Arquitetura Docker-First funcional
- ✅ Filtros por namespace e tipo de recurso
- ✅ Modo detalhado com --explain
- ✅ Documentação técnica completa

## 🛣️ Roadmap para v1.0.0

### v0.9.x (Novembro - Dezembro 2025)
- Refinamentos e ajustes
- Correções de bugs descobertos em uso real
- Melhorias de performance
- Documentação adicional

### v0.10.0 (Dezembro 2025)
- **Release Candidate (RC)**
- Feature freeze - sem novas funcionalidades
- Testes intensivos
- Validação com usuários beta

### v1.0.0 (Janeiro 2026)
- **Primeira Versão Estável - CLI-First**
- Arquitetura CLI-First (sem Docker obrigatório)
- Suporte multi-provider IA (Ollama, OpenAI, Anthropic, Gemini)
- Sistema de configuração config.yaml
- Documentação profissional completa
- Garantia de backward compatibility a partir deste ponto

## 🎓 Lições Aprendidas

### O que Aprendemos
1. **Honestidade > Números Bonitos:** É melhor ter v0.9.0 honesto que v2.0.5 inflacionado
2. **SemVer é Sério:** Semantic Versioning não é apenas números, é um contrato com usuários
3. **v1.0.0 é um Compromisso:** Significa "estável, testado, pronto para produção"
4. **Correção Requer Coragem:** Resetar é difícil, mas é a coisa certa a fazer

### Por que Isso Importa
- **Confiança:** Usuários precisam confiar que os números de versão significam algo
- **Expectativas:** v2.x sugere maturidade que ainda não atingimos
- **Comunidade:** Open source depende de transparência e honestidade
- **Longo Prazo:** Melhor corrigir agora que ter que fazer em v10.0.0

## 🔄 Processo de Reset

### Arquivos Modificados
1. `VERSION` - 2.0.5 → 0.9.0
2. `cmd/nautikube/main.go` - Version constant → "0.9.0-beta"
3. `CHANGELOG.md` - Adicionada seção de reset explicando a mudança
4. `README.md` - Banner de beta warning adicionado
5. `docs/VERSION-RESET-BRUTAL.md` - Este documento criado

### Git Workflow
```bash
git add -A
git commit -m "feat: brutal version reset v2.0.5 → v0.9.0-beta

BREAKING CHANGE: Version numbering has been reset to correctly reflect
project maturity. This is an honest reset - we never had v1.0.0 stable.

- Previous versions (v2.0.0-v2.0.5) are preserved in git history
- All functionality remains the same
- v0.9.0 signals we're 90% to stable v1.0.0
- Roadmap: v0.9.x → v0.10.0 (RC) → v1.0.0 (Jan 2026)

See docs/VERSION-RESET-BRUTAL.md for full rationale."

git tag -a v0.9.0 -m "Version 0.9.0-beta - Honest reset, functional beta"
git push origin develop
git push origin v0.9.0
```

## 📢 Comunicação

### Mensagem aos Usuários
> "Estamos fazendo um reset honesto do versionamento. O Nautikube v2.0.5 se torna v0.9.0-beta, refletindo corretamente que estamos em beta funcional, não em produção estável. Todo o código funciona perfeitamente, apenas os números mudaram para serem honestos. v1.0.0 chegará em Janeiro/2026 com CLI-First."

### Benefícios para a Comunidade
- ✅ Transparência total sobre estado do projeto
- ✅ Expectativas alinhadas com realidade
- ✅ Permite crescimento estruturado e sustentável
- ✅ Demonstra maturidade ao admitir e corrigir erro

## 🎯 Compromisso

A partir de v0.9.0, nos comprometemos a:

1. **Seguir SemVer rigorosamente:** Sem atalhos, sem pulos
2. **v1.0.0 será real:** Só lançaremos quando estivermos prontos de verdade
3. **Transparência sempre:** Comunicar claramente o estado do projeto
4. **Aprender com o erro:** Usar isso como exemplo de como fazer certo

---

## 🔗 Referências

- **Semantic Versioning 2.0.0:** https://semver.org/
- **Git Tagging:** https://git-scm.com/book/en/v2/Git-Basics-Tagging
- **Changelog Format:** https://keepachangelog.com/
- **Beta/RC Conventions:** https://en.wikipedia.org/wiki/Software_release_life_cycle

---

**Conclusão:** Este reset não é um fracasso, é uma demonstração de maturidade e honestidade. Estamos construindo algo sólido, e isso começa com ter coragem de fazer o que é certo, mesmo quando é difícil.

_"A honestidade é a melhor política, especialmente em versionamento de software."_ 🚀
