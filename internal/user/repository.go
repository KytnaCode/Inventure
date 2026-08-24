package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
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

// Data is the authentication data for a user.
type Data struct {
	ID      string
	RoleIDs []string
}

// Model is the user's database model. For the domain object see [user.User].
type Model struct {
	gorm.Model

	ID uuid.UUID `gorm:"primaryKey"`

	// Name is user display name.
	Name string `validate:"required,min=3,max=80,resourcename"`

	Email string `validate:"required,email"`

	PasswordHash *string

	// ResourceType is the type of the resource the user belongs to.
	ResourceType string

	// ResourceID is the ID of the resource the user belongs to.
	ResourceID uuid.UUID

	CreatedAt time.Time
	UpdatedAt time.Time
}

// TableName implements [gorm/schema.Tabler].
func (m *Model) TableName() string {
	return "users"
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
	table gorm.Interface[Model]
	v     *validator.Validate
}

// NewRepository creates a new [Repository].
func NewRepository(db gorm.Interface[Model], v *validator.Validate) *Repository {
	return &Repository{
		table: db,
		v:     v,
	}
}

// SignUp creates a new user with password based authentication.
func (r *Repository) SignUp(ctx context.Context, data *SignUpData) (userData *Data, err error) {
	m := &Model{
		ID:           uuid.New(),
		Name:         data.Name,
		Email:        data.Email,
		PasswordHash: &data.PasswordHash,
	}

	if err := r.v.Struct(m); err != nil {
		return nil, fmt.Errorf("invalid user model: %w", err)
	}

	err = r.table.Create(ctx, m)
	if err != nil {
		return nil, fmt.Errorf("could not create user: %w", err)
	}

	userData = &Data{
		ID: m.ID.String(),
	}

	return userData, nil
}

// SignIn search for a user by the given data and verifies its credentials, if credentials
// are valid and no error occurs, then its id is returned.
//
// If the user cannot be found, [ErrUserNotFound] is returned, if user account doesn't support
// password-based authentication, [ErrNotPasswordAuth] is returned, at last, if user was found
// but their credentials don't match, [ErrWrongCredentials] is returned.
func (r *Repository) SignIn(ctx context.Context, data *SignInData) (userData *Data, err error) {
	u, err := r.table.Where("email = ?", data.Email).Take(ctx)
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

	userData = &Data{
		ID: u.ID.String(),
	}

	return userData, nil
}
