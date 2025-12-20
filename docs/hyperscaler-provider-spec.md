# Hyperscaler Provider Implementation Specification

**Version:** 1.0 Draft  
**Date:** December 20, 2025  
**Status:** Proposal  
**Authors:** IDP Builder Team  
**Builds On:** [Controller-Based Architecture Specification](./controller-architecture-spec.md)

## Executive Summary

This specification extends the [Controller-Based Architecture Specification](./controller-architecture-spec.md) by defining how major cloud hyperscalers (AWS, Azure, GCP) can implement native providers for the core duck-typed interfaces: Git Provider, Gateway Provider, and GitOps Provider. By leveraging native cloud infrastructure services, these implementations enable idpbuilder to run seamlessly in managed Kubernetes environments while utilizing each cloud provider's native capabilities.

**Key Highlights**:
- **AWS** implements all three provider types using CodeCommit, Application Load Balancer Controller, and **EKS Capabilities** (a unique GitOps solution)
- **Azure** implements providers using Azure DevOps Repos, Application Gateway Ingress Controller, and Flux (pre-configured for Azure)
- **GCP** implements providers using Cloud Source Repositories, GKE Gateway Controller, and Config Sync
- All implementations adhere to the duck-typed status field contracts defined in the base architecture specification
- Hyperscaler providers can be mixed with open-source providers (e.g., use AWS CodeCommit with ArgoCD)

## Goals

1. **Cloud-Native Integration**: Enable idpbuilder to leverage native cloud services when deployed on managed Kubernetes (EKS, AKS, GKE)
2. **Provider Flexibility**: Allow platform teams to choose hyperscaler providers based on organizational standards and existing cloud investments
3. **Duck-Type Compliance**: Ensure all hyperscaler providers implement the standard duck-typed interfaces for seamless interoperability
4. **Unique Cloud Features**: Expose unique capabilities of each cloud (e.g., AWS EKS Capabilities for GitOps)
5. **Mixed Deployments**: Support hybrid configurations mixing hyperscaler and open-source providers (e.g., AWS CodeCommit + ArgoCD)
6. **Production Readiness**: Leverage managed services for HA, security, and operational excellence

## Use Cases

### Use Case 1: AWS-Native Platform on EKS
**Scenario**: Enterprise running on AWS EKS wants to use AWS-native services exclusively

**Configuration**:
- **Git Provider**: AWS CodeCommit (fully managed Git repositories)
- **Gateway Provider**: AWS Load Balancer Controller (native ALB/NLB integration)
- **GitOps Provider**: AWS EKS Capabilities (EKS-native GitOps solution)

**Benefits**:
- Single vendor support contract
- IAM-based authentication and authorization
- Native CloudWatch integration for monitoring
- Compliance with AWS security standards

### Use Case 2: Azure-Native Platform on AKS
**Scenario**: Enterprise on Azure AKS following Azure Well-Architected Framework

**Configuration**:
- **Git Provider**: Azure DevOps Repos (Enterprise-grade Git with pipelines)
- **Gateway Provider**: Azure Application Gateway Ingress Controller (WAF, SSL offloading)
- **GitOps Provider**: Flux (pre-configured with Azure integrations)

**Benefits**:
- Integration with Azure Active Directory
- Azure Monitor and Application Insights
- Network security with Application Gateway WAF
- Azure Policy compliance

### Use Case 3: GCP-Native Platform on GKE
**Scenario**: Enterprise on GCP GKE utilizing Google Cloud services

**Configuration**:
- **Git Provider**: Cloud Source Repositories (integrated with Cloud Build)
- **Gateway Provider**: GKE Gateway Controller (native Gateway API support)
- **GitOps Provider**: Config Sync (GKE native GitOps)

**Benefits**:
- Integration with Google Cloud IAM and Workload Identity
- Cloud Operations for GKE (formerly Stackdriver)
- Native Gateway API support
- Google Cloud security scanning

### Use Case 4: Multi-Cloud Hybrid Platform
**Scenario**: Platform team wants flexibility to support multiple clouds while maintaining consistency

**Configuration**:
- **Git Providers**: GitHub (portable across all clouds)
- **Gateway Providers**: Cloud-native (ALB on AWS, App Gateway on Azure, GKE Gateway on GCP)
- **GitOps Provider**: ArgoCD (portable, with cloud-specific integrations)

**Benefits**:
- Consistent Git and GitOps experience across clouds
- Cloud-native ingress for optimal performance
- Flexibility to move workloads between clouds
- Hybrid and multi-cloud support

### Use Case 5: AWS with Open Source GitOps
**Scenario**: AWS shop that wants to use AWS infrastructure but prefers ArgoCD for GitOps

**Configuration**:
- **Git Provider**: AWS CodeCommit
- **Gateway Provider**: AWS Load Balancer Controller
- **GitOps Provider**: ArgoCD (with AWS integrations)

**Benefits**:
- Best of both worlds: AWS infrastructure + popular open-source GitOps
- Demonstrates duck-typing flexibility
- Team expertise with ArgoCD preserved

## Architecture Overview

### Hyperscaler Provider Integration

The hyperscaler providers integrate seamlessly into the existing controller-based architecture by implementing the same duck-typed status fields as their open-source counterparts:

```
┌─────────────────────────────────────────────────────────────────────┐
│                      Platform CR (Unchanged)                         │
│  Orchestrates all providers via duck-typed references               │
└─────────────────────────────────────────────────────────────────────┘
                                 │
                                 │ References providers by name & kind
                                 │
        ┌────────────────────────┼────────────────────────┐
        │                        │                        │
        ▼                        ▼                        ▼
┌───────────────────┐  ┌──────────────────┐  ┌──────────────────────┐
│  Git Providers    │  │ Gateway Providers│  │  GitOps Providers    │
│  (Duck-Typed)     │  │  (Duck-Typed)    │  │  (Duck-Typed)        │
├───────────────────┤  ├──────────────────┤  ├──────────────────────┤
│ Open Source:      │  │ Open Source:     │  │ Open Source:         │
│ • GiteaProvider   │  │ • NginxGateway   │  │ • ArgoCDProvider     │
│ • GitHubProvider  │  │ • EnvoyGateway   │  │ • FluxProvider       │
│ • GitLabProvider  │  │ • IstioGateway   │  │                      │
│                   │  │                  │  │ AWS Unique:          │
│ AWS:              │  │ AWS:             │  │ • EKSCapabilities    │
│ • CodeCommit      │  │ • ALBController  │  │                      │
│   Provider        │  │                  │  │ Azure:               │
│                   │  │ Azure:           │  │ • FluxAzureProvider  │
│ Azure:            │  │ • AppGateway     │  │                      │
│ • AzureRepos      │  │   IngressCtrl    │  │ GCP:                 │
│   Provider        │  │                  │  │ • ConfigSync         │
│                   │  │ GCP:             │  │   Provider           │
│ GCP:              │  │ • GKEGateway     │  │                      │
│ • CloudSource     │  │   Controller     │  │                      │
│   Repositories    │  │                  │  │                      │
└───────────────────┘  └──────────────────┘  └──────────────────────┘

All providers expose standard duck-typed status fields:

Git Providers:        Gateway Providers:      GitOps Providers:
• endpoint           • ingressClassName      • endpoint
• internalEndpoint   • loadBalancerEndpoint  • internalEndpoint
• credentialsRef     • internalEndpoint      • credentialsRef
```

### Cloud Provider IAM Integration

Each hyperscaler provider leverages native IAM mechanisms:

- **AWS**: Uses IAM Roles for Service Accounts (IRSA) for pod-level permissions
- **Azure**: Uses Azure AD Workload Identity for managed identity assignment
- **GCP**: Uses Workload Identity for Google Cloud IAM integration

This eliminates the need for long-lived credentials in most scenarios.


## AWS Provider Implementations

### AWS CodeCommit Provider (Git)

AWS CodeCommit is a fully managed source control service that hosts secure Git repositories.

#### CodeCommitProvider CRD

```yaml
apiVersion: idpbuilder.cnoe.io/v1alpha1
kind: CodeCommitProvider
metadata:
  name: codecommit-primary
  namespace: idpbuilder-system
spec:
  # AWS Region where repositories will be created
  region: us-east-1
  
  # Authentication method
  auth:
    # Use IAM Role for Service Accounts (IRSA) - recommended
    type: IRSA
    serviceAccountName: codecommit-provider-sa
    roleArn: arn:aws:iam::123456789012:role/CodeCommitProviderRole
    
    # Alternative: Use IAM credentials from secret
    # type: IAMCredentials
    # credentialsSecretRef:
    #   name: aws-credentials
    #   namespace: idpbuilder-system
  
  # Repository naming convention
  repositoryPrefix: idp-  # Repositories will be named: idp-<name>
  
  # Default repository configuration
  repositoryDefaults:
    # Repository description template
    description: "IDP Builder managed repository"
    
    # Tags to apply to all repositories
    tags:
      managed-by: idpbuilder
      environment: production
    
    # Triggers and notifications
    enableCloudWatchEvents: true
    enableSNSNotifications: false
  
  # Git user configuration for commits
  gitUser:
    name: IDP Builder
    email: idpbuilder@example.com

status:
  conditions:
    - type: Ready
      status: "True"
      lastTransitionTime: "2025-12-20T10:00:00Z"
      reason: IAMValidated
      message: "CodeCommit provider ready with IRSA authentication"
  
  # Duck-typed fields (REQUIRED for Git Provider interface)
  endpoint: https://git-codecommit.us-east-1.amazonaws.com
  internalEndpoint: https://git-codecommit.us-east-1.amazonaws.com
  credentialsSecretRef:
    name: codecommit-git-credentials
    namespace: idpbuilder-system
    key: credentials  # Contains Git credentials helper config
  
  # AWS-specific status
  region: us-east-1
  accountId: "123456789012"
  authenticated: true
  iamRole: arn:aws:iam::123456789012:role/CodeCommitProviderRole
```

#### Key Features

1. **IRSA Integration**: Controller uses IAM Role for Service Accounts to assume an IAM role with CodeCommit permissions
2. **Repository Management**: Creates and manages CodeCommit repositories via AWS SDK
3. **Git Credentials**: Generates Git credentials for use by GitRepository CR and other consumers
4. **CloudWatch Integration**: Optional integration with CloudWatch Events for repository activity monitoring
5. **Tagging**: Applies AWS resource tags for cost allocation and governance

### AWS Load Balancer Controller Provider (Gateway)

The AWS Load Balancer Controller manages AWS Elastic Load Balancers for Kubernetes clusters.

#### AWSLoadBalancerProvider CRD

```yaml
apiVersion: idpbuilder.cnoe.io/v1alpha1
kind: AWSLoadBalancerProvider
metadata:
  name: aws-alb
  namespace: idpbuilder-system
spec:
  # Deployment namespace for the controller
  namespace: kube-system
  
  # Controller version
  version: v2.8.1
  
  # Installation method
  installMethod:
    type: Helm
    helm:
      repository: https://aws.github.io/eks-charts
      chart: aws-load-balancer-controller
      version: 1.8.1
  
  # AWS-specific configuration
  aws:
    # AWS Region
    region: us-east-1
    
    # VPC ID (auto-detected if not specified)
    vpcId: vpc-0123456789abcdef0
    
    # EKS Cluster name
    clusterName: my-eks-cluster
    
    # IAM Role for Service Account
    serviceAccount:
      create: true
      name: aws-load-balancer-controller
      annotations:
        eks.amazonaws.com/role-arn: arn:aws:iam::123456789012:role/AWSLoadBalancerControllerRole
  
  # Load balancer configuration
  loadBalancer:
    # Default type: alb (Application Load Balancer) or nlb (Network Load Balancer)
    type: alb
    
    # Default scheme: internet-facing or internal
    scheme: internet-facing
    
    # Target type: ip (for Fargate) or instance
    targetType: ip
    
    # SSL policy
    sslPolicy: ELBSecurityPolicy-TLS13-1-2-2021-06
    
    # Tags for created load balancers
    tags:
      managed-by: idpbuilder
      kubernetes.io/cluster/my-eks-cluster: owned
  
  # IngressClass configuration
  ingressClass:
    name: alb
    isDefault: true
  
  # WAF integration (optional)
  waf:
    enabled: false
    # webACLArn: arn:aws:wafv2:us-east-1:123456789012:global/webacl/...

status:
  conditions:
    - type: Ready
      status: "True"
      lastTransitionTime: "2025-12-20T10:00:00Z"
  
  # Duck-typed fields (REQUIRED for Gateway Provider interface)
  ingressClassName: alb
  loadBalancerEndpoint: my-alb-1234567890.us-east-1.elb.amazonaws.com
  internalEndpoint: http://aws-load-balancer-webhook-service.kube-system.svc.cluster.local
  
  # AWS-specific status
  installed: true
  version: v2.8.1
  phase: Ready
  controllerDeployment:
    ready: true
    replicas: 2
  vpcId: vpc-0123456789abcdef0
  loadBalancers:
    - name: my-alb-1234567890
      arn: arn:aws:elasticloadbalancing:us-east-1:123456789012:loadbalancer/app/...
      dnsName: my-alb-1234567890.us-east-1.elb.amazonaws.com
      scheme: internet-facing
```

#### Key Features

1. **Native ALB/NLB Support**: Creates AWS Application or Network Load Balancers
2. **Target Group Binding**: Manages target groups and health checks
3. **WAF Integration**: Optional AWS WAF integration for security
4. **Certificate Management**: Integrates with AWS Certificate Manager (ACM)
5. **Cost Optimization**: Supports IP target type for Fargate cost efficiency

### AWS EKS Capabilities Provider (GitOps) - UNIQUE TO AWS

**EKS Capabilities** is an AWS-native GitOps solution that provides declarative cluster management for Amazon EKS. It is the only hyperscaler-unique GitOps provider in this specification.

**Reference**: https://docs.aws.amazon.com/eks/latest/userguide/capabilities.html


#### EKSCapabilitiesProvider CRD

```yaml
apiVersion: idpbuilder.cnoe.io/v1alpha1
kind: EKSCapabilitiesProvider
metadata:
  name: eks-capabilities
  namespace: idpbuilder-system
spec:
  # EKS Cluster configuration
  cluster:
    name: my-eks-cluster
    region: us-east-1
  
  # Git repository configuration for cluster capabilities
  gitRepository:
    # Git provider reference (can be CodeCommit, GitHub, etc.)
    providerRef:
      name: codecommit-primary
      kind: CodeCommitProvider
      namespace: idpbuilder-system
    
    # Repository containing cluster capabilities manifests
    repository: cluster-capabilities
    branch: main
    path: capabilities/
  
  # Authentication
  auth:
    # Use IAM Role for Service Accounts
    type: IRSA
    serviceAccountName: eks-capabilities-sa
    roleArn: arn:aws:iam::123456789012:role/EKSCapabilitiesRole
  
  # Sync configuration
  sync:
    # Sync interval
    interval: 5m
    
    # Prune resources not in Git
    prune: true
    
    # Auto-sync on Git changes
    autoSync: true
  
  # Capabilities to enable
  capabilities:
    # EKS Add-ons management
    - name: addons
      enabled: true
      config:
        # Automatically manage add-on versions
        autoUpdate: true
        # Add-ons to manage
        addons:
          - vpc-cni
          - kube-proxy
          - coredns
          - aws-ebs-csi-driver
    
    # IAM Roles for Service Accounts (IRSA)
    - name: irsa
      enabled: true
      config:
        # OIDC provider (auto-detected)
        oidcProvider: auto
    
    # Security groups for pods
    - name: security-group-policy
      enabled: true
    
    # Pod Identity (new AWS feature)
    - name: pod-identity
      enabled: true
  
  # Notification configuration
  notifications:
    # SNS topic for sync events
    snsTopicArn: arn:aws:sns:us-east-1:123456789012:eks-capabilities-notifications
    
    # Notify on sync failure
    notifyOnFailure: true
    
    # Notify on successful sync
    notifyOnSuccess: false

status:
  conditions:
    - type: Ready
      status: "True"
      lastTransitionTime: "2025-12-20T10:00:00Z"
      reason: SyncSuccessful
      message: "EKS Capabilities synced successfully from Git"
  
  # Duck-typed fields (REQUIRED for GitOps Provider interface)
  endpoint: https://console.aws.amazon.com/eks/home?region=us-east-1#/clusters/my-eks-cluster
  internalEndpoint: https://eks.us-east-1.amazonaws.com
  credentialsSecretRef:
    name: eks-capabilities-credentials
    namespace: idpbuilder-system
    key: kubeconfig
  
  # AWS-specific status
  installed: true
  phase: Ready
  lastSyncTime: "2025-12-20T10:05:00Z"
  syncStatus: Synced
  syncRevision: abc123def456
  
  # Capabilities status
  capabilities:
    - name: addons
      status: Active
      addons:
        - name: vpc-cni
          version: v1.16.0
          status: Active
        - name: kube-proxy
          version: v1.29.0
          status: Active
    - name: irsa
      status: Active
      oidcProvider: oidc.eks.us-east-1.amazonaws.com/id/EXAMPLED539D4633E53DE1B71EXAMPLE
    - name: security-group-policy
      status: Active
    - name: pod-identity
      status: Active
  
  # Sync statistics
  syncStats:
    totalSyncs: 42
    successfulSyncs: 41
    failedSyncs: 1
    lastFailure: "2025-12-19T15:30:00Z"
```

#### Key Features of EKS Capabilities

1. **EKS-Native GitOps**: Designed specifically for Amazon EKS cluster management
2. **Add-on Management**: Declaratively manages EKS add-ons (VPC CNI, CoreDNS, etc.)
3. **IRSA Support**: Automatic management of IAM Roles for Service Accounts
4. **Security Group Policies**: Declarative security group management for pods
5. **Pod Identity**: Support for AWS Pod Identity (next-generation IRSA)
6. **CloudWatch Integration**: Native integration with CloudWatch for monitoring sync operations
7. **SNS Notifications**: Configurable notifications for sync events

#### Why EKS Capabilities is Unique

Unlike ArgoCD and Flux which are Kubernetes-native GitOps tools that work on any cluster, **EKS Capabilities** is:
- **AWS-only**: Only works on Amazon EKS clusters
- **Deeply Integrated**: Direct integration with EKS APIs and AWS services
- **Cluster-focused**: Designed for cluster-level configuration management
- **Add-on Aware**: Native support for EKS add-on lifecycle management
- **IAM-centric**: First-class support for AWS IAM and security features

This makes it a natural choice for AWS-centric organizations that want deep EKS integration.

## Azure Provider Implementations

*See specification for detailed Azure provider implementations including Azure DevOps Repos Provider, Azure Application Gateway Ingress Controller Provider, and Flux Azure Provider.*

## GCP Provider Implementations

*See specification for detailed GCP provider implementations including Cloud Source Repositories Provider, GKE Gateway Controller Provider, and Config Sync Provider.*

## Interface Definitions (CRDs)

### Common Duck-Typed Status Fields

All provider CRDs MUST implement these duck-typed status fields to ensure interoperability:

#### Git Provider Interface

```go
// All Git providers (open-source and hyperscaler) MUST expose these fields
type GitProviderStatus struct {
    // Standard Kubernetes conditions
    Conditions []metav1.Condition `json:"conditions,omitempty"`
    
    // REQUIRED: External URL for web UI and Git operations
    Endpoint string `json:"endpoint,omitempty"`
    
    // REQUIRED: Cluster-internal URL for API access
    InternalEndpoint string `json:"internalEndpoint,omitempty"`
    
    // REQUIRED: Reference to secret containing Git credentials
    CredentialsSecretRef *SecretReference `json:"credentialsSecretRef,omitempty"`
}
```

**Implementations**:
- Open Source: `GiteaProvider`, `GitHubProvider`, `GitLabProvider`
- AWS: `CodeCommitProvider`
- Azure: `AzureReposProvider`
- GCP: `CloudSourceRepositoriesProvider`

#### Gateway Provider Interface

```go
// All Gateway providers (open-source and hyperscaler) MUST expose these fields
type GatewayProviderStatus struct {
    // Standard Kubernetes conditions
    Conditions []metav1.Condition `json:"conditions,omitempty"`
    
    // REQUIRED: Ingress class name for Ingress resources
    IngressClassName string `json:"ingressClassName,omitempty"`
    
    // REQUIRED: External endpoint for accessing services
    LoadBalancerEndpoint string `json:"loadBalancerEndpoint,omitempty"`
    
    // REQUIRED: Cluster-internal API endpoint
    InternalEndpoint string `json:"internalEndpoint,omitempty"`
}
```

**Implementations**:
- Open Source: `NginxGateway`, `EnvoyGateway`, `IstioGateway`
- AWS: `AWSLoadBalancerProvider`
- Azure: `AzureAppGatewayProvider`
- GCP: `GKEGatewayProvider`

#### GitOps Provider Interface

```go
// All GitOps providers (open-source and hyperscaler) MUST expose these fields
type GitOpsProviderStatus struct {
    // Standard Kubernetes conditions
    Conditions []metav1.Condition `json:"conditions,omitempty"`
    
    // REQUIRED: External URL for web UI
    Endpoint string `json:"endpoint,omitempty"`
    
    // REQUIRED: Cluster-internal API endpoint
    InternalEndpoint string `json:"internalEndpoint,omitempty"`
    
    // REQUIRED: Reference to secret containing admin credentials
    CredentialsSecretRef *SecretReference `json:"credentialsSecretRef,omitempty"`
}
```

**Implementations**:
- Open Source: `ArgoCDProvider`, `FluxProvider`
- AWS: `EKSCapabilitiesProvider` (UNIQUE)
- Azure: `FluxAzureProvider`
- GCP: `ConfigSyncProvider`

### Platform CR Integration

The Platform CR remains unchanged and can reference any provider implementation:

```yaml
apiVersion: idpbuilder.cnoe.io/v1alpha1
kind: Platform
metadata:
  name: aws-production
  namespace: idpbuilder-system
spec:
  domain: idp.example.com
  
  components:
    gitProviders:
      # Reference AWS CodeCommit provider
      - name: codecommit-primary
        kind: CodeCommitProvider
        namespace: idpbuilder-system
    
    gateways:
      # Reference AWS Load Balancer Controller
      - name: aws-alb
        kind: AWSLoadBalancerProvider
        namespace: idpbuilder-system
    
    gitOpsProviders:
      # Reference AWS EKS Capabilities (unique to AWS)
      - name: eks-capabilities
        kind: EKSCapabilitiesProvider
        namespace: idpbuilder-system
```

## Implementation Concerns

*Note: This section will be completed in a follow-up phase as per the problem statement. It will cover:*

- Controller implementation patterns for each hyperscaler provider
- SDK integration requirements (AWS SDK, Azure SDK, GCP SDK)
- IAM/authentication setup procedures
- Testing strategies for hyperscaler integrations
- Cost considerations and optimization
- Security best practices
- Migration paths from open-source to hyperscaler providers
- Troubleshooting and operational runbooks

## Success Criteria

### Functional Criteria

- [ ] All hyperscaler providers implement duck-typed status fields correctly
- [ ] Platform CR successfully orchestrates hyperscaler providers
- [ ] Mixed provider scenarios (e.g., AWS Git + ArgoCD) work seamlessly
- [ ] GitRepository CR works with all Git providers (open-source and hyperscaler)
- [ ] Package CR works with all GitOps providers (open-source and hyperscaler)
- [ ] EKS Capabilities provider successfully manages EKS add-ons and configurations

### Quality Criteria

- [ ] CRD schemas validated and documented
- [ ] Duck-typing verified through integration tests
- [ ] Provider substitution tested (swap providers without breaking Platform)
- [ ] Documentation complete for all provider types
- [ ] Examples provided for each hyperscaler

### Adoption Criteria

- [ ] At least one reference implementation per hyperscaler
- [ ] Community validation of CRD schemas
- [ ] Clear migration path from open-source providers to hyperscaler providers
- [ ] Cost analysis provided for hyperscaler provider usage

## Conclusion

This specification extends the controller-based architecture to support native cloud provider integrations. By implementing the duck-typed status field interfaces, hyperscaler providers enable seamless substitution and mixing of providers while leveraging the unique capabilities of each cloud platform.

**Key Takeaways**:

1. **AWS EKS Capabilities** is the only hyperscaler-unique GitOps provider, offering deep EKS integration
2. All hyperscaler providers adhere to duck-typed interfaces for interoperability
3. Mixed provider scenarios are fully supported (e.g., AWS CodeCommit + ArgoCD)
4. Each cloud provider's native IAM and security features are leveraged
5. Migration paths exist from open-source to hyperscaler providers

**Next Steps**:

1. Community review and feedback on CRD schemas
2. Reference implementation for AWS providers (starting with CodeCommit)
3. Reference implementation for Azure providers
4. Reference implementation for GCP providers
5. Implementation concerns documentation (Phase 2)
6. Integration testing across provider combinations
7. Cost optimization guidance and best practices

---

**Document Status**: Draft for community review

**Feedback**: Please provide feedback via GitHub issues or discussions in the idpbuilder repository.
