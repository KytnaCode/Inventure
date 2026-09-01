package retail

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
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
	Items []ItemData
}

// ItemData is the necessary data to create an item.
type ItemData struct {
	// Name is item's name.
	Name string `validate:"required,max=80,resourcename"`

	// Desc is item's description.
	Desc string `validate:"max=65565,max=0|resourcename"`

	// Stock is item's stock, must be positive.
	Stock int `validate:"gte=0"`

	// Attrs are optional per-item custom attributes.
	Attrs map[string]any `validate:"dive,keys,required,min=1,max=80,endkeys"`
}

var _ tenantRepository = &Repository{}

// Repository handles high level persistence logic for tenants.
type Repository struct {
	db        *gorm.DB
	userDAO   *user.DAO
	retailDAO *DAO
}

// NewRepository creates a new [Repository].
func NewRepository(db *gorm.DB, v *validator.Validate) *Repository {
	return &Repository{
		db:        db,
		userDAO:   user.NewDAO(),
		retailDAO: NewDAO(v),
	}
}

// CreateFullTenant creates a tenant and a list of retails owned by it, it also set the given user
// ids as members of the tenant, it drops non-existing user ids.
func (r *Repository) CreateFullTenant(
	ctx context.Context,
	data *TenantData,
	retails []Data,
	userIDs []uuid.UUID,
) (id uuid.UUID, err error) {
	users := make([]user.Model, 0, len(userIDs))

	for _, id := range userIDs {
		found, err := r.userDAO.Exists(r.db, id)
		if err != nil {
			return uuid.UUID{}, fmt.Errorf("could not check if user with ID '%v' exists", id)
		}

		if found {
			users = append(users, user.Model{
				ID: id,
			})
		}
	}

	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		tenantID := uuid.New()

		var retailModels []RetailModel

		if len(retails) != 0 {
			ids, err := r.retailDAO.CreateRetailsList(tx, dataFromRetailData(retails, tenantID))
			if err != nil {
				return err
			}

			retailModels = make([]RetailModel, 0, len(ids))

			for _, id := range ids {
				retailModels = append(retailModels, RetailModel{
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

func itemModelToDomain(models []ItemModel) []Item {
	items := make([]Item, 0, len(models))

	for _, it := range models {
		items = append(items, *it.ToDomain())
	}

	return items
}
