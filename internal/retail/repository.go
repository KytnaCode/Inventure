package retail

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/kytnacode/inventure/internal/retail/placepath"
	"github.com/kytnacode/inventure/internal/user"
	"gorm.io/gorm"
)

// TenantData is the necessary data to create a tenant.
type TenantData struct {
	Name string
}

// Data is the necessary data to create a retail.
type Data struct {
	// Name is retail's display name.
	Name string

	// TenantID is the tenant the retail belongs, required. MUST exists.
	TenantID uuid.UUID

	// Storage is the data for retail's root storage. If missing a default storage will be created.
	Storage *PlaceData
}

// PlaceData is the necessary data to create a place.
type PlaceData struct {
	// Name is place's display name.
	Name string

	// Children are other places contained in this one.
	Children []PlaceData

	// Items are items directly contained in this place.
	Items []StockItemData
}

// ItemData is the necessary data to create an item definition.
type ItemData struct {
	// Name is item's name.
	Name string `validate:"required,max=80,resourcename"`

	// Desc is item's description.
	Desc string `validate:"max=65565,max=0|resourcename"`

	// Attrs are optional per-item custom attributes.
	Attrs map[string]any `validate:"dive,keys,required,min=1,max=80,endkeys"`
}

// StockItemData is the necessary data to create a new [StockItem].
type StockItemData struct {
	// Data is the definition data of the stored item.
	Data *Item

	// Stock is the amount of units stored in this place.
	Stock int

	// PlaceID is the place where these items are stored.
	PlaceID uuid.UUID
}

var _ tenantRepository = &Repository{}

// Repository handles high level persistence logic for tenants.
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new [Repository].
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{
		db: db,
	}
}

// CreateFullTenant creates a tenant and a list of retails owned by it, it also set the given user
// ids as members of the tenant.
func (r *Repository) CreateFullTenant(
	ctx context.Context,
	data *TenantData,
	retails []Data,
	userIDs []uuid.UUID,
) (id uuid.UUID, err error) {
	users := make([]user.Model, 0, len(userIDs))

	for _, id := range userIDs {
		users = append(users, user.Model{
			ID: id,
		})
	}

	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		tenantID := uuid.New()

		var retailModels []Model

		if len(retails) != 0 {
			ids, err := r.createRetailsList(tx, dataFromRetailData(retails, tenantID))
			if err != nil {
				return err
			}

			retailModels = make([]Model, 0, len(ids))

			for _, id := range ids {
				retailModels = append(retailModels, Model{
					ID: id,
				})
			}
		}

		m := &TenantModel{
			ID:      tenantID,
			Name:    data.Name,
			Users:   users,
			Retails: retailModels,
		}

		err = tx.Create(&m).Error
		if err != nil {
			return err
		}

		id = m.ID

		return nil
	})
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("could not create tenant: %w", err)
	}

	return id, nil
}

// CreateRetail creates a new retail.
func (r *Repository) CreateRetail(
	ctx context.Context,
	data *Data,
) (uuid.UUID, error) {
	m, err := r.createRetail(r.db.WithContext(ctx), data)
	if err != nil {
		return uuid.UUID{}, err
	}

	err = r.db.WithContext(ctx).Create(m).Error
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("could not insert model: %w", err)
	}

	return m.ID, nil
}

// CreateItemType creates a new [Item].
func (r *Repository) CreateItemType(ctx context.Context, data *ItemData) (*Item, error) {
	m := &ItemModel{
		ID:    uuid.New(),
		Name:  data.Name,
		Desc:  data.Desc,
		Attrs: data.Attrs,
	}

	err := r.db.WithContext(ctx).Create(&m).Error
	if err != nil {
		return nil, fmt.Errorf("could not create item type: %w", err)
	}

	return m.ToDomain(), nil
}

func dataFromRetailData(retailData []Data, tenantID uuid.UUID) []Data {
	data := make([]Data, 0, len(retailData))

	for _, d := range retailData {
		data = append(data, Data{
			Name:     d.Name,
			Storage:  d.Storage,
			TenantID: tenantID,
		})
	}

	return data
}

func itemModelToDomain(models []StockItemModel) []StockItem {
	items := make([]StockItem, 0, len(models))

	for _, it := range models {
		items = append(items, *it.ToDomain())
	}

	return items
}

func (r *Repository) createRetail(tx *gorm.DB, data *Data) (*Model, error) {
	m := Model{
		ID:       uuid.New(),
		Name:     data.Name,
		TenantID: data.TenantID,
	}

	storage := data.Storage
	if storage == nil {
		storage = &PlaceData{
			Name: "Storage",
		}
	}

	err := r.createPlace(tx, m.ID, storage, "/")
	if err != nil {
		return nil, fmt.Errorf("could not create retail: %w", err)
	}

	return &m, nil
}

// createRetailsList creates a list of retails and returns its IDs.
func (r *Repository) createRetailsList(tx *gorm.DB, data []Data) ([]uuid.UUID, error) {
	models := make([]*Model, 0, len(data))

	for _, retailData := range data {
		m, err := r.createRetail(tx, &retailData)
		if err != nil {
			return nil, err
		}

		models = append(models, m)
	}

	err := tx.Create(models).Error
	if err != nil {
		return nil, fmt.Errorf("could not create retail: %w", err)
	}

	ids := make([]uuid.UUID, 0, len(models))

	for _, m := range models {
		ids = append(ids, m.ID)
	}

	return ids, nil
}

// createPlace creates a place in a retail.
func (r *Repository) createPlace(
	tx *gorm.DB,
	retailID uuid.UUID,
	data *PlaceData,
	placePath string,
) error {
	m := &PlaceModel{
		ID:       uuid.New(),
		Name:     data.Name,
		RetailID: retailID,
	}

	placePath = placepath.Join(placePath, m.ID.String())

	m.Path = placepath.TrimLeftPath(placePath)

	err := tx.Create(m).Error
	if err != nil {
		return fmt.Errorf("could not create place: %w", err)
	}

	items := make([]*StockItemModel, 0, len(data.Items))

	for _, itemData := range data.Items {
		itemModel := NewStockItemModel(&itemData)

		itemModel.PlaceID = m.ID

		items = append(items, itemModel)
	}

	if len(items) > 0 {
		err = tx.Create(&items).Error
		if err != nil {
			return fmt.Errorf("could not create place items: %w", err)
		}
	}

	for _, child := range data.Children {
		err = r.createPlace(tx, retailID, &child, placePath)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetRetailStorage the retails root storage, is equivalent to call [GetRetailPlaceByPath] with
// path argument set to '/'.
func (r *Repository) GetRetailStorage(tx *gorm.DB, retailID uuid.UUID) (*Place, error) {
	return r.GetRetailPlaceByPath(tx, retailID, "/")
}

// GetRetailPlaceByPath gets a retail place by its path, populating children places.
func (r *Repository) GetRetailPlaceByPath(
	tx *gorm.DB,
	retailID uuid.UUID,
	path string,
) (*Place, error) {
	var places []PlaceModel

	err := tx.Where("retail_id = ?", retailID).
		Where("path LIKE ?", path+"%").
		Preload("Items").
		Preload("Items.Data").
		Find(&places).Error
	if err != nil {
		return nil, fmt.Errorf("could not get retail places: %w", err)
	}

	fmt.Println(places)

	p := PlaceFromModel(places)

	return p, nil
}
