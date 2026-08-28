package rbac_test

import (
	"cmp"
	"slices"
	"testing"

	"github.com/google/uuid"
	"github.com/kytnacode/inventure/internal/auth/rbac"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newRepo(t *testing.T) (*rbac.Repository, *gorm.DB) {
	db, err := gorm.Open(sqlite.Open(":memory:"))
	if err != nil {
		t.Fatalf("could not open sqlite3 instance: %v", err)
	}

	err = db.AutoMigrate(&rbac.RoleModel{}, &rbac.AccessModel{})
	if err != nil {
		t.Fatalf("could not run test database migrations: %v", err)
	}

	repo := rbac.NewRepository(db)

	return repo, db
}

func compareModelAndData(
	t *testing.T,
	got *rbac.RoleModel,
	expectedRole *rbac.RoleData,
	expectedAccess []rbac.AccessData,
) {
	if got.Name != expectedRole.Name {
		t.Errorf("expected role name to be '%v': got '%v'", expectedRole.Name, got.Name)
	}

	gotResource := rbac.NewResource(got.ResourceType, got.ResourceID, nil)

	if !gotResource.Equal(expectedRole.On) {
		t.Errorf("expected role resource to be '%v': got '%v'", expectedRole.On, gotResource)
	}

	if len(got.Accesses) != len(expectedAccess) {
		t.Errorf("expected role to have '%v' accesses: got '%v'", len(expectedAccess), len(got.Accesses))
	}

	for _, a := range got.Accesses {
		accessRes := rbac.NewResource(a.ResourceType, a.ResourceID, nil)

		var otherAccess *rbac.AccessData

		for _, accessData := range expectedAccess {
			if accessRes.Equal(accessData.On) {
				otherAccess = &accessData

				break
			}
		}

		if otherAccess == nil {
			t.Fatalf("not found access on resource '%v'", accessRes)
		}

		gotPerms := make([]rbac.Perm, 0, len(a.Perms))

		for _, v := range a.Perms {
			gotPerms = append(gotPerms, rbac.Perm(v))
		}

		// Ensure sorted to avoid false positives.
		slices.Sort(gotPerms)
		slices.Sort(otherAccess.Perms)

		if !slices.Equal(gotPerms, otherAccess.Perms) {
			t.Errorf("expected access permissions to be '%v': got '%v'", otherAccess.Perms, gotPerms)
		}
	}
}

func compareDomainAndData(
	t *testing.T,
	got *rbac.Role,
	expectedRole *rbac.RoleData,
	expectedAccess []rbac.AccessData,
) {
	if got.Name != expectedRole.Name {
		t.Errorf("expected role name to be '%v': got '%v'", expectedRole.Name, got.Name)
	}

	if !got.On.Equal(expectedRole.On) {
		t.Errorf("expected role resource to be '%v': got '%v'", expectedRole.On, got.On)
	}

	if len(got.Accesses) != len(expectedAccess) {
		t.Errorf("expected role to have '%v' accesses: got '%v'", len(expectedAccess), len(got.Accesses))
	}

	for _, a := range got.Accesses {
		var otherAccess *rbac.AccessData

		for _, accessData := range expectedAccess {
			if a.On.Equal(accessData.On) {
				otherAccess = &accessData

				break
			}
		}

		if otherAccess == nil {
			t.Fatalf("not found access on resource '%v'", a.On)
		}

		// Ensure sorted to avoid false positives.
		slices.Sort(a.Perms)
		slices.Sort(otherAccess.Perms)

		if !slices.Equal(a.Perms, otherAccess.Perms) {
			t.Errorf("expected access permissions to be '%v': got '%v'", otherAccess.Perms, a.Perms)
		}
	}
}

func TestRepository_CreateRoleShouldCreateRole(t *testing.T) {
	t.Parallel()

	repo, db := newRepo(t)

	res := rbac.NewResource("merchant", uuid.New(), nil)

	roleData := rbac.RoleData{
		Name: "admin",
		On:   res,
	}

	accesses := []rbac.AccessData{
		{
			Perms: []rbac.Perm{"user-read"},
			On:    res,
		},
		{
			Perms: []rbac.Perm{"item-read", "item-del", "item-add"},
			On:    rbac.NewResource("place", uuid.New(), &res),
		},
	}

	id, err := repo.CreateRole(t.Context(), &roleData, accesses...)
	if err != nil {
		t.Fatalf("could not create roles: %v", err)
	}

	var got rbac.RoleModel

	err = db.Where("id = ?", id).
		Preload("Accesses").
		Take(&got).Error
	if err != nil {
		t.Fatalf("could not get role from database: %v", err)
	}

	compareModelAndData(t, &got, &roleData, accesses)
}

func TestRepository_CreateRoleShouldAppendToExistingRole(t *testing.T) {
	t.Parallel()

	repo, db := newRepo(t)

	res := rbac.NewResource("merchant", uuid.New(), nil)

	subRes := rbac.NewResource("place", uuid.New(), &res)

	roleData := &rbac.RoleData{
		Name: "admin",
		On:   res,
	}

	baseAccesses := []rbac.AccessData{
		{
			On:    res,
			Perms: []rbac.Perm{"user-add"},
		},
	}

	newAccesses := []rbac.AccessData{
		{
			On:    res,
			Perms: []rbac.Perm{"user-del"},
		},
		{
			On:    subRes,
			Perms: []rbac.Perm{"item-read"},
		},
	}

	expectedAccesses := []rbac.AccessData{
		{
			On:    res,
			Perms: []rbac.Perm{"user-add", "user-del"},
		},
		{
			On:    subRes,
			Perms: []rbac.Perm{"item-read"},
		},
	}

	_, err := repo.CreateRole(t.Context(), roleData, baseAccesses...)
	if err != nil {
		t.Fatalf("could not create role: %v", err)
	}

	id, err := repo.CreateRole(t.Context(), roleData, newAccesses...)
	if err != nil {
		t.Fatalf("could not append to existing role: %v", err)
	}

	var got rbac.RoleModel

	err = db.Where("id = ?", id).
		Preload("Accesses").
		Take(&got).Error
	if err != nil {
		t.Fatalf("could not get inserted role: %v", err)
	}

	compareModelAndData(t, &got, roleData, expectedAccesses)
}

func TestRepository_GetRolesShouldReturnAllRoles(t *testing.T) {
	t.Parallel()

	repo, _ := newRepo(t)

	res := rbac.NewResource("merchant", uuid.New(), nil)

	firstRole := &rbac.RoleData{
		Name: "admin",
		On:   res,
	}

	secondRole := &rbac.RoleData{
		Name: "moderator",
		On:   res,
	}

	accesses := map[string][]rbac.AccessData{
		firstRole.Name: {
			{
				On:    res,
				Perms: []rbac.Perm{"user-add", "user-read", "user-del"},
			},
		},
		secondRole.Name: {
			{
				On:    res,
				Perms: []rbac.Perm{"user-add", "user-read"},
			},
		},
	}

	firstID, err := repo.CreateRole(t.Context(), firstRole, accesses[firstRole.Name]...)
	if err != nil {
		t.Fatalf("could not create role: %v", err)
	}

	secondID, err := repo.CreateRole(t.Context(), secondRole, accesses[secondRole.Name]...)
	if err != nil {
		t.Fatalf("could not append to existing role: %v", err)
	}

	gotRoles, err := repo.GetRoles(t.Context(), firstID, secondID)
	if err != nil {
		t.Fatalf("expected a nil error: %v", err)
	}

	slices.SortFunc(gotRoles, func(a, b rbac.Role) int {
		return cmp.Compare(a.Name, b.Name)
	})

	expectedRoles := []*rbac.RoleData{firstRole, secondRole}

	slices.SortFunc(expectedRoles, func(a, b *rbac.RoleData) int {
		return cmp.Compare(a.Name, b.Name)
	})

	for i, got := range gotRoles {
		expected := expectedRoles[i]

		compareDomainAndData(t, &got, expected, accesses[expected.Name])
	}
}

func TestRepository_GetRolesShouldReturnOnlyExistingRoles(t *testing.T) {
	t.Parallel()

	repo, _ := newRepo(t)

	res := rbac.NewResource("merchant", uuid.New(), nil)

	firstRole := &rbac.RoleData{
		Name: "admin",
		On:   res,
	}

	secondRole := &rbac.RoleData{
		Name: "moderator",
		On:   res,
	}

	accesses := map[string][]rbac.AccessData{
		firstRole.Name: {
			{
				On:    res,
				Perms: []rbac.Perm{"user-add", "user-read", "user-del"},
			},
		},
		secondRole.Name: {
			{
				On:    res,
				Perms: []rbac.Perm{"user-add", "user-read"},
			},
		},
	}

	firstID, err := repo.CreateRole(t.Context(), firstRole, accesses[firstRole.Name]...)
	if err != nil {
		t.Fatalf("could not create role: %v", err)
	}

	secondID, err := repo.CreateRole(t.Context(), secondRole, accesses[secondRole.Name]...)
	if err != nil {
		t.Fatalf("could not append to existing role: %v", err)
	}

	gotRoles, err := repo.GetRoles(
		t.Context(),
		uuid.New(), // Non-existing.
		firstID,
		uuid.New(), // Non-existing.
		secondID,
		uuid.New(), // Non-existing.
		uuid.New(), // Non-existing.
	)
	if err != nil {
		t.Fatalf("expected a nil error: %v", err)
	}

	slices.SortFunc(gotRoles, func(a, b rbac.Role) int {
		return cmp.Compare(a.Name, b.Name)
	})

	expectedRoles := []*rbac.RoleData{firstRole, secondRole}

	slices.SortFunc(expectedRoles, func(a, b *rbac.RoleData) int {
		return cmp.Compare(a.Name, b.Name)
	})

	for i, got := range gotRoles {
		expected := expectedRoles[i]

		compareDomainAndData(t, &got, expected, accesses[expected.Name])
	}
}

func TestRepository_GetRolesShouldReturnEmptyListWhenAllRolesNotExists(t *testing.T) {
	t.Parallel()

	repo, _ := newRepo(t)

	gotRoles, err := repo.GetRoles(
		t.Context(),
		uuid.New(), // Non-existing.
		uuid.New(), // Non-existing.
		uuid.New(), // Non-existing.
	)
	if err != nil {
		t.Fatalf("expected a nil error: %v", err)
	}

	if len(gotRoles) != 0 {
		t.Fatalf("expected roles to be an empty slice: got '%v'", gotRoles)
	}
}
