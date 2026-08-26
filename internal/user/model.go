package user

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Model is the user's database model. For the domain object see [user.User].
type Model struct {
	gorm.Model

	// ID is the user's unique ID.
	ID uuid.UUID `gorm:"primaryKey"`

	// Name is user display name.
	Name string `validate:"required,min=3,max=80,resourcename"`

	// Email is user's unique email.
	Email string `validate:"required,email"`

	// PasswordHash is user's password hash for users using baked-in auth, nil for non baked-in auth
	// users.
	PasswordHash *string

	// ResourceType is the type of the resource the user belongs to.
	ResourceType string

	// ResourceID is the ID of the resource the user belongs to.
	ResourceID uuid.UUID
}

// TableName implements [gorm/schema.Tabler].
func (m *Model) TableName() string {
	return "users"
}
