package controller

import (
	"context"
	"fmt"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	volsyncv1alpha1 "github.com/rafaribe/homelab-assistant/api/v1alpha1"
)

// VolSyncUnlockReconciler reconciles a VolSyncUnlock object
type VolSyncUnlockReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=homelab.rafaribe.com,resources=volsyncunlocks,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=homelab.rafaribe.com,resources=volsyncunlocks/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=homelab.rafaribe.com,resources=volsyncunlocks/finalizers,verbs=update
//+kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *VolSyncUnlockReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the VolSyncUnlock instance
	volsyncUnlock := &volsyncv1alpha1.VolSyncUnlock{}
	err := r.Get(ctx, req.NamespacedName, volsyncUnlock)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("VolSyncUnlock resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get VolSyncUnlock")
		return ctrl.Result{}, err
	}

	// Handle deletion
	if volsyncUnlock.DeletionTimestamp != nil {
		return r.handleDeletion(ctx, volsyncUnlock)
	}

	// Add finalizer if not present
	if !controllerutil.ContainsFinalizer(volsyncUnlock, "homelab.rafaribe.com/finalizer") {
		controllerutil.AddFinalizer(volsyncUnlock, "homelab.rafaribe.com/finalizer")
		return ctrl.Result{}, r.Update(ctx, volsyncUnlock)
	}

	// Check current phase and handle accordingly
	switch volsyncUnlock.Status.Phase {
	case "":
		return r.handlePending(ctx, volsyncUnlock)
	case volsyncv1alpha1.VolSyncUnlockPhasePending:
		return r.handlePending(ctx, volsyncUnlock)
	case volsyncv1alpha1.VolSyncUnlockPhaseRunning:
		return r.handleRunning(ctx, volsyncUnlock)
	case volsyncv1alpha1.VolSyncUnlockPhaseSucceeded, volsyncv1alpha1.VolSyncUnlockPhaseFailed:
		// Terminal states - requeue after some time for cleanup
		return ctrl.Result{RequeueAfter: time.Minute * 5}, nil
	}

	return ctrl.Result{}, nil
}

func (r *VolSyncUnlockReconciler) handlePending(ctx context.Context, volsyncUnlock *volsyncv1alpha1.VolSyncUnlock) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Update status to pending if not already set
	if volsyncUnlock.Status.Phase != volsyncv1alpha1.VolSyncUnlockPhasePending {
		volsyncUnlock.Status.Phase = volsyncv1alpha1.VolSyncUnlockPhasePending
		volsyncUnlock.Status.Message = "Preparing unlock operation"
		volsyncUnlock.Status.StartTime = &metav1.Time{Time: time.Now()}

		if err := r.Status().Update(ctx, volsyncUnlock); err != nil {
			logger.Error(err, "Failed to update status to pending")
			return ctrl.Result{}, err
		}
	}

	// Create the unlock job
	job, err := r.createUnlockJob(ctx, volsyncUnlock)
	if err != nil {
		logger.Error(err, "Failed to create unlock job")
		volsyncUnlock.Status.Phase = volsyncv1alpha1.VolSyncUnlockPhaseFailed
		volsyncUnlock.Status.Message = fmt.Sprintf("Failed to create unlock job: %v", err)
		if updateErr := r.Status().Update(ctx, volsyncUnlock); updateErr != nil {
			logger.Error(updateErr, "Failed to update status after job creation failure")
		}
		return ctrl.Result{}, err
	}

	// Update status to running
	volsyncUnlock.Status.Phase = volsyncv1alpha1.VolSyncUnlockPhaseRunning
	volsyncUnlock.Status.Message = "Unlock job is running"
	volsyncUnlock.Status.JobName = job.Name

	if err := r.Status().Update(ctx, volsyncUnlock); err != nil {
		logger.Error(err, "Failed to update status to running")
		return ctrl.Result{}, err
	}

	logger.Info("Created unlock job", "job", job.Name)
	return ctrl.Result{RequeueAfter: time.Second * 30}, nil
}

func (r *VolSyncUnlockReconciler) handleRunning(ctx context.Context, volsyncUnlock *volsyncv1alpha1.VolSyncUnlock) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Check job status
	job := &batchv1.Job{}
	err := r.Get(ctx, types.NamespacedName{
		Name:      volsyncUnlock.Status.JobName,
		Namespace: volsyncUnlock.Namespace,
	}, job)

	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Job not found, marking as failed")
			volsyncUnlock.Status.Phase = volsyncv1alpha1.VolSyncUnlockPhaseFailed
			volsyncUnlock.Status.Message = "Unlock job was deleted or not found"
			volsyncUnlock.Status.CompletionTime = &metav1.Time{Time: time.Now()}
			return ctrl.Result{}, r.Status().Update(ctx, volsyncUnlock)
		}
		return ctrl.Result{}, err
	}

	// Check job completion
	if job.Status.Succeeded > 0 {
		volsyncUnlock.Status.Phase = volsyncv1alpha1.VolSyncUnlockPhaseSucceeded
		volsyncUnlock.Status.Message = "Unlock operation completed successfully"
		volsyncUnlock.Status.CompletionTime = &metav1.Time{Time: time.Now()}
		logger.Info("Unlock job succeeded")
	} else if job.Status.Failed > 0 {
		volsyncUnlock.Status.Phase = volsyncv1alpha1.VolSyncUnlockPhaseFailed
		volsyncUnlock.Status.Message = "Unlock operation failed"
		volsyncUnlock.Status.CompletionTime = &metav1.Time{Time: time.Now()}
		logger.Info("Unlock job failed")
	} else {
		// Job is still running
		return ctrl.Result{RequeueAfter: time.Second * 30}, nil
	}

	return ctrl.Result{}, r.Status().Update(ctx, volsyncUnlock)
}

func (r *VolSyncUnlockReconciler) handleDeletion(ctx context.Context, volsyncUnlock *volsyncv1alpha1.VolSyncUnlock) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Clean up the job if it exists
	if volsyncUnlock.Status.JobName != "" {
		job := &batchv1.Job{}
		err := r.Get(ctx, types.NamespacedName{
			Name:      volsyncUnlock.Status.JobName,
			Namespace: volsyncUnlock.Namespace,
		}, job)

		if err == nil {
			logger.Info("Deleting unlock job", "job", job.Name)
			if err := r.Delete(ctx, job); err != nil && !errors.IsNotFound(err) {
				logger.Error(err, "Failed to delete unlock job")
				return ctrl.Result{}, err
			}
		}
	}

	// Remove finalizer
	controllerutil.RemoveFinalizer(volsyncUnlock, "homelab.rafaribe.com/finalizer")
	return ctrl.Result{}, r.Update(ctx, volsyncUnlock)
}

func (r *VolSyncUnlockReconciler) createUnlockJob(ctx context.Context, volsyncUnlock *volsyncv1alpha1.VolSyncUnlock) (*batchv1.Job, error) {
	logger := log.FromContext(ctx)

	// Determine repository secret name
	secretName := volsyncUnlock.Spec.RepositorySecret
	if secretName == "" {
		// Try to find secret by app name
		secretList := &corev1.SecretList{}
		err := r.List(ctx, secretList, client.InNamespace(volsyncUnlock.Spec.Namespace), client.MatchingLabels{
			"app.kubernetes.io/name": volsyncUnlock.Spec.AppName,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list secrets: %w", err)
		}

		if len(secretList.Items) == 0 {
			return nil, fmt.Errorf("no repository secret found for app %s", volsyncUnlock.Spec.AppName)
		}

		secretName = secretList.Items[0].Name
		logger.Info("Found repository secret", "secret", secretName)
	}

	// Set default TTL if not specified
	ttl := int32(300) // 5 minutes
	if volsyncUnlock.Spec.TTLSecondsAfterFinished != nil {
		ttl = *volsyncUnlock.Spec.TTLSecondsAfterFinished
	}

	jobName := fmt.Sprintf("volsync-unlock-%s-%s", volsyncUnlock.Spec.AppName, volsyncUnlock.Name)

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: volsyncUnlock.Spec.Namespace,
			Labels: map[string]string{
				"app":                            "volsync-unlock",
				"homelab.rafaribe.com/app":       volsyncUnlock.Spec.AppName,
				"homelab.rafaribe.com/object":    volsyncUnlock.Spec.ObjectName,
				"homelab.rafaribe.com/unlock-cr": volsyncUnlock.Name,
			},
		},
		Spec: batchv1.JobSpec{
			TTLSecondsAfterFinished: &ttl,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyNever,
					InitContainers: []corev1.Container{
						{
							Name:  "install-kubectl",
							Image: "alpine/k8s:1.28.4",
							Command: []string{
								"/bin/sh", "-c",
								"cp /usr/bin/kubectl /shared/kubectl && chmod +x /shared/kubectl",
							},
							VolumeMounts: []corev1.VolumeMount{
								{Name: "kubectl-binary", MountPath: "/shared"},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:    "restic-unlock",
							Image:   "restic/restic:latest",
							Command: []string{"/bin/sh", "-c"},
							Args: []string{
								fmt.Sprintf(`
echo "Starting VolSync unlock for %s/%s"
echo "Using repository secret: %s"

# Extract repository information from secret
REPO_URL=$(kubectl get secret %s -n %s -o jsonpath='{.data.RESTIC_REPOSITORY}' | base64 -d)
REPO_PASSWORD=$(kubectl get secret %s -n %s -o jsonpath='{.data.RESTIC_PASSWORD}' | base64 -d)

if [ -z "$REPO_URL" ] || [ -z "$REPO_PASSWORD" ]; then
    echo "Repository URL or password not found in secret %s"
    exit 1
fi

export RESTIC_REPOSITORY="$REPO_URL"
export RESTIC_PASSWORD="$REPO_PASSWORD"

echo "Attempting to unlock repository: $RESTIC_REPOSITORY"
if [ "%t" = "true" ]; then
    echo "Force unlock enabled"
    restic unlock --remove-all
else
    restic unlock
fi

if [ $? -eq 0 ]; then
    echo "Successfully unlocked repository for %s/%s"
else
    echo "Failed to unlock repository for %s/%s"
    exit 1
fi
`,
									volsyncUnlock.Spec.AppName, volsyncUnlock.Spec.ObjectName,
									secretName,
									secretName, volsyncUnlock.Spec.Namespace,
									secretName, volsyncUnlock.Spec.Namespace,
									secretName,
									volsyncUnlock.Spec.ForceUnlock,
									volsyncUnlock.Spec.AppName, volsyncUnlock.Spec.ObjectName,
									volsyncUnlock.Spec.AppName, volsyncUnlock.Spec.ObjectName,
								),
							},
							VolumeMounts: []corev1.VolumeMount{
								{Name: "kubectl-binary", MountPath: "/usr/local/bin/kubectl", SubPath: "kubectl"},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "kubectl-binary",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
					},
				},
			},
		},
	}

	// Set owner reference
	if err := controllerutil.SetControllerReference(volsyncUnlock, job, r.Scheme); err != nil {
		return nil, fmt.Errorf("failed to set controller reference: %w", err)
	}

	// Create the job
	if err := r.Create(ctx, job); err != nil {
		return nil, fmt.Errorf("failed to create job: %w", err)
	}

	return job, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *VolSyncUnlockReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&volsyncv1alpha1.VolSyncUnlock{}).
		Owns(&batchv1.Job{}).
		Complete(r)
}
