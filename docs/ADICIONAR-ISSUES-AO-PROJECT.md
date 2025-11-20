# 📋 Guia: Adicionar Issues ao GitHub Project

## Opção 1: Via GitHub CLI (Recomendado - Automático)

### Passo 1: Instalar GitHub CLI
```powershell
# Executar no PowerShell (como Administrador)
.\scripts\install-github-cli.ps1
```

Ou instalar manualmente:
- **Via WinGet:** `winget install --id GitHub.cli`
- **Via Download:** https://github.com/cli/cli/releases/latest

### Passo 2: Autenticar no GitHub
```powershell
# Fechar e reabrir PowerShell após instalação
gh auth login
```

Escolha:
1. `GitHub.com`
2. `HTTPS`
3. `Login with a web browser`
4. Copie o código e cole no navegador

### Passo 3: Adicionar Issues ao Project
```powershell
.\scripts\add-issues-to-project.ps1
```

Isso vai adicionar automaticamente todas as 24 issues ao Project #3.

---

## Opção 2: Via Interface Web (Manual - 5 minutos)

### Passo 1: Abrir o Project
https://github.com/users/jorgegabrielti/projects/3

### Passo 2: Adicionar Issues
1. Clique no botão **"+"** (Add items) no topo da primeira coluna
2. Ou pressione **`Ctrl + Space`**

### Passo 3: Buscar e Adicionar Issues
Na busca que aparecer, digite cada número de issue e pressione Enter:

**Sprint 1:**
```
#13  ← Milestone Sprint 1
#9   ← Sistema de Severidade
#10  ← Exportação JSON
#11  ← Exportação YAML
#12  ← Scanner Deployments
```

**Sprint 2:**
```
#14  ← Milestone Sprint 2
#17  ← Scanner StatefulSets
#19  ← Scanner DaemonSets
#18  ← Scanner ConfigMaps/Secrets
#20  ← Scanner Ingress
```

**Sprint 3:**
```
#16  ← Milestone Sprint 3
#21  ← Interface Abstrata Providers
#22  ← Provider OpenAI
#23  ← Provider Anthropic
#24  ← Provider Gemini
```

**Sprint 4:**
```
#15  ← Milestone Sprint 4
#26  ← Modo CI/CD
#25  ← Testes Integração
#27  ← Testes E2E
#28  ← Cobertura >80%
```

### Passo 4: Organizar nas Colunas (Opcional)
Arraste as issues para as colunas apropriadas:
- **Backlog:** Issues #10-28 (exceto #9)
- **Sprint Atual:** Issue #9 (primeira do Sprint 1)

---

## Opção 3: Via API GitHub (Para Desenvolvedores)

```powershell
# Configurar variáveis
$GITHUB_TOKEN = "seu_token_aqui"
$PROJECT_ID = "PVT_kwDOQSZ4LM4AzXy8"  # ID do Project (v2)

# Exemplo para adicionar issue #9
gh api graphql -f query='
  mutation {
    addProjectV2ItemById(input: {
      projectId: "PVT_kwDOQSZ4LM4AzXy8"
      contentId: "I_kwDOQSZ4LM56vBLh"
    }) {
      item {
        id
      }
    }
  }
'
```

**Nota:** Você precisa obter o Node ID de cada issue. Use o script automático (Opção 1) que já faz isso.

---

## ✅ Validação

Após adicionar as issues, verifique:

1. **Total de issues no project:** 24
   - 4 Milestones (tracking)
   - 20 Issues técnicas

2. **Todas as issues aparecem** na view do projeto:
   https://github.com/users/jorgegabrielti/projects/3

3. **Issues organizadas** (opcional mas recomendado):
   - Coluna "Sprint Atual": #9
   - Coluna "Backlog": Resto das issues

---

## 🆘 Troubleshooting

### Erro: "gh: command not found"
**Solução:** Feche e reabra o PowerShell após instalar o GitHub CLI.

### Erro: "authentication required"
**Solução:** Execute `gh auth login` e autentique via navegador.

### Erro: "project not found"
**Solução:** Verifique se o Project #3 existe em https://github.com/users/jorgegabrielti/projects

### Erro: "issue not found"
**Solução:** Verifique se todas as issues foram criadas corretamente em https://github.com/jorgegabrielti/nautikube/issues

---

## 📊 Próximos Passos

Após adicionar as issues ao project:

1. ✅ Configurar colunas do Kanban (se ainda não fez):
   - Backlog
   - Sprint Atual
   - Em Desenvolvimento
   - Revisão
   - Concluído

2. ✅ Mover Issue #9 para "Em Desenvolvimento" (primeira tarefa)

3. ✅ Começar a trabalhar! 🚀
   - Ler a issue #9 completa
   - Implementar sistema de severidade
   - Fazer commit e tag v0.9.1

---

**Dica:** Depois que adicionar as issues uma vez, não precisa fazer de novo. O Project vai manter as issues organizadas automaticamente! 🎯
