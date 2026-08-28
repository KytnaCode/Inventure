package rbac

import (
	"github.com/google/uuid"
	"github.com/kytnacode/inventure/sqltypes"
	"gorm.io/gorm"
)

// RoleModel is the database representation for a role.
type RoleModel struct {
	gorm.Model

	// ID is role's unique ID.
	ID uuid.UUID `gorm:"primaryKey"`

	// Name is role's name, must be locally unique within the resource the role belongs.
	Name string `gorm:"uniqueIndex:idx_name,where:deleted_at IS NULL,priority:1"`

	// Accesses are the permission grants for a role.
	Accesses []AccessModel `gorm:"foreignKey:RoleID"`

	// ResourceType is the type of the resource the role belongs to, role permission are not
	// applicable to resources that are not the one the role is declared or a child of it.
	ResourceType string `gorm:"uniqueIndex:idx_name,where:deleted_at IS NULL,priority:3"`

	// ResourceID is the ID of the resource the role belongs to, role permission are not applicable
	// to resources that are not the one the role is declared or a child of it.
	ResourceID uuid.UUID `gorm:"uniqueIndex:idx_name,where:deleted_at IS NULL,priority:2"`
}

// TableName return the name for roles table. Implements [gorm/schema.Tabler].
func (m *RoleModel) TableName() string {
	return "roles"
}

// AccessModel is the database representation of a permission grant.
type AccessModel struct {
	gorm.Model

	// ID is access' unique ID.
	ID uuid.UUID `gorm:"primaryKey"`

	// RoleID is the role which the access is granted on.
	RoleID uuid.UUID `gorm:"uniqueIndex:idx_resource,where:deleted_at IS NULL,priority:2"`

	// Perms is the list of permissions granted to the role.
	Perms sqltypes.List

	// ResourceType is the type of the resource the access is granting permissions on.
	ResourceType string `gorm:"uniqueIndex:idx_resource,where:deleted_at IS NULL,priority:3"`

	// ResourceID is the ID of the resource the access is granting permissions on.
	ResourceID uuid.UUID `gorm:"uniqueIndex:idx_resource,where:deleted_at IS NULL,priority:1"`
}

// TableName returns the name for accesses table. Implements [gorm/schema.Tabler].
func (m *AccessModel) TableName() string {
	return "accesses"
}

func convertList[S, T ~string](s []S) []T {
	target := make([]T, 0, len(s))

	for _, v := range s {
		target = append(target, T(v))
	}

	return target
}

// sliceToPerm converts a slice of string to a slices of [Perm].
func sliceToPerm(l []string) []Perm {
	return convertList[string, Perm](l)
}

// permToSlice converts a slice of [Perm] to a slice of strings.
func permToSlice(perms []Perm) []string {
	return convertList[Perm, string](perms)
}
