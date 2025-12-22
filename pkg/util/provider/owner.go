package provider

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetPlatformOwnerReference extracts the Platform owner reference from a provider CR
// Returns nil if no Platform owner reference is found
func GetPlatformOwnerReference(obj client.Object) *metav1.OwnerReference {
	for i := range obj.GetOwnerReferences() {
		ref := &obj.GetOwnerReferences()[i]
		if ref.APIVersion == "idpbuilder.cnoe.io/v1alpha2" && ref.Kind == "Platform" {
			return ref
		}
	}
	return nil
}
