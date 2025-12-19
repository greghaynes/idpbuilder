package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ArgoCDProviderSpec defines the desired state of ArgoCDProvider
type ArgoCDProviderSpec struct {
	// Namespace where ArgoCD will be deployed
	// +optional
	Namespace string `json:"namespace,omitempty"`

	// Version of ArgoCD to install
	// +optional
	Version string `json:"version,omitempty"`

	// InstallMethod specifies how ArgoCD should be installed
	// +optional
	InstallMethod InstallMethod `json:"installMethod,omitempty"`

	// Config contains ArgoCD-specific configuration
	// +optional
	Config ArgoCDConfig `json:"config,omitempty"`

	// AdminUser specifies admin user configuration
	// +optional
	AdminUser AdminUser `json:"adminUser,omitempty"`

	// Projects lists ArgoCD projects to create
	// +optional
	Projects []ArgoCDProject `json:"projects,omitempty"`
}

// ArgoCDConfig contains ArgoCD-specific configuration
type ArgoCDConfig struct {
	// Server configuration
	// +optional
	Server ServerConfig `json:"server,omitempty"`

	// Controller configuration
	// +optional
	Controller ApplicationControllerConfig `json:"controller,omitempty"`

	// RepoServer configuration
	// +optional
	RepoServer RepoServerConfig `json:"repoServer,omitempty"`

	// Notifications configuration
	// +optional
	Notifications NotificationsConfig `json:"notifications,omitempty"`

	// Dex configuration
	// +optional
	Dex DexConfig `json:"dex,omitempty"`

	// Ingress configuration
	// +optional
	Ingress IngressSpec `json:"ingress,omitempty"`
}

// ServerConfig defines ArgoCD server configuration
type ServerConfig struct {
	// Replicas for the server
	// +optional
	Replicas int32 `json:"replicas,omitempty"`

	// Resources for the server
	// +optional
	Resources *ResourceRequirements `json:"resources,omitempty"`
}

// ApplicationControllerConfig defines ArgoCD application controller configuration
type ApplicationControllerConfig struct {
	// Replicas for the controller
	// +optional
	Replicas int32 `json:"replicas,omitempty"`

	// Resources for the controller
	// +optional
	Resources *ResourceRequirements `json:"resources,omitempty"`
}

// RepoServerConfig defines ArgoCD repo server configuration
type RepoServerConfig struct {
	// Replicas for the repo server
	// +optional
	Replicas int32 `json:"replicas,omitempty"`

	// Resources for the repo server
	// +optional
	Resources *ResourceRequirements `json:"resources,omitempty"`
}

// NotificationsConfig defines ArgoCD notifications configuration
type NotificationsConfig struct {
	// Enabled indicates if notifications should be enabled
	// +optional
	Enabled bool `json:"enabled,omitempty"`
}

// DexConfig defines ArgoCD Dex configuration
type DexConfig struct {
	// Enabled indicates if Dex should be enabled
	// +optional
	Enabled bool `json:"enabled,omitempty"`
}

// ArgoCDProject defines an ArgoCD project
type ArgoCDProject struct {
	// Name of the project
	Name string `json:"name"`

	// Description of the project
	// +optional
	Description string `json:"description,omitempty"`

	// SourceRepos lists allowed source repositories
	// +optional
	SourceRepos []string `json:"sourceRepos,omitempty"`

	// Destinations lists allowed destinations
	// +optional
	Destinations []ProjectDestination `json:"destinations,omitempty"`
}

// ProjectDestination defines an ArgoCD project destination
type ProjectDestination struct {
	// Server URL
	// +optional
	Server string `json:"server,omitempty"`

	// Namespace
	// +optional
	Namespace string `json:"namespace,omitempty"`
}

// ArgoCDProviderStatus defines the observed state of ArgoCDProvider
type ArgoCDProviderStatus struct {
	// Conditions represent the latest available observations of the provider's state
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// Common duck-typed status fields for GitOps providers
	// Endpoint is the external URL for the ArgoCD UI
	// +optional
	Endpoint string `json:"endpoint,omitempty"`

	// InternalEndpoint is the cluster-internal URL for API access
	// +optional
	InternalEndpoint string `json:"internalEndpoint,omitempty"`

	// CredentialsSecretRef references the secret containing access credentials
	// +optional
	CredentialsSecretRef *SecretReference `json:"credentialsSecretRef,omitempty"`

	// ArgoCD-specific status fields
	// Installed indicates if ArgoCD is installed
	// +optional
	Installed bool `json:"installed,omitempty"`

	// Version is the installed version
	// +optional
	Version string `json:"version,omitempty"`

	// Phase represents the current phase (Pending, Installing, Ready, Error)
	// +optional
	Phase string `json:"phase,omitempty"`

	// AdminUser contains admin user information
	// +optional
	AdminUser *AdminUserStatus `json:"adminUser,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced

// ArgoCDProvider is the Schema for the argocdproviders API
type ArgoCDProvider struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ArgoCDProviderSpec   `json:"spec,omitempty"`
	Status ArgoCDProviderStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ArgoCDProviderList contains a list of ArgoCDProvider
type ArgoCDProviderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ArgoCDProvider `json:"items"`
}
