package resourcemapper

import (
	"fmt"

	"github.com/jon-whit/openfga-authorizer/internal/rebac"
	rbacv1 "k8s.io/api/rbac/v1"
)

func RoleToRelationshipTuples(role rbacv1.Role) []rebac.RelationshipTuple {
	roleNamespace := role.GetNamespace()
	roleName := role.GetName()

	var relationshipTuples []rebac.RelationshipTuple
	relationshipTuples = append(relationshipTuples, rebac.RelationshipTuple{
		Object: rebac.Object{
			Type: "k8s_role",
			ID:   fmt.Sprintf("namespace/%s/roles/%s", roleNamespace, roleName),
		},
		Relation: "contains",
		Subject: rebac.Object{
			Type: "k8s_namespace",
			ID:   roleNamespace,
		},
	})

	for _, rule := range role.Rules {
		for _, verb := range rule.Verbs {
			for _, apiGroup := range rule.APIGroups {
				for _, resourceName := range rule.ResourceNames {
					tuple := rebac.RelationshipTuple{
						Object: rebac.Object{
							Type: "k8s_resource",
							ID:   fmt.Sprintf("%s/namespaces/%s/%s", apiGroup, roleNamespace, resourceName),
						},
						Relation: verb,
						Subject: rebac.SubjectSet{
							Object: rebac.Object{
								Type: "k8s_role",
								ID:   fmt.Sprintf("namespace/%s/roles/%s", roleNamespace, roleName),
							},
							Relation: "assignee",
						},
					}

					relationshipTuples = append(relationshipTuples, tuple)
				}
			}
		}

		//resources := rule.Resources[0]
		// _ = rule.NonResourceURLs
	}
	return relationshipTuples
}
