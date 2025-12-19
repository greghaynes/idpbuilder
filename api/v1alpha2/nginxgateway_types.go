package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NginxGatewaySpec defines the desired state of NginxGateway
type NginxGatewaySpec struct {
	// Namespace where Nginx Ingress will be deployed
	// +optional
	Namespace string `json:"namespace,omitempty"`

	// Version of Nginx Ingress to install
	// +optional
	Version string `json:"version,omitempty"`

	// InstallMethod specifies how Nginx should be installed
	// +optional
	InstallMethod InstallMethod `json:"installMethod,omitempty"`

	// Config contains Nginx-specific configuration
	// +optional
	Config NginxConfig `json:"config,omitempty"`

	// IngressClass specifies ingress class configuration
	// +optional
	IngressClass IngressClassConfig `json:"ingressClass,omitempty"`

	// DefaultTLS specifies default TLS configuration
	// +optional
	DefaultTLS *TLSConfig `json:"defaultTLS,omitempty"`
}

// NginxConfig contains Nginx-specific configuration
type NginxConfig struct {
	// Controller configuration
	// +optional
	Controller ControllerConfig `json:"controller,omitempty"`
}

// ControllerConfig defines Nginx controller configuration
type ControllerConfig struct {
	// Service configuration
	// +optional
	Service ServiceConfig `json:"service,omitempty"`

	// Resources for the controller
	// +optional
	Resources *ResourceRequirements `json:"resources,omitempty"`

	// AdmissionWebhooks configuration
	// +optional
	AdmissionWebhooks *AdmissionWebhooksConfig `json:"admissionWebhooks,omitempty"`

	// Config is a map of Nginx configuration options
	// +optional
	Config map[string]string `json:"config,omitempty"`
}

// ServiceConfig defines service configuration
type ServiceConfig struct {
	// Type of service (NodePort, LoadBalancer, ClusterIP)
	// +optional
	Type string `json:"type,omitempty"`

	// NodePorts configuration for NodePort service type
	// +optional
	NodePorts *NodePortsConfig `json:"nodePorts,omitempty"`
}

// NodePortsConfig defines NodePort configuration
type NodePortsConfig struct {
	// HTTP port
	// +optional
	HTTP int32 `json:"http,omitempty"`

	// HTTPS port
	// +optional
	HTTPS int32 `json:"https,omitempty"`
}

// ResourceRequirements defines resource requirements
type ResourceRequirements struct {
	// Limits defines resource limits
	// +optional
	Limits map[string]string `json:"limits,omitempty"`

	// Requests defines resource requests
	// +optional
	Requests map[string]string `json:"requests,omitempty"`
}

// AdmissionWebhooksConfig defines admission webhooks configuration
type AdmissionWebhooksConfig struct {
	// Enabled indicates if admission webhooks should be enabled
	// +optional
	Enabled bool `json:"enabled,omitempty"`
}

// IngressClassConfig defines ingress class configuration
type IngressClassConfig struct {
	// Name of the ingress class
	// +optional
	Name string `json:"name,omitempty"`

	// IsDefault indicates if this should be the default ingress class
	// +optional
	IsDefault bool `json:"isDefault,omitempty"`
}

// TLSConfig defines TLS configuration
type TLSConfig struct {
	// SecretRef references a secret containing TLS certificates
	// +optional
	SecretRef *SecretReference `json:"secretRef,omitempty"`
}

// NginxGatewayStatus defines the observed state of NginxGateway
type NginxGatewayStatus struct {
	// Conditions represent the latest available observations of the gateway's state
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// Common duck-typed status fields for Gateway providers
	// IngressClassName is the name of the ingress class to use in Ingress resources
	// +optional
	IngressClassName string `json:"ingressClassName,omitempty"`

	// LoadBalancerEndpoint is the external endpoint for accessing services
	// +optional
	LoadBalancerEndpoint string `json:"loadBalancerEndpoint,omitempty"`

	// InternalEndpoint is the cluster-internal API endpoint
	// +optional
	InternalEndpoint string `json:"internalEndpoint,omitempty"`

	// Nginx-specific status fields
	// Installed indicates if Nginx is installed
	// +optional
	Installed bool `json:"installed,omitempty"`

	// Version is the installed version
	// +optional
	Version string `json:"version,omitempty"`

	// Phase represents the current phase (Pending, Installing, Ready, Error)
	// +optional
	Phase string `json:"phase,omitempty"`

	// Controller status
	// +optional
	Controller *ControllerStatus `json:"controller,omitempty"`
}

// ControllerStatus contains controller status information
type ControllerStatus struct {
	// Replicas is the desired number of replicas
	// +optional
	Replicas int32 `json:"replicas,omitempty"`

	// ReadyReplicas is the number of ready replicas
	// +optional
	ReadyReplicas int32 `json:"readyReplicas,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced

// NginxGateway is the Schema for the nginxgateways API
type NginxGateway struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NginxGatewaySpec   `json:"spec,omitempty"`
	Status NginxGatewayStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// NginxGatewayList contains a list of NginxGateway
type NginxGatewayList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NginxGateway `json:"items"`
}
