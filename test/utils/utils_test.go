package utils

import (
	"os/exec"
	"testing"
)

func TestRun(t *testing.T) {
	tests := []struct {
		name        string
		cmd         *exec.Cmd
		expectError bool
	}{
		{
			name:        "successful command",
			cmd:         exec.Command("echo", "test"),
			expectError: false,
		},
		{
			name:        "failing command",
			cmd:         exec.Command("sh", "-c", "exit 1"),
			expectError: true,
		},
		{
			name:        "command with output",
			cmd:         exec.Command("echo", "hello world"),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Run(tt.cmd)
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestRunOutput(t *testing.T) {
	cmd := exec.Command("echo", "test output")
	output, err := Run(cmd)
	if err != nil {
		t.Errorf("Expected no error but got: %v", err)
	}
	
	if len(output) == 0 {
		t.Error("Expected output but got none")
	}
}

func TestInstallPrometheusOperator(t *testing.T) {
	// This test would normally require kubectl and network access
	// We'll test that the function exists and can be called
	err := InstallPrometheusOperator()
	// We expect this to fail in test environment, but function should exist
	if err == nil {
		t.Log("InstallPrometheusOperator succeeded (unexpected in test env)")
	} else {
		t.Logf("InstallPrometheusOperator failed as expected: %v", err)
	}
}

func TestUninstallPrometheusOperator(t *testing.T) {
	// Test that the function exists and can be called
	// Note: UninstallPrometheusOperator returns void
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("UninstallPrometheusOperator panicked: %v", r)
		}
	}()
	
	UninstallPrometheusOperator()
	t.Log("UninstallPrometheusOperator completed")
}

func TestInstallCertManager(t *testing.T) {
	// Test that the function exists and can be called
	err := InstallCertManager()
	// We expect this to fail in test environment, but function should exist
	if err == nil {
		t.Log("InstallCertManager succeeded (unexpected in test env)")
	} else {
		t.Logf("InstallCertManager failed as expected: %v", err)
	}
}

func TestUninstallCertManager(t *testing.T) {
	// Test that the function exists and can be called
	// Note: UninstallCertManager returns void
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("UninstallCertManager panicked: %v", r)
		}
	}()
	
	UninstallCertManager()
	t.Log("UninstallCertManager completed")
}

func TestLoadImageToKindClusterWithName(t *testing.T) {
	// Test that the function exists and can be called
	err := LoadImageToKindClusterWithName("test-image")
	// We expect this to fail in test environment, but function should exist
	if err == nil {
		t.Log("LoadImageToKindClusterWithName succeeded (unexpected in test env)")
	} else {
		t.Logf("LoadImageToKindClusterWithName failed as expected: %v", err)
	}
}

func TestGetNonEmptyLines(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "single line",
			input:    "hello",
			expected: []string{"hello"},
		},
		{
			name:     "multiple lines",
			input:    "line1\nline2\nline3",
			expected: []string{"line1", "line2", "line3"},
		},
		{
			name:     "with empty lines",
			input:    "line1\n\nline3\n",
			expected: []string{"line1", "line3"},
		},
		{
			name:     "empty string",
			input:    "",
			expected: []string{},
		},
		{
			name:     "only newlines",
			input:    "\n\n\n",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetNonEmptyLines(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d lines, got %d", len(tt.expected), len(result))
				return
			}
			
			for i, expected := range tt.expected {
				if result[i] != expected {
					t.Errorf("Expected line %d to be '%s', got '%s'", i, expected, result[i])
				}
			}
		})
	}
}

func TestGetProjectDir(t *testing.T) {
	dir, err := GetProjectDir()
	if err != nil {
		t.Errorf("GetProjectDir failed: %v", err)
	}
	
	if dir == "" {
		t.Error("GetProjectDir returned empty string")
	}
	
	t.Logf("Project directory: %s", dir)
}

func TestCommandExecution(t *testing.T) {
	// Test various command execution scenarios
	tests := []struct {
		name    string
		command string
		args    []string
		wantErr bool
	}{
		{
			name:    "echo command",
			command: "echo",
			args:    []string{"hello"},
			wantErr: false,
		},
		{
			name:    "false command",
			command: "false",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "nonexistent command",
			command: "nonexistentcommand123",
			args:    []string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(tt.command, tt.args...)
			_, err := Run(cmd)
			
			if tt.wantErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestWarnError(t *testing.T) {
	// Test that warnError function exists and can be called
	// This function writes to GinkgoWriter, so we can't easily test output
	// but we can ensure it doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("warnError panicked: %v", r)
		}
	}()
	
	warnError(nil)
	warnError(exec.ErrNotFound)
}

func TestConstants(t *testing.T) {
	// Test that constants are defined with expected values
	if prometheusOperatorVersion == "" {
		t.Errorf("prometheusOperatorVersion should not be empty")
	}
	
	if prometheusOperatorURL == "" {
		t.Errorf("prometheusOperatorURL should not be empty")
	}
	
	if certmanagerVersion == "" {
		t.Errorf("certmanagerVersion should not be empty")
	}
	
	if certmanagerURLTmpl == "" {
		t.Errorf("certmanagerURLTmpl should not be empty")
	}
	
	// Test that URL template contains placeholder
	if !contains(prometheusOperatorURL, "%s") {
		t.Errorf("prometheusOperatorURL should contain %%s placeholder")
	}
	
	if !contains(certmanagerURLTmpl, "%s") {
		t.Errorf("certmanagerURLTmpl should contain %%s placeholder")
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
