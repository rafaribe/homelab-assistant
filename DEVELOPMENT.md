# Development Guide

This guide covers development workflows for the Homelab Assistant project.

## 🛠️ Tool Management with Mise

We use [mise](https://mise.jdx.dev/) for managing development tools and automating common tasks.

### Installation

```bash
# Install mise
curl https://mise.run | sh

# Or via package manager
# macOS
brew install mise

# Ubuntu/Debian
sudo install -dm 755 /etc/apt/keyrings
wget -qO - https://mise.jdx.dev/gpg-key.pub | gpg --dearmor | sudo tee /etc/apt/keyrings/mise-archive-keyring.gpg 1> /dev/null
echo "deb [signed-by=/etc/apt/keyrings/mise-archive-keyring.gpg arch=amd64] https://mise.jdx.dev/deb stable main" | sudo tee /etc/apt/sources.list.d/mise.list
sudo apt update
sudo apt install mise
```

### Quick Start

```bash
# Install all project tools
mise install

# Run the full CI pipeline
mise run ci

# Set up development environment
mise run dev-setup
```

## 📋 Available Tasks

### Core Development Tasks

```bash
# Code quality
mise run lint          # Run golangci-lint
mise run vet           # Run go vet  
mise run fmt           # Format Go code
mise run tidy          # Tidy Go modules

# Building and testing
mise run build         # Build the project
mise run test          # Run all tests
mise run test-unit     # Run unit tests only

# Code generation
mise run generate      # Generate deepcopy methods
mise run manifests     # Generate CRDs and RBAC
```

### Kubernetes Development

```bash
# Local cluster management
mise run kind-create   # Create kind cluster
mise run kind-delete   # Delete kind cluster
mise run install-crds  # Install CRDs to cluster

# Container operations
mise run docker-build  # Build Docker image
```

### Composite Tasks

```bash
# Development workflows
mise run dev-setup     # Complete development setup
mise run ci            # Full CI pipeline (lint, vet, build, test)
mise run pre-commit    # Pre-commit checks
```

## 🔧 Configuration Files

### `.mise.toml`
Main configuration file with tools, environment variables, and tasks.

### `.tool-versions`
Compatible with asdf for tool version management.

### `mise.toml`
Alternative configuration format (simpler).

## 🚀 Development Workflow

### 1. Initial Setup
```bash
# Clone and setup
git clone https://github.com/rafaribe/homelab-assistant.git
cd homelab-assistant
mise install
mise run dev-setup
```

### 2. Daily Development
```bash
# Before starting work
mise run pre-commit

# During development
mise run lint          # Check code quality
mise run test          # Run tests
mise run build         # Build project

# Before committing
mise run ci            # Full CI pipeline
```

### 3. Testing Changes
```bash
# Unit tests
mise run test-unit

# Integration tests (requires cluster)
mise run kind-create
mise run install-crds
# Run your integration tests
mise run kind-delete
```

## 🔄 Integration with Make

The project supports both mise and traditional Make workflows:

```bash
# These are equivalent
make lint              # Traditional
mise run lint          # Mise

# Make can use mise if available
make mise-ci           # Uses mise for CI
make mise-dev-setup    # Uses mise for setup
```

## 🐳 Docker Development

```bash
# Build image
mise run docker-build

# Test locally
docker run --rm homelab-assistant:latest --help
```

## 🧪 Testing

### Unit Tests
```bash
# All unit tests
mise run test-unit

# Specific package
mise exec -- go test ./api/v1alpha1 -v

# With coverage
mise exec -- go test ./... -coverprofile=coverage.out
```

### Integration Tests
```bash
# Create test cluster
mise run kind-create

# Install CRDs
mise run install-crds

# Run controller tests
mise exec -- go test ./internal/controller -v

# Cleanup
mise run kind-delete
```

## 🔍 Debugging

### Tool Issues
```bash
# Check tool versions
mise list

# Reinstall tools
mise install --force

# Check mise configuration
mise config
```

### Build Issues
```bash
# Clean and rebuild
mise run clean
mise run build

# Check Go environment
mise exec -- go env
```

## 📚 Additional Resources

- [Mise Documentation](https://mise.jdx.dev/)
- [Kubebuilder Book](https://book.kubebuilder.io/)
- [Controller Runtime](https://pkg.go.dev/sigs.k8s.io/controller-runtime)
- [Kubernetes API Reference](https://kubernetes.io/docs/reference/kubernetes-api/)

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Run `mise run pre-commit` before committing
4. Submit a pull request

The CI pipeline will automatically run `mise run ci` to validate your changes.
