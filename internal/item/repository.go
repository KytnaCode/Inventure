package item

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Data contains fields necessary to create an item with [Repository].
type Data struct {
	// Name is item's name.
	Name string `validate:"required,max=80,resourcename"`

	// Desc is item's description.
	Desc string `validate:"max=65565,resourcename"`

	// Stock is item's stock, must be positive.
	Stock int `validate:"gte=0"`

	// Attrs are optional per-item custom attributes.
	Attrs map[string]any `validate:"dive,keys,required,min=1,max=80,endkeys"`
}

// Repository handles business logic for [Item].
//
// Safe for concurrent use.
type Repository struct {
	db *gorm.DB
	v  *validator.Validate
}

// NewRepository creates a new [Repository].
func NewRepository(db *gorm.DB, v *validator.Validate) *Repository {
	return &Repository{
		db: db,
		v:  v,
	}
}

// CreateItem creates a new item from data a returns its ID. If data violates item field
// constraints a [validator.ValidationErrors] is returned.
func (r *Repository) CreateItem(ctx context.Context, data *Data) (id string, err error) {
	if err := r.v.Struct(data); err != nil {
		return "", fmt.Errorf("invalid item data: %w", err)
	}

	newID := uuid.New()

	err = r.db.WithContext(ctx).
		Create(&Model{
			ID:    newID,
			Name:  data.Name,
			Desc:  data.Desc,
			Stock: data.Stock,
			Attrs: data.Attrs,
		}).Error
	if err != nil {
		return "", fmt.Errorf("could not create item: %w", err)
	}

	return newID.String(), nil
}
