#!/bin/sh

# Ajusta o kubeconfig substituindo 127.0.0.1 por host.docker.internal
if [ -f /root/.kube/config ]; then
    sed 's/127\.0\.0\.1/host.docker.internal/g' /root/.kube/config > /root/.kube/config_mod
    export KUBECONFIG=/root/.kube/config_mod
fi

# Configura K8sGPT com Ollama
if ! k8sgpt auth list | grep -A1 "Active:" | grep -q "ollama"; then
    echo "Configurando K8sGPT com Ollama..."
    
    # Remove configurações anteriores se existirem
    k8sgpt auth remove --backend localai 2>/dev/null || true
    k8sgpt auth remove --backend ollama 2>/dev/null || true
    
    # Adiciona backend Ollama e configura como padrão
    k8sgpt auth add --backend ollama --model gemma:7b --baseurl http://localhost:11434
    k8sgpt auth default -p ollama
    echo "K8sGPT configurado!"
else
    echo "K8sGPT já está configurado com ollama ativo"
fi

# Mantém o container rodando
while true; do
    sleep 3600
done
