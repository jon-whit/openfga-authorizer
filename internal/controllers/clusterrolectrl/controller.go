package clusterrolectrl

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

type ClusterRoleReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	OpenFGAClient openfgasdk.SdkClient
}

// Reconcile implements reconcile.Reconciler.
func (r *ClusterRoleReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	logger := log.FromContext(ctx)

	if strings.HasPrefix(req.Name, "system:kcp") {
		return ctrl.Result{Requeue: false}, nil
	}

	var clusterRole rbacv1.ClusterRole
	if err := r.Client.Get(ctx, req.NamespacedName, &clusterRole); err != nil {
		return ctrl.Result{}, err
	}

	logger.Info("clusterrole mutated", "clusterrole", clusterRole)

	clusterRoleNamespace := clusterRole.GetNamespace()
	clusterRoleName := clusterRole.GetName()

	var writes []openfgasdk.ClientTupleKey

	_ = clusterRoleName
	_ = clusterRoleNamespace

	for _, rule := range clusterRole.Rules {
		_ = rule.APIGroups
		_ = rule.Resources
		_ = rule.ResourceNames
		_ = rule.Verbs
	}

	_, err := r.OpenFGAClient.
		Write(ctx).
		Body(openfgasdk.ClientWriteRequest{
			Writes: writes,
		}).
		Execute()

	return ctrl.Result{}, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterRoleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&rbacv1.ClusterRole{}).
		Complete(r)
}
