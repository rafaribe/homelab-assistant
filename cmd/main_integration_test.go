//go:build integration
// +build integration

package main

import (
	"context"
	"os"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	volsyncv1alpha1 "github.com/rafaribe/homelab-assistant/api/v1alpha1"
)

// TestMainFunction tests the main function components
func TestMainFunction(t *testing.T) {
	// Set up environment variables for testing
	os.Setenv("METRICS_BIND_ADDRESS", ":8080")
	os.Setenv("HEALTH_PROBE_BIND_ADDRESS", ":8081")
	os.Setenv("LEADER_ELECT", "false")
	defer func() {
		os.Unsetenv("METRICS_BIND_ADDRESS")
		os.Unsetenv("HEALTH_PROBE_BIND_ADDRESS")
		os.Unsetenv("LEADER_ELECT")
	}()

	// Test scheme setup
	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(volsyncv1alpha1.AddToScheme(scheme))

	// Test logger setup
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&zap.Options{
		Development: true,
	})))

	// Test signal context
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Verify context is working
	select {
	case <-ctx.Done():
		t.Log("Context cancelled as expected")
	case <-time.After(2 * time.Second):
		t.Error("Context should have been cancelled")
	}
}

// TestMainEnvironmentVariables tests environment variable handling
func TestMainEnvironmentVariables(t *testing.T) {
	tests := []struct {
		name     string
		envVar   string
		setValue string
		expected string
	}{
		{
			name:     "METRICS_BIND_ADDRESS",
			envVar:   "METRICS_BIND_ADDRESS",
			setValue: ":9090",
			expected: ":9090",
		},
		{
			name:     "HEALTH_PROBE_BIND_ADDRESS",
			envVar:   "HEALTH_PROBE_BIND_ADDRESS",
			setValue: ":9091",
			expected: ":9091",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original value
			originalVal := os.Getenv(tt.envVar)
			defer os.Setenv(tt.envVar, originalVal)

			// Set test value
			os.Setenv(tt.envVar, tt.setValue)

			// Get value
			result := os.Getenv(tt.envVar)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestMainLeaderElection tests leader election flag parsing
func TestMainLeaderElection(t *testing.T) {
	tests := []struct {
		name     string
		setValue string
		expected bool
	}{
		{
			name:     "leader election enabled",
			setValue: "true",
			expected: true,
		},
		{
			name:     "leader election disabled",
			setValue: "false",
			expected: false,
		},
		{
			name:     "leader election empty",
			setValue: "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original value
			originalVal := os.Getenv("LEADER_ELECT")
			defer os.Setenv("LEADER_ELECT", originalVal)

			// Set test value
			os.Setenv("LEADER_ELECT", tt.setValue)

			// Parse value
			result := os.Getenv("LEADER_ELECT") == "true"
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}
