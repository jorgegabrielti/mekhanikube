#!/bin/sh

echo "🚀 Iniciando K8sGPT MCP Server..."

# Ajusta o kubeconfig substituindo 127.0.0.1 por host.docker.internal
if [ -f /root/.kube/config ]; then
    echo "📝 Ajustando kubeconfig..."
    sed 's/127\.0\.0\.1/host.docker.internal/g' /root/.kube/config > /root/.kube/config_mod
    export KUBECONFIG=/root/.kube/config_mod
fi

# Verificar conexão com cluster Kubernetes
echo "🔍 Verificando conexão com cluster Kubernetes..."
if kubectl cluster-info >/dev/null 2>&1; then
    echo "✅ Cluster Kubernetes acessível!"
else
    echo "⚠️  Aviso: Não foi possível conectar ao cluster Kubernetes"
    echo "   Verifique se o kubeconfig está correto"
fi

# Configurar backend fake (K8sGPT exige, mas MCP não usará - Copilot fará a IA)
echo "⚙️  Configurando backend fake (apenas para validação do K8sGPT)..."
k8sgpt auth add --backend openai --model gpt-3.5-turbo --password fake-key-not-used 2>/dev/null || true
k8sgpt auth default -p openai 2>/dev/null || true
echo "✅ Backend configurado (não será usado - Copilot fará todo o trabalho de IA)"

# Inicia o servidor MCP (sem backend de IA - Copilot fará o trabalho)
echo "🎯 Iniciando servidor MCP na porta 3000..."
echo "💡 GitHub Copilot será responsável pela inteligência artificial"
echo ""

exec k8sgpt serve --mcp
