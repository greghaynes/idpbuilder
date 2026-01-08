package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	CNOEURIScheme = "cnoe://"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type CustomPackage struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CustomPackageSpec   `json:"spec,omitempty"`
	Status CustomPackageStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type CustomPackageList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CustomPackage `json:"items"`
}

// CustomPackageSpec controls the installation of the custom applications.
type CustomPackageSpec struct {
	ArgoCD ArgoCDPackageSpec `json:"argoCD,omitempty"`

	// PlatformRef references the Platform resource to use for git provider configuration.
	// If not specified, the controller will look for a Platform named "platform" in the same namespace.
	// If no Platform is found, it falls back to the GitServerURL and InternalGitServeURL fields.
	// +optional
	PlatformRef *PlatformReference `json:"platformRef,omitempty"`

	// GitServerURL specifies the base URL for the git server for API calls.
	// for example, https://gitea.cnoe.localtest.me:8443
	// Deprecated: Use PlatformRef instead. This field is kept for backward compatibility.
	// +optional
	GitServerURL           string          `json:"gitServerURL,omitempty"`
	GitServerAuthSecretRef SecretReference `json:"gitServerAuthSecretRef,omitempty"`
	// InternalGitServeURL specifies the base URL for the git server accessible within the cluster.
	// for example, http://my-gitea-http.gitea.svc.cluster.local:3000
	// Deprecated: Use PlatformRef instead. This field is kept for backward compatibility.
	// +optional
	InternalGitServeURL string               `json:"internalGitServeURL,omitempty"`
	RemoteRepository    RemoteRepositorySpec `json:"remoteRepository"`
	// Replicate specifies whether to replicate remote or local contents to the local gitea server.
	// +kubebuilder:default:=false
	Replicate bool `json:"replicate"`
}

// PlatformReference references a Platform resource
type PlatformReference struct {
	// Name is the name of the Platform resource
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Namespace is the namespace of the Platform resource.
	// If not specified, defaults to the same namespace as the CustomPackage.
	// +optional
	Namespace string `json:"namespace,omitempty"`
}

// RemoteRepositorySpec specifies information about remote repositories.
type RemoteRepositorySpec struct {
	CloneSubmodules bool   `json:"cloneSubmodules"`
	Path            string `json:"path"`
	// Url specifies the url to the repository containing the ArgoCD application file
	Url string `json:"url"`
	// Ref specifies the specific ref supported by git fetch
	Ref string `json:"ref"`
}

type ArgoCDPackageSpec struct {
	// ApplicationFile specifies the absolute path to the ArgoCD application file
	ApplicationFile string `json:"applicationFile"`
	Name            string `json:"name"`
	Namespace       string `json:"namespace"`
	// +kubebuilder:validation:Enum:=Application;ApplicationSet
	Type string `json:"type"`
}

type CustomPackageStatus struct {
	// A Custom package is considered synced when the in-cluster repository url is set as the repository URL
	// This only applies for a package that references local directories
	Synced            bool        `json:"synced,omitempty"`
	GitRepositoryRefs []ObjectRef `json:"gitRepositoryRefs,omitempty"`
}

type ObjectRef struct {
	APIVersion string `json:"apiVersion,omitempty"`
	Name       string `json:"name,omitempty"`
	Namespace  string `json:"namespace,omitempty"`
	Kind       string `json:"kind,omitempty"`
	UID        string `json:"uid,omitempty"`
}
