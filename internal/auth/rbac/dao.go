package rbac

import (
	"errors"
	"fmt"
	"maps"
	"slices"

	"github.com/google/uuid"
	"github.com/kytnacode/inventure/sqltypes"
	"gorm.io/gorm"
)

// DAO handles low-level persistence layer role and accesses operations.
type DAO struct{}

// NewDAO creates a new [DAO].
func NewDAO() *DAO {
	return &DAO{}
}

// CreateRole creates a new Role with the given accesses and returns role's ID, if role already
// exists accesses will be appended to existing role object.
func (d *DAO) CreateRole(
	tx *gorm.DB,
	data *RoleData,
	accesses ...AccessData,
) (uuid.UUID, error) {
	var newID uuid.UUID

	err := tx.Transaction(func(tx *gorm.DB) error {
		roleID, err := d.getOrCreateRole(tx, data)
		if err != nil {
			return err
		}

		for _, a := range accesses {
			err := d.createAccess(tx, roleID, &a)
			if err != nil {
				return err
			}
		}

		newID = roleID

		return nil
	})
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("could not create role: %w", err)
	}

	return newID, nil
}

// getOrCreateRole get a role's ID if already exists, if not exists it will create it.
func (d *DAO) getOrCreateRole(tx *gorm.DB, data *RoleData) (id uuid.UUID, err error) {
	var roleID RoleModel

	err = tx.Model(&RoleModel{}).
		Where(&RoleModel{
			Name:         data.Name,
			ResourceType: data.On.Typ,
			ResourceID:   data.On.ID,
		}).
		Attrs(&RoleModel{
			ID:           uuid.New(),
			Name:         data.Name,
			ResourceType: data.On.Typ,
			ResourceID:   data.On.ID,
		}).
		FirstOrCreate(&roleID).Error
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("could not create role: %w", err)
	}

	return roleID.ID, nil
}

// createAccess creates a new access if not exists, or if it already exists, update it.
func (d *DAO) createAccess(
	tx *gorm.DB,
	roleID uuid.UUID,
	access *AccessData,
) error {
	var m AccessModel

	err := tx.Model(&AccessModel{}).
		Where(&AccessModel{
			RoleID:       roleID,
			ResourceType: access.On.Typ,
			ResourceID:   access.On.ID,
		}).
		Select("id", "perms").
		Take(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return d.newAccess(tx, roleID, access)
		}

		return fmt.Errorf("could not get access: %w", err)
	}

	return d.updateAccess(tx, &m, access)
}

// updateAccess updates access permissions, automatically removes repeated ones.
func (d *DAO) updateAccess(tx *gorm.DB, access *AccessModel, data *AccessData) error {
	currentPerms := sliceToPerm(access.Perms)

	newPerms := permToSlice(union(currentPerms, data.Perms))

	err := tx.Model(&AccessModel{}).
		Where("id = ?", access.ID).
		Update("perms", sqltypes.List(newPerms)).Error
	if err != nil {
		return fmt.Errorf("could not update access: %w", err)
	}

	return nil
}

// newAccess creates a new access on the database.
func (d *DAO) newAccess(tx *gorm.DB, roleID uuid.UUID, access *AccessData) error {
	err := tx.Create(&AccessModel{
		ID:           uuid.New(),
		Perms:        permToSlice(access.Perms),
		ResourceType: access.On.Typ,
		ResourceID:   access.On.ID,
		RoleID:       roleID,
	}).Error
	if err != nil {
		return fmt.Errorf("could not create access: %w", err)
	}

	return nil
}

// union is set union.
func union[E comparable](a, b []E) []E {
	got := make(map[E]struct{}, len(a)+len(b))

	for _, e := range a {
		got[e] = struct{}{}
	}

	for _, e := range b {
		got[e] = struct{}{}
	}

	return slices.Collect(maps.Keys(got))
}
