# Phase 1.3 Controller Architecture Examples

This directory contains example YAML files for Phase 1.3 of the controller-based architecture implementation.

## Overview

Phase 1.3 introduces support for Gateway and GitOps providers in addition to the existing Git providers. This enables a complete platform deployment using Kubernetes Custom Resources.

## Provider Types

### Git Providers
- **GiteaProvider**: In-cluster Git server
- Exposes duck-typed fields: `endpoint`, `internalEndpoint`, `credentialsSecretRef`

### Gateway Providers
- **NginxGateway**: Nginx Ingress Controller for routing
- Exposes duck-typed fields: `ingressClassName`, `loadBalancerEndpoint`, `internalEndpoint`

### GitOps Providers
- **ArgoCDProvider**: ArgoCD for GitOps-based deployments
- Exposes duck-typed fields: `endpoint`, `internalEndpoint`, `credentialsSecretRef`

## Example Files

1. **giteaprovider.yaml**: Example GiteaProvider configuration
2. **nginxgateway.yaml**: Example NginxGateway configuration
3. **argocdprovider.yaml**: Example ArgoCDProvider configuration
4. **platform-full.yaml**: Complete Platform configuration referencing all three provider types

## Usage

### Deploy Individual Providers

```bash
# Deploy Gitea provider
kubectl apply -f giteaprovider.yaml

# Deploy Nginx gateway
kubectl apply -f nginxgateway.yaml

# Deploy ArgoCD provider
kubectl apply -f argocdprovider.yaml
```

### Deploy Complete Platform

```bash
# Deploy all providers and platform in one command
kubectl apply -f .
```

### Check Provider Status

```bash
# Check Gitea provider status
kubectl get giteaprovider gitea-local -n idpbuilder-system -o yaml

# Check Nginx gateway status
kubectl get nginxgateway nginx-gateway -n idpbuilder-system -o yaml

# Check ArgoCD provider status
kubectl get argocdprovider argocd -n idpbuilder-system -o yaml

# Check platform status (aggregates all provider statuses)
kubectl get platform localdev -n idpbuilder-system -o yaml
```

## Duck-Typed Interface

All providers implement duck-typed interfaces through their status fields. This means the Platform controller can work with any provider implementation without tight coupling.

### Git Provider Duck-Typed Fields
```yaml
status:
  endpoint: string              # External URL
  internalEndpoint: string      # Cluster-internal URL
  credentialsSecretRef:         # Admin credentials
    name: string
    namespace: string
    key: string
```

### Gateway Provider Duck-Typed Fields
```yaml
status:
  ingressClassName: string      # Ingress class name
  loadBalancerEndpoint: string  # External endpoint
  internalEndpoint: string      # Cluster-internal endpoint
```

### GitOps Provider Duck-Typed Fields
```yaml
status:
  endpoint: string              # External URL
  internalEndpoint: string      # Cluster-internal URL
  credentialsSecretRef:         # Admin credentials
    name: string
    namespace: string
    key: string
```

## Platform Status

The Platform CR aggregates the status of all referenced providers:

```yaml
status:
  phase: Ready
  conditions:
    - type: Ready
      status: "True"
      reason: AllProvidersReady
      message: All providers are ready
  providers:
    gitProviders:
      - name: gitea-local
        kind: GiteaProvider
        ready: true
    gateways:
      - name: nginx-gateway
        kind: NginxGateway
        ready: true
    gitOpsProviders:
      - name: argocd
        kind: ArgoCDProvider
        ready: true
```

## Next Steps

- See the [controller architecture spec](../../docs/controller-architecture-spec.md) for more details
- Check provider-specific documentation for advanced configuration options
- Review the duck-typing pattern for adding custom provider implementations
