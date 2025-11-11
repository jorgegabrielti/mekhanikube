#!/bin/sh
set -e

echo "⚓ NautiKube - Seu navegador de diagnósticos Kubernetes"
echo ""

# Detectar tipo de ambiente
CLUSTER_TYPE="local"
if [ -d "/root/.aws" ] && [ "$(ls -A /root/.aws 2>/dev/null)" ]; then
    if [ -f "/usr/bin/aws" ]; then
        CLUSTER_TYPE="eks"
        echo "🔍 Ambiente detectado: EKS (AWS)"
    fi
else
    echo "🔍 Ambiente detectado: Kubernetes Local"
fi

# Ajusta kubeconfig para funcionar dentro do container
if [ -f "/root/.kube/config" ]; then
    echo "📋 Configurando acesso ao cluster..."
    
    # Cria versão modificada do kubeconfig
    sed 's|https://127.0.0.1|https://host.docker.internal|g; s|https://localhost|https://host.docker.internal|g' \
        /root/.kube/config > /root/.kube/config_mod
    
    export KUBECONFIG=/root/.kube/config_mod
    echo "✅ Kubeconfig configurado"
    
    # Se EKS, verifica credenciais AWS
    if [ "$CLUSTER_TYPE" = "eks" ]; then
        echo "✅ Credenciais AWS disponíveis"
    fi
else
    echo "⚠️  Kubeconfig não encontrado em /root/.kube/config"
    echo "   O container pode não conseguir acessar o cluster"
fi

# Verifica conectividade com o cluster
echo ""
echo "🔍 Verificando conectividade com o cluster..."
if kubectl cluster-info > /dev/null 2>&1; then
    echo "✅ Cluster acessível"
    kubectl get nodes --no-headers 2>/dev/null | wc -l | xargs -I {} echo "   {} node(s) encontrado(s)"
else
    echo "❌ Não foi possível conectar ao cluster"
    echo "   Verifique se o kubeconfig está correto"
fi

# Verifica conectividade com Ollama
echo ""
echo "🤖 Verificando Ollama..."
if curl -s http://host.docker.internal:11434/api/tags > /dev/null 2>&1; then
    echo "✅ Ollama acessível em http://host.docker.internal:11434"
else
    echo "⚠️  Ollama não acessível"
    echo "   Use 'analyze' sem --explain para análise básica"
fi

echo ""
echo "🚀 NautiKube pronto!"
echo "   Teste com: docker exec nautikube nautikube analyze --explain --language Portuguese"
echo ""

# Mantém container rodando
tail -f /dev/null
