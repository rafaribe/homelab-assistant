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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// VolSyncUnlockSpec defines the desired state of VolSyncUnlock
type VolSyncUnlockSpec struct {
	// AppName is the name of the application that owns the VolSync resource
	AppName string `json:"appName"`

	// Namespace is the namespace where the VolSync resource is located
	Namespace string `json:"namespace"`

	// ObjectName is the name of the VolSync object (e.g., "prowlarr-nfs")
	ObjectName string `json:"objectName"`

	// RepositorySecret is the name of the secret containing restic repository credentials
	// +optional
	RepositorySecret string `json:"repositorySecret,omitempty"`

	// ForceUnlock indicates whether to force unlock even if repository is in use
	// +optional
	ForceUnlock bool `json:"forceUnlock,omitempty"`

	// TTLSecondsAfterFinished specifies the TTL for the unlock job
	// +optional
	TTLSecondsAfterFinished *int32 `json:"ttlSecondsAfterFinished,omitempty"`
}

// VolSyncUnlockStatus defines the observed state of VolSyncUnlock
type VolSyncUnlockStatus struct {
	// Phase represents the current phase of the unlock operation
	// +optional
	Phase VolSyncUnlockPhase `json:"phase,omitempty"`

	// Message provides additional information about the current phase
	// +optional
	Message string `json:"message,omitempty"`

	// JobName is the name of the created unlock job
	// +optional
	JobName string `json:"jobName,omitempty"`

	// StartTime is when the unlock operation started
	// +optional
	StartTime *metav1.Time `json:"startTime,omitempty"`

	// CompletionTime is when the unlock operation completed
	// +optional
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`

	// Conditions represent the latest available observations of the unlock operation
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// VolSyncUnlockPhase represents the phase of a VolSyncUnlock operation
type VolSyncUnlockPhase string

const (
	// VolSyncUnlockPhasePending indicates the unlock operation is pending
	VolSyncUnlockPhasePending VolSyncUnlockPhase = "Pending"
	// VolSyncUnlockPhaseRunning indicates the unlock operation is running
	VolSyncUnlockPhaseRunning VolSyncUnlockPhase = "Running"
	// VolSyncUnlockPhaseSucceeded indicates the unlock operation succeeded
	VolSyncUnlockPhaseSucceeded VolSyncUnlockPhase = "Succeeded"
	// VolSyncUnlockPhaseFailed indicates the unlock operation failed
	VolSyncUnlockPhaseFailed VolSyncUnlockPhase = "Failed"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// VolSyncUnlock is the Schema for the volsyncunlocks API
type VolSyncUnlock struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VolSyncUnlockSpec   `json:"spec,omitempty"`
	Status VolSyncUnlockStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// VolSyncUnlockList contains a list of VolSyncUnlock
type VolSyncUnlockList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VolSyncUnlock `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VolSyncUnlock{}, &VolSyncUnlockList{})
}
