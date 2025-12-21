# Test Coverage Improvement Plan

## Executive Summary

This document identifies modules in the idpbuilder codebase with significant functionality but very low test coverage (< 30%). It provides recommendations for functional tests, particularly using fake Kubernetes clients, that would provide valuable coverage for groups of related modules.

**Current Overall Coverage: 27.3%**

## Analysis Methodology

1. Generated test coverage report using `go test ./... -coverprofile cover.out`
2. Analyzed coverage by file, excluding:
   - Generated code (e.g., `zz_generated.deepcopy.go`)
   - Pure type definitions in `/api/` directories
3. Identified modules with < 30% coverage that contain significant business logic
4. Grouped modules by functional area for targeted testing

## Modules with Low Coverage

### Critical Business Logic Modules (0% Coverage)

The following modules have zero test coverage but contain significant business logic:

| Module | Functions | Description |
|--------|-----------|-------------|
| `pkg/controllers/localbuild/controller.go` | 23 (7.4%) | Core localbuild reconciliation logic |
| `pkg/controllers/gitopsprovider/argocdprovider_controller.go` | 7 | ArgoCD provider management |
| `pkg/logger/handler.go` | 11 | Custom structured logging handler |
| `pkg/cmd/get/clusters.go` | 9 | Cluster information retrieval |
| `pkg/cmd/create/root.go` | 7 | Cluster creation command logic |
| `pkg/cmd/get/packages.go` | 6 | Package information retrieval |
| `pkg/util/files/files.go` | 6 | File and directory utilities |
| `pkg/cmd/version/root.go` | 4 | Version command |
| `pkg/util/k8s.go` | 4 | Kubernetes client utilities |
| `pkg/printer/*` | 9 | Output formatting (JSON/YAML/Table) |
| `pkg/controllers/localbuild/argo.go` | 3 | ArgoCD installation |
| `pkg/controllers/localbuild/gitea.go` | 3 | Gitea installation |
| `pkg/controllers/crd.go` | 3 | CRD management |
| `pkg/util/argocd.go` | 2 | ArgoCD utilities |
| `pkg/util/idp.go` | 2 | IDP configuration utilities |
| `pkg/resources/localbuild/application.go` | 2 | Argo application specs |

### Logger and Utility Modules (0-10% Coverage)

| Module | Functions | Coverage | Description |
|--------|-----------|----------|-------------|
| `pkg/kind/kindlogger.go` | 10 | 10.0% | Kind cluster logging adapter |
| `pkg/cmd/helpers/logger.go` | 3 | 0% | CLI logging setup |

## Recommended Testing Strategy

### Group 1: Controller Tests with Fake Kubernetes Client

**Target Modules:**
- `pkg/controllers/localbuild/controller.go`
- `pkg/controllers/gitopsprovider/argocdprovider_controller.go`
- `pkg/controllers/localbuild/argo.go`
- `pkg/controllers/localbuild/gitea.go`
- `pkg/controllers/crd.go`

**Testing Approach:**

Use the existing testing pattern found in `pkg/controllers/gatewayprovider/nginxgateway_functional_test.go` and `pkg/controllers/gitprovider/giteaprovider_controller_test.go`:

```go
// Example test structure for LocalbuildReconciler
func TestLocalbuildReconciler_BasicReconciliation(t *testing.T) {
    scheme := k8s.GetScheme()
    
    // Create test resources
    localbuild := &v1alpha1.Localbuild{
        ObjectMeta: metav1.ObjectMeta{
            Name:      "test-localbuild",
            Namespace: "test-namespace",
        },
        Spec: v1alpha1.LocalbuildSpec{
            // Test spec
        },
    }
    
    // Create fake client with resources
    fakeClient := fake.NewClientBuilder().
        WithScheme(scheme).
        WithObjects(localbuild).
        WithStatusSubresource(&v1alpha1.Localbuild{}).
        Build()
    
    reconciler := &LocalbuildReconciler{
        Client: fakeClient,
        Scheme: scheme,
        Config: v1alpha1.BuildCustomizationSpec{
            Host:     "test.example.com",
            Port:     "8443",
            Protocol: "https",
        },
    }
    
    // Test reconciliation
    req := ctrl.Request{
        NamespacedName: types.NamespacedName{
            Name:      localbuild.Name,
            Namespace: localbuild.Namespace,
        },
    }
    
    result, err := reconciler.Reconcile(context.Background(), req)
    
    // Assertions
    require.NoError(t, err)
    assert.NotNil(t, result)
}
```

**Specific Test Cases:**

1. **LocalbuildReconciler Tests** (`pkg/controllers/localbuild/controller_test.go`):
   - Test reconciliation adds finalizers
   - Test project namespace creation
   - Test core package installation coordination
   - Test static password updates (ArgoCD)
   - Test ArgoCD app reconciliation with Gitea
   - Test shutdown behavior and cleanup
   - Test error handling and requeue logic
   - Test status updates (ObservedGeneration)

2. **ArgoCDProviderReconciler Tests** (`pkg/controllers/gitopsprovider/argocdprovider_controller_test.go`):
   - Test provider creation and initialization
   - Test ArgoCD installation via manifests
   - Test admin credential generation
   - Test status updates (phase, conditions)
   - Test namespace creation
   - Test error handling during installation
   - Test reconciliation idempotency

3. **CRD Management Tests** (`pkg/controllers/crd_test.go`):
   - Test CRD existence checking
   - Test CRD installation from embedded resources
   - Test CRD update/patching
   - Test multiple CRD installation
   - Test error handling for malformed CRDs

### Group 2: Utility Function Tests

**Target Modules:**
- `pkg/util/k8s.go`
- `pkg/util/argocd.go`
- `pkg/util/idp.go`
- `pkg/util/files/files.go`

**Testing Approach:**

Standard unit tests with mocks and temporary directories:

```go
// Example for pkg/util/k8s_test.go
func TestGetKubeConfigPath(t *testing.T) {
    // Test default path
    path := GetKubeConfigPath()
    assert.Contains(t, path, ".kube/config")
    
    // Test custom path
    KubeConfigPath = "/custom/path/config"
    defer func() { KubeConfigPath = "" }()
    path = GetKubeConfigPath()
    assert.Equal(t, "/custom/path/config", path)
}

func TestLoadKubeConfig(t *testing.T) {
    // Create temporary kubeconfig
    tmpDir := t.TempDir()
    configPath := filepath.Join(tmpDir, "config")
    
    // Write valid kubeconfig
    kubeconfig := `
apiVersion: v1
kind: Config
clusters:
- cluster:
    server: https://localhost:6443
  name: test-cluster
contexts:
- context:
    cluster: test-cluster
    user: test-user
  name: test-context
current-context: test-context
users:
- name: test-user
  user:
    token: test-token
`
    err := os.WriteFile(configPath, []byte(kubeconfig), 0644)
    require.NoError(t, err)
    
    KubeConfigPath = configPath
    defer func() { KubeConfigPath = "" }()
    
    config, err := LoadKubeConfig()
    require.NoError(t, err)
    assert.NotNil(t, config)
    assert.Equal(t, "test-context", config.CurrentContext)
}
```

**Specific Test Cases:**

1. **K8s Utilities** (`pkg/util/k8s_test.go`):
   - Test `GetKubeConfigPath()` with default and custom paths
   - Test `LoadKubeConfig()` with valid and invalid configs
   - Test `GetKubeConfig()` error handling
   - Test `GetKubeClient()` with mock rest.Config

2. **ArgoCD Utilities** (`pkg/util/argocd_test.go`):
   - Test `ArgocdBaseUrl()` with path routing enabled/disabled
   - Test `ArgocdInitialAdminSecretObject()` structure
   - Verify secret name, namespace, and type

3. **IDP Utilities** (`pkg/util/idp_test.go`):
   - Test `GetConfig()` with fake client and localbuild resources
   - Test error handling when localbuild not found
   - Test default values

4. **File Utilities** (`pkg/util/files/files_test.go`):
   - Test `CopyDirectory()` with nested directories
   - Test `Copy()` for individual files
   - Test `Exists()` for files and directories
   - Test `CreateIfNotExists()` for new and existing directories
   - Test `ApplyTemplate()` with various template data
   - Test template function `indentNewLines()`

### Group 3: Printer/Formatter Tests

**Target Modules:**
- `pkg/printer/printer.go`
- `pkg/printer/cluster.go`
- `pkg/printer/package.go`
- `pkg/printer/secret.go`

**Testing Approach:**

Test output formatting with sample data:

```go
// Example for pkg/printer/printer_test.go
func TestPrintDataAsJson(t *testing.T) {
    data := map[string]string{
        "key1": "value1",
        "key2": "value2",
    }
    
    var buf bytes.Buffer
    err := PrintDataAsJson(data, &buf)
    
    require.NoError(t, err)
    
    // Verify JSON output
    var result map[string]string
    err = json.Unmarshal(buf.Bytes(), &result)
    require.NoError(t, err)
    assert.Equal(t, data, result)
}

func TestPrintDataAsTable(t *testing.T) {
    table := metav1.Table{
        ColumnDefinitions: []metav1.TableColumnDefinition{
            {Name: "Name", Type: "string"},
            {Name: "Value", Type: "string"},
        },
        Rows: []metav1.TableRow{
            {Cells: []interface{}{"test", "123"}},
        },
    }
    
    var buf bytes.Buffer
    err := PrintDataAsTable(table, &buf)
    
    require.NoError(t, err)
    assert.Contains(t, buf.String(), "NAME")
    assert.Contains(t, buf.String(), "test")
}
```

**Specific Test Cases:**

1. **Generic Printer Tests** (`pkg/printer/printer_test.go`):
   - Test `PrintDataAsJson()` with various data types
   - Test `PrintDataAsYaml()` with various data types
   - Test `PrintDataAsTable()` with different table structures
   - Test special characters and escaping
   - Test empty data handling

2. **Cluster Printer Tests** (`pkg/printer/cluster_test.go`):
   - Test `ClusterPrinter.PrintOutput()` for each format
   - Test `generateClusterTable()` with multiple clusters
   - Test `generateNodeData()` with multiple nodes
   - Test empty cluster list
   - Test format validation

3. **Package Printer Tests** (`pkg/printer/package_test.go`):
   - Test package output in all formats
   - Test table generation with package data
   - Test empty package list

4. **Secret Printer Tests** (`pkg/printer/secret_test.go`):
   - Test secret output (ensuring sensitive data is handled correctly)
   - Test table generation with secret metadata
   - Test format switching

### Group 4: CLI Command Tests

**Target Modules:**
- `pkg/cmd/get/clusters.go`
- `pkg/cmd/get/packages.go`
- `pkg/cmd/create/root.go`
- `pkg/cmd/version/root.go`

**Testing Approach:**

Use cobra command testing with mock clients:

```go
// Example for pkg/cmd/get/clusters_test.go
func TestClustersCommand(t *testing.T) {
    // Setup fake Kind cluster list
    // Mock kubernetes client
    
    cmd := &cobra.Command{
        Use: "clusters",
        RunE: func(cmd *cobra.Command, args []string) error {
            // Test implementation
            return nil
        },
    }
    
    // Execute command
    err := cmd.Execute()
    require.NoError(t, err)
}
```

**Specific Test Cases:**

1. **Get Clusters Command** (`pkg/cmd/get/clusters_test.go`):
   - Test cluster listing with multiple clusters
   - Test cluster filtering by name
   - Test output format options (json, yaml, table)
   - Test error handling when no clusters exist
   - Test port mapping extraction
   - Test resource allocation display

2. **Get Packages Command** (`pkg/cmd/get/packages_test.go`):
   - Test package listing
   - Test package filtering by name
   - Test output formats
   - Test error handling

3. **Create Command** (`pkg/cmd/create/root_test.go`):
   - Test validation of package custom files
   - Test configuration validation
   - Test package path resolution
   - Test error messages for invalid inputs

4. **Version Command** (`pkg/cmd/version/root_test.go`):
   - Test version output in different formats
   - Test version information accuracy
   - Test JSON and YAML output parsing

### Group 5: Logger Tests

**Target Modules:**
- `pkg/logger/handler.go`
- `pkg/kind/kindlogger.go`
- `pkg/cmd/helpers/logger.go`

**Testing Approach:**

Test logging output and configuration:

```go
// Example for pkg/logger/handler_test.go
func TestHandler_BasicLogging(t *testing.T) {
    var buf bytes.Buffer
    handler := NewHandler(&buf, Options{
        Colored: false,
        Level:   slog.LevelInfo,
    })
    
    logger := slog.New(handler)
    logger.Info("test message", "key", "value")
    
    output := buf.String()
    assert.Contains(t, output, "test message")
    assert.Contains(t, output, "key=value")
}

func TestHandler_ColoredOutput(t *testing.T) {
    var buf bytes.Buffer
    handler := NewHandler(&buf, Options{
        Colored: true,
        Level:   slog.LevelInfo,
    })
    
    logger := slog.New(handler)
    logger.Error("error message")
    
    output := buf.String()
    // Should contain ANSI color codes
    assert.Contains(t, output, "\033[")
}
```

**Specific Test Cases:**

1. **Custom Handler Tests** (`pkg/logger/handler_test.go`):
   - Test basic logging at different levels
   - Test colored vs non-colored output
   - Test log level filtering
   - Test attribute handling
   - Test group handling
   - Test source location (when enabled)
   - Test concurrent logging (thread safety)
   - Test buffer pool management

2. **Kind Logger Tests** (`pkg/kind/kindlogger_test.go`):
   - Test logger adapter methods (Warn, Warnf, Error, Errorf)
   - Test verbosity levels
   - Test integration with logr.Logger
   - Test enabled/disabled state

3. **CLI Logger Setup Tests** (`pkg/cmd/helpers/logger_test.go`):
   - Test logger configuration from CLI flags
   - Test log level parsing
   - Test default values

## Implementation Priority

### High Priority (Critical for Core Functionality)
1. **Controller Tests**: LocalbuildReconciler, ArgoCDProviderReconciler, CRD management
2. **Utility Tests**: k8s.go, argocd.go, idp.go (these are widely used)

### Medium Priority (User-Facing Features)
3. **Printer Tests**: All printer modules (affects CLI output)
4. **CLI Command Tests**: get commands, create command

### Low Priority (Infrastructure)
5. **Logger Tests**: Handler, kindlogger, CLI logger setup
6. **File Utility Tests**: files.go

## Testing Best Practices

1. **Use Fake Kubernetes Clients**: Leverage `sigs.k8s.io/controller-runtime/pkg/client/fake` for controller tests
2. **Test Status Updates**: Use `WithStatusSubresource()` when building fake clients
3. **Test Error Paths**: Ensure error handling is tested, not just happy paths
4. **Use Table-Driven Tests**: For testing multiple scenarios efficiently
5. **Mock External Dependencies**: Avoid real network calls, file I/O where possible
6. **Test Idempotency**: Controllers should handle repeated reconciliation
7. **Verify State Changes**: Check that resources are created/updated as expected
8. **Test Finalizers**: Ensure cleanup logic is tested for controllers
9. **Use Temporary Directories**: For file system tests, use `t.TempDir()`
10. **Parallel Tests**: Use `t.Parallel()` where tests are independent

## Expected Coverage Improvements

With the recommended tests implemented, we expect:

- **Controllers**: 0-7% → 60-70% coverage
- **Utilities**: 0% → 80-90% coverage  
- **Printers**: 0% → 90-95% coverage
- **CLI Commands**: 0% → 50-60% coverage
- **Loggers**: 0-10% → 70-80% coverage

**Overall Project**: 27.3% → 50-55% coverage

## References

- Existing test patterns: `pkg/controllers/gatewayprovider/nginxgateway_functional_test.go`
- Existing test patterns: `pkg/controllers/gitprovider/giteaprovider_controller_test.go`
- Fake client documentation: https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/client/fake
- Controller testing guide: https://book.kubebuilder.io/reference/writing-tests.html

## Appendix: Coverage Report Summary

```
Package                                                           Coverage
================================================================================
pkg/controllers/localbuild                                        3.9%
pkg/controllers/gitprovider                                       13.8%
pkg/build                                                         25.1%
pkg/cmd/get                                                       24.5%
pkg/controllers/custompackage                                     58.4%
pkg/controllers/gatewayprovider                                   42.4%
pkg/controllers/gitrepository                                     52.4%
pkg/controllers/platform                                          61.4%
pkg/k8s                                                           56.9%
pkg/kind                                                          58.9%
pkg/resources/gitea                                               11.4%
pkg/util                                                          45.4%
pkg/util/fs                                                       52.9%
pkg/util/provider                                                 86.3%

Total                                                             27.3%
```
