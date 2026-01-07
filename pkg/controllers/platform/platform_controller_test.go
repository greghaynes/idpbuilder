package platform

import (
	"context"
	"testing"

	argov1alpha1 "github.com/cnoe-io/argocd-api/api/argo/application/v1alpha1"
	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"github.com/cnoe-io/idpbuilder/api/v1alpha2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestPlatformReconciler_Reconcile(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = v1alpha2.AddToScheme(scheme)

	tests := []struct {
		name           string
		platform       *v1alpha2.Platform
		providers      []client.Object
		expectRequeue  bool
		expectError    bool
		validateStatus func(*testing.T, *v1alpha2.Platform)
	}{
		{
			name: "platform with no providers",
			platform: &v1alpha2.Platform{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-platform",
					Namespace: "idpbuilder-system",
				},
				Spec: v1alpha2.PlatformSpec{
					Domain:     "test.local",
					Components: v1alpha2.PlatformComponents{},
				},
			},
			providers:     []client.Object{},
			expectRequeue: false,
			expectError:   false,
			validateStatus: func(t *testing.T, p *v1alpha2.Platform) {
				// Platform requires at least one git provider to be Ready
				assert.Equal(t, "Initializing", p.Status.Phase)
			},
		},
		{
			name: "platform with git provider reference",
			platform: &v1alpha2.Platform{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-platform",
					Namespace: "idpbuilder-system",
				},
				Spec: v1alpha2.PlatformSpec{
					Domain: "test.local",
					Components: v1alpha2.PlatformComponents{
						GitProviders: []v1alpha2.ProviderReference{
							{
								Name:      "gitea-test",
								Kind:      "GiteaProvider",
								Namespace: "idpbuilder-system",
							},
						},
					},
				},
			},
			providers: []client.Object{
				&v1alpha2.GiteaProvider{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "gitea-test",
						Namespace: "idpbuilder-system",
					},
					Status: v1alpha2.GiteaProviderStatus{
						Conditions: []metav1.Condition{
							{
								Type:    "Ready",
								Status:  metav1.ConditionTrue,
								Message: "Gitea is ready",
								Reason:  "GiteaReady",
							},
						},
					},
				},
			},
			expectRequeue: false,
			expectError:   false,
			validateStatus: func(t *testing.T, p *v1alpha2.Platform) {
				assert.Equal(t, "Ready", p.Status.Phase)
				assert.Len(t, p.Status.Providers.GitProviders, 1)
				assert.True(t, p.Status.Providers.GitProviders[0].Ready)
				assert.Equal(t, "Gitea is ready", p.Status.Providers.GitProviders[0].Message)
				assert.Equal(t, "GiteaReady", p.Status.Providers.GitProviders[0].Reason)
			},
		},
		{
			name: "platform with gateway provider reference",
			platform: &v1alpha2.Platform{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-platform",
					Namespace: "idpbuilder-system",
				},
				Spec: v1alpha2.PlatformSpec{
					Domain: "test.local",
					Components: v1alpha2.PlatformComponents{
						Gateways: []v1alpha2.ProviderReference{
							{
								Name:      "nginx-test",
								Kind:      "NginxGateway",
								Namespace: "idpbuilder-system",
							},
						},
					},
				},
			},
			providers: []client.Object{
				&v1alpha2.NginxGateway{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "nginx-test",
						Namespace: "idpbuilder-system",
					},
					Status: v1alpha2.NginxGatewayStatus{
						Conditions: []metav1.Condition{
							{
								Type:   "Ready",
								Status: metav1.ConditionTrue,
							},
						},
					},
				},
			},
			expectRequeue: false,
			expectError:   false,
			validateStatus: func(t *testing.T, p *v1alpha2.Platform) {
				// Platform requires at least one git provider to be Ready, so even though gateway is ready, platform is not
				assert.Equal(t, "Initializing", p.Status.Phase)
				assert.Len(t, p.Status.Providers.Gateways, 1)
				assert.True(t, p.Status.Providers.Gateways[0].Ready)
			},
		},
		{
			name: "platform with not ready providers",
			platform: &v1alpha2.Platform{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-platform",
					Namespace: "idpbuilder-system",
				},
				Spec: v1alpha2.PlatformSpec{
					Domain: "test.local",
					Components: v1alpha2.PlatformComponents{
						GitProviders: []v1alpha2.ProviderReference{
							{
								Name:      "gitea-test",
								Kind:      "GiteaProvider",
								Namespace: "idpbuilder-system",
							},
						},
					},
				},
			},
			providers: []client.Object{
				&v1alpha2.GiteaProvider{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "gitea-test",
						Namespace: "idpbuilder-system",
					},
					Status: v1alpha2.GiteaProviderStatus{
						Conditions: []metav1.Condition{
							{
								Type:    "Ready",
								Status:  metav1.ConditionFalse,
								Message: "Deployment not ready",
								Reason:  "DeploymentNotReady",
							},
						},
					},
				},
			},
			expectRequeue: true,
			expectError:   false,
			validateStatus: func(t *testing.T, p *v1alpha2.Platform) {
				assert.Equal(t, "Initializing", p.Status.Phase)
				assert.Len(t, p.Status.Providers.GitProviders, 1)
				assert.False(t, p.Status.Providers.GitProviders[0].Ready)
				assert.Equal(t, "Deployment not ready", p.Status.Providers.GitProviders[0].Message)
				assert.Equal(t, "DeploymentNotReady", p.Status.Providers.GitProviders[0].Reason)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			objs := append(tt.providers, tt.platform)
			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(objs...).
				WithStatusSubresource(&v1alpha2.Platform{}).
				Build()

			reconciler := &PlatformReconciler{
				Client: fakeClient,
				Scheme: scheme,
			}

			req := ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name:      tt.platform.Name,
					Namespace: tt.platform.Namespace,
				},
			}

			result, err := reconciler.Reconcile(context.Background(), req)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			if tt.expectRequeue {
				assert.True(t, result.Requeue || result.RequeueAfter > 0)
			}

			if tt.validateStatus != nil {
				platform := &v1alpha2.Platform{}
				err := fakeClient.Get(context.Background(), req.NamespacedName, platform)
				require.NoError(t, err)
				tt.validateStatus(t, platform)
			}
		})
	}
}

func TestPlatformReconciler_aggregateGitProviders(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = v1alpha2.AddToScheme(scheme)

	tests := []struct {
		name          string
		platform      *v1alpha2.Platform
		providers     []client.Object
		expectReady   bool
		expectSummary int
	}{
		{
			name: "no git providers",
			platform: &v1alpha2.Platform{
				Spec: v1alpha2.PlatformSpec{
					Components: v1alpha2.PlatformComponents{},
				},
			},
			providers:     []client.Object{},
			expectReady:   true,
			expectSummary: 0,
		},
		{
			name: "one ready git provider",
			platform: &v1alpha2.Platform{
				Spec: v1alpha2.PlatformSpec{
					Components: v1alpha2.PlatformComponents{
						GitProviders: []v1alpha2.ProviderReference{
							{
								Name:      "gitea",
								Kind:      "GiteaProvider",
								Namespace: "idpbuilder-system",
							},
						},
					},
				},
			},
			providers: []client.Object{
				&v1alpha2.GiteaProvider{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "gitea",
						Namespace: "idpbuilder-system",
					},
					Status: v1alpha2.GiteaProviderStatus{
						Conditions: []metav1.Condition{
							{
								Type:    "Ready",
								Status:  metav1.ConditionTrue,
								Message: "All components ready",
								Reason:  "Ready",
							},
						},
					},
				},
			},
			expectReady:   true,
			expectSummary: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(tt.providers...).
				Build()

			reconciler := &PlatformReconciler{
				Client: fakeClient,
				Scheme: scheme,
			}

			summaries, ready, err := reconciler.aggregateGitProviders(context.Background(), tt.platform)
			require.NoError(t, err)
			assert.Equal(t, tt.expectReady, ready)
			assert.Len(t, summaries, tt.expectSummary)
		})
	}
}

func TestPlatformReconciler_aggregateGateways(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = v1alpha2.AddToScheme(scheme)

	tests := []struct {
		name          string
		platform      *v1alpha2.Platform
		providers     []client.Object
		expectReady   bool
		expectSummary int
	}{
		{
			name: "no gateways",
			platform: &v1alpha2.Platform{
				Spec: v1alpha2.PlatformSpec{
					Components: v1alpha2.PlatformComponents{},
				},
			},
			providers:     []client.Object{},
			expectReady:   true,
			expectSummary: 0,
		},
		{
			name: "one ready gateway",
			platform: &v1alpha2.Platform{
				Spec: v1alpha2.PlatformSpec{
					Components: v1alpha2.PlatformComponents{
						Gateways: []v1alpha2.ProviderReference{
							{
								Name:      "nginx",
								Kind:      "NginxGateway",
								Namespace: "idpbuilder-system",
							},
						},
					},
				},
			},
			providers: []client.Object{
				&v1alpha2.NginxGateway{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "nginx",
						Namespace: "idpbuilder-system",
					},
					Status: v1alpha2.NginxGatewayStatus{
						Conditions: []metav1.Condition{
							{
								Type:   "Ready",
								Status: metav1.ConditionTrue,
							},
						},
					},
				},
			},
			expectReady:   true,
			expectSummary: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(tt.providers...).
				Build()

			reconciler := &PlatformReconciler{
				Client: fakeClient,
				Scheme: scheme,
			}

			summaries, ready, err := reconciler.aggregateGateways(context.Background(), tt.platform)
			require.NoError(t, err)
			assert.Equal(t, tt.expectReady, ready)
			assert.Len(t, summaries, tt.expectSummary)
		})
	}
}

func TestPlatformReconciler_createBootstrapResources(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = v1alpha2.AddToScheme(scheme)
	_ = v1alpha1.AddToScheme(scheme)
	_ = argov1alpha1.AddToScheme(scheme)

	tests := []struct {
		name          string
		platform      *v1alpha2.Platform
		providers     []client.Object
		expectError   bool
		validateRepos func(*testing.T, client.Client, *v1alpha2.Platform)
	}{
		{
			name: "creates bootstrap resources successfully",
			platform: &v1alpha2.Platform{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-platform",
					Namespace: "default",
					UID:       "platform-uid-123",
				},
				Spec: v1alpha2.PlatformSpec{
					Domain: "test.local",
					Components: v1alpha2.PlatformComponents{
						GitProviders: []v1alpha2.ProviderReference{
							{
								Name:      "test-gitea",
								Kind:      "GiteaProvider",
								Namespace: "gitea",
							},
						},
					},
				},
			},
			providers: []client.Object{
				&v1alpha2.GiteaProvider{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-gitea",
						Namespace: "gitea",
					},
					Status: v1alpha2.GiteaProviderStatus{
						Endpoint:         "http://gitea.test.local",
						InternalEndpoint: "http://gitea-http.gitea.svc.cluster.local:3000",
						CredentialsSecretRef: &v1alpha2.SecretReference{
							Name:      "gitea-credential",
							Namespace: "gitea",
						},
						Conditions: []metav1.Condition{
							{
								Type:   "Ready",
								Status: metav1.ConditionTrue,
							},
						},
					},
				},
			},
			expectError: false,
			validateRepos: func(t *testing.T, c client.Client, p *v1alpha2.Platform) {
				// Check that GitRepository was created
				repo := &v1alpha1.GitRepository{}
				err := c.Get(context.Background(), types.NamespacedName{
					Name:      "argocd",
					Namespace: "idpbuilder-test",
				}, repo)
				require.NoError(t, err)
				assert.Equal(t, "embedded", repo.Spec.Source.Type)
				assert.Equal(t, "argocd", repo.Spec.Source.EmbeddedAppName)
				assert.Equal(t, "http://gitea.test.local", repo.Spec.Provider.GitURL)
				assert.Equal(t, "http://gitea-http.gitea.svc.cluster.local:3000", repo.Spec.Provider.InternalGitURL)

				// Check that platform annotation was set
				updatedPlatform := &v1alpha2.Platform{}
				err = c.Get(context.Background(), types.NamespacedName{
					Name:      p.Name,
					Namespace: p.Namespace,
				}, updatedPlatform)
				require.NoError(t, err)
				assert.Equal(t, "true", updatedPlatform.Annotations[bootstrapReposCreatedFlag])
			},
		},
		{
			name: "skips creation when already created",
			platform: &v1alpha2.Platform{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-platform",
					Namespace: "default",
					UID:       "platform-uid-456",
					Annotations: map[string]string{
						bootstrapReposCreatedFlag: "true",
					},
				},
				Spec: v1alpha2.PlatformSpec{
					Domain: "test.local",
					Components: v1alpha2.PlatformComponents{
						GitProviders: []v1alpha2.ProviderReference{
							{
								Name:      "test-gitea",
								Kind:      "GiteaProvider",
								Namespace: "gitea",
							},
						},
					},
				},
			},
			providers:   []client.Object{},
			expectError: false,
			validateRepos: func(t *testing.T, c client.Client, p *v1alpha2.Platform) {
				// Should not create any repositories
				repo := &v1alpha1.GitRepository{}
				err := c.Get(context.Background(), types.NamespacedName{
					Name:      "argocd",
					Namespace: "idpbuilder-test",
				}, repo)
				assert.True(t, errors.IsNotFound(err))
			},
		},
		{
			name: "returns error when no git providers",
			platform: &v1alpha2.Platform{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-platform",
					Namespace: "default",
				},
				Spec: v1alpha2.PlatformSpec{
					Domain:     "test.local",
					Components: v1alpha2.PlatformComponents{},
				},
			},
			providers:   []client.Object{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(tt.platform).
				WithObjects(tt.providers...).
				Build()

			reconciler := &PlatformReconciler{
				Client: fakeClient,
				Scheme: scheme,
			}

			err := reconciler.createBootstrapResources(context.Background(), tt.platform)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				if tt.validateRepos != nil {
					tt.validateRepos(t, fakeClient, tt.platform)
				}
			}
		})
	}
}

func TestPlatformReconciler_validateGitURL(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		fieldName   string
		expectError bool
	}{
		{
			name:        "valid http URL",
			url:         "http://gitea.test.local",
			fieldName:   "test",
			expectError: false,
		},
		{
			name:        "valid https URL",
			url:         "https://gitea.test.local",
			fieldName:   "test",
			expectError: false,
		},
		{
			name:        "empty URL",
			url:         "",
			fieldName:   "test",
			expectError: true,
		},
		{
			name:        "invalid protocol",
			url:         "ftp://gitea.test.local",
			fieldName:   "test",
			expectError: true,
		},
		{
			name:        "protocol only",
			url:         "http://",
			fieldName:   "test",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateGitURL(tt.url, tt.fieldName)
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
