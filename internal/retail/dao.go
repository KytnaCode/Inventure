package retail

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/kytnacode/inventure/internal/retail/placepath"
	"gorm.io/gorm"
)

// DAO handles low-level persistence operations on retails and places.
type DAO struct{}

// NewDAO creates a new [DAO].
func NewDAO() *DAO {
	return &DAO{}
}

// CreateRetailsList creates a list of retails and returns its IDs.
func (d *DAO) CreateRetailsList(tx *gorm.DB, data []Data) ([]uuid.UUID, error) {
	models := make([]Model, 0, len(data))

	for _, retailData := range data {
		m := Model{
			ID:       uuid.New(),
			Name:     retailData.Name,
			TenantID: retailData.TenantID,
		}

		storage := retailData.Storage
		if storage == nil {
			storage = &PlaceData{
				Name: "Storage",
			}
		}

		err := d.CreatePlace(tx, m.ID, storage, "/")
		if err != nil {
			return nil, fmt.Errorf("could not create retail: %w", err)
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

// CreateRetail creates a new retail and returns its ID.
func (d *DAO) CreateRetail(tx *gorm.DB, data *Data) (uuid.UUID, error) {
	ids, err := d.CreateRetailsList(tx, []Data{*data})
	if err != nil {
		return uuid.UUID{}, err
	}

	return ids[0], nil
}

// CreatePlace creates a place in a retail.
func (d *DAO) CreatePlace(
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

	items := make([]*ItemModel, 0, len(data.Items))

	for _, itemData := range data.Items {
		itemModel := NewItemModel(&itemData)

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
		err = d.CreatePlace(tx, retailID, &child, placePath)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetRetailPlaces get all places that belong to a given retail.
func (d *DAO) GetRetailPlaces(tx *gorm.DB, retailID uuid.UUID) ([]PlaceModel, error) {
	var places []PlaceModel

	err := tx.Where("retail_id = ?", retailID).
		Where("path LIKE '/%'").
		Preload("Items").
		Find(&places).Error
	if err != nil {
		return nil, fmt.Errorf("could not get retail places: %w", err)
	}

	return places, nil
}
