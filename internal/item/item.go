// Package item contains item's domain and business logic, and routes for items.
package item

// Item represents a limited-count item.
type Item struct {
	// ID is the item's unique ID.
	ID string

	// Name is the item's name. Must not ever be empty.
	Name string

	// Desc is the item description. May be empty.
	Desc string

	// Stock how many units remains
	Stock int

	// Attrs contains custom per-item attributes.
	Attrs map[string]any
}
