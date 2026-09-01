package user_test

import (
	"errors"
	"path"
	"testing"

	"github.com/google/uuid"
	"github.com/kytnacode/inventure/internal/user"
	"github.com/kytnacode/inventure/validation"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newSqliteRepo(t *testing.T) (*user.Repository, gorm.Interface[user.Model]) {
	t.Helper()

	dbPath := path.Join(t.TempDir(), "test.db")

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("could not open database: %v", err)
	}

	if err := db.AutoMigrate(user.Model{}); err != nil {
		t.Fatalf("could not migrate user's schema: %v", err)
	}

	table := gorm.G[user.Model](db)

	repo := user.NewRepository(table, validation.New())

	return repo, table
}

func TestRepositorySqliteShouldCreateUser(t *testing.T) {
	t.Parallel()

	//nolint:gosec // fake credentials.
	data := user.SignUpData{
		Name:         "my user name",
		Email:        "my-user@email.com",
		PasswordHash: "$argon2id$v=19$m=65536,t=3,p=4$uiOyD7yhvKuJwK2B+mYX9w$ZwOB3SQDZWgSI17gaRUJ5Slzr5SH8XErpN/ihlacVR8",
	}

	repo, table := newSqliteRepo(t)

	userData, err := repo.SignUp(t.Context(), &data)
	if err != nil {
		t.Fatalf("could not create user: %v", err)
	}

	user, err := table.Where("id = ?", userData.ID).Take(t.Context())
	if err != nil {
		t.Fatalf("could not get created user: %v", err)
	}

	if user.Name != data.Name {
		t.Errorf("expected name to be '%v': got '%v'", data.Name, user.Name)
	}

	if user.Email != data.Email {
		t.Errorf("expected email to be '%v': got '%v'", data.Email, user.Email)
	}

	if user.PasswordHash == nil {
		t.Fatalf("expected a non nil password hash")
	}

	if *user.PasswordHash != data.PasswordHash {
		t.Errorf(
			"expected password hash to be '%v': got '%v'",
			data.PasswordHash,
			user.PasswordHash,
		)
	}
}

func TestRepository_SignInShouldReturnUserNotFoundSqlite(t *testing.T) {
	t.Parallel()

	r, _ := newSqliteRepo(t)

	userData, err := r.SignIn(t.Context(), &user.SignInData{
		Email:         "super-real@email.com",
		ClearPassword: "my-super-real-password",
	})
	if userData != nil {
		t.Errorf("expected empty user data: got '%v'", userData)
	}

	if err == nil {
		t.Fatal("exected a non-nil error")
	}

	if !errors.Is(err, user.ErrUserNotFound) {
		t.Fatalf("expected user not found error: got '%v'", err)
	}
}

func TestRepository_SignInShouldReturnNoPasswordAuthErrorSqlite(t *testing.T) {
	t.Parallel()

	r, g := newSqliteRepo(t)

	u := user.Model{
		ID:    uuid.New(),
		Name:  "My user name",
		Email: "my-user@email.com",
		// No Password hash.
	}

	err := g.Create(t.Context(), &u)
	if err != nil {
		t.Fatalf("could not insert new user: %v", err)
	}

	userData, err := r.SignIn(t.Context(), &user.SignInData{
		Email:         u.Email,
		ClearPassword: "some random password",
	})
	if userData != nil {
		t.Errorf("expected empty user data: got '%v'", userData)
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

	u := user.Model{
		Name:         "my valid user name",
		Email:        "my-valid@email.com",
		PasswordHash: &actualHash,
	}

	r, g := newSqliteRepo(t)

	err := g.Create(t.Context(), &u)
	if err != nil {
		t.Fatalf("could not create test user: %v", err)
	}

	userData, err := r.SignIn(t.Context(), &user.SignInData{
		Email:         u.Email,
		ClearPassword: otherPassword,
	})
	if userData != nil {
		t.Errorf("expected empty user data: got '%v'", userData)
	}

	if err == nil {
		t.Fatal("expected a non-nil error")
	}

	if !errors.Is(err, user.ErrWrongCredentials) {
		t.Errorf("expected wrong credentials error: got '%v'", err)
	}
}
