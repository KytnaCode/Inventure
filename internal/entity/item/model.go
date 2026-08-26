package item

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// Model is a database representation of an item.
type Model struct {
	// ID is item's unique ID.
	ID uuid.UUID `gorm:"primaryKey"`

	// Name is item's name.
	Name string

	// Desc is item's description.
	Desc string

	// Stock is item's available stock.
	Stock int

	// Attrs a per-item custom attributes.
	Attrs datatypes.JSONMap

	// PlaceID is the ID of the place the item is residing on.
	PlaceID uuid.UUID
}

// TableName returns item's table name. Implements [gorm/schema.Tabler].
func (m *Model) TableName() string {
	return "items"
}

// NewModel validates data and returns a new [Model].
func NewModel(v *validator.Validate, data *Data) (*Model, error) {
	if err := v.Struct(data); err != nil {
		return nil, fmt.Errorf("invalid item data: %w", err)
	}

	m := &Model{
		ID:    uuid.New(),
		Name:  data.Name,
		Desc:  data.Desc,
		Stock: data.Stock,
		Attrs: data.Attrs,
	}

	return m, nil
}

// ToDomain converts a [Model] into an [Item].
func (m *Model) ToDomain() *Item {
	return &Item{
		ID:      m.ID.String(),
		Name:    m.Name,
		Desc:    m.Desc,
		Stock:   m.Stock,
		Attrs:   m.Attrs,
		PlaceID: m.PlaceID.String(),
	}
}
