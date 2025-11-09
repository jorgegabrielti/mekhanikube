#!/bin/bash

# Quick analysis script with formatted output
set -e

NAMESPACE=${1:-""}
FILTER=${2:-""}

echo "🔍 Mekhanikube - Kubernetes Analysis"
echo "===================================="
echo ""

# Build command
CMD="k8sgpt analyze --explain"

if [ -n "$NAMESPACE" ]; then
    CMD="$CMD -n $NAMESPACE"
    echo "📦 Namespace: $NAMESPACE"
fi

if [ -n "$FILTER" ]; then
    CMD="$CMD --filter=$FILTER"
    echo "🔎 Filter: $FILTER"
fi

echo ""
echo "Running analysis..."
echo ""

# Execute analysis
docker exec mekhanikube-k8sgpt $CMD

echo ""
echo "===================================="
echo "✓ Analysis complete"
