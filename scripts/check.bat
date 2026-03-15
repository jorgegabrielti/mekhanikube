@echo off
setlocal enabledelayedexpansion

echo ===================================
echo   NautiKube Local Quality Gate
echo ===================================

echo [1/4] Running go fmt...
go fmt ./...
if %errorlevel% neq 0 (
    echo [ERROR] go fmt failed.
    exit /b %errorlevel%
)

echo [2/4] Running go vet...
go vet ./...
if %errorlevel% neq 0 (
    echo [ERROR] go vet failed.
    exit /b %errorlevel%
)

echo [3/4] Running golangci-lint...
where golangci-lint >nul 2>nul
if %errorlevel% neq 0 (
    echo [WARNING] golangci-lint not installed locally. Skipping.
    echo To install: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
) else (
    golangci-lint run
    if !errorlevel! neq 0 (
        echo [ERROR] golangci-lint failed.
        exit /b !errorlevel!
    )
)

echo [4/4] Running govulncheck...
where govulncheck >nul 2>nul
if %errorlevel% neq 0 (
    echo [WARNING] govulncheck not installed locally. Skipping.
    echo To install: go install golang.org/x/vuln/cmd/govulncheck@latest
) else (
    govulncheck ./...
    if !errorlevel! neq 0 (
        echo [ERROR] govulncheck failed.
        exit /b !errorlevel!
    )
)

echo [5/5] Running Tests...
go test ./...
if %errorlevel% neq 0 (
    echo [ERROR] Tests failed.
    exit /b %errorlevel%
)

echo ===================================
echo   ALL CHECKS PASSED SUCCESSFULLY!
echo ===================================
exit /b 0
