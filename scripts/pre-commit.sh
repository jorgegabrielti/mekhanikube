#!/bin/sh
# NautiKube pre-commit hook
# This script runs automatically before every git commit to ensure code quality.

echo "Running NautiKube pre-commit checks..."

# Check if we are on Windows or Unix
if [ "$OS" = "Windows_NT" ]; then
    cmd.exe /c ".\scripts\check.bat"
else
    ./scripts/check.sh
fi

if [ $? -ne 0 ]; then
    echo ""
    echo "[!] PRE-COMMIT FAILED: Please fix the errors above before committing."
    echo "[!] To bypass this check (not recommended), use: git commit --no-verify"
    exit 1
fi

echo "Pre-commit checks passed! Committing..."
exit 0
