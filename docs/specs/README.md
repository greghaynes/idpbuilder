# Technical Specifications

This directory contains technical specifications and architectural design documents for IDP Builder.

## Documents

### [Client Architecture Specification](./client-architecture-spec.md)

**Status:** Proposal  
**Version:** 1.0 Draft  
**Date:** December 23, 2024

A comprehensive specification for a new unified client architecture that replaces the current localbuilder implementation and CLI. The new client provides three primary capabilities:

- **Infrastructure Management** - Provisions local/remote Kubernetes infrastructure with pluggable providers (kind, k3s, external)
- **Flavor Selection** - Enables selection of pre-defined platform configurations ("flavors") that generate appropriate Custom Resources
- **Status Monitoring** - Watches Platform resources and provides real-time, user-friendly status information

**Key Topics:**
- Modular client component design (Infrastructure Manager, Flavor Manager, Status Watcher)
- Flavor definition format and built-in flavors
- New CLI command structure and UX improvements
- Iterative implementation plan with 6 phases
- Migration strategy from current implementation
- Security and performance considerations

**Goals:**
- Clear separation of concerns between infrastructure, configuration, and monitoring
- Extensible flavor system for different deployment scenarios
- Excellent user experience with real-time feedback
- Support for both development and production use cases

**See also:**
- [Controller Architecture Spec](./controller-architecture-spec.md) - Controller design that client orchestrates
- [Resource Creation Sequencing](./resource-creation-sequencing.md) - How resources are tracked and sequenced
- [Implementation Roadmap](../implementation/implementation-roadmap.md) - Current implementation status and remaining work

### [Resource Creation Sequencing and State Transitions](./resource-creation-sequencing.md)

**Status:** Technical Specification  
**Version:** 1.0  
**Date:** December 22, 2024

Comprehensive technical specification for resource creation sequencing and state transitions in the idpbuilder controller-based architecture.

**Key Topics:**
- Resource creation sequence (CLI → Platform → Providers)
- Controller-driven sequencing and state machines
- Owner reference pattern for coordination
- Status aggregation and monitoring
- Cross-provider dependencies
- Configuration discovery patterns
- Client resource tracking implementation
- Error handling and retry logic
- Duck-typing for provider independence

**Use Cases:**
- Understanding controller coordination patterns
- Implementing new providers
- Debugging state transition issues
- Learning the controller-based architecture

**See also:**
- [Controller Architecture Spec](./controller-architecture-spec.md) - High-level architecture design
- [Implementation Roadmap](../implementation/implementation-roadmap.md) - Current implementation status and remaining work

### [Controller-Based Architecture Specification](./controller-architecture-spec.md)

**Status:** Proposal  
**Version:** 1.0 Draft  
**Date:** December 19, 2025

A comprehensive specification for evolving idpbuilder from a CLI-driven installation model to a controller-based architecture. This architectural change enables:

- Kubernetes-native management through Custom Resources
- Separation between CLI (infrastructure provisioning) and controllers (platform reconciliation)
- Support for both development (CLI-driven) and production (GitOps-driven) deployment modes
- Duck-typed provider interfaces for Git, Gateway, and GitOps components
- Pluggable provider architecture (Gitea/GitHub/GitLab, Nginx/Envoy/Istio, ArgoCD/Flux)

**Key Topics:**
- Platform CR and provider CRs design
- Duck-typing pattern for provider independence
- Phased implementation plan
- Migration strategy from v1alpha1 to v1alpha2

### [Pluggable and Configurable Packaging Proposal](./pluggable-packages.md)

**Status:** Implemented  
**Date:** Original proposal

A design document outlining the approach for making packages installed by idpbuilder configurable and pluggable.

**Key Topics:**
- ArgoCD-based package management
- In-cluster Git server (Gitea) for GitOps workflows
- Runtime content generation
- Support for Helm charts, Kustomize, and raw manifests
- Local file handling for development workflows

**Goals:**
- Make packages configurable without recompiling
- Minimize runtime dependencies
- Enable fast local development feedback loops
- Support imperative pipelines via ArgoCD resource hooks

### [Validation and Testing Strategy](./validation-testing-strategy.md)

**Status:** Proposal  
**Version:** 1.0  
**Date:** March 2026

A specification for a multi-tier validation and testing strategy that can be run remotely in GitHub Codespaces while providing significant validation that idpbuilder meets its functional requirements.

**Key Topics:**
- Four-tier validation pyramid (static analysis, binary smoke, functional/integration, e2e)
- Mapping of functional requirements to test tiers
- `make validate` target for fast Codespaces-friendly validation (< 5 minutes)
- `make smoke` target for binary build verification
- Test writing guidelines for controllers, CLI, and e2e scenarios
- GitHub Actions workflow for automated Codespaces-style CI

**Goals:**
- Fill the gap between fast unit tests and heavy e2e tests
- Enable meaningful validation in Codespaces without a real Kubernetes cluster
- Align developer workflow with CI/CD pipeline validation

**See also:**
- [Test Coverage Improvement Plan](../implementation/test-coverage-improvement-plan.md) - Coverage gap analysis
- [Implementation Roadmap](../implementation/implementation-roadmap.md) - Current implementation status

### [Hyperscaler Provider Implementation Specification](./hyperscaler-provider-spec.md)

**Status:** Proposal  
**Version:** 1.0 Draft  
**Date:** December 20, 2025

Extends the controller-based architecture specification to describe how major cloud hyperscalers (AWS, Azure, GCP) can implement native providers for the core duck-typed interfaces: Git Provider, Gateway Provider, and GitOps Provider.

**Key Topics:**
- AWS provider implementations (CodeCommit, ALB Controller, EKS Capabilities)
- Azure provider implementations (Azure Repos, App Gateway, Flux Azure)
- GCP provider implementations (Cloud Source Repos, GKE Gateway, Config Sync)
- Duck-typed status field contracts for provider interoperability
- Mixed provider scenarios (e.g., AWS infrastructure with ArgoCD)
- Cloud IAM integration (IRSA, Azure AD Workload Identity, GCP Workload Identity)
- Cost considerations and authentication setup guides

**Goals:**
- Enable cloud-native integration with managed Kubernetes services
- Leverage native cloud services for Git, Gateway, and GitOps
- Maintain provider flexibility through duck-typing
- Support hybrid and multi-cloud deployments

## Purpose

These specifications serve as:

1. **Design Documentation** - Detailed explanation of architectural decisions and rationale
2. **Implementation Guides** - Reference for developers implementing features
3. **Historical Record** - Documentation of why certain design choices were made
4. **Community Input** - Proposals for feedback and discussion

## Related Documentation

- [Implementation Documentation](../implementation/) - Developer docs and testing information
- [User Documentation](../user/) - User-facing guides and how-tos
