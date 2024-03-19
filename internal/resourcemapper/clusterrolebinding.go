package resourcemapper

import (
	"github.com/jon-whit/openfga-authorizer/internal/rebac"
	rbacv1 "k8s.io/api/rbac/v1"
)

func ClusterRoleBindingToRelationshipTuples(crBinding rbacv1.ClusterRoleBinding) []rebac.RelationshipTuple {
	clusterRoleBindingName := crBinding.GetName()

	clusterRoleRef := crBinding.RoleRef
	clusterRoleName := clusterRoleRef.Name

	var relationshipTuples []rebac.RelationshipTuple

	relationshipTuples = append(relationshipTuples, rebac.RelationshipTuple{
		Object: rebac.Object{
			Type: "k8s_clusterrole",
			ID:   clusterRoleName,
		},
		Relation: "assignee",
		Subject: rebac.SubjectSet{
			Object: rebac.Object{
				Type: "k8s_clusterrolebinding",
				ID:   clusterRoleBindingName,
			},
			Relation: "assignee",
		},
	})

	for _, subject := range crBinding.Subjects {
		if subject.Kind == rbacv1.GroupKind {
			relationshipTuples = append(relationshipTuples, rebac.RelationshipTuple{
				Object: rebac.Object{
					Type: "k8s_clusterrolebinding",
					ID:   clusterRoleBindingName,
				},
				Relation: "assignee",
				Subject: rebac.SubjectSet{
					Object: rebac.Object{
						Type: "group",
						ID:   subject.Name,
					},
					Relation: "member",
				},
			})
		}
	}

	return relationshipTuples
}
