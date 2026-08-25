package retail

import (
	"cmp"
	"slices"

	"github.com/google/uuid"
	"github.com/kytnacode/inventure/internal/item"
	"github.com/kytnacode/inventure/internal/retail/placepath"
	"github.com/kytnacode/inventure/internal/user"
)

// Model is the database representation of a [Retail].
type Model struct {
	// ID is the retail's unique ID.
	ID uuid.UUID `gorm:"primaryKey"`

	// Name is the retail's display name.
	Name string `validate:"required,resourcename"`

	// Users are the users the retail has.
	Users []user.Model `gorm:"polymorphic:Resource;" validate:"dive"`

	// Storage is the root place where items are stored.
	Storage PlaceModel `gorm:"foreignKey:RetailID"`
}

// TableName returns retail's table name. Implements [gorm/schema.Tabler].
func (m *Model) TableName() string {
	return "retails"
}

// PlaceModel is the database representation of a [Place].
type PlaceModel struct {
	ID uuid.UUID `gorm:"primaryKey"`

	Path string `gorm:"uniqueIndex:idx_name"`

	// Name is place's name, must be unique between siblings.
	Name string `gorm:"uniqueIndex:idx_name"`

	RetailID uuid.UUID

	// Items are the items that reside directly on this place.
	Items []item.Model `gorm:"foreignKey:PlaceID"`
}

// TableName returns place's table name. Implements [gorm/schema.Tabler].
func (m *PlaceModel) TableName() string {
	return "places"
}

// PlaceFromModel creates a [Place] from a list of [PlaceModel]. It's not necessary to pass all
// place models contained in the root place, this would give a partial place tree, however,
// if a place is included, its parent MUST be included also.
func PlaceFromModel(places []PlaceModel) *Place {
	ps := slices.SortedFunc(slices.Values(places), func(a, b PlaceModel) int {
		byLen := cmp.Compare(len(a.Path), len(b.Path))
		if byLen == 0 {
			return cmp.Compare(a.Path, b.Path)
		}

		return byLen
	})

	var p Place

	_ = placeFromModel(&p, nil, "/", ps, -1)

	return &p
}

func placeFromModel(place, parent *Place, currPath string, places []PlaceModel, i int) int {
	children := make(map[string]*Place)

	for {
		i++

		if i >= len(places) {
			break
		}

		p := places[i]

		localPath := placepath.CutPrefix(p.Path, currPath)

		if localPath == "/" {
			place.Name = p.Name
			place.Children = make([]Place, 0, 3)
			place.Parent = parent
			place.Items = itemModelToDomain(p.Items)

			continue
		}

		components := placepath.Components(localPath)

		if len(components) == 0 {
			continue
		}

		if len(components) == 1 {
			children[components[0]] = &Place{
				Name:     p.Name,
				Children: make([]Place, 0, 2),
				Parent:   place,
				Items:    itemModelToDomain(p.Items),
			}

			continue
		}

		v, ok := children[components[0]]
		if !ok {
			break
		}

		i = placeFromModel(v, place, placepath.Join(currPath, components[0]), places, i-1)
	}

	for _, c := range children {
		place.Children = append(place.Children, *c)
	}

	return i
}
