# Script para adicionar issues ao GitHub Project
# Nautikube Roadmap v1.0.0 - Project ID: 3

Write-Host "🎯 Adicionando issues ao GitHub Project..." -ForegroundColor Cyan
Write-Host ""

# Verificar se gh CLI está instalado
try {
    $ghVersion = gh --version
    Write-Host "✅ GitHub CLI encontrado" -ForegroundColor Green
} catch {
    Write-Host "❌ GitHub CLI não encontrado. Instale com: winget install GitHub.cli" -ForegroundColor Red
    exit 1
}

# Configurações
$PROJECT_NUMBER = 3
$OWNER = "jorgegabrielti"

# Lista de issues para adicionar (todas as 20 issues técnicas + 4 milestones)
$issues = @(
    # Sprint 1
    13,  # Milestone Sprint 1
    9,   # Sistema de Severidade
    10,  # Exportação JSON
    11,  # Exportação YAML
    12,  # Scanner Deployments
    
    # Sprint 2
    14,  # Milestone Sprint 2
    17,  # Scanner StatefulSets
    19,  # Scanner DaemonSets
    18,  # Scanner ConfigMaps/Secrets
    20,  # Scanner Ingress
    
    # Sprint 3
    16,  # Milestone Sprint 3
    21,  # Interface Abstrata Providers
    22,  # Provider OpenAI
    23,  # Provider Anthropic
    24,  # Provider Gemini
    
    # Sprint 4
    15,  # Milestone Sprint 4
    26,  # Modo CI/CD
    25,  # Testes Integração
    27,  # Testes E2E
    28   # Cobertura >80%
)

Write-Host "📋 Adicionando $($issues.Count) issues ao projeto #$PROJECT_NUMBER" -ForegroundColor Yellow
Write-Host ""

$success = 0
$failed = 0

foreach ($issueNumber in $issues) {
    try {
        Write-Host "   Adicionando issue #$issueNumber... " -NoNewline
        
        # Comando gh para adicionar issue ao projeto
        $result = gh project item-add $PROJECT_NUMBER --owner $OWNER --url "https://github.com/$OWNER/nautikube/issues/$issueNumber" 2>&1
        
        if ($LASTEXITCODE -eq 0) {
            Write-Host "✅" -ForegroundColor Green
            $success++
        } else {
            Write-Host "❌ $result" -ForegroundColor Red
            $failed++
        }
        
        # Pequeno delay para não sobrecarregar API
        Start-Sleep -Milliseconds 200
        
    } catch {
        Write-Host "❌ Erro: $_" -ForegroundColor Red
        $failed++
    }
}

Write-Host ""
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
Write-Host "✅ Sucesso: $success issues" -ForegroundColor Green
Write-Host "❌ Falhas: $failed issues" -ForegroundColor Red
Write-Host ""

if ($success -gt 0) {
    Write-Host "🎉 Issues adicionadas ao projeto!" -ForegroundColor Green
    Write-Host "📊 Visualizar: https://github.com/users/$OWNER/projects/$PROJECT_NUMBER" -ForegroundColor Cyan
} else {
    Write-Host "⚠️  Nenhuma issue foi adicionada. Verifique se você está autenticado:" -ForegroundColor Yellow
    Write-Host "   gh auth status" -ForegroundColor Gray
    Write-Host "   gh auth login" -ForegroundColor Gray
}

Write-Host ""
