package rolectrl

import (
	"context"
	"fmt"
	"strings"

	openfgasdk "github.com/openfga/go-sdk/client"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type RoleReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	OpenFGAClient openfgasdk.SdkClient
}

// Reconcile implements reconcile.Reconciler.
func (r *RoleReconciler) Reconcile(
	ctx context.Context,
	req reconcile.Request,
) (reconcile.Result, error) {
	logger := log.FromContext(ctx)

	if strings.HasPrefix(req.Name, "system:kcp") {
		return ctrl.Result{Requeue: false}, nil
	}

	var role rbacv1.Role
	if err := r.Client.Get(ctx, req.NamespacedName, &role); err != nil {
		return ctrl.Result{}, err
	}

	logger.Info("role mutated", "role", role)

	// todo: reconile the OpenFGA relationship tuples based on the role
	roleNamespace := role.GetNamespace()
	roleName := role.GetName()

	mappedObject := fmt.Sprintf("k8s_role:namespace/%s/roles/%s", roleNamespace, roleName)

	writes := []openfgasdk.ClientTupleKey{
		{
			Object:   mappedObject,
			Relation: "contains",
			User:     fmt.Sprintf("k8s_namespace:%s", roleNamespace),
		},
	}

	for _, rule := range role.Rules {
		for _, verb := range rule.Verbs {
			for _, apiGroup := range rule.APIGroups {
				for _, resourceName := range rule.ResourceNames {
					tuple := openfgasdk.ClientTupleKey{
						Object:   fmt.Sprintf("k8s_resource:%s/namespaces/%s/%s", apiGroup, roleNamespace, resourceName),
						Relation: verb,
						User:     fmt.Sprintf("%s#assignee", mappedObject),
					}

					tupleStr := fmt.Sprintf("%s#%s@%s", tuple.GetObject(), tuple.GetRelation(), tuple.GetUser())

					logger.Info("adding tuple", "tupleKey", tupleStr)

					writes = append(writes, tuple)
				}
			}
		}

		//resources := rule.Resources[0]
		// _ = rule.NonResourceURLs
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
func (r *RoleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&rbacv1.Role{}).
		Complete(r)
}
