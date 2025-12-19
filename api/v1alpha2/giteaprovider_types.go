package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GiteaProviderSpec defines the desired state of GiteaProvider
type GiteaProviderSpec struct {
	// Namespace where Gitea will be deployed
	// +optional
	Namespace string `json:"namespace,omitempty"`

	// Version of Gitea to install
	// +optional
	Version string `json:"version,omitempty"`

	// InstallMethod specifies how Gitea should be installed
	// +optional
	InstallMethod InstallMethod `json:"installMethod,omitempty"`

	// Config contains Gitea-specific configuration
	// +optional
	Config GiteaConfig `json:"config,omitempty"`

	// AdminUser specifies admin user configuration
	// +optional
	AdminUser AdminUser `json:"adminUser,omitempty"`

	// Organizations lists organizations to create
	// +optional
	Organizations []Organization `json:"organizations,omitempty"`
}

// InstallMethod specifies how a component should be installed
type InstallMethod struct {
	// Type of installation (Helm, Manifests)
	Type string `json:"type"`

	// Helm specifies Helm chart details
	// +optional
	Helm *HelmInstall `json:"helm,omitempty"`

	// Manifests specifies raw manifest details
	// +optional
	Manifests *ManifestsInstall `json:"manifests,omitempty"`
}

// HelmInstall specifies Helm chart installation details
type HelmInstall struct {
	// Repository URL for the Helm chart
	Repository string `json:"repository"`

	// Chart name
	Chart string `json:"chart"`

	// Version of the chart
	// +optional
	Version string `json:"version,omitempty"`

	// Values are Helm values to override defaults (YAML string)
	// +optional
	Values string `json:"values,omitempty"`
}

// ManifestsInstall specifies raw manifest installation details
type ManifestsInstall struct {
	// Path to manifests (embedded or external)
	Path string `json:"path"`
}

// GiteaConfig contains Gitea-specific configuration
type GiteaConfig struct {
	// Ingress configuration
	// +optional
	Ingress IngressSpec `json:"ingress,omitempty"`

	// Persistence configuration
	// +optional
	Persistence PersistenceSpec `json:"persistence,omitempty"`

	// Database configuration
	// +optional
	Database DatabaseSpec `json:"database,omitempty"`
}

// IngressSpec defines ingress configuration
type IngressSpec struct {
	// Enabled indicates if ingress should be created
	// +optional
	Enabled bool `json:"enabled,omitempty"`

	// ClassName specifies the ingress class to use
	// +optional
	ClassName string `json:"className,omitempty"`

	// Host specifies the hostname for the ingress
	// +optional
	Host string `json:"host,omitempty"`
}

// PersistenceSpec defines persistence configuration
type PersistenceSpec struct {
	// Enabled indicates if persistence should be enabled
	// +optional
	Enabled bool `json:"enabled,omitempty"`

	// Size of the persistent volume
	// +optional
	Size string `json:"size,omitempty"`
}

// DatabaseSpec defines database configuration
type DatabaseSpec struct {
	// Type of database (sqlite, postgres, mysql)
	// +optional
	Type string `json:"type,omitempty"`
}

// AdminUser defines admin user configuration
type AdminUser struct {
	// Username for the admin user
	// +optional
	Username string `json:"username,omitempty"`

	// Email for the admin user
	// +optional
	Email string `json:"email,omitempty"`

	// PasswordSecretRef references a secret containing the password
	// +optional
	PasswordSecretRef *SecretReference `json:"passwordSecretRef,omitempty"`

	// AutoGenerate indicates if credentials should be auto-generated
	// +optional
	AutoGenerate bool `json:"autoGenerate,omitempty"`
}

// Organization defines a Gitea organization
type Organization struct {
	// Name of the organization
	Name string `json:"name"`

	// Description of the organization
	// +optional
	Description string `json:"description,omitempty"`
}

// GiteaProviderStatus defines the observed state of GiteaProvider
type GiteaProviderStatus struct {
	// Conditions represent the latest available observations of the provider's state
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// Common duck-typed status fields for Git providers
	// Endpoint is the external URL for web UI and cloning
	// +optional
	Endpoint string `json:"endpoint,omitempty"`

	// InternalEndpoint is the cluster-internal URL for API access
	// +optional
	InternalEndpoint string `json:"internalEndpoint,omitempty"`

	// CredentialsSecretRef references the secret containing access credentials
	// +optional
	CredentialsSecretRef *SecretReference `json:"credentialsSecretRef,omitempty"`

	// Gitea-specific status fields
	// Installed indicates if Gitea is installed
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

// AdminUserStatus contains admin user status information
type AdminUserStatus struct {
	// Username of the admin user
	// +optional
	Username string `json:"username,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced

// GiteaProvider is the Schema for the giteaproviders API
type GiteaProvider struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GiteaProviderSpec   `json:"spec,omitempty"`
	Status GiteaProviderStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// GiteaProviderList contains a list of GiteaProvider
type GiteaProviderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GiteaProvider `json:"items"`
}
