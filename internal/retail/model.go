package retail

import (
	"cmp"
	"slices"

	"github.com/google/uuid"
	"github.com/kytnacode/inventure/internal/retail/placepath"
	"github.com/kytnacode/inventure/internal/user"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ItemModel is the database representation of an item type.
type ItemModel struct {
	// ID is item's unique ID.
	ID uuid.UUID `gorm:"primaryKey"`

	// Name is item's name.
	Name string

	// Desc is item's description.
	Desc string

	// Attrs a per-item custom attributes.
	Attrs datatypes.JSONMap
}

// TableName returns item's table name. Implements [gorm/schema.Tabler].
func (m *ItemModel) TableName() string {
	return EntityItems
}

// NewItemModel validates data and returns a new [ItemModel].
func NewItemModel(data *ItemData) *ItemModel {
	m := &ItemModel{
		ID:    uuid.New(),
		Name:  data.Name,
		Desc:  data.Desc,
		Attrs: data.Attrs,
	}

	return m
}

// StockItemModel is the database representation of a stock item.
type StockItemModel struct {
	// ID is the stock item's unique ID.
	ID uuid.UUID

	// Data is the data of the item stored.
	Data ItemModel

	DataID uuid.UUID

	// Stock is the amount of items stored in this location.
	Stock int

	// PlaceID is the ID of the place where these items are stored.
	PlaceID uuid.UUID
}

// TableName returns stock items table name. Implements [gorm/schema.Tabler].
func (m *StockItemModel) TableName() string {
	return EntityStockItems
}

// ToDomain converts a [StockItemModel] into a [StockItem] domain object.
func (m *StockItemModel) ToDomain() *StockItem {
	return &StockItem{
		ID:      m.ID,
		Data:    *m.Data.ToDomain(),
		Stock:   m.Stock,
		PlaceID: m.PlaceID,
	}
}

// NewStockItemModel validates data and returns a new [StockItemModel].
func NewStockItemModel(data *StockItemData) *StockItemModel {
	m := &StockItemModel{
		ID:      uuid.New(),
		PlaceID: data.PlaceID,
		Stock:   data.Stock,
		DataID:  data.Data.ID,
	}

	return m
}

// ToDomain converts an [ItemModel] into an [Item].
func (m *ItemModel) ToDomain() *Item {
	return &Item{
		ID:    m.ID,
		Name:  m.Name,
		Desc:  m.Desc,
		Attrs: m.Attrs,
	}
}

// Model is the database representation of a retail.
type Model struct {
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
func (m *Model) TableName() string {
	return EntityRetails
}

// PlaceModel is the database representation of a [Place].
type PlaceModel struct {
	ID uuid.UUID `gorm:"primaryKey"`

	Path string `gorm:"uniqueIndex:idx_name"`

	// Name is place's name, must be unique between siblings.
	Name string `gorm:"uniqueIndex:idx_name"`

	RetailID uuid.UUID `gorm:"uniqueIndex:idx_name"`

	// Items are the items that reside directly on this place.
	Items []StockItemModel `gorm:"foreignKey:PlaceID"`
}

// TableName returns place's table name. Implements [gorm/schema.Tabler].
func (m *PlaceModel) TableName() string {
	return EntityPlaces
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

// TenantModel is the database representation of a tenant.
type TenantModel struct {
	gorm.Model

	// ID is tenant's unique ID.
	ID uuid.UUID

	// Name is tenant's display name.
	Name string

	// Users are tenant-scoped users.
	Users []user.Model `gorm:"many2many:tenant_users;"`

	// Retails are tenant-owned retails.
	Retails []Model `gorm:"foreignKey:TenantID"`
}

// TableName returns tenant's table name. Implements [gorm/schema.Tabler].
func (m *TenantModel) TableName() string {
	return EntityTenants
}
