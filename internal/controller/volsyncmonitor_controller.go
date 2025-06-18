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
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"

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

	// Get the VolSyncMonitor instance
	var monitor volsyncv1alpha1.VolSyncMonitor
	if err := r.Get(ctx, req.NamespacedName, &monitor); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("VolSyncMonitor resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get VolSyncMonitor")
		return ctrl.Result{}, err
	}

	// Check if monitor is enabled
	if !monitor.Spec.Enabled {
		logger.Info("VolSyncMonitor is disabled, skipping reconciliation")
		return ctrl.Result{RequeueAfter: time.Minute * 5}, nil
	}

	// Update status phase
	if monitor.Status.Phase == "" {
		monitor.Status.Phase = volsyncv1alpha1.VolSyncMonitorPhaseActive
	}

	// Main reconciliation logic
	result, err := r.reconcileMonitor(ctx, &monitor)
	if err != nil {
		monitor.Status.Phase = volsyncv1alpha1.VolSyncMonitorPhaseError
		monitor.Status.LastError = err.Error()
		logger.Error(err, "Failed to reconcile VolSyncMonitor")
	} else {
		monitor.Status.Phase = volsyncv1alpha1.VolSyncMonitorPhaseActive
		monitor.Status.LastError = ""
	}

	// Update status
	monitor.Status.ObservedGeneration = monitor.Generation
	if err := r.Status().Update(ctx, &monitor); err != nil {
		logger.Error(err, "Failed to update VolSyncMonitor status")
		return ctrl.Result{}, err
	}

	return result, err
}

func (r *VolSyncMonitorReconciler) reconcileMonitor(ctx context.Context, monitor *volsyncv1alpha1.VolSyncMonitor) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Step 1: Find failed VolSync jobs
	failedJobs, err := r.findFailedVolSyncJobs(ctx, monitor)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to find failed VolSync jobs: %w", err)
	}

	// Step 2: Process each failed job
	for _, job := range failedJobs {
		// Skip if already processed
		if r.isJobAlreadyProcessed(monitor, job) {
			continue
		}

		// Check if job has lock errors
		hasLockError, lockError, err := r.checkJobForLockErrors(ctx, job, monitor.Spec.LockErrorPatterns)
		if err != nil {
			logger.Error(err, "Failed to check job for lock errors", "job", job.Name, "namespace", job.Namespace)
			continue
		}

		if hasLockError {
			logger.Info("Lock error detected in failed job", "job", job.Name, "namespace", job.Namespace, "error", lockError)
			
			// Create unlock job
			unlockJob, err := r.createUnlockJob(ctx, monitor, job, lockError)
			if err != nil {
				logger.Error(err, "Failed to create unlock job", "job", job.Name)
				continue
			}

			// Remove failed job if configured to do so
			if monitor.Spec.RemoveFailedJobs {
				if err := r.removeFailedJob(ctx, job); err != nil {
					logger.Error(err, "Failed to remove failed job", "job", job.Name)
					// Continue anyway - we still want to track the unlock job
				} else {
					monitor.Status.TotalFailedJobsRemoved++
					logger.Info("Removed failed job", "job", job.Name, "namespace", job.Namespace)
				}
			}

			// Track the processed job
			processedJob := volsyncv1alpha1.ProcessedJob{
				JobName:       job.Name,
				Namespace:     job.Namespace,
				ProcessedTime: metav1.Now(),
				UnlockJobName: unlockJob.Name,
				Removed:       monitor.Spec.RemoveFailedJobs,
				LockError:     lockError,
			}
			monitor.Status.ProcessedJobs = append(monitor.Status.ProcessedJobs, processedJob)

			// Update counters
			monitor.Status.TotalLockErrorsDetected++
			monitor.Status.TotalUnlocksCreated++
			monitor.Status.LastUnlockTime = &metav1.Time{Time: time.Now()}
		}
	}

	// Step 3: Update active unlocks status
	if err := r.updateActiveUnlocks(ctx, monitor); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to update active unlocks: %w", err)
	}

	// Step 4: Clean up old processed jobs (keep last 50)
	r.cleanupProcessedJobs(monitor)

	// Requeue after 30 seconds to continuously monitor
	return ctrl.Result{RequeueAfter: time.Second * 30}, nil
}

func (r *VolSyncMonitorReconciler) findFailedVolSyncJobs(ctx context.Context, monitor *volsyncv1alpha1.VolSyncMonitor) ([]batchv1.Job, error) {
	var failedJobs []batchv1.Job

	// Determine namespaces to search
	namespaces := []string{}
	if monitor.Spec.JobSelector != nil {
		namespaces = monitor.Spec.JobSelector.Namespaces
	}
	if len(namespaces) == 0 {
		// Get all namespaces
		var nsList corev1.NamespaceList
		if err := r.List(ctx, &nsList); err != nil {
			return nil, fmt.Errorf("failed to list namespaces: %w", err)
		}
		for _, ns := range nsList.Items {
			namespaces = append(namespaces, ns.Name)
		}
	}

	// Search in each namespace
	for _, namespace := range namespaces {
		var jobList batchv1.JobList
		listOpts := []client.ListOption{
			client.InNamespace(namespace),
		}

		if err := r.List(ctx, &jobList, listOpts...); err != nil {
			return nil, fmt.Errorf("failed to list jobs in namespace %s: %w", namespace, err)
		}

		// Filter jobs based on selector
		for _, job := range jobList.Items {
			if r.matchesJobSelector(job, monitor.Spec.JobSelector) && r.isJobFailed(job) {
				failedJobs = append(failedJobs, job)
			}
		}
	}

	return failedJobs, nil
}

func (r *VolSyncMonitorReconciler) matchesJobSelector(job batchv1.Job, selector *volsyncv1alpha1.JobSelector) bool {
	if selector == nil {
		// Default: match jobs with "volsync-" prefix
		return strings.HasPrefix(job.Name, "volsync-")
	}

	// Check name prefix
	namePrefix := selector.NamePrefix
	if namePrefix == "" {
		namePrefix = "volsync-"
	}
	if !strings.HasPrefix(job.Name, namePrefix) {
		return false
	}

	// Check label selector
	if len(selector.LabelSelector) > 0 {
		for key, value := range selector.LabelSelector {
			if jobValue, exists := job.Labels[key]; !exists || jobValue != value {
				return false
			}
		}
	}

	return true
}

func (r *VolSyncMonitorReconciler) isJobFailed(job batchv1.Job) bool {
	for _, condition := range job.Status.Conditions {
		if condition.Type == batchv1.JobFailed && condition.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

func (r *VolSyncMonitorReconciler) isJobAlreadyProcessed(monitor *volsyncv1alpha1.VolSyncMonitor, job batchv1.Job) bool {
	for _, processed := range monitor.Status.ProcessedJobs {
		if processed.JobName == job.Name && processed.Namespace == job.Namespace {
			return true
		}
	}
	return false
}

func (r *VolSyncMonitorReconciler) checkJobForLockErrors(ctx context.Context, job batchv1.Job, patterns []string) (bool, string, error) {
	// Default lock error patterns if none specified
	defaultPatterns := []string{
		"repository is already locked",
		"unable to create lock",
		"repository.*locked",
		"lock.*already exists",
	}

	if len(patterns) == 0 {
		patterns = defaultPatterns
	}

	// Compile regex patterns
	var regexes []*regexp.Regexp
	for _, pattern := range patterns {
		regex, err := regexp.Compile("(?i)" + pattern) // Case insensitive
		if err != nil {
			return false, "", fmt.Errorf("invalid regex pattern %s: %w", pattern, err)
		}
		regexes = append(regexes, regex)
	}

	// Get pods for this job
	var podList corev1.PodList
	listOpts := []client.ListOption{
		client.InNamespace(job.Namespace),
		client.MatchingLabels{"job-name": job.Name},
	}

	if err := r.List(ctx, &podList, listOpts...); err != nil {
		return false, "", fmt.Errorf("failed to list pods for job %s: %w", job.Name, err)
	}

	// Check logs of each pod
	for _, pod := range podList.Items {
		logs, err := helpers.GetPodLogs(ctx, r.Client, pod.Namespace, pod.Name, "")
		if err != nil {
			continue // Skip pods we can't get logs from
		}

		// Check each line against patterns
		lines := strings.Split(logs, "\n")
		for _, line := range lines {
			for _, regex := range regexes {
				if regex.MatchString(line) {
					return true, strings.TrimSpace(line), nil
				}
			}
		}
	}

	return false, "", nil
}

func (r *VolSyncMonitorReconciler) createUnlockJob(ctx context.Context, monitor *volsyncv1alpha1.VolSyncMonitor, failedJob batchv1.Job, lockError string) (*batchv1.Job, error) {
	logger := log.FromContext(ctx)

	// Generate unique name for unlock job
	unlockJobName := fmt.Sprintf("volsync-unlock-%s-%d", failedJob.Name, time.Now().Unix())

	// Build job spec from template
	jobSpec := r.buildUnlockJobSpec(monitor, failedJob, unlockJobName, lockError)

	unlockJob := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      unlockJobName,
			Namespace: failedJob.Namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":       "homelab-assistant",
				"app.kubernetes.io/component":  "volsync-unlock",
				"app.kubernetes.io/created-by": "volsync-monitor",
				"homelab.rafaribe.com/monitor": monitor.Name,
				"homelab.rafaribe.com/failed-job": failedJob.Name,
			},
			Annotations: map[string]string{
				"homelab.rafaribe.com/lock-error": lockError,
				"homelab.rafaribe.com/failed-job": fmt.Sprintf("%s/%s", failedJob.Namespace, failedJob.Name),
			},
		},
		Spec: *jobSpec,
	}

	// Set owner reference to the monitor
	if err := controllerutil.SetControllerReference(monitor, unlockJob, r.Scheme); err != nil {
		return nil, fmt.Errorf("failed to set controller reference: %w", err)
	}

	// Create the job
	if err := r.Create(ctx, unlockJob); err != nil {
		return nil, fmt.Errorf("failed to create unlock job: %w", err)
	}

	logger.Info("Created unlock job", "job", unlockJobName, "namespace", failedJob.Namespace, "failedJob", failedJob.Name)
	return unlockJob, nil
}

func (r *VolSyncMonitorReconciler) buildUnlockJobSpec(monitor *volsyncv1alpha1.VolSyncMonitor, failedJob batchv1.Job, unlockJobName, lockError string) *batchv1.JobSpec {
	template := monitor.Spec.UnlockJobTemplate
	
	// Default values
	if len(template.Command) == 0 {
		template.Command = []string{"/bin/sh"}
	}
	if len(template.Args) == 0 {
		template.Args = []string{"-c", "restic unlock"}
	}

	// Build container spec
	container := corev1.Container{
		Name:    "unlock",
		Image:   template.Image,
		Command: template.Command,
		Args:    template.Args,
		Env: []corev1.EnvVar{
			{
				Name:  "FAILED_JOB_NAME",
				Value: failedJob.Name,
			},
			{
				Name:  "LOCK_ERROR",
				Value: lockError,
			},
		},
	}

	// Add resource requirements if specified
	if template.Resources != nil {
		container.Resources = corev1.ResourceRequirements{}
		if template.Resources.Limits != nil {
			container.Resources.Limits = make(corev1.ResourceList)
			for k, v := range template.Resources.Limits {
				container.Resources.Limits[corev1.ResourceName(k)] = resource.MustParse(v)
			}
		}
		if template.Resources.Requests != nil {
			container.Resources.Requests = make(corev1.ResourceList)
			for k, v := range template.Resources.Requests {
				container.Resources.Requests[corev1.ResourceName(k)] = resource.MustParse(v)
			}
		}
	}

	// Build pod spec
	podSpec := corev1.PodSpec{
		RestartPolicy: corev1.RestartPolicyNever,
		Containers:    []corev1.Container{container},
	}

	// Add service account if specified
	if template.ServiceAccount != "" {
		podSpec.ServiceAccountName = template.ServiceAccount
	}

	// Add security context if specified
	if template.SecurityContext != nil {
		podSpec.SecurityContext = &corev1.PodSecurityContext{
			RunAsUser:  template.SecurityContext.RunAsUser,
			RunAsGroup: template.SecurityContext.RunAsGroup,
			FSGroup:    template.SecurityContext.FSGroup,
		}
	}

	// Build job spec
	jobSpec := &batchv1.JobSpec{
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"app.kubernetes.io/name":      "homelab-assistant",
					"app.kubernetes.io/component": "volsync-unlock",
					"homelab.rafaribe.com/unlock-job": unlockJobName,
				},
			},
			Spec: podSpec,
		},
		BackoffLimit: helpers.Int32Ptr(3),
	}

	// Add TTL if specified
	if monitor.Spec.TTLSecondsAfterFinished != nil {
		jobSpec.TTLSecondsAfterFinished = monitor.Spec.TTLSecondsAfterFinished
	}

	return jobSpec
}

func (r *VolSyncMonitorReconciler) removeFailedJob(ctx context.Context, job batchv1.Job) error {
	// Delete the job with propagation policy to clean up pods
	deletePolicy := metav1.DeletePropagationForeground
	deleteOptions := &client.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}

	return r.Delete(ctx, &job, deleteOptions)
}

func (r *VolSyncMonitorReconciler) updateActiveUnlocks(ctx context.Context, monitor *volsyncv1alpha1.VolSyncMonitor) error {
	var activeUnlocks []volsyncv1alpha1.ActiveUnlock

	// Find all unlock jobs created by this monitor
	var jobList batchv1.JobList
	listOpts := []client.ListOption{
		client.MatchingLabels{"homelab.rafaribe.com/monitor": monitor.Name},
	}

	if err := r.List(ctx, &jobList, listOpts...); err != nil {
		return fmt.Errorf("failed to list unlock jobs: %w", err)
	}

	// Check status of each unlock job
	for _, job := range jobList.Items {
		if r.isJobActive(job) {
			failedJobName := job.Labels["homelab.rafaribe.com/failed-job"]
			if failedJobName == "" {
				failedJobName = job.Annotations["homelab.rafaribe.com/failed-job"]
			}

			activeUnlock := volsyncv1alpha1.ActiveUnlock{
				AppName:          r.extractAppName(failedJobName),
				Namespace:        job.Namespace,
				ObjectName:       failedJobName,
				JobName:          job.Name,
				StartTime:        job.CreationTimestamp,
				AlertFingerprint: fmt.Sprintf("%s-%s", job.Namespace, job.Name),
			}
			activeUnlocks = append(activeUnlocks, activeUnlock)
		} else if r.isJobSucceeded(job) {
			monitor.Status.TotalUnlocksSucceeded++
		} else if r.isJobFailed(job) {
			monitor.Status.TotalUnlocksFailed++
		}
	}

	monitor.Status.ActiveUnlocks = activeUnlocks
	return nil
}

func (r *VolSyncMonitorReconciler) isJobActive(job batchv1.Job) bool {
	return job.Status.Active > 0
}

func (r *VolSyncMonitorReconciler) isJobSucceeded(job batchv1.Job) bool {
	for _, condition := range job.Status.Conditions {
		if condition.Type == batchv1.JobComplete && condition.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

func (r *VolSyncMonitorReconciler) extractAppName(jobName string) string {
	// Extract app name from job name (e.g., "volsync-prowlarr-backup" -> "prowlarr")
	parts := strings.Split(jobName, "-")
	if len(parts) >= 2 && parts[0] == "volsync" {
		return parts[1]
	}
	return jobName
}

func (r *VolSyncMonitorReconciler) cleanupProcessedJobs(monitor *volsyncv1alpha1.VolSyncMonitor) {
	// Keep only the last 50 processed jobs
	if len(monitor.Status.ProcessedJobs) > 50 {
		monitor.Status.ProcessedJobs = monitor.Status.ProcessedJobs[len(monitor.Status.ProcessedJobs)-50:]
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *VolSyncMonitorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&volsyncv1alpha1.VolSyncMonitor{}).
		Watches(
			&batchv1.Job{},
			handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []ctrl.Request {
				return r.findVolSyncMonitorsForJob(ctx, obj)
			}),
		).
		Complete(r)
}

// findVolSyncMonitorsForJob finds VolSyncMonitors that should be triggered by job events
func (r *VolSyncMonitorReconciler) findVolSyncMonitorsForJob(ctx context.Context, obj client.Object) []ctrl.Request {
	job, ok := obj.(*batchv1.Job)
	if !ok {
		return nil
	}

	// Only trigger for failed VolSync jobs
	if !strings.HasPrefix(job.Name, "volsync-") || !r.isJobFailed(*job) {
		return nil
	}

	// Find all VolSyncMonitors in the cluster
	var monitorList volsyncv1alpha1.VolSyncMonitorList
	if err := r.List(ctx, &monitorList); err != nil {
		return nil
	}

	var requests []ctrl.Request
	for _, monitor := range monitorList.Items {
		// Check if this monitor should handle this job
		if r.shouldMonitorHandleJob(monitor, *job) {
			requests = append(requests, ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name:      monitor.Name,
					Namespace: monitor.Namespace,
				},
			})
		}
	}

	return requests
}

func (r *VolSyncMonitorReconciler) shouldMonitorHandleJob(monitor volsyncv1alpha1.VolSyncMonitor, job batchv1.Job) bool {
	if !monitor.Spec.Enabled {
		return false
	}

	// Check if job matches the monitor's selector
	return r.matchesJobSelector(job, monitor.Spec.JobSelector)
}
