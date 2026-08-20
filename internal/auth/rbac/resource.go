package rbac

import (
	"fmt"
)

// Resource represents a resource that can contain roles or can be subject of permissions.
type Resource struct {
	Typ string
	ID  string
}

// String returns a string representation of a resource.
func (r *Resource) String() string {
	return fmt.Sprintf("%v:%v", r.Typ, r.ID)
}

// NewResource creates a new [Resource].
func NewResource(typ, id string) Resource {
	return Resource{
		Typ: typ,
		ID:  id,
	}
}
