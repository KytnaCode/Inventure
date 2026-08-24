package item

import (
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

	// PlaceType is the type of the place the item is residing on.
	PlaceType string

	// PlaceID is the ID of the place the item is residing on.
	PlaceID uuid.UUID
}
