package retail

import (
	"github.com/google/uuid"
	"github.com/kytnacode/inventure/internal/entity/item"
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

func itemModelToDomain(models []item.Model) []item.Item {
	items := make([]item.Item, 0, len(models))

	for _, it := range models {
		items = append(items, *it.ToDomain())
	}

	return items
}
