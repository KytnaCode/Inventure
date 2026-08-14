package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/kytnacode/inventure/pkg/passhash"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

var (
	// ErrUserNotFound is returned when the user with the given credentials could not be found.
	ErrUserNotFound = errors.New("user not found")

	// ErrNotPasswordAuth is returned if the user signed up with other method than password based
	// authentication.
	ErrNotPasswordAuth = errors.New("user doesn't accept password authentication")

	// ErrWrongCredentials is returned on user's credentials mismatch.
	ErrWrongCredentials = errors.New("wrong email or password")
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

// SignInUserData is the necessary data for a user to sign in.
type SignInUserData struct {
	Email         string
	ClearPassword string
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
		ID:           datatypes.NewUUIDv4(),
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

// SignInUser search for a user by the given data and verifies its credentials, if credentials
// are valid and no error occurs, then its id is returned.
//
// If the user cannot be found, [ErrUserNotFound] is returned, if user account doesn't support
// password-based authentication, [ErrNotPasswordAuth] is returned, at last, if user was found
// but their credentials don't match, [ErrWrongCredentials] is returned.
func (r *Repository) SignInUser(ctx context.Context, data *SignInUserData) (id string, err error) {
	u, err := r.table.Where("email = ?", data.Email).Take(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrUserNotFound
		}

		return "", fmt.Errorf("user query failed: %w", err)
	}

	if u.PasswordHash == nil {
		return "", ErrNotPasswordAuth
	}

	ok, err := passhash.Verify(*u.PasswordHash, []byte(data.ClearPassword))
	if err != nil {
		return "", fmt.Errorf("could not authenticate user: %w", err)
	}

	if !ok {
		return "", ErrWrongCredentials
	}

	return u.ID.String(), nil
}
