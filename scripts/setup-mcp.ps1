# Script de configuração MCP para GitHub Copilot
# Este script configura a integração entre Mekhanikube e GitHub Copilot via MCP

Write-Host "🔧 Configurando Mekhanikube MCP para GitHub Copilot..." -ForegroundColor Cyan
Write-Host ""

# Definir configuração MCP
$mcpConfig = @{
    mcpServers = @{
        mekhanikube = @{
            command = "docker"
            args = @(
                "exec",
                "-i",
                "mekhanikube-k8sgpt-mcp",
                "k8sgpt",
                "serve",
                "--mcp"
            )
        }
    }
} | ConvertTo-Json -Depth 10

# Determinar caminho do arquivo de configuração do VS Code
$vscodeConfigPath = "$env:APPDATA\Code\User"
$copilotConfigFile = "$vscodeConfigPath\globalStorage\github.copilot-chat\mcpServers.json"

# Criar diretório se não existir
$copilotConfigDir = Split-Path -Parent $copilotConfigFile
if (-not (Test-Path $copilotConfigDir)) {
    New-Item -ItemType Directory -Path $copilotConfigDir -Force | Out-Null
}

# Verificar se já existe configuração
if (Test-Path $copilotConfigFile) {
    Write-Host "⚠️  Arquivo de configuração já existe: $copilotConfigFile" -ForegroundColor Yellow
    Write-Host ""
    $backup = "${copilotConfigFile}.backup"
    Copy-Item $copilotConfigFile $backup -Force
    Write-Host "✅ Backup criado: $backup" -ForegroundColor Green
}

# Salvar configuração
$mcpConfig | Out-File -FilePath $copilotConfigFile -Encoding utf8 -Force

Write-Host ""
Write-Host "✅ Configuração MCP criada com sucesso!" -ForegroundColor Green
Write-Host ""
Write-Host "📁 Arquivo: $copilotConfigFile" -ForegroundColor Gray
Write-Host ""
Write-Host "🚀 Próximos passos:" -ForegroundColor Cyan
Write-Host ""
Write-Host "1. Iniciar serviço MCP:" -ForegroundColor White
Write-Host "   docker-compose --profile mcp up -d" -ForegroundColor Gray
Write-Host ""
Write-Host "2. Reiniciar VS Code" -ForegroundColor White
Write-Host ""
Write-Host "3. Abrir GitHub Copilot Chat e testar:" -ForegroundColor White
Write-Host "   'Analise meu cluster Kubernetes'" -ForegroundColor Gray
Write-Host "   'Quais problemas existem no namespace default?'" -ForegroundColor Gray
Write-Host ""
Write-Host "📖 Documentação completa: docs/MCP.md" -ForegroundColor Cyan
Write-Host ""
