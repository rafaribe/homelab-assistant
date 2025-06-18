# VolSync Monitor Examples

This directory contains real-world examples of how to use the VolSync Monitor Controller.

## Basic Example

The simplest configuration that monitors all VolSync jobs and creates unlock jobs when lock errors are detected:

```yaml
apiVersion: homelab.rafaribe.com/v1alpha1
kind: VolSyncMonitor
metadata:
  name: volsync-monitor
  namespace: homelab-assistant-system
spec:
  enabled: true
  removeFailedJobs: true
  unlockJobTemplate:
    image: "restic/restic:latest"
    command: ["/bin/sh"]
    args: ["-c", "restic unlock"]
```

## Complete Workflow Example

When your Prowlarr VolSync backup fails with "repository is already locked":

1. **Detection**: Controller detects the lock error in pod logs
2. **Metrics**: Records `volsync_lock_errors_detected_total`
3. **Action**: Creates unlock job `volsync-unlock-prowlarr-prowlarr-nfs-123456`
4. **Monitoring**: Updates VolSyncMonitor status with active unlock info
5. **Cleanup**: Removes the failed job (if `removeFailedJobs: true`)
6. **Completion**: Records success metrics and cleans up

## Status Monitoring

Check the status of your VolSync Monitor:

```bash
kubectl get volsyncmonitor -n homelab-assistant-system
```

Output:
```
NAME              PHASE    ACTIVE UNLOCKS   TOTAL CREATED   JOBS REMOVED   AGE
volsync-monitor   Active   1                5               3              2d
```

View detailed status:
```bash
kubectl describe volsyncmonitor volsync-monitor -n homelab-assistant-system
```

## Files

- `volsync-monitor-complete.yaml` - Complete example with all options
- `volsync-monitor-with-nfs.yaml` - Advanced example with NFS repository mounting
- `volsync-monitor-basic.yaml` - Minimal configuration for quick setup

## Metrics

The controller exposes these Prometheus metrics:

- `volsync_unlock_jobs_created_total` - Total unlock jobs created
- `volsync_unlock_jobs_succeeded_total` - Successful unlock jobs
- `volsync_unlock_jobs_failed_total` - Failed unlock jobs
- `volsync_active_unlock_jobs` - Currently active unlock jobs
- `volsync_lock_errors_detected_total` - Lock errors detected

## Troubleshooting

### Check if monitor is working:
```bash
kubectl logs -n homelab-assistant-system deployment/homelab-assistant-controller-manager
```

### View processed jobs:
```bash
kubectl get volsyncmonitor volsync-monitor -n homelab-assistant-system -o jsonpath='{.status.processedJobs}' | jq
```

### Check unlock job logs:
```bash
kubectl logs -n <namespace> job/volsync-unlock-<job-name>-<timestamp>
```
