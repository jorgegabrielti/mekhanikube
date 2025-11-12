<#
.SYNOPSIS
    Script de release do NautiKube
    
.DESCRIPTION
    Automatiza o processo de criação de release:
    - Atualiza VERSION file
    - Atualiza version no main.go
    - Atualiza CHANGELOG.md
    - Cria commit e tag
    - Faz push (opcional)
    
.PARAMETER Version
    Nova versão no formato semver (ex: 2.0.3)
    
.PARAMETER Push
    Se especificado, faz push automático da tag
    
.PARAMETER DryRun
    Executa sem fazer mudanças reais (teste)
    
.EXAMPLE
    .\release.ps1 -Version 2.0.3
    
.EXAMPLE
    .\release.ps1 -Version 2.0.3 -Push
    
.EXAMPLE
    .\release.ps1 -Version 2.0.3 -DryRun
#>

param(
    [Parameter(Mandatory=$true)]
    [ValidatePattern('^\d+\.\d+\.\d+$')]
    [string]$Version,
    
    [switch]$Push,
    [switch]$DryRun
)

$ErrorActionPreference = "Stop"

# Cores
function Write-ColorOutput {
    param(
        [string]$Message,
        [string]$Color = "White"
    )
    Write-Host $Message -ForegroundColor $Color
}

# Banner
Write-ColorOutput "`n🚢 NautiKube Release Script v2.0" -Color Cyan
Write-ColorOutput "================================`n" -Color Cyan

# Verificar se está no diretório correto
if (-not (Test-Path "VERSION")) {
    Write-ColorOutput "❌ Erro: Execute o script na raiz do projeto" -Color Red
    exit 1
}

# Ler versão atual
$currentVersion = Get-Content VERSION -Raw
$currentVersion = $currentVersion.Trim()

Write-ColorOutput "📋 Versão atual: $currentVersion" -Color Yellow
Write-ColorOutput "📋 Nova versão:  $Version`n" -Color Green

# Validar se não é a mesma versão
if ($currentVersion -eq $Version) {
    Write-ColorOutput "⚠️  Aviso: Nova versão é igual à atual!" -Color Yellow
    $continue = Read-Host "Continuar mesmo assim? (s/N)"
    if ($continue -ne 's' -and $continue -ne 'S') {
        exit 0
    }
}

# Verificar se git está limpo
$gitStatus = git status --porcelain
if ($gitStatus -and -not $DryRun) {
    Write-ColorOutput "⚠️  Aviso: Existem mudanças não commitadas" -Color Yellow
    git status --short
    Write-Host ""
    $continue = Read-Host "Continuar mesmo assim? (s/N)"
    if ($continue -ne 's' -and $continue -ne 'S') {
        exit 0
    }
}

# Verificar se a tag já existe
$tagExists = git tag -l "v$Version"
if ($tagExists) {
    Write-ColorOutput "❌ Erro: Tag v$Version já existe!" -Color Red
    exit 1
}

# Confirmação final
if (-not $DryRun) {
    Write-ColorOutput "`n⚠️  Isso irá:" -Color Yellow
    Write-Host "  1. Atualizar VERSION file"
    Write-Host "  2. Atualizar cmd/nautikube/main.go"
    Write-Host "  3. Criar commit 'chore: release v$Version'"
    Write-Host "  4. Criar tag 'v$Version'"
    if ($Push) {
        Write-Host "  5. Fazer push da tag (DISPARA RELEASE NO GITHUB!)"
    }
    Write-Host ""
    $confirm = Read-Host "Confirma a criação da release v$Version? (s/N)"
    if ($confirm -ne 's' -and $confirm -ne 'S') {
        Write-ColorOutput "`n❌ Release cancelada pelo usuário" -Color Yellow
        exit 0
    }
}

Write-ColorOutput "`n🔧 Iniciando processo de release...`n" -Color Cyan

# 1. Atualizar VERSION file
Write-ColorOutput "1️⃣  Atualizando VERSION file..." -Color Cyan
if (-not $DryRun) {
    Set-Content -Path "VERSION" -Value $Version -NoNewline
    Write-ColorOutput "   ✅ VERSION atualizado para $Version" -Color Green
} else {
    Write-ColorOutput "   [DRY-RUN] VERSION seria atualizado para $Version" -Color Yellow
}

# 2. Atualizar main.go
Write-ColorOutput "2️⃣  Atualizando cmd/nautikube/main.go..." -Color Cyan
$mainGoPath = "cmd\nautikube\main.go"
if (Test-Path $mainGoPath) {
    if (-not $DryRun) {
        $content = Get-Content $mainGoPath -Raw
        $content = $content -replace 'Version = "\d+\.\d+\.\d+"', "Version = `"$Version`""
        Set-Content -Path $mainGoPath -Value $content -NoNewline
        Write-ColorOutput "   ✅ main.go atualizado" -Color Green
    } else {
        Write-ColorOutput "   [DRY-RUN] main.go seria atualizado" -Color Yellow
    }
} else {
    Write-ColorOutput "   ⚠️  main.go não encontrado" -Color Yellow
}

# 3. Atualizar CHANGELOG.md (se existir entrada [Unreleased])
Write-ColorOutput "3️⃣  Verificando CHANGELOG.md..." -Color Cyan
if (Test-Path "CHANGELOG.md") {
    if (-not $DryRun) {
        $changelog = Get-Content "CHANGELOG.md" -Raw
        $today = Get-Date -Format "yyyy-MM-dd"
        
        # Substitui [Unreleased] por [Version] - Data
        $changelog = $changelog -replace '\[Unreleased\]', "[$Version] - $today"
        
        # Adiciona nova seção Unreleased no topo
        $unreleasedSection = @"
## [Unreleased]

### Adicionado
### Modificado
### Corrigido
### Removido

## 
"@
        $changelog = $changelog -replace '(# Changelog\s+)', "`$1`n$unreleasedSection"
        
        Set-Content -Path "CHANGELOG.md" -Value $changelog -NoNewline
        Write-ColorOutput "   ✅ CHANGELOG.md atualizado" -Color Green
    } else {
        Write-ColorOutput "   [DRY-RUN] CHANGELOG.md seria atualizado" -Color Yellow
    }
} else {
    Write-ColorOutput "   ⚠️  CHANGELOG.md não encontrado" -Color Yellow
}

# 4. Criar commit
Write-ColorOutput "4️⃣  Criando commit..." -Color Cyan
if (-not $DryRun) {
    git add VERSION cmd/nautikube/main.go CHANGELOG.md 2>$null
    git commit -m "chore: release v$Version" --no-verify
    Write-ColorOutput "   ✅ Commit criado" -Color Green
} else {
    Write-ColorOutput "   [DRY-RUN] Commit seria criado: 'chore: release v$Version'" -Color Yellow
}

# 5. Criar tag
Write-ColorOutput "5️⃣  Criando tag v$Version..." -Color Cyan
if (-not $DryRun) {
    git tag -a "v$Version" -m "Release v$Version"
    Write-ColorOutput "   ✅ Tag v$Version criada" -Color Green
} else {
    Write-ColorOutput "   [DRY-RUN] Tag v$Version seria criada" -Color Yellow
}

# 6. Push (se solicitado)
if ($Push -and -not $DryRun) {
    Write-ColorOutput "6️⃣  Fazendo push..." -Color Cyan
    git push origin main
    git push origin "v$Version"
    Write-ColorOutput "   ✅ Push realizado (release será criada automaticamente!)" -Color Green
} elseif ($Push -and $DryRun) {
    Write-ColorOutput "6️⃣  [DRY-RUN] Push seria realizado" -Color Yellow
} else {
    Write-ColorOutput "`n⚠️  Tag criada localmente. Para publicar, execute:" -Color Yellow
    Write-ColorOutput "   git push origin main" -Color White
    Write-ColorOutput "   git push origin v$Version" -Color White
}

# Resumo final
Write-ColorOutput "`n✅ Release v$Version preparada com sucesso!`n" -Color Green

if (-not $Push -and -not $DryRun) {
    Write-ColorOutput "📋 Próximos passos:" -Color Cyan
    Write-ColorOutput "   1. Revisar as mudanças: git show" -Color White
    Write-ColorOutput "   2. Fazer push: git push origin main && git push origin v$Version" -Color White
    Write-ColorOutput "   3. Aguardar GitHub Actions criar a release" -Color White
    Write-ColorOutput "   4. Verificar: https://github.com/jorgegabrielti/nautikube/releases`n" -Color White
}

if ($DryRun) {
    Write-ColorOutput "`n💡 Dry-run concluído. Nenhuma mudança foi feita." -Color Cyan
    Write-ColorOutput "   Execute sem -DryRun para aplicar as mudanças.`n" -Color White
}
