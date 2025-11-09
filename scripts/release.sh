#!/bin/bash

# Release script for Mekhanikube
set -e

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Check if version is provided
if [ -z "$1" ]; then
    echo -e "${RED}Error: Version number required${NC}"
    echo "Usage: $0 <version> (e.g., 1.0.1)"
    exit 1
fi

NEW_VERSION=$1
CURRENT_VERSION=$(cat VERSION)

echo -e "${BLUE}Mekhanikube Release Script${NC}"
echo "=========================="
echo ""
echo "Current version: $CURRENT_VERSION"
echo "New version: $NEW_VERSION"
echo ""

# Validate version format (semver)
if ! [[ $NEW_VERSION =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo -e "${RED}Error: Invalid version format. Use semantic versioning (e.g., 1.0.0)${NC}"
    exit 1
fi

# Check if git is clean
if [ -n "$(git status --porcelain)" ]; then
    echo -e "${YELLOW}Warning: You have uncommitted changes${NC}"
    read -p "Continue anyway? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

echo -e "${GREEN}Updating version files...${NC}"

# Update VERSION file
echo "$NEW_VERSION" > VERSION

# Update CHANGELOG.md
TODAY=$(date +%Y-%m-%d)
sed -i.bak "s/## \[Unreleased\]/## [Unreleased]\n\n## [$NEW_VERSION] - $TODAY/" CHANGELOG.md
rm CHANGELOG.md.bak 2>/dev/null || true

echo -e "${GREEN}Building and testing...${NC}"

# Run tests
make test || {
    echo -e "${RED}Tests failed! Fix issues before releasing.${NC}"
    exit 1
}

# Build images
make build || {
    echo -e "${RED}Build failed!${NC}"
    exit 1
}

echo -e "${GREEN}Committing changes...${NC}"

# Git operations
git add VERSION CHANGELOG.md
git commit -m "Release v$NEW_VERSION"
git tag -a "v$NEW_VERSION" -m "Release version $NEW_VERSION"

echo ""
echo -e "${GREEN}✓ Release v$NEW_VERSION prepared!${NC}"
echo ""
echo "Next steps:"
echo "  1. Review the changes: git log -1 -p"
echo "  2. Push the release: git push && git push --tags"
echo "  3. Create GitHub release: https://github.com/jorgegabrielti/mekhanikube/releases/new"
echo ""
