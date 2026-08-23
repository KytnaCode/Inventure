package rbac_test

import (
	"slices"
	"testing"

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

	gotResource := rbac.NewResource(got.ResourceType, got.ResourceID.String(), nil)

	if !gotResource.Equal(expectedRole.On) {
		t.Errorf("expected role resource to be '%v': got '%v'", expectedRole.On, gotResource)
	}

	if len(got.Accesses) != len(expectedAccess) {
		t.Errorf("expected role to have '%v' accesses: got '%v'", len(expectedAccess), len(got.Accesses))
	}

	for _, a := range got.Accesses {
		accessRes := rbac.NewResource(a.ResourceType, a.ResourceID.String(), nil)

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

		if !slices.Equal(gotPerms, otherAccess.Perms) {
			t.Errorf("expected access permissions to be '%v': got '%v'", otherAccess.Perms, got)
		}
	}
}

func TestRepository_CreateRoleShouldCreateRole(t *testing.T) {
	t.Parallel()

	repo, db := newRepo(t)

	res := rbac.NewResource("merchant", "308824cd-ba05-4a10-bb11-3e9cdeb8b276", nil)

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
			On:    rbac.NewResource("place", "f77cb426-bc42-4698-8755-815c8445dea6", &res),
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

	res := rbac.NewResource("merchant", "c2c84010-5c5c-43f3-a7fd-bd92cdc5f212", nil)

	subRes := rbac.NewResource("place", "f4f4b9a9-7345-41c7-87b6-64763637df50", &res)

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
