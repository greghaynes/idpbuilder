package gatewayprovider

import (
	"context"
	"embed"
	"fmt"

	"github.com/cnoe-io/idpbuilder/api/v1alpha2"
	"github.com/cnoe-io/idpbuilder/globals"
	"github.com/cnoe-io/idpbuilder/pkg/k8s"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

//go:embed resources/nginx/k8s/*
var installNginxFS embed.FS

// NginxGatewayReconciler reconciles a NginxGateway object
type NginxGatewayReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=idpbuilder.cnoe.io,resources=nginxgateways,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=idpbuilder.cnoe.io,resources=nginxgateways/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=idpbuilder.cnoe.io,resources=nginxgateways/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch;create;update;patch
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch
// +kubebuilder:rbac:groups=networking.k8s.io,resources=ingressclasses,verbs=get;list;watch;create;update;patch

func (r *NginxGatewayReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling NginxGateway", "name", req.Name, "namespace", req.Namespace)

	// Fetch the NginxGateway instance
	nginxGateway := &v1alpha2.NginxGateway{}
	if err := r.Get(ctx, req.NamespacedName, nginxGateway); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("NginxGateway resource not found, ignoring")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get NginxGateway")
		return ctrl.Result{}, err
	}

	// Set initial status if not set
	if nginxGateway.Status.Phase == "" {
		nginxGateway.Status.Phase = "Pending"
		if err := r.Status().Update(ctx, nginxGateway); err != nil {
			logger.Error(err, "Failed to update NginxGateway status")
			return ctrl.Result{}, err
		}
	}

	// Install Nginx
	if err := r.installNginx(ctx, nginxGateway); err != nil {
		logger.Error(err, "Failed to install Nginx")
		r.setCondition(nginxGateway, metav1.Condition{
			Type:    "Ready",
			Status:  metav1.ConditionFalse,
			Reason:  "InstallationFailed",
			Message: fmt.Sprintf("Failed to install Nginx: %v", err),
		})
		nginxGateway.Status.Phase = "Failed"
		if statusErr := r.Status().Update(ctx, nginxGateway); statusErr != nil {
			logger.Error(statusErr, "Failed to update status after installation failure")
		}
		return ctrl.Result{}, err
	}

	// Update status with duck-typed fields
	if err := r.updateStatus(ctx, nginxGateway); err != nil {
		logger.Error(err, "Failed to update NginxGateway status")
		return ctrl.Result{}, err
	}

	logger.Info("Successfully reconciled NginxGateway")
	return ctrl.Result{}, nil
}

func (r *NginxGatewayReconciler) installNginx(ctx context.Context, nginxGateway *v1alpha2.NginxGateway) error {
	logger := log.FromContext(ctx)
	logger.Info("Installing Nginx", "namespace", nginxGateway.Spec.Namespace)

	// Ensure namespace exists
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: nginxGateway.Spec.Namespace,
		},
	}
	if err := k8s.EnsureNamespace(ctx, r.Client, nginxGateway.Spec.Namespace); err != nil {
		return fmt.Errorf("failed to ensure namespace: %w", err)
	}

	// Load and apply embedded manifests
	installObjs, err := k8s.BuildCustomizedObjects("", "resources/nginx/k8s", installNginxFS, r.Scheme, nil)
	if err != nil {
		return fmt.Errorf("failed to build nginx manifests: %w", err)
	}

	nsClient := client.NewNamespacedClient(r.Client, nginxGateway.Spec.Namespace)
	for _, obj := range installObjs {
		if err := k8s.EnsureObject(ctx, nsClient, obj, nginxGateway.Spec.Namespace); err != nil {
			return fmt.Errorf("failed to create nginx resource %s: %w", obj.GetName(), err)
		}
	}

	logger.Info("Nginx resources created successfully")
	return nil
}

func (r *NginxGatewayReconciler) updateStatus(ctx context.Context, nginxGateway *v1alpha2.NginxGateway) error {
	logger := log.FromContext(ctx)

	// Get the ingress controller service to determine the load balancer endpoint
	svc := &corev1.Service{}
	err := r.Get(ctx, types.NamespacedName{
		Name:      "ingress-nginx-controller",
		Namespace: nginxGateway.Spec.Namespace,
	}, svc)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("Nginx controller service not found yet, will retry")
			return nil
		}
		return fmt.Errorf("failed to get nginx controller service: %w", err)
	}

	// Update duck-typed status fields
	nginxGateway.Status.Installed = true
	nginxGateway.Status.Phase = "Ready"
	nginxGateway.Status.Version = nginxGateway.Spec.Version
	
	// Set ingress class name
	ingressClassName := nginxGateway.Spec.IngressClass.Name
	if ingressClassName == "" {
		ingressClassName = "nginx"
	}
	nginxGateway.Status.IngressClassName = ingressClassName

	// Set internal endpoint
	nginxGateway.Status.InternalEndpoint = fmt.Sprintf("http://%s.%s.svc.cluster.local", 
		svc.Name, svc.Namespace)

	// Try to get load balancer endpoint
	if len(svc.Status.LoadBalancer.Ingress) > 0 {
		if svc.Status.LoadBalancer.Ingress[0].IP != "" {
			nginxGateway.Status.LoadBalancerEndpoint = fmt.Sprintf("http://%s", svc.Status.LoadBalancer.Ingress[0].IP)
		} else if svc.Status.LoadBalancer.Ingress[0].Hostname != "" {
			nginxGateway.Status.LoadBalancerEndpoint = fmt.Sprintf("http://%s", svc.Status.LoadBalancer.Ingress[0].Hostname)
		}
	} else if svc.Spec.Type == corev1.ServiceTypeNodePort {
		// For NodePort, we can try to get the node IP
		nginxGateway.Status.LoadBalancerEndpoint = "http://localhost" // Placeholder for now
	}

	// Set Ready condition
	r.setCondition(nginxGateway, metav1.Condition{
		Type:    "Ready",
		Status:  metav1.ConditionTrue,
		Reason:  "NginxInstalled",
		Message: "Nginx Ingress Controller is installed and ready",
	})

	if err := r.Status().Update(ctx, nginxGateway); err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	logger.Info("Updated NginxGateway status", "phase", nginxGateway.Status.Phase)
	return nil
}

func (r *NginxGatewayReconciler) setCondition(nginxGateway *v1alpha2.NginxGateway, condition metav1.Condition) {
	condition.LastTransitionTime = metav1.Now()
	meta.SetStatusCondition(&nginxGateway.Status.Conditions, condition)
}

// SetupWithManager sets up the controller with the Manager.
func (r *NginxGatewayReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha2.NginxGateway{}).
		Complete(r)
}
