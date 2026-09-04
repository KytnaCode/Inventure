package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/kytnacode/inventure/internal/auth/rbac"
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

// UpdateByEmail updates a user by email. The updateFn function takes a user model with the user
// data, changes made to the given user will be uploaded to the database. Atomically fetches and
// updates user.
func (r *Repository) UpdateByEmail(
	ctx context.Context,
	email string,
	updateFn func(u *Model) error,
) error {
	var u Model

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Model(&Model{}).
			Where("email = ?", email).
			Take(&u).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("user not exists: %w", errors.Join(err, ErrUserNotFound))
		}

		if err != nil {
			return err
		}

		err = updateFn(&u)
		if err != nil {
			return err
		}

		return tx.Model(u).Updates(u).Error
	})
	if err != nil {
		return fmt.Errorf("could not update user: %w", err)
	}

	return nil
}

// AssingRoles roles assign the roles with the given IDs to the user with the given ID.
func (r *Repository) AssingRoles(
	ctx context.Context,
	userID uuid.UUID,
	roleIDs ...uuid.UUID,
) error {
	roles := make([]rbac.RoleModel, 0, len(roleIDs))

	for _, id := range roleIDs {
		roles = append(roles, rbac.RoleModel{
			ID: id,
		})
	}

	err := r.db.WithContext(ctx).
		Model(&Model{ID: userID}).
		Association("Roles").
		Append(&roles)
	if err != nil {
		return fmt.Errorf("could not assing roles to user: %w", err)
	}

	return nil
}
