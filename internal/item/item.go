// Package item contains item's domain and business logic, and routes for items.
package item

// Item represents a limited-count item.
type Item struct {
	// ID is the item's unique ID.
	ID string

	// Name is the item's name. Must not ever be empty.
	Name string `validate:"required,max=80,alphanumspace"`

	// Desc is the item description. May be empty.
	Desc string `validate:"max=65536,alphanumspace"`

	// Stock how many units remains
	Stock int `validate:"gte=0"`

	// Attrs contains custom per-item attributes.
	Attrs map[string]any `validate:"dive,keys,required,max=80,alphanumspace,endKeys"`

	// PlaceID is the ID of the place where the item resides.
	PlaceID string `validate:"required"`
}
