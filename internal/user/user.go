package user

import "github.com/kytnacode/inventure/internal/role"

// User is the domain model for a user.
type User struct {
	// TODO: Allow Unicode alphabetic characters.
	// Name is user display name.
	Name string `validate:"required,min=3,max=80,alphanumspace"`

	// Roles are the roles the user belongs to.
	Roles []role.Role
}
