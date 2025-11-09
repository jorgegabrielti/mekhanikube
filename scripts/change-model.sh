#!/bin/bash

# Script to change AI model
set -e

MODEL=${1:-"gemma:7b"}

echo "🔄 Changing AI model to: $MODEL"
echo ""

# Check if model is installed
echo "Checking if model is installed..."
if docker exec mekhanikube-ollama ollama list | grep -q "$MODEL"; then
    echo "✓ Model $MODEL is already installed"
else
    echo "⚠ Model $MODEL not found. Downloading..."
    docker exec mekhanikube-ollama ollama pull $MODEL
    echo "✓ Model downloaded successfully"
fi

echo ""
echo "Reconfiguring K8sGPT..."

# Remove old backend
docker exec mekhanikube-k8sgpt k8sgpt auth remove --backend ollama 2>/dev/null || true

# Add new backend with specified model
docker exec mekhanikube-k8sgpt k8sgpt auth add --backend ollama --model $MODEL --baseurl http://localhost:11434

# Set as default
docker exec mekhanikube-k8sgpt k8sgpt auth default -p ollama

echo ""
echo "✓ Model changed successfully to: $MODEL"
echo ""
echo "Current configuration:"
docker exec mekhanikube-k8sgpt k8sgpt auth list
