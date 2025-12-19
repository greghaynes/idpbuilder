package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PlatformSpec defines the desired state of Platform
type PlatformSpec struct {
	// Domain is the base domain for all platform services
	// +optional
	Domain string `json:"domain,omitempty"`

	// IngressConfig configures ingress/gateway behavior
	// +optional
	IngressConfig IngressConfig `json:"ingressConfig,omitempty"`

	// Components specifies the platform components and their provider references
	Components ComponentsSpec `json:"components,omitempty"`

	// Bootstrap specifies GitOps bootstrap configuration
	// +optional
	Bootstrap BootstrapSpec `json:"bootstrap,omitempty"`
}

// IngressConfig defines ingress configuration
type IngressConfig struct {
	// Provider specifies which gateway provider to use (e.g., "nginx")
	// +optional
	Provider string `json:"provider,omitempty"`

	// UsePathRouting enables path-based routing instead of subdomain routing
	// +optional
	UsePathRouting bool `json:"usePathRouting,omitempty"`

	// TLSSecretRef references a secret containing TLS certificates
	// +optional
	TLSSecretRef *SecretReference `json:"tlsSecretRef,omitempty"`
}

// ComponentsSpec defines all platform components
type ComponentsSpec struct {
	// GitProviders lists references to Git provider CRs
	// +optional
	GitProviders []ProviderReference `json:"gitProviders,omitempty"`

	// Gateways lists references to Gateway provider CRs
	// +optional
	Gateways []ProviderReference `json:"gateways,omitempty"`

	// GitOpsProviders lists references to GitOps provider CRs
	// +optional
	GitOpsProviders []ProviderReference `json:"gitOpsProviders,omitempty"`
}

// ProviderReference references a provider CR by name, kind, and namespace
type ProviderReference struct {
	// Name of the provider CR
	Name string `json:"name"`

	// Kind of the provider CR (e.g., "GiteaProvider", "NginxGateway")
	Kind string `json:"kind"`

	// Namespace of the provider CR
	// +optional
	Namespace string `json:"namespace,omitempty"`
}

// BootstrapSpec defines GitOps bootstrap configuration
type BootstrapSpec struct {
	// GitServerRef references the git server to use for bootstrap repositories
	// +optional
	GitServerRef *ProviderReference `json:"gitServerRef,omitempty"`

	// Repositories lists bootstrap repositories to create
	// +optional
	Repositories []BootstrapRepository `json:"repositories,omitempty"`
}

// BootstrapRepository defines a bootstrap repository
type BootstrapRepository struct {
	// Name of the repository
	Name string `json:"name"`

	// Path to the content to sync into the repository
	Path string `json:"path"`

	// AutoSync enables automatic syncing of repository content
	// +optional
	AutoSync bool `json:"autoSync,omitempty"`
}

// SecretReference references a secret by name, namespace, and key
type SecretReference struct {
	// Name of the secret
	Name string `json:"name"`

	// Namespace of the secret
	// +optional
	Namespace string `json:"namespace,omitempty"`

	// Key within the secret
	// +optional
	Key string `json:"key,omitempty"`
}

// PlatformStatus defines the observed state of Platform
type PlatformStatus struct {
	// Conditions represent the latest available observations of the platform's state
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// Providers aggregates status from all provider CRs
	// +optional
	Providers ProvidersStatus `json:"providers,omitempty"`

	// ObservedGeneration reflects the generation of the most recently observed Platform
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Phase represents the current phase of the platform (Pending, Ready, Error)
	// +optional
	Phase string `json:"phase,omitempty"`
}

// ProvidersStatus aggregates status from provider CRs
type ProvidersStatus struct {
	// GitProviders lists status of Git provider CRs
	// +optional
	GitProviders []ProviderStatus `json:"gitProviders,omitempty"`

	// Gateways lists status of Gateway provider CRs
	// +optional
	Gateways []ProviderStatus `json:"gateways,omitempty"`

	// GitOpsProviders lists status of GitOps provider CRs
	// +optional
	GitOpsProviders []ProviderStatus `json:"gitOpsProviders,omitempty"`
}

// ProviderStatus represents the status of a provider CR
type ProviderStatus struct {
	// Name of the provider CR
	Name string `json:"name"`

	// Kind of the provider CR
	Kind string `json:"kind"`

	// Ready indicates if the provider is ready
	Ready bool `json:"ready"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced

// Platform is the Schema for the platforms API
type Platform struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PlatformSpec   `json:"spec,omitempty"`
	Status PlatformStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PlatformList contains a list of Platform
type PlatformList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Platform `json:"items"`
}
