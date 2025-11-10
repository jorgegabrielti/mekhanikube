#!/bin/sh

# Ajusta o kubeconfig substituindo 127.0.0.1 por host.docker.internal
if [ -f /root/.kube/config ]; then
    sed 's/127\.0\.0\.1/host.docker.internal/g' /root/.kube/config > /root/.kube/config_mod
    export KUBECONFIG=/root/.kube/config_mod
fi

# Aguardar Ollama estar disponível
echo "Aguardando Ollama estar disponível..."
until curl -f ${OLLAMA_URL:-http://host.docker.internal:11434}/api/tags >/dev/null 2>&1; do
    echo "Ollama não está pronto ainda. Tentando novamente em 5 segundos..."
    sleep 5
done
echo "Ollama está pronto!"

# Configura K8sGPT com Ollama se ainda não estiver configurado
if ! k8sgpt auth list | grep -A1 "Active:" | grep -q "ollama"; then
    echo "Configurando K8sGPT com Ollama para MCP..."
    
    # Remove configurações anteriores se existirem
    k8sgpt auth remove --backends localai 2>/dev/null || true
    k8sgpt auth remove --backends ollama 2>/dev/null || true
    
    # Adiciona backend Ollama e configura como padrão
    k8sgpt auth add --backend ollama --model ${K8SGPT_MODEL:-llama3.1:8b} --baseurl ${OLLAMA_URL:-http://host.docker.internal:11434}
    k8sgpt auth default -p ollama
    echo "K8sGPT configurado para MCP!"
else
    echo "K8sGPT já está configurado com ollama ativo"
fi

# Inicia o servidor MCP
echo "Iniciando servidor MCP na porta 3000..."
exec k8sgpt serve --mcp --port 3000
