package database

import (
	"fmt"

	"github.com/kytnacode/inventure/internal/auth/rbac"
	"github.com/kytnacode/inventure/internal/retail"
	"github.com/kytnacode/inventure/internal/user"
	"gorm.io/gorm"
)

// RunMigrations run database migrations.
func RunMigrations(db *gorm.DB) error {
	err := db.AutoMigrate(
		user.Model{},
		retail.Model{},
		retail.TenantModel{},
		retail.StockItemModel{},
		retail.ItemModel{},
		retail.PlaceModel{},
		rbac.RoleModel{},
		rbac.AccessModel{},
	)
	if err != nil {
		return fmt.Errorf("could not run database migrations: %w", err)
	}

	return nil
}
