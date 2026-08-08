package repository

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// User is the user's database model. For the domain object see [user.User].
type User struct {
	gorm.Model

	ID datatypes.UUID `gorm:"primaryKey"`

	// TODO: Allow Unicode alphabetic characters.
	// Name is user display name.
	Name string `validate:"required,min=3,max=80,alphanumspace"`

	Email string `validate:"required,email"`

	PasswordHash *string

	CreatedAt time.Time
	UpdatedAt time.Time
}
