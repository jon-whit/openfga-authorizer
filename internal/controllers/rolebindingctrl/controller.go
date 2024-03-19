package rolebindingctrl

import (
	"context"
	"strings"

	"github.com/jon-whit/openfga-authorizer/internal/resourcemapper"
	openfgasdk "github.com/openfga/go-sdk/client"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type RoleBindingReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	OpenFGAClient openfgasdk.SdkClient
}

// Reconcile implements reconcile.Reconciler.
func (r *RoleBindingReconciler) Reconcile(
	ctx context.Context,
	req reconcile.Request,
) (reconcile.Result, error) {
	logger := log.FromContext(ctx)

	if strings.HasPrefix(req.Name, "system:kcp") {
		return ctrl.Result{Requeue: false}, nil
	}

	var roleBinding rbacv1.RoleBinding
	if err := r.Client.Get(ctx, req.NamespacedName, &roleBinding); err != nil {
		return ctrl.Result{}, err
	}

	logger.Info("rolebinding mutated", "rolebinding", roleBinding)

	relationshipTuples := resourcemapper.RoleBindingToRelationshipTuples(roleBinding)

	var writes []openfgasdk.ClientTupleKey
	for _, tuple := range relationshipTuples {
		writes = append(writes, openfgasdk.ClientTupleKey{
			Object:   tuple.Object.String(),
			Relation: tuple.Relation,
			User:     tuple.Subject.String(),
		})
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
func (r *RoleBindingReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&rbacv1.RoleBinding{}).
		Complete(r)
}
