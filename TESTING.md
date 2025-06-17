# Testing Guide

This document describes the testing approach for the Homelab Assistant project.

## Testing Philosophy

**Safety First**: This project is designed to never accidentally connect to or modify real Kubernetes clusters during automated testing. All tests use mocked or fake Kubernetes environments.

## Test Types

### 1. Unit Tests
- **Location**: `internal/controller/*_test.go`, `api/v1alpha1/*_test.go`
- **Environment**: Pure Go unit tests with mocked dependencies
- **Purpose**: Test individual functions and methods in isolation
- **Run with**: `make test`

### 2. Controller Integration Tests
- **Location**: `internal/controller/suite_test.go`
- **Environment**: Uses `envtest` (fake Kubernetes API server)
- **Purpose**: Test controller logic against a fake Kubernetes API
- **Safety**: ✅ **Safe** - Uses fake API server, no real cluster connection
- **Run with**: `make test`

### 3. Helm Chart Tests
- **Location**: `charts/*/tests/*.yaml`
- **Environment**: `helm unittest` with template rendering
- **Purpose**: Test Helm chart template generation and values
- **Safety**: ✅ **Safe** - Only tests template rendering
- **Run with**: `make chart-test`

### 4. Chart Linting
- **Purpose**: Validate Helm chart structure and best practices
- **Safety**: ✅ **Safe** - Static analysis only
- **Run with**: `make chart-lint`

## Removed Tests

### ❌ E2E Tests (Removed for Safety)
- **Previously**: `test/e2e/` directory
- **Why Removed**: These tests used real Kubernetes clusters via `kubectl`
- **Risk**: Could accidentally modify cert-manager, prometheus-operator, or other critical namespaces
- **Alternative**: Use controller integration tests with envtest instead

### ❌ Chart Installation Tests (Disabled)
- **Previously**: `ct install` in GitHub Actions
- **Why Disabled**: Creates real Kind clusters and installs charts
- **Risk**: Could connect to wrong cluster context
- **Alternative**: Helm unittest covers template validation

## Running Tests

### Local Development
```bash
# Run all safe tests
make test

# Run chart tests
make chart-test

# Run chart linting
make chart-lint

# Generate chart documentation
make chart-docs
```

### CI/CD Pipeline
- **GitHub Actions**: Automatically runs all safe tests on PR/push
- **No Real Clusters**: CI never connects to real Kubernetes clusters
- **Chart Testing**: Only lints and unit tests charts, no installation

## Manual Testing (Use with Caution)

If you need to test against a real cluster, use these commands with extreme caution:

```bash
# WARNING: These commands use your current kubeconfig context!

# Check current context first
kubectl config current-context

# Install to test cluster (with confirmation prompt)
make chart-install-local

# Uninstall from test cluster (with confirmation prompt)  
make chart-uninstall-local
```

**⚠️ Important**: These commands will prompt for confirmation and show your current Kubernetes context before proceeding.

## Test Coverage

### What's Tested
- ✅ Controller reconciliation logic
- ✅ Custom resource validation
- ✅ Helm chart template generation
- ✅ Chart values and configuration
- ✅ RBAC permissions and security contexts
- ✅ Metrics and monitoring setup

### What's Not Tested (By Design)
- ❌ Real cluster installation
- ❌ Real VolSync integration
- ❌ Real restic repository operations
- ❌ Actual job execution in clusters

## Adding New Tests

### Guidelines
1. **Never connect to real clusters** in automated tests
2. **Use envtest** for controller testing
3. **Use helm unittest** for chart testing
4. **Mock external dependencies** (Kubernetes API, etc.)
5. **Add confirmation prompts** for any manual cluster operations

### Example: Adding a Controller Test
```go
// Use envtest environment from suite_test.go
var _ = Describe("VolSyncMonitor Controller", func() {
    It("should create unlock jobs", func() {
        // Test uses fake Kubernetes API from envtest
        monitor := &volsyncv1alpha1.VolSyncMonitor{...}
        Expect(k8sClient.Create(ctx, monitor)).To(Succeed())
        // ... test logic
    })
})
```

### Example: Adding a Chart Test
```yaml
# charts/homelab-assistant/tests/new_test.yaml
suite: test new feature
templates:
  - deployment.yaml
tests:
  - it: should configure new feature
    set:
      newFeature.enabled: true
    asserts:
      - equal:
          path: spec.template.spec.containers[0].env[0].name
          value: NEW_FEATURE_ENABLED
```

## Troubleshooting

### "No tests found" Error
- Ensure test files end with `_test.go`
- Check that test functions start with `Test` or use Ginkgo `It` blocks

### "Cannot connect to cluster" Error
- This should never happen in CI - indicates a test is trying to use real clusters
- Check that tests use `envtest` or mocked clients only

### Chart Test Failures
- Run `helm unittest charts/chart-name` locally to debug
- Check that test values match actual chart structure
- Verify template paths in assertions

## Security Considerations

- **No kubeconfig access** in automated tests
- **No real cluster modifications** during CI/CD
- **Confirmation prompts** for manual cluster operations
- **Isolated test environments** using envtest and mocks
- **Safe defaults** that prevent accidental cluster access
