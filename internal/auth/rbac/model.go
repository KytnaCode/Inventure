package rbac

import (
	"github.com/kytnacode/inventure/sqltypes"
	"gorm.io/datatypes"
)

// RoleModel is the database representation for a role.
type RoleModel struct {
	// ID is role's unique ID.
	ID datatypes.UUID `gorm:"primaryKey"`

	// Name is role's name, must be locally unique within the resource the role belongs.
	Name string `gorm:"uniqueIndex:idx_name,where:deleted_at IS NULL,priority:1"`

	// Accesses are the permission grants for a role.
	Accesses []AccessModel `gorm:"foreignKey:RoleID"`

	// ResourceType is the type of the resource the role belongs to, role permission are not
	// applicable to resources that are not the one the role is declared or a child of it.
	ResourceType string `gorm:"uniqueIndex:idx_name,where:deleted_at IS NULL,priority:3"`

	// ResourceID is the ID of the resource the role belongs to, role permission are not applicable
	// to resources that are not the one the role is declared or a child of it.
	ResourceID string `gorm:"uniqueIndex:idx_name,where:deleted_at IS NULL,priority:2"`
}

// TableName return the name for roles table. Implements [gorm/schema.Tabler].
func (m *RoleModel) TableName() string {
	return "roles"
}

// AccessModel is the database representation of a permission grant.
type AccessModel struct {
	// ID is access' unique ID.
	ID datatypes.UUID `gorm:"primaryKey"`

	// RoleID is the role which the access is granted on.
	RoleID datatypes.UUID

	// Perms is the list of permissions granted to the role.
	Perms sqltypes.List

	// ResourceType is the type of the resource the access is granting permissions on.
	ResourceType string `gorm:"uniqueIndex:idx_resource,where:deleted_at IS NULL,priority:2"`

	// ResourceID is the ID of the resource the access is granting permissions on.
	ResourceID datatypes.UUID `gorm:"uniqueIndex:idx_resource,where:deleted_at IS NULL,priority:1"`
}

// TableName returns the name for accesses table. Implements [gorm/schema.Tabler].
func (m *AccessModel) TableName() string {
	return "accesses"
}
