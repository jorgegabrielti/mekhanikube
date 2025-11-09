#!/bin/bash

# Healthcheck script for Mekhanikube
set -e

echo "🏥 Mekhanikube Health Check"
echo "=========================="
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

check_status=0

# Check Docker daemon
echo -n "🐳 Docker daemon: "
if docker info > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Running${NC}"
else
    echo -e "${RED}✗ Not running${NC}"
    check_status=1
fi

# Check Ollama container
echo -n "🤖 Ollama container: "
if docker ps | grep -q mekhanikube-ollama; then
    echo -e "${GREEN}✓ Running${NC}"
    
    # Check Ollama API
    echo -n "   └─ Ollama API: "
    if docker exec mekhanikube-ollama curl -f http://localhost:11434/api/tags > /dev/null 2>&1; then
        echo -e "${GREEN}✓ Healthy${NC}"
    else
        echo -e "${YELLOW}⚠ Unhealthy${NC}"
        check_status=1
    fi
    
    # List installed models
    echo -n "   └─ Installed models: "
    models=$(docker exec mekhanikube-ollama ollama list 2>/dev/null | tail -n +2 | wc -l)
    if [ "$models" -gt 0 ]; then
        echo -e "${GREEN}$models model(s)${NC}"
        docker exec mekhanikube-ollama ollama list 2>/dev/null | tail -n +2 | while read line; do
            echo "      • $line"
        done
    else
        echo -e "${YELLOW}⚠ No models installed${NC}"
        echo "      Run: make install-model"
    fi
else
    echo -e "${RED}✗ Not running${NC}"
    check_status=1
fi

echo ""

# Check K8sGPT container
echo -n "🔧 K8sGPT container: "
if docker ps | grep -q mekhanikube-k8sgpt; then
    echo -e "${GREEN}✓ Running${NC}"
    
    # Check K8sGPT version
    echo -n "   └─ K8sGPT version: "
    version=$(docker exec mekhanikube-k8sgpt k8sgpt version 2>/dev/null | head -n1)
    echo -e "${GREEN}$version${NC}"
    
    # Check K8sGPT auth
    echo -n "   └─ K8sGPT backend: "
    if docker exec mekhanikube-k8sgpt k8sgpt auth list 2>/dev/null | grep -q "Active.*true"; then
        backend=$(docker exec mekhanikube-k8sgpt k8sgpt auth list 2>/dev/null | grep "Active.*true" -B3 | grep "Provider:" | awk '{print $2}')
        echo -e "${GREEN}✓ $backend${NC}"
    else
        echo -e "${YELLOW}⚠ Not configured${NC}"
        check_status=1
    fi
    
    # Check Kubernetes connectivity
    echo -n "   └─ Kubernetes API: "
    if docker exec mekhanikube-k8sgpt kubectl cluster-info > /dev/null 2>&1; then
        echo -e "${GREEN}✓ Connected${NC}"
    else
        echo -e "${RED}✗ Not accessible${NC}"
        check_status=1
    fi
else
    echo -e "${RED}✗ Not running${NC}"
    check_status=1
fi

echo ""

# Check Docker volumes
echo "💾 Docker volumes:"
for volume in mekhanikube-ollama-data mekhanikube-k8sgpt-config; do
    echo -n "   └─ $volume: "
    if docker volume ls | grep -q $volume; then
        size=$(docker system df -v 2>/dev/null | grep $volume | awk '{print $3}')
        echo -e "${GREEN}✓ ${size}${NC}"
    else
        echo -e "${YELLOW}⚠ Not found${NC}"
    fi
done

echo ""
echo "=========================="

if [ $check_status -eq 0 ]; then
    echo -e "${GREEN}✓ All systems operational!${NC}"
    exit 0
else
    echo -e "${YELLOW}⚠ Some issues detected${NC}"
    exit 1
fi
