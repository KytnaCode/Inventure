package rbac_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/kytnacode/inventure/internal/auth/rbac"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newDAO(t *testing.T) (dao *rbac.DAO, db *gorm.DB) {
	db, err := gorm.Open(sqlite.Open(":memory:"))
	if err != nil {
		t.Fatalf("could not open sqlite3 instance: %v", err)
	}

	err = db.AutoMigrate(&rbac.RoleModel{}, &rbac.AccessModel{})
	if err != nil {
		t.Fatalf("could not run test database migrations: %v", err)
	}

	dao = rbac.NewDAO()

	return dao, db
}

func TestDAO_CreateRoleShouldCreateRole(t *testing.T) {
	t.Parallel()

	dao, db := newDAO(t)

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

	id, err := dao.CreateRole(db.WithContext(t.Context()), &roleData, accesses...)
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

func TestDAO_CreateRoleShouldAppendToExistingRole(t *testing.T) {
	t.Parallel()

	dao, db := newDAO(t)

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

	_, err := dao.CreateRole(db.WithContext(t.Context()), roleData, baseAccesses...)
	if err != nil {
		t.Fatalf("could not create role: %v", err)
	}

	id, err := dao.CreateRole(db.WithContext(t.Context()), roleData, newAccesses...)
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
