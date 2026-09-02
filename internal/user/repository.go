package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/kytnacode/inventure/passhash"
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

// Claims are authentication claims of a user.
type Claims struct {
	// ID is user's unique ID.
	ID string

	// RoleIDs are the IDs of the roles the user belongs, can be empty, but never nil.
	RoleIDs []string
}

// SignUpData is the necessary data to create a new user with password based authentication
// using [Repository.SignUp]. For validation rules see [repository.User].
type SignUpData struct {
	// Name is user display name.
	Name string

	// Email is user email.
	Email string

	// PasswordHash is a PHC formatted argon2id hash.
	PasswordHash string
}

// SignInData is the necessary data for a user to sign in.
type SignInData struct {
	// Email is user email.
	Email string

	// ClearPassword is the password in clear text.
	ClearPassword string
}

// Repository handles user related business logic.
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new [Repository].
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{
		db: db,
	}
}

// SignUp creates a new user with password based authentication.
func (r *Repository) SignUp(ctx context.Context, data *SignUpData) (userData *Claims, err error) {
	m := &Model{
		ID:           uuid.New(),
		Name:         data.Name,
		Email:        data.Email,
		PasswordHash: &data.PasswordHash,
	}

	err = r.db.WithContext(ctx).Create(m).Error
	if err != nil {
		return nil, fmt.Errorf("could not create user: %w", err)
	}

	userData = &Claims{
		ID:      m.ID.String(),
		RoleIDs: make([]string, 0),
	}

	return userData, nil
}

// SignIn search for a user by the given data and verifies its credentials, if credentials
// are valid and no error occurs, then its id is returned.
//
// If the user cannot be found, [ErrUserNotFound] is returned, if user account doesn't support
// password-based authentication, [ErrNotPasswordAuth] is returned, at last, if user was found
// but their credentials don't match, [ErrWrongCredentials] is returned.
func (r *Repository) SignIn(ctx context.Context, data *SignInData) (userData *Claims, err error) {
	var u Model

	err = r.db.WithContext(ctx).Where("email = ?", data.Email).Take(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}

		return nil, fmt.Errorf("user query failed: %w", err)
	}

	if u.PasswordHash == nil {
		return nil, ErrNotPasswordAuth
	}

	ok, err := passhash.Verify(*u.PasswordHash, []byte(data.ClearPassword))
	if err != nil {
		return nil, fmt.Errorf("could not authenticate user: %w", err)
	}

	if !ok {
		return nil, ErrWrongCredentials
	}

	userData = &Claims{
		ID:      u.ID.String(),
		RoleIDs: make([]string, 0),
	}

	return userData, nil
}
