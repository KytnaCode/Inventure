package user

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/kytnacode/inventure/passhash"
)

// userRepo is an abstraction over user repository.
type userRepo interface {
	// CreateUser creates a user with the given data a return its ID.
	CreateUser(ctx context.Context, data *Data) (id uuid.UUID, err error)

	// UserByEmail returns a user by its email, must return [ErrUserNotFound]
	// if the user could not be found.
	UserByEmail(ctx context.Context, email string) (u *User, err error)
}

// Service handles high-level user operations.
type Service struct {
	repo userRepo
}

// NewService creates a new [Service].
func NewService(repo userRepo) *Service {
	return &Service{
		repo: repo,
	}
}

// Claims are authentication claims of a user.
type Claims struct {
	// ID is user's unique ID.
	ID string

	// RoleIDs are the IDs of the roles the user belongs, can be empty, but never nil.
	RoleIDs []string
}

// SignUpData is the necessary data to create a new user with password based authentication. Data
// is NOT validated.
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

// SignUp creates a new user with password based authentication.
func (s *Service) SignUp(ctx context.Context, data *SignUpData) (userData *Claims, err error) {
	m := &Data{
		Name:         data.Name,
		Email:        data.Email,
		PasswordHash: &data.PasswordHash,
	}

	id, err := s.repo.CreateUser(ctx, m)
	if err != nil {
		return nil, fmt.Errorf("could not create user: %w", err)
	}

	userData = &Claims{
		ID:      id.String(),
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
func (s *Service) SignIn(ctx context.Context, data *SignInData) (userData *Claims, err error) {
	u, err := s.repo.UserByEmail(ctx, data.Email)
	if err != nil {
		return nil, fmt.Errorf("could not sign in user: %w", err)
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
