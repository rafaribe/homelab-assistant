/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// VolSyncMonitorSpec defines the desired state of VolSyncMonitor
type VolSyncMonitorSpec struct {
	// UnlockJobTemplate defines the template for unlock jobs
	UnlockJobTemplate UnlockJobTemplate `json:"unlockJobTemplate"`

	// TTLSecondsAfterFinished specifies the TTL for unlock jobs
	// +optional
	TTLSecondsAfterFinished *int32 `json:"ttlSecondsAfterFinished,omitempty"`

	// MaxConcurrentUnlocks limits the number of concurrent unlock operations
	// +optional
	MaxConcurrentUnlocks int32 `json:"maxConcurrentUnlocks,omitempty"`

	// Enabled controls whether the monitor is active
	// +optional
	Enabled bool `json:"enabled,omitempty"`

	// LockErrorPatterns are regex patterns to match in job logs that indicate lock issues
	// Default patterns will be used if not specified
	// +optional
	LockErrorPatterns []string `json:"lockErrorPatterns,omitempty"`
}

// UnlockJobTemplate defines the template for creating unlock jobs
type UnlockJobTemplate struct {
	// Image is the container image to use for unlock jobs
	Image string `json:"image"`

	// Command is the command to run in the unlock job
	// +optional
	Command []string `json:"command,omitempty"`

	// Args are the arguments to pass to the command
	// +optional
	Args []string `json:"args,omitempty"`

	// Resources defines resource requirements for unlock jobs
	// +optional
	Resources *ResourceRequirements `json:"resources,omitempty"`

	// ServiceAccount to use for unlock jobs
	// +optional
	ServiceAccount string `json:"serviceAccount,omitempty"`

	// SecurityContext for unlock jobs
	// +optional
	SecurityContext *SecurityContext `json:"securityContext,omitempty"`
}

// VolSyncMonitorStatus defines the observed state of VolSyncMonitor
type VolSyncMonitorStatus struct {
	// ActiveUnlocks tracks currently running unlock operations
	// +optional
	ActiveUnlocks []ActiveUnlock `json:"activeUnlocks,omitempty"`

	// TotalUnlocksCreated is the total number of unlock jobs created
	// +optional
	TotalUnlocksCreated int32 `json:"totalUnlocksCreated,omitempty"`

	// TotalUnlocksSucceeded is the total number of unlock jobs that succeeded
	// +optional
	TotalUnlocksSucceeded int32 `json:"totalUnlocksSucceeded,omitempty"`

	// TotalUnlocksFailed is the total number of unlock jobs that failed
	// +optional
	TotalUnlocksFailed int32 `json:"totalUnlocksFailed,omitempty"`

	// TotalLockErrorsDetected is the total number of lock errors detected
	// +optional
	TotalLockErrorsDetected int32 `json:"totalLockErrorsDetected,omitempty"`

	// LastUnlockTime is the timestamp of the last unlock operation
	// +optional
	LastUnlockTime *metav1.Time `json:"lastUnlockTime,omitempty"`

	// LastError contains the last error encountered
	// +optional
	LastError string `json:"lastError,omitempty"`

	// Conditions represent the latest available observations
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// Phase represents the current phase of the monitor
	// +optional
	Phase VolSyncMonitorPhase `json:"phase,omitempty"`

	// ObservedGeneration is the last generation observed by the controller
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// ActiveUnlock represents an active unlock operation
type ActiveUnlock struct {
	// AppName is the name of the application
	AppName string `json:"appName"`

	// Namespace is the namespace of the VolSync resource
	Namespace string `json:"namespace"`

	// ObjectName is the name of the VolSync object
	ObjectName string `json:"objectName"`

	// JobName is the name of the unlock job
	JobName string `json:"jobName"`

	// StartTime is when the unlock started
	StartTime metav1.Time `json:"startTime"`

	// AlertFingerprint is the unique identifier of the alert
	AlertFingerprint string `json:"alertFingerprint"`
}

// VolSyncMonitorPhase represents the phase of the monitor
type VolSyncMonitorPhase string

const (
	// VolSyncMonitorPhaseActive indicates the monitor is actively checking
	VolSyncMonitorPhaseActive VolSyncMonitorPhase = "Active"
	// VolSyncMonitorPhasePaused indicates the monitor is paused
	VolSyncMonitorPhasePaused VolSyncMonitorPhase = "Paused"
	// VolSyncMonitorPhaseError indicates the monitor has encountered an error
	VolSyncMonitorPhaseError VolSyncMonitorPhase = "Error"
)

// ResourceRequirements defines resource requirements
type RepositoryMount struct {
	// Type of mount (nfs, pvc, hostPath, etc.)
	Type RepositoryMountType `json:"type"`

	// NFS configuration (when type is "nfs")
	// +optional
	NFS *NFSMount `json:"nfs,omitempty"`

	// PVC configuration (when type is "pvc")
	// +optional
	PVC *PVCMount `json:"pvc,omitempty"`

	// HostPath configuration (when type is "hostPath")
	// +optional
	HostPath *HostPathMount `json:"hostPath,omitempty"`

	// MountPath is where to mount the repository in the container
	// +optional
	MountPath string `json:"mountPath,omitempty"`
}

// RepositoryMountType defines the type of repository mount
type RepositoryMountType string

const (
	// RepositoryMountTypeNFS uses NFS for repository access
	RepositoryMountTypeNFS RepositoryMountType = "nfs"
	// RepositoryMountTypePVC uses a PVC for repository access
	RepositoryMountTypePVC RepositoryMountType = "pvc"
	// RepositoryMountTypeHostPath uses hostPath for repository access
	RepositoryMountTypeHostPath RepositoryMountType = "hostPath"
)

// NFSMount defines NFS mount configuration
type NFSMount struct {
	// Server is the NFS server hostname or IP
	Server string `json:"server"`

	// Path is the path on the NFS server
	Path string `json:"path"`

	// ReadOnly specifies if the mount should be read-only
	// +optional
	ReadOnly bool `json:"readOnly,omitempty"`
}

// PVCMount defines PVC mount configuration
type PVCMount struct {
	// ClaimName is the name of the PVC
	ClaimName string `json:"claimName"`

	// ReadOnly specifies if the mount should be read-only
	// +optional
	ReadOnly bool `json:"readOnly,omitempty"`
}

// HostPathMount defines hostPath mount configuration
type HostPathMount struct {
	// Path on the host
	Path string `json:"path"`

	// Type of hostPath
	// +optional
	Type string `json:"type,omitempty"`

	// ReadOnly specifies if the mount should be read-only
	// +optional
	ReadOnly bool `json:"readOnly,omitempty"`
}

// ResourceRequirements defines resource requirements
// ResourceRequirements defines resource requirements
type ResourceRequirements struct {
	// Limits describes the maximum amount of compute resources allowed
	// +optional
	Limits map[string]string `json:"limits,omitempty"`

	// Requests describes the minimum amount of compute resources required
	// +optional
	Requests map[string]string `json:"requests,omitempty"`
}

// SecurityContext defines security context
type SecurityContext struct {
	// RunAsUser is the UID to run the entrypoint of the container process
	// +optional
	RunAsUser *int64 `json:"runAsUser,omitempty"`

	// RunAsGroup is the GID to run the entrypoint of the container process
	// +optional
	RunAsGroup *int64 `json:"runAsGroup,omitempty"`

	// FSGroup defines a file system group ID for all containers
	// +optional
	FSGroup *int64 `json:"fsGroup,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase"
//+kubebuilder:printcolumn:name="Active Unlocks",type="integer",JSONPath=".status.activeUnlocks"
//+kubebuilder:printcolumn:name="Total Created",type="integer",JSONPath=".status.totalUnlocksCreated"
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// VolSyncMonitor is the Schema for the volsyncmonitors API
type VolSyncMonitor struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VolSyncMonitorSpec   `json:"spec,omitempty"`
	Status VolSyncMonitorStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// VolSyncMonitorList contains a list of VolSyncMonitor
type VolSyncMonitorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VolSyncMonitor `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VolSyncMonitor{}, &VolSyncMonitorList{})
}
