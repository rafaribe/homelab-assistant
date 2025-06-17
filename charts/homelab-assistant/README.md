# homelab-assistant

A collection of Kubernetes controllers for homelab automation and management

![Version: 0.1.0](https://img.shields.io/badge/Version-0.1.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 0.1.0](https://img.shields.io/badge/AppVersion-0.1.0-informational?style=flat-square)

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

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| commonAnnotations | object | `{}` | Additional annotations to add to all resources |
| commonLabels | object | `{}` | Additional labels to add to all resources |
| controller.affinity | object | `{}` | Affinity for controller pod |
| controller.image.pullPolicy | string | `"IfNotPresent"` | Controller image pull policy |
| controller.image.repository | string | `"ghcr.io/rafaribe/homelab-assistant"` | Controller image repository |
| controller.image.tag | string | `"latest"` | Controller image tag |
| controller.nodeSelector | object | `{}` | Node selector for controller pod |
| controller.resources | object | `{"limits":{"cpu":"500m","memory":"128Mi"},"requests":{"cpu":"10m","memory":"64Mi"}}` | Resource requirements for the controller |
| controller.securityContext | object | `{"allowPrivilegeEscalation":false,"capabilities":{"drop":["ALL"]},"readOnlyRootFilesystem":true,"runAsGroup":1000,"runAsNonRoot":true,"runAsUser":1000}` | Security context for the controller |
| controller.tolerations | list | `[]` | Tolerations for controller pod |
| global.imagePullPolicy | string | `"IfNotPresent"` | Image pull policy |
| global.imageRegistry | string | `"ghcr.io"` | Image registry for all components |
| metrics.enabled | bool | `true` | Enable metrics endpoint |
| metrics.port | int | `8080` | Metrics port |
| metrics.serviceMonitor.additionalLabels | object | `{}` | Additional labels for ServiceMonitor |
| metrics.serviceMonitor.enabled | bool | `false` | Enable ServiceMonitor creation |
| metrics.serviceMonitor.interval | string | `"30s"` | Scrape interval |
| namespace.create | bool | `true` | Create namespace if it doesn't exist |
| namespace.name | string | `""` | Namespace name (defaults to Release.Namespace) |
| networkPolicy.egress | list | `[]` | Egress rules   |
| networkPolicy.enabled | bool | `false` | Enable network policy |
| networkPolicy.ingress | list | `[]` | Ingress rules |
| podDisruptionBudget.enabled | bool | `false` | Enable pod disruption budget |
| podDisruptionBudget.minAvailable | int | `1` | Minimum available pods |
| podSecurityPolicy.create | bool | `false` | Specifies whether a PodSecurityPolicy should be created |
| rbac.create | bool | `true` | Specifies whether RBAC resources should be created |
| serviceAccount.annotations | object | `{}` | Annotations to add to the service account |
| serviceAccount.create | bool | `true` | Specifies whether a service account should be created |
| serviceAccount.name | string | `""` | The name of the service account to use. If not set and create is true, a name is generated using the fullname template |
| volsyncMonitor.enabled | bool | `true` | Enable the VolSync monitor controller |
| volsyncMonitor.lockErrorPatterns | list | `[]` | Custom lock error patterns (optional) If not specified, sensible defaults will be used |
| volsyncMonitor.maxConcurrentUnlocks | int | `3` | Maximum number of concurrent unlock operations |
| volsyncMonitor.ttlSecondsAfterFinished | int | `3600` | TTL for unlock jobs (in seconds) - 1 hour default |
| volsyncMonitor.unlockJob.args | list | `["unlock","--remove-all"]` | Arguments for unlock jobs |
| volsyncMonitor.unlockJob.command | list | `["restic"]` | Command and args for unlock jobs |
| volsyncMonitor.unlockJob.image.pullPolicy | string | `"IfNotPresent"` | Unlock job image pull policy |
| volsyncMonitor.unlockJob.image.repository | string | `"quay.io/backube/volsync"` | Unlock job image repository |
| volsyncMonitor.unlockJob.image.tag | string | `"0.13.0-rc.2"` | Unlock job image tag |
| volsyncMonitor.unlockJob.resources | object | `{"limits":{"cpu":"500m","memory":"512Mi"},"requests":{"cpu":"100m","memory":"128Mi"}}` | Resource requirements for unlock jobs |
| volsyncMonitor.unlockJob.securityContext | object | `{"fsGroup":1000,"runAsGroup":1000,"runAsUser":1000}` | Security context for unlock jobs |
| volsyncMonitor.unlockJob.serviceAccount | string | `""` | Service account for unlock jobs (optional) |
| webhook.enabled | bool | `false` | Enable admission webhook |
| webhook.port | int | `9443` | Webhook port |

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

## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| rafaribe | <rafa@rafaribe.com> |  |

## Source Code

* <https://github.com/rafaribe/homelab-assistant>

