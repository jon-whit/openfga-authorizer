package resourcemapper

import (
	"fmt"

	"github.com/jon-whit/openfga-authorizer/internal/rebac"
	rbacv1 "k8s.io/api/rbac/v1"
)

func RoleBindingToRelationshipTuples(roleBinding rbacv1.RoleBinding) []rebac.RelationshipTuple {
	roleBindingNamespace := roleBinding.GetNamespace()
	roleBindingName := roleBinding.GetName()

	roleRef := roleBinding.RoleRef

	var relationshipTuples []rebac.RelationshipTuple

	relationshipTuples = append(relationshipTuples, rebac.RelationshipTuple{
		Object: rebac.Object{
			Type: "k8s_role",
			ID:   fmt.Sprintf("namespace/%s/roles/%s", roleBindingNamespace, roleRef.Name),
		},
		Relation: "namespaced_assignee",
		Subject: rebac.SubjectSet{
			Object: rebac.Object{
				Type: "k8s_rolebinding",
				ID:   fmt.Sprintf("namespace/%s/rolebindings/%s", roleBindingNamespace, roleBindingName),
			},
			Relation: "namespaced_assignee",
		},
	})

	for _, subject := range roleBinding.Subjects {
		if subject.Kind == rbacv1.GroupKind {
			relationshipTuples = append(relationshipTuples, rebac.RelationshipTuple{
				Object: rebac.Object{
					Type: "k8s_rolebinding",
					ID:   fmt.Sprintf("namespace/%s/rolebindings/%s", roleBindingNamespace, roleBindingName),
				},
				Relation: "namespaced_assignee",
				Subject: rebac.SubjectSet{
					Object: rebac.Object{
						Type: "group",
						ID:   subject.Name,
					},
					Relation: "assignee",
				},
			})
		}
	}

	return relationshipTuples
}
