# IDP Builder Documentation

This directory contains all documentation for the IDP Builder project, organized into three main categories:

## 📋 Documentation Structure

### [Technical Specifications](./specs/)
Technical specifications and architectural design documents that define how the system should work.

- [**Specifications Overview**](./specs/README.md) - Index of all technical specifications
- [Controller-Based Architecture Specification](./specs/controller-architecture-spec.md) - Comprehensive spec for the v2 controller-based architecture (Dec 2025)
- [Hyperscaler Provider Implementation Specification](./specs/hyperscaler-provider-spec.md) - Cloud-native providers for AWS, Azure, GCP (Dec 2025)
- [Pluggable and Configurable Packaging Proposal](./specs/pluggable-packages.md) - Design for flexible package installation ✅ **IMPLEMENTED**

### [Implementation Documentation](./implementation/)
Implementation details, developer documentation, and testing information.

**🎯 Start Here:**
- [**Implementation Status**](./implementation/IMPLEMENTATION_STATUS.md) - Quick reference to project status and roadmap
- [**Phased Implementation Plan**](./implementation/phased-implementation-plan.md) - Complete roadmap for all specifications (24-36 weeks)

**Architecture Transition (v1alpha1 → v1alpha2):**
- [Architecture Transition Guide](./implementation/architecture-transition.md) - Visual guide with diagrams
- [Next Steps to Remove Localbuild](./implementation/next-steps-remove-localbuild.md) - Detailed implementation steps
- [Resource Creation Sequencing](./implementation/resource-creation-sequencing.md) - Controller coordination patterns
- [Phase 1-2 Final Status](./implementation/phase-1-2-final-status.md) - Current progress report

**Testing:**
- [Test Coverage Improvement Plan](./implementation/test-coverage-improvement-plan.md) - Strategy to improve test coverage
- [Test Duration Summary](./implementation/test-duration-summary.md) - Summary of test execution times
- [Test Timing Analysis](./implementation/test-timing-analysis.md) - Detailed test performance analysis

### [User Documentation](./user/)
User-facing guides and documentation for using IDP Builder.

- [Minimum Requirements](./user/minimum-requirements.md) - System requirements for running IDP Builder
- [Private Registry Authentication](./user/private-registries.md) - Guide for using private container registries

## 🚀 Quick Start

For getting started with IDP Builder, see the main [README.md](../README.md) in the repository root.

## 📚 Additional Resources

- [Examples](../examples/) - Example configurations and usage patterns
- [Scripts](../scripts/) - Utility scripts for development and testing
