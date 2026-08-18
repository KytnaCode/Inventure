package rbac_test

import (
	"errors"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/kytnacode/inventure/internal/auth/rbac"
	"github.com/kytnacode/inventure/pkg/validation"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newRepo(t *testing.T) (*rbac.Repository, gorm.Interface[rbac.Model]) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"))
	if err != nil {
		t.Fatalf("could not open database: %v", err)
	}

	table := gorm.G[rbac.Model](db)

	repo := rbac.NewRepository(table, validation.New())

	return repo, table
}

func TestRepository_CreateRoleShouldRequireName(t *testing.T) {
	t.Parallel()

	repo, _ := newRepo(t)

	id, err := repo.CreateRole(t.Context(), &rbac.CreateRoleData{
		Name:   "",
		Scopes: []string{"user_add", "user_del"},
	})
	if id != "" {
		t.Errorf("expected an empty ID: got '%v'", id)
	}

	if err == nil {
		t.Fatal("expected a non-nil error")
	}

	if _, ok := errors.AsType[validator.ValidationErrors](err); !ok {
		t.Fatalf("expected error to be validation errors: got '%v'", err)
	}
}
