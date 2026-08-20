package role_test

import (
	"errors"
	"slices"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/kytnacode/inventure/internal/role"
	"github.com/kytnacode/inventure/pkg/validation"
	"gorm.io/datatypes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newRepo(t *testing.T) (*role.Repository, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"))
	if err != nil {
		t.Fatalf("could not open database: %v", err)
	}

	err = db.AutoMigrate(role.RoleModel{})
	if err != nil {
		t.Fatalf("could not run test database migrations: %v", err)
	}

	table := gorm.G[role.RoleModel](db)

	repo := role.NewRepository(table, validation.New())

	return repo, db
}

func TestRepository_CreateRoleShouldRequireName(t *testing.T) {
	t.Parallel()

	repo, _ := newRepo(t)

	id, err := repo.CreateRole(t.Context(), &role.CreateRoleData{
		Name:   "",
		Allow:  []string{"user_add", "user_del"},
		Forbid: []string{"item-read"},
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

	roleData, err := r.GetRoleByID(t.Context(), datatypes.NewBinUUIDv4().String())
	if roleData != nil {
		t.Errorf("expected a nil role: got '%v'", roleData)
	}

	if err == nil {
		t.Fatal("expected a non-nil error")
	}

	if !errors.Is(err, role.ErrRoleNotFound) {
		t.Errorf("expected error to be '%T': got '%v'", role.ErrRoleNotFound, err)
	}
}

func TestRepository_GetRoleByIDShouldReturnRole(t *testing.T) {
	t.Parallel()

	r, db := newRepo(t)

	g := gorm.G[role.RoleModel](db)

	m := &role.RoleModel{
		ID:     datatypes.NewUUIDv4(),
		Name:   "realrole",
		Allow:  role.ScopeString([]string{"user-add", "user-del", "user-read", "item-read"}),
		Forbid: role.ScopeString([]string{"item-del"}),
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

	if got := role.Allow; !slices.Equal(got, expected.Allow) {
		t.Errorf("expected allowed permissions to be '%v': got '%v'", expected.Allow, got)
	}

	if got := role.Forbid; !slices.Equal(got, expected.Forbid) {
		t.Errorf("expected forbidden permissions to be '%v': got '%v'", expected.Forbid, got)
	}
}

func TestRepository_GetRolesByIDsShouldReturnRoleNotFound(t *testing.T) {
	t.Parallel()

	r, _ := newRepo(t)

	const missingID = "real-id"

	roleData, err := r.GetRolesByIDs(t.Context(), []string{missingID})
	if len(roleData) != 0 {
		t.Fatal("expected a empty role list: got list of length: '%'", len(roleData))
	}

	if err == nil {
		t.Fatal("expected a non-nil error")
	}

	if !errors.Is(err, role.ErrRoleNotFound) {
		t.Errorf("expected error to be '%T': got '%v'", role.ErrRoleNotFound, err)
	}
}

func TestRepository_GetRolesByIDsShouldReturnRoles(t *testing.T) {
	t.Parallel()

	r, db := newRepo(t)

	roles := []role.RoleModel{
		{
			ID:     datatypes.NewUUIDv4(),
			Name:   "realrole",
			Allow:  role.ScopeString([]string{"user-add", "user-del", "user-read", "item-read"}),
			Forbid: role.ScopeString([]string{"item-del"}),
		},
		{
			ID:     datatypes.NewUUIDv4(),
			Name:   "other-real-role",
			Allow:  role.ScopeString([]string{"user-add", "user-del", "user-read", "item-read"}),
			Forbid: role.ScopeString([]string{"item-del"}),
		},
	}

	ids := make([]string, 0, len(roles))

	for _, v := range roles {
		ids = append(ids, v.ID.String())
	}

	db.Create(&roles)

	gotRoles, err := r.GetRolesByIDs(t.Context(), ids)
	if err != nil {
		t.Errorf("expected a non-nil error: %v", err)
	}

	if gotRoles == nil {
		t.Fatal("expected a non-nil role")
	}

	expected := make([]role.Role, 0, len(roles))

	for _, v := range roles {
		expected = append(expected, *v.ToDomain())
	}

	for _, got := range gotRoles {
		var expectedRole *role.Role

		for _, exp := range expected {
			if exp.ID == got.ID {
				expectedRole = &exp

				break
			}
		}

		if expectedRole == nil {
			t.Fatalf("role not found with ID: %v", got.ID)
		}

		if got.Name != expectedRole.Name {
			t.Errorf("expected name to be '%v': got '%v'", expectedRole.Name, got.Name)
		}

		if !slices.Equal(got.Allow, expectedRole.Allow) {
			t.Errorf("expected allowed permissions to be '%v': got '%v'", expectedRole.Allow, got.Allow)
		}

		if !slices.Equal(got.Forbid, expectedRole.Forbid) {
			t.Errorf("expected forbidden permissions to be '%v': got '%v'", expectedRole.Forbid, got.Forbid)
		}
	}
}
