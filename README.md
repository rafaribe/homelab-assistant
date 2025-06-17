# Homelab Assistant

[![CI](https://github.com/rafaribe/homelab-assistant/workflows/CI/badge.svg)](https://github.com/rafaribe/homelab-assistant/actions)
[![Built with Mise](https://img.shields.io/badge/built%20with-mise-blue)](https://mise.jdx.dev/)

A collection of Kubernetes controllers designed to automate and simplify homelab operations.

## 🎯 **Controllers**

### VolSync Monitor Controller
Automatically detects and resolves restic repository lock issues in VolSync backup jobs.

**Features:**
- 🔍 **Smart Detection**: Monitors VolSync jobs for lock errors using configurable regex patterns
- 🔓 **Auto-Unlock**: Creates unlock jobs when lock errors are detected
- 📊 **Prometheus Metrics**: Comprehensive metrics for monitoring and alerting
- ⚙️ **Configurable**: Customizable unlock job templates and error patterns
- 🛡️ **Secure**: Follows Kubernetes security best practices

### Future Controllers
- **Backup Controller**: Automated backup management and scheduling
- **Network Controller**: Dynamic network configuration and monitoring
- **Storage Controller**: Intelligent storage provisioning and cleanup

## 🚀 **Quick Start**

### Flux GitOps (Recommended)

```bash
# Add to your Flux repository
kubectl apply -f examples/flux-crds-deployment.yaml
kubectl apply -f examples/flux-app-template-deployment.yaml
```

### Helm Installation (GHCR OCI Registry)

```bash
# Install CRDs
helm install homelab-assistant-crds \
  oci://ghcr.io/rafaribe/homelab-assistant-crds \
  --version 0.1.0 \
  --namespace homelab-assistant-system --create-namespace

# Install Controllers
helm install homelab-assistant \
  oci://ghcr.io/rafaribe/homelab-assistant \
  --version 0.1.0 \
  --namespace homelab-assistant-system
```

### Traditional Helm Repository

```bash
# Install CRDs
helm repo add homelab-assistant https://rafaribe.github.io/homelab-assistant
helm install homelab-assistant-crds homelab-assistant/homelab-assistant-crds \
  --namespace homelab-assistant-system --create-namespace

# Install Controllers
helm install homelab-assistant homelab-assistant/homelab-assistant \
  --namespace homelab-assistant-system
```

## 📊 **Monitoring**

The controllers expose Prometheus metrics for comprehensive monitoring:

- `volsync_unlock_jobs_created_total` - Total unlock jobs created
- `volsync_unlock_jobs_succeeded_total` - Successful unlock jobs
- `volsync_unlock_jobs_failed_total` - Failed unlock jobs
- `volsync_active_unlock_jobs` - Currently active unlock jobs
- `volsync_lock_errors_detected_total` - Lock errors detected

## 🏠 **Perfect for Homelabs**

- **GitOps Ready**: Full Flux and ArgoCD support
- **App-Template Compatible**: Works with bjw-s app-template
- **Lightweight**: Minimal resource requirements
- **Secure**: Non-root containers, RBAC, security contexts
- **Observable**: Rich metrics and logging

## 📚 **Documentation**

- [Schema Documentation](https://rafaribe.github.io/homelab-assistant/) - CRD schemas and installation
- [Deployment Guide](DEPLOYMENT.md) - Complete deployment instructions
- [VolSync Monitor](VOLSYNC_MONITOR.md) - Technical documentation
- [Examples](examples/) - Real-world deployment examples

## 🛠️ **Development**

### Using Mise (Recommended)

This project uses [mise](https://mise.jdx.dev/) for tool management and task automation everywhere - local development, CI, and production builds:

```bash
# Install mise (if not already installed)
curl https://mise.run | sh

# Install all project tools
mise install

# Run common development tasks
mise run fmt           # Format code
mise run tidy          # Tidy dependencies  
mise run lint          # Run linting
mise run vet           # Run go vet
mise run build         # Build project

# Testing
mise run test-unit-only    # Fast unit tests only
mise run test-controller   # Controller tests (requires envtest)
mise run test-all          # All tests
mise run coverage          # Generate coverage report

# CI pipelines
mise run ci-fast       # Fast CI (unit tests only) - 1.15s
mise run ci            # Full CI (all tests) - comprehensive
mise run dev-setup     # Set up development environment

# Kubernetes development
mise run k8s-setup     # Create kind cluster + install CRDs
mise run k8s-teardown  # Clean up kind cluster
mise run generate      # Generate deepcopy methods
mise run manifests     # Generate CRDs and RBAC

# Schema documentation
mise run generate-schemas  # Generate CRD schemas
mise run validate-schemas  # Validate schemas
mise run preview-schemas   # Preview schemas locally
mise run serve-schemas     # Serve docs at http://localhost:8000
```

### Traditional Make (Fallback)

The Makefile provides fallbacks and integrates with mise when available:

```bash
# Clone the repository
git clone https://github.com/rafaribe/homelab-assistant.git
cd homelab-assistant

# These automatically use mise if available
make lint              # Linting
make test              # Testing
make build             # Building
make docker-build      # Docker image

# Mise-specific targets
make mise-ci           # Full CI via mise
make mise-k8s-setup    # Kubernetes setup via mise
make mise-k8s-teardown # Kubernetes cleanup via mise
```

### Available Tools

The project automatically manages these tools via mise:
- **Go 1.22** - Programming language
- **golangci-lint** - Code linting
- **controller-gen** - Kubernetes code generation
- **kubectl** - Kubernetes CLI (with kind-only safeguards)
- **kind** - Local Kubernetes clusters for testing
- **kustomize** - Configuration management

**Safety Features**: All kubectl operations are restricted to kind clusters only.

## 📈 **Real-World Example**

When your Prowlarr VolSync backup fails with "repository is already locked":

1. **Detection**: Controller detects the lock error in pod logs
2. **Metrics**: Records `volsync_lock_errors_detected_total`
3. **Action**: Creates unlock job `volsync-unlock-prowlarr-prowlarr-nfs-123456`
4. **Monitoring**: Updates VolSyncMonitor status with active unlock info
5. **Completion**: Records success metrics and cleans up

## 🤝 **Contributing**

Contributions are welcome! Please see our [Contributing Guide](CONTRIBUTING.md).

## 📄 **License**

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## 🔧 **CI/CD**

The project uses GitHub Actions with mise for consistent tooling:

- **Main CI** (`.github/workflows/ci.yml`) - Runs on every push/PR
- **Schema Publishing** (`.github/workflows/schemas.yml`) - Publishes CRD schemas to GitHub Pages

**Troubleshooting**: If you encounter issues with the `jdx/mise-action`, you can install mise manually in CI:
```yaml
- name: Install mise manually
  run: |
    curl https://mise.run | sh
    echo "$HOME/.local/bin" >> $GITHUB_PATH
```
