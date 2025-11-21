# 🔄 Development Workflow - Nautikube

## Processo Definido para Cada Nova Feature

Este documento estabelece o workflow padrão para desenvolvimento de features no Nautikube, garantindo qualidade e rastreabilidade.

---

## 📋 Workflow Completo (8 Etapas)

### 1. 📝 Planejamento (GitHub Issues)
**Objetivo:** Definir escopo antes de codificar

**Checklist:**
- [ ] Issue criada no GitHub com número (#N)
- [ ] Descrição clara do objetivo
- [ ] Critérios de aceitação definidos
- [ ] Story Points estimados (1 SP = 1 hora)
- [ ] Sprint atribuído
- [ ] Labels aplicadas (enhancement, v0.9.x, etc)
- [ ] Issue adicionada ao GitHub Project

**Exemplo:**
```
Issue #9: Sistema de Severidade e Score
SP: 3 (3 horas)
Sprint: 1
Labels: enhancement, v0.9.x, v1.0.0
```

---

### 2. 🌿 Criar Branch Feature
**Objetivo:** Isolar desenvolvimento da branch principal

**Comando:**
```powershell
git checkout develop
git pull origin develop
git checkout -b feature/issue-N-nome-descritivo
```

**Padrão de nomenclatura:**
- `feature/issue-9-severity-system`
- `feature/issue-10-json-export`
- `bugfix/issue-15-crash-on-startup`

---

### 3. 💻 Implementação
**Objetivo:** Desenvolver funcionalidade conforme critérios

**Checklist:**
- [ ] Código implementado seguindo critérios de aceitação
- [ ] Comentários e godoc adicionados
- [ ] Convenções do Go seguidas (gofmt, golint)
- [ ] Tratamento de erros adequado
- [ ] Logging apropriado

**Boas práticas:**
- Commits incrementais durante desenvolvimento
- Mensagens de commit descritivas
- Código limpo e legível

---

### 4. 🧪 Testes Unitários
**Objetivo:** Garantir cobertura de testes

**Checklist:**
- [ ] Testes unitários escritos para nova funcionalidade
- [ ] Casos de borda cobertos
- [ ] Testes de integração (se aplicável)
- [ ] **Todos os testes passando: `go test ./... -v`**
- [ ] Cobertura mínima: 70% (meta: >80%)

**Comando:**
```powershell
# Rodar todos os testes
go test ./... -v

# Verificar cobertura
go test ./... -cover

# Cobertura detalhada
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

**Critério obrigatório:** ❌ Se testes falharem, NÃO avançar!

---

### 5. 🏗️ Compilação
**Objetivo:** Validar que código compila sem erros

**Checklist:**
- [ ] **Build bem-sucedido: `go build ./cmd/nautikube`**
- [ ] Nenhum erro de compilação
- [ ] Nenhum warning crítico
- [ ] Binário gerado com sucesso

**Comando:**
```powershell
go build -o nautikube.exe ./cmd/nautikube
```

**Critério obrigatório:** ❌ Se build falhar, NÃO avançar!

---

### 6. 🔬 Teste Manual (Cluster Real)
**Objetivo:** Validar funcionalidade em ambiente real

**Checklist:**
- [ ] Cluster Kind/Minikube iniciado
- [ ] Recursos de teste criados (pods, services, etc)
- [ ] Funcionalidade testada end-to-end
- [ ] Output validado conforme esperado
- [ ] Edge cases testados manualmente
- [ ] Screenshots/logs salvos (opcional)

**Exemplo - Issue #9:**
```powershell
# 1. Criar cluster Kind
kind create cluster --name nautikube-test

# 2. Criar pods com problemas
kubectl apply -f test-pods.yaml

# 3. Executar nautikube
.\nautikube.exe analyze

# 4. Validar output:
# ✅ Severity atribuída corretamente
# ✅ Score calculado com contexto
# ✅ Ícones exibidos
# ✅ Problemas detectados

# 5. Limpar
kubectl delete -f test-pods.yaml
```

**Critério obrigatório:** ❌ Se teste manual falhar, corrigir e repetir etapas 3-6!

---

### 7. 📦 Commit e Push
**Objetivo:** Versionar código validado

**Checklist:**
- [ ] Arquivos temporários removidos (*.exe, test-*.yaml, etc)
- [ ] CHANGELOG.md atualizado
- [ ] README.md atualizado (se necessário)
- [ ] VERSION atualizado
- [ ] Commit com mensagem convencional
- [ ] Push da branch feature

**Formato de mensagem de commit:**
```
<tipo>: <descrição curta>

<descrição detalhada>
- Bullet points de mudanças
- Detalhes técnicos
- Testes realizados

Tested:
- X unit tests passing
- Manual test on Kind cluster
- Y scenarios validated

Closes #N
Sprint X - Issue #N (Y SP)
Release: vX.Y.Z
```

**Tipos de commit:**
- `feat:` - Nova funcionalidade
- `fix:` - Correção de bug
- `docs:` - Documentação
- `test:` - Testes
- `refactor:` - Refatoração
- `chore:` - Tarefas de manutenção

**Comando:**
```powershell
# Stage arquivos relevantes
git add <arquivos>

# Commit
git commit -m "feat: descrição

Detalhes...

Closes #N"

# Push
git push -u origin feature/issue-N-nome
```

---

### 8. 🔀 Pull Request e Merge
**Objetivo:** Integrar código na branch develop

**Checklist:**
- [ ] Pull Request criado no GitHub
- [ ] Título descritivo
- [ ] Descrição completa (o que, por que, como)
- [ ] Issue vinculada (#N)
- [ ] Screenshots/GIFs (se UI mudou)
- [ ] Self-review realizado
- [ ] CI/CD passou (quando implementado)
- [ ] Merge aprovado
- [ ] Branch feature deletada após merge

**Template de PR:**
```markdown
## Issue #N: [Título]

### O que foi implementado
- Feature X
- Melhoria Y
- Correção Z

### Como testar
1. Passo 1
2. Passo 2
3. Resultado esperado

### Testes realizados
- ✅ 28 unit tests passing
- ✅ Manual test on Kind cluster
- ✅ 3 scenarios validated

### Checklist
- [x] Testes passando
- [x] Build bem-sucedido
- [x] Teste manual validado
- [x] CHANGELOG atualizado
- [x] Documentação atualizada

Closes #N
```

**Após merge:**
```powershell
# Voltar para develop
git checkout develop

# Atualizar local
git pull origin develop

# Deletar branch local
git branch -d feature/issue-N-nome
```

---

## 🏷️ Criação de Release (após merge)

### 9. 📌 Tag e Release
**Objetivo:** Versionar release oficialmente

**Checklist:**
- [ ] Código em develop atualizado
- [ ] Todos os testes passando
- [ ] VERSION atualizado
- [ ] CHANGELOG atualizado
- [ ] Tag criada (vX.Y.Z)
- [ ] Push da tag
- [ ] GitHub Release criada (opcional)

**Comando:**
```powershell
# Criar tag anotada
git tag -a v0.9.1 -m "Release v0.9.1 - Sistema de Severidade

Funcionalidades:
- Enum Severity (CRITICAL, HIGH, MEDIUM, LOW, INFO)
- Score 0-100 com cálculo inteligente
- Ajustes contextuais

Sprint 1 - Issue #9
Data: $(Get-Date -Format 'yyyy-MM-dd')"

# Push tag
git push origin v0.9.1

# (Opcional) Criar release no GitHub
# Via web UI ou GitHub CLI
```

---

## 📊 Métricas de Qualidade

### Critérios de Aceitação (Definition of Done)
Todo código mergeado DEVE atender:

✅ **Testes:**
- Todos os testes unitários passando
- Cobertura mínima 70%
- Testes de integração (quando aplicável)

✅ **Compilação:**
- Build sem erros
- Sem warnings críticos

✅ **Validação:**
- Teste manual em cluster real
- Cenários principais validados

✅ **Documentação:**
- CHANGELOG.md atualizado
- README.md atualizado (se necessário)
- Godoc para código novo

✅ **Versionamento:**
- Commit convencional
- Issue fechada (#N)
- Tag criada (vX.Y.Z)

---

## 🚨 Quando Pular Etapas?

**NUNCA pule:** Etapas 4 (Testes), 5 (Compilação), 6 (Teste Manual)

**Pode simplificar:**
- Etapa 1: Issues muito pequenas (<1h) podem ser informais
- Etapa 8: Hotfixes críticos podem ir direto para develop
- Etapa 9: Patches menores (v0.9.1 → v0.9.2) podem ter tags simples

---

## 🎯 Resumo Visual

```
┌─────────────────────────────────────────────────────────────────┐
│                    DEVELOPMENT WORKFLOW                         │
└─────────────────────────────────────────────────────────────────┘

1. 📝 Issue Created (#N)
        ↓
2. 🌿 Create feature/issue-N-name branch
        ↓
3. 💻 Implement code
        ↓
4. 🧪 Write & run unit tests  ← ❌ MUST PASS
        ↓
5. 🏗️ Build binary            ← ❌ MUST BUILD
        ↓
6. 🔬 Manual test (cluster)   ← ❌ MUST WORK
        ↓
7. 📦 Commit & Push branch
        ↓
8. 🔀 Pull Request → Merge to develop
        ↓
9. 🏷️ Tag vX.Y.Z → Release

┌─────────────────────────────────────────────────────────────────┐
│  ⚠️  IF ANY STEP FAILS: Fix → Repeat from step 4               │
└─────────────────────────────────────────────────────────────────┘
```

---

## 📚 Referências

- **Semantic Versioning:** https://semver.org/
- **Conventional Commits:** https://www.conventionalcommits.org/
- **Git Flow:** https://nvie.com/posts/a-successful-git-branching-model/
- **Go Testing:** https://golang.org/pkg/testing/

---

## 🔄 Revisão do Workflow

Este workflow deve ser revisado a cada Sprint Review para otimizações baseadas em aprendizados.

**Última atualização:** 2025-11-20
**Versão:** 1.0
**Status:** ✅ Validado com Issue #9
