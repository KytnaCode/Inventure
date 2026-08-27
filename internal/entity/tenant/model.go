package tenant

import (
	"github.com/google/uuid"
	"github.com/kytnacode/inventure/internal/entity/retail"
	"github.com/kytnacode/inventure/internal/entity/user"
	"gorm.io/gorm"
)

// Model is the database representation of a tenant.
type Model struct {
	gorm.Model

	// ID is tenant's unique ID.
	ID uuid.UUID

	// Name is tenant's display name.
	Name string

	// Users are tenant-scoped users.
	Users []user.Model `gorm:"many2many:tenant_users;"`

	// Retails are tenant-owned retails.
	Retails []retail.Model
}

// TableName returns tenant's table name. Implements [gorm/schema.Tabler].
func (m *Model) TableName() string {
	return "tenants"
}
