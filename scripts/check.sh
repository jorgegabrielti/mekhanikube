#!/usr/bin/env bash
set -e

echo "==================================="
echo "  NautiKube Local Quality Gate"
echo "==================================="

echo "[1/4] Running go fmt..."
go fmt ./...

echo "[2/4] Running go vet..."
go vet ./...

echo "[3/4] Running golangci-lint..."
if ! command -v golangci-lint &> /dev/null; then
    echo "[WARNING] golangci-lint not installed locally. Skipping."
    echo "To install: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b \$(go env GOPATH)/bin"
else
    golangci-lint run
fi

echo "[4/4] Running govulncheck..."
if ! command -v govulncheck &> /dev/null; then
    echo "[WARNING] govulncheck not installed locally. Skipping."
    echo "To install: go install golang.org/x/vuln/cmd/govulncheck@latest"
else
    govulncheck ./...
fi

echo "[5/5] Running Tests..."
go test ./...

echo "==================================="
echo "  ALL CHECKS PASSED SUCCESSFULLY!"
echo "==================================="
