package retail

import (
	"os/user"
	"slices"

	"github.com/google/uuid"
	"github.com/kytnacode/inventure/internal/entity/item"
)

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
	Items []item.Item
}

// GetItems returns all place items, including the ones in children places.
func (p *Place) GetItems() []item.Item {
	itemSlices := make([][]item.Item, 0, p.num())

	itemSlices = p.getItems(itemSlices)

	return slices.Concat(itemSlices...)
}

func (p *Place) getItems(itemSlices [][]item.Item) [][]item.Item {
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
