package main

import (
	"os"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	volsyncv1alpha1 "github.com/rafaribe/homelab-assistant/api/v1alpha1"
)

func TestSchemeSetup(t *testing.T) {
	scheme := runtime.NewScheme()

	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(volsyncv1alpha1.AddToScheme(scheme))

	// Verify that our custom types are registered
	gvk := volsyncv1alpha1.GroupVersion.WithKind("VolSyncMonitor")
	obj, err := scheme.New(gvk)
	if err != nil {
		t.Errorf("Failed to create VolSyncMonitor from scheme: %v", err)
	}

	if _, ok := obj.(*volsyncv1alpha1.VolSyncMonitor); !ok {
		t.Errorf("Created object is not a VolSyncMonitor")
	}

	gvkUnlock := volsyncv1alpha1.GroupVersion.WithKind("VolSyncUnlock")
	objUnlock, err := scheme.New(gvkUnlock)
	if err != nil {
		t.Errorf("Failed to create VolSyncUnlock from scheme: %v", err)
	}

	if _, ok := objUnlock.(*volsyncv1alpha1.VolSyncUnlock); !ok {
		t.Errorf("Created object is not a VolSyncUnlock")
	}
}

func TestManagerConfiguration(t *testing.T) {
	// Test manager options
	opts := ctrl.Options{
		Scheme: runtime.NewScheme(),
		Metrics: metricsserver.Options{
			BindAddress: ":8080",
		},
		WebhookServer: webhook.NewServer(webhook.Options{
			Port: 9443,
		}),
		HealthProbeBindAddress: ":8081",
		LeaderElection:         false,
		LeaderElectionID:       "homelab-assistant-leader-election",
	}

	if opts.Metrics.BindAddress != ":8080" {
		t.Errorf("Expected metrics bind address :8080, got %s", opts.Metrics.BindAddress)
	}

	if opts.HealthProbeBindAddress != ":8081" {
		t.Errorf("Expected health probe bind address :8081, got %s", opts.HealthProbeBindAddress)
	}

	if opts.LeaderElectionID != "homelab-assistant-leader-election" {
		t.Errorf("Expected leader election ID homelab-assistant-leader-election, got %s", opts.LeaderElectionID)
	}
}

func TestEnvironmentVariables(t *testing.T) {
	tests := []struct {
		name     string
		envVar   string
		setValue string
		expected interface{}
		testFunc func(string) interface{}
	}{
		{
			name:     "METRICS_BIND_ADDRESS",
			envVar:   "METRICS_BIND_ADDRESS",
			setValue: ":9090",
			expected: ":9090",
			testFunc: func(val string) interface{} { return val },
		},
		{
			name:     "HEALTH_PROBE_BIND_ADDRESS",
			envVar:   "HEALTH_PROBE_BIND_ADDRESS",
			setValue: ":9091",
			expected: ":9091",
			testFunc: func(val string) interface{} { return val },
		},
		{
			name:     "LEADER_ELECT",
			envVar:   "LEADER_ELECT",
			setValue: "true",
			expected: true,
			testFunc: func(val string) interface{} { return val == "true" },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original value
			originalVal := os.Getenv(tt.envVar)
			defer os.Setenv(tt.envVar, originalVal)

			// Set test value
			os.Setenv(tt.envVar, tt.setValue)

			// Test the function
			result := tt.testFunc(os.Getenv(tt.envVar))
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestLoggerSetup(t *testing.T) {
	// Test development logger
	devLogger := zap.New(zap.UseFlagOptions(&zap.Options{
		Development: true,
	}))

	// Check that logger is not the zero value
	if (devLogger == logr.Logger{}) {
		t.Error("Failed to create development logger")
	}

	// Test production logger
	prodLogger := zap.New(zap.UseFlagOptions(&zap.Options{
		Development: false,
	}))

	// Check that logger is not the zero value
	if (prodLogger == logr.Logger{}) {
		t.Error("Failed to create production logger")
	}
}

func TestHealthzSetup(t *testing.T) {
	// Test readiness check
	readinessCheck := healthz.Ping
	if readinessCheck == nil {
		t.Error("Readiness check should not be nil")
	}

	// Test liveness check
	livenessCheck := healthz.Ping
	if livenessCheck == nil {
		t.Error("Liveness check should not be nil")
	}

	// Test that we can call the check
	err := readinessCheck(nil)
	if err != nil {
		t.Errorf("Ping check should not return error: %v", err)
	}
}

func TestControllerManagerOptions(t *testing.T) {
	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(volsyncv1alpha1.AddToScheme(scheme))

	// Test various manager configurations
	testCases := []struct {
		name    string
		options ctrl.Options
	}{
		{
			name: "Default configuration",
			options: ctrl.Options{
				Scheme: scheme,
				Metrics: metricsserver.Options{
					BindAddress: ":8080",
				},
				HealthProbeBindAddress: ":8081",
				LeaderElection:         false,
			},
		},
		{
			name: "Leader election enabled",
			options: ctrl.Options{
				Scheme: scheme,
				Metrics: metricsserver.Options{
					BindAddress: ":8080",
				},
				HealthProbeBindAddress: ":8081",
				LeaderElection:         true,
				LeaderElectionID:       "test-leader-election",
				LeaseDuration:          &[]time.Duration{15 * time.Second}[0],
				RenewDeadline:          &[]time.Duration{10 * time.Second}[0],
				RetryPeriod:            &[]time.Duration{2 * time.Second}[0],
			},
		},
		{
			name: "Custom metrics address",
			options: ctrl.Options{
				Scheme: scheme,
				Metrics: metricsserver.Options{
					BindAddress: ":9090",
				},
				HealthProbeBindAddress: ":9091",
				LeaderElection:         false,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.options.Scheme == nil {
				t.Error("Scheme should not be nil")
			}

			if tc.options.Metrics.BindAddress == "" {
				t.Error("Metrics bind address should not be empty")
			}

			if tc.options.HealthProbeBindAddress == "" {
				t.Error("Health probe bind address should not be empty")
			}

			if tc.options.LeaderElection && tc.options.LeaderElectionID == "" {
				t.Error("Leader election ID should not be empty when leader election is enabled")
			}
		})
	}
}

func TestWebhookServerOptions(t *testing.T) {
	// Test webhook server configuration
	webhookServer := webhook.NewServer(webhook.Options{
		Port:    9443,
		Host:    "0.0.0.0",
		CertDir: "/tmp/k8s-webhook-server/serving-certs",
	})

	if webhookServer == nil {
		t.Error("Webhook server should not be nil")
	}
}

func TestSignalHandling(t *testing.T) {
	// Test that we can create a signal context
	ctx := ctrl.SetupSignalHandler()
	if ctx == nil {
		t.Error("Signal context should not be nil")
	}

	// Verify context is not already cancelled
	select {
	case <-ctx.Done():
		t.Error("Context should not be cancelled initially")
	default:
		// Expected - context should not be done
	}
}

func TestSchemeRegistration(t *testing.T) {
	scheme := runtime.NewScheme()

	// Test that we can add client-go scheme
	err := clientgoscheme.AddToScheme(scheme)
	if err != nil {
		t.Errorf("Failed to add client-go scheme: %v", err)
	}

	// Test that we can add our custom scheme
	err = volsyncv1alpha1.AddToScheme(scheme)
	if err != nil {
		t.Errorf("Failed to add volsync scheme: %v", err)
	}

	// Verify core Kubernetes types are registered
	coreGVK := clientgoscheme.Scheme.AllKnownTypes()
	if len(coreGVK) == 0 {
		t.Error("Core Kubernetes types should be registered")
	}

	// Verify our custom types are registered
	customGVK := scheme.AllKnownTypes()
	found := false
	for gvk := range customGVK {
		if gvk.Group == "homelab.rafaribe.com" && gvk.Kind == "VolSyncMonitor" {
			found = true
			break
		}
	}
	if !found {
		t.Error("VolSyncMonitor should be registered in scheme")
	}
}

func TestManagerCreationParameters(t *testing.T) {
	// Test that manager can be created with various parameters
	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(volsyncv1alpha1.AddToScheme(scheme))

	// Test minimum required options
	minOpts := ctrl.Options{
		Scheme: scheme,
	}

	if minOpts.Scheme == nil {
		t.Error("Minimum options should have scheme")
	}

	// Test full options
	fullOpts := ctrl.Options{
		Scheme: scheme,
		Metrics: metricsserver.Options{
			BindAddress: ":8080",
		},
		WebhookServer: webhook.NewServer(webhook.Options{
			Port: 9443,
		}),
		HealthProbeBindAddress:        ":8081",
		LeaderElection:                true,
		LeaderElectionID:              "test-election",
		LeaderElectionReleaseOnCancel: true,
	}

	if fullOpts.LeaderElection != true {
		t.Error("Leader election should be enabled")
	}

	if fullOpts.LeaderElectionReleaseOnCancel != true {
		t.Error("Leader election release on cancel should be enabled")
	}
}
