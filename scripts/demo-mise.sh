#!/bin/bash
# Demo script showing mise capabilities

set -e

echo "🚀 Homelab Assistant - Mise Demo"
echo "================================="

# Check if mise is available
if ! command -v mise &> /dev/null; then
    echo "❌ mise is not installed. Please install it first:"
    echo "   curl https://mise.run | sh"
    exit 1
fi

echo "✅ mise is available"

# Show mise version
echo "📦 Mise version: $(mise --version)"

# Install tools
echo ""
echo "🔧 Installing project tools..."
mise install

# Show installed tools
echo ""
echo "📋 Installed tools:"
mise list

# Run development tasks
echo ""
echo "🧪 Running development tasks..."

echo "  → Formatting code..."
mise run fmt

echo "  → Tidying modules..."
mise run tidy

echo "  → Running linter..."
if mise run lint; then
    echo "  ✅ Linting passed"
else
    echo "  ❌ Linting failed"
fi

echo "  → Running vet..."
if mise run vet; then
    echo "  ✅ Vet passed"
else
    echo "  ❌ Vet failed"
fi

echo "  → Building project..."
if mise run build; then
    echo "  ✅ Build passed"
else
    echo "  ❌ Build failed"
fi

echo "  → Running unit tests..."
if mise run test-unit; then
    echo "  ✅ Unit tests passed"
else
    echo "  ❌ Unit tests failed"
fi

echo ""
echo "🎉 Demo completed!"
echo ""
echo "💡 Try these commands:"
echo "   mise run ci          # Full CI pipeline"
echo "   mise run dev-setup   # Development setup"
echo "   mise run lint        # Run linting"
echo "   mise run test        # Run tests"
echo "   mise run build       # Build project"
echo ""
echo "🐳 For local Kubernetes testing:"
echo "   mise run k8s-setup   # Create cluster + install CRDs"
echo "   mise run k8s-teardown # Clean up everything"
echo "   mise run kind-create # Just create cluster"
echo "   mise run kind-delete # Just delete cluster"
echo ""
echo "🔧 Code generation:"
echo "   mise run generate    # Generate deepcopy methods"
echo "   mise run manifests   # Generate CRDs and RBAC"
