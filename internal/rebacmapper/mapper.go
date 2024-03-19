// Package rebacmapper provides mapping utilities to map Kubernetes
// resources to ReBAC mappings.
package rebacmapper

import (
	"fmt"
)

type RelationshipTuple struct {
	Object   Object
	Relation string
	Subject  isSubjectRef
}

type isSubjectRef interface {
	isSubjectRef()
	String() string
}

type Object struct {
	Type string
	ID   string
}

func (o Object) String() string {
	return fmt.Sprintf("%s:%s", o.Type, o.ID)
}

func (o Object) isSubjectRef() {}

type SubjectSet struct {
	Object   Object
	Relation string
}

func (ss SubjectSet) isSubjectRef() {}

func (ss SubjectSet) String() string {
	return fmt.Sprintf("%s#%s", ss.Object, ss.Relation)
}

// MapObjectToRelationshipTuples maps the provided Kubernetes Object to
// zero or more RelationshipTuples.
func MapObjectToRelationshipTuples(object any) []RelationshipTuple {
	var relationshipTuples []RelationshipTuple

	return relationshipTuples
}
