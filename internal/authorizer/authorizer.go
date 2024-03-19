package authorizer

import (
	"context"
	"fmt"
	"log"

	"github.com/jon-whit/openfga-authorizer/internal/resourcemapper"
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

	relationshipTuple := resourcemapper.AttributesToRelationshipTuple(a)

	object := relationshipTuple.Object.String()
	relation := relationshipTuple.Relation
	user := relationshipTuple.Subject.String()

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
	if err != nil {
		return authorizer.DecisionNoOpinion, "", err
	}

	if resp.GetAllowed() {
		return authorizer.DecisionAllow, "", nil
	}

	return authorizer.DecisionDeny, "", nil
}
