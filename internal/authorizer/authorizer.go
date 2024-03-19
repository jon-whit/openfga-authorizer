package authorizer

import (
	"context"
	"fmt"
	"log"
	"slices"

	openfgasdk "github.com/openfga/go-sdk/client"
	"k8s.io/apiserver/pkg/authorization/authorizer"
)

var _ authorizer.Authorizer = (*OpenFGAAuthorizer)(nil)

type OpenFGAAuthorizer struct {
	OpenFGAClient openfgasdk.SdkClient
}

// Authorize implements authorizer.Authorizer.
func (o *OpenFGAAuthorizer) Authorize(
	ctx context.Context, a authorizer.Attributes,
) (authorizer.Decision, string, error) {

	log.Printf("verb: %s, ns: %s, name: %s, resource: %s\n", a.GetVerb(), a.
		GetNamespace(), a.GetName(), a.GetResource())

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

	object := fmt.Sprintf("k8s_resource:%s", resource)
	relation := a.GetVerb()
	user := fmt.Sprintf("user:%s", a.GetUser().GetName())

	log.Printf("object='%s', relation='%s', user='%s'\n", object, relation, user)

	resp, err := o.OpenFGAClient.
		Check(context.Background()).
		Body(openfgasdk.ClientCheckRequest{
			Object:   object,
			Relation: relation,
			User:     user,
			ContextualTuples: []openfgasdk.ClientContextualTupleKey{
				{
					Object:   fmt.Sprintf("k8s_namespace:%s", a.GetNamespace()),
					Relation: "operates_in",
					User:     user,
				},
			},
		}).
		Execute()
	/*resp, err := o.OpenFGAClient.Check(context.Background(), &openfgav1.CheckRequest{
		StoreId: o.StoreID,
		TupleKey: &openfgav1.CheckRequestTupleKey{
			Object:   object,
			Relation: relation,
			User:     user,
		},
		ContextualTuples: &openfgav1.ContextualTupleKeys{
			TupleKeys: []*openfgav1.TupleKey{
				{
					Object:   fmt.Sprintf("k8s_namespace:%s", a.GetNamespace()),
					Relation: "operates_in",
					User:     user,
				},
			},
		},
	})*/
	if err != nil {
		return authorizer.DecisionNoOpinion, "", err
	}

	if resp.GetAllowed() {
		return authorizer.DecisionAllow, "", nil
	}

	return authorizer.DecisionDeny, "", nil
}
