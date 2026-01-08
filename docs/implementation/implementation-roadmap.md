# Implementation Roadmap: Controller-Based Architecture

**Version:** 2.0  
**Date:** January 2026  
**Status:** Current

## Executive Summary

This document outlines the remaining work to complete the migration from the v1alpha1 (Localbuild) architecture to the v1alpha2 (Platform-based) controller architecture. Much of the foundation has been completed, and this roadmap focuses on the final steps needed to fully deprecate the Localbuild controller.

## Current Implementation Status

### ✅ Completed Work

The following components have been successfully implemented:

#### Phase 1: v1alpha2 CRDs and Controllers
- ✅ **GiteaProvider CRD** - Defined in `api/v1alpha2/giteaprovider_types.go`
- ✅ **GiteaProvider Controller** - Implemented in `pkg/controllers/gitprovider/giteaprovider_controller.go`
  - Waits for Platform owner reference before reconciling
  - Discovers configuration from Platform CR
  - Full unit test coverage

- ✅ **NginxGateway CRD** - Defined in `api/v1alpha2/nginxgateway_types.go`
- ✅ **NginxGateway Controller** - Implemented in `pkg/controllers/gatewayprovider/nginxgateway_controller.go`
  - Complete implementation with tests
  - Integration tests in place

- ✅ **ArgoCDProvider CRD** - Defined in `api/v1alpha2/argocdprovider_types.go`
- ✅ **ArgoCDProvider Controller** - Implemented in `pkg/controllers/gitopsprovider/argocdprovider_controller.go`
  - Full controller implementation
  - Unit tests included

- ✅ **Platform CRD** - Defined in `api/v1alpha2/platform_types.go`
- ✅ **Platform Controller** - Implemented in `pkg/controllers/platform/platform_controller.go`
  - Owner reference pattern fully implemented
  - Aggregates status from all three provider types (Git, Gateway, GitOps)
  - Duck-typing support for provider independence
  - RBAC permissions for updating provider CRs

#### Phase 2: CLI Integration
- ✅ **CLI creates GiteaProvider CR** - `pkg/build/build.go::createGiteaProvider()` (line 394)
- ✅ **CLI creates ArgoCDProvider CR** - `pkg/build/build.go::createArgoCDProvider()` (line 428)
- ✅ **CLI creates NginxGateway CR** - `pkg/build/build.go::createNginxGateway()` (line 456)
- ✅ **CLI creates Platform CR** - `pkg/build/build.go::createPlatform()` (line 481)
  - Platform references all three provider types
  - Proper ordering: Providers created before Platform

#### Phase 3: Provider Duck-Typing
- ✅ **Duck-typing utilities** - `pkg/util/provider/` directory
  - `git.go` - Git provider duck-typing
  - `gateway.go` - Gateway provider duck-typing
  - `gitops.go` - GitOps provider duck-typing

### 🚧 Partially Complete

- **Owner Reference Pattern** ✅ Implemented but providers created AFTER Platform
  - Platform controller adds owner references ✅
  - Providers wait for owner references ✅
  - **Issue:** CLI creates Platform AFTER providers, causing initial reconcile to happen before owner ref is set
  - **Fix needed:** See Priority 1 below

### ✅ Recently Completed

#### Priority 4: Localbuild CR Creation Removed (January 2026)
- ✅ **CLI no longer creates Localbuild CR** - Removed from `pkg/build/build.go`
- ✅ **Only v1alpha2 CRs created** - Platform, GiteaProvider, ArgoCDProvider, NginxGateway
- ✅ **Compatibility check removed** - `isCompatible()` function and related code deleted
- ✅ **Tests updated** - Removed obsolete `TestIsCompatible`
- Note: Localbuild controller remains for backward compatibility with existing installations

### ❌ Not Yet Implemented

#### Critical Path Items

1. **Bootstrap Repository Creation** - Currently handled by Localbuild controller
   - Platform should create GitRepository CRs for bootstrap apps (argocd, gitea, nginx)
   - Platform should create ArgoCD Application CRs
   - Code exists in `pkg/controllers/localbuild/controller.go` but needs to be moved

2. **Custom Package Handling** - Currently handled by Localbuild controller
   - CustomPackage controller exists but may need updates
   - Platform may need to orchestrate custom package deployment
   - Code in `pkg/controllers/localbuild/controller.go::reconcileCustomPkg()`

3. **Localbuild Controller Deprecation** - Still active
   - Controller still exists in `pkg/controllers/localbuild/`
   - Still registered in `pkg/controllers/run.go`
   - Need migration plan and deprecation timeline

## Remaining Work - Prioritized

### Priority 1: Fix Provider Creation Order (Low Risk) ⚡

**Problem:** Currently the CLI creates providers first, then Platform. This causes providers to reconcile before the Platform can add owner references, leading to a "WaitingForPlatform" phase initially.

**Solution:** Reorder the CLI to create Platform CR before provider CRs, or make providers wait longer before attempting reconciliation.

**Implementation:**
```go
// In pkg/build/build.go, reorder around line 320-354:

// Create Platform CR FIRST (before providers)
setupLog.V(1).Info("Creating platform resource")
if err := b.createPlatform(ctx, kubeClient); err != nil {
    if b.statusReporter != nil {
        b.statusReporter.FailStep("resources", err)
    }
    return fmt.Errorf("creating platform resource: %w", err)
}

// Then create providers (they will wait for Platform to add owner refs)
setupLog.V(1).Info("Creating giteaprovider resource")
if err := b.createGiteaProvider(ctx, kubeClient); err != nil {
    // ...
}

// ... continue with other providers
```

**Files to modify:**
- `pkg/build/build.go` - Reorder CR creation (1 line move, ~20 lines)

**Testing:**
- Verify Platform creates owner references quickly
- Verify providers transition from WaitingForPlatform → Installing → Ready
- Check that all components reach Ready state
- Run integration tests

**Timeline:** 1-2 days

**Risk:** Low - Simple reordering, no logic changes

---

### Priority 2: Move Bootstrap Repository Creation to Platform (Medium Risk) 🔧

**Problem:** Platform controller currently only aggregates status. It should also create the bootstrap GitRepository and ArgoCD Application CRs, which is currently done by the Localbuild controller.

**Background:** Bootstrap repositories are the initial Git repos for core components (argocd, gitea, nginx) that ArgoCD uses to manage the platform.

**Implementation Steps:**

1. **Add bootstrap creation to Platform controller** (`pkg/controllers/platform/platform_controller.go`)
   - Add `createBootstrapRepositories()` method
   - Call it after all providers are Ready (phase == "Ready")
   - Only create if they don't already exist

2. **Move code from Localbuild controller:**
   - Move `reconcileGitRepo()` from `pkg/controllers/localbuild/controller.go` (lines ~739-849)
   - Move `reconcileEmbeddedApp()` (lines ~362-415)
   - Move `ReconcileArgoAppsWithGitea()` (lines ~280-360)
   - Adapt to use duck-typed git provider instead of hardcoded Gitea

3. **Add RBAC permissions:**
```go
//+kubebuilder:rbac:groups=idpbuilder.cnoe.io,resources=gitrepositories,verbs=get;list;watch;create;update;patch
//+kubebuilder:rbac:groups=argoproj.io,resources=applications,verbs=get;list;watch;create;update;patch
```

**Files to modify:**
- `pkg/controllers/platform/platform_controller.go` - Add bootstrap creation (~200 lines)
- Move helper functions from localbuild controller

**Testing:**
- Verify GitRepository CRs are created when Platform becomes Ready
- Verify ArgoCD Applications are created
- Verify bootstrap apps sync successfully
- Check that Gitea repositories are populated
- Run e2e tests

**Timeline:** 1 week

**Risk:** Medium - Moving significant logic, but well-tested in Localbuild

---

### Priority 3: Custom Package Migration (Medium Risk) 📦

**Problem:** Custom packages are currently reconciled by the Localbuild controller. Need to determine the best approach for v1alpha2 architecture.

**Options:**

**Option A: Keep in Separate CustomPackage Controller**
- CustomPackage controller already exists in `pkg/controllers/custompackage/`
- Update it to discover git provider from Platform (duck-typing)
- Ensure it works independently of Localbuild

**Option B: Move to Platform Controller**
- Platform orchestrates custom package deployment
- Simpler dependency chain
- Platform can ensure providers are ready before creating custom packages

**Recommendation:** Option A - Keep CustomPackage controller independent but update it to use Platform for provider discovery.

**Implementation:**
1. Review existing CustomPackage controller
2. Update to use Platform CR for git provider discovery
3. Ensure proper ownership and lifecycle management
4. Add integration tests

**Files to modify:**
- `pkg/controllers/custompackage/` - Update provider discovery
- Tests for custom package scenarios

**Testing:**
- Verify custom packages work with v1alpha2 architecture
- Test with --package-dir, --package-file, --package-url flags
- Verify package priority ordering works
- Check ArgoCD Application creation

**Timeline:** 1 week

**Risk:** Medium - Custom packages have complex interactions with Git and ArgoCD

---

### Priority 4: Remove Localbuild CR Creation from CLI (Low Risk) 🗑️ ✅

**Status:** COMPLETED (January 2026)

**Problem:** CLI still creates Localbuild CR alongside v1alpha2 CRs, creating redundancy and potential confusion.

**Prerequisites:**
- Priority 2 complete (bootstrap repositories in Platform)
- Priority 3 complete (custom packages working)
- All integration tests passing with v1alpha2 path

**Implementation:** ✅ COMPLETED
- ✅ Removed Localbuild CR creation block (lines 281-318 in `pkg/build/build.go`)
- ✅ Removed `isCompatible()` function that checked existing Localbuild CR
- ✅ Removed `isBuildCustomizationSpecEqual()` helper function
- ✅ Cleaned up unused imports (`time`, `k8serrors`)
- ✅ Updated tests (removed `TestIsCompatible`)
- CLI now creates only v1alpha2 CRs: Platform, GiteaProvider, ArgoCDProvider, NginxGateway

**Files modified:**
- `pkg/build/build.go` - Removed Localbuild CR creation (~80 lines deleted total)
- `pkg/build/build_test.go` - Removed obsolete tests

**Testing:** ✅ Unit tests passing
- ✅ Package builds successfully
- ✅ Unit tests pass (`go test ./pkg/build/...`)
- ✅ All pkg tests pass (`go test ./pkg/...`)
- ✅ Code formatting checks pass
- ✅ Go vet checks pass
- ⏳ Full e2e test suite (pending)
- ⏳ Performance comparison (pending)

**Impact:**
- CLI creates only v1alpha2 CRs
- Localbuild controller remains registered for backward compatibility
- Existing Localbuild CRs can still be read/queried via `get` commands
- No breaking changes for users with existing installations

**Timeline:** Completed in 1 day

**Risk:** Low - Simple removal, well-isolated changes

---

### Priority 5: Deprecate Localbuild Controller (Low Risk) 📝

**Problem:** Localbuild controller is no longer needed once v1alpha2 architecture is fully operational.

**Prerequisites:**
- Priority 4 complete (CLI doesn't create Localbuild CRs)
- Multiple releases with v1alpha2 architecture stable
- Migration documentation complete

**Implementation Phases:**

**Phase A: Mark as Deprecated** (Release N)
1. Add deprecation warnings to Localbuild CRD
2. Update documentation stating Localbuild is deprecated
3. Add runtime warnings when Localbuild CR is created
4. Keep controller functional for migration period

**Phase B: Remove Controller** (Release N+2 or later)
1. Remove controller registration from `pkg/controllers/run.go`
2. Delete `pkg/controllers/localbuild/` directory
3. Keep CRD definition for data migration
4. Provide migration guide

**Files to modify:**
- `api/v1alpha1/localbuild_types.go` - Add deprecation notices
- `pkg/controllers/localbuild/controller.go` - Add deprecation warnings
- `pkg/controllers/run.go` - Eventually remove registration
- Documentation - Update with migration guide

**Timeline:** 
- Phase A: 1 week
- Phase B: 1 week (after 2+ releases)

**Risk:** Low - Controlled deprecation with long migration period

---

### Priority 6: Documentation Updates (Ongoing) 📚

**Continuous Updates Needed:**

1. **User Documentation**
   - Update getting started guides
   - Document v1alpha2 CRs and usage
   - Provide examples for each provider type
   - Migration guide from v1alpha1 to v1alpha2

2. **Developer Documentation**  
   - Architecture diagrams with current state
   - Controller interaction patterns
   - Testing guidelines
   - Adding new provider types

3. **API Documentation**
   - CRD field descriptions
   - Status field meanings
   - Provider interface contracts

**Files to update:**
- `docs/user/` - User guides
- `examples/v1alpha2/` - Example CRs
- `README.md` - Main getting started
- API documentation (godoc comments)

**Timeline:** Ongoing throughout implementation

**Risk:** Low - Documentation only

---

## Implementation Timeline

### Sprint 1 (Week 1-2)
- ✅ Priority 1: Fix provider creation order
- ✅ Start Priority 2: Platform bootstrap creation
- Documentation: Update architecture diagrams

### Sprint 2 (Week 3-4)
- ✅ Complete Priority 2: Platform bootstrap creation
- ✅ Priority 3: Custom package migration
- Testing: Integration tests for v1alpha2 path

### Sprint 3 (Week 5-6)
- ✅ Priority 4: Remove Localbuild CR creation
- ✅ Priority 5A: Mark Localbuild as deprecated
- Testing: Full e2e test suite
- Documentation: Migration guide

### Sprint 4+ (Week 7+)
- Monitoring and bug fixes
- Performance optimization
- Additional provider implementations (GitHub, GitLab, etc.)
- Release planning

### Future (After 2+ stable releases)
- ✅ Priority 5B: Remove Localbuild controller
- Cloud provider implementations (AWS, Azure, GCP)

**Total Timeline:** ~6-8 weeks for core migration (Priorities 1-4)

---

## Testing Strategy

### Unit Tests
- [x] Platform controller tests exist
- [x] Provider controller tests exist
- [ ] Bootstrap repository creation tests (Priority 2)
- [ ] Custom package integration tests (Priority 3)

### Integration Tests
- [x] Provider wait for owner reference
- [x] Platform aggregation of all provider types
- [ ] Full workflow: CLI → Platform → Providers → Bootstrap → Ready
- [ ] Custom packages with v1alpha2 architecture

### E2E Tests
- [ ] `idpbuilder create` with v1alpha2 only (no Localbuild)
- [ ] All core components operational
- [ ] Custom packages deployment
- [ ] Multi-provider scenarios

### Performance Tests
- [ ] Benchmark v1alpha2 vs v1alpha1 performance
- [ ] Resource utilization comparison
- [ ] Time to Ready comparison

---

## Success Criteria

- [ ] Platform controller creates bootstrap repositories
- [ ] Custom packages work with v1alpha2 architecture  
- [x] CLI creates only v1alpha2 CRs (no Localbuild) ✅ **COMPLETED**
- [ ] All integration tests pass
- [ ] All e2e tests pass
- [ ] Performance is equal or better than v1alpha1
- [ ] Documentation is complete and accurate
- [ ] Migration guide is available
- [ ] Localbuild controller marked as deprecated

---

## Risks and Mitigations

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|-----------|
| Breaking existing users | High | Low | Keep both paths working during transition, long deprecation period |
| Performance regression | Medium | Low | Benchmark and compare, optimize as needed |
| Missing functionality | High | Medium | Comprehensive feature parity checklist, thorough testing |
| Complex migration | Medium | Medium | Detailed step-by-step guide, automation where possible |
| Bootstrap creation issues | High | Low | Reuse well-tested code from Localbuild, extensive testing |
| Custom package breakage | Medium | Medium | Maintain backward compatibility, test all package scenarios |

---

## Reference Documentation

### Technical Specifications
- [Controller Architecture Specification](../specs/controller-architecture-spec.md) - Overall architecture design
- [Resource Creation Sequencing](../specs/resource-creation-sequencing.md) - State transitions and coordination
- [Client Architecture Specification](../specs/client-architecture-spec.md) - Future CLI enhancements

### Archived Implementation Docs
- [Archived Implementation Documentation](./archived/) - Historical planning documents

### Current Codebase
- `api/v1alpha2/` - v1alpha2 CRD definitions
- `pkg/controllers/platform/` - Platform controller
- `pkg/controllers/gitprovider/` - Git provider controllers
- `pkg/controllers/gatewayprovider/` - Gateway provider controllers
- `pkg/controllers/gitopsprovider/` - GitOps provider controllers
- `pkg/controllers/localbuild/` - Legacy Localbuild controller (to be deprecated)
- `pkg/build/build.go` - CLI CR creation logic
- `examples/v1alpha2/` - Example CRs

---

## Questions and Open Issues

### Open Questions
1. **Bootstrap timing:** Should Platform wait for providers to be fully Ready before creating bootstrap resources, or create them in parallel?
   - **Recommendation:** Wait for Ready to avoid race conditions

2. **Custom package controller:** Should it be independent or part of Platform?
   - **Recommendation:** Keep independent for separation of concerns

3. **Migration automation:** Should we provide a tool to migrate v1alpha1 to v1alpha2?
   - **Recommendation:** Not needed - v1alpha2 is new deployment, not migration

### Known Issues
- Platform creates owner references after providers start reconciling (Priority 1 addresses this)
- No e2e tests for v1alpha2-only path yet
- Bootstrap repository creation still in Localbuild controller

---

## Getting Help

For questions about this roadmap or the implementation:

1. Check the [technical specifications](../specs/) for design details
2. Review [archived implementation docs](./archived/) for historical context
3. Open a GitHub issue for discussion
4. Refer to the [Controller Architecture Spec](../specs/controller-architecture-spec.md) for architectural guidance

---

**Last Updated:** January 2026  
**Next Review:** After Priority 1 completion
