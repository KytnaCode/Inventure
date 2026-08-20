package user

import "github.com/kytnacode/inventure/internal/role"

type User struct {
	// TODO: Allow Unicode alphabetic characters.
	// Name is user display name.
	Name string `validate:"required,min=3,max=80,alphanumspace"`

	Roles []role.Role
}
