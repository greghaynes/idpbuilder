package platform

import (
	"context"
	"fmt"

	"github.com/cnoe-io/idpbuilder/api/v1alpha2"
	"github.com/cnoe-io/idpbuilder/pkg/util/provider"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// PlatformReconciler reconciles a Platform object
type PlatformReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=idpbuilder.cnoe.io,resources=platforms,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=idpbuilder.cnoe.io,resources=platforms/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=idpbuilder.cnoe.io,resources=platforms/finalizers,verbs=update
// +kubebuilder:rbac:groups=idpbuilder.cnoe.io,resources=giteaproviders,verbs=get;list;watch
// +kubebuilder:rbac:groups=idpbuilder.cnoe.io,resources=nginxgateways,verbs=get;list;watch
// +kubebuilder:rbac:groups=idpbuilder.cnoe.io,resources=argocdproviders,verbs=get;list;watch

func (r *PlatformReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling Platform", "name", req.Name, "namespace", req.Namespace)

	// Fetch the Platform instance
	platform := &v1alpha2.Platform{}
	if err := r.Get(ctx, req.NamespacedName, platform); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("Platform resource not found, ignoring")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get Platform")
		return ctrl.Result{}, err
	}

	// Set initial status if not set
	if platform.Status.Phase == "" {
		platform.Status.Phase = "Pending"
		if err := r.Status().Update(ctx, platform); err != nil {
			logger.Error(err, "Failed to update Platform status")
			return ctrl.Result{}, err
		}
	}

	// Aggregate provider status
	allReady := true

	// Check Git providers
	gitProviderStatuses := []v1alpha2.ProviderStatusSummary{}
	for _, providerRef := range platform.Spec.Components.GitProviders {
		ready, err := r.checkProviderReady(ctx, providerRef, "git")
		if err != nil {
			logger.Error(err, "Failed to check git provider", "provider", providerRef.Name)
			allReady = false
		} else {
			allReady = allReady && ready
		}
		gitProviderStatuses = append(gitProviderStatuses, v1alpha2.ProviderStatusSummary{
			Name:  providerRef.Name,
			Kind:  providerRef.Kind,
			Ready: ready,
		})
	}

	// Check Gateways
	gatewayStatuses := []v1alpha2.ProviderStatusSummary{}
	for _, providerRef := range platform.Spec.Components.Gateways {
		ready, err := r.checkProviderReady(ctx, providerRef, "gateway")
		if err != nil {
			logger.Error(err, "Failed to check gateway provider", "provider", providerRef.Name)
			allReady = false
		} else {
			allReady = allReady && ready
		}
		gatewayStatuses = append(gatewayStatuses, v1alpha2.ProviderStatusSummary{
			Name:  providerRef.Name,
			Kind:  providerRef.Kind,
			Ready: ready,
		})
	}

	// Check GitOps providers
	gitopsProviderStatuses := []v1alpha2.ProviderStatusSummary{}
	for _, providerRef := range platform.Spec.Components.GitOpsProviders {
		ready, err := r.checkProviderReady(ctx, providerRef, "gitops")
		if err != nil {
			logger.Error(err, "Failed to check gitops provider", "provider", providerRef.Name)
			allReady = false
		} else {
			allReady = allReady && ready
		}
		gitopsProviderStatuses = append(gitopsProviderStatuses, v1alpha2.ProviderStatusSummary{
			Name:  providerRef.Name,
			Kind:  providerRef.Kind,
			Ready: ready,
		})
	}

	// Update provider status
	platform.Status.Providers.GitProviders = gitProviderStatuses
	platform.Status.Providers.Gateways = gatewayStatuses
	platform.Status.Providers.GitOpsProviders = gitopsProviderStatuses

	// Update overall status
	if allReady {
		platform.Status.Phase = "Ready"
		r.setCondition(platform, metav1.Condition{
			Type:    "Ready",
			Status:  metav1.ConditionTrue,
			Reason:  "AllComponentsReady",
			Message: "All platform components are operational",
		})
	} else {
		platform.Status.Phase = "Pending"
		r.setCondition(platform, metav1.Condition{
			Type:    "Ready",
			Status:  metav1.ConditionFalse,
			Reason:  "ComponentsNotReady",
			Message: "Some platform components are not ready yet",
		})
	}

	platform.Status.ObservedGeneration = platform.GetGeneration()

	if err := r.Status().Update(ctx, platform); err != nil {
		logger.Error(err, "Failed to update Platform status")
		return ctrl.Result{}, err
	}

	logger.Info("Successfully reconciled Platform", "phase", platform.Status.Phase)
	return ctrl.Result{}, nil
}

func (r *PlatformReconciler) checkProviderReady(ctx context.Context, providerRef v1alpha2.ProviderReference, providerType string) (bool, error) {
	logger := log.FromContext(ctx)

	// Create unstructured object to fetch the provider
	obj := &unstructured.Unstructured{}
	gvk := schema.GroupVersionKind{
		Group:   "idpbuilder.cnoe.io",
		Version: "v1alpha2",
		Kind:    providerRef.Kind,
	}
	obj.SetGroupVersionKind(gvk)

	// Fetch the provider
	key := client.ObjectKey{
		Name:      providerRef.Name,
		Namespace: providerRef.Namespace,
	}
	if err := r.Get(ctx, key, obj); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("Provider not found", "kind", providerRef.Kind, "name", providerRef.Name)
			return false, nil
		}
		return false, err
	}

	// Use duck-typing to check if provider is ready
	var ready bool
	var err error

	switch providerType {
	case "git":
		ready, err = provider.IsGitProviderReady(obj)
	case "gateway":
		ready, err = provider.IsGatewayProviderReady(obj)
	case "gitops":
		ready, err = provider.IsGitOpsProviderReady(obj)
	default:
		return false, fmt.Errorf("unknown provider type: %s", providerType)
	}

	if err != nil {
		logger.Error(err, "Failed to check provider readiness", "kind", providerRef.Kind, "name", providerRef.Name)
		return false, err
	}

	return ready, nil
}

func (r *PlatformReconciler) setCondition(platform *v1alpha2.Platform, condition metav1.Condition) {
	condition.LastTransitionTime = metav1.Now()
	meta.SetStatusCondition(&platform.Status.Conditions, condition)
}

// SetupWithManager sets up the controller with the Manager.
func (r *PlatformReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha2.Platform{}).
		Complete(r)
}
