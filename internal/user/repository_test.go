package user_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/kytnacode/inventure/internal/auth/rbac"
	"github.com/kytnacode/inventure/internal/testutil"
	"github.com/kytnacode/inventure/internal/testutil/dbtest"
	"github.com/kytnacode/inventure/internal/user"
	"gorm.io/gorm"
)

func runMigrations(db *gorm.DB) error {
	return db.AutoMigrate(user.Model{})
}

func newRepo(db *gorm.DB) *user.Repository {
	return user.NewRepository(db)
}

func TestRepositorySqlite_CreateUserShouldCreateUser(t *testing.T) {
	testutil.Integration(t)

	t.Parallel()

	dbtest.RunWithDatabases(t, runMigrations, func(t *testing.T, db *gorm.DB) {
		t.Parallel()

		repo := newRepo(db)

		data := user.Data{
			Name:  "my user name",
			Email: "my-user@email.com",
			PasswordHash: new(
				"$argon2id$v=19$m=65536,t=3,p=4$uiOyD7yhvKuJwK2B+mYX9w$ZwOB3SQDZWgSI17gaRUJ5Slzr5SH8XErpN/ihlacVR8",
			),
		}

		id, err := repo.CreateUser(t.Context(), &data)
		if err != nil {
			t.Fatalf("could not create user: %v", err)
		}

		var user user.Model

		err = db.WithContext(t.Context()).Where("id = ?", id).Take(&user).Error
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

		if *user.PasswordHash != *data.PasswordHash {
			t.Errorf(
				"expected password hash to be '%v': got '%v'",
				data.PasswordHash,
				user.PasswordHash,
			)
		}
	})
}

func TestRepositorySqlite_UserByEmailShouldReturnUserNotFound(t *testing.T) {
	testutil.Integration(t)

	t.Parallel()

	dbtest.RunWithDatabases(t, runMigrations, func(t *testing.T, db *gorm.DB) {
		t.Parallel()

		r := newRepo(db)

		u, err := r.UserByEmail(t.Context(), "super-real@email.com")
		if u != nil {
			t.Errorf("expected empty user data: got '%v'", u)
		}

		if err == nil {
			t.Fatal("exected a non-nil error")
		}

		if !errors.Is(err, user.ErrUserNotFound) {
			t.Fatalf("expected user not found error: got '%v'", err)
		}
	})
}

func TestRepositorySqlite_AssingRolesShouldAssingRoles(t *testing.T) {
	testutil.Integration(t)

	t.Parallel()

	dbtest.RunWithDatabases(t, runMigrations, func(t *testing.T, db *gorm.DB) {
		t.Parallel()

		repo := newRepo(db)

		const (
			expectedRolesNum = 1
			expectedRoleName = "test"
		)

		roleRepo := rbac.NewRepository(db)

		res := rbac.NewResource("tenant", uuid.New(), nil)

		roleID, err := rbac.RoleBuilder(roleRepo).
			Name(expectedRoleName).BelongsTo(res).Build(t.Context())
		if err != nil {
			t.Fatalf("could not create test role: %v", err)
		}

		id, err := repo.CreateUser(t.Context(), &user.Data{
			Name:  "akko",
			Email: "akko@lunanova.edu",
			PasswordHash: new(
				"$argon2id$v=19$m=65536,t=3,p=4$uiOyD7yhvKuJwK2B+mYX9w$ZwOB3SQDZWgSI17gaRUJ5Slzr5SH8XErpN/ihlacVR8",
			),
		})
		if err != nil {
			t.Fatalf("could not create test user: %v", err)
		}

		err = repo.AssingRoles(t.Context(), id, roleID)
		if err != nil {
			t.Fatalf("could not assing roles: %v", err)
		}

		var got user.Model

		err = db.WithContext(t.Context()).Where("id = ?", id).Preload("Roles").Take(&got).Error
		if err != nil {
			t.Fatalf("could not get updated user: %v", err)
		}

		if len(got.Roles) != expectedRolesNum {
			t.Fatalf("expected %v roles: got %v", expectedRolesNum, len(got.Roles))
		}

		if got.Roles[0].Name != expectedRoleName {
			t.Errorf("expected role name to be '%v': got '%v'", expectedRoleName, got.Roles[0].Name)
		}
	})
}

func TestRepositorySqlite_UpdateByEmailShouldReturnUserNotFound(t *testing.T) {
	testutil.Integration(t)

	t.Parallel()

	dbtest.RunWithDatabases(t, runMigrations, func(t *testing.T, db *gorm.DB) {
		t.Parallel()

		repo := newRepo(db)

		err := repo.UpdateByEmail(t.Context(), "non-existing", func(u *user.Model) error {
			u.Name = "hello " + u.Name

			return nil
		})
		if err == nil {
			t.Fatal("expected a non-nil error")
		}

		if !errors.Is(err, user.ErrUserNotFound) {
			t.Errorf("expected error to be UserNotFound: %v", err)
		}
	})
}

func TestRepositorySqlite_UpdateByEmailShouldUpdateUser(t *testing.T) {
	testutil.Integration(t)

	t.Parallel()

	dbtest.RunWithDatabases(t, runMigrations, func(t *testing.T, db *gorm.DB) {
		t.Parallel()

		repo := newRepo(db)

		const email = "diana.cavendish@lunanova.edu"

		data := &user.Data{
			Name:  "Diana Cavendish",
			Email: email,
			PasswordHash: new(
				"$argon2id$v=19$m=65536,t=3,p=4$uiOyD7yhvKuJwK2B+mYX9w$ZwOB3SQDZWgSI17gaRUJ5Slzr5SH8XErpN/ihlacVR8",
			),
		}

		extraPart := " Cavendish"
		expectedName := data.Name + extraPart

		_, err := repo.CreateUser(t.Context(), data)
		if err != nil {
			t.Fatalf("could not create test user: %v", err)
		}

		err = repo.UpdateByEmail(t.Context(), email, func(u *user.Model) error {
			u.Name += extraPart

			return nil
		})
		if err != nil {
			t.Fatalf("expected a nil error: %v", err)
		}

		u, err := repo.UserByEmail(t.Context(), email)
		if err != nil {
			t.Fatalf("could not get updated user: %v", err)
		}

		if u.Name != expectedName {
			t.Errorf("expected name to be '%v': got '%v'", expectedName, u.Name)
		}

		if u.PasswordHash == nil || data.PasswordHash == nil {
			t.Fatalf(
				"expected non-nil password hash field: user: '%v' data: '%v'",
				u.PasswordHash,
				data.PasswordHash,
			)
		}

		if *u.PasswordHash != *data.PasswordHash {
			t.Errorf(
				"expected unchanged pasword field to be '%v': got '%v'",
				*u.PasswordHash,
				*data.PasswordHash,
			)
		}
	})
}

func TestRepositorySqlite_UpdateByEmailShouldWrapUpdateFnError(t *testing.T) {
	testutil.Integration(t)

	t.Parallel()

	dbtest.RunWithDatabases(t, runMigrations, func(t *testing.T, db *gorm.DB) {
		t.Parallel()

		repo := newRepo(db)

		ErrExpected := errors.New("expected")

		const email = "diana.cavendish@lunanova.edu"

		_, err := repo.CreateUser(t.Context(), &user.Data{
			Name:  "Diana Cavendish",
			Email: email,
			PasswordHash: new(
				"$argon2id$v=19$m=65536,t=3,p=4$uiOyD7yhvKuJwK2B+mYX9w$ZwOB3SQDZWgSI17gaRUJ5Slzr5SH8XErpN/ihlacVR8",
			),
		})
		if err != nil {
			t.Fatalf("could not create test user: %v", err)
		}

		err = repo.UpdateByEmail(t.Context(), email, func(_ *user.Model) error {
			return ErrExpected
		})
		if err == nil {
			t.Fatal("expected a non-nil error")
		}

		if !errors.Is(err, ErrExpected) {
			t.Errorf("expected error to be '%v': got '%v'", ErrExpected, err)
		}
	})
}
