package v1alpha1

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestVolSyncUnlockSpec_Fields(t *testing.T) {
	spec := VolSyncUnlockSpec{
		AppName:          "test-app",
		Namespace:        "test-namespace",
		ObjectName:       "test-object",
		RepositorySecret: "test-secret",
		ForceUnlock:      true,
	}

	if spec.AppName != "test-app" {
		t.Errorf("Expected AppName to be 'test-app', got %v", spec.AppName)
	}

	if spec.Namespace != "test-namespace" {
		t.Errorf("Expected Namespace to be 'test-namespace', got %v", spec.Namespace)
	}

	if spec.ObjectName != "test-object" {
		t.Errorf("Expected ObjectName to be 'test-object', got %v", spec.ObjectName)
	}

	if spec.RepositorySecret != "test-secret" {
		t.Errorf("Expected RepositorySecret to be 'test-secret', got %v", spec.RepositorySecret)
	}

	if !spec.ForceUnlock {
		t.Errorf("Expected ForceUnlock to be true, got %v", spec.ForceUnlock)
	}
}

func TestVolSyncUnlockSpec_OptionalFields(t *testing.T) {
	spec := VolSyncUnlockSpec{
		AppName:    "test-app",
		Namespace:  "test-namespace",
		ObjectName: "test-object",
		// Optional fields not set
	}

	if spec.RepositorySecret != "" {
		t.Errorf("Expected RepositorySecret to be empty, got %v", spec.RepositorySecret)
	}

	if spec.ForceUnlock {
		t.Errorf("Expected ForceUnlock to be false, got %v", spec.ForceUnlock)
	}

	if spec.TTLSecondsAfterFinished != nil {
		t.Errorf("Expected TTLSecondsAfterFinished to be nil, got %v", spec.TTLSecondsAfterFinished)
	}
}

func TestVolSyncUnlockSpec_WithTTL(t *testing.T) {
	ttl := int32(3600)
	spec := VolSyncUnlockSpec{
		AppName:                 "test-app",
		Namespace:               "test-namespace",
		ObjectName:              "test-object",
		TTLSecondsAfterFinished: &ttl,
	}

	if spec.TTLSecondsAfterFinished == nil {
		t.Error("Expected TTLSecondsAfterFinished to be set")
	}

	if *spec.TTLSecondsAfterFinished != 3600 {
		t.Errorf("Expected TTLSecondsAfterFinished to be 3600, got %v", *spec.TTLSecondsAfterFinished)
	}
}

func TestVolSyncUnlockStatus_Phases(t *testing.T) {
	tests := []struct {
		name     string
		phase    VolSyncUnlockPhase
		expected string
	}{
		{"Pending phase", VolSyncUnlockPhasePending, "Pending"},
		{"Running phase", VolSyncUnlockPhaseRunning, "Running"},
		{"Succeeded phase", VolSyncUnlockPhaseSucceeded, "Succeeded"},
		{"Failed phase", VolSyncUnlockPhaseFailed, "Failed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.phase) != tt.expected {
				t.Errorf("Expected phase %s, got %s", tt.expected, string(tt.phase))
			}
		})
	}
}

func TestVolSyncUnlockStatus_WithJobInfo(t *testing.T) {
	now := metav1.Now()
	status := VolSyncUnlockStatus{
		Phase:          VolSyncUnlockPhaseRunning,
		JobName:        "test-job",
		StartTime:      &now,
		CompletionTime: nil,
		Message:        "Unlock in progress",
	}

	if status.Phase != VolSyncUnlockPhaseRunning {
		t.Errorf("Expected Phase to be Running, got %v", status.Phase)
	}

	if status.JobName != "test-job" {
		t.Errorf("Expected JobName to be 'test-job', got %v", status.JobName)
	}

	if status.StartTime == nil {
		t.Errorf("Expected StartTime to be set")
	}

	if status.CompletionTime != nil {
		t.Errorf("Expected CompletionTime to be nil for running job")
	}

	if status.Message != "Unlock in progress" {
		t.Errorf("Expected Message to be 'Unlock in progress', got %v", status.Message)
	}
}

func TestVolSyncUnlock_ObjectMeta(t *testing.T) {
	unlock := VolSyncUnlock{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "homelab.rafaribe.com/v1alpha1",
			Kind:       "VolSyncUnlock",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-unlock",
			Namespace: "test-namespace",
		},
		Spec: VolSyncUnlockSpec{
			AppName:    "test-app",
			Namespace:  "test-namespace",
			ObjectName: "test-object",
		},
	}

	if unlock.Name != "test-unlock" {
		t.Errorf("Expected Name to be 'test-unlock', got %v", unlock.Name)
	}

	if unlock.Namespace != "test-namespace" {
		t.Errorf("Expected Namespace to be 'test-namespace', got %v", unlock.Namespace)
	}

	if unlock.APIVersion != "homelab.rafaribe.com/v1alpha1" {
		t.Errorf("Expected APIVersion to be 'homelab.rafaribe.com/v1alpha1', got %v", unlock.APIVersion)
	}

	if unlock.Kind != "VolSyncUnlock" {
		t.Errorf("Expected Kind to be 'VolSyncUnlock', got %v", unlock.Kind)
	}
}

func TestVolSyncUnlockList_Items(t *testing.T) {
	list := VolSyncUnlockList{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "homelab.rafaribe.com/v1alpha1",
			Kind:       "VolSyncUnlockList",
		},
		Items: []VolSyncUnlock{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "unlock1",
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "unlock2",
				},
			},
		},
	}

	if len(list.Items) != 2 {
		t.Errorf("Expected 2 items, got %v", len(list.Items))
	}

	if list.Items[0].Name != "unlock1" {
		t.Errorf("Expected first item name to be 'unlock1', got %v", list.Items[0].Name)
	}

	if list.Items[1].Name != "unlock2" {
		t.Errorf("Expected second item name to be 'unlock2', got %v", list.Items[1].Name)
	}
}

func TestVolSyncUnlock_DeepCopy(t *testing.T) {
	ttl := int32(3600)
	original := &VolSyncUnlock{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-unlock",
			Namespace: "test-namespace",
		},
		Spec: VolSyncUnlockSpec{
			AppName:                 "test-app",
			Namespace:               "test-namespace",
			ObjectName:              "test-object",
			TTLSecondsAfterFinished: &ttl,
		},
		Status: VolSyncUnlockStatus{
			Phase: VolSyncUnlockPhaseRunning,
		},
	}

	copied := original.DeepCopy()

	if copied.Name != original.Name {
		t.Errorf("DeepCopy failed: Name mismatch")
	}

	if copied.Spec.AppName != original.Spec.AppName {
		t.Errorf("DeepCopy failed: AppName mismatch")
	}

	// Modify original to ensure deep copy
	*original.Spec.TTLSecondsAfterFinished = 7200
	if *copied.Spec.TTLSecondsAfterFinished == 7200 {
		t.Errorf("DeepCopy failed: TTL was not deeply copied")
	}
}

func TestVolSyncUnlock_DeepCopyObject(t *testing.T) {
	original := &VolSyncUnlock{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-unlock",
		},
	}

	copied := original.DeepCopyObject()
	copiedUnlock, ok := copied.(*VolSyncUnlock)
	if !ok {
		t.Errorf("DeepCopyObject failed: returned object is not *VolSyncUnlock")
	}

	if copiedUnlock.Name != original.Name {
		t.Errorf("DeepCopyObject failed: Name mismatch")
	}
}

func TestVolSyncUnlockList_DeepCopy(t *testing.T) {
	original := &VolSyncUnlockList{
		Items: []VolSyncUnlock{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "unlock1",
				},
			},
		},
	}

	copied := original.DeepCopy()

	if len(copied.Items) != len(original.Items) {
		t.Errorf("DeepCopy failed: Items length mismatch")
	}

	if copied.Items[0].Name != original.Items[0].Name {
		t.Errorf("DeepCopy failed: Item name mismatch")
	}
}

func TestVolSyncUnlockList_DeepCopyObject(t *testing.T) {
	original := &VolSyncUnlockList{
		Items: []VolSyncUnlock{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "unlock1",
				},
			},
		},
	}

	copied := original.DeepCopyObject()
	copiedList, ok := copied.(*VolSyncUnlockList)
	if !ok {
		t.Errorf("DeepCopyObject failed: returned object is not *VolSyncUnlockList")
	}

	if len(copiedList.Items) != len(original.Items) {
		t.Errorf("DeepCopyObject failed: Items length mismatch")
	}
}

func TestVolSyncUnlockStatus_CompletedJob(t *testing.T) {
	startTime := metav1.Now()
	completionTime := metav1.Now()

	status := VolSyncUnlockStatus{
		Phase:          VolSyncUnlockPhaseSucceeded,
		JobName:        "completed-job",
		StartTime:      &startTime,
		CompletionTime: &completionTime,
		Message:        "Unlock completed successfully",
	}

	if status.Phase != VolSyncUnlockPhaseSucceeded {
		t.Errorf("Expected Phase to be Succeeded, got %v", status.Phase)
	}

	if status.StartTime == nil {
		t.Errorf("Expected StartTime to be set")
	}

	if status.CompletionTime == nil {
		t.Errorf("Expected CompletionTime to be set for completed job")
	}

	if status.Message != "Unlock completed successfully" {
		t.Errorf("Expected success message, got %v", status.Message)
	}
}

func TestVolSyncUnlock_SchemeRegistration(t *testing.T) {
	scheme := runtime.NewScheme()
	err := AddToScheme(scheme)
	if err != nil {
		t.Errorf("Failed to add to scheme: %v", err)
	}

	// Verify that VolSyncUnlock types are registered
	gvk := GroupVersion.WithKind("VolSyncUnlock")
	obj, err := scheme.New(gvk)
	if err != nil {
		t.Errorf("Failed to create VolSyncUnlock from scheme: %v", err)
	}

	if _, ok := obj.(*VolSyncUnlock); !ok {
		t.Errorf("Created object is not a VolSyncUnlock")
	}

	gvkList := GroupVersion.WithKind("VolSyncUnlockList")
	objList, err := scheme.New(gvkList)
	if err != nil {
		t.Errorf("Failed to create VolSyncUnlockList from scheme: %v", err)
	}

	if _, ok := objList.(*VolSyncUnlockList); !ok {
		t.Errorf("Created object is not a VolSyncUnlockList")
	}
}
