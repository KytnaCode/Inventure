package user_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/kytnacode/inventure/internal/auth/rbac"
	"github.com/kytnacode/inventure/internal/user"
)

type testRepository struct {
	id  uuid.UUID
	err error
	u   *user.User
}

func (r *testRepository) CreateUser(_ context.Context, _ *user.Data) (uuid.UUID, error) {
	if r.err != nil {
		return uuid.UUID{}, r.err
	}

	return r.id, nil
}

func (r *testRepository) UserByEmail(_ context.Context, _ string) (*user.User, error) {
	if r.err != nil {
		return nil, r.err
	}

	return r.u, nil
}

func newService(repo *testRepository) *user.Service {
	return user.NewService(repo)
}

func TestService_SignUpShouldReturnRepositoryError(t *testing.T) {
	t.Parallel()

	ErrExpected := errors.New("expected")

	s := user.NewService(&testRepository{
		err: ErrExpected,
	})

	//nolint:gosec // fake credentials.
	claims, err := s.SignUp(t.Context(), &user.SignUpData{
		Name:         "username",
		Email:        "user@email.com",
		PasswordHash: "$argon2id$v=19$m=65536,t=3,p=4$0B7fXhc2KKWuhCsyyYqDGQ$3hiOlhK7ubLJPFFt5dLN8zq8PnZX+mHlogk/toFxKaQ",
	})
	if claims != nil {
		t.Errorf("expected nil claims: got '%v'", claims)
	}

	if err == nil {
		t.Fatal("expected non-nil error")
	}

	if !errors.Is(err, ErrExpected) {
		t.Errorf("expected repository error: got '%v'", err)
	}
}

func TestService_SignUpShouldReturnClaims(t *testing.T) {
	t.Parallel()

	u := &user.User{
		ID:    uuid.New(),
		Name:  "Luz Noceda",
		Email: "luz.noceda@gmail.com",
		PasswordHash: new(
			"$argon2id$v=19$m=65536,t=3,p=4$0B7fXhc2KKWuhCsyyYqDGQ$3hiOlhK7ubLJPFFt5dLN8zq8PnZX+mHlogk/toFxKaQ",
		),
		On: rbac.NewResource("tenant", uuid.New(), nil),
	}

	s := user.NewService(&testRepository{
		u: u,
	})

	claims, err := s.SignUp(t.Context(), &user.SignUpData{
		Name:         u.Name,
		Email:        u.Email,
		PasswordHash: *u.PasswordHash,
	})
	if err != nil {
		t.Errorf("expected nil error: %v", err)
	}

	if claims == nil {
		t.Fatal("expected non-nil claims")
	}
}

func TestRepository_SignInShouldReturnNoPasswordAuthErrorSqlite(t *testing.T) {
	t.Parallel()

	u := &user.User{
		ID:           uuid.New(),
		Email:        "expected@email.com",
		PasswordHash: nil,
	}

	s := newService(&testRepository{
		u: u,
	})

	claims, err := s.SignIn(t.Context(), &user.SignInData{
		Email: u.Email,
	})
	if claims != nil {
		t.Errorf("expected empty user claims: got '%v'", claims)
	}

	if err == nil {
		t.Fatal("expected a non-nil error")
	}

	if !errors.Is(err, user.ErrNotPasswordAuth) {
		t.Errorf("expected not password auth error: got '%v'", err)
	}
}

func TestRepository_SignInShouldReturnWrongCredentialsErrorSqlite(t *testing.T) {
	t.Parallel()

	// "hello"
	actualHash := "$argon2id$v=19$m=65536,t=3,p=4$0B7fXhc2KKWuhCsyyYqDGQ$3hiOlhK7ubLJPFFt5dLN8zq8PnZX+mHlogk/toFxKaQ"

	const otherPassword = "luz-noceda"

	u := &user.User{
		Name:         "my valid user name",
		Email:        "my-valid@email.com",
		PasswordHash: &actualHash,
	}

	s := newService(&testRepository{
		u: u,
	})

	claims, err := s.SignIn(t.Context(), &user.SignInData{
		Email:         u.Email,
		ClearPassword: otherPassword,
	})
	if claims != nil {
		t.Errorf("expected empty user claims: got '%v'", claims)
	}

	if err == nil {
		t.Fatal("expected a non-nil error")
	}

	if !errors.Is(err, user.ErrWrongCredentials) {
		t.Errorf("expected wrong credentials error: got '%v'", err)
	}
}
