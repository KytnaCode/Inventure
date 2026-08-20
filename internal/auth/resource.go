package auth

import (
	"errors"
	"fmt"

	"gorm.io/datatypes"
)

// ErrInvalidUUID is returned if a string is not a valid UUID.
var ErrInvalidUUID = errors.New("string is not a valid UUID")

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

func toUUID(id string) (datatypes.UUID, error) {
	const uuidLen = 16

	if len(id) != uuidLen {
		return [uuidLen]byte{}, ErrInvalidUUID
	}

	var v [16]byte

	copy(v[:], id)

	return datatypes.UUID(v), nil
}
