package rbac

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
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
	db  *gorm.DB
	dao *DAO
}

// NewRepository creates a new [Repository].
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{
		db:  db,
		dao: &DAO{},
	}
}

// CreateRole creates a new Role with the given accesses and returns role's ID, if role already
// exists accesses will be appended to existing role object.
func (r *Repository) CreateRole(
	ctx context.Context,
	roleData *RoleData,
	accesses ...AccessData,
) (id uuid.UUID, err error) {
	return r.dao.CreateRole(r.db.WithContext(ctx), roleData, accesses...)
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
