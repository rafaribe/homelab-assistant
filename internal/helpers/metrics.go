package helpers

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	// unlockJobsCreatedTotal tracks the total number of unlock jobs created
	unlockJobsCreatedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "volsync_unlock_jobs_created_total",
			Help: "Total number of VolSync unlock jobs created",
		},
		[]string{"namespace", "app", "object"},
	)

	// unlockJobsSucceededTotal tracks the total number of successful unlock jobs
	unlockJobsSucceededTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "volsync_unlock_jobs_succeeded_total",
			Help: "Total number of VolSync unlock jobs that succeeded",
		},
		[]string{"namespace", "app", "object"},
	)

	// unlockJobsFailedTotal tracks the total number of failed unlock jobs
	unlockJobsFailedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "volsync_unlock_jobs_failed_total",
			Help: "Total number of VolSync unlock jobs that failed",
		},
		[]string{"namespace", "app", "object"},
	)

	// activeUnlockJobs tracks the current number of active unlock jobs
	activeUnlockJobs = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "volsync_active_unlock_jobs",
			Help: "Current number of active VolSync unlock jobs",
		},
		[]string{"namespace", "app", "object"},
	)

	// lockErrorsDetectedTotal tracks the total number of lock errors detected
	lockErrorsDetectedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "volsync_lock_errors_detected_total",
			Help: "Total number of VolSync lock errors detected",
		},
		[]string{"namespace", "app", "object", "error_pattern"},
	)

	// monitorReconciliationsTotal tracks the total number of monitor reconciliations
	monitorReconciliationsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "volsync_monitor_reconciliations_total",
			Help: "Total number of VolSync monitor reconciliations",
		},
		[]string{"namespace", "monitor", "result"},
	)
)

func init() {
	// Register metrics with controller-runtime's metrics registry
	metrics.Registry.MustRegister(
		unlockJobsCreatedTotal,
		unlockJobsSucceededTotal,
		unlockJobsFailedTotal,
		activeUnlockJobs,
		lockErrorsDetectedTotal,
		monitorReconciliationsTotal,
	)
}

// RecordUnlockJobCreated increments the counter for created unlock jobs
func RecordUnlockJobCreated(namespace, app, object string) {
	unlockJobsCreatedTotal.WithLabelValues(namespace, app, object).Inc()
}

// RecordUnlockJobSucceeded increments the counter for successful unlock jobs
func RecordUnlockJobSucceeded(namespace, app, object string) {
	unlockJobsSucceededTotal.WithLabelValues(namespace, app, object).Inc()
	activeUnlockJobs.WithLabelValues(namespace, app, object).Dec()
}

// RecordUnlockJobFailed increments the counter for failed unlock jobs
func RecordUnlockJobFailed(namespace, app, object string) {
	unlockJobsFailedTotal.WithLabelValues(namespace, app, object).Inc()
	activeUnlockJobs.WithLabelValues(namespace, app, object).Dec()
}

// RecordActiveUnlockJob increments the gauge for active unlock jobs
func RecordActiveUnlockJob(namespace, app, object string) {
	activeUnlockJobs.WithLabelValues(namespace, app, object).Inc()
}

// RecordLockErrorDetected increments the counter for detected lock errors
func RecordLockErrorDetected(namespace, app, object, errorPattern string) {
	lockErrorsDetectedTotal.WithLabelValues(namespace, app, object, errorPattern).Inc()
}

// RecordMonitorReconciliation increments the counter for monitor reconciliations
func RecordMonitorReconciliation(namespace, monitor, result string) {
	monitorReconciliationsTotal.WithLabelValues(namespace, monitor, result).Inc()
}
