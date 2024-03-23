package resourcemapper

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

var (
	Scheme = runtime.NewScheme()
	Codecs = serializer.NewCodecFactory(Scheme)
)

func TestRoleToRelationshipTuples(t *testing.T) {
	rbacv1.AddToScheme(Scheme)

	tests := []struct {
		name     string
		path     string
		expected []string
	}{
		{
			path: `testdata/role.yaml`,
			expected: []string{
				"k8s_role:namespace/fga-backend/deployment-reader#contains@k8s_namespace:fga-backend",
				"k8s_resource:/namespaces/fga-backend/deployments#get@k8s_role:namespace/fga-backend/deployment-reader#assignee",
				"k8s_resource:/namespaces/fga-backend/deployments#list@k8s_role:namespace/fga-backend/deployment-reader#assignee",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fileData, err := os.ReadFile(test.path)
			require.NoError(t, err)

			obj, _, err := Codecs.UniversalDeserializer().Decode(fileData, nil, nil)
			require.NoError(t, err)

			role := obj.(*rbacv1.Role)

			tuples := RoleToRelationshipTuples(*role)

			actual := make([]string, 0, len(tuples))
			for i := 0; i < len(tuples); i++ {
				actual[i] = tuples[i].String()
			}

			require.ElementsMatch(t, actual, test.expected)
		})
	}
}
