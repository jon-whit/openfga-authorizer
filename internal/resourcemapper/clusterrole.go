package resourcemapper

import (
	"fmt"

	"github.com/jon-whit/openfga-authorizer/internal/rebac"
	rbacv1 "k8s.io/api/rbac/v1"
)

func ClusterRoleToRelationshipTuples(clusterRole rbacv1.ClusterRole) []rebac.RelationshipTuple {
	clusterRoleNamespace := clusterRole.GetNamespace()
	clusterRoleName := clusterRole.GetName()

	var relationshipTuples []rebac.RelationshipTuple

	for _, rule := range clusterRole.Rules {
		for _, verb := range rule.Verbs {
			for _, apiGroup := range rule.APIGroups {
				for _, resourceName := range rule.ResourceNames {
					tuple := rebac.RelationshipTuple{
						Object: rebac.Object{
							Type: "k8s_resource",
							ID:   fmt.Sprintf("%s/namespaces/%s/%s", apiGroup, clusterRoleNamespace, resourceName),
						},
						Relation: verb,
						Subject: rebac.SubjectSet{
							Object: rebac.Object{
								Type: "k8s_role",
								ID:   fmt.Sprintf("namespace/%s/roles/%s", clusterRoleNamespace, clusterRoleName),
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
