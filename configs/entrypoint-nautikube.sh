#!/bin/sh
set -e

echo "⚓ NautiKube - Seu navegador de diagnósticos Kubernetes"
echo ""

# Função para configurar o kubeconfig de forma agnóstica
configure_kubeconfig() {
    if [ ! -f "/root/.kube/config" ]; then
        echo "⚠️  Kubeconfig não encontrado em /root/.kube/config"
        return 1
    fi
    
    echo "🔧 Configurando acesso ao cluster..."
    cp /root/.kube/config /root/.kube/config_mod
    
    # Usa Python para manipular o kubeconfig de forma segura e robusta
    python3 -c "
import yaml
import sys

try:
    # Lê o kubeconfig original
    with open('/root/.kube/config', 'r') as f:
        config = yaml.safe_load(f)

    if not config or 'clusters' not in config:
        print('⚠️  Kubeconfig inválido ou vazio')
        sys.exit(0)

    # Processa cada cluster
    for cluster in config.get('clusters', []):
        if 'cluster' in cluster:
            server = cluster['cluster'].get('server', '')
            
            # Substitui localhost/127.0.0.1 por host.docker.internal (para Docker Desktop/Kind)
            if 'localhost' in server or '127.0.0.1' in server:
                server = server.replace('https://127.0.0.1', 'https://host.docker.internal')
                server = server.replace('https://localhost', 'https://host.docker.internal')
                cluster['cluster']['server'] = server
            
            # Remove certificate-authority-data para evitar erros de CA local
            if 'certificate-authority-data' in cluster['cluster']:
                del cluster['cluster']['certificate-authority-data']
            
            # Adiciona insecure-skip-tls-verify para facilitar conexão local
            cluster['cluster']['insecure-skip-tls-verify'] = True

    # Salva o kubeconfig modificado
    with open('/root/.kube/config_mod', 'w') as f:
        yaml.dump(config, f, default_flow_style=False)
    
    print('✅ Kubeconfig processado com sucesso')

except Exception as e:
    print(f'❌ Erro ao processar kubeconfig: {e}')
    sys.exit(1)
"
    
    export KUBECONFIG=/root/.kube/config_mod
    return 0
}

# Executa configuração
configure_kubeconfig

# --- DETECÇÃO DE PROVEDOR (Feature Avançada) ---
echo ""
echo "🔍 Analisando ambiente..."

SERVER=$(grep -m 1 "server:" /root/.kube/config_mod | awk '{print $2}')
PROVIDER="Desconhecido"
ICON="❓"

if echo "$SERVER" | grep -q "eks.amazonaws.com"; then
    PROVIDER="AWS EKS"
    ICON="☁️ "
elif echo "$SERVER" | grep -q "azmk8s.io"; then
    PROVIDER="Azure AKS"
    ICON="☁️ "
elif echo "$SERVER" | grep -q "googleapis.com"; then
    PROVIDER="Google GKE"
    ICON="☁️ "
elif echo "$SERVER" | grep -q "host.docker.internal"; then
    PROVIDER="Cluster Local (Docker/Kind)"
    ICON="🏠"
elif echo "$SERVER" | grep -q "192.168"; then
    PROVIDER="Cluster Local (LAN)"
    ICON="🏠"
else
    PROVIDER="Cluster Customizado"
    ICON="🌐"
fi

echo "   $ICON Tipo: $PROVIDER"
echo "   🔗 Endpoint: $SERVER"

# --- TESTE DE CONECTIVIDADE ---
echo ""
echo "🔌 Testando conexão..."

if kubectl cluster-info > /dev/null 2>&1; then
    echo "✅ Conectado com sucesso!"
    
    # Coleta métricas básicas
    NODE_COUNT=$(kubectl get nodes --no-headers 2>/dev/null | wc -l || echo "0")
    K8S_VERSION=$(kubectl version --short 2>/dev/null | grep "Server Version" | awk '{print $3}' || echo "N/A")
    
    echo "   📊 Nodes: $NODE_COUNT"
    echo "   🐳 Versão: $K8S_VERSION"
else
    echo "❌ Falha na conexão"
    echo "   ⚠️  O NautiKube não conseguiu falar com o cluster."
    echo "   💡 Dica: Verifique se o cluster está rodando e se o kubeconfig está montado corretamente."
fi

# --- OLLAMA CHECK ---
echo ""
echo "🤖 Verificando IA (Ollama)..."
if curl -s http://host.docker.internal:11434/api/tags > /dev/null 2>&1; then
    echo "✅ Ollama detectado"
else
    echo "⚠️  Ollama não encontrado (IA desativada)"
fi

echo ""
echo "🚀 NautiKube v2.0.5 pronto!"
echo "   Uso: docker exec nautikube nautikube analyze --explain"
echo ""

# Mantém container rodando
tail -f /dev/null
