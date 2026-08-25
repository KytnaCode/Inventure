package retail

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/kytnacode/inventure/internal/item"
	"github.com/kytnacode/inventure/internal/retail/placepath"
	"gorm.io/gorm"
)

// Data is used by [Repository] to create a new retail.
type Data struct {
	// Name is retail's display name.
	Name string
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

// Repository handles business logic for retails and places.
type Repository struct {
	db *gorm.DB
	v  *validator.Validate
}

// NewRepository creates a new [Repository].
func NewRepository(db *gorm.DB, v *validator.Validate) *Repository {
	return &Repository{
		db: db,
		v:  v,
	}
}

// CreateRetail creates a new retail and returns its ID.
func (r *Repository) CreateRetail(
	ctx context.Context,
	retailData *Data,
	storage *PlaceData,
) (id string, err error) {
	var newID string

	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		retailID, err := r.createRetail(tx, retailData)
		if err != nil {
			return err
		}

		err = r.createPlace(tx, retailID, storage, "/")
		if err != nil {
			return err
		}

		newID = retailID.String()

		return nil
	})
	if err != nil {
		return "", fmt.Errorf("could not create retail: %w", err)
	}

	return newID, nil
}

func (r *Repository) createRetail(tx *gorm.DB, data *Data) (uuid.UUID, error) {
	retailID := uuid.New()

	m := &Model{
		ID:   retailID,
		Name: data.Name,
	}

	if err := r.v.Struct(m); err != nil {
		return uuid.UUID{}, fmt.Errorf("invalid retail data: %w", err)
	}

	err := tx.Create(m).Error
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("could not create retail: %w", err)
	}

	return retailID, nil
}

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

	items := make([]*item.Model, 0, len(data.Items))

	for _, itemData := range data.Items {
		itemModel, err := item.NewModel(r.v, &itemData)
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
		err = r.createPlace(tx, retailID, &child, placePath)
		if err != nil {
			return err
		}
	}

	return nil
}

func itemModelToDomain(models []item.Model) []item.Item {
	items := make([]item.Item, 0, len(models))

	for _, it := range models {
		items = append(items, *it.ToDomain())
	}

	return items
}
