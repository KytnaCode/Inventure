package repository

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

// SignUpUserData is the necessary data to create a new user with password based authentication
// using [Repository.SignUpUser]. For validation rules see [repository.User].
type SignUpUserData struct {
	// Name is user display name.
	Name string

	// Email is user email.
	Email string

	// PasswordHash is a PHC formatted argon2id hash.
	PasswordHash string
}

// Repository handles user related business logic.
type Repository struct {
	table gorm.Interface[User]
	v     *validator.Validate
}

// New creates a new [Repository].
func New(db gorm.Interface[User], v *validator.Validate) *Repository {
	return &Repository{
		table: db,
		v:     v,
	}
}

// SignUpUser creates a new user with password based authentication.
func (r *Repository) SignUpUser(ctx context.Context, data *SignUpUserData) (id string, err error) {
	m := &User{
		Name:         data.Name,
		Email:        data.Email,
		PasswordHash: &data.PasswordHash,
	}

	if err := r.v.Struct(m); err != nil {
		return "", fmt.Errorf("invalid user model: %w", err)
	}

	err = r.table.Create(ctx, m)
	if err != nil {
		return "", fmt.Errorf("could not create user: %w", err)
	}

	return m.ID.String(), nil
}
