package rbac

import (
	"fmt"

	"github.com/google/uuid"
)

// Role is a role for RBAC.
type Role struct {
	// ID is role's unique ID.
	ID uuid.UUID

	// Name is role's unique name.
	Name string

	// Accesses are a list of permission grants over a resource.
	Accesses []Access

	// On is the resource the role belongs to, role permission are not applicable to resources
	// that is not the role's resource or a child of it.
	On Resource
}

// Access is a permission grant for a role on a given resource.
type Access struct {
	// ID is access' unique ID.
	ID uuid.UUID

	// Role is the role which the permission is granted on.
	Role *Role

	// Perms are the permissions granted.
	Perms []Perm

	// On are the resource on which the access is granted.
	On Resource
}

// Resource is a resource that can contain roles and be subject of permission grants.
type Resource struct {
	// Parent is the resource parent, nil for a root resource.
	Parent *Resource

	// Typ is the resource type.
	Typ string

	// ID is the resource unique ID.
	ID uuid.UUID
}

// NewResource creates a new [Resource]. If parent is nil the resource will be interpreted as a
// root resource.
func NewResource(typ string, id uuid.UUID, parent *Resource) Resource {
	return Resource{
		Typ:    typ,
		ID:     id,
		Parent: parent,
	}
}

// String returns a string representation of the resource.
func (r *Resource) String() string {
	return fmt.Sprintf("[%v: %v]", r.Typ, r.ID)
}

// Equal check if a resource is equal to another. Resource's parent is not considered for equality.
func (r *Resource) Equal(other Resource) bool {
	return r.Typ == other.Typ && r.ID == other.ID
}

// Reference is a resource reference, can optionally contains a parent reference, but it should not
// be interpreted as a root resource if parent's reference is missing.
type Reference struct {
	Typ    string
	ID     uuid.UUID
	Parent *Reference
}

// NewReference returns a resource reference, can optionally contain a parent reference, a nil
// parent reference must be interpreted as 'unknown' parent, resource may or not have a parent
// but is not loaded.
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

// Perm is a permission that can be granted.
type Perm string
