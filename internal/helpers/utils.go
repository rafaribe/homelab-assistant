package helpers

import (
	"context"
	"fmt"
	"io"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

// Int32Ptr returns a pointer to an int32 value
func Int32Ptr(i int32) *int32 {
	return &i
}

// Int64Ptr returns a pointer to an int64 value
func Int64Ptr(i int64) *int64 {
	return &i
}

// StringPtr returns a pointer to a string value
func StringPtr(s string) *string {
	return &s
}

// GetPodLogs retrieves logs from a pod container
func GetPodLogs(ctx context.Context, c client.Client, namespace, podName, containerName string) (string, error) {
	// Get the rest config
	cfg, err := config.GetConfig()
	if err != nil {
		return "", fmt.Errorf("failed to get config: %w", err)
	}

	// Create a clientset
	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return "", fmt.Errorf("failed to create clientset: %w", err)
	}

	// Set log options
	logOptions := &corev1.PodLogOptions{
		Container: containerName,
		Follow:    false,
	}

	// Get the logs
	req := clientset.CoreV1().Pods(namespace).GetLogs(podName, logOptions)
	readCloser, err := req.Stream(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get pod logs: %w", err)
	}
	defer readCloser.Close()

	// Read all logs
	logs, err := io.ReadAll(readCloser)
	if err != nil {
		return "", fmt.Errorf("failed to read pod logs: %w", err)
	}

	return string(logs), nil
}
