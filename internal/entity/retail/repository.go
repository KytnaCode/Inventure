package retail

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/kytnacode/inventure/internal/entity/item"
	"gorm.io/gorm"
)

// Data is used by [Repository] to create a new retail.
type Data struct {
	// Name is retail's display name.
	Name string

	// TenantID is the tenant the retail belongs, required. MUST exists.
	TenantID uuid.UUID

	// Storage is the data for retail's root storage. If missing a default storage will be created.
	Storage *PlaceData
}

// Reference contains a reference to a resource.
type Reference struct {
	// Typ is resource's type.
	Typ string

	// ID is resource' ID.
	ID string
}

// PlaceData is used by [Repository] to create new places.
type PlaceData struct {
	// Name is place's display name.
	Name string

	// Children are other places contained in this one.
	Children []PlaceData

	// Items are items directly contained in this place.
	Items []item.Data
}

// Repository handles business logic for retails and places.
type Repository struct {
	db  *gorm.DB
	v   *validator.Validate
	dao *DAO
}

// NewRepository creates a new [Repository].
func NewRepository(db *gorm.DB, v *validator.Validate) *Repository {
	return &Repository{
		db: db,
		v:  v,
		dao: &DAO{
			v: v,
		},
	}
}

// CreateRetail creates a new retail and returns its ID.
func (r *Repository) CreateRetail(
	ctx context.Context,
	retailData *Data,
) (id string, err error) {
	var newID string

	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		retailID, err := r.dao.CreateRetail(tx, retailData)
		if err != nil {
			return err
		}

		newID = retailID.String()

		return nil
	})
	if err != nil {
		return "", fmt.Errorf("could not create retail: %w", err)
	}

	return newID, nil
}

func itemModelToDomain(models []item.Model) []item.Item {
	items := make([]item.Item, 0, len(models))

	for _, it := range models {
		items = append(items, *it.ToDomain())
	}

	return items
}
