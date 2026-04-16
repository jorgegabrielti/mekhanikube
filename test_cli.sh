#!/usr/bin/env bash

# Build binary
go build -o nautikube.exe ./cmd/nautikube

echo "=== Test 1: Invalid Namespace ==="
./nautikube.exe scan --namespace this-does-not-exist

echo -e "\n=== Test 2: Filter by Resource (Deployment) ==="
./nautikube.exe scan --namespace nautikube-qa --resource Deployment

echo -e "\n=== Test 3: Filter by Min-Severity (critical) ==="
./nautikube.exe scan --namespace nautikube-qa --min-severity critical

echo -e "\n=== Test 4: Output JSON ==="
./nautikube.exe scan --namespace nautikube-qa --output json | head -n 20

echo -e "\n=== Test 5: Output YAML ==="
./nautikube.exe scan --namespace nautikube-qa --output yaml | head -n 20

echo -e "\n=== Test 6: No Color ==="
./nautikube.exe scan --namespace nautikube-qa --no-color | head -n 10

echo -e "\n=== Test 7: Invalid Kubeconfig ==="
./nautikube.exe scan --kubeconfig /path/to/nowhere

echo -e "\n=== Test 8: Invalid Context ==="
./nautikube.exe scan --context non-existent-context
