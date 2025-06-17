# Homelab Assistant Helm Chart

A Helm chart for deploying the Homelab Assistant controllers on Kubernetes.

## Prerequisites

- Kubernetes 1.25+
- Helm 3.8+
- VolSync operator (for VolSync Monitor controller)

## Installation

### Install CRDs First

```bash
helm install homelab-assistant-crds \
  oci://ghcr.io/rafaribe/homelab-assistant-crds \
  --version 0.1.0 \
  --namespace homelab-assistant-system \
  --create-namespace
```

### Install Controllers

```bash
helm install homelab-assistant \
  oci://ghcr.io/rafaribe/homelab-assistant \
  --version 0.1.0 \
  --namespace homelab-assistant-system
```

## Configuration

The following table lists the configurable parameters and their default values.

### Controller Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `controller.image.repository` | Controller image repository | `ghcr.io/rafaribe/homelab-assistant` |
| `controller.image.tag` | Controller image tag | `latest` |
| `controller.image.pullPolicy` | Image pull policy | `IfNotPresent` |
| `controller.resources.limits.cpu` | CPU limit | `500m` |
| `controller.resources.limits.memory` | Memory limit | `128Mi` |
| `controller.resources.requests.cpu` | CPU request | `10m` |
| `controller.resources.requests.memory` | Memory request | `64Mi` |

### VolSync Monitor Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `volsyncMonitor.enabled` | Enable VolSync Monitor controller | `true` |
| `volsyncMonitor.maxConcurrentUnlocks` | Maximum concurrent unlock operations | `3` |
| `volsyncMonitor.ttlSecondsAfterFinished` | TTL for unlock jobs | `3600` |
| `volsyncMonitor.lockErrorPatterns` | Custom lock error patterns | `[]` |
| `volsyncMonitor.unlockJob.image.repository` | Unlock job image repository | `quay.io/backube/volsync` |
| `volsyncMonitor.unlockJob.image.tag` | Unlock job image tag | `0.13.0-rc.2` |

### RBAC Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `rbac.create` | Create RBAC resources | `true` |
| `serviceAccount.create` | Create service account | `true` |
| `serviceAccount.name` | Service account name | `""` |

### Metrics Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `metrics.enabled` | Enable metrics endpoint | `true` |
| `metrics.port` | Metrics port | `8080` |
| `metrics.serviceMonitor.enabled` | Create ServiceMonitor | `false` |

## Examples

### Basic Installation

```yaml
# values.yaml
controller:
  image:
    tag: "v0.1.0"

volsyncMonitor:
  enabled: true
  maxConcurrentUnlocks: 5
  
metrics:
  enabled: true
  serviceMonitor:
    enabled: true
    additionalLabels:
      release: prometheus
```

### Minimal Installation

```yaml
# values.yaml
volsyncMonitor:
  enabled: false

metrics:
  enabled: false
```

### Custom Lock Error Patterns

```yaml
# values.yaml
volsyncMonitor:
  enabled: true
  lockErrorPatterns:
    - "repository is already locked"
    - "unable to create lock"
    - "custom error pattern"
```

## Uninstalling

```bash
# Remove the controllers
helm uninstall homelab-assistant -n homelab-assistant-system

# Remove the CRDs (optional, will remove all custom resources)
helm uninstall homelab-assistant-crds -n homelab-assistant-system
```

## Development

### Running Tests

```bash
# Install helm unittest plugin
helm plugin install https://github.com/helm-unittest/helm-unittest.git

# Run tests
helm unittest charts/homelab-assistant
```

### Linting

```bash
# Install chart-testing
pip install chart-testing

# Lint charts
ct lint --config .github/ct.yaml
```
