package user

import (
	"github.com/google/uuid"
	"github.com/kytnacode/inventure/internal/auth/rbac"
)

// User is the domain model for a user.
type User struct {
	// ID is user's unique ID.
	ID uuid.UUID

	// Name is user display name.
	Name string `validate:"required,min=3,max=80,resourcename"`

	// Email is user's unique email.
	Email string `validate:"required,email"`

	PasswordHash *string

	// On is the resource the user is scoped on.
	On rbac.Resource
}
