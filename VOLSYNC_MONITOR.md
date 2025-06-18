# VolSync Monitor Controller

## Overview

The VolSync Monitor Controller provides a complete automated solution for handling failed VolSync backup jobs with repository lock issues. It monitors all VolSync jobs, detects lock errors in failed jobs, creates unlock jobs, and optionally removes the failed jobs - all with a single Custom Resource.

## Complete Workflow

1. **Continuous Monitoring**: Watches all VolSync jobs across specified namespaces
2. **Failure Detection**: Identifies failed VolSync jobs automatically  
3. **Lock Error Analysis**: Scans pod logs for repository lock error patterns
4. **Unlock Job Creation**: Creates restic unlock jobs when lock errors are detected
5. **Failed Job Cleanup**: Optionally removes failed jobs after creating unlock jobs
6. **Status Tracking**: Maintains comprehensive status of all operations

## Key Features

- **Single Controller**: One VolSyncMonitor CR handles the entire workflow
- **Automatic Detection**: No manual intervention required
- **Configurable Patterns**: Customizable lock error detection patterns
- **Job Cleanup**: Optional removal of failed jobs after processing
- **Comprehensive Monitoring**: Tracks active unlocks, success/failure rates, and processed jobs
- **Namespace Scoping**: Monitor specific namespaces or all namespaces
- **Resource Management**: Configurable resource limits and TTL for unlock jobs

2. **Failure Detection**: When a VolSync job fails, the controller examines the failed job's pod status and container termination messages

3. **Lock Error Detection**: Checks for restic lock-related error patterns such as:
   - "repository is already locked"
   - "unable to create lock"
   - "failed to create lock"
   - "repository locked by another process"

4. **Automatic Volume Discovery**: Discovers the exact volume configuration from the failed VolSync job, including:
   - NFS mounts (like your `truenas.rafaribe.com:/mnt/storage-0/volsync`)
   - PVC mounts
   - Any other volume types used by VolSync

5. **Automatic Unlock**: When lock errors are detected, creates an unlock job that:
   - Uses the same image and environment variables as the failed job
   - Runs `restic unlock --remove-all`
   - Mounts the exact same volumes as the failed job
   - Has proper resource limits and security context

## Real-World Example: Prowlarr

When your prowlarr VolSync job fails:

**Failed Job:**
```yaml
Name: volsync-src-prowlarr-nfs
Namespace: downloads
Secret: prowlarr-volsync-nfs
Volumes:
  - name: repository
    nfs:
      server: truenas.rafaribe.com
      path: /mnt/storage-0/volsync
```

**Automatic Unlock Job Created:**
```yaml
Name: volsync-unlock-prowlarr-prowlarr-nfs-<timestamp>
Namespace: downloads
Environment: Same variables from prowlarr-volsync-nfs secret
Volumes: Exact same NFS mount discovered from failed job
Command: restic unlock --remove-all
```

**No configuration needed** - everything is discovered automatically!

## Configuration

### Basic Setup

```yaml
apiVersion: homelab.rafaribe.com/v1alpha1
kind: VolSyncMonitor
metadata:
  name: volsync-monitor-main
  namespace: homelab-assistant-system
spec:
  enabled: true
  maxConcurrentUnlocks: 3
  ttlSecondsAfterFinished: 3600
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

### Advanced Configuration

```yaml
spec:
  # Custom lock error patterns
  lockErrorPatterns:
    - "repository is already locked"
    - "unable to create lock"
    - "repository locked"
    - "lock.*already.*exists"
    - "failed to create lock"
    - "repository.*locked.*by.*another.*process"
    - "timeout.*waiting.*for.*lock"
  
  # Use different restic image if needed
  unlockJobTemplate:
    image: "restic/restic:latest"
    # Service account for additional permissions
    serviceAccount: "volsync-unlock-sa"
```

## Secret Discovery

The controller automatically discovers restic secrets using these naming patterns:

1. `{app}-volsync-nfs` (e.g., `prowlarr-volsync-nfs`)
2. `{app}-restic-secret` (e.g., `prowlarr-restic-secret`)
3. `{app}-volsync` (e.g., `prowlarr-volsync`)
4. `{objectName}-secret` (e.g., `prowlarr-nfs-secret`)

## Volume Discovery

The controller automatically discovers and replicates volume configurations from failed VolSync jobs:

- **NFS Mounts**: Copies exact server, path, and mount options
- **PVC Mounts**: Replicates PVC references and mount paths
- **Any Volume Type**: Works with whatever VolSync is already using

**Example Discovery:**
```yaml
# From failed VolSync job
volumes:
- name: repository
  nfs:
    server: truenas.rafaribe.com
    path: /mnt/storage-0/volsync
    readOnly: false

# Automatically replicated in unlock job
volumes:
- name: repository
  nfs:
    server: truenas.rafaribe.com
    path: /mnt/storage-0/volsync
    readOnly: false
```

## Monitoring

### Check Controller Status

```bash
kubectl get volsyncmonitor -o wide
```

### View Active Unlock Jobs

```bash
kubectl get jobs -l homelab.rafaribe.com/monitor=volsync-monitor-main --all-namespaces
```

### Check Unlock Job Logs

```bash
kubectl logs job/volsync-unlock-prowlarr-prowlarr-nfs-1234567890 -n downloads
```

### Monitor Controller Logs

```bash
kubectl logs -l app=homelab-assistant-controller -n homelab-assistant-system
```

## Supported Applications

The controller works with any VolSync-managed application, including:

- **Media Applications**: prowlarr, lidarr, sonarr, radarr
- **Home Automation**: home-assistant, node-red
- **Databases**: postgresql, mysql, redis
- **Any application** using VolSync with restic backend

## Troubleshooting

### Common Issues

1. **No unlock jobs created**: 
   - Check if VolSync jobs are actually failing
   - Verify lock error patterns match actual error messages
   - Ensure controller is enabled (`spec.enabled: true`)

2. **Secret not found errors**:
   - Verify secret naming follows expected patterns
   - Check secret exists in the same namespace as failed job

3. **Volume mount issues**:
   - Check that failed job has proper volume specifications
   - Verify NFS server is accessible

4. **Permission errors**:
   - Verify RBAC permissions for the controller
   - Check service account has necessary permissions

### Debug Commands

```bash
# Check failed VolSync jobs
kubectl get jobs -l app.kubernetes.io/created-by=volsync --all-namespaces | grep -v Complete

# Check pod status of failed jobs
kubectl describe pod -l job-name=volsync-src-prowlarr-nfs -n downloads

# Check controller events
kubectl get events -n homelab-assistant-system --sort-by='.lastTimestamp'

# Examine failed job volumes
kubectl get job volsync-src-prowlarr-nfs -n downloads -o yaml | grep -A 20 volumes:
```

## Benefits

- **Zero Configuration**: Automatically discovers everything from existing VolSync setup
- **Volume Agnostic**: Works with NFS, PVC, hostPath, or any volume type
- **Environment Aware**: Uses the exact same secrets and configurations as VolSync
- **Safe**: Concurrency limits prevent resource exhaustion
- **Observable**: Full status reporting and job tracking
- **Generic**: Works across all namespaces and applications
- **Efficient**: Event-driven, only acts on actual failures

## How It Replaces Manual Tasks

Instead of running:
```bash
task volsync:unlock cluster=main ns=downloads app=prowlarr
```

The controller automatically:
1. Detects the failed `volsync-src-prowlarr-nfs` job
2. Discovers it uses `prowlarr-volsync-nfs` secret
3. Finds the NFS mount to `truenas.rafaribe.com:/mnt/storage-0/volsync`
4. Creates an unlock job with identical configuration
5. Runs `restic unlock --remove-all` automatically

**Result**: Your VolSync jobs get unlocked without any manual intervention!
