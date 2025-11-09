#!/bin/bash

# Integration tests for Mekhanikube
set -e

echo "🧪 Mekhanikube Integration Tests"
echo "================================"
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

test_count=0
passed_count=0
failed_count=0

run_test() {
    local test_name=$1
    local test_command=$2
    
    test_count=$((test_count + 1))
    echo -n "Test $test_count: $test_name... "
    
    if eval "$test_command" > /dev/null 2>&1; then
        echo -e "${GREEN}✓ PASSED${NC}"
        passed_count=$((passed_count + 1))
        return 0
    else
        echo -e "${RED}✗ FAILED${NC}"
        failed_count=$((failed_count + 1))
        return 1
    fi
}

# Test 1: Docker Compose configuration is valid
run_test "Docker Compose config validation" \
    "docker-compose config -q"

# Test 2: Ollama container is running
run_test "Ollama container is running" \
    "docker ps | grep -q mekhanikube-ollama"

# Test 3: K8sGPT container is running
run_test "K8sGPT container is running" \
    "docker ps | grep -q mekhanikube-k8sgpt"

# Test 4: Ollama API is accessible
run_test "Ollama API responds" \
    "docker exec mekhanikube-ollama curl -f http://localhost:11434/api/tags"

# Test 5: K8sGPT binary is functional
run_test "K8sGPT version command" \
    "docker exec mekhanikube-k8sgpt k8sgpt version"

# Test 6: K8sGPT auth is configured
run_test "K8sGPT auth configuration" \
    "docker exec mekhanikube-k8sgpt k8sgpt auth list | grep -q 'Active.*true'"

# Test 7: K8sGPT can list filters
run_test "K8sGPT filters list" \
    "docker exec mekhanikube-k8sgpt k8sgpt filters list"

# Test 8: Kubernetes connection works
run_test "Kubernetes API connectivity" \
    "docker exec mekhanikube-k8sgpt kubectl cluster-info"

# Test 9: K8sGPT analyze (without --explain) works
run_test "K8sGPT basic analysis" \
    "docker exec mekhanikube-k8sgpt k8sgpt analyze"

# Test 10: Volumes exist
run_test "Ollama data volume exists" \
    "docker volume ls | grep -q mekhanikube-ollama-data"

run_test "K8sGPT config volume exists" \
    "docker volume ls | grep -q mekhanikube-k8sgpt-config"

# Test 11: Entrypoint script exists and is executable
run_test "Entrypoint script exists" \
    "docker exec mekhanikube-k8sgpt test -x /entrypoint.sh"

echo ""
echo "================================"
echo "Test Results:"
echo "  Total: $test_count"
echo -e "  ${GREEN}Passed: $passed_count${NC}"
echo -e "  ${RED}Failed: $failed_count${NC}"
echo ""

if [ $failed_count -eq 0 ]; then
    echo -e "${GREEN}✓ All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}✗ Some tests failed${NC}"
    exit 1
fi
