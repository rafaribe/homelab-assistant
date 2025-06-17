package controller

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/handler"

	volsyncv1alpha1 "github.com/rafaribe/homelab-assistant/api/v1alpha1"
	"github.com/rafaribe/homelab-assistant/internal/helpers"
)

// VolSyncMonitorReconciler reconciles a VolSyncMonitor object
type VolSyncMonitorReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=homelab.rafaribe.com,resources=volsyncmonitors,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=homelab.rafaribe.com,resources=volsyncmonitors/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=homelab.rafaribe.com,resources=volsyncmonitors/finalizers,verbs=update
//+kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=pods/log,verbs=get;list
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch
//+kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch

func (r *VolSyncMonitorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Check if this is a VolSync job that failed
	if strings.Contains(req.Name, "volsync-") {
		return r.handleVolSyncJob(ctx, req)
	}

	// Otherwise, handle VolSyncMonitor resource
	monitor := &volsyncv1alpha1.VolSyncMonitor{}
	err := r.Get(ctx, req.NamespacedName, monitor)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("VolSyncMonitor resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get VolSyncMonitor")
		return ctrl.Result{}, err
	}

	// Add finalizer if not present
	if !controllerutil.ContainsFinalizer(monitor, "homelab.rafaribe.com/finalizer") {
		controllerutil.AddFinalizer(monitor, "homelab.rafaribe.com/finalizer")
		return ctrl.Result{}, r.Update(ctx, monitor)
	}

	// Handle deletion
	if monitor.DeletionTimestamp != nil {
		return r.handleDeletion(ctx, monitor)
	}

	// Update status
	monitor.Status.Phase = volsyncv1alpha1.VolSyncMonitorPhaseActive
	if !monitor.Spec.Enabled {
		monitor.Status.Phase = volsyncv1alpha1.VolSyncMonitorPhasePaused
	}

	return ctrl.Result{}, r.Status().Update(ctx, monitor)
}

func (r *VolSyncMonitorReconciler) handleVolSyncJob(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Get the job
	job := &batchv1.Job{}
	err := r.Get(ctx, req.NamespacedName, job)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Only process failed VolSync jobs
	if !r.isVolSyncJob(job) || !r.isJobFailed(job) {
		return ctrl.Result{}, nil
	}

	logger.Info("Processing failed VolSync job", "job", job.Name, "namespace", job.Namespace)

	// Check if we have any active monitors
	monitors, err := r.getActiveMonitors(ctx)
	if err != nil {
		logger.Error(err, "Failed to get active monitors")
		return ctrl.Result{}, err
	}

	if len(monitors) == 0 {
		logger.Info("No active VolSync monitors found")
		return ctrl.Result{}, nil
	}

	// Check job logs for lock errors
	hasLockError, err := r.checkJobLogsForLockErrors(ctx, job, monitors[0])
	if err != nil {
		logger.Error(err, "Failed to check job logs")
		return ctrl.Result{}, err
	}

	if !hasLockError {
		logger.Info("No lock errors found in job logs", "job", job.Name)
		return ctrl.Result{}, nil
	}

	// Extract app info from job
	appName, objectName := r.extractAppInfoFromJob(job)
	if appName == "" || objectName == "" {
		logger.Info("Could not extract app info from job", "job", job.Name)
		return ctrl.Result{}, nil
	}

	// Create unlock job
	for _, monitor := range monitors {
		if r.canCreateUnlockJob(monitor) {
			err := r.createUnlockJobForFailedJob(ctx, &monitor, job, appName, objectName)
			if err != nil {
				logger.Error(err, "Failed to create unlock job")
				continue
			}
			logger.Info("Created unlock job for failed VolSync job", 
				"volsyncJob", job.Name, 
				"app", appName, 
				"object", objectName)
			break
		}
	}

	return ctrl.Result{}, nil
}

func (r *VolSyncMonitorReconciler) isVolSyncJob(job *batchv1.Job) bool {
	// Check if job is created by VolSync
	if job.Labels != nil {
		if createdBy, exists := job.Labels["app.kubernetes.io/created-by"]; exists && createdBy == "volsync" {
			return true
		}
	}
	
	// Check job name patterns
	return strings.HasPrefix(job.Name, "volsync-src-") || strings.HasPrefix(job.Name, "volsync-dst-")
}

func (r *VolSyncMonitorReconciler) isJobFailed(job *batchv1.Job) bool {
	return job.Status.Failed > 0
}

func (r *VolSyncMonitorReconciler) getActiveMonitors(ctx context.Context) ([]volsyncv1alpha1.VolSyncMonitor, error) {
	monitorList := &volsyncv1alpha1.VolSyncMonitorList{}
	err := r.List(ctx, monitorList)
	if err != nil {
		return nil, err
	}

	var activeMonitors []volsyncv1alpha1.VolSyncMonitor
	for _, monitor := range monitorList.Items {
		if monitor.Spec.Enabled {
			activeMonitors = append(activeMonitors, monitor)
		}
	}

	return activeMonitors, nil
}

func (r *VolSyncMonitorReconciler) checkJobLogsForLockErrors(ctx context.Context, job *batchv1.Job, monitor volsyncv1alpha1.VolSyncMonitor) (bool, error) {
	logger := log.FromContext(ctx)

	// Get pods for this job
	podList := &corev1.PodList{}
	err := r.List(ctx, podList, client.InNamespace(job.Namespace), client.MatchingLabels{
		"job-name": job.Name,
	})
	if err != nil {
		return false, err
	}

	if len(podList.Items) == 0 {
		logger.Info("No pods found for job", "job", job.Name)
		return false, nil
	}

	// Default lock error patterns if none specified
	lockPatterns := monitor.Spec.LockErrorPatterns
	if len(lockPatterns) == 0 {
		lockPatterns = []string{
			"repository is already locked",
			"unable to create lock",
			"repository locked",
			"lock.*already.*exists",
			"failed to create lock",
		}
	}

	// Compile regex patterns
	var regexPatterns []*regexp.Regexp
	for _, pattern := range lockPatterns {
		regex, err := regexp.Compile("(?i)" + pattern) // Case insensitive
		if err != nil {
			logger.Error(err, "Invalid regex pattern", "pattern", pattern)
			continue
		}
		regexPatterns = append(regexPatterns, regex)
	}

	// Check logs of each pod
	for _, pod := range podList.Items {
		if pod.Status.Phase == corev1.PodFailed {
			logs, err := r.getPodLogs(ctx, &pod)
			if err != nil {
				logger.Error(err, "Failed to get pod logs", "pod", pod.Name)
				continue
			}

			// Check for lock error patterns
			for _, regex := range regexPatterns {
				if regex.MatchString(logs) {
					// Extract app info for metrics
					appName, objectName := r.extractAppInfoFromJob(job)
					
					// Record metrics
					helpers.RecordLockErrorDetected(job.Namespace, appName, objectName, regex.String())
					
					logger.Info("Lock error detected in pod logs", 
						"pod", pod.Name, 
						"job", job.Name,
						"pattern", regex.String(),
						"app", appName,
						"object", objectName)
					return true, nil
				}
			}
		}
	}

	return false, nil
}

func (r *VolSyncMonitorReconciler) getPodLogs(ctx context.Context, pod *corev1.Pod) (string, error) {
	// For now, we'll check the pod status and events instead of logs
	// This is simpler and doesn't require additional RBAC permissions for log streaming
	
	// Check pod status message
	if pod.Status.Message != "" {
		return pod.Status.Message, nil
	}
	
	// Check container statuses
	var messages []string
	for _, containerStatus := range pod.Status.ContainerStatuses {
		if containerStatus.State.Terminated != nil && containerStatus.State.Terminated.Message != "" {
			messages = append(messages, containerStatus.State.Terminated.Message)
		}
		if containerStatus.State.Waiting != nil && containerStatus.State.Waiting.Message != "" {
			messages = append(messages, containerStatus.State.Waiting.Message)
		}
	}
	
	return strings.Join(messages, "\n"), nil
}

func (r *VolSyncMonitorReconciler) extractAppInfoFromJob(job *batchv1.Job) (string, string) {
	// Try to extract from job name
	// Pattern: volsync-src-<objectName> or volsync-dst-<objectName>
	name := job.Name
	if strings.HasPrefix(name, "volsync-src-") {
		objectName := strings.TrimPrefix(name, "volsync-src-")
		return r.guessAppNameFromObjectName(objectName), objectName
	}
	if strings.HasPrefix(name, "volsync-dst-") {
		objectName := strings.TrimPrefix(name, "volsync-dst-")
		return r.guessAppNameFromObjectName(objectName), objectName
	}

	// Try to extract from labels
	if job.Labels != nil {
		if app, exists := job.Labels["app"]; exists {
			return app, job.Name
		}
	}

	return "", ""
}

func (r *VolSyncMonitorReconciler) guessAppNameFromObjectName(objectName string) string {
	// Common patterns: app-nfs, app-pvc, etc.
	parts := strings.Split(objectName, "-")
	if len(parts) > 1 {
		return parts[0] // Return first part as app name
	}
	return objectName
}

func (r *VolSyncMonitorReconciler) canCreateUnlockJob(monitor volsyncv1alpha1.VolSyncMonitor) bool {
	maxConcurrent := monitor.Spec.MaxConcurrentUnlocks
	if maxConcurrent == 0 {
		maxConcurrent = 3 // Default
	}
	
	return len(monitor.Status.ActiveUnlocks) < int(maxConcurrent)
}

func (r *VolSyncMonitorReconciler) createUnlockJobForFailedJob(ctx context.Context, monitor *volsyncv1alpha1.VolSyncMonitor, failedJob *batchv1.Job, appName, objectName string) error {
	logger := log.FromContext(ctx)
	
	// Record metrics
	helpers.RecordUnlockJobCreated(failedJob.Namespace, appName, objectName)
	helpers.RecordActiveUnlockJob(failedJob.Namespace, appName, objectName)

	// Update monitor status
	now := metav1.Now()
	monitor.Status.TotalUnlocksCreated++
	monitor.Status.LastUnlockTime = &now
	
	// Add to active unlocks
	jobName := fmt.Sprintf("volsync-unlock-%s-%s-%d", appName, objectName, time.Now().Unix())
	activeUnlock := volsyncv1alpha1.ActiveUnlock{
		AppName:          appName,
		Namespace:        failedJob.Namespace,
		ObjectName:       objectName,
		JobName:          jobName,
		StartTime:        now,
		AlertFingerprint: fmt.Sprintf("%s-%s-%s", failedJob.Namespace, appName, objectName),
	}
	
	monitor.Status.ActiveUnlocks = append(monitor.Status.ActiveUnlocks, activeUnlock)

	// Update status
	if err := r.Status().Update(ctx, monitor); err != nil {
		logger.Error(err, "Failed to update monitor status")
		return err
	}

	// Build restic unlock command
	command := []string{"restic"}
	args := []string{"unlock", "--remove-all"}

	// Override with custom command/args if provided
	if len(monitor.Spec.UnlockJobTemplate.Command) > 0 {
		command = monitor.Spec.UnlockJobTemplate.Command
	}
	if len(monitor.Spec.UnlockJobTemplate.Args) > 0 {
		args = monitor.Spec.UnlockJobTemplate.Args
	}

	// Get environment variables from the secret (same pattern as VolSync)
	envVars, err := r.buildResticEnvVars(ctx, appName, failedJob.Namespace, objectName)
	if err != nil {
		return fmt.Errorf("failed to build environment variables: %w", err)
	}

	// Discover volume configuration from the failed VolSync job
	volumes, volumeMounts, err := r.discoverVolSyncVolumeConfig(ctx, failedJob)
	if err != nil {
		return fmt.Errorf("failed to discover VolSync volume configuration: %w", err)
	}

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: failedJob.Namespace,
			Labels: map[string]string{
				"app":                              "volsync-unlock",
				"homelab.rafaribe.com/app":         appName,
				"homelab.rafaribe.com/object":      objectName,
				"homelab.rafaribe.com/monitor":     monitor.Name,
				"homelab.rafaribe.com/failed-job":  failedJob.Name,
			},
			// Copy annotations that might trigger admission policies
			Annotations: map[string]string{
				"homelab.rafaribe.com/volsync-unlock": "true",
				"homelab.rafaribe.com/app":            appName,
				"homelab.rafaribe.com/object":         objectName,
			},
		},
		Spec: batchv1.JobSpec{
			TTLSecondsAfterFinished: monitor.Spec.TTLSecondsAfterFinished,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":                              "volsync-unlock",
						"homelab.rafaribe.com/app":         appName,
						"homelab.rafaribe.com/object":      objectName,
					},
					// Copy annotations to pod template for admission policies
					Annotations: map[string]string{
						"homelab.rafaribe.com/volsync-unlock": "true",
						"homelab.rafaribe.com/app":            appName,
						"homelab.rafaribe.com/object":         objectName,
					},
				},
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyNever,
					Containers: []corev1.Container{
						{
							Name:         "restic-unlock",
							Image:        monitor.Spec.UnlockJobTemplate.Image,
							Command:      command,
							Args:         args,
							Env:          envVars,
							VolumeMounts: volumeMounts,
							Resources: corev1.ResourceRequirements{
								Limits:   r.convertResources(r.getResourceLimits(monitor.Spec.UnlockJobTemplate.Resources)),
								Requests: r.convertResources(r.getResourceRequests(monitor.Spec.UnlockJobTemplate.Resources)),
							},
						},
					},
					Volumes: volumes,
				},
			},
		},
	}

	// Set security context if provided
	if monitor.Spec.UnlockJobTemplate.SecurityContext != nil {
		sc := monitor.Spec.UnlockJobTemplate.SecurityContext
		job.Spec.Template.Spec.SecurityContext = &corev1.PodSecurityContext{
			RunAsUser:  sc.RunAsUser,
			RunAsGroup: sc.RunAsGroup,
			FSGroup:    sc.FSGroup,
		}
	}

	// Set service account if provided
	if monitor.Spec.UnlockJobTemplate.ServiceAccount != "" {
		job.Spec.Template.Spec.ServiceAccountName = monitor.Spec.UnlockJobTemplate.ServiceAccount
	}

	// Set owner reference
	if err := controllerutil.SetControllerReference(monitor, job, r.Scheme); err != nil {
		return fmt.Errorf("failed to set controller reference: %w", err)
	}

	if err := r.Create(ctx, job); err != nil {
		return fmt.Errorf("failed to create job: %w", err)
	}

	// Update monitor status
	activeUnlock := volsyncv1alpha1.ActiveUnlock{
		AppName:          appName,
		Namespace:        failedJob.Namespace,
		ObjectName:       objectName,
		JobName:          jobName,
		StartTime:        metav1.Now(),
		AlertFingerprint: failedJob.Name, // Use failed job name as fingerprint
	}
	monitor.Status.ActiveUnlocks = append(monitor.Status.ActiveUnlocks, activeUnlock)
	monitor.Status.TotalUnlocksCreated++

	return r.Status().Update(ctx, monitor)
}

func (r *VolSyncMonitorReconciler) buildResticEnvVars(ctx context.Context, appName, namespace, objectName string) ([]corev1.EnvVar, error) {
	// Try different secret naming patterns that VolSync might use
	secretNames := []string{
		fmt.Sprintf("%s-volsync-nfs", appName),     // Pattern from your example
		fmt.Sprintf("%s-restic-secret", appName),   // Common pattern
		fmt.Sprintf("%s-volsync", appName),         // Alternative pattern
		fmt.Sprintf("%s-secret", objectName),       // Object-based naming
	}

	var secretName string
	for _, name := range secretNames {
		secret := &corev1.Secret{}
		err := r.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, secret)
		if err == nil {
			secretName = name
			break
		}
	}

	if secretName == "" {
		return nil, fmt.Errorf("could not find restic secret for app %s in namespace %s", appName, namespace)
	}

	// Build comprehensive environment variables (same as VolSync uses)
	envVars := []corev1.EnvVar{
		{
			Name: "RESTIC_REPOSITORY",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: secretName},
					Key: "RESTIC_REPOSITORY",
				},
			},
		},
		{
			Name: "RESTIC_PASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: secretName},
					Key: "RESTIC_PASSWORD",
				},
			},
		},
	}

	// Add all the optional environment variables that VolSync supports
	optionalEnvVars := []string{
		"RESTIC_COMPRESSION", "RESTIC_PACK_SIZE", "RESTIC_READ_CONCURRENCY",
		"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY", "AWS_SESSION_TOKEN", "AWS_DEFAULT_REGION", "AWS_PROFILE",
		"RESTIC_AWS_ASSUME_ROLE_ARN", "RESTIC_AWS_ASSUME_ROLE_SESSION_NAME", "RESTIC_AWS_ASSUME_ROLE_EXTERNAL_ID",
		"RESTIC_AWS_ASSUME_ROLE_POLICY", "RESTIC_AWS_ASSUME_ROLE_REGION", "RESTIC_AWS_ASSUME_ROLE_STS_ENDPOINT",
		"ST_AUTH", "ST_USER", "ST_KEY",
		"OS_AUTH_URL", "OS_REGION_NAME", "OS_USERNAME", "OS_USER_ID", "OS_PASSWORD",
		"OS_TENANT_ID", "OS_TENANT_NAME", "OS_USER_DOMAIN_NAME", "OS_USER_DOMAIN_ID",
		"OS_PROJECT_NAME", "OS_PROJECT_DOMAIN_NAME", "OS_PROJECT_DOMAIN_ID", "OS_TRUST_ID",
		"OS_APPLICATION_CREDENTIAL_ID", "OS_APPLICATION_CREDENTIAL_NAME", "OS_APPLICATION_CREDENTIAL_SECRET",
		"OS_STORAGE_URL", "OS_AUTH_TOKEN",
		"B2_ACCOUNT_ID", "B2_ACCOUNT_KEY",
		"AZURE_ACCOUNT_NAME", "AZURE_ACCOUNT_KEY", "AZURE_ACCOUNT_SAS", "AZURE_ENDPOINT_SUFFIX",
		"GOOGLE_PROJECT_ID",
		"RESTIC_REST_USERNAME", "RESTIC_REST_PASSWORD",
	}

	optional := true
	for _, envVar := range optionalEnvVars {
		envVars = append(envVars, corev1.EnvVar{
			Name: envVar,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: secretName},
					Key:      envVar,
					Optional: &optional,
				},
			},
		})
	}

	return envVars, nil
}

func (r *VolSyncMonitorReconciler) discoverVolSyncVolumeConfig(ctx context.Context, failedJob *batchv1.Job) ([]corev1.Volume, []corev1.VolumeMount, error) {
	logger := log.FromContext(ctx)

	// Extract volumes and volume mounts from the failed job
	var volumes []corev1.Volume
	var volumeMounts []corev1.VolumeMount

	// Look for repository-related volumes (typically named "repository")
	for _, volume := range failedJob.Spec.Template.Spec.Volumes {
		if volume.Name == "repository" {
			volumes = append(volumes, volume)
			break
		}
	}

	// Look for repository-related volume mounts
	if len(failedJob.Spec.Template.Spec.Containers) > 0 {
		for _, mount := range failedJob.Spec.Template.Spec.Containers[0].VolumeMounts {
			if mount.Name == "repository" {
				volumeMounts = append(volumeMounts, mount)
				break
			}
		}
	}

	logger.Info("Discovered VolSync volume configuration from failed job", 
		"volumes", len(volumes), 
		"volumeMounts", len(volumeMounts),
		"failedJob", failedJob.Name)

	return volumes, volumeMounts, nil
}
func (r *VolSyncMonitorReconciler) getResourceLimits(resources *volsyncv1alpha1.ResourceRequirements) map[string]string {
	if resources == nil {
		return nil
	}
	return resources.Limits
}

func (r *VolSyncMonitorReconciler) getResourceRequests(resources *volsyncv1alpha1.ResourceRequirements) map[string]string {
	if resources == nil {
		return nil
	}
	return resources.Requests
}

func (r *VolSyncMonitorReconciler) convertResources(resources map[string]string) corev1.ResourceList {
	if resources == nil {
		return nil
	}

	result := make(corev1.ResourceList)
	for k, v := range resources {
		if quantity, err := resource.ParseQuantity(v); err == nil {
			result[corev1.ResourceName(k)] = quantity
		}
	}
	return result
}

func (r *VolSyncMonitorReconciler) handleDeletion(ctx context.Context, monitor *volsyncv1alpha1.VolSyncMonitor) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Handling VolSyncMonitor deletion")

	// Clean up any remaining jobs
	jobList := &batchv1.JobList{}
	err := r.List(ctx, jobList, client.InNamespace(monitor.Namespace), client.MatchingLabels{
		"homelab.rafaribe.com/monitor": monitor.Name,
	})
	if err != nil {
		logger.Error(err, "Failed to list jobs for cleanup")
	} else {
		for _, job := range jobList.Items {
			if err := r.Delete(ctx, &job); err != nil && !errors.IsNotFound(err) {
				logger.Error(err, "Failed to delete job", "job", job.Name)
			}
		}
	}

	// Remove finalizer
	controllerutil.RemoveFinalizer(monitor, "homelab.rafaribe.com/finalizer")
	return ctrl.Result{}, r.Update(ctx, monitor)
}

// SetupWithManager sets up the controller with the Manager.
func (r *VolSyncMonitorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&volsyncv1alpha1.VolSyncMonitor{}).
		Watches(
			&batchv1.Job{},
			&handler.EnqueueRequestForObject{},
		).
		Owns(&batchv1.Job{}).
		Complete(r)
}
