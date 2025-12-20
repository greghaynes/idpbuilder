package gatewayprovider

import (
	"context"
	"embed"
	"fmt"
	"time"

	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"github.com/cnoe-io/idpbuilder/api/v1alpha2"
	"github.com/cnoe-io/idpbuilder/pkg/k8s"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

//go:embed resources/nginx/k8s/*
var nginxResourcesFS embed.FS

const (
	nginxResourcePath = "resources/nginx/k8s"
)

// NginxGatewayReconciler reconciles a NginxGateway object
type NginxGatewayReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Config v1alpha1.BuildCustomizationSpec
}

// +kubebuilder:rbac:groups=idpbuilder.cnoe.io,resources=nginxgateways,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=idpbuilder.cnoe.io,resources=nginxgateways/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=idpbuilder.cnoe.io,resources=nginxgateways/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch;create;update;patch
// +kubebuilder:rbac:groups="apps",resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="networking.k8s.io",resources=ingressclasses,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop
func (r *NginxGatewayReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling NginxGateway", "name", req.Name, "namespace", req.Namespace)

	// Fetch the NginxGateway instance
	nginx := &v1alpha2.NginxGateway{}
	err := r.Get(ctx, req.NamespacedName, nginx)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("NginxGateway resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get NginxGateway")
		return ctrl.Result{}, err
	}

	// Set initial status if not set
	if nginx.Status.Phase == "" {
		nginx.Status.Phase = "Pending"
		if err := r.Status().Update(ctx, nginx); err != nil {
			logger.Error(err, "Failed to update NginxGateway status")
			return ctrl.Result{}, err
		}
	}

	// Install Nginx
	if err := r.installNginx(ctx, nginx); err != nil {
		logger.Error(err, "Failed to install Nginx")
		r.setCondition(nginx, "Ready", metav1.ConditionFalse, "InstallationFailed", err.Error())
		nginx.Status.Phase = "Failed"
		if statusErr := r.Status().Update(ctx, nginx); statusErr != nil {
			logger.Error(statusErr, "Failed to update status")
		}
		return ctrl.Result{RequeueAfter: 30 * time.Second}, err
	}

	// Wait for Nginx to be ready
	if err := r.waitForNginxReady(ctx, nginx); err != nil {
		logger.Error(err, "Nginx is not ready yet")
		nginx.Status.Phase = "Installing"
		if statusErr := r.Status().Update(ctx, nginx); statusErr != nil {
			logger.Error(statusErr, "Failed to update status")
		}
		return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
	}

	// Update status with nginx information
	if err := r.updateNginxStatus(ctx, nginx); err != nil {
		logger.Error(err, "Failed to update Nginx status")
		return ctrl.Result{RequeueAfter: 10 * time.Second}, err
	}

	// Set Ready condition and phase
	r.setCondition(nginx, "Ready", metav1.ConditionTrue, "NginxReady", "Nginx Ingress Controller is ready")
	nginx.Status.Phase = "Ready"
	nginx.Status.Installed = true
	nginx.Status.Version = nginx.Spec.Version

	if err := r.Status().Update(ctx, nginx); err != nil {
		logger.Error(err, "Failed to update NginxGateway status")
		return ctrl.Result{}, err
	}

	logger.Info("NginxGateway reconciled successfully")
	return ctrl.Result{}, nil
}

func (r *NginxGatewayReconciler) installNginx(ctx context.Context, nginx *v1alpha2.NginxGateway) error {
	logger := log.FromContext(ctx)
	logger.Info("Installing Nginx", "namespace", nginx.Spec.Namespace)

	// Ensure namespace exists
	if err := k8s.EnsureNamespace(ctx, r.Client, nginx.Spec.Namespace); err != nil {
		return fmt.Errorf("failed to ensure namespace: %w", err)
	}

	// Load and apply nginx manifests
	// Use empty customization since we're using embedded manifests as-is
	customization := v1alpha1.PackageCustomization{}

	installObjs, err := k8s.BuildCustomizedObjects(customization.FilePath, nginxResourcePath, nginxResourcesFS, r.Scheme, r.Config)
	if err != nil {
		return fmt.Errorf("failed to build nginx objects: %w", err)
	}

	nsClient := client.NewNamespacedClient(r.Client, nginx.Spec.Namespace)
	for _, obj := range installObjs {
		if err := k8s.EnsureObject(ctx, nsClient, obj, nginx.Spec.Namespace); err != nil {
			return fmt.Errorf("failed to ensure nginx object %s: %w", obj.GetName(), err)
		}
	}

	logger.Info("Nginx manifests applied successfully")
	return nil
}

func (r *NginxGatewayReconciler) waitForNginxReady(ctx context.Context, nginx *v1alpha2.NginxGateway) error {
	logger := log.FromContext(ctx)

	// Check if nginx controller deployment is ready
	deployment := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{
		Name:      "ingress-nginx-controller",
		Namespace: nginx.Spec.Namespace,
	}, deployment)
	if err != nil {
		return fmt.Errorf("nginx controller deployment not found: %w", err)
	}

	if deployment.Status.AvailableReplicas < 1 {
		logger.Info("Waiting for nginx controller deployment to be ready",
			"available", deployment.Status.AvailableReplicas,
			"desired", deployment.Status.Replicas)
		return fmt.Errorf("nginx controller not ready: %d/%d replicas available",
			deployment.Status.AvailableReplicas, deployment.Status.Replicas)
	}

	logger.Info("Nginx controller deployment is ready")
	return nil
}

func (r *NginxGatewayReconciler) updateNginxStatus(ctx context.Context, nginx *v1alpha2.NginxGateway) error {
	logger := log.FromContext(ctx)

	// Get deployment status
	deployment := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{
		Name:      "ingress-nginx-controller",
		Namespace: nginx.Spec.Namespace,
	}, deployment)
	if err != nil {
		return fmt.Errorf("failed to get nginx controller deployment: %w", err)
	}

	nginx.Status.Controller.Replicas = deployment.Status.Replicas
	nginx.Status.Controller.ReadyReplicas = deployment.Status.ReadyReplicas

	// Set ingress class name from spec or default
	ingressClassName := nginx.Spec.IngressClass.Name
	if ingressClassName == "" {
		ingressClassName = "nginx"
	}
	nginx.Status.IngressClassName = ingressClassName

	// Get service to determine load balancer endpoint
	service := &corev1.Service{}
	err = r.Get(ctx, types.NamespacedName{
		Name:      "ingress-nginx-controller",
		Namespace: nginx.Spec.Namespace,
	}, service)
	if err != nil {
		logger.Info("Warning: could not get nginx controller service", "error", err.Error())
	} else {
		// For NodePort services, we'll use the node port information
		if service.Spec.Type == corev1.ServiceTypeNodePort {
			// Get node IP for load balancer endpoint
			nodeList := &corev1.NodeList{}
			if err := r.List(ctx, nodeList); err == nil && len(nodeList.Items) > 0 {
				// Use the first node's internal IP
				for _, addr := range nodeList.Items[0].Status.Addresses {
					if addr.Type == corev1.NodeInternalIP {
						nginx.Status.LoadBalancerEndpoint = fmt.Sprintf("http://%s", addr.Address)
						break
					}
				}
			}
		} else if service.Spec.Type == corev1.ServiceTypeLoadBalancer {
			// For LoadBalancer type
			if len(service.Status.LoadBalancer.Ingress) > 0 {
				ing := service.Status.LoadBalancer.Ingress[0]
				if ing.IP != "" {
					nginx.Status.LoadBalancerEndpoint = fmt.Sprintf("http://%s", ing.IP)
				} else if ing.Hostname != "" {
					nginx.Status.LoadBalancerEndpoint = fmt.Sprintf("http://%s", ing.Hostname)
				}
			}
		}
	}

	// Set internal endpoint
	nginx.Status.InternalEndpoint = fmt.Sprintf("http://ingress-nginx-controller.%s.svc.cluster.local", nginx.Spec.Namespace)

	logger.Info("Updated nginx status",
		"ingressClassName", nginx.Status.IngressClassName,
		"loadBalancerEndpoint", nginx.Status.LoadBalancerEndpoint)

	return nil
}

func (r *NginxGatewayReconciler) setCondition(nginx *v1alpha2.NginxGateway, conditionType string, status metav1.ConditionStatus, reason, message string) {
	condition := metav1.Condition{
		Type:               conditionType,
		Status:             status,
		ObservedGeneration: nginx.Generation,
		LastTransitionTime: metav1.Now(),
		Reason:             reason,
		Message:            message,
	}

	// Find and update existing condition or append new one
	found := false
	for i, existingCond := range nginx.Status.Conditions {
		if existingCond.Type == conditionType {
			// Only update if status has changed
			if existingCond.Status != status {
				nginx.Status.Conditions[i] = condition
			}
			found = true
			break
		}
	}
	if !found {
		nginx.Status.Conditions = append(nginx.Status.Conditions, condition)
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *NginxGatewayReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha2.NginxGateway{}).
		Complete(r)
}
