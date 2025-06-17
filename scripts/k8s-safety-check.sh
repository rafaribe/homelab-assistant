#!/bin/bash
# Safety check script to ensure we only use kind clusters

set -e

CURRENT_CONTEXT=$(kubectl config current-context 2>/dev/null || echo "none")

echo "🔒 Kubernetes Safety Check"
echo "=========================="
echo "Current context: $CURRENT_CONTEXT"

if [[ "$CURRENT_CONTEXT" == "none" ]]; then
    echo "❌ No Kubernetes context set"
    echo "💡 Run 'mise run kind-create' to create a test cluster"
    exit 1
fi

if [[ "$CURRENT_CONTEXT" =~ ^kind- ]]; then
    echo "✅ Safe: Using kind cluster ($CURRENT_CONTEXT)"
    kubectl cluster-info --context "$CURRENT_CONTEXT"
    exit 0
else
    echo "🚨 DANGER: Not using a kind cluster!"
    echo "Current context: $CURRENT_CONTEXT"
    echo ""
    echo "For safety, this project only works with kind clusters."
    echo "Switch to a kind cluster or create one:"
    echo "  mise run kind-create"
    echo "  kubectl config use-context kind-homelab-assistant"
    exit 1
fi
