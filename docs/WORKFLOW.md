# 🌊 Workflow de Desenvolvimento - NautiKube

O NautiKube usa **GitHub Flow** - uma estratégia simples, eficiente e baseada em Pull Requests.

## 📋 Índice

- [Visão Geral](#visão-geral)
- [Estrutura de Branches](#estrutura-de-branches)
- [Workflow Passo a Passo](#workflow-passo-a-passo)
- [Convenções](#convenções)
- [Exemplos Práticos](#exemplos-práticos)

---

## 🎯 Visão Geral

### Princípios do GitHub Flow

1. **Branch `main` sempre estável** - Pronta para produção a qualquer momento
2. **Features em branches temporárias** - Isolamento e experimentação
3. **Pull Requests para tudo** - Code review e discussão
4. **Deploy após merge** - CI/CD automático via tags

### Fluxo Visual

```
main ─────●─────●─────●─────●─────●─────
           │     │     │     │     │
           │     │     │     │     └─ feature/dashboard
           │     │     │     └─────── hotfix/critical-bug
           │     │     └───────────── feature/gpu-support
           │     └─────────────────── feature/new-scanner
           └───────────────────────── v2.0.2 (tag)

         PR→merge    PR→merge    PR→merge
```

---

## 🌳 Estrutura de Branches

### Branch Principal

**`main`** - Branch de Produção
- ✅ Sempre estável e funcional
- ✅ Protegida: requer PR aprovado
- ✅ Cada release recebe uma tag (v2.0.2, v2.1.0)
- ✅ CI/CD automático quando recebe tag
- 🚫 **NUNCA commitar diretamente**

### Branches Temporárias

**`feature/*`** - Novas funcionalidades
- Exemplos: `feature/gpu-support`, `feature/dashboard`
- Criada de: `main`
- Merge para: `main` (via PR)
- Vida útil: dias/semanas
- Deletada após merge

**`fix/*`** - Correções não-urgentes
- Exemplos: `fix/scanner-timeout`, `fix/typo-docs`
- Criada de: `main`
- Merge para: `main` (via PR)
- Vida útil: horas/dias

**`hotfix/*`** - Correções urgentes em produção
- Exemplos: `hotfix/critical-security`, `hotfix/data-loss`
- Criada de: `main`
- Merge para: `main` (PR rápido)
- Deploy imediato após merge

**`docs/*`** - Apenas documentação
- Exemplos: `docs/update-readme`, `docs/gpu-guide`
- Criada de: `main`
- Merge para: `main` (via PR)

---

## 🚀 Workflow Passo a Passo

### 1️⃣ Iniciar Nova Feature

```powershell
# Atualizar main
git checkout main
git pull origin main

# Criar feature branch
git checkout -b feature/gpu-support

# Fazer push inicial (para backup e colaboração)
git push -u origin feature/gpu-support
```

### 2️⃣ Desenvolver

```powershell
# Trabalhar normalmente
# Fazer commits frequentes e atômicos

git add docker-compose.gpu.yml
git commit -m "feat(gpu): adicionar docker-compose override para GPU"

git add docs/GPU-SETUP.md
git commit -m "docs(gpu): guia completo de setup"

git add scripts/nautikube-helpers.ps1
git commit -m "feat(scripts): helpers PowerShell para GPU"

# Push frequente (backup + visibilidade)
git push origin feature/gpu-support
```

### 3️⃣ Manter Atualizado com Main (Opcional)

Se `main` recebeu atualizações durante seu trabalho:

```powershell
# Atualizar main local
git checkout main
git pull origin main

# Voltar para feature e trazer mudanças
git checkout feature/gpu-support
git rebase main
# ou
git merge main

# Resolver conflitos se houver
git push origin feature/gpu-support --force-with-lease
```

### 4️⃣ Criar Pull Request

Quando a feature estiver pronta:

**No GitHub:**
1. Vá para: https://github.com/jorgegabrielti/nautikube
2. Clique em "Pull Requests" → "New Pull Request"
3. Base: `main` ← Compare: `feature/gpu-support`
4. Preencha o template:

```markdown
## 📝 Descrição

Adiciona suporte opcional para aceleração GPU NVIDIA, mantendo CPU como padrão.

## 🎯 Tipo de Mudança

- [x] Nova feature
- [ ] Bug fix
- [ ] Breaking change
- [ ] Documentação

## ✅ Checklist

- [x] Código compila sem erros
- [x] Testes locais passaram
- [x] Documentação atualizada
- [x] CHANGELOG.md atualizado
- [x] Testado manualmente

## 🧪 Como Testar

```powershell
git checkout feature/gpu-support
docker-compose -f docker-compose.yml -f docker-compose.gpu.yml up -d
docker exec nautikube-ollama nvidia-smi
```

## 📸 Screenshots (se aplicável)

[Adicionar prints se for UI]
```

### 5️⃣ Code Review

- **Auto-review** (você é owner): Revisar próprio código
- **Aguardar CI**: GitHub Actions deve passar
- **Fazer ajustes** se necessário (novos commits no mesmo branch)

### 6️⃣ Merge

Quando aprovado:

**Opções de merge:**
- **Squash and merge** ✅ (recomendado): Agrupa commits em um
- **Merge commit**: Mantém histórico completo
- **Rebase and merge**: Histórico linear

```powershell
# No GitHub: clicar em "Squash and merge"
# Editar mensagem do commit:
# "feat: adicionar suporte para GPU NVIDIA (#42)"

# Deletar branch automaticamente (marcar checkbox)
```

### 7️⃣ Limpar Localmente

```powershell
# Atualizar main
git checkout main
git pull origin main

# Deletar branch local
git branch -d feature/gpu-support

# Limpar branches remotas deletadas
git fetch --prune
```

### 8️⃣ Criar Release

Quando tiver features suficientes ou correções importantes:

```powershell
# Garantir que está na main atualizada
git checkout main
git pull origin main

# Usar script de release
.\scripts\release.ps1 -Version 2.1.0 -DryRun  # Testar
.\scripts\release.ps1 -Version 2.1.0 -Push    # Criar e publicar

# Ou manualmente:
.\scripts\release.ps1 -Version 2.1.0
git push origin main
git push origin v2.1.0  # 🚀 Dispara GitHub Actions!
```

---

## 📐 Convenções

### Nomes de Branches

| Tipo | Formato | Exemplo |
|------|---------|---------|
| Feature | `feature/<nome>` | `feature/gpu-support` |
| Bug Fix | `fix/<nome>` | `fix/scanner-timeout` |
| Hotfix | `hotfix/<nome>` | `hotfix/security-patch` |
| Docs | `docs/<nome>` | `docs/update-readme` |
| Chore | `chore/<nome>` | `chore/update-deps` |

**Regras:**
- ✅ Usar kebab-case (palavras-separadas-por-hífen)
- ✅ Ser descritivo mas conciso
- ✅ Evitar números ou IDs de issue

### Mensagens de Commit (Conventional Commits)

Formato:
```
<tipo>(<escopo>): <descrição curta>

<descrição longa opcional>

<footer opcional>
```

**Tipos:**
- `feat`: Nova funcionalidade
- `fix`: Correção de bug
- `docs`: Apenas documentação
- `style`: Formatação, ponto e vírgula, etc
- `refactor`: Refatoração sem mudar funcionalidade
- `perf`: Melhorias de performance
- `test`: Adicionar/modificar testes
- `chore`: Manutenção (deps, build, etc)
- `ci`: Mudanças em CI/CD

**Exemplos:**
```bash
feat(gpu): adicionar suporte para NVIDIA GPU
fix(scanner): corrigir timeout em clusters grandes
docs(readme): atualizar instruções de instalação
refactor(analyzer): simplificar lógica de detecção
chore(deps): atualizar golang para 1.21.5
ci(release): adicionar build para ARM64
```

### Pull Requests

**Título:** Seguir formato de commit
```
feat: adicionar suporte para GPU NVIDIA
fix: corrigir timeout no scanner de ConfigMaps
docs: atualizar guia de instalação
```

**Labels:** Usar labels do GitHub
- `feature` - Nova funcionalidade
- `bug` - Correção de bug
- `documentation` - Mudanças em docs
- `enhancement` - Melhoria de algo existente
- `breaking-change` - Quebra compatibilidade

---

## 💡 Exemplos Práticos

### Exemplo 1: Nova Feature (GPU Support)

```powershell
# 1. Criar branch
git checkout main
git pull
git checkout -b feature/gpu-support

# 2. Implementar
# ... editar arquivos ...

git add docker-compose.gpu.yml docs/GPU-SETUP.md
git commit -m "feat(gpu): adicionar suporte opcional para GPU NVIDIA"

git add .env.example
git commit -m "docs(env): adicionar variáveis de GPU"

git push origin feature/gpu-support

# 3. PR no GitHub
# Base: main ← feature/gpu-support
# Review → Merge

# 4. Limpar
git checkout main
git pull
git branch -d feature/gpu-support
```

### Exemplo 2: Bug Fix

```powershell
# 1. Criar branch
git checkout main
git pull
git checkout -b fix/scanner-timeout

# 2. Corrigir
# ... editar internal/scanner/scanner.go ...

git add internal/scanner/scanner.go
git commit -m "fix(scanner): aumentar timeout para 60s

Clusters grandes demoravam mais que 30s para responder.
Aumentado timeout de 30s para 60s."

git push origin fix/scanner-timeout

# 3. PR → Merge → Limpar
```

### Exemplo 3: Hotfix Urgente

```powershell
# 1. Criar branch direto da main
git checkout main
git pull
git checkout -b hotfix/security-vuln

# 2. Corrigir rapidamente
# ... aplicar patch ...

git add .
git commit -m "fix(security): aplicar patch CVE-2024-XXXX"

git push origin hotfix/security-vuln

# 3. PR urgente (fast-track)
# 4. Merge imediato
# 5. Release patch
.\scripts\release.ps1 -Version 2.0.3 -Push
```

### Exemplo 4: Apenas Documentação

```powershell
# 1. Criar branch
git checkout -b docs/gpu-screenshots

# 2. Adicionar imagens/docs
git add docs/GPU-SETUP.md assets/gpu-*.png
git commit -m "docs(gpu): adicionar screenshots do setup"

git push origin docs/gpu-screenshots

# 3. PR → Merge
```

---

## 🛡️ Proteção de Branches

Configure no GitHub:

**Settings → Branches → Branch protection rules → `main`**

Configurações recomendadas:
- ✅ **Require pull request before merging**
- ✅ **Require approvals**: 1 (você mesmo, se solo)
- ✅ **Dismiss stale reviews**: Sim
- ✅ **Require status checks**: GitHub Actions
- ✅ **Require conversation resolution**: Sim
- ❌ **Allow force pushes**: NUNCA
- ❌ **Allow deletions**: NUNCA

---

## 🎯 Vantagens do GitHub Flow

✅ **Simples**: Fácil de entender e seguir  
✅ **Rápido**: Features vão para produção rapidamente  
✅ **Seguro**: PRs garantem qualidade  
✅ **Flexível**: Funciona para projetos de qualquer tamanho  
✅ **GitHub-native**: Integra perfeitamente com Actions  

---

## 📚 Referências

- [GitHub Flow Guide](https://guides.github.com/introduction/flow/)
- [Conventional Commits](https://www.conventionalcommits.org/)
- [Semantic Versioning](https://semver.org/)

---

**Última atualização:** 2025-11-11  
**Versão do documento:** 1.0
