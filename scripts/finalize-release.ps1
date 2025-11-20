# Script genérico para finalizar releases seguindo GitFlow
# Uso: .\finalize-release.ps1 -Version "2.0.5"

param(
    [Parameter(Mandatory = $true)]
    [string]$Version
)

$ErrorActionPreference = "Stop"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Finalizando Release v$Version" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Valida se estamos na branch correta
$currentBranch = git branch --show-current
if ($currentBranch -ne "release/v$Version") {
    Write-Host "AVISO: Você não está na branch release/v$Version" -ForegroundColor Yellow
    Write-Host "Branch atual: $currentBranch" -ForegroundColor Yellow
    $continue = Read-Host "Deseja continuar mesmo assim? (s/N)"
    if ($continue -ne "s") {
        Write-Host "Operação cancelada." -ForegroundColor Red
        exit 1
    }
}

# 1. Volta para a branch main e atualiza
Write-Host "[1/6] Atualizando branch main..." -ForegroundColor Yellow
git checkout main
git pull origin main

# 2. Verifica se o merge do PR foi feito
$lastCommit = git log -1 --pretty=%B
Write-Host "Último commit na main: $lastCommit" -ForegroundColor Gray
Write-Host ""

# 3. Cria a tag
Write-Host "[2/6] Criando tag v$Version..." -ForegroundColor Yellow
git tag -a "v$Version" -m "v$Version - Release"

# 4. Faz push da tag
Write-Host ""
Write-Host "[3/6] Enviando tag para o GitHub..." -ForegroundColor Yellow
git push origin "v$Version"

# 5. Deleta a branch de release local
Write-Host ""
Write-Host "[4/6] Limpando branch de release local..." -ForegroundColor Yellow
git branch -d "release/v$Version"

# 6. Deleta a branch de release remota
Write-Host ""
Write-Host "[5/6] Limpando branch de release remota..." -ForegroundColor Yellow
git push origin --delete "release/v$Version"

# 7. Exibe informações finais
Write-Host ""
Write-Host "[6/6] Release finalizada!" -ForegroundColor Green
Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Próximos Passos" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "1. Criar release no GitHub:" -ForegroundColor White
Write-Host "   https://github.com/jorgegabrielti/nautikube/releases/new?tag=v$Version" -ForegroundColor Cyan
Write-Host ""
Write-Host "2. Ou execute:" -ForegroundColor White
Write-Host "   Start-Process 'https://github.com/jorgegabrielti/nautikube/releases/new?tag=v$Version'" -ForegroundColor Cyan
Write-Host ""
