# 📦 Guia de Release do NautiKube

Este documento descreve o processo completo de criação de releases do NautiKube.

## 📋 Índice

- [Visão Geral](#visão-geral)
- [Pré-requisitos](#pré-requisitos)
- [Processo de Release](#processo-de-release)
- [Versionamento](#versionamento)
- [Troubleshooting](#troubleshooting)

---

## 🎯 Visão Geral

O NautiKube usa um processo de release automatizado que:

1. ✅ Atualiza arquivos de versão (VERSION, main.go, CHANGELOG.md)
2. ✅ Cria commit e tag Git
3. ✅ Dispara GitHub Actions ao fazer push da tag
4. ✅ Compila binários para múltiplas plataformas
5. ✅ Publica Docker images no GitHub Container Registry
6. ✅ Cria release no GitHub com assets

**Processo completo:** `~5-10 minutos` (a maioria automatizado)

---

## 🔧 Pré-requisitos

### Localmente

- [x] Git configurado
- [x] Go 1.21+ instalado
- [x] Docker Desktop rodando
- [x] Permissões de escrita no repositório

### GitHub

- [x] Acesso ao repositório `jorgegabrielti/nautikube`
- [x] GitHub Actions habilitado
- [x] Token `GITHUB_TOKEN` (configurado automaticamente)

---

## 🚀 Processo de Release

### 1. Preparação

Antes de criar uma release, certifique-se que:

```powershell
# Todas as mudanças estão commitadas
git status

# Testes passando
go test ./...

# Build funciona
docker-compose build

# Funcionalidade OK
docker-compose up -d
docker exec nautikube nautikube analyze --explain
```

### 2. Atualizar CHANGELOG.md

Edite `CHANGELOG.md` e mova as mudanças de `[Unreleased]` para uma nova seção de versão:

```markdown
## [Unreleased]

## [2.0.3] - 2025-11-12

### Adicionado
- Nova feature X
- Melhoria Y

### Corrigido
- Bug Z
```

**Dica:** O script de release pode fazer isso automaticamente!

### 3. Executar Script de Release

#### Opção A: Modo Interativo (Recomendado)

```powershell
# Teste primeiro (dry-run)
.\scripts\release.ps1 -Version 2.0.3 -DryRun

# Executar de verdade
.\scripts\release.ps1 -Version 2.0.3

# Ou com push automático
.\scripts\release.ps1 -Version 2.0.3 -Push
```

O script irá:
- ✅ Atualizar `VERSION` file
- ✅ Atualizar `cmd/nautikube/main.go`
- ✅ Atualizar `CHANGELOG.md` (se tiver [Unreleased])
- ✅ Criar commit: `chore: release v2.0.3`
- ✅ Criar tag: `v2.0.3`
- ✅ Fazer push (se `-Push` foi usado)

#### Opção B: Manual

Se preferir fazer manualmente:

```powershell
# 1. Atualizar VERSION
"2.0.3" | Out-File -FilePath VERSION -NoNewline

# 2. Atualizar main.go
# Edite cmd/nautikube/main.go e mude:
# Version = "2.0.2" → Version = "2.0.3"

# 3. Commitar
git add VERSION cmd/nautikube/main.go CHANGELOG.md
git commit -m "chore: release v2.0.3"

# 4. Criar tag
git tag -a v2.0.3 -m "Release v2.0.3"

# 5. Push
git push origin main
git push origin v2.0.3
```

### 4. Aguardar GitHub Actions

Após o push da tag, o GitHub Actions automaticamente:

1. **Compila binários** para:
   - Linux (AMD64, ARM64)
   - macOS (AMD64, ARM64/Apple Silicon)
   - Windows (AMD64)

2. **Gera checksums** (SHA256)

3. **Cria release** no GitHub com:
   - Release notes extraídas do CHANGELOG.md
   - Binários anexados
   - Arquivo de checksums

4. **Publica Docker images**:
   - `ghcr.io/jorgegabrielti/nautikube:2.0.3`
   - `ghcr.io/jorgegabrielti/nautikube:latest`

**Tempo estimado:** 5-8 minutos

### 5. Validação

Após o Actions completar, valide:

```powershell
# Verificar release criada
# https://github.com/jorgegabrielti/nautikube/releases

# Testar binário Windows
Invoke-WebRequest -Uri "https://github.com/jorgegabrielti/nautikube/releases/download/v2.0.3/nautikube-windows-amd64.exe" -OutFile nautikube.exe
.\nautikube.exe version

# Testar Docker image
docker pull ghcr.io/jorgegabrielti/nautikube:2.0.3
docker run --rm ghcr.io/jorgegabrielti/nautikube:2.0.3 version
```

### 6. Pós-Release

- [ ] Atualizar README.md se necessário (badges são automáticos)
- [ ] Fechar issues/PRs resolvidos nesta versão
- [ ] Anunciar release (se apropriado)
- [ ] Criar milestone para próxima versão

---

## 🔢 Versionamento

O NautiKube segue [Semantic Versioning 2.0.0](https://semver.org/):

```
MAJOR.MINOR.PATCH
```

### Quando incrementar cada número:

**MAJOR** (ex: 2.0.0 → 3.0.0)
- Mudanças incompatíveis na API
- Remoção de features
- Mudanças que quebram compatibilidade

**MINOR** (ex: 2.0.0 → 2.1.0)
- Novas features (compatíveis)
- Melhorias significativas
- Deprecações

**PATCH** (ex: 2.0.0 → 2.0.1)
- Correção de bugs
- Melhorias de performance
- Atualizações de documentação
- Refatorações internas

### Exemplos Práticos

```
2.0.0 → 2.0.1  ✅ Correção de bug no scanner
2.0.1 → 2.1.0  ✅ Adição de suporte a GPU
2.1.0 → 3.0.0  ✅ Mudança na interface CLI (breaking change)
```

---

## 🧪 Testando Localmente

Antes de criar a release, você pode testar o build local:

```powershell
# Build manual dos binários
$Version = "2.0.3"

# Linux AMD64
$env:GOOS="linux"; $env:GOARCH="amd64"; $env:CGO_ENABLED="0"
go build -ldflags="-w -s" -o "nautikube-linux-amd64" ./cmd/nautikube

# Windows AMD64
$env:GOOS="windows"; $env:GOARCH="amd64"; $env:CGO_ENABLED="0"
go build -ldflags="-w -s" -o "nautikube-windows-amd64.exe" ./cmd/nautikube

# Testar
.\nautikube-windows-amd64.exe version
```

---

## 🐛 Troubleshooting

### Erro: "Tag já existe"

```powershell
# Remover tag local
git tag -d v2.0.3

# Remover tag remota (CUIDADO!)
git push --delete origin v2.0.3
```

### GitHub Actions falhou

1. **Verificar logs**: GitHub → Actions → Release workflow
2. **Causas comuns**:
   - VERSION file não corresponde à tag
   - Erro de compilação (teste com `go build`)
   - Erro no CHANGELOG.md (formato inválido)

### Binários não foram anexados

- Verificar se workflow completou 100%
- Verificar permissões: Settings → Actions → General → Workflow permissions
- Re-executar workflow manualmente

### Docker image não foi publicada

```powershell
# Verificar se existe
docker pull ghcr.io/jorgegabrielti/nautikube:2.0.3

# Se não existir, fazer build manual
docker build -t ghcr.io/jorgegabrielti/nautikube:2.0.3 -f configs/Dockerfile.nautikube .
docker push ghcr.io/jorgegabrielti/nautikube:2.0.3
```

### Reverter Release

Se algo deu muito errado:

```powershell
# 1. Deletar tag local
git tag -d v2.0.3

# 2. Deletar tag remota
git push --delete origin v2.0.3

# 3. Deletar release no GitHub (manualmente via interface)

# 4. Reverter commit
git revert HEAD
git push origin main
```

---

## 📚 Referências

- [Keep a Changelog](https://keepachangelog.com/)
- [Semantic Versioning](https://semver.org/)
- [GitHub Actions - Creating Releases](https://docs.github.com/en/actions/publishing-packages/about-packaging-with-github-actions)
- [Docker Multi-platform Builds](https://docs.docker.com/build/building/multi-platform/)

---

## 🤝 Contribuindo

Se você encontrar problemas ou tiver sugestões para melhorar o processo de release:

1. Abra uma issue: [github.com/jorgegabrielti/nautikube/issues](https://github.com/jorgegabrielti/nautikube/issues)
2. Proponha mudanças via PR
3. Documente no CHANGELOG.md

---

**Última atualização:** 2025-11-11
**Versão do documento:** 1.0
