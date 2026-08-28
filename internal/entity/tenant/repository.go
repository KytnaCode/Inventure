package tenant

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/kytnacode/inventure/internal/entity/retail"
	"github.com/kytnacode/inventure/internal/entity/user"
	"gorm.io/gorm"
)

// Data is the necessary data to create a tenant with [Repository].
type Data struct {
	Name string
}

// RetailData is the necessary data to create retail with [Repository].
type RetailData struct {
	Name string

	Storage *retail.PlaceData
}

// Ensure [Repository] implements [tenantRepository].
var _ tenantRepository = &Repository{}

// Repository handles high level persistence logic for tenants.
type Repository struct {
	db        *gorm.DB
	userDAO   *user.DAO
	retailDAO *retail.DAO
}

// NewRepository creates a new [Repository].
func NewRepository(db *gorm.DB, v *validator.Validate) *Repository {
	return &Repository{
		db:        db,
		userDAO:   user.NewDAO(),
		retailDAO: retail.NewDAO(v),
	}
}

// CreateFullTenant creates a tenant and a list of retails owned by it, it also set the given user
// ids as members of the tenant, it drops non-existing user ids.
func (r *Repository) CreateFullTenant(
	ctx context.Context,
	data *Data,
	retails []RetailData,
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

		var retailModels []retail.Model

		if len(retails) != 0 {
			ids, err := r.retailDAO.CreateRetailsList(tx, dataFromRetailData(retails, tenantID))
			if err != nil {
				return err
			}

			retailModels = make([]retail.Model, 0, len(ids))

			for _, id := range ids {
				retailModels = append(retailModels, retail.Model{
					ID: id,
				})
			}
		}

		m := &Model{
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

func dataFromRetailData(retailData []RetailData, tenantID uuid.UUID) []retail.Data {
	data := make([]retail.Data, 0, len(retailData))

	for _, d := range retailData {
		data = append(data, retail.Data{
			Name:     d.Name,
			Storage:  d.Storage,
			TenantID: tenantID,
		})
	}

	return data
}
