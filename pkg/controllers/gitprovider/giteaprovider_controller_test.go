package gitprovider

import (
	"context"
	"testing"

	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"github.com/cnoe-io/idpbuilder/api/v1alpha2"
	"github.com/cnoe-io/idpbuilder/pkg/k8s"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestGiteaProviderReconciler_Reconcile(t *testing.T) {
	t.Skip("Requires envtest environment for Gitea installation")
	scheme := k8s.GetScheme()

	tests := []struct {
		name            string
		provider        *v1alpha2.GiteaProvider
		expectedPhase   string
		expectFinalizer bool
		expectError     bool
		skipReconcile   bool
	}{
		{
			name: "new giteaprovider gets finalizer",
			provider: &v1alpha2.GiteaProvider{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-gitea",
					Namespace: "gitea",
				},
				Spec: v1alpha2.GiteaProviderSpec{
					Namespace: "gitea",
					Version:   "1.24.3",
					AdminUser: v1alpha2.GiteaAdminUser{
						Username:     "giteaAdmin",
						Email:        "admin@test.com",
						AutoGenerate: true,
					},
				},
			},
			expectedPhase:   "", // We'll check after the first reconcile that adds finalizer
			expectFinalizer: true,
			expectError:     false,
			skipReconcile:   false, // Only do one reconcile to add finalizer
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fake client with objects
			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithRuntimeObjects(tt.provider).
				WithStatusSubresource(&v1alpha2.GiteaProvider{}).
				Build()

			reconciler := &GiteaProviderReconciler{
				Client: fakeClient,
				Scheme: scheme,
				Config: v1alpha1.BuildCustomizationSpec{
					Host:     "test.example.com",
					Port:     "8443",
					Protocol: "https",
				},
			}

			req := ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name:      tt.provider.Name,
					Namespace: tt.provider.Namespace,
				},
			}

			// First reconcile to add finalizer
			_, err := reconciler.Reconcile(context.Background(), req)
			if tt.expectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			// Get updated provider
			provider := &v1alpha2.GiteaProvider{}
			err = fakeClient.Get(context.Background(), req.NamespacedName, provider)
			require.NoError(t, err)

			// Verify finalizer
			if tt.expectFinalizer {
				assert.Contains(t, provider.Finalizers, giteaProviderFinalizer)
			}
		})
	}
}

func TestGiteaProviderReconciler_DeletionHandling(t *testing.T) {
	scheme := k8s.GetScheme()

	now := metav1.Now()
	provider := &v1alpha2.GiteaProvider{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-gitea",
			Namespace:         "gitea",
			Finalizers:        []string{giteaProviderFinalizer},
			DeletionTimestamp: &now,
		},
		Spec: v1alpha2.GiteaProviderSpec{
			Namespace: "gitea",
			Version:   "1.24.3",
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithRuntimeObjects(provider).
		WithStatusSubresource(&v1alpha2.GiteaProvider{}).
		Build()

	reconciler := &GiteaProviderReconciler{
		Client: fakeClient,
		Scheme: scheme,
		Config: v1alpha1.BuildCustomizationSpec{},
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      provider.Name,
			Namespace: provider.Namespace,
		},
	}

	// Reconcile to handle deletion
	_, err := reconciler.Reconcile(context.Background(), req)
	require.NoError(t, err)

	// Try to get updated provider - should not be found since it was deleted
	updatedProvider := &v1alpha2.GiteaProvider{}
	err = fakeClient.Get(context.Background(), req.NamespacedName, updatedProvider)
	// The object should be deleted (not found), or if still there, finalizer should be removed
	if err == nil {
		// If object is still there, finalizer should be removed
		assert.NotContains(t, updatedProvider.Finalizers, "giteaprovider.idpbuilder.cnoe.io/finalizer")
	}
}

func TestBuildConfigFromSpec(t *testing.T) {
	tests := []struct {
		name     string
		provider *v1alpha2.GiteaProvider
		expected v1alpha1.BuildCustomizationSpec
	}{
		{
			name: "all fields set in provider",
			provider: &v1alpha2.GiteaProvider{
				Spec: v1alpha2.GiteaProviderSpec{
					Protocol:       "https",
					Host:           "example.com",
					Port:           "8443",
					UsePathRouting: true,
				},
			},
			expected: v1alpha1.BuildCustomizationSpec{
				Protocol:       "https",
				Host:           "example.com",
				IngressHost:    "example.com",
				Port:           "8443",
				UsePathRouting: true,
			},
		},
		{
			name: "empty fields use defaults",
			provider: &v1alpha2.GiteaProvider{
				Spec: v1alpha2.GiteaProviderSpec{},
			},
			expected: v1alpha1.BuildCustomizationSpec{
				Protocol:       "http",
				Host:           "cnoe.localtest.me",
				IngressHost:    "cnoe.localtest.me",
				Port:           "8080",
				UsePathRouting: false,
			},
		},
		{
			name: "partial fields with defaults",
			provider: &v1alpha2.GiteaProvider{
				Spec: v1alpha2.GiteaProviderSpec{
					Host: "my-custom-host.com",
				},
			},
			expected: v1alpha1.BuildCustomizationSpec{
				Protocol:       "http",
				Host:           "my-custom-host.com",
				IngressHost:    "my-custom-host.com",
				Port:           "8080",
				UsePathRouting: false,
			},
		},
		{
			name: "IngressHost defaults to Host when empty",
			provider: &v1alpha2.GiteaProvider{
				Spec: v1alpha2.GiteaProviderSpec{
					Protocol: "https",
					Host:     "test.localtest.me",
					Port:     "9443",
				},
			},
			expected: v1alpha1.BuildCustomizationSpec{
				Protocol:       "https",
				Host:           "test.localtest.me",
				IngressHost:    "test.localtest.me",
				Port:           "9443",
				UsePathRouting: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reconciler := &GiteaProviderReconciler{}
			result := reconciler.buildConfigFromSpec(tt.provider)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsNginxAdmissionWebhookReady(t *testing.T) {
	scheme := k8s.GetScheme()

	t.Run("webhook service not found", func(t *testing.T) {
		fakeClient := fake.NewClientBuilder().
			WithScheme(scheme).
			Build()

		reconciler := &GiteaProviderReconciler{
			Client: fakeClient,
			Scheme: scheme,
		}

		ready, err := reconciler.isNginxAdmissionWebhookReady(context.Background())
		require.NoError(t, err)
		assert.False(t, ready)
	})

	t.Run("webhook service exists but no endpoints", func(t *testing.T) {
		service := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      nginxAdmissionWebhookServiceName,
				Namespace: nginxNamespace,
			},
		}

		fakeClient := fake.NewClientBuilder().
			WithScheme(scheme).
			WithRuntimeObjects(service).
			Build()

		reconciler := &GiteaProviderReconciler{
			Client: fakeClient,
			Scheme: scheme,
		}

		ready, err := reconciler.isNginxAdmissionWebhookReady(context.Background())
		require.NoError(t, err)
		assert.False(t, ready)
	})

	t.Run("webhook service and endpoints exist with ready addresses", func(t *testing.T) {
		service := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      nginxAdmissionWebhookServiceName,
				Namespace: nginxNamespace,
			},
		}

		endpoints := &corev1.Endpoints{
			ObjectMeta: metav1.ObjectMeta{
				Name:      nginxAdmissionWebhookServiceName,
				Namespace: nginxNamespace,
			},
			Subsets: []corev1.EndpointSubset{
				{
					Addresses: []corev1.EndpointAddress{
						{
							IP: "10.0.0.1",
						},
					},
				},
			},
		}

		fakeClient := fake.NewClientBuilder().
			WithScheme(scheme).
			WithRuntimeObjects(service, endpoints).
			Build()

		reconciler := &GiteaProviderReconciler{
			Client: fakeClient,
			Scheme: scheme,
		}

		ready, err := reconciler.isNginxAdmissionWebhookReady(context.Background())
		require.NoError(t, err)
		assert.True(t, ready)
	})

	t.Run("webhook service and endpoints exist but no ready addresses", func(t *testing.T) {
		service := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      nginxAdmissionWebhookServiceName,
				Namespace: nginxNamespace,
			},
		}

		endpoints := &corev1.Endpoints{
			ObjectMeta: metav1.ObjectMeta{
				Name:      nginxAdmissionWebhookServiceName,
				Namespace: nginxNamespace,
			},
			Subsets: []corev1.EndpointSubset{},
		}

		fakeClient := fake.NewClientBuilder().
			WithScheme(scheme).
			WithRuntimeObjects(service, endpoints).
			Build()

		reconciler := &GiteaProviderReconciler{
			Client: fakeClient,
			Scheme: scheme,
		}

		ready, err := reconciler.isNginxAdmissionWebhookReady(context.Background())
		require.NoError(t, err)
		assert.False(t, ready)
	})
}

func TestGiteaProviderReconciler_FinalizerAddition(t *testing.T) {
	scheme := k8s.GetScheme()

	provider := &v1alpha2.GiteaProvider{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-gitea",
			Namespace: "gitea",
		},
		Spec: v1alpha2.GiteaProviderSpec{
			Namespace: "gitea",
			Version:   "1.24.3",
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(provider).
		WithStatusSubresource(&v1alpha2.GiteaProvider{}).
		Build()

	reconciler := &GiteaProviderReconciler{
		Client: fakeClient,
		Scheme: scheme,
		Config: v1alpha1.BuildCustomizationSpec{},
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      provider.Name,
			Namespace: provider.Namespace,
		},
	}

	// First reconcile to add finalizer
	_, err := reconciler.Reconcile(context.Background(), req)
	require.NoError(t, err)

	// Get updated provider
	updatedProvider := &v1alpha2.GiteaProvider{}
	err = fakeClient.Get(context.Background(), req.NamespacedName, updatedProvider)
	require.NoError(t, err)

	// Verify finalizer was added
	assert.Contains(t, updatedProvider.Finalizers, giteaProviderFinalizer)
}

func TestGiteaProviderReconciler_PhaseTransitions(t *testing.T) {
	scheme := k8s.GetScheme()

	tests := []struct {
		name          string
		initialPhase  string
		expectedPhase string
	}{
		{
			name:          "empty phase transitions to Installing",
			initialPhase:  "",
			expectedPhase: "Installing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &v1alpha2.GiteaProvider{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "test-gitea",
					Namespace:  "gitea",
					Finalizers: []string{giteaProviderFinalizer},
				},
				Spec: v1alpha2.GiteaProviderSpec{
					Namespace: "gitea",
					Version:   "1.24.3",
				},
				Status: v1alpha2.GiteaProviderStatus{
					Phase: tt.initialPhase,
				},
			}

			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(provider).
				WithStatusSubresource(&v1alpha2.GiteaProvider{}).
				Build()

			reconciler := &GiteaProviderReconciler{
				Client: fakeClient,
				Scheme: scheme,
				Config: v1alpha1.BuildCustomizationSpec{},
			}

			req := ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name:      provider.Name,
					Namespace: provider.Namespace,
				},
			}

			// Reconcile
			_, err := reconciler.Reconcile(context.Background(), req)
			// May error due to missing resources, but phase should still update
			if err != nil {
				t.Logf("Reconcile error (expected): %v", err)
			}

			// Get updated provider
			updatedProvider := &v1alpha2.GiteaProvider{}
			err = fakeClient.Get(context.Background(), req.NamespacedName, updatedProvider)
			require.NoError(t, err)

			// Verify phase was updated
			assert.Equal(t, tt.expectedPhase, updatedProvider.Status.Phase)
		})
	}
}

func TestGiteaProviderReconciler_NotFound(t *testing.T) {
	scheme := k8s.GetScheme()

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		Build()

	reconciler := &GiteaProviderReconciler{
		Client: fakeClient,
		Scheme: scheme,
		Config: v1alpha1.BuildCustomizationSpec{},
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "nonexistent",
			Namespace: "gitea",
		},
	}

	// Reconcile should handle not found gracefully
	result, err := reconciler.Reconcile(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, ctrl.Result{}, result)
}

func TestGiteaProviderReconciler_CreateAdminSecret(t *testing.T) {
	scheme := k8s.GetScheme()

	tests := []struct {
		name             string
		provider         *v1alpha2.GiteaProvider
		expectedUsername string
	}{
		{
			name: "creates secret with custom username",
			provider: &v1alpha2.GiteaProvider{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-gitea",
					Namespace: "gitea",
				},
				Spec: v1alpha2.GiteaProviderSpec{
					Namespace: "gitea",
					AdminUser: v1alpha2.GiteaAdminUser{
						Username: "customadmin",
					},
				},
			},
			expectedUsername: "customadmin",
		},
		{
			name: "creates secret with default username",
			provider: &v1alpha2.GiteaProvider{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-gitea",
					Namespace: "gitea",
				},
				Spec: v1alpha2.GiteaProviderSpec{
					Namespace: "gitea",
					AdminUser: v1alpha2.GiteaAdminUser{
						Username: "",
					},
				},
			},
			expectedUsername: "giteaAdmin",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(tt.provider).
				Build()

			reconciler := &GiteaProviderReconciler{
				Client: fakeClient,
				Scheme: scheme,
			}

			secret, err := reconciler.createAdminSecretIfNotExists(context.Background(), tt.provider)
			require.NoError(t, err)
			require.NotNil(t, secret)

			// Retrieve the secret - fake client keeps StringData available
			retrievedSecret := &corev1.Secret{}
			err = fakeClient.Get(context.Background(), types.NamespacedName{
				Name:      secret.Name,
				Namespace: secret.Namespace,
			}, retrievedSecret)
			require.NoError(t, err)
			
			// Fake client doesn't convert StringData to Data, so check StringData
			assert.Equal(t, tt.expectedUsername, retrievedSecret.StringData["username"])
			assert.NotEmpty(t, retrievedSecret.StringData["password"])
		})
	}
}

func TestGiteaProviderReconciler_AdminSecretIdempotency(t *testing.T) {
	scheme := k8s.GetScheme()

	provider := &v1alpha2.GiteaProvider{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-gitea",
			Namespace: "gitea",
		},
		Spec: v1alpha2.GiteaProviderSpec{
			Namespace: "gitea",
			AdminUser: v1alpha2.GiteaAdminUser{
				Username: "testadmin",
			},
		},
	}

	// Pre-create the secret
	existingSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "gitea-credential",
			Namespace: "gitea",
		},
		Data: map[string][]byte{
			"username": []byte("testadmin"),
			"password": []byte("existingpassword"),
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(provider, existingSecret).
		Build()

	reconciler := &GiteaProviderReconciler{
		Client: fakeClient,
		Scheme: scheme,
	}

	// Call createAdminSecretIfNotExists - should return existing secret
	secret, err := reconciler.createAdminSecretIfNotExists(context.Background(), provider)
	require.NoError(t, err)
	require.NotNil(t, secret)

	// Verify it's the existing secret (password should match)
	assert.Equal(t, "existingpassword", string(secret.Data["password"]))
	assert.Equal(t, "testadmin", string(secret.Data["username"]))
}

func TestGiteaProviderReconciler_StatusConditions(t *testing.T) {
	scheme := k8s.GetScheme()

	provider := &v1alpha2.GiteaProvider{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-gitea",
			Namespace:  "gitea",
			Finalizers: []string{giteaProviderFinalizer},
		},
		Spec: v1alpha2.GiteaProviderSpec{
			Namespace: "gitea",
			Version:   "1.24.3",
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(provider).
		WithStatusSubresource(&v1alpha2.GiteaProvider{}).
		Build()

	reconciler := &GiteaProviderReconciler{
		Client: fakeClient,
		Scheme: scheme,
		Config: v1alpha1.BuildCustomizationSpec{},
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      provider.Name,
			Namespace: provider.Namespace,
		},
	}

	// Reconcile - will fail because resources don't exist
	_, err := reconciler.Reconcile(context.Background(), req)
	// Error is expected due to missing resources
	if err != nil {
		t.Logf("Expected error during reconcile: %v", err)
	}

	// Get updated provider
	updatedProvider := &v1alpha2.GiteaProvider{}
	err = fakeClient.Get(context.Background(), req.NamespacedName, updatedProvider)
	require.NoError(t, err)

	// Verify status conditions are set
	assert.NotEmpty(t, updatedProvider.Status.Conditions)

	// Find Ready condition
	var readyCondition *metav1.Condition
	for i := range updatedProvider.Status.Conditions {
		if updatedProvider.Status.Conditions[i].Type == "Ready" {
			readyCondition = &updatedProvider.Status.Conditions[i]
			break
		}
	}

	require.NotNil(t, readyCondition)
	// Should be False because installation failed
	assert.Equal(t, metav1.ConditionFalse, readyCondition.Status)
}

func TestGiteaProviderReconciler_Idempotency(t *testing.T) {
	scheme := k8s.GetScheme()

	provider := &v1alpha2.GiteaProvider{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-gitea",
			Namespace: "gitea",
		},
		Spec: v1alpha2.GiteaProviderSpec{
			Namespace: "gitea",
			Version:   "1.24.3",
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(provider).
		WithStatusSubresource(&v1alpha2.GiteaProvider{}).
		Build()

	reconciler := &GiteaProviderReconciler{
		Client: fakeClient,
		Scheme: scheme,
		Config: v1alpha1.BuildCustomizationSpec{},
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      provider.Name,
			Namespace: provider.Namespace,
		},
	}

	// First reconcile
	_, err1 := reconciler.Reconcile(context.Background(), req)
	if err1 != nil {
		t.Logf("First reconcile error (expected): %v", err1)
	}

	// Get provider after first reconcile
	provider1 := &v1alpha2.GiteaProvider{}
	err := fakeClient.Get(context.Background(), req.NamespacedName, provider1)
	require.NoError(t, err)

	// Second reconcile
	_, err2 := reconciler.Reconcile(context.Background(), req)
	if err2 != nil {
		t.Logf("Second reconcile error (expected): %v", err2)
	}

	// Get provider after second reconcile
	provider2 := &v1alpha2.GiteaProvider{}
	err = fakeClient.Get(context.Background(), req.NamespacedName, provider2)
	require.NoError(t, err)

	// Finalizer should still be present
	assert.Contains(t, provider2.Finalizers, giteaProviderFinalizer)

	// Resource version should change if updates occurred
	// but reconciliation should be idempotent (no duplicate resources)
}

func TestGiteaProviderReconciler_EnsureAdminSecretWithoutToken(t *testing.T) {
	scheme := k8s.GetScheme()

	provider := &v1alpha2.GiteaProvider{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-gitea",
			Namespace: "gitea",
		},
		Spec: v1alpha2.GiteaProviderSpec{
			Namespace: "gitea",
			AdminUser: v1alpha2.GiteaAdminUser{
				Username: "testuser",
			},
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(provider).
		Build()

	reconciler := &GiteaProviderReconciler{
		Client: fakeClient,
		Scheme: scheme,
	}

	// Ensure admin secret without token
	err := reconciler.ensureAdminSecretWithoutToken(context.Background(), provider)
	require.NoError(t, err)

	// Verify secret was created
	secret := &corev1.Secret{}
	err = fakeClient.Get(context.Background(), types.NamespacedName{
		Name:      "gitea-credential",
		Namespace: "gitea",
	}, secret)
	require.NoError(t, err)

	// Fake client doesn't convert StringData to Data - check StringData
	assert.Equal(t, "testuser", secret.StringData["username"])
	assert.NotEmpty(t, secret.StringData["password"])
	assert.Empty(t, secret.StringData["token"])
}

func TestGiteaProviderReconciler_IsGiteaReadyDeploymentNotFound(t *testing.T) {
	scheme := k8s.GetScheme()

	provider := &v1alpha2.GiteaProvider{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-gitea",
			Namespace: "gitea",
		},
		Spec: v1alpha2.GiteaProviderSpec{
			Namespace: "gitea",
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		Build()

	reconciler := &GiteaProviderReconciler{
		Client: fakeClient,
		Scheme: scheme,
		Config: v1alpha1.BuildCustomizationSpec{},
	}

	ready, err := reconciler.isGiteaReady(context.Background(), provider)
	require.NoError(t, err)
	assert.False(t, ready)
}

func TestGiteaProviderReconciler_IsGiteaReadyDeploymentNotReady(t *testing.T) {
	scheme := k8s.GetScheme()

	provider := &v1alpha2.GiteaProvider{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-gitea",
			Namespace: "gitea",
		},
		Spec: v1alpha2.GiteaProviderSpec{
			Namespace: "gitea",
		},
	}

	// Create deployment with 0 available replicas
	deployment := map[string]interface{}{
		"apiVersion": "apps/v1",
		"kind":       "Deployment",
		"metadata": map[string]interface{}{
			"name":      "my-gitea",
			"namespace": "gitea",
		},
		"status": map[string]interface{}{
			"availableReplicas": int64(0),
		},
	}

	deploymentObj := &unstructured.Unstructured{Object: deployment}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(deploymentObj).
		Build()

	reconciler := &GiteaProviderReconciler{
		Client: fakeClient,
		Scheme: scheme,
		Config: v1alpha1.BuildCustomizationSpec{},
	}

	ready, err := reconciler.isGiteaReady(context.Background(), provider)
	require.NoError(t, err)
	assert.False(t, ready)
}

func TestGiteaProviderReconciler_HandleDeletionRemovesFinalizer(t *testing.T) {
	scheme := k8s.GetScheme()

	now := metav1.Now()
	provider := &v1alpha2.GiteaProvider{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-gitea",
			Namespace:         "gitea",
			Finalizers:        []string{giteaProviderFinalizer},
			DeletionTimestamp: &now,
		},
		Spec: v1alpha2.GiteaProviderSpec{
			Namespace: "gitea",
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(provider).
		Build()

	reconciler := &GiteaProviderReconciler{
		Client: fakeClient,
		Scheme: scheme,
	}

	result, err := reconciler.handleDeletion(context.Background(), provider)
	require.NoError(t, err)
	assert.Equal(t, ctrl.Result{}, result)

	// Try to get the updated provider
	updatedProvider := &v1alpha2.GiteaProvider{}
	err = fakeClient.Get(context.Background(), types.NamespacedName{
		Name:      provider.Name,
		Namespace: provider.Namespace,
	}, updatedProvider)
	
	// Finalizer should be removed
	if err == nil {
		assert.NotContains(t, updatedProvider.Finalizers, giteaProviderFinalizer)
	}
}

func TestGiteaProviderReconciler_ConfigurationVariations(t *testing.T) {
	scheme := k8s.GetScheme()

	tests := []struct {
		name     string
		spec     v1alpha2.GiteaProviderSpec
		wantHost string
		wantPort string
	}{
		{
			name: "custom host and port",
			spec: v1alpha2.GiteaProviderSpec{
				Namespace: "gitea",
				Host:      "custom.example.com",
				Port:      "9443",
			},
			wantHost: "custom.example.com",
			wantPort: "9443",
		},
		{
			name: "default values",
			spec: v1alpha2.GiteaProviderSpec{
				Namespace: "gitea",
			},
			wantHost: "cnoe.localtest.me",
			wantPort: "8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &v1alpha2.GiteaProvider{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-gitea",
					Namespace: "gitea",
				},
				Spec: tt.spec,
			}

			reconciler := &GiteaProviderReconciler{
				Client: nil,
				Scheme: scheme,
			}

			config := reconciler.buildConfigFromSpec(provider)
			assert.Equal(t, tt.wantHost, config.Host)
			assert.Equal(t, tt.wantPort, config.Port)
		})
	}
}

func TestGiteaProviderReconciler_PasswordGeneration(t *testing.T) {
	scheme := k8s.GetScheme()

	provider := &v1alpha2.GiteaProvider{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-gitea",
			Namespace: "gitea",
		},
		Spec: v1alpha2.GiteaProviderSpec{
			Namespace: "gitea",
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(provider).
		Build()

	reconciler := &GiteaProviderReconciler{
		Client: fakeClient,
		Scheme: scheme,
	}

	_, err := reconciler.createAdminSecretIfNotExists(context.Background(), provider)
	require.NoError(t, err)

	// Retrieve the secret - fake client keeps StringData
	retrievedSecret := &corev1.Secret{}
	err = fakeClient.Get(context.Background(), types.NamespacedName{
		Name:      "gitea-credential",
		Namespace: "gitea",
	}, retrievedSecret)
	require.NoError(t, err)

	password := retrievedSecret.StringData["password"]
	
	// Verify password is not empty
	assert.NotEmpty(t, password)
	
	// Verify password has reasonable length (should be 40+ chars based on GeneratePassword)
	assert.Greater(t, len(password), 30)
}

func TestGiteaProviderReconciler_MultipleReconciliationsNoConflict(t *testing.T) {
	scheme := k8s.GetScheme()

	provider := &v1alpha2.GiteaProvider{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-gitea",
			Namespace: "gitea",
		},
		Spec: v1alpha2.GiteaProviderSpec{
			Namespace: "gitea",
			Version:   "1.24.3",
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(provider).
		WithStatusSubresource(&v1alpha2.GiteaProvider{}).
		Build()

	reconciler := &GiteaProviderReconciler{
		Client: fakeClient,
		Scheme: scheme,
		Config: v1alpha1.BuildCustomizationSpec{},
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      provider.Name,
			Namespace: provider.Namespace,
		},
	}

	// Perform multiple reconciliations
	for i := 0; i < 3; i++ {
		_, err := reconciler.Reconcile(context.Background(), req)
		// May error due to missing resources, but should not conflict
		if err != nil {
			t.Logf("Reconcile iteration %d error (may be expected): %v", i, err)
		}

		// Get provider to verify it's still accessible
		updatedProvider := &v1alpha2.GiteaProvider{}
		err = fakeClient.Get(context.Background(), req.NamespacedName, updatedProvider)
		require.NoError(t, err)
	}
}

func TestGiteaProviderReconciler_StatusFieldsUpdate(t *testing.T) {
	scheme := k8s.GetScheme()

	provider := &v1alpha2.GiteaProvider{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-gitea",
			Namespace:  "gitea",
			Finalizers: []string{giteaProviderFinalizer},
		},
		Spec: v1alpha2.GiteaProviderSpec{
			Namespace: "gitea",
			Version:   "1.24.3",
			Protocol:  "https",
			Host:      "test.example.com",
			Port:      "8443",
			AdminUser: v1alpha2.GiteaAdminUser{
				Username: "customadmin",
			},
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(provider).
		WithStatusSubresource(&v1alpha2.GiteaProvider{}).
		Build()

	reconciler := &GiteaProviderReconciler{
		Client: fakeClient,
		Scheme: scheme,
		Config: v1alpha1.BuildCustomizationSpec{
			Protocol: "https",
			Host:     "test.example.com",
			Port:     "8443",
		},
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      provider.Name,
			Namespace: provider.Namespace,
		},
	}

	// Reconcile
	_, err := reconciler.Reconcile(context.Background(), req)
	// May error, that's ok
	if err != nil {
		t.Logf("Expected reconcile error: %v", err)
	}

	// Get updated provider
	updatedProvider := &v1alpha2.GiteaProvider{}
	err = fakeClient.Get(context.Background(), req.NamespacedName, updatedProvider)
	require.NoError(t, err)

	// Phase should be set
	assert.NotEmpty(t, updatedProvider.Status.Phase)
}

func TestGiteaProviderReconciler_ReconcileGiteaNamespaceCreation(t *testing.T) {
	scheme := k8s.GetScheme()

	provider := &v1alpha2.GiteaProvider{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-gitea",
			Namespace: "gitea",
		},
		Spec: v1alpha2.GiteaProviderSpec{
			Namespace: "custom-namespace",
			Version:   "1.24.3",
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(provider).
		Build()

	reconciler := &GiteaProviderReconciler{
		Client: fakeClient,
		Scheme: scheme,
		Config: v1alpha1.BuildCustomizationSpec{},
	}

	// Call reconcileGitea which should try to ensure namespace
	_, err := reconciler.reconcileGitea(context.Background(), provider)
	// Will error on installation, but namespace creation is attempted
	if err != nil {
		t.Logf("Expected error: %v", err)
	}

	// Verify namespace was created
	ns := &corev1.Namespace{}
	err = fakeClient.Get(context.Background(), types.NamespacedName{
		Name: "custom-namespace",
	}, ns)
	require.NoError(t, err)
	assert.Equal(t, "custom-namespace", ns.Name)
}

func TestGiteaProviderReconciler_NginxWebhookNotReadyRequeues(t *testing.T) {
	scheme := k8s.GetScheme()

	provider := &v1alpha2.GiteaProvider{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-gitea",
			Namespace: "gitea",
		},
		Spec: v1alpha2.GiteaProviderSpec{
			Namespace: "gitea",
			Version:   "1.24.3",
		},
	}

	// Create nginx service but no endpoints (not ready)
	nginxService := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nginxAdmissionWebhookServiceName,
			Namespace: nginxNamespace,
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(provider, nginxService).
		Build()

	reconciler := &GiteaProviderReconciler{
		Client: fakeClient,
		Scheme: scheme,
		Config: v1alpha1.BuildCustomizationSpec{},
	}

	// Call reconcileGitea - should requeue when nginx webhook not ready
	result, err := reconciler.reconcileGitea(context.Background(), provider)
	
	// Error or requeue is expected
	if err != nil {
		t.Logf("reconcileGitea error (expected): %v", err)
	}
	
	// Should requeue
	assert.Equal(t, defaultRequeueTime, result.RequeueAfter)
}

func TestGiteaProviderReconciler_AdminUsernameDefaulting(t *testing.T) {
	scheme := k8s.GetScheme()

	tests := []struct {
		name             string
		specUsername     string
		expectedUsername string
	}{
		{
			name:             "empty username defaults to giteaAdmin",
			specUsername:     "",
			expectedUsername: "giteaAdmin",
		},
		{
			name:             "custom username is preserved",
			specUsername:     "myadmin",
			expectedUsername: "myadmin",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &v1alpha2.GiteaProvider{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-gitea",
					Namespace: "gitea",
				},
				Spec: v1alpha2.GiteaProviderSpec{
					Namespace: "gitea",
					AdminUser: v1alpha2.GiteaAdminUser{
						Username: tt.specUsername,
					},
				},
			}

			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(provider).
				Build()

			reconciler := &GiteaProviderReconciler{
				Client: fakeClient,
				Scheme: scheme,
			}

			_, err := reconciler.createAdminSecretIfNotExists(context.Background(), provider)
			require.NoError(t, err)
			
			// Retrieve the secret - fake client keeps StringData
			retrievedSecret := &corev1.Secret{}
			err = fakeClient.Get(context.Background(), types.NamespacedName{
				Name:      "gitea-credential",
				Namespace: "gitea",
			}, retrievedSecret)
			require.NoError(t, err)
			
			assert.Equal(t, tt.expectedUsername, retrievedSecret.StringData["username"])
		})
	}
}

func TestGiteaProviderReconciler_UsePathRouting(t *testing.T) {
	tests := []struct {
		name           string
		usePathRouting bool
	}{
		{
			name:           "path routing enabled",
			usePathRouting: true,
		},
		{
			name:           "path routing disabled",
			usePathRouting: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &v1alpha2.GiteaProvider{
				Spec: v1alpha2.GiteaProviderSpec{
					UsePathRouting: tt.usePathRouting,
					Host:           "example.com",
					Port:           "8080",
					Protocol:       "http",
				},
			}

			reconciler := &GiteaProviderReconciler{}
			config := reconciler.buildConfigFromSpec(provider)
			
			assert.Equal(t, tt.usePathRouting, config.UsePathRouting)
		})
	}
}
