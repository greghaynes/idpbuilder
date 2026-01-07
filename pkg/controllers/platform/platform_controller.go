package platform

import (
	"context"
	"fmt"
	"strings"
	"time"

	argov1alpha1 "github.com/cnoe-io/argocd-api/api/argo/application/v1alpha1"
	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"github.com/cnoe-io/idpbuilder/api/v1alpha2"
	"github.com/cnoe-io/idpbuilder/globals"
	"github.com/cnoe-io/idpbuilder/pkg/resources/localbuild"
	"github.com/cnoe-io/idpbuilder/pkg/util"
	"github.com/cnoe-io/idpbuilder/pkg/util/provider"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	platformFinalizer         = "platform.idpbuilder.cnoe.io/finalizer"
	defaultRequeueTime        = time.Second * 30
	defaultArgoCDProjectName  = "default"
	bootstrapReposCreatedFlag = "platform.idpbuilder.cnoe.io/bootstrap-repos-created"
)

// PlatformReconciler reconciles a Platform object
type PlatformReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=idpbuilder.cnoe.io,resources=platforms,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=idpbuilder.cnoe.io,resources=platforms/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=idpbuilder.cnoe.io,resources=platforms/finalizers,verbs=update
//+kubebuilder:rbac:groups=idpbuilder.cnoe.io,resources=giteaproviders,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=idpbuilder.cnoe.io,resources=nginxgateways,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=idpbuilder.cnoe.io,resources=argocdproviders,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=idpbuilder.cnoe.io,resources=gitrepositories,verbs=get;list;watch;create;update;patch
//+kubebuilder:rbac:groups=argoproj.io,resources=applications,verbs=get;list;watch;create;update;patch

// Reconcile is part of the main kubernetes reconciliation loop
func (r *PlatformReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.V(1).Info("Reconciling Platform", "resource", req.NamespacedName)

	// Fetch the Platform instance
	platform := &v1alpha2.Platform{}
	if err := r.Get(ctx, req.NamespacedName, platform); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Platform resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get Platform")
		return ctrl.Result{}, err
	}

	// Add finalizer if not present
	if !controllerutil.ContainsFinalizer(platform, platformFinalizer) {
		controllerutil.AddFinalizer(platform, platformFinalizer)
		if err := r.Update(ctx, platform); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Handle deletion
	if !platform.DeletionTimestamp.IsZero() {
		return r.handleDeletion(ctx, platform)
	}

	// Update phase to Pending if not set
	if platform.Status.Phase == "" {
		platform.Status.Phase = "Pending"
		if err := r.Status().Update(ctx, platform); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Establish owner references to all provider CRs
	if err := r.ensureProviderOwnerReferences(ctx, platform); err != nil {
		logger.Error(err, "Failed to ensure provider owner references")
		return ctrl.Result{RequeueAfter: defaultRequeueTime}, nil
	}

	// Aggregate provider statuses
	allReady := true

	// Aggregate Git Providers
	gitProviderStatuses, gitReady, err := r.aggregateGitProviders(ctx, platform)
	if err != nil {
		logger.Error(err, "Failed to aggregate git providers")
		return ctrl.Result{RequeueAfter: defaultRequeueTime}, nil
	}
	platform.Status.Providers.GitProviders = gitProviderStatuses
	if !gitReady {
		allReady = false
	}

	// Aggregate Gateway Providers
	gatewayStatuses, gatewayReady, err := r.aggregateGateways(ctx, platform)
	if err != nil {
		logger.Error(err, "Failed to aggregate gateways")
		return ctrl.Result{RequeueAfter: defaultRequeueTime}, nil
	}
	platform.Status.Providers.Gateways = gatewayStatuses
	if !gatewayReady {
		allReady = false
	}

	// Aggregate GitOps Providers
	gitopsStatuses, gitopsReady, err := r.aggregateGitOpsProviders(ctx, platform)
	if err != nil {
		logger.Error(err, "Failed to aggregate gitops providers")
		return ctrl.Result{RequeueAfter: defaultRequeueTime}, nil
	}
	platform.Status.Providers.GitOpsProviders = gitopsStatuses
	if !gitopsReady {
		allReady = false
	}

	// Update observed generation
	platform.Status.ObservedGeneration = platform.Generation

	// Set condition and phase based on provider readiness
	if allReady && len(platform.Spec.Components.GitProviders) > 0 {
		platform.Status.Phase = "Ready"
		meta.SetStatusCondition(&platform.Status.Conditions, metav1.Condition{
			Type:    "Ready",
			Status:  metav1.ConditionTrue,
			Reason:  "AllComponentsReady",
			Message: "All platform components are operational",
		})
	} else {
		platform.Status.Phase = "Initializing"
		meta.SetStatusCondition(&platform.Status.Conditions, metav1.Condition{
			Type:    "Ready",
			Status:  metav1.ConditionFalse,
			Reason:  "ComponentsNotReady",
			Message: "Waiting for platform components to be ready",
		})
	}

	if err := r.Status().Update(ctx, platform); err != nil {
		return ctrl.Result{}, err
	}

	if !allReady {
		logger.Info("Platform not fully ready, requeuing")
		return ctrl.Result{RequeueAfter: defaultRequeueTime}, nil
	}

	// Create bootstrap repositories and ArgoCD Applications after all providers are Ready
	if err := r.createBootstrapResources(ctx, platform); err != nil {
		logger.Error(err, "Failed to create bootstrap resources")
		return ctrl.Result{RequeueAfter: defaultRequeueTime}, nil
	}

	logger.Info("Platform reconciliation complete", "phase", platform.Status.Phase)
	return ctrl.Result{}, nil
}

// aggregateGitProviders aggregates status from all Git providers
func (r *PlatformReconciler) aggregateGitProviders(ctx context.Context, platform *v1alpha2.Platform) ([]v1alpha2.ProviderStatusSummary, bool, error) {
	logger := log.FromContext(ctx)
	summaries := []v1alpha2.ProviderStatusSummary{}
	allReady := true

	for _, gitProviderRef := range platform.Spec.Components.GitProviders {
		// Fetch provider using unstructured client to support duck-typing
		gvk := schema.GroupVersionKind{
			Group:   "idpbuilder.cnoe.io",
			Version: "v1alpha2",
			Kind:    gitProviderRef.Kind,
		}

		providerObj := &unstructured.Unstructured{}
		providerObj.SetGroupVersionKind(gvk)

		err := r.Get(ctx, types.NamespacedName{
			Name:      gitProviderRef.Name,
			Namespace: gitProviderRef.Namespace,
		}, providerObj)

		if err != nil {
			if errors.IsNotFound(err) {
				logger.Info("Git provider not found", "name", gitProviderRef.Name, "kind", gitProviderRef.Kind)
				summaries = append(summaries, v1alpha2.ProviderStatusSummary{
					Name:  gitProviderRef.Name,
					Kind:  gitProviderRef.Kind,
					Ready: false,
				})
				allReady = false
				continue
			}
			return nil, false, fmt.Errorf("getting git provider %s: %w", gitProviderRef.Name, err)
		}

		// Extract status using duck-typing
		status, err := provider.GetGitProviderStatus(providerObj)
		if err != nil {
			logger.Error(err, "Failed to get git provider status", "name", gitProviderRef.Name)
			summaries = append(summaries, v1alpha2.ProviderStatusSummary{
				Name:  gitProviderRef.Name,
				Kind:  gitProviderRef.Kind,
				Ready: false,
			})
			allReady = false
			continue
		}

		summaries = append(summaries, v1alpha2.ProviderStatusSummary{
			Name:    gitProviderRef.Name,
			Kind:    gitProviderRef.Kind,
			Ready:   status.Ready,
			Message: status.Message,
			Reason:  status.Reason,
		})

		if !status.Ready {
			allReady = false
		}
	}

	return summaries, allReady, nil
}

// aggregateGateways aggregates status from all Gateway providers
func (r *PlatformReconciler) aggregateGateways(ctx context.Context, platform *v1alpha2.Platform) ([]v1alpha2.ProviderStatusSummary, bool, error) {
	logger := log.FromContext(ctx)
	summaries := []v1alpha2.ProviderStatusSummary{}
	allReady := true

	for _, gatewayRef := range platform.Spec.Components.Gateways {
		// Fetch provider using unstructured client to support duck-typing
		gvk := schema.GroupVersionKind{
			Group:   "idpbuilder.cnoe.io",
			Version: "v1alpha2",
			Kind:    gatewayRef.Kind,
		}

		providerObj := &unstructured.Unstructured{}
		providerObj.SetGroupVersionKind(gvk)

		err := r.Get(ctx, types.NamespacedName{
			Name:      gatewayRef.Name,
			Namespace: gatewayRef.Namespace,
		}, providerObj)

		if err != nil {
			if errors.IsNotFound(err) {
				logger.Info("Gateway provider not found", "name", gatewayRef.Name, "kind", gatewayRef.Kind)
				summaries = append(summaries, v1alpha2.ProviderStatusSummary{
					Name:  gatewayRef.Name,
					Kind:  gatewayRef.Kind,
					Ready: false,
				})
				allReady = false
				continue
			}
			return nil, false, fmt.Errorf("getting gateway provider %s: %w", gatewayRef.Name, err)
		}

		// Extract status using duck-typing
		status, err := provider.GetGatewayProviderStatus(providerObj)
		if err != nil {
			logger.Error(err, "Failed to get gateway provider status", "name", gatewayRef.Name)
			summaries = append(summaries, v1alpha2.ProviderStatusSummary{
				Name:  gatewayRef.Name,
				Kind:  gatewayRef.Kind,
				Ready: false,
			})
			allReady = false
			continue
		}

		summaries = append(summaries, v1alpha2.ProviderStatusSummary{
			Name:    gatewayRef.Name,
			Kind:    gatewayRef.Kind,
			Ready:   status.Ready,
			Message: status.Message,
			Reason:  status.Reason,
		})

		if !status.Ready {
			allReady = false
		}
	}

	return summaries, allReady, nil
}

// aggregateGitOpsProviders aggregates status from all GitOps providers
func (r *PlatformReconciler) aggregateGitOpsProviders(ctx context.Context, platform *v1alpha2.Platform) ([]v1alpha2.ProviderStatusSummary, bool, error) {
	logger := log.FromContext(ctx)
	summaries := []v1alpha2.ProviderStatusSummary{}
	allReady := true

	for _, gitopsProviderRef := range platform.Spec.Components.GitOpsProviders {
		// Fetch provider using unstructured client to support duck-typing
		gvk := schema.GroupVersionKind{
			Group:   "idpbuilder.cnoe.io",
			Version: "v1alpha2",
			Kind:    gitopsProviderRef.Kind,
		}

		providerObj := &unstructured.Unstructured{}
		providerObj.SetGroupVersionKind(gvk)

		err := r.Get(ctx, types.NamespacedName{
			Name:      gitopsProviderRef.Name,
			Namespace: gitopsProviderRef.Namespace,
		}, providerObj)

		if err != nil {
			if errors.IsNotFound(err) {
				logger.Info("GitOps provider not found", "name", gitopsProviderRef.Name, "kind", gitopsProviderRef.Kind)
				summaries = append(summaries, v1alpha2.ProviderStatusSummary{
					Name:  gitopsProviderRef.Name,
					Kind:  gitopsProviderRef.Kind,
					Ready: false,
				})
				allReady = false
				continue
			}
			return nil, false, fmt.Errorf("getting gitops provider %s: %w", gitopsProviderRef.Name, err)
		}

		// Extract status using duck-typing
		status, err := provider.GetGitOpsProviderStatus(providerObj)
		if err != nil {
			logger.Error(err, "Failed to get gitops provider status", "name", gitopsProviderRef.Name)
			summaries = append(summaries, v1alpha2.ProviderStatusSummary{
				Name:  gitopsProviderRef.Name,
				Kind:  gitopsProviderRef.Kind,
				Ready: false,
			})
			allReady = false
			continue
		}

		summaries = append(summaries, v1alpha2.ProviderStatusSummary{
			Name:    gitopsProviderRef.Name,
			Kind:    gitopsProviderRef.Kind,
			Ready:   status.Ready,
			Message: status.Message,
			Reason:  status.Reason,
		})

		if !status.Ready {
			allReady = false
		}
	}

	return summaries, allReady, nil
}

// ensureProviderOwnerReferences ensures that all provider CRs have the Platform as an owner reference
func (r *PlatformReconciler) ensureProviderOwnerReferences(ctx context.Context, platform *v1alpha2.Platform) error {
	// Process all Git providers
	for _, providerRef := range platform.Spec.Components.GitProviders {
		if err := r.ensureOwnerReference(ctx, platform, providerRef); err != nil {
			return err
		}
	}

	// Process all Gateway providers
	for _, providerRef := range platform.Spec.Components.Gateways {
		if err := r.ensureOwnerReference(ctx, platform, providerRef); err != nil {
			return err
		}
	}

	// Process all GitOps providers
	for _, providerRef := range platform.Spec.Components.GitOpsProviders {
		if err := r.ensureOwnerReference(ctx, platform, providerRef); err != nil {
			return err
		}
	}

	return nil
}

// ensureOwnerReference ensures that a specific provider CR has the Platform as an owner reference
func (r *PlatformReconciler) ensureOwnerReference(ctx context.Context, platform *v1alpha2.Platform, providerRef v1alpha2.ProviderReference) error {
	logger := log.FromContext(ctx)

	// Fetch provider as unstructured (works for any provider type)
	provider := &unstructured.Unstructured{}
	provider.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "idpbuilder.cnoe.io",
		Version: "v1alpha2",
		Kind:    providerRef.Kind,
	})

	key := types.NamespacedName{
		Name:      providerRef.Name,
		Namespace: providerRef.Namespace,
	}

	if err := r.Get(ctx, key, provider); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Provider not found, skipping owner reference", "name", providerRef.Name, "kind", providerRef.Kind)
			return nil
		}
		return fmt.Errorf("getting provider %s/%s: %w", providerRef.Kind, providerRef.Name, err)
	}

	// Check if Platform is already an owner
	hasOwnerRef := false
	for _, ref := range provider.GetOwnerReferences() {
		if ref.UID == platform.UID {
			hasOwnerRef = true
			break
		}
	}

	if !hasOwnerRef {
		// Add Platform as owner reference (non-controller owner)
		ownerRef := metav1.OwnerReference{
			APIVersion: platform.APIVersion,
			Kind:       platform.Kind,
			Name:       platform.Name,
			UID:        platform.UID,
			Controller: func() *bool { b := false; return &b }(), // Not a controller owner
		}

		refs := provider.GetOwnerReferences()
		refs = append(refs, ownerRef)
		provider.SetOwnerReferences(refs)

		if err := r.Update(ctx, provider); err != nil {
			return fmt.Errorf("updating provider %s/%s with owner reference: %w", providerRef.Kind, providerRef.Name, err)
		}

		logger.Info("Added Platform as owner reference", "provider", providerRef.Name, "kind", providerRef.Kind)
	}

	return nil
}

// handleDeletion handles the deletion of Platform
func (r *PlatformReconciler) handleDeletion(ctx context.Context, platform *v1alpha2.Platform) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	if controllerutil.ContainsFinalizer(platform, platformFinalizer) {
		// Perform cleanup if needed
		logger.Info("Cleaning up Platform resources")

		// Remove finalizer
		controllerutil.RemoveFinalizer(platform, platformFinalizer)
		if err := r.Update(ctx, platform); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// createBootstrapResources creates bootstrap GitRepository and ArgoCD Application CRs
func (r *PlatformReconciler) createBootstrapResources(ctx context.Context, platform *v1alpha2.Platform) error {
	logger := log.FromContext(ctx)

	// Check if we've already created bootstrap resources
	if platform.Annotations != nil {
		if _, ok := platform.Annotations[bootstrapReposCreatedFlag]; ok {
			logger.V(1).Info("Bootstrap resources already created")
			return nil
		}
	}

	// Extract build name from platform name (format: {buildname}-platform)
	buildName := strings.TrimSuffix(platform.Name, "-platform")
	if buildName == platform.Name {
		return fmt.Errorf("platform name does not follow expected format: {buildname}-platform")
	}

	// Get the first Git provider to use for creating repositories
	if len(platform.Spec.Components.GitProviders) == 0 {
		return fmt.Errorf("no git providers configured")
	}

	gitProviderRef := platform.Spec.Components.GitProviders[0]

	// Fetch the Git provider to get its status
	gitProvider := &unstructured.Unstructured{}
	gitProvider.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "idpbuilder.cnoe.io",
		Version: "v1alpha2",
		Kind:    gitProviderRef.Kind,
	})

	err := r.Get(ctx, types.NamespacedName{
		Name:      gitProviderRef.Name,
		Namespace: gitProviderRef.Namespace,
	}, gitProvider)
	if err != nil {
		return fmt.Errorf("getting git provider %s: %w", gitProviderRef.Name, err)
	}

	// Get provider status using duck-typing
	gitProviderStatus, err := provider.GetGitProviderStatus(gitProvider)
	if err != nil {
		return fmt.Errorf("getting git provider status: %w", err)
	}

	// Validate URLs
	if err := validateGitURL(gitProviderStatus.Endpoint, "Git provider endpoint"); err != nil {
		return err
	}
	if err := validateGitURL(gitProviderStatus.InternalEndpoint, "Git provider internal endpoint"); err != nil {
		return err
	}

	// Create bootstrap repositories for core components
	bootstrapApps := []string{v1alpha1.ArgoCDPackageName}

	for _, appName := range bootstrapApps {
		logger.V(1).Info("Creating bootstrap GitRepository and Application", "app", appName)

		// Create GitRepository CR
		repo, err := r.createGitRepository(ctx, platform, buildName, appName, gitProviderStatus)
		if err != nil {
			return fmt.Errorf("creating GitRepository for %s: %w", appName, err)
		}

		// Create ArgoCD Application CR
		if err := r.createArgoCDApplication(ctx, platform, appName, repo); err != nil {
			return fmt.Errorf("creating ArgoCD Application for %s: %w", appName, err)
		}
	}

	// Mark bootstrap resources as created
	if platform.Annotations == nil {
		platform.Annotations = make(map[string]string)
	}
	platform.Annotations[bootstrapReposCreatedFlag] = "true"
	if err := r.Update(ctx, platform); err != nil {
		return fmt.Errorf("updating platform annotation: %w", err)
	}

	logger.Info("Bootstrap resources created successfully")
	return nil
}

// createGitRepository creates a GitRepository CR for a bootstrap app
func (r *PlatformReconciler) createGitRepository(ctx context.Context, platform *v1alpha2.Platform, buildName, appName string, gitProviderStatus *provider.GitProviderStatus) (*v1alpha1.GitRepository, error) {
	logger := log.FromContext(ctx)

	repo := &v1alpha1.GitRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      appName,
			Namespace: globals.GetProjectNamespace(buildName),
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, repo, func() error {
		// Set Platform as owner (non-controller)
		ownerRef := metav1.OwnerReference{
			APIVersion: platform.APIVersion,
			Kind:       platform.Kind,
			Name:       platform.Name,
			UID:        platform.UID,
			Controller: func() *bool { b := false; return &b }(),
		}

		// Check if owner reference already exists
		hasOwnerRef := false
		for _, ref := range repo.GetOwnerReferences() {
			if ref.UID == platform.UID {
				hasOwnerRef = true
				break
			}
		}

		if !hasOwnerRef {
			refs := repo.GetOwnerReferences()
			refs = append(refs, ownerRef)
			repo.SetOwnerReferences(refs)
		}

		// Set up labels
		util.SetPackageLabels(repo)

		// Get credentials secret ref from provider status
		var secretName, secretNamespace string
		if gitProviderStatus.CredentialsSecretRef.Name != "" {
			secretName = gitProviderStatus.CredentialsSecretRef.Name
			secretNamespace = gitProviderStatus.CredentialsSecretRef.Namespace
		} else {
			logger.V(1).Info("Warning: Git provider credentials secret ref is not set")
			// Fallback to defaults for Gitea
			secretName = util.GiteaAdminSecret
			secretNamespace = util.GiteaNamespace
		}

		repo.Spec = v1alpha1.GitRepositorySpec{
			Source: v1alpha1.GitRepositorySource{
				Type:            v1alpha1.SourceTypeEmbedded,
				EmbeddedAppName: appName,
			},
			Provider: v1alpha1.Provider{
				Name:             v1alpha1.GitProviderGitea,
				GitURL:           gitProviderStatus.Endpoint,
				InternalGitURL:   gitProviderStatus.InternalEndpoint,
				OrganizationName: v1alpha1.GiteaAdminUserName,
			},
			SecretRef: v1alpha1.SecretReference{
				Name:      secretName,
				Namespace: secretNamespace,
			},
		}

		return nil
	})

	return repo, err
}

// createArgoCDApplication creates an ArgoCD Application CR for a bootstrap app
func (r *PlatformReconciler) createArgoCDApplication(ctx context.Context, platform *v1alpha2.Platform, appName string, repo *v1alpha1.GitRepository) error {
	app := &argov1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name:      appName,
			Namespace: globals.ArgoCDNamespace,
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, app, func() error {
		// Set Platform as owner (non-controller)
		ownerRef := metav1.OwnerReference{
			APIVersion: platform.APIVersion,
			Kind:       platform.Kind,
			Name:       platform.Name,
			UID:        platform.UID,
			Controller: func() *bool { b := false; return &b }(),
		}

		// Check if owner reference already exists
		hasOwnerRef := false
		for _, ref := range app.GetOwnerReferences() {
			if ref.UID == platform.UID {
				hasOwnerRef = true
				break
			}
		}

		if !hasOwnerRef {
			refs := app.GetOwnerReferences()
			refs = append(refs, ownerRef)
			app.SetOwnerReferences(refs)
		}

		// Set up labels
		util.SetPackageLabels(app)

		// Set Application spec
		localbuild.SetApplicationSpec(
			app,
			repo.Status.InternalGitRepositoryUrl,
			".",
			defaultArgoCDProjectName,
			appName,
			nil,
		)

		return nil
	})

	return err
}

// validateGitURL validates that a Git URL is properly formatted
func validateGitURL(url, fieldName string) error {
	if url == "" {
		return fmt.Errorf("%s is not set", fieldName)
	}
	// Validate URL format - must match the GitRepository CRD validation pattern: ^https?:\/\/.+$
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return fmt.Errorf("%s must start with http:// or https://, got: %s", fieldName, url)
	}
	// Check that there's content after the protocol (e.g., not just "http://" or "https://")
	if url == "http://" || url == "https://" {
		return fmt.Errorf("%s is too short: %s", fieldName, url)
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PlatformReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha2.Platform{}).
		Complete(r)
}
