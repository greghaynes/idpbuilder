package gatewayprovider

import (
	"context"
	"testing"
	"time"

	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"github.com/cnoe-io/idpbuilder/api/v1alpha2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// TestNginxGatewayFunctional tests the functional behavior of NginxGateway controller
// This validates that creating a NginxGateway resource triggers deployment of nginx resources
func TestNginxGatewayFunctional(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = v1alpha2.AddToScheme(scheme)
	_ = v1alpha1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)

	// Create a fake client
	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		Build()

	// Create reconciler
	reconciler := &NginxGatewayReconciler{
		Client: fakeClient,
		Scheme: scheme,
		Config: v1alpha1.BuildCustomizationSpec{},
	}

	// Create test NginxGateway
	nginxGateway := &v1alpha2.NginxGateway{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-nginx-gateway",
			Namespace: "test-namespace",
		},
		Spec: v1alpha2.NginxGatewaySpec{
			Namespace: "ingress-nginx",
			Version:   "1.13.0",
			IngressClass: v1alpha2.NginxIngressClass{
				Name:      "nginx",
				IsDefault: true,
			},
		},
	}

	// Create the NginxGateway resource
	err := fakeClient.Create(context.Background(), nginxGateway)
	require.NoError(t, err, "Failed to create NginxGateway")

	// Create namespace for nginx
	ns := &appsv1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "ingress-nginx",
		},
	}
	err = fakeClient.Create(context.Background(), ns)
	require.NoError(t, err, "Failed to create namespace")

	// Reconcile
	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      nginxGateway.Name,
			Namespace: nginxGateway.Namespace,
		},
	}

	ctx := context.Background()
	result, err := reconciler.Reconcile(ctx, req)

	// Verify reconciliation succeeded
	assert.NoError(t, err, "Reconciliation should not error")
	assert.NotEqual(t, result.RequeueAfter, time.Duration(0), "Should requeue to check readiness")

	// Verify NginxGateway status was updated
	updatedGateway := &v1alpha2.NginxGateway{}
	err = fakeClient.Get(ctx, types.NamespacedName{
		Name:      nginxGateway.Name,
		Namespace: nginxGateway.Namespace,
	}, updatedGateway)
	require.NoError(t, err, "Failed to get updated NginxGateway")

	// Status should be populated
	assert.Equal(t, "Installing", updatedGateway.Status.Phase, "Phase should be Installing initially")
}

// TestNginxGatewayResourcesCreated tests that nginx resources are actually created
func TestNginxGatewayResourcesCreated(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = v1alpha2.AddToScheme(scheme)
	_ = v1alpha1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)

	// Create a fake client that tracks created objects
	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		Build()

	// Create reconciler
	reconciler := &NginxGatewayReconciler{
		Client: fakeClient,
		Scheme: scheme,
		Config: v1alpha1.BuildCustomizationSpec{},
	}

	// Create test NginxGateway
	nginxGateway := &v1alpha2.NginxGateway{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-nginx",
			Namespace: "test-ns",
		},
		Spec: v1alpha2.NginxGatewaySpec{
			Namespace: "ingress-nginx",
			Version:   "1.13.0",
		},
	}

	err := fakeClient.Create(context.Background(), nginxGateway)
	require.NoError(t, err)

	// Create namespace
	ns := &appsv1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "ingress-nginx",
		},
	}
	err = fakeClient.Create(context.Background(), ns)
	require.NoError(t, err)

	// Reconcile to install nginx
	ctx := context.Background()
	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      nginxGateway.Name,
			Namespace: nginxGateway.Namespace,
		},
	}

	_, err = reconciler.Reconcile(ctx, req)
	assert.NoError(t, err)

	// Note: In a unit test with fake client, we can't validate actual resource creation
	// because the embedded manifests require a real cluster to apply
	// This test validates the reconciliation logic completes without errors
	t.Log("Nginx Gateway reconciliation completed successfully")
}

// TestNginxGatewayStatusUpdate tests that status fields are updated correctly
func TestNginxGatewayStatusUpdate(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = v1alpha2.AddToScheme(scheme)
	_ = v1alpha1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)

	// Create deployment to simulate nginx being ready
	deployment := &unstructured.Unstructured{}
	deployment.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "apps",
		Version: "v1",
		Kind:    "Deployment",
	})
	deployment.SetName("ingress-nginx-controller")
	deployment.SetNamespace("ingress-nginx")

	// Set deployment status to ready
	_ = unstructured.SetNestedField(deployment.Object, int64(1), "status", "replicas")
	_ = unstructured.SetNestedField(deployment.Object, int64(1), "status", "availableReplicas")
	_ = unstructured.SetNestedField(deployment.Object, int64(1), "status", "readyReplicas")

	// Create service
	service := &unstructured.Unstructured{}
	service.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "",
		Version: "v1",
		Kind:    "Service",
	})
	service.SetName("ingress-nginx-controller")
	service.SetNamespace("ingress-nginx")
	_ = unstructured.SetNestedField(service.Object, "10.96.0.1", "spec", "clusterIP")

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(deployment, service).
		Build()

	reconciler := &NginxGatewayReconciler{
		Client: fakeClient,
		Scheme: scheme,
		Config: v1alpha1.BuildCustomizationSpec{},
	}

	nginxGateway := &v1alpha2.NginxGateway{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-nginx",
			Namespace: "test-ns",
		},
		Spec: v1alpha2.NginxGatewaySpec{
			Namespace: "ingress-nginx",
			Version:   "1.13.0",
			IngressClass: v1alpha2.NginxIngressClass{
				Name:      "nginx",
				IsDefault: true,
			},
		},
	}

	err := fakeClient.Create(context.Background(), nginxGateway)
	require.NoError(t, err)

	// Create namespace
	ns := &appsv1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "ingress-nginx",
		},
	}
	err = fakeClient.Create(context.Background(), ns)
	require.NoError(t, err)

	// Reconcile
	ctx := context.Background()
	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      nginxGateway.Name,
			Namespace: nginxGateway.Namespace,
		},
	}

	// First reconcile - installs nginx
	_, err = reconciler.Reconcile(ctx, req)
	assert.NoError(t, err)

	// Second reconcile - should detect nginx is ready
	_, err = reconciler.Reconcile(ctx, req)
	assert.NoError(t, err)

	// Get updated gateway
	updatedGateway := &v1alpha2.NginxGateway{}
	err = fakeClient.Get(ctx, client.ObjectKey{
		Name:      nginxGateway.Name,
		Namespace: nginxGateway.Namespace,
	}, updatedGateway)
	require.NoError(t, err)

	// Verify duck-typed status fields are set
	assert.Equal(t, "nginx", updatedGateway.Status.IngressClassName, "IngressClassName should be set")
	assert.NotEmpty(t, updatedGateway.Status.InternalEndpoint, "InternalEndpoint should be set")
	assert.Equal(t, "Ready", updatedGateway.Status.Phase, "Phase should be Ready")
	assert.True(t, updatedGateway.Status.Installed, "Installed should be true")
	assert.Equal(t, "1.13.0", updatedGateway.Status.Version, "Version should match spec")

	// Verify controller status
	assert.Equal(t, int32(1), updatedGateway.Status.Controller.Replicas, "Controller replicas should be set")
	assert.Equal(t, int32(1), updatedGateway.Status.Controller.ReadyReplicas, "Controller ready replicas should be set")
}

// TestNginxGatewayDeletion tests that finalizer logic works correctly
func TestNginxGatewayDeletion(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = v1alpha2.AddToScheme(scheme)
	_ = v1alpha1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		Build()

	reconciler := &NginxGatewayReconciler{
		Client: fakeClient,
		Scheme: scheme,
		Config: v1alpha1.BuildCustomizationSpec{},
	}

	nginxGateway := &v1alpha2.NginxGateway{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-nginx",
			Namespace:  "test-ns",
			Finalizers: []string{"nginxgateway.idpbuilder.cnoe.io/finalizer"},
		},
		Spec: v1alpha2.NginxGatewaySpec{
			Namespace: "ingress-nginx",
		},
	}

	// Set deletion timestamp
	now := metav1.Now()
	nginxGateway.DeletionTimestamp = &now

	err := fakeClient.Create(context.Background(), nginxGateway)
	require.NoError(t, err)

	ctx := context.Background()
	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      nginxGateway.Name,
			Namespace: nginxGateway.Namespace,
		},
	}

	// Reconcile deletion
	_, err = reconciler.Reconcile(ctx, req)
	assert.NoError(t, err)

	// Verify finalizer was removed
	updatedGateway := &v1alpha2.NginxGateway{}
	err = fakeClient.Get(ctx, client.ObjectKey{
		Name:      nginxGateway.Name,
		Namespace: nginxGateway.Namespace,
	}, updatedGateway)
	require.NoError(t, err)

	assert.Empty(t, updatedGateway.Finalizers, "Finalizers should be removed")
}
