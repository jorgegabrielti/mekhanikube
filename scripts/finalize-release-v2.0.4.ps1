# Script para finalizar release v2.0.4 após merge do PR
# Execute este script APÓS fazer o merge do Pull Request

$ErrorActionPreference = "Stop"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Finalizando Release v2.0.4" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# 1. Volta para a branch main e atualiza
Write-Host "[1/5] Atualizando branch main..." -ForegroundColor Yellow
git checkout main
git pull origin main

# 2. Cria a tag v2.0.4
Write-Host ""
Write-Host "[2/5] Criando tag v2.0.4..." -ForegroundColor Yellow
git tag -a v2.0.4 -m "v2.0.4 - Fix kubeconfig manipulation using PyYAML"

# 3. Faz push da tag
Write-Host ""
Write-Host "[3/5] Enviando tag para o GitHub..." -ForegroundColor Yellow
git push origin v2.0.4

# 4. Deleta a branch de release local e remota
Write-Host ""
Write-Host "[4/5] Limpando branch de release..." -ForegroundColor Yellow
git branch -d release/v2.0.4
git push origin --delete release/v2.0.4

# 5. Exibe informações finais
Write-Host ""
Write-Host "[5/5] Release finalizada!" -ForegroundColor Green
Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Próximos Passos" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "1. Acesse: https://github.com/jorgegabrielti/nautikube/releases/new?tag=v2.0.4" -ForegroundColor White
Write-Host "2. Título: 'v2.0.4 - Fix kubeconfig manipulation'" -ForegroundColor White
Write-Host "3. Copie o conteúdo de 'release-notes-v2.0.4.md' para a descrição" -ForegroundColor White
Write-Host "4. Clique em 'Publish release'" -ForegroundColor White
Write-Host ""
Write-Host "Ou execute o comando abaixo para abrir a página:" -ForegroundColor Yellow
Write-Host "Start-Process 'https://github.com/jorgegabrielti/nautikube/releases/new?tag=v2.0.4'" -ForegroundColor Cyan
Write-Host ""
