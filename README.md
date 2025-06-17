# Homelab Assistant

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

- [Deployment Guide](DEPLOYMENT.md) - Complete deployment instructions
- [VolSync Monitor](VOLSYNC_MONITOR.md) - Technical documentation
- [Examples](examples/) - Real-world deployment examples

## 🛠️ **Development**

```bash
# Clone the repository
git clone https://github.com/rafaribe/homelab-assistant.git
cd homelab-assistant

# Run tests
make test

# Build and run locally
make run
```

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
