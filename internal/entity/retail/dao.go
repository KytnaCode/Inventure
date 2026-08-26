package retail

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/kytnacode/inventure/internal/entity/item"
	"github.com/kytnacode/inventure/internal/entity/retail/placepath"
	"gorm.io/gorm"
)

// DAO handles low-level persistence operations on retails and places.
type DAO struct {
	v *validator.Validate
}

// NewDAO creates a new [DAO].
func NewDAO(v *validator.Validate) *DAO {
	return &DAO{
		v: v,
	}
}

// CreateRetail creates a new retail and returns its ID.
func (d *DAO) CreateRetail(tx *gorm.DB, data *Data) (uuid.UUID, error) {
	m := &Model{
		ID:       uuid.New(),
		Name:     data.Name,
		TenantID: data.TenantID,
	}

	if err := d.v.Struct(m); err != nil {
		return uuid.UUID{}, fmt.Errorf("invalid retail data: %w", err)
	}

	err := tx.Create(m).Error
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("could not create retail: %w", err)
	}

	return m.ID, nil
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

	items := make([]*item.Model, 0, len(data.Items))

	for _, itemData := range data.Items {
		itemModel, err := item.NewModel(d.v, &itemData)
		if err != nil {
			return err
		}

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
