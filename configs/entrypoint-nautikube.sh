#!/bin/sh
set -e

echo "⚓ NautiKube - Seu navegador de diagnósticos Kubernetes"
echo ""

# Função para ajustar kubeconfig de forma agnóstica
configure_kubeconfig() {
    if [ ! -f "/root/.kube/config" ]; then
        echo "⚠️  Kubeconfig não encontrado em /root/.kube/config"
        return 1
    fi
    
    echo "🔧 Configurando acesso ao cluster..."
    cp /root/.kube/config /root/.kube/config_mod
    
    # Ajuste simples para clusters locais (localhost -> host.docker.internal)
    sed -i 's|https://127.0.0.1|https://host.docker.internal|g; \
            s|https://localhost|https://host.docker.internal|g' \
        /root/.kube/config_mod
        
    # Remove CA e adiciona insecure-skip-tls-verify para desenvolvimento local
    sed -i '/certificate-authority-data:/d' /root/.kube/config_mod
    sed -i '/server: https:\/\/host.docker.internal/a\    insecure-skip-tls-verify: true' \
        /root/.kube/config_mod
    
    export KUBECONFIG=/root/.kube/config_mod
    echo "✅ Kubeconfig configurado"
    return 0
}

# Configura o kubeconfig
configure_kubeconfig

# Verificação inteligente de conectividade
echo ""
echo "🔍 Testando conectividade com o cluster..."

# Primeira tentativa: conexão direta
if kubectl cluster-info > /dev/null 2>&1; then
    echo "✅ Cluster acessível!"
    
    # Informações do cluster
    NODE_COUNT=$(kubectl get nodes --no-headers 2>/dev/null | wc -l || echo "0")
    CURRENT_CONTEXT=$(kubectl config current-context 2>/dev/null || echo "N/A")
    K8S_VERSION=$(kubectl version --short 2>/dev/null | grep "Server Version" | awk '{print $3}' || echo "N/A")
    
    echo "   📊 Nodes: $NODE_COUNT"
    echo "   🎯 Contexto: $CURRENT_CONTEXT"
    echo "   🐳 Versão K8s: $K8S_VERSION"
else
    echo "⚠️  Primeira tentativa falhou, tentando estratégias alternativas..."
    
    # Estratégia 2: Limpa duplicatas e garante insecure-skip-tls-verify
    echo "   🔄 Limpando configuração e forçando insecure-skip-tls-verify..."
    
    # Remove todas as linhas duplicadas de server e insecure
    awk '!seen[$0]++' /root/.kube/config_mod > /root/.kube/config_mod.tmp
    mv /root/.kube/config_mod.tmp /root/.kube/config_mod
    
    # Remove CA completamente
    sed -i '/certificate-authority-data:/d' /root/.kube/config_mod
    
    # Garante que insecure está presente se ainda não estiver
    if ! grep -q "insecure-skip-tls-verify: true" /root/.kube/config_mod; then
        sed -i '/server: https:\/\//a\    insecure-skip-tls-verify: true' /root/.kube/config_mod
    fi
    
    if kubectl cluster-info > /dev/null 2>&1; then
        echo "   ✅ Conectado após ajustes!"
        
        # Mostra informações do cluster
        NODE_COUNT=$(kubectl get nodes --no-headers 2>/dev/null | wc -l || echo "0")
        CURRENT_CONTEXT=$(kubectl config current-context 2>/dev/null || echo "N/A")
        K8S_VERSION=$(kubectl version --short 2>/dev/null | grep "Server Version" | awk '{print $3}' || echo "N/A")
        
        echo "   📊 Nodes: $NODE_COUNT"
        echo "   🎯 Contexto: $CURRENT_CONTEXT"
        echo "   🐳 Versão K8s: $K8S_VERSION"
    else
        echo "   ❌ Ainda sem conexão"
        echo "   💡 Dicas de troubleshooting:"
        echo "      - Verifique se o cluster está rodando: kubectl cluster-info"
        echo "      - Confirme o kubeconfig montado: docker exec nautikube cat /root/.kube/config"
        echo "      - Para clusters EKS: verifique ~/.aws/credentials"
        echo "      - Servidor detectado: $(grep -m 1 'server:' /root/.kube/config_mod | awk '{print $2}')"
    fi
fi

# Verifica conectividade com Ollama
echo ""
echo "🤖 Verificando Ollama..."
if curl -s http://host.docker.internal:11434/api/tags > /dev/null 2>&1; then
    echo "✅ Ollama acessível em http://host.docker.internal:11434"
    
    # Lista modelos instalados
    MODEL_COUNT=$(curl -s http://host.docker.internal:11434/api/tags 2>/dev/null | grep -o '"name"' | wc -l || echo "0")
    if [ "$MODEL_COUNT" -gt 0 ]; then
        echo "   $MODEL_COUNT modelo(s) instalado(s)"
    fi
else
    echo "⚠️  Ollama não acessível"
    echo "   Use 'analyze' sem --explain para análise básica"
fi

echo ""
echo "🚀 NautiKube v2.0.3 pronto!"
echo "   Uso: docker exec nautikube nautikube analyze --explain"
echo ""

# Mantém container rodando
tail -f /dev/null
