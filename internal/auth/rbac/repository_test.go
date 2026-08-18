package rbac_test

import (
	"errors"
	"slices"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/kytnacode/inventure/internal/auth/rbac"
	"github.com/kytnacode/inventure/pkg/validation"
	"gorm.io/datatypes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newRepo(t *testing.T) (*rbac.Repository, gorm.Interface[rbac.Model]) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"))
	if err != nil {
		t.Fatalf("could not open database: %v", err)
	}

	err = db.AutoMigrate(rbac.Model{})
	if err != nil {
		t.Fatalf("could not run test database migrations: %v", err)
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

func TestRepository_GetRoleByIDShouldReturnRecordNotFound(t *testing.T) {
	t.Parallel()

	r, _ := newRepo(t)

	role, err := r.GetRoleByID(t.Context(), datatypes.NewBinUUIDv4().String())
	if role != nil {
		t.Errorf("expected a nil role: got '%v'", role)
	}

	if err == nil {
		t.Fatal("expected a non-nil error")
	}

	if !errors.Is(err, rbac.ErrRoleNotFound) {
		t.Errorf("expected error to be '%T': got '%v'", rbac.ErrRoleNotFound, err)
	}
}

func TestRepository_GetRoleByIDShouldReturnRole(t *testing.T) {
	t.Parallel()

	r, g := newRepo(t)

	m := &rbac.Model{
		ID:     datatypes.NewUUIDv4(),
		Name:   "realrole",
		Scopes: rbac.ScopeString([]string{"user-add", "user-del", "user-read", "item-read"}),
	}

	err := g.Create(t.Context(), m)
	if err != nil {
		t.Fatalf("could not create test user: %v", err)
	}

	role, err := r.GetRoleByID(t.Context(), m.ID.String())
	if err != nil {
		t.Errorf("expected a non-nil error: %v", err)
	}

	if role == nil {
		t.Fatal("expected a non-nil role")
	}

	expected := m.ToDomain()

	if got := role.Name; got != expected.Name {
		t.Errorf("expected name to be '%v': got '%v'", expected.Name, got)
	}

	if got := role.Scopes; !slices.Equal(got, expected.Scopes) {
		t.Errorf("expected scopes to be '%v': got '%v'", expected.Scopes, got)
	}
}
