package entity

import (
	"fmt"

	"github.com/google/uuid"
)

// Reference are entity reference names for each entity type.
const (
	ReferenceItem   = "items"
	ReferenceRetail = "retails"
	ReferenceTenant = "tenants"
	ReferenceUser   = "users"
)

// Reference is an entity reference, can optionally contains a parent reference, but it should not
// be interpreted as a root entity if parent's reference is missing.
type Reference struct {
	Typ    string
	ID     uuid.UUID
	Parent *Reference
}

// NewReference returns an entity reference, can optionally contain a parent reference, a nil
// parent reference must be interpreted as 'unknown' parent, resource may or not have a parent
// but it's not loaded.
func NewReference(typ string, id uuid.UUID, parent *Reference) Reference {
	return Reference{
		Parent: parent,
		Typ:    typ,
		ID:     id,
	}
}

// String returns a string representation of the reference.
func (r *Reference) String() string {
	return fmt.Sprintf("[%v: %v]", r.Typ, r.ID)
}

// Equal check if a reference is equal to another. Reference's parent is not considered for
// equality.
func (r *Reference) Equal(other Reference) bool {
	return r.Typ == other.Typ && r.ID == other.ID
}
