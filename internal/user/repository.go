package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var _ userRepo = &Repository{}

var (
	// ErrUserNotFound is returned when the user with the given credentials could not be found.
	ErrUserNotFound = errors.New("user not found")

	// ErrNotPasswordAuth is returned if the user signed up with other method than password based
	// authentication.
	ErrNotPasswordAuth = errors.New("user doesn't accept password authentication")

	// ErrWrongCredentials is returned on user's credentials mismatch.
	ErrWrongCredentials = errors.New("wrong email or password")
)

// Data is the necessary data to create a user.
type Data struct {
	// Name is the new user's name.
	Name string

	// Email is the new user's email.
	Email string

	// PasswordHash is new user's password hash, nil for non password-based authentication users.
	PasswordHash *string
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

// CreateUser creates a new user with the given data. Returns [ErrDuplicatedUser] if the email
// is already registered.
func (r *Repository) CreateUser(ctx context.Context, data *Data) (id uuid.UUID, err error) {
	m := &Model{
		ID:           uuid.New(),
		Name:         data.Name,
		Email:        data.Email,
		PasswordHash: data.PasswordHash,
	}

	err = r.db.WithContext(ctx).Create(m).Error
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("could not create user: %w", err)
	}

	return m.ID, nil
}

// UserByEmail returns a user by its email address, returns [ErrUserNotFound] if user with the
// given email could not be found.
func (r *Repository) UserByEmail(ctx context.Context, email string) (u *User, err error) {
	var m Model

	err = r.db.WithContext(ctx).Where("email = ?", email).Take(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("user not exists: %w", errors.Join(err, ErrUserNotFound))
	}

	if err != nil {
		return nil, fmt.Errorf("could not get user: %w", err)
	}

	return m.ToDomain(), nil
}
