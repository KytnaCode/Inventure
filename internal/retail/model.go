package retail

import (
	"cmp"
	"fmt"
	"slices"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/kytnacode/inventure/internal/retail/placepath"
	"github.com/kytnacode/inventure/internal/user"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type ItemModel struct {
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
func (m *ItemModel) TableName() string {
	return "items"
}

// NewModel validates data and returns a new [ItemModel].
func NewModel(v *validator.Validate, data *ItemData) (*ItemModel, error) {
	if err := v.Struct(data); err != nil {
		return nil, fmt.Errorf("invalid item data: %w", err)
	}

	m := &ItemModel{
		ID:    uuid.New(),
		Name:  data.Name,
		Desc:  data.Desc,
		Stock: data.Stock,
		Attrs: data.Attrs,
	}

	return m, nil
}

// ToDomain converts an [ItemModel] into an [Item].
func (m *ItemModel) ToDomain() *Item {
	return &Item{
		ID:      m.ID.String(),
		Name:    m.Name,
		Desc:    m.Desc,
		Stock:   m.Stock,
		Attrs:   m.Attrs,
		PlaceID: m.PlaceID.String(),
	}
}

type RetailModel struct {
	// ID is the retail's unique ID.
	ID uuid.UUID `gorm:"primaryKey"`

	// Name is the retail's display name.
	Name string `validate:"required,resourcename"`

	// Users are the users the retail has.
	Users []user.Model `gorm:"many2many:retail_users;" validate:"dive"`

	// Storage is the root place where items are stored.
	Storage PlaceModel `gorm:"foreignKey:RetailID"`

	// TenantID is the ID of the tenant the retails belongs to, a retail MUST be part of a tenant.
	TenantID uuid.UUID
}

// TableName returns retail's table name. Implements [gorm/schema.Tabler].
func (m *RetailModel) TableName() string {
	return "retails"
}

// PlaceModel is the database representation of a [Place].
type PlaceModel struct {
	ID uuid.UUID `gorm:"primaryKey"`

	Path string `gorm:"uniqueIndex:idx_name"`

	// Name is place's name, must be unique between siblings.
	Name string `gorm:"uniqueIndex:idx_name"`

	RetailID uuid.UUID `gorm:"uniqueIndex:idx_name"`

	// Items are the items that reside directly on this place.
	Items []ItemModel `gorm:"foreignKey:PlaceID"`
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

type TenantModel struct {
	gorm.Model

	// ID is tenant's unique ID.
	ID uuid.UUID

	// Name is tenant's display name.
	Name string

	// Users are tenant-scoped users.
	Users []user.Model `gorm:"many2many:tenant_users;"`

	// Retails are tenant-owned retails.
	Retails []RetailModel `gorm:"foreignKey:TenantID"`
}

// TableName returns tenant's table name. Implements [gorm/schema.Tabler].
func (m *TenantModel) TableName() string {
	return "tenants"
}
