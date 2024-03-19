package clusterrolebindingctrl

import (
	"context"
	"strings"

	openfgasdk "github.com/openfga/go-sdk/client"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type ClusterRoleBindingReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	OpenFGAClient openfgasdk.SdkClient
}

// Reconcile implements reconcile.Reconciler.
func (r *ClusterRoleBindingReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	logger := log.FromContext(ctx)

	if strings.HasPrefix(req.Name, "system:kcp") {
		return ctrl.Result{Requeue: false}, nil
	}

	var clusterRoleBinding rbacv1.ClusterRoleBinding
	if err := r.Client.Get(ctx, req.NamespacedName, &clusterRoleBinding); err != nil {
		return ctrl.Result{}, err
	}

	logger.Info("clusterrolebinding mutated", "clusterrolebinding", clusterRoleBinding)

	_ = clusterRoleBinding

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterRoleBindingReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&rbacv1.ClusterRoleBinding{}).
		Complete(r)
}
