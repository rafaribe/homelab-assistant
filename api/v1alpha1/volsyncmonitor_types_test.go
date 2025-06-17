package v1alpha1

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestVolSyncMonitorSpec_Defaults(t *testing.T) {
	spec := VolSyncMonitorSpec{
		Enabled:              true,
		MaxConcurrentUnlocks: 3,
		UnlockJobTemplate: UnlockJobTemplate{
			Image: "test-image",
		},
	}

	if !spec.Enabled {
		t.Errorf("Expected Enabled to be true, got %v", spec.Enabled)
	}

	if spec.MaxConcurrentUnlocks != 3 {
		t.Errorf("Expected MaxConcurrentUnlocks to be 3, got %v", spec.MaxConcurrentUnlocks)
	}

	if spec.UnlockJobTemplate.Image != "test-image" {
		t.Errorf("Expected Image to be 'test-image', got %v", spec.UnlockJobTemplate.Image)
	}
}

func TestVolSyncMonitorStatus_Phases(t *testing.T) {
	tests := []struct {
		name     string
		phase    VolSyncMonitorPhase
		expected string
	}{
		{"Active phase", VolSyncMonitorPhaseActive, "Active"},
		{"Paused phase", VolSyncMonitorPhasePaused, "Paused"},
		{"Error phase", VolSyncMonitorPhaseError, "Error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.phase) != tt.expected {
				t.Errorf("Expected phase %s, got %s", tt.expected, string(tt.phase))
			}
		})
	}
}

func TestActiveUnlock_Fields(t *testing.T) {
	now := metav1.Now()
	unlock := ActiveUnlock{
		AppName:          "test-app",
		Namespace:        "test-namespace",
		ObjectName:       "test-object",
		JobName:          "test-job",
		StartTime:        now,
		AlertFingerprint: "test-fingerprint",
	}

	if unlock.AppName != "test-app" {
		t.Errorf("Expected AppName to be 'test-app', got %v", unlock.AppName)
	}

	if unlock.Namespace != "test-namespace" {
		t.Errorf("Expected Namespace to be 'test-namespace', got %v", unlock.Namespace)
	}

	if unlock.ObjectName != "test-object" {
		t.Errorf("Expected ObjectName to be 'test-object', got %v", unlock.ObjectName)
	}

	if unlock.JobName != "test-job" {
		t.Errorf("Expected JobName to be 'test-job', got %v", unlock.JobName)
	}

	if unlock.AlertFingerprint != "test-fingerprint" {
		t.Errorf("Expected AlertFingerprint to be 'test-fingerprint', got %v", unlock.AlertFingerprint)
	}
}

func TestUnlockJobTemplate_OptionalFields(t *testing.T) {
	template := UnlockJobTemplate{
		Image:   "test-image",
		Command: []string{"restic"},
		Args:    []string{"unlock", "--remove-all"},
		Resources: &ResourceRequirements{
			Limits: map[string]string{
				"cpu":    "500m",
				"memory": "512Mi",
			},
			Requests: map[string]string{
				"cpu":    "100m",
				"memory": "128Mi",
			},
		},
		ServiceAccount: "test-sa",
		SecurityContext: &SecurityContext{
			RunAsUser:  &[]int64{1000}[0],
			RunAsGroup: &[]int64{1000}[0],
			FSGroup:    &[]int64{1000}[0],
		},
	}

	if len(template.Command) != 1 || template.Command[0] != "restic" {
		t.Errorf("Expected Command to be ['restic'], got %v", template.Command)
	}

	if len(template.Args) != 2 || template.Args[0] != "unlock" {
		t.Errorf("Expected Args to start with 'unlock', got %v", template.Args)
	}

	if template.Resources.Limits["cpu"] != "500m" {
		t.Errorf("Expected CPU limit to be '500m', got %v", template.Resources.Limits["cpu"])
	}

	if template.ServiceAccount != "test-sa" {
		t.Errorf("Expected ServiceAccount to be 'test-sa', got %v", template.ServiceAccount)
	}

	if *template.SecurityContext.RunAsUser != 1000 {
		t.Errorf("Expected RunAsUser to be 1000, got %v", *template.SecurityContext.RunAsUser)
	}
}

func TestResourceRequirements_EmptyValues(t *testing.T) {
	resources := ResourceRequirements{}

	if resources.Limits != nil {
		t.Errorf("Expected Limits to be nil, got %v", resources.Limits)
	}

	if resources.Requests != nil {
		t.Errorf("Expected Requests to be nil, got %v", resources.Requests)
	}
}

func TestSecurityContext_EmptyValues(t *testing.T) {
	securityContext := SecurityContext{}

	if securityContext.RunAsUser != nil {
		t.Errorf("Expected RunAsUser to be nil, got %v", securityContext.RunAsUser)
	}

	if securityContext.RunAsGroup != nil {
		t.Errorf("Expected RunAsGroup to be nil, got %v", securityContext.RunAsGroup)
	}

	if securityContext.FSGroup != nil {
		t.Errorf("Expected FSGroup to be nil, got %v", securityContext.FSGroup)
	}
}

func TestVolSyncMonitor_ObjectMeta(t *testing.T) {
	monitor := VolSyncMonitor{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "homelab.rafaribe.com/v1alpha1",
			Kind:       "VolSyncMonitor",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-monitor",
			Namespace: "test-namespace",
		},
		Spec: VolSyncMonitorSpec{
			Enabled: true,
			UnlockJobTemplate: UnlockJobTemplate{
				Image: "test-image",
			},
		},
	}

	if monitor.Name != "test-monitor" {
		t.Errorf("Expected Name to be 'test-monitor', got %v", monitor.Name)
	}

	if monitor.Namespace != "test-namespace" {
		t.Errorf("Expected Namespace to be 'test-namespace', got %v", monitor.Namespace)
	}

	if monitor.APIVersion != "homelab.rafaribe.com/v1alpha1" {
		t.Errorf("Expected APIVersion to be 'homelab.rafaribe.com/v1alpha1', got %v", monitor.APIVersion)
	}

	if monitor.Kind != "VolSyncMonitor" {
		t.Errorf("Expected Kind to be 'VolSyncMonitor', got %v", monitor.Kind)
	}
}

func TestVolSyncMonitorList_Items(t *testing.T) {
	list := VolSyncMonitorList{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "homelab.rafaribe.com/v1alpha1",
			Kind:       "VolSyncMonitorList",
		},
		Items: []VolSyncMonitor{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "monitor1",
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "monitor2",
				},
			},
		},
	}

	if len(list.Items) != 2 {
		t.Errorf("Expected 2 items, got %v", len(list.Items))
	}

	if list.Items[0].Name != "monitor1" {
		t.Errorf("Expected first item name to be 'monitor1', got %v", list.Items[0].Name)
	}

	if list.Items[1].Name != "monitor2" {
		t.Errorf("Expected second item name to be 'monitor2', got %v", list.Items[1].Name)
	}
}

func TestVolSyncMonitorStatus_WithActiveUnlocks(t *testing.T) {
	now := metav1.Now()
	status := VolSyncMonitorStatus{
		ActiveUnlocks: []ActiveUnlock{
			{
				AppName:          "app1",
				Namespace:        "ns1",
				ObjectName:       "obj1",
				JobName:          "job1",
				StartTime:        now,
				AlertFingerprint: "fp1",
			},
			{
				AppName:          "app2",
				Namespace:        "ns2",
				ObjectName:       "obj2",
				JobName:          "job2",
				StartTime:        now,
				AlertFingerprint: "fp2",
			},
		},
		TotalUnlocksCreated: 5,
		LastError:           "test error",
		Phase:               VolSyncMonitorPhaseActive,
	}

	if len(status.ActiveUnlocks) != 2 {
		t.Errorf("Expected 2 active unlocks, got %v", len(status.ActiveUnlocks))
	}

	if status.TotalUnlocksCreated != 5 {
		t.Errorf("Expected TotalUnlocksCreated to be 5, got %v", status.TotalUnlocksCreated)
	}

	if status.LastError != "test error" {
		t.Errorf("Expected LastError to be 'test error', got %v", status.LastError)
	}

	if status.Phase != VolSyncMonitorPhaseActive {
		t.Errorf("Expected Phase to be Active, got %v", status.Phase)
	}
}

func TestVolSyncMonitor_DeepCopy(t *testing.T) {
	original := &VolSyncMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-monitor",
			Namespace: "test-namespace",
		},
		Spec: VolSyncMonitorSpec{
			Enabled: true,
			UnlockJobTemplate: UnlockJobTemplate{
				Image: "test-image",
				Command: []string{"restic"},
				Args:    []string{"unlock"},
			},
		},
		Status: VolSyncMonitorStatus{
			Phase: VolSyncMonitorPhaseActive,
		},
	}

	copied := original.DeepCopy()

	if copied.Name != original.Name {
		t.Errorf("DeepCopy failed: Name mismatch")
	}

	if copied.Spec.UnlockJobTemplate.Image != original.Spec.UnlockJobTemplate.Image {
		t.Errorf("DeepCopy failed: Image mismatch")
	}

	// Modify original to ensure deep copy
	original.Spec.UnlockJobTemplate.Command[0] = "modified"
	if copied.Spec.UnlockJobTemplate.Command[0] == "modified" {
		t.Errorf("DeepCopy failed: Command was not deeply copied")
	}
}

func TestVolSyncMonitor_DeepCopyObject(t *testing.T) {
	original := &VolSyncMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-monitor",
		},
	}

	copied := original.DeepCopyObject()
	copiedMonitor, ok := copied.(*VolSyncMonitor)
	if !ok {
		t.Errorf("DeepCopyObject failed: returned object is not *VolSyncMonitor")
	}

	if copiedMonitor.Name != original.Name {
		t.Errorf("DeepCopyObject failed: Name mismatch")
	}
}

func TestVolSyncMonitorList_DeepCopy(t *testing.T) {
	original := &VolSyncMonitorList{
		Items: []VolSyncMonitor{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "monitor1",
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

func TestVolSyncMonitorList_DeepCopyObject(t *testing.T) {
	original := &VolSyncMonitorList{
		Items: []VolSyncMonitor{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "monitor1",
				},
			},
		},
	}

	copied := original.DeepCopyObject()
	copiedList, ok := copied.(*VolSyncMonitorList)
	if !ok {
		t.Errorf("DeepCopyObject failed: returned object is not *VolSyncMonitorList")
	}

	if len(copiedList.Items) != len(original.Items) {
		t.Errorf("DeepCopyObject failed: Items length mismatch")
	}
}

func TestSchemeBuilder_Registration(t *testing.T) {
	scheme := runtime.NewScheme()
	err := AddToScheme(scheme)
	if err != nil {
		t.Errorf("Failed to add to scheme: %v", err)
	}

	// Verify that our types are registered
	gvk := GroupVersion.WithKind("VolSyncMonitor")
	obj, err := scheme.New(gvk)
	if err != nil {
		t.Errorf("Failed to create VolSyncMonitor from scheme: %v", err)
	}

	if _, ok := obj.(*VolSyncMonitor); !ok {
		t.Errorf("Created object is not a VolSyncMonitor")
	}

	gvkList := GroupVersion.WithKind("VolSyncMonitorList")
	objList, err := scheme.New(gvkList)
	if err != nil {
		t.Errorf("Failed to create VolSyncMonitorList from scheme: %v", err)
	}

	if _, ok := objList.(*VolSyncMonitorList); !ok {
		t.Errorf("Created object is not a VolSyncMonitorList")
	}
}
