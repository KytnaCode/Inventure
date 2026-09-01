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
}

// TableName implements [gorm/schema.Tabler].
func (m *Model) TableName() string {
	return "users"
}

// ToDomain converts a model into a domain user representation.
func (m *Model) ToDomain() *User {
	return &User{
		ID:    m.ID,
		Name:  m.Name,
		Email: m.Email,
	}
}
