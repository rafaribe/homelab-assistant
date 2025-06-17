# Homelab Assistant CRDs Helm Chart

Custom Resource Definitions for the Homelab Assistant controllers.

## Prerequisites

- Kubernetes 1.25+
- Helm 3.8+

## Installation

```bash
helm install homelab-assistant-crds \
  oci://ghcr.io/rafaribe/homelab-assistant-crds \
  --version 0.1.0 \
  --namespace homelab-assistant-system \
  --create-namespace
```

## Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `installCRDs` | Install CRDs | `true` |

## Custom Resources

This chart installs the following Custom Resource Definitions:

### VolSyncMonitor

Configures the VolSync Monitor controller to automatically detect and resolve restic repository lock issues.

```yaml
apiVersion: homelab.rafaribe.com/v1alpha1
kind: VolSyncMonitor
metadata:
  name: main-monitor
spec:
  enabled: true
  maxConcurrentUnlocks: 3
  ttlSecondsAfterFinished: 3600
  lockErrorPatterns:
    - "repository is already locked"
    - "unable to create lock"
  unlockJobTemplate:
    image: "quay.io/backube/volsync:0.13.0-rc.2"
    command: ["restic"]
    args: ["unlock", "--remove-all"]
```

### VolSyncUnlock

Represents an active unlock operation for a specific VolSync backup.

```yaml
apiVersion: homelab.rafaribe.com/v1alpha1
kind: VolSyncUnlock
metadata:
  name: unlock-prowlarr-backup
spec:
  appName: prowlarr
  namespace: downloads
  objectName: prowlarr-nfs
  jobTemplate:
    image: "quay.io/backube/volsync:0.13.0-rc.2"
    command: ["restic"]
    args: ["unlock", "--remove-all"]
```

## Uninstalling

⚠️ **Warning**: Uninstalling this chart will remove all CRDs and their associated custom resources.

```bash
helm uninstall homelab-assistant-crds -n homelab-assistant-system
```

## Development

### Running Tests

```bash
# Install helm unittest plugin
helm plugin install https://github.com/helm-unittest/helm-unittest.git

# Run tests
helm unittest charts/homelab-assistant-crds
```
