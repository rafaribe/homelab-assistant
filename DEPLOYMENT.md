# Homelab Assistant Deployment Guide

This guide covers different ways to deploy the Homelab Assistant controllers in your homelab environment.

## Prerequisites

- Kubernetes cluster (1.25+)
- VolSync operator already installed (for VolSync Monitor controller)
- Prometheus operator (optional, for metrics)
- ArgoCD (optional, for GitOps deployment)

## Deployment Options

### Option 1: Helm Chart with GHCR OCI Registry (Recommended)

#### 1.1 Install CRDs First

```bash
# Install CRDs from GHCR OCI registry
helm install homelab-assistant-crds \
  oci://ghcr.io/rafaribe/homelab-assistant-crds \
  --version 0.1.0 \
  --namespace homelab-assistant-system \
  --create-namespace
```

#### 1.2 Install the Controllers

```bash
# Install the controllers from GHCR OCI registry
helm install homelab-assistant \
  oci://ghcr.io/rafaribe/homelab-assistant \
  --version 0.1.0 \
  --namespace homelab-assistant-system \
  --values values.yaml
```

#### 1.3 Alternative: Traditional Helm Repository

```bash
# Add the Helm repository (if you prefer traditional approach)
helm repo add homelab-assistant https://rafaribe.github.io/homelab-assistant
helm repo update

# Install CRDs
helm install homelab-assistant-crds homelab-assistant/homelab-assistant-crds \
  --namespace homelab-assistant-system \
  --create-namespace

# Install Controllers
helm install homelab-assistant homelab-assistant/homelab-assistant \
  --namespace homelab-assistant-system \
  --values values.yaml
```

**Example values.yaml:**

```yaml
# Controller configuration
controller:
  image:
    repository: ghcr.io/rafaribe/homelab-assistant
    tag: "latest"
    pullPolicy: IfNotPresent
  
  resources:
    limits:
      cpu: 500m
      memory: 128Mi
    requests:
      cpu: 10m
      memory: 64Mi

# VolSync Monitor Controller configuration
volsyncMonitor:
  enabled: true
  maxConcurrentUnlocks: 3
  ttlSecondsAfterFinished: 3600
  
  # Custom lock error patterns
  lockErrorPatterns:
    - "repository is already locked"
    - "unable to create lock"
    - "repository locked"
    - "lock.*already.*exists"
    - "failed to create lock"
    - "repository.*locked.*by.*another.*process"
  
  unlockJob:
    image:
      repository: quay.io/backube/volsync
      tag: "0.13.0-rc.2"
    
    resources:
      limits:
        cpu: "500m"
        memory: "512Mi"
      requests:
        cpu: "100m"
        memory: "128Mi"
    
    securityContext:
      runAsUser: 1000
      runAsGroup: 1000
      fsGroup: 1000

# Future controllers can be added here
# backupController:
#   enabled: false
# 
# networkController:
#   enabled: false

# Metrics configuration
metrics:
  enabled: true
  serviceMonitor:
    enabled: true
    additionalLabels:
      release: prometheus
```

### Option 2: App-Template with Flux (Recommended for Homelabs)

Perfect for homelabs already using Flux and the popular app-template pattern:

#### 2.1 Flux Repository Structure

```
clusters/homelab/
├── infrastructure/
│   └── homelab-assistant/
│       ├── kustomization.yaml
│       ├── namespace.yaml
│       ├── helmrepository.yaml
│       └── helmrelease-crds.yaml
└── apps/
    └── homelab-assistant/
        ├── kustomization.yaml
        ├── helmrepository.yaml
        ├── helmrelease.yaml
        └── volsync-monitor.yaml
```

#### 2.2 Infrastructure Layer (CRDs)

```yaml
# clusters/homelab/infrastructure/homelab-assistant/helmrelease-crds.yaml
apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
metadata:
  name: homelab-assistant-crds
  namespace: homelab-assistant-system
spec:
  interval: 30m
  chart:
    spec:
      chart: homelab-assistant-crds
      version: "0.1.x"
      sourceRef:
        kind: HelmRepository
        name: homelab-assistant
        namespace: flux-system
  values:
    installCRDs: true
```

#### 2.3 Application Layer (Controller)

```yaml
# clusters/homelab/apps/homelab-assistant/helmrelease.yaml
apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
metadata:
  name: homelab-assistant
  namespace: homelab-assistant-system
spec:
  interval: 30m
  dependsOn:
    - name: homelab-assistant-crds
      namespace: homelab-assistant-system
  chart:
    spec:
      chart: app-template
      version: "3.0.4"
      sourceRef:
        kind: HelmRepository
        name: bjw-s
        namespace: flux-system
  values:
    controllers:
      homelab-assistant:
        type: deployment
        containers:
          app:
            image:
              repository: ghcr.io/rafaribe/homelab-assistant
              tag: latest
            command: ["/manager"]
            args:
              - --leader-elect
              - --health-probe-bind-address=:8081
              - --metrics-bind-address=:8080
            probes:
              liveness:
                enabled: true
                custom: true
                spec:
                  httpGet:
                    path: /healthz
                    port: 8081
              readiness:
                enabled: true
                custom: true
                spec:
                  httpGet:
                    path: /readyz
                    port: 8081
            resources:
              limits:
                cpu: 500m
                memory: 128Mi
              requests:
                cpu: 10m
                memory: 64Mi
    
    # RBAC is handled by app-template
    serviceAccount:
      create: true
    
    rbac:
      create: true
      rules:
        - apiGroups: ["homelab.rafaribe.com"]
          resources: ["volsyncmonitors", "volsyncunlocks"]
          verbs: ["*"]
        - apiGroups: ["batch"]
          resources: ["jobs"]
          verbs: ["*"]
        - apiGroups: [""]
          resources: ["pods", "pods/log", "secrets", "events", "namespaces", "configmaps"]
          verbs: ["get", "list", "watch", "create", "patch"]
        - apiGroups: ["coordination.k8s.io"]
          resources: ["leases"]
          verbs: ["*"]
    
    service:
      app:
        controller: homelab-assistant
        ports:
          metrics:
            port: 8080
    
    serviceMonitor:
      app:
        serviceName: homelab-assistant
        endpoints:
          - port: metrics
            path: /metrics
            interval: 30s
```

### Option 3: GitOps with Flux

#### 3.1 CRDs HelmRelease

```yaml
apiVersion: source.toolkit.fluxcd.io/v1beta2
kind: HelmRepository
metadata:
  name: homelab-assistant
  namespace: flux-system
spec:
  interval: 30m
  url: https://rafaribe.github.io/homelab-assistant

---
apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
metadata:
  name: homelab-assistant-crds
  namespace: homelab-assistant-system
spec:
  interval: 30m
  chart:
    spec:
      chart: homelab-assistant-crds
      version: "0.1.x"
      sourceRef:
        kind: HelmRepository
        name: homelab-assistant
        namespace: flux-system
  install:
    createNamespace: true
    remediation:
      retries: 3
  upgrade:
    remediation:
      retries: 3
  values:
    installCRDs: true
```

#### 3.2 Controller HelmRelease

```yaml
apiVersion: source.toolkit.fluxcd.io/v1beta2
kind: HelmRepository
metadata:
  name: bjw-s-labs
  namespace: flux-system
spec:
  interval: 30m
  url: https://bjw-s-labs.github.io/helm-charts

---
apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
metadata:
  name: homelab-assistant
  namespace: homelab-assistant-system
spec:
  interval: 30m
  dependsOn:
    - name: homelab-assistant-crds
      namespace: homelab-assistant-system
  chart:
    spec:
      chart: app-template
      version: "4.1.1"
      sourceRef:
        kind: HelmRepository
        name: bjw-s-labs
        namespace: flux-system
  values:
    # ... (same as above)
```

## Configuration

### VolSyncMonitor Resource

The core configuration is done through the `VolSyncMonitor` custom resource:

```yaml
apiVersion: homelab.rafaribe.com/v1alpha1
kind: VolSyncMonitor
metadata:
  name: homelab-main
  namespace: volsync-monitor-system
spec:
  # Enable/disable the monitor
  enabled: true
  
  # Maximum concurrent unlock operations
  maxConcurrentUnlocks: 3
  
  # TTL for unlock jobs (seconds)
  ttlSecondsAfterFinished: 3600
  
  # Custom lock error patterns (regex)
  lockErrorPatterns:
    - "repository is already locked"
    - "unable to create lock"
    - "repository locked"
    - "lock.*already.*exists"
    - "failed to create lock"
    - "repository.*locked.*by.*another.*process"
    - "timeout.*waiting.*for.*lock"
  
  # Unlock job template
  unlockJobTemplate:
    image: "quay.io/backube/volsync:0.13.0-rc.2"
    command: ["restic"]
    args: ["unlock", "--remove-all"]
    
    resources:
      limits:
        cpu: "500m"
        memory: "512Mi"
      requests:
        cpu: "100m"
        memory: "128Mi"
    
    securityContext:
      runAsUser: 1000
      runAsGroup: 1000
      fsGroup: 1000
```

## Monitoring and Observability

### Prometheus Metrics

The controller exposes several metrics:

- `volsync_unlock_jobs_created_total` - Total unlock jobs created
- `volsync_unlock_jobs_succeeded_total` - Total successful unlock jobs
- `volsync_unlock_jobs_failed_total` - Total failed unlock jobs
- `volsync_active_unlock_jobs` - Current active unlock jobs
- `volsync_lock_errors_detected_total` - Total lock errors detected
- `volsync_monitor_reconciliations_total` - Total reconciliations

### Grafana Dashboard

Import the provided dashboard JSON or create custom panels:

```promql
# Unlock success rate
sum(volsync_unlock_jobs_succeeded_total) / sum(volsync_unlock_jobs_created_total) * 100

# Lock errors by application
sum by (namespace, app) (volsync_lock_errors_detected_total)

# Active unlock jobs
sum(volsync_active_unlock_jobs)
```

### Alerting Rules

```yaml
groups:
- name: volsync-monitor
  rules:
  - alert: VolSyncUnlockJobsFailing
    expr: increase(volsync_unlock_jobs_failed_total[5m]) > 0
    for: 2m
    labels:
      severity: warning
    annotations:
      summary: "VolSync unlock jobs are failing"
  
  - alert: VolSyncHighLockErrorRate
    expr: rate(volsync_lock_errors_detected_total[5m]) > 0.1
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "High rate of VolSync lock errors"
```

## Real-World Example: Prowlarr

When prowlarr's VolSync job fails with a lock error:

1. **Failed Job**: `volsync-src-prowlarr-nfs` in `downloads` namespace
2. **Detection**: Controller detects "repository is already locked" in pod logs
3. **Metrics**: Increments `volsync_lock_errors_detected_total{namespace="downloads",app="prowlarr",object="prowlarr-nfs"}`
4. **Unlock Job**: Creates `volsync-unlock-prowlarr-prowlarr-nfs-1234567890`
5. **Status Update**: Updates VolSyncMonitor status with active unlock info
6. **Completion**: Records success/failure metrics and updates status

## Troubleshooting

### Check Controller Status

```bash
kubectl get volsyncmonitor -o wide
kubectl describe volsyncmonitor homelab-main
```

### View Unlock Jobs

```bash
kubectl get jobs -l app=volsync-unlock --all-namespaces
kubectl logs job/volsync-unlock-prowlarr-prowlarr-nfs-1234567890 -n downloads
```

### Monitor Metrics

```bash
kubectl port-forward svc/volsync-monitor-metrics 8080:8080
curl http://localhost:8080/metrics | grep volsync
```

### Common Issues

1. **No unlock jobs created**: Check if VolSync jobs are actually failing with lock errors
2. **Permission errors**: Verify RBAC permissions for the controller
3. **Secret not found**: Ensure restic secrets follow naming conventions
4. **Volume mount issues**: Check that failed jobs have proper volume specifications

## Upgrading

### Helm Upgrade

```bash
helm repo update
helm upgrade volsync-monitor volsync-monitor/volsync-monitor \
  --namespace volsync-monitor-system
```

### ArgoCD Upgrade

Update the `targetRevision` in your ArgoCD applications and sync.

## Uninstalling

### Helm

```bash
helm uninstall volsync-monitor --namespace volsync-monitor-system
helm uninstall volsync-monitor-crds --namespace volsync-monitor-system
```

### Manual Cleanup

```bash
kubectl delete volsyncmonitor --all --all-namespaces
kubectl delete crd volsyncmonitors.homelab.rafaribe.com
kubectl delete crd volsyncunlocks.homelab.rafaribe.com
kubectl delete namespace volsync-monitor-system
```

## Support

- **GitHub Issues**: https://github.com/rafaribe/homelab-assistant/issues
- **Documentation**: https://github.com/rafaribe/homelab-assistant/blob/main/VOLSYNC_MONITOR.md
- **Examples**: https://github.com/rafaribe/homelab-assistant/tree/main/examples
