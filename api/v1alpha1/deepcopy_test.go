package v1alpha1

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	testModifiedValue = "modified"
)

func TestResourceRequirements_DeepCopy(t *testing.T) {
	original := &ResourceRequirements{
		Limits: map[string]string{
			"cpu":    "500m",
			"memory": "512Mi",
		},
		Requests: map[string]string{
			"cpu":    "100m",
			"memory": "128Mi",
		},
	}

	copied := original.DeepCopy()

	if copied == nil {
		t.Error("DeepCopy returned nil")
		return
	}

	if len(copied.Limits) != len(original.Limits) {
		t.Error("DeepCopy failed: Limits length mismatch")
	}

	if len(copied.Requests) != len(original.Requests) {
		t.Error("DeepCopy failed: Requests length mismatch")
	}

	// Modify original to ensure deep copy
	original.Limits["cpu"] = "1000m"
	if copied.Limits["cpu"] == "1000m" {
		t.Error("DeepCopy failed: Limits was not deeply copied")
	}
}

func TestSecurityContext_DeepCopy(t *testing.T) {
	uid := int64(1000)
	gid := int64(1000)
	fsg := int64(1000)

	original := &SecurityContext{
		RunAsUser:  &uid,
		RunAsGroup: &gid,
		FSGroup:    &fsg,
	}

	copied := original.DeepCopy()

	if copied == nil {
		t.Error("DeepCopy returned nil")
		return
	}

	if copied.RunAsUser == nil || *copied.RunAsUser != *original.RunAsUser {
		t.Error("DeepCopy failed: RunAsUser mismatch")
	}

	if copied.RunAsGroup == nil || *copied.RunAsGroup != *original.RunAsGroup {
		t.Error("DeepCopy failed: RunAsGroup mismatch")
	}

	if copied.FSGroup == nil || *copied.FSGroup != *original.FSGroup {
		t.Error("DeepCopy failed: FSGroup mismatch")
	}

	// Modify original to ensure deep copy
	*original.RunAsUser = 2000
	if *copied.RunAsUser == 2000 {
		t.Error("DeepCopy failed: RunAsUser was not deeply copied")
	}
}

func TestUnlockJobTemplate_DeepCopy(t *testing.T) {
	uid := int64(1000)
	original := &UnlockJobTemplate{
		Image:   "test-image",
		Command: []string{"restic"},
		Args:    []string{"unlock", "--remove-all"},
		Resources: &ResourceRequirements{
			Limits: map[string]string{
				"cpu": "500m",
			},
		},
		SecurityContext: &SecurityContext{
			RunAsUser: &uid,
		},
		ServiceAccount: "test-sa",
	}

	copied := original.DeepCopy()

	if copied == nil {
		t.Error("DeepCopy returned nil")
		return
	}

	if copied.Image != original.Image {
		t.Error("DeepCopy failed: Image mismatch")
	}

	if len(copied.Command) != len(original.Command) {
		t.Error("DeepCopy failed: Command length mismatch")
	}

	if len(copied.Args) != len(original.Args) {
		t.Error("DeepCopy failed: Args length mismatch")
	}

	// Modify original to ensure deep copy
	original.Command[0] = testModifiedValue
	if copied.Command[0] == testModifiedValue {
		t.Error("DeepCopy failed: Command was not deeply copied")
	}
}

func TestVolSyncMonitorStatus_DeepCopy(t *testing.T) {
	now := metav1.Now()
	original := &VolSyncMonitorStatus{
		ActiveUnlocks: []ActiveUnlock{
			{
				AppName:          "test-app",
				Namespace:        "test-ns",
				ObjectName:       "test-obj",
				JobName:          "test-job",
				StartTime:        now,
				AlertFingerprint: "test-fp",
			},
		},
		TotalUnlocksCreated: 5,
		LastError:           "test error",
		Phase:               VolSyncMonitorPhaseActive,
		Conditions: []metav1.Condition{
			{
				Type:   "Ready",
				Status: metav1.ConditionTrue,
				Reason: "Available",
			},
		},
	}

	copied := original.DeepCopy()

	if copied == nil {
		t.Error("DeepCopy returned nil")
		return
	}

	if len(copied.ActiveUnlocks) != len(original.ActiveUnlocks) {
		t.Error("DeepCopy failed: ActiveUnlocks length mismatch")
	}

	if copied.TotalUnlocksCreated != original.TotalUnlocksCreated {
		t.Error("DeepCopy failed: TotalUnlocksCreated mismatch")
	}

	if len(copied.Conditions) != len(original.Conditions) {
		t.Error("DeepCopy failed: Conditions length mismatch")
	}

	// Modify original to ensure deep copy
	original.ActiveUnlocks[0].AppName = testModifiedValue
	if copied.ActiveUnlocks[0].AppName == testModifiedValue {
		t.Error("DeepCopy failed: ActiveUnlocks was not deeply copied")
	}
}

func TestActiveUnlock_DeepCopy(t *testing.T) {
	now := metav1.Now()
	original := &ActiveUnlock{
		AppName:          "test-app",
		Namespace:        "test-ns",
		ObjectName:       "test-obj",
		JobName:          "test-job",
		StartTime:        now,
		AlertFingerprint: "test-fp",
	}

	copied := original.DeepCopy()

	if copied == nil {
		t.Error("DeepCopy returned nil")
		return
	}

	if copied.AppName != original.AppName {
		t.Error("DeepCopy failed: AppName mismatch")
	}

	if copied.StartTime != original.StartTime {
		t.Error("DeepCopy failed: StartTime mismatch")
	}
}
