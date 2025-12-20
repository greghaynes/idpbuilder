package gitopsprovider

import (
	"context"
	"embed"
	"fmt"
	"time"

	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"github.com/cnoe-io/idpbuilder/api/v1alpha2"
	"github.com/cnoe-io/idpbuilder/pkg/k8s"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

//go:embed resources/argo/*
var argoResourcesFS embed.FS

const (
	argoResourcePath = "resources/argo"
)

// ArgoCDProviderReconciler reconciles an ArgoCDProvider object
type ArgoCDProviderReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Config v1alpha1.BuildCustomizationSpec
}

// +kubebuilder:rbac:groups=idpbuilder.cnoe.io,resources=argocdproviders,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=idpbuilder.cnoe.io,resources=argocdproviders/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=idpbuilder.cnoe.io,resources=argocdproviders/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch;create;update;patch
// +kubebuilder:rbac:groups="apps",resources=deployments;statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services;configmaps;secrets,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop
func (r *ArgoCDProviderReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling ArgoCDProvider", "name", req.Name, "namespace", req.Namespace)

	// Fetch the ArgoCDProvider instance
	argocd := &v1alpha2.ArgoCDProvider{}
	err := r.Get(ctx, req.NamespacedName, argocd)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("ArgoCDProvider resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get ArgoCDProvider")
		return ctrl.Result{}, err
	}

	// Set initial status if not set
	if argocd.Status.Phase == "" {
		argocd.Status.Phase = "Pending"
		if err := r.Status().Update(ctx, argocd); err != nil {
			logger.Error(err, "Failed to update ArgoCDProvider status")
			return ctrl.Result{}, err
		}
	}

	// Install ArgoCD
	if err := r.installArgoCD(ctx, argocd); err != nil {
		logger.Error(err, "Failed to install ArgoCD")
		r.setCondition(argocd, "Ready", metav1.ConditionFalse, "InstallationFailed", err.Error())
		argocd.Status.Phase = "Failed"
		if statusErr := r.Status().Update(ctx, argocd); statusErr != nil {
			logger.Error(statusErr, "Failed to update status")
		}
		return ctrl.Result{RequeueAfter: 30 * time.Second}, err
	}

	// Wait for ArgoCD to be ready
	if err := r.waitForArgoCDReady(ctx, argocd); err != nil {
		logger.Info("ArgoCD is not ready yet", "error", err.Error())
		argocd.Status.Phase = "Installing"
		if statusErr := r.Status().Update(ctx, argocd); statusErr != nil {
			logger.Error(statusErr, "Failed to update status")
		}
		return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
	}

	// Update status with ArgoCD information
	if err := r.updateArgoCDStatus(ctx, argocd); err != nil {
		logger.Error(err, "Failed to update ArgoCD status")
		return ctrl.Result{RequeueAfter: 10 * time.Second}, err
	}

	// Set Ready condition and phase
	r.setCondition(argocd, "Ready", metav1.ConditionTrue, "ArgoCDReady", "ArgoCD is ready")
	argocd.Status.Phase = "Ready"
	argocd.Status.Installed = true
	argocd.Status.Version = argocd.Spec.Version

	if err := r.Status().Update(ctx, argocd); err != nil {
		logger.Error(err, "Failed to update ArgoCDProvider status")
		return ctrl.Result{}, err
	}

	logger.Info("ArgoCDProvider reconciled successfully")
	return ctrl.Result{}, nil
}

func (r *ArgoCDProviderReconciler) installArgoCD(ctx context.Context, argocd *v1alpha2.ArgoCDProvider) error {
	logger := log.FromContext(ctx)
	logger.Info("Installing ArgoCD", "namespace", argocd.Spec.Namespace)

	// Ensure namespace exists
	if err := k8s.EnsureNamespace(ctx, r.Client, argocd.Spec.Namespace); err != nil {
		return fmt.Errorf("failed to ensure namespace: %w", err)
	}

	// Load and apply ArgoCD manifests
	// Use empty customization since we're using embedded manifests as-is
	customization := v1alpha1.PackageCustomization{}

	installObjs, err := k8s.BuildCustomizedObjects(customization.FilePath, argoResourcePath, argoResourcesFS, r.Scheme, r.Config)
	if err != nil {
		return fmt.Errorf("failed to build ArgoCD objects: %w", err)
	}

	nsClient := client.NewNamespacedClient(r.Client, argocd.Spec.Namespace)
	for _, obj := range installObjs {
		if err := k8s.EnsureObject(ctx, nsClient, obj, argocd.Spec.Namespace); err != nil {
			return fmt.Errorf("failed to ensure ArgoCD object %s: %w", obj.GetName(), err)
		}
	}

	logger.Info("ArgoCD manifests applied successfully")
	return nil
}

func (r *ArgoCDProviderReconciler) waitForArgoCDReady(ctx context.Context, argocd *v1alpha2.ArgoCDProvider) error {
	logger := log.FromContext(ctx)

	// Check if ArgoCD server deployment is ready
	serverDeployment := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{
		Name:      "argocd-server",
		Namespace: argocd.Spec.Namespace,
	}, serverDeployment)
	if err != nil {
		return fmt.Errorf("argocd-server deployment not found: %w", err)
	}

	if serverDeployment.Status.AvailableReplicas < 1 {
		logger.Info("Waiting for argocd-server deployment to be ready",
			"available", serverDeployment.Status.AvailableReplicas,
			"desired", serverDeployment.Status.Replicas)
		return fmt.Errorf("argocd-server not ready: %d/%d replicas available",
			serverDeployment.Status.AvailableReplicas, serverDeployment.Status.Replicas)
	}

	// Check if ArgoCD repo server deployment is ready
	repoDeployment := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{
		Name:      "argocd-repo-server",
		Namespace: argocd.Spec.Namespace,
	}, repoDeployment)
	if err != nil {
		return fmt.Errorf("argocd-repo-server deployment not found: %w", err)
	}

	if repoDeployment.Status.AvailableReplicas < 1 {
		logger.Info("Waiting for argocd-repo-server deployment to be ready",
			"available", repoDeployment.Status.AvailableReplicas,
			"desired", repoDeployment.Status.Replicas)
		return fmt.Errorf("argocd-repo-server not ready: %d/%d replicas available",
			repoDeployment.Status.AvailableReplicas, repoDeployment.Status.Replicas)
	}

	// Check if ArgoCD application controller statefulset is ready
	appController := &appsv1.StatefulSet{}
	err = r.Get(ctx, types.NamespacedName{
		Name:      "argocd-application-controller",
		Namespace: argocd.Spec.Namespace,
	}, appController)
	if err != nil {
		return fmt.Errorf("argocd-application-controller statefulset not found: %w", err)
	}

	if appController.Status.ReadyReplicas < 1 {
		logger.Info("Waiting for argocd-application-controller statefulset to be ready",
			"ready", appController.Status.ReadyReplicas,
			"desired", appController.Status.Replicas)
		return fmt.Errorf("argocd-application-controller not ready: %d/%d replicas ready",
			appController.Status.ReadyReplicas, appController.Status.Replicas)
	}

	logger.Info("ArgoCD components are ready")
	return nil
}

func (r *ArgoCDProviderReconciler) updateArgoCDStatus(ctx context.Context, argocd *v1alpha2.ArgoCDProvider) error {
	logger := log.FromContext(ctx)

	// Get server deployment status
	serverDeployment := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{
		Name:      "argocd-server",
		Namespace: argocd.Spec.Namespace,
	}, serverDeployment)
	if err != nil {
		return fmt.Errorf("failed to get argocd-server deployment: %w", err)
	}

	// Set server health status
	if serverDeployment.Status.AvailableReplicas >= 1 {
		argocd.Status.ServerHealth.Status = "Healthy"
	} else {
		argocd.Status.ServerHealth.Status = "Degraded"
	}

	// Check application controller
	appController := &appsv1.StatefulSet{}
	err = r.Get(ctx, types.NamespacedName{
		Name:      "argocd-application-controller",
		Namespace: argocd.Spec.Namespace,
	}, appController)
	if err == nil && appController.Status.ReadyReplicas >= 1 {
		argocd.Status.ApplicationControllerReady = true
	}

	// Set endpoints
	// TODO: Get actual domain from cluster configuration
	domain := "cnoe.localtest.me"
	argocd.Status.Endpoint = fmt.Sprintf("https://argocd.%s", domain)
	argocd.Status.InternalEndpoint = fmt.Sprintf("http://argocd-server.%s.svc.cluster.local", argocd.Spec.Namespace)

	// Set credentials secret ref
	// ArgoCD stores admin password in a secret
	argocd.Status.CredentialsSecretRef = &v1alpha2.SecretReference{
		Name:      "argocd-initial-admin-secret",
		Namespace: argocd.Spec.Namespace,
		Key:       "password",
	}

	logger.Info("Updated ArgoCD status",
		"endpoint", argocd.Status.Endpoint,
		"serverHealth", argocd.Status.ServerHealth.Status)

	return nil
}

func (r *ArgoCDProviderReconciler) setCondition(argocd *v1alpha2.ArgoCDProvider, conditionType string, status metav1.ConditionStatus, reason, message string) {
	condition := metav1.Condition{
		Type:               conditionType,
		Status:             status,
		ObservedGeneration: argocd.Generation,
		LastTransitionTime: metav1.Now(),
		Reason:             reason,
		Message:            message,
	}

	// Find and update existing condition or append new one
	found := false
	for i, existingCond := range argocd.Status.Conditions {
		if existingCond.Type == conditionType {
			// Only update if status has changed
			if existingCond.Status != status {
				argocd.Status.Conditions[i] = condition
			}
			found = true
			break
		}
	}
	if !found {
		argocd.Status.Conditions = append(argocd.Status.Conditions, condition)
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *ArgoCDProviderReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha2.ArgoCDProvider{}).
		Complete(r)
}
