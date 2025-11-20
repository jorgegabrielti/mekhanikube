# GitFlow - Guia Rápido NautiKube

## 📋 Fluxo de Release

### 1. Criar Branch de Release

```powershell
# A partir da main (ou develop)
git checkout main
git pull origin main

# Criar branch de release
git checkout -b release/vX.Y.Z

# Atualizar VERSION
echo "X.Y.Z" > VERSION

# Atualizar CHANGELOG.md
# Adicionar entrada para a nova versão

# Atualizar version no entrypoint
# configs/entrypoint-nautikube.sh: linha ~107

# Commit
git add -A
git commit -m "release: vX.Y.Z - [descrição]"

# Push
git push origin release/vX.Y.Z
```

### 2. Criar Pull Request

```powershell
# Abrir página de PR
Start-Process "https://github.com/jorgegabrielti/nautikube/compare/main...release/vX.Y.Z?expand=1"

# Ou criar via script (se tiver gh CLI)
gh pr create --base main --head release/vX.Y.Z --title "Release vX.Y.Z - [título]" --body-file release-notes-vX.Y.Z.md
```

### 3. Após Merge do PR

```powershell
# Executar script de finalização
.\scripts\finalize-release.ps1 -Version "X.Y.Z"

# Ou manualmente:
git checkout main
git pull origin main
git tag -a vX.Y.Z -m "vX.Y.Z - [descrição]"
git push origin vX.Y.Z
git branch -d release/vX.Y.Z
git push origin --delete release/vX.Y.Z
```

### 4. Publicar Release no GitHub

```powershell
# Abrir página de criação de release
Start-Process "https://github.com/jorgegabrielti/nautikube/releases/new?tag=vX.Y.Z"

# Preencher:
# - Título: "vX.Y.Z - [título]"
# - Descrição: Copiar de release-notes-vX.Y.Z.md
# - Clicar em "Publish release"
```

## 🌿 Branches

- **`main`**: Código de produção (sempre estável)
- **`release/vX.Y.Z`**: Preparação de release
- **`feature/*`**: Novas funcionalidades
- **`hotfix/vX.Y.Z`**: Correções urgentes em produção

## 📝 Versionamento Semântico

- **MAJOR** (X.0.0): Mudanças incompatíveis na API
- **MINOR** (x.Y.0): Novas funcionalidades (compatíveis)
- **PATCH** (x.y.Z): Correções de bugs

## 🔧 Scripts Disponíveis

- `scripts/finalize-release.ps1`: Finaliza release após merge do PR
- `scripts/finalize-release-vX.Y.Z.ps1`: Versão específica para uma release

## 📚 Referências

- [GitFlow](https://nvie.com/posts/a-successful-git-branching-model/)
- [Semantic Versioning](https://semver.org/)
- [Keep a Changelog](https://keepachangelog.com/)
