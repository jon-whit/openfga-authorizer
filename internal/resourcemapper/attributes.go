// Package resourcemapper provides utilities to map Kubernetes resources
// to ReBAC Relationship Tuples.
package resourcemapper

import (
	"fmt"
	"slices"

	"github.com/jon-whit/openfga-authorizer/internal/rebac"
	"k8s.io/apiserver/pkg/authorization/authorizer"
)

// AttributesToRelationshipTuple converts the provided Kubernetes
// authorizer attributes to the appropriate Relationship Tuple
// representing the authorization request.
func AttributesToRelationshipTuple(a authorizer.Attributes) rebac.RelationshipTuple {
	resourceBasePath := fmt.Sprintf("%s/%s", a.GetAPIGroup(), a.GetAPIVersion())

	var resourceName string
	if slices.Contains([]string{"watch", "list", "create"}, a.GetVerb()) {
		resourceName = a.GetResource()
	} else {
		resourceName = fmt.Sprintf("%s/%s", a.GetResource(), a.GetName())
	}

	resource := fmt.Sprintf("%s/%s", resourceBasePath, resourceName)

	if a.GetNamespace() != "" {
		resource = fmt.Sprintf("%s/namespaces/%s/%s", resourceBasePath, a.GetNamespace(), resourceName)
	}

	return rebac.RelationshipTuple{
		Object: rebac.Object{
			Type: "k8s_resource",
			ID:   resource,
		},
		Relation: a.GetVerb(),
		Subject: rebac.Object{
			Type: "user",
			ID:   a.GetUser().GetName(),
		},
	}
}
