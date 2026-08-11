package repository_test

import (
	"path"
	"testing"

	"github.com/kytnacode/inventure/internal/user/repository"
	"github.com/kytnacode/inventure/pkg/validation"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newSqliteRepo(t *testing.T) (*repository.Repository, gorm.Interface[repository.User]) {
	t.Helper()

	dbPath := path.Join(t.TempDir(), "test.db")

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("could not open database: %v", err)
	}

	if err := db.AutoMigrate(repository.User{}); err != nil {
		t.Fatalf("could not migrate user's schema: %v", err)
	}

	table := gorm.G[repository.User](db)

	repo := repository.New(table, validation.New())

	return repo, table
}

func TestRepositorySqliteShouldCreateUser(t *testing.T) {
	t.Parallel()

	//nolint:gosec // fake credentials.
	data := repository.SignUpUserData{
		Name:         "my user name",
		Email:        "my-user@email.com",
		PasswordHash: "$argon2id$v=19$m=65536,t=3,p=4$uiOyD7yhvKuJwK2B+mYX9w$ZwOB3SQDZWgSI17gaRUJ5Slzr5SH8XErpN/ihlacVR8",
	}

	repo, table := newSqliteRepo(t)

	id, err := repo.SignUpUser(t.Context(), &data)
	if err != nil {
		t.Fatalf("could not create user: %v", err)
	}

	user, err := table.Where("id = ?", id).Take(t.Context())
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
