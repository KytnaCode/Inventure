package user

import (
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DAO handles low-level database operations for users.
type DAO struct{}

// NewDAO creates a new [DAO].
func NewDAO() *DAO {
	return &DAO{}
}

// Exists check if a given user correspond to an existing user or not.
func (d *DAO) Exists(tx *gorm.DB, id uuid.UUID) (found bool, err error) {
	err = tx.Model(&Model{}).
		Where("id = ?", id).
		Select("COUNT(*) > 0").
		Take(&found).Error
	if err != nil {
		return false, fmt.Errorf("could not check for user existence: %w", err)
	}

	return found, nil
}
