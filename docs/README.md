# IDP Builder Documentation

This directory contains all documentation for the IDP Builder project, organized into three main categories:

## 📋 Documentation Structure

### [Technical Specifications](./specs/)
Technical specifications and architectural design documents that define how the system should work.

- [Client Architecture Specification](./specs/client-architecture-spec.md) - New unified client design replacing localbuilder and current CLI
- [Controller-Based Architecture Specification](./specs/controller-architecture-spec.md) - Comprehensive spec for the v2 controller-based architecture
- [Resource Creation Sequencing and State Transitions](./specs/resource-creation-sequencing.md) - Technical spec for resource tracking and sequencing
- [Pluggable and Configurable Packaging Proposal](./specs/pluggable-packages.md) - Design for flexible package installation
- [Hyperscaler Provider Implementation Specification](./specs/hyperscaler-provider-spec.md) - Cloud provider integration design

### [Implementation Documentation](./implementation/)
Implementation details, developer documentation, and testing information.

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
