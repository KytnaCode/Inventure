package rbac

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"

	"github.com/google/uuid"
	"github.com/kytnacode/inventure/sqltypes"
	"gorm.io/gorm"
)

// Ensure [Repository] implements [roleGetter].
var _ roleGetter = &Repository{}

// RoleData contains necessary data to create a new role with [Repository].
type RoleData struct {
	// Name is role's name.
	Name string

	// On is the resource which the role is scoped on.
	On Resource
}

// AccessData is the necessary data to create a new access with [Repository].
type AccessData struct {
	// Perms are the permissions granted.
	Perms []Perm

	// On is the resource on which the permissions are being granted.
	On Resource
}

// Repository handles role and access business logic.
//
// Safe for concurrent use.
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new [Repository].
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{
		db: db,
	}
}

// GetRoles gets a list of roles by theirs IDs, returned roles may, and possibly will not match
// IDs order. Non-existing role IDs will be dropped without error, so, returned roles slice length
// will be less or equal to the number of IDs passed.
func (r *Repository) GetRoles(ctx context.Context, ids ...uuid.UUID) ([]Role, error) {
	var models []RoleModel

	err := r.db.WithContext(ctx).
		Where("id IN ?", ids).
		Preload("Accesses").
		Find(&models).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return []Role{}, nil
	}

	if err != nil {
		return nil, fmt.Errorf("could not get roles: %w", err)
	}

	roles := make([]Role, 0, len(models))

	for _, m := range models {
		roles = append(roles, *m.ToDomain())
	}

	return roles, nil
}

// CreateRole creates a new Role with the given accesses and returns role's ID, if role already
// exists accesses will be appended to existing role object.
func (r *Repository) CreateRole(
	ctx context.Context,
	data *RoleData,
	accesses ...AccessData,
) (uuid.UUID, error) {
	var newID uuid.UUID

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		roleID, err := r.getOrCreateRole(tx, data)
		if err != nil {
			return err
		}

		for _, a := range accesses {
			err := r.createAccess(tx, roleID, &a)
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
func (r *Repository) getOrCreateRole(tx *gorm.DB, data *RoleData) (id uuid.UUID, err error) {
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
func (r *Repository) createAccess(
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
			return r.newAccess(tx, roleID, access)
		}

		return fmt.Errorf("could not get access: %w", err)
	}

	return r.updateAccess(tx, &m, access)
}

// updateAccess updates access permissions, automatically removes repeated ones.
func (r *Repository) updateAccess(tx *gorm.DB, access *AccessModel, data *AccessData) error {
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
func (r *Repository) newAccess(tx *gorm.DB, roleID uuid.UUID, access *AccessData) error {
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
