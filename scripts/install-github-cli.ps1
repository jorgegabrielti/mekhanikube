# Instalador do GitHub CLI (gh)
# Nautikube Project Setup

Write-Host "🔧 Instalando GitHub CLI..." -ForegroundColor Cyan
Write-Host ""

# Verificar se winget está disponível
try {
    $wingetVersion = winget --version
    Write-Host "✅ WinGet encontrado: $wingetVersion" -ForegroundColor Green
    Write-Host ""
    
    # Instalar GitHub CLI
    Write-Host "📦 Instalando GitHub.cli via WinGet..." -ForegroundColor Yellow
    winget install --id GitHub.cli
    
    Write-Host ""
    Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
    Write-Host "✅ GitHub CLI instalado!" -ForegroundColor Green
    Write-Host ""
    Write-Host "⚠️  IMPORTANTE: Feche e reabra o PowerShell para usar o comando 'gh'" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "📋 Próximos passos:" -ForegroundColor Cyan
    Write-Host "   1. Feche este PowerShell" -ForegroundColor Gray
    Write-Host "   2. Abra um novo PowerShell" -ForegroundColor Gray
    Write-Host "   3. Execute: gh auth login" -ForegroundColor Gray
    Write-Host "   4. Execute: .\scripts\add-issues-to-project.ps1" -ForegroundColor Gray
    Write-Host ""
    
} catch {
    Write-Host "❌ WinGet não encontrado." -ForegroundColor Red
    Write-Host ""
    Write-Host "📥 Instalação manual:" -ForegroundColor Yellow
    Write-Host "   1. Baixe: https://github.com/cli/cli/releases/latest" -ForegroundColor Gray
    Write-Host "   2. Execute o instalador" -ForegroundColor Gray
    Write-Host "   3. Reinicie o PowerShell" -ForegroundColor Gray
    Write-Host "   4. Execute: gh auth login" -ForegroundColor Gray
    Write-Host ""
}
