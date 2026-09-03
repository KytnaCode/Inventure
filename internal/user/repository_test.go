package user_test

import (
	"errors"
	"path"
	"testing"

	"github.com/google/uuid"
	"github.com/kytnacode/inventure/internal/auth/rbac"
	"github.com/kytnacode/inventure/internal/user"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newSqliteRepo(t *testing.T) (*user.Repository, *gorm.DB) {
	t.Helper()

	dbPath := path.Join(t.TempDir(), "test.db")

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("could not open database: %v", err)
	}

	if err := db.AutoMigrate(user.Model{}); err != nil {
		t.Fatalf("could not migrate user's schema: %v", err)
	}

	repo := user.NewRepository(db)

	return repo, db
}

func TestRepositorySqlite_CreateUserShouldCreateUser(t *testing.T) {
	t.Parallel()

	data := user.Data{
		Name:  "my user name",
		Email: "my-user@email.com",
		PasswordHash: new(
			"$argon2id$v=19$m=65536,t=3,p=4$uiOyD7yhvKuJwK2B+mYX9w$ZwOB3SQDZWgSI17gaRUJ5Slzr5SH8XErpN/ihlacVR8",
		),
	}

	repo, db := newSqliteRepo(t)

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
}

func TestRepositorySqlite_UserByEmailShouldReturnUserNotFound(t *testing.T) {
	t.Parallel()

	r, _ := newSqliteRepo(t)

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
}

func TestRepositorySqlite_AssingRolesShouldAssingRoles(t *testing.T) {
	t.Parallel()

	const (
		expectedRolesNum = 1
		expectedRoleName = "test"
	)

	repo, db := newSqliteRepo(t)

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
}
