package provider

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

type mockObject struct {
	metav1.ObjectMeta
	metav1.TypeMeta
}

func (m *mockObject) GetObjectKind() schema.ObjectKind {
	return m
}

func (m *mockObject) DeepCopyObject() runtime.Object {
	return m
}

func TestGetPlatformOwnerReference(t *testing.T) {
	tests := []struct {
		name      string
		ownerRefs []metav1.OwnerReference
		want      *metav1.OwnerReference
	}{
		{
			name:      "no owner references",
			ownerRefs: nil,
			want:      nil,
		},
		{
			name: "platform owner reference present",
			ownerRefs: []metav1.OwnerReference{
				{
					APIVersion: "idpbuilder.cnoe.io/v1alpha2",
					Kind:       "Platform",
					Name:       "test-platform",
					UID:        types.UID("12345"),
				},
			},
			want: &metav1.OwnerReference{
				APIVersion: "idpbuilder.cnoe.io/v1alpha2",
				Kind:       "Platform",
				Name:       "test-platform",
				UID:        types.UID("12345"),
			},
		},
		{
			name: "multiple owner references including platform",
			ownerRefs: []metav1.OwnerReference{
				{
					APIVersion: "apps/v1",
					Kind:       "Deployment",
					Name:       "some-deployment",
					UID:        types.UID("11111"),
				},
				{
					APIVersion: "idpbuilder.cnoe.io/v1alpha2",
					Kind:       "Platform",
					Name:       "test-platform",
					UID:        types.UID("12345"),
				},
			},
			want: &metav1.OwnerReference{
				APIVersion: "idpbuilder.cnoe.io/v1alpha2",
				Kind:       "Platform",
				Name:       "test-platform",
				UID:        types.UID("12345"),
			},
		},
		{
			name: "owner references but no platform",
			ownerRefs: []metav1.OwnerReference{
				{
					APIVersion: "apps/v1",
					Kind:       "Deployment",
					Name:       "some-deployment",
					UID:        types.UID("11111"),
				},
			},
			want: nil,
		},
		{
			name: "wrong API version",
			ownerRefs: []metav1.OwnerReference{
				{
					APIVersion: "idpbuilder.cnoe.io/v1alpha1",
					Kind:       "Platform",
					Name:       "test-platform",
					UID:        types.UID("12345"),
				},
			},
			want: nil,
		},
		{
			name: "wrong kind",
			ownerRefs: []metav1.OwnerReference{
				{
					APIVersion: "idpbuilder.cnoe.io/v1alpha2",
					Kind:       "NotPlatform",
					Name:       "test-platform",
					UID:        types.UID("12345"),
				},
			},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &mockObject{
				ObjectMeta: metav1.ObjectMeta{
					OwnerReferences: tt.ownerRefs,
				},
			}

			got := GetPlatformOwnerReference(obj)

			if tt.want == nil {
				if got != nil {
					t.Errorf("GetPlatformOwnerReference() = %v, want nil", got)
				}
				return
			}

			if got == nil {
				t.Errorf("GetPlatformOwnerReference() = nil, want %v", tt.want)
				return
			}

			if got.APIVersion != tt.want.APIVersion {
				t.Errorf("APIVersion = %v, want %v", got.APIVersion, tt.want.APIVersion)
			}
			if got.Kind != tt.want.Kind {
				t.Errorf("Kind = %v, want %v", got.Kind, tt.want.Kind)
			}
			if got.Name != tt.want.Name {
				t.Errorf("Name = %v, want %v", got.Name, tt.want.Name)
			}
			if got.UID != tt.want.UID {
				t.Errorf("UID = %v, want %v", got.UID, tt.want.UID)
			}
		})
	}
}
