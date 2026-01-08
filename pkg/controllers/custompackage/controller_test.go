package custompackage

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	argov1alpha1 "github.com/cnoe-io/argocd-api/api/argo/application/v1alpha1"
	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"github.com/cnoe-io/idpbuilder/api/v1alpha2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

type testCase struct {
	expectedGitRepo        v1alpha1.GitRepository
	expectedApplicationSet argov1alpha1.ApplicationSet
	input                  v1alpha1.CustomPackage
}

func TestReconcileCustomPkg(t *testing.T) {
	s := k8sruntime.NewScheme()
	sb := k8sruntime.NewSchemeBuilder(
		v1.AddToScheme,
		argov1alpha1.AddToScheme,
		v1alpha1.AddToScheme,
		v1alpha2.AddToScheme,
	)
	require.NoError(t, sb.AddToScheme(s))

	cwd, err := os.Getwd()
	require.NoError(t, err)

	customPkgs := []v1alpha1.CustomPackage{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test1",
				Namespace: "test",
				UID:       "abc",
			},
			Spec: v1alpha1.CustomPackageSpec{
				Replicate:           true,
				GitServerURL:        "https://cnoe.io",
				InternalGitServeURL: "http://internal.cnoe.io",
				ArgoCD: v1alpha1.ArgoCDPackageSpec{
					ApplicationFile: filepath.Join(cwd, "test/resources/customPackages/testDir/app.yaml"),
					Name:            "my-app",
					Namespace:       "argocd",
					Type:            "Application",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test2",
				Namespace: "test",
				UID:       "abc",
			},
			Spec: v1alpha1.CustomPackageSpec{
				Replicate:           false,
				GitServerURL:        "https://cnoe.io",
				InternalGitServeURL: "http://cnoe.io/internal",
				ArgoCD: v1alpha1.ArgoCDPackageSpec{
					ApplicationFile: filepath.Join(cwd, "test/resources/customPackages/testDir2/exampleApp.yaml"),
					Name:            "guestbook",
					Namespace:       "argocd",
					Type:            "Application",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test3",
				Namespace: "test",
				UID:       "abc",
			},
			Spec: v1alpha1.CustomPackageSpec{
				Replicate:           true,
				GitServerURL:        "https://cnoe.io",
				InternalGitServeURL: "http://internal.cnoe.io",
				ArgoCD: v1alpha1.ArgoCDPackageSpec{
					ApplicationFile: filepath.Join(cwd, "test/resources/customPackages/testDir/app2.yaml"),
					Name:            "my-app2",
					Namespace:       "argocd",
					Type:            "Application",
				},
			},
		},
	}

	// Create namespaces
	ns1 := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "argocd",
		},
	}
	ns2 := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
	}

	// Create fake client with initial objects
	fakeClient := fake.NewClientBuilder().
		WithScheme(s).
		WithObjects(ns1, ns2).
		WithStatusSubresource(&v1alpha1.CustomPackage{}, &v1alpha1.GitRepository{}).
		Build()

	r := &Reconciler{
		Client:   fakeClient,
		Scheme:   s,
		Recorder: record.NewFakeRecorder(100),
	}

	// Reconcile each custom package
	for i := range customPkgs {
		_, err = r.reconcileCustomPackage(context.Background(), &customPkgs[i])
		if err != nil {
			t.Fatalf("reconciling custom packages %v", err)
		}
	}

	// Give some time for reconciliation to complete (simulate async processing)
	time.Sleep(100 * time.Millisecond)

	// verify repo.
	repo := v1alpha1.GitRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      localRepoName("my-app", "test/resources/customPackages/testDir/app1"),
			Namespace: "test",
		},
	}
	err = fakeClient.Get(context.Background(), client.ObjectKeyFromObject(&repo), &repo)
	if err != nil {
		t.Fatalf("getting my-app-app1 git repo %v", err)
	}

	p, _ := filepath.Abs("test/resources/customPackages/testDir/app1")
	expectedRepo := v1alpha1.GitRepository{
		Spec: v1alpha1.GitRepositorySpec{
			Source: v1alpha1.GitRepositorySource{
				Type: "local",
				Path: p,
			},
			Provider: v1alpha1.Provider{
				Name:             v1alpha1.GitProviderGitea,
				GitURL:           "https://cnoe.io",
				InternalGitURL:   "http://internal.cnoe.io",
				OrganizationName: v1alpha1.GiteaAdminUserName,
			},
		},
	}
	assert.Equal(t, repo.Spec, expectedRepo.Spec)
	ok := reflect.DeepEqual(repo.Spec, expectedRepo.Spec)
	assert.True(t, ok)

	tcs := []struct {
		name string
	}{
		{
			name: "my-app",
		},
		{
			name: "my-app2",
		},
		{
			name: "guestbook",
		},
	}

	for _, tc := range tcs {
		app := argov1alpha1.Application{
			ObjectMeta: metav1.ObjectMeta{
				Name:      tc.name,
				Namespace: "argocd",
			},
		}
		err = fakeClient.Get(context.Background(), client.ObjectKeyFromObject(&app), &app)
		assert.NoError(t, err)

		if app.ObjectMeta.Labels == nil {
			t.Fatalf("labels not set")
		}

		_, ok := app.ObjectMeta.Labels[v1alpha1.PackageTypeLabelKey]
		if !ok {
			t.Fatalf("label %s not set", v1alpha1.PackageTypeLabelKey)
		}

		_, ok = app.ObjectMeta.Labels[v1alpha1.PackageNameLabelKey]
		if !ok {
			t.Fatalf("label %s not set", v1alpha1.PackageNameLabelKey)
		}

		if app.Spec.Sources == nil {
			if strings.HasPrefix(app.Spec.Source.RepoURL, v1alpha1.CNOEURIScheme) {
				t.Fatalf("%s prefix should be removed", v1alpha1.CNOEURIScheme)
			}
			continue
		}
		for _, s := range app.Spec.Sources {
			if strings.HasPrefix(s.RepoURL, v1alpha1.CNOEURIScheme) {
				t.Fatalf("%s prefix should be removed", v1alpha1.CNOEURIScheme)
			}
		}

	}
}

func TestReconcileCustomPkgAppSet(t *testing.T) {
	s := k8sruntime.NewScheme()
	sb := k8sruntime.NewSchemeBuilder(
		v1.AddToScheme,
		argov1alpha1.AddToScheme,
		v1alpha1.AddToScheme,
		v1alpha2.AddToScheme,
	)
	require.NoError(t, sb.AddToScheme(s))

	cwd, err := os.Getwd()
	require.NoError(t, err)

	// Create namespaces
	ns1 := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "argocd",
		},
	}
	ns2 := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
	}

	// Create fake client with initial objects
	fakeClient := fake.NewClientBuilder().
		WithScheme(s).
		WithObjects(ns1, ns2).
		WithStatusSubresource(&v1alpha1.CustomPackage{}, &v1alpha1.GitRepository{}).
		Build()

	r := &Reconciler{
		Client:   fakeClient,
		Scheme:   s,
		Recorder: record.NewFakeRecorder(100),
	}

	cases := []testCase{
		{
			input: v1alpha1.CustomPackage{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test1",
					Namespace: "test",
					UID:       "abc",
				},
				Spec: v1alpha1.CustomPackageSpec{
					Replicate:           true,
					GitServerURL:        "https://cnoe.io",
					InternalGitServeURL: "http://internal.cnoe.io",
					ArgoCD: v1alpha1.ArgoCDPackageSpec{
						ApplicationFile: filepath.Join(cwd, "test/resources/customPackages/applicationSet/generator-single-source.yaml"),
						Type:            "ApplicationSet",
					},
				},
			},
			expectedGitRepo: v1alpha1.GitRepository{
				ObjectMeta: metav1.ObjectMeta{
					Name:      localRepoName("generator-single-source", "test/resources/customPackages/applicationSet/test1"),
					Namespace: "test",
				},
				Spec: v1alpha1.GitRepositorySpec{
					Source: v1alpha1.GitRepositorySource{
						Type: "local",
						Path: filepath.Join(cwd, "test/resources/customPackages/applicationSet/test1"),
					},
					Provider: v1alpha1.Provider{
						Name:             v1alpha1.GitProviderGitea,
						GitURL:           "https://cnoe.io",
						InternalGitURL:   "http://internal.cnoe.io",
						OrganizationName: v1alpha1.GiteaAdminUserName,
					},
				},
			},
			expectedApplicationSet: argov1alpha1.ApplicationSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "generator-single-source",
					Namespace: "argocd",
				},
				Spec: argov1alpha1.ApplicationSetSpec{
					Generators: []argov1alpha1.ApplicationSetGenerator{
						{
							Git: &argov1alpha1.GitGenerator{
								RepoURL: "",
							},
						},
					},
					Template: argov1alpha1.ApplicationSetTemplate{
						Spec: argov1alpha1.ApplicationSpec{
							Source: &argov1alpha1.ApplicationSource{
								RepoURL: "",
							},
						},
					},
				},
			},
		},
		{
			input: v1alpha1.CustomPackage{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test2",
					Namespace: "test",
					UID:       "test2",
				},
				Spec: v1alpha1.CustomPackageSpec{
					Replicate:           true,
					GitServerURL:        "https://cnoe.io",
					InternalGitServeURL: "http://internal.cnoe.io",
					ArgoCD: v1alpha1.ArgoCDPackageSpec{
						ApplicationFile: filepath.Join(cwd, "test/resources/customPackages/applicationSet/generator-multi-sources.yaml"),
						Type:            "ApplicationSet",
					},
				},
			},
			expectedGitRepo: v1alpha1.GitRepository{
				ObjectMeta: metav1.ObjectMeta{
					Name:      localRepoName("generator-multi-sources", "test/resources/customPackages/applicationSet/test1"),
					Namespace: "test",
				},
				Spec: v1alpha1.GitRepositorySpec{
					Source: v1alpha1.GitRepositorySource{
						Type: "local",
						Path: filepath.Join(cwd, "test/resources/customPackages/applicationSet/test1"),
					},
					Provider: v1alpha1.Provider{
						Name:             v1alpha1.GitProviderGitea,
						GitURL:           "https://cnoe.io",
						InternalGitURL:   "http://internal.cnoe.io",
						OrganizationName: v1alpha1.GiteaAdminUserName,
					},
				},
			},
			expectedApplicationSet: argov1alpha1.ApplicationSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "generator-multi-sources",
					Namespace: "argocd",
				},
				Spec: argov1alpha1.ApplicationSetSpec{
					Generators: []argov1alpha1.ApplicationSetGenerator{
						{
							Git: &argov1alpha1.GitGenerator{
								RepoURL: "",
							},
						},
					},
					Template: argov1alpha1.ApplicationSetTemplate{
						Spec: argov1alpha1.ApplicationSpec{
							Sources: []argov1alpha1.ApplicationSource{
								{
									RepoURL: "",
								},
							},
						},
					},
				},
			},
		},
		{
			input: v1alpha1.CustomPackage{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test3",
					Namespace: "test",
					UID:       "test3",
				},
				Spec: v1alpha1.CustomPackageSpec{
					Replicate:           true,
					GitServerURL:        "https://cnoe.io",
					InternalGitServeURL: "http://internal.cnoe.io",
					ArgoCD: v1alpha1.ArgoCDPackageSpec{
						ApplicationFile: filepath.Join(cwd, "test/resources/customPackages/applicationSet/no-generator-single-source.yaml"),
						Type:            "ApplicationSet",
					},
				},
			},
			expectedGitRepo: v1alpha1.GitRepository{
				ObjectMeta: metav1.ObjectMeta{
					Name:      localRepoName("no-generator-single-source", "test/resources/customPackages/applicationSet/test1"),
					Namespace: "test",
				},
				Spec: v1alpha1.GitRepositorySpec{
					Source: v1alpha1.GitRepositorySource{
						Type: "local",
						Path: filepath.Join(cwd, "test/resources/customPackages/applicationSet/test1"),
					},
					Provider: v1alpha1.Provider{
						Name:             v1alpha1.GitProviderGitea,
						GitURL:           "https://cnoe.io",
						InternalGitURL:   "http://internal.cnoe.io",
						OrganizationName: v1alpha1.GiteaAdminUserName,
					},
				},
			},
			expectedApplicationSet: argov1alpha1.ApplicationSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "no-generator-single-source",
					Namespace: "argocd",
				},
				Spec: argov1alpha1.ApplicationSetSpec{
					Template: argov1alpha1.ApplicationSetTemplate{
						Spec: argov1alpha1.ApplicationSpec{
							Source: &argov1alpha1.ApplicationSource{
								RepoURL: "",
							},
						},
					},
				},
			},
		},
		{
			input: v1alpha1.CustomPackage{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test4",
					Namespace: "test",
					UID:       "test4",
				},
				Spec: v1alpha1.CustomPackageSpec{
					Replicate:           true,
					GitServerURL:        "https://cnoe.io",
					InternalGitServeURL: "http://internal.cnoe.io",
					ArgoCD: v1alpha1.ArgoCDPackageSpec{
						ApplicationFile: filepath.Join(cwd, "test/resources/customPackages/applicationSet/generator-matrix.yaml"),
						Type:            "ApplicationSet",
					},
				},
			},
			expectedGitRepo: v1alpha1.GitRepository{
				ObjectMeta: metav1.ObjectMeta{
					Name:      localRepoName("generator-matrix", "test/resources/customPackages/applicationSet/test1"),
					Namespace: "test",
				},
				Spec: v1alpha1.GitRepositorySpec{
					Source: v1alpha1.GitRepositorySource{
						Type: "local",
						Path: filepath.Join(cwd, "test/resources/customPackages/applicationSet/test1"),
					},
					Provider: v1alpha1.Provider{
						Name:             v1alpha1.GitProviderGitea,
						GitURL:           "https://cnoe.io",
						InternalGitURL:   "http://internal.cnoe.io",
						OrganizationName: v1alpha1.GiteaAdminUserName,
					},
				},
			},
			expectedApplicationSet: argov1alpha1.ApplicationSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "generator-matrix",
					Namespace: "argocd",
				},
				Spec: argov1alpha1.ApplicationSetSpec{
					Generators: []argov1alpha1.ApplicationSetGenerator{
						{
							Matrix: &argov1alpha1.MatrixGenerator{
								Generators: []argov1alpha1.ApplicationSetNestedGenerator{
									{
										Git: &argov1alpha1.GitGenerator{
											RepoURL: "",
										},
									},
								},
							},
						},
					},
					Template: argov1alpha1.ApplicationSetTemplate{
						Spec: argov1alpha1.ApplicationSpec{
							Source: &argov1alpha1.ApplicationSource{
								RepoURL: "",
							},
						},
					},
				},
			},
		},
	}

	for i := range cases {
		tc := cases[i]
		_, err = r.reconcileCustomPackage(context.Background(), &tc.input)
		assert.NoError(t, err)
		time.Sleep(100 * time.Millisecond)

		repo := v1alpha1.GitRepository{}
		err = fakeClient.Get(context.Background(), client.ObjectKeyFromObject(&tc.expectedGitRepo), &repo)
		assert.NoError(t, err)

		assert.Equal(t, tc.expectedGitRepo.Spec, repo.Spec)

		// verify argocd applicationSet
		appset := argov1alpha1.ApplicationSet{}
		err = fakeClient.Get(context.Background(), client.ObjectKeyFromObject(&tc.expectedApplicationSet), &appset)
		assert.NoError(t, err)

		if len(tc.expectedApplicationSet.Spec.Template.Spec.Sources) > 0 {
			for j := range tc.expectedApplicationSet.Spec.Template.Spec.Sources {
				exs := tc.expectedApplicationSet.Spec.Template.Spec.Sources[j]
				assert.Equal(t, exs.RepoURL, appset.Spec.Template.Spec.Sources[j].RepoURL)
				assert.False(t, strings.HasPrefix(appset.Spec.Template.Spec.Sources[j].RepoURL, v1alpha1.CNOEURIScheme))
			}
		} else {
			assert.Equal(t, tc.expectedApplicationSet.Spec.Template.Spec.Source.RepoURL, appset.Spec.Template.Spec.Source.RepoURL)
			assert.False(t, strings.HasPrefix(appset.Spec.Template.Spec.Source.RepoURL, v1alpha1.CNOEURIScheme))
		}

		if len(tc.expectedApplicationSet.Spec.Generators) > 0 {
			for j := range tc.expectedApplicationSet.Spec.Generators {
				exg := tc.expectedApplicationSet.Spec.Generators[j]
				if exg.Git != nil {
					assert.Equal(t, exg.Git.RepoURL, appset.Spec.Generators[j].Git.RepoURL)
				}
				if exg.Matrix != nil {
					for k := range exg.Matrix.Generators {
						if exg.Matrix.Generators[k].Git != nil {
							assert.Equal(t, exg.Matrix.Generators[k].Git.RepoURL, appset.Spec.Generators[j].Matrix.Generators[k].Git.RepoURL)
						}
					}
				}
			}
		}
	}
}

func TestReconcileHelmValueObject(t *testing.T) {
	s := k8sruntime.NewScheme()
	sb := k8sruntime.NewSchemeBuilder(
		v1.AddToScheme,
		argov1alpha1.AddToScheme,
		v1alpha1.AddToScheme,
		v1alpha2.AddToScheme,
	)
	require.NoError(t, sb.AddToScheme(s))

	cwd, err := os.Getwd()
	require.NoError(t, err)

	// Create namespaces
	ns1 := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "argocd",
		},
	}
	ns2 := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
	}

	// Create fake client with initial objects
	fakeClient := fake.NewClientBuilder().
		WithScheme(s).
		WithObjects(ns1, ns2).
		WithStatusSubresource(&v1alpha1.CustomPackage{}, &v1alpha1.GitRepository{}).
		Build()

	r := &Reconciler{
		Client:   fakeClient,
		Scheme:   s,
		Recorder: record.NewFakeRecorder(100),
	}

	resource := v1alpha1.CustomPackage{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test1",
			Namespace: "test",
			UID:       "abc",
		},
		Spec: v1alpha1.CustomPackageSpec{
			Replicate:           true,
			GitServerURL:        "https://cnoe.io",
			InternalGitServeURL: "http://internal.cnoe.io",
			ArgoCD: v1alpha1.ArgoCDPackageSpec{
				ApplicationFile: filepath.Join(cwd, "test/resources/customPackages/helm/app.yaml"),
				Name:            "my-app",
				Namespace:       "argocd",
				Type:            "Application",
			},
		},
	}

	source := &argov1alpha1.ApplicationSource{
		Helm: &argov1alpha1.ApplicationSourceHelm{
			ValuesObject: &k8sruntime.RawExtension{
				Raw: []byte(`{
				 "repoURLGit": "cnoe://test",
				 "nested": {
				   "repoURLGit": "cnoe://test",
				   "bool": true,
				   "int": 123
				 },
				 "bool": false,
				 "int": 456,
				 "arrayString": [
				   "abc",
				   "cnoe://test"
				 ],
				 "arrayMap": [
				   {
				     "test": "cnoe://test",
				     "nested": {
				       "test": "cnoe://test"
				     }
				   }
				 ]
				}`),
			},
		},
	}

	_, err = r.reconcileHelmValueObject(context.Background(), source, &resource, "test")
	assert.NoError(t, err)
	expectJson := `{"arrayMap":[{"nested":{"test":""},"test":""}],"arrayString":["abc",""],"bool":false,"int":456,"nested":{"bool":true,"int":123,"repoURLGit":""},"repoURLGit":""}`
	assert.JSONEq(t, expectJson, string(source.Helm.ValuesObject.Raw))
}

func TestPackagePriority(t *testing.T) {
	s := k8sruntime.NewScheme()
	sb := k8sruntime.NewSchemeBuilder(
		v1.AddToScheme,
		argov1alpha1.AddToScheme,
		v1alpha1.AddToScheme,
		v1alpha2.AddToScheme,
	)
	require.NoError(t, sb.AddToScheme(s))

	cwd, err := os.Getwd()
	require.NoError(t, err)

	// Create namespaces
	ns1 := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "argocd",
		},
	}
	ns2 := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
	}

	// Create two CustomPackages with the same app name but different priorities
	// Package 1 has priority 0 (from first --package argument)
	pkg1 := &v1alpha1.CustomPackage{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pkg1-my-app",
			Namespace: "test",
			UID:       "pkg1",
			Annotations: map[string]string{
				v1alpha1.PackagePriorityAnnotation:   "0",
				v1alpha1.PackageSourcePathAnnotation: "/path/to/package1",
			},
		},
		Spec: v1alpha1.CustomPackageSpec{
			Replicate:           true,
			GitServerURL:        "https://cnoe.io",
			InternalGitServeURL: "http://internal.cnoe.io",
			ArgoCD: v1alpha1.ArgoCDPackageSpec{
				ApplicationFile: filepath.Join(cwd, "test/resources/customPackages/testDir/app.yaml"),
				Name:            "my-app",
				Namespace:       "argocd",
				Type:            "Application",
			},
		},
	}

	// Package 2 has priority 1 (from second --package argument, should win)
	pkg2 := &v1alpha1.CustomPackage{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pkg2-my-app",
			Namespace: "test",
			UID:       "pkg2",
			Annotations: map[string]string{
				v1alpha1.PackagePriorityAnnotation:   "1",
				v1alpha1.PackageSourcePathAnnotation: "/path/to/package2",
			},
		},
		Spec: v1alpha1.CustomPackageSpec{
			Replicate:           true,
			GitServerURL:        "https://cnoe.io",
			InternalGitServeURL: "http://internal.cnoe.io",
			ArgoCD: v1alpha1.ArgoCDPackageSpec{
				ApplicationFile: filepath.Join(cwd, "test/resources/customPackages/testDir/app.yaml"),
				Name:            "my-app",
				Namespace:       "argocd",
				Type:            "Application",
			},
		},
	}

	// Create fake client with initial objects including both packages
	fakeClient := fake.NewClientBuilder().
		WithScheme(s).
		WithObjects(ns1, ns2, pkg1, pkg2).
		WithStatusSubresource(&v1alpha1.CustomPackage{}, &v1alpha1.GitRepository{}).
		Build()

	r := &Reconciler{
		Client:   fakeClient,
		Scheme:   s,
		Recorder: record.NewFakeRecorder(100),
	}

	// Test priority resolution
	t.Run("lower priority package should not reconcile", func(t *testing.T) {
		shouldReconcile, err := r.shouldReconcile(context.Background(), pkg1)
		assert.NoError(t, err)
		assert.False(t, shouldReconcile, "pkg1 (priority 0) should not reconcile when pkg2 (priority 1) exists")
	})

	t.Run("higher priority package should reconcile", func(t *testing.T) {
		shouldReconcile, err := r.shouldReconcile(context.Background(), pkg2)
		assert.NoError(t, err)
		assert.True(t, shouldReconcile, "pkg2 (priority 1) should reconcile as it has highest priority")
	})

	t.Run("getPackagePriority should extract priority correctly", func(t *testing.T) {
		priority1, err := getPackagePriority(pkg1)
		assert.NoError(t, err)
		assert.Equal(t, 0, priority1)

		priority2, err := getPackagePriority(pkg2)
		assert.NoError(t, err)
		assert.Equal(t, 1, priority2)
	})
}

func TestGetPackagePriority(t *testing.T) {
	t.Run("valid priority annotation", func(t *testing.T) {
		pkg := &v1alpha1.CustomPackage{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					v1alpha1.PackagePriorityAnnotation: "5",
				},
			},
		}
		priority, err := getPackagePriority(pkg)
		assert.NoError(t, err)
		assert.Equal(t, 5, priority)
	})

	t.Run("missing annotations", func(t *testing.T) {
		pkg := &v1alpha1.CustomPackage{
			ObjectMeta: metav1.ObjectMeta{},
		}
		_, err := getPackagePriority(pkg)
		assert.Error(t, err)
	})

	t.Run("missing priority annotation", func(t *testing.T) {
		pkg := &v1alpha1.CustomPackage{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					"other": "value",
				},
			},
		}
		_, err := getPackagePriority(pkg)
		assert.Error(t, err)
	})

	t.Run("invalid priority format", func(t *testing.T) {
		pkg := &v1alpha1.CustomPackage{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					v1alpha1.PackagePriorityAnnotation: "invalid",
				},
			},
		}
		_, err := getPackagePriority(pkg)
		assert.Error(t, err)
	})

	t.Run("zero priority", func(t *testing.T) {
		pkg := &v1alpha1.CustomPackage{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					v1alpha1.PackagePriorityAnnotation: "0",
				},
			},
		}
		priority, err := getPackagePriority(pkg)
		assert.NoError(t, err)
		assert.Equal(t, 0, priority)
	})

	t.Run("large priority value", func(t *testing.T) {
		pkg := &v1alpha1.CustomPackage{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					v1alpha1.PackagePriorityAnnotation: "1000",
				},
			},
		}
		priority, err := getPackagePriority(pkg)
		assert.NoError(t, err)
		assert.Equal(t, 1000, priority)
	})
}

func TestDiscoverGitProviderFromPlatform(t *testing.T) {
	s := k8sruntime.NewScheme()
	sb := k8sruntime.NewSchemeBuilder(
		v1.AddToScheme,
		argov1alpha1.AddToScheme,
		v1alpha1.AddToScheme,
		v1alpha2.AddToScheme,
	)
	require.NoError(t, sb.AddToScheme(s))

	t.Run("discover git provider with explicit Platform reference", func(t *testing.T) {
		ctx := context.Background()

		// Create a GiteaProvider CR
		giteaProvider := &v1alpha2.GiteaProvider{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "gitea",
				Namespace: "test-ns",
			},
			Spec: v1alpha2.GiteaProviderSpec{
				Namespace: "gitea",
				Host:      "gitea.test.com",
			},
			Status: v1alpha2.GiteaProviderStatus{
				Endpoint:         "https://gitea.test.com",
				InternalEndpoint: "http://gitea.svc.cluster.local:3000",
				CredentialsSecretRef: &v1alpha2.SecretReference{
					Name:      "gitea-credentials",
					Namespace: "test-ns",
					Key:       "password",
				},
				Conditions: []metav1.Condition{
					{
						Type:   "Ready",
						Status: metav1.ConditionTrue,
					},
				},
			},
		}

		// Create a Platform CR referencing the GiteaProvider
		platform := &v1alpha2.Platform{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-platform",
				Namespace: "test-ns",
			},
			Spec: v1alpha2.PlatformSpec{
				Domain: "test.com",
				Components: v1alpha2.PlatformComponents{
					GitProviders: []v1alpha2.ProviderReference{
						{
							Name:      "gitea",
							Kind:      "GiteaProvider",
							Namespace: "test-ns",
						},
					},
				},
			},
		}

		// Create a CustomPackage with explicit Platform reference
		customPkg := &v1alpha1.CustomPackage{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pkg",
				Namespace: "test-ns",
			},
			Spec: v1alpha1.CustomPackageSpec{
				PlatformRef: &v1alpha1.PlatformReference{
					Name: "my-platform",
				},
			},
		}

		// Create fake client with initial objects
		fakeClient := fake.NewClientBuilder().
			WithScheme(s).
			WithObjects(giteaProvider, platform).
			Build()

		reconciler := &Reconciler{
			Client:   fakeClient,
			Recorder: record.NewFakeRecorder(10),
			Scheme:   s,
		}

		// Test discovery with explicit reference
		providerInfo, err := reconciler.discoverGitProviderFromPlatform(ctx, customPkg)
		require.NoError(t, err)
		require.NotNil(t, providerInfo)

		assert.Equal(t, "https://gitea.test.com", providerInfo.externalURL)
		assert.Equal(t, "http://gitea.svc.cluster.local:3000", providerInfo.internalURL)
		assert.Equal(t, "gitea-credentials", providerInfo.secretName)
		assert.Equal(t, "test-ns", providerInfo.secretNamespace)
		assert.Equal(t, v1alpha1.GiteaAdminUserName, providerInfo.organizationName)
	})

	t.Run("discover git provider with default Platform", func(t *testing.T) {
		ctx := context.Background()

		// Create a GiteaProvider CR
		giteaProvider := &v1alpha2.GiteaProvider{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "gitea",
				Namespace: "test-ns",
			},
			Spec: v1alpha2.GiteaProviderSpec{
				Namespace: "gitea",
				Host:      "gitea.test.com",
			},
			Status: v1alpha2.GiteaProviderStatus{
				Endpoint:         "https://gitea.test.com",
				InternalEndpoint: "http://gitea.svc.cluster.local:3000",
				CredentialsSecretRef: &v1alpha2.SecretReference{
					Name:      "gitea-credentials",
					Namespace: "test-ns",
					Key:       "password",
				},
				Conditions: []metav1.Condition{
					{
						Type:   "Ready",
						Status: metav1.ConditionTrue,
					},
				},
			},
		}

		// Create a default Platform CR named "platform"
		platform := &v1alpha2.Platform{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "platform",
				Namespace: "test-ns",
			},
			Spec: v1alpha2.PlatformSpec{
				Domain: "test.com",
				Components: v1alpha2.PlatformComponents{
					GitProviders: []v1alpha2.ProviderReference{
						{
							Name:      "gitea",
							Kind:      "GiteaProvider",
							Namespace: "test-ns",
						},
					},
				},
			},
		}

		// Create a CustomPackage without explicit Platform reference
		customPkg := &v1alpha1.CustomPackage{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pkg",
				Namespace: "test-ns",
			},
			Spec: v1alpha1.CustomPackageSpec{
				// No PlatformRef - should use default "platform"
			},
		}

		// Create fake client with initial objects
		fakeClient := fake.NewClientBuilder().
			WithScheme(s).
			WithObjects(giteaProvider, platform).
			Build()

		reconciler := &Reconciler{
			Client:   fakeClient,
			Recorder: record.NewFakeRecorder(10),
			Scheme:   s,
		}

		// Test discovery with default platform
		providerInfo, err := reconciler.discoverGitProviderFromPlatform(ctx, customPkg)
		require.NoError(t, err)
		require.NotNil(t, providerInfo)

		assert.Equal(t, "https://gitea.test.com", providerInfo.externalURL)
		assert.Equal(t, "http://gitea.svc.cluster.local:3000", providerInfo.internalURL)
	})

	t.Run("no Platform CR - fallback to spec", func(t *testing.T) {
		ctx := context.Background()

		// Create a CustomPackage without Platform reference and no default Platform
		customPkg := &v1alpha1.CustomPackage{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pkg",
				Namespace: "test-ns",
			},
			Spec: v1alpha1.CustomPackageSpec{
				// No PlatformRef and no default "platform" exists
			},
		}

		fakeClient := fake.NewClientBuilder().
			WithScheme(s).
			Build()

		reconciler := &Reconciler{
			Client:   fakeClient,
			Recorder: record.NewFakeRecorder(10),
			Scheme:   s,
		}

		// Test discovery with no Platform
		providerInfo, err := reconciler.discoverGitProviderFromPlatform(ctx, customPkg)
		require.NoError(t, err)
		assert.Nil(t, providerInfo) // Should return nil for backward compatibility
	})

	t.Run("explicit Platform reference not found - error", func(t *testing.T) {
		ctx := context.Background()

		// Create a CustomPackage with explicit Platform reference that doesn't exist
		customPkg := &v1alpha1.CustomPackage{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pkg",
				Namespace: "test-ns",
			},
			Spec: v1alpha1.CustomPackageSpec{
				PlatformRef: &v1alpha1.PlatformReference{
					Name: "nonexistent-platform",
				},
			},
		}

		fakeClient := fake.NewClientBuilder().
			WithScheme(s).
			Build()

		reconciler := &Reconciler{
			Client:   fakeClient,
			Recorder: record.NewFakeRecorder(10),
			Scheme:   s,
		}

		// Test discovery with nonexistent explicit reference - should error
		providerInfo, err := reconciler.discoverGitProviderFromPlatform(ctx, customPkg)
		require.Error(t, err)
		assert.Nil(t, providerInfo)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("Platform CR with no git providers - error", func(t *testing.T) {
		ctx := context.Background()

		// Create a Platform CR with no git providers
		platform := &v1alpha2.Platform{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "platform",
				Namespace: "test-ns",
			},
			Spec: v1alpha2.PlatformSpec{
				Domain:     "test.com",
				Components: v1alpha2.PlatformComponents{},
			},
		}

		// Create a CustomPackage without explicit reference (will use default)
		customPkg := &v1alpha1.CustomPackage{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pkg",
				Namespace: "test-ns",
			},
			Spec: v1alpha1.CustomPackageSpec{
				// No PlatformRef - will try default "platform"
			},
		}

		fakeClient := fake.NewClientBuilder().
			WithScheme(s).
			WithObjects(platform).
			Build()

		reconciler := &Reconciler{
			Client:   fakeClient,
			Recorder: record.NewFakeRecorder(10),
			Scheme:   s,
		}

		// Test discovery - should error since platform has no git providers
		providerInfo, err := reconciler.discoverGitProviderFromPlatform(ctx, customPkg)
		require.Error(t, err)
		assert.Nil(t, providerInfo)
		assert.Contains(t, err.Error(), "no git providers")
	})

	t.Run("fallback to CustomPackage spec when no Platform", func(t *testing.T) {
		ctx := context.Background()

		customPkg := &v1alpha1.CustomPackage{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pkg",
				Namespace: "test-ns",
			},
			Spec: v1alpha1.CustomPackageSpec{
				GitServerURL:        "https://custom.gitea.io",
				InternalGitServeURL: "http://custom.internal",
				GitServerAuthSecretRef: v1alpha1.SecretReference{
					Name:      "custom-secret",
					Namespace: "test-ns",
				},
			},
		}

		fakeClient := fake.NewClientBuilder().
			WithScheme(s).
			WithObjects(customPkg).
			Build()

		reconciler := &Reconciler{
			Client:   fakeClient,
			Recorder: record.NewFakeRecorder(10),
			Scheme:   s,
		}

		// Test getGitProviderInfo fallback
		providerInfo, err := reconciler.getGitProviderInfo(ctx, customPkg)
		require.NoError(t, err)
		require.NotNil(t, providerInfo)

		assert.Equal(t, "https://custom.gitea.io", providerInfo.externalURL)
		assert.Equal(t, "http://custom.internal", providerInfo.internalURL)
		assert.Equal(t, "custom-secret", providerInfo.secretName)
		assert.Equal(t, "test-ns", providerInfo.secretNamespace)
	})
}
