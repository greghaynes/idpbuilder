# Validation and Testing Strategy for Codespaces

**Status:** Proposal  
**Version:** 1.0  
**Date:** March 2026  
**Authors:** IDP Builder Team

## Executive Summary

This document defines a validation and testing strategy that can be executed remotely in GitHub Codespaces while providing significant validation that idpbuilder meets its functional requirements. The strategy establishes a multi-tier testing pyramid where each tier adds coverage depth at increasing resource cost, with clear guidance on which tiers are appropriate for different execution contexts (local development, Codespaces, CI/CD).

## Problem Statement

idpbuilder currently has two distinct testing modes:

1. **Unit Tests** (`make test`): Fast, no Docker required, runs Go tests with fake Kubernetes clients.
2. **End-to-End Tests** (`make e2e`): Heavy, requires a running Kind cluster, 10–15 minute execution time.

This leaves a significant gap: developers working in GitHub Codespaces can run unit tests but cannot conveniently validate that the platform's core functional requirements are met (controller reconciliation, provider lifecycle, CLI behavior) without committing to a full e2e run. The goal is to fill this gap with a structured, tiered strategy.

## Goals

1. **Runnable in Codespaces**: All tier 1–4 tests must complete in a Codespaces environment (4 CPUs, Docker-in-Docker) without requiring pre-provisioned external infrastructure.
2. **Significant Functional Validation**: Tests must verify that controller reconciliation logic, provider interactions, and CLI behaviors work correctly against real or realistic test harnesses.
3. **Fast Feedback**: The default Codespaces validation target (`make validate`) must complete in under 5 minutes.
4. **Progressive Depth**: Each successive tier provides deeper validation, allowing developers to select the right level of confidence for their workflow.
5. **CI Alignment**: The same targets used in Codespaces must be usable in CI/CD pipelines without modification.

## Functional Requirements Mapped to Test Coverage

The following functional requirements are derived from the project architecture and must be validated at one or more tiers:

| Functional Requirement | Tier |
|------------------------|------|
| Code follows Go formatting conventions | 1 |
| Code passes static analysis | 1 |
| All documentation is linked in site navigation | 1 |
| Binary builds successfully with version metadata | 2 |
| CLI produces correct help/version output | 2 |
| Build does not produce uncommitted file changes | 2 |
| Platform controller aggregates provider status | 3 |
| GiteaProvider controller reconciles resources | 3 |
| NginxGateway controller reconciles resources | 3 |
| ArgoCDProvider controller reconciles resources | 3 |
| CustomPackage controller resolves Git provider | 3 |
| Status reporter tracks phases and sub-steps | 3 |
| Duck-typing utilities extract provider endpoints | 3 |
| Platform installs ArgoCD, Gitea, and Nginx | 4 |
| ArgoCD applications sync and become healthy | 4 |
| Gitea endpoints respond and repositories exist | 4 |
| Container registry push/pull works | 4 |
| Custom packages deploy via ArgoCD | 4 |

## Validation Tiers

### Tier 1: Static Analysis (< 30 seconds)

**Purpose**: Catch formatting, linting, and documentation consistency issues before any compilation.

**Tests included**:
- `go fmt ./...` — ensures all Go source files are properly formatted
- `go vet ./...` — catches common Go programmer errors (unreachable code, incorrect format strings, etc.)
- `make validate-docs` — ensures all documentation files are reachable from the site navigation

**Infrastructure requirements**: Go toolchain only (no Docker, no Kubernetes)

**Makefile target**: `make fmt-check vet validate-docs`

---

### Tier 2: Binary Smoke Tests (< 2 minutes)

**Purpose**: Validate that the project compiles, embeds resources correctly, and the CLI behaves as expected.

**Tests included**:
- `make build` — compiles the binary with embedded manifests and version metadata
- `git diff-index --quiet HEAD` — detects uncommitted file changes caused by code generation
- `./idpbuilder --help` — verifies the binary runs and produces expected output
- `./idpbuilder version` — verifies version metadata is embedded at build time

**Infrastructure requirements**: Go toolchain + `kustomize` + `helm` (auto-downloaded by `make`)

**Makefile target**: `make smoke`

---

### Tier 3: Functional/Integration Tests (< 3 minutes)

**Purpose**: Validate controller reconciliation logic, provider interactions, and status tracking using the Go testing framework with fake Kubernetes clients. No real Kubernetes cluster or network access is required.

**Tests included**:
- All tests in `pkg/controllers/` — controller reconciliation with `sigs.k8s.io/controller-runtime/pkg/client/fake`
- All tests in `pkg/status/` — phase and sub-step tracking
- All tests in `pkg/util/provider/` — duck-typing utilities for Git, Gateway, and GitOps providers
- All tests in `pkg/k8s/` — Kubernetes utility functions
- All tests in `pkg/kind/` — Kind configuration generation
- All tests in `pkg/cmd/` — CLI command parsing and helpers

**Infrastructure requirements**: Go toolchain only

**Makefile target**: `make test`

**Testing patterns used**:
- `fake.NewClientBuilder()` from `controller-runtime` for in-memory Kubernetes API server
- `github.com/stretchr/testify/assert` and `require` for assertions
- Table-driven tests (`[]struct{ name, input, expected }`) for comprehensive scenario coverage

---

### Tier 4: End-to-End Integration Tests (10–15 minutes)

**Purpose**: Validate that idpbuilder correctly provisions a complete IDP platform including a real Kind cluster, live ArgoCD and Gitea instances, and working container registry.

**Tests included**:
- `tests/e2e/docker/docker_test.go` — full lifecycle tests using the compiled binary:
  - `testCreate` — default install with subdomain routing
  - `testCreatePath` — install with path-based routing
  - `testCustomPkg` — custom ArgoCD package installation
  - `testPackagePriority` — package installation priority ordering
  - `TestGiteaRegistry` — container image push/pull via Gitea registry
  - `TestGiteaRegistryInCluster` — in-cluster container image pull

**Infrastructure requirements**: Docker (Docker-in-Docker in Codespaces), `kind` CLI (installed by devcontainer `postCreateCommand.sh`)

**Makefile target**: `make e2e`

**Execution context**: Supported in Codespaces (4+ CPU required), CI/CD with Docker access

---

## Codespaces Validation Target

The primary validation target for Codespaces is `make validate`, which combines Tiers 1–3:

```
make validate = make fmt-check + vet + build (dirty check) + test
```

This target:
- Takes under 5 minutes on a 4-CPU Codespace
- Requires no external services or network access during the test phase
- Validates all functional requirements except full end-to-end provisioning (Tier 4)
- Is safe to run repeatedly during development

Tier 4 (`make e2e`) is also supported in Codespaces given that Docker-in-Docker is enabled. It requires approximately 15 minutes and is recommended before raising a pull request on infrastructure-affecting changes.

## Implementation

### Makefile Targets

The following targets implement the strategy:

| Target | Description | Tiers |
|--------|-------------|-------|
| `make fmt-check` | Check Go formatting | 1 |
| `make vet` | Run `go vet` | 1 |
| `make validate-docs` | Validate documentation links | 1 |
| `make build` | Build binary | 2 |
| `make smoke` | Build binary and verify CLI output | 2 |
| `make test` | Run unit and functional tests | 3 |
| `make validate` | Run Tiers 1–3 (Codespaces default) | 1, 2, 3 |
| `make e2e` | Run full end-to-end tests | 4 |

### GitHub Actions Workflow

A dedicated workflow file (`.github/workflows/codespaces-validation.yaml`) runs `make validate` on every pull request using `ubuntu-latest` runners that closely match the Codespaces environment. This provides automated validation without requiring dedicated self-hosted runners.

### devcontainer Integration

The `.devcontainer/devcontainer.json` configuration documents `make validate` as the recommended first command to run after the container starts, ensuring developers can quickly confirm their environment is correctly set up.

## Test Writing Guidelines

When writing tests for idpbuilder, follow these guidelines to ensure they fit appropriately within the tier system:

### Controller Tests (Tier 3)

Use `fake.NewClientBuilder()` for all controller tests. Tests requiring a real Kubernetes API server should use the `//go:build integration` build tag and be excluded from `make test`.

```go
fakeClient := fake.NewClientBuilder().
    WithScheme(scheme).
    WithObjects(resource).
    WithStatusSubresource(&v1alpha2.MyResource{}).
    Build()

reconciler := &MyReconciler{Client: fakeClient, Scheme: scheme}
result, err := reconciler.Reconcile(ctx, req)
```

### CLI Tests (Tier 3)

Use `cobra` command testing patterns with captured output buffers:

```go
cmd := NewMyCommand()
buf := &bytes.Buffer{}
cmd.SetOut(buf)
err := cmd.Execute()
assert.NoError(t, err)
assert.Contains(t, buf.String(), "expected output")
```

### E2E Tests (Tier 4)

E2E tests must use the `//go:build e2e` build tag. Tests call the compiled `idpbuilder` binary directly using `exec.CommandContext` with appropriate timeouts (8 minutes per sub-test recommended).

### Skipping Tests

Tests that require infrastructure not available in the tier should use `t.Skip()` with a clear message:

```go
t.Skip("Requires envtest environment - run with 'make test-integration'")
```

## Future Considerations

### envtest Integration Tests

The `integration_test.go` files using the `//go:build integration` tag currently skip by default. A future `make test-integration` target could run these tests using `sigs.k8s.io/controller-runtime/tools/setup-envtest` to provide a real Kubernetes API server without a full cluster. This would add coverage for:

- Controller-manager startup and RBAC validation
- Webhook admission logic
- CRD validation rules

### Codespaces Quick E2E

A `make e2e-quick` target could run a reduced subset of e2e tests focused on the most critical paths (default install + ArgoCD health check) with a shorter timeout (5 minutes), suitable for rapid validation in Codespaces without the full 15-minute suite.

## Related Documents

- [Controller Architecture Specification](./controller-architecture-spec.md) — Architecture this strategy validates
- [Implementation Roadmap](../implementation/implementation-roadmap.md) — Current implementation status
- [Test Coverage Improvement Plan](../implementation/test-coverage-improvement-plan.md) — Coverage gap analysis
