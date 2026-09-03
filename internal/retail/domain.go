package retail

import (
	"slices"

	"github.com/google/uuid"
	"github.com/kytnacode/inventure/internal/user"
)

// Entity reference and table names.
const (
	EntityItems   = "items"
	EntityRetails = "retails"
	EntityTenants = "tenants"
	EntityPlaces  = "places"
)

// Item represents a retail's item.
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

// Retail is a physic or logic sell point that manage users and storage.
type Retail struct {
	// ID is retail's unique ID.
	ID string

	// Name is retail's display name.
	Name string

	// Users are the list of retail-scoped users.
	Users []user.User

	// Storage manages retail's stock
	Storage Place

	// TenantID is the ID of the tenant the retails belongs to.
	TenantID uuid.UUID
}

// Place is a physical or logic, place where items are stored, can have children places and can
// be child of another place.
type Place struct {
	// Parent place, nil for retail's root storage.
	Parent *Place

	// Name is place name, must be unique between siblings.
	Name string

	// Children are children places.
	Children []Place

	// Items are items directly stored in this places, items stored in inner places are not included,
	// to get all items including the ones stored in children places use [Place.GetItems].
	Items []Item
}

// GetItems returns all place items, including the ones in children places.
func (p *Place) GetItems() []Item {
	itemSlices := make([][]Item, 0, p.num())

	itemSlices = p.getItems(itemSlices)

	return slices.Concat(itemSlices...)
}

func (p *Place) getItems(itemSlices [][]Item) [][]Item {
	itemSlices = append(itemSlices, p.Items)

	for _, child := range p.Children {
		itemSlices = child.getItems(itemSlices)
	}

	return itemSlices
}

func (p *Place) num() int {
	n := 1

	for _, child := range p.Children {
		n += child.num()
	}

	return n
}

// Reference contains a reference to a resource.
type Reference struct {
	// Typ is resource's type.
	Typ string

	// ID is resource' ID.
	ID string
}

// Tenant represents a company that can own retails and have users.
type Tenant struct {
	// ID is tenant's unique ID.
	ID uuid.UUID

	// Name is tenant's display name.
	Name string

	// Users are tenant-scoped users for this tenant.
	Users []user.User

	// Retails are retails owned by this tenant.
	Retails []Retail
}
