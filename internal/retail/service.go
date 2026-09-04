package retail

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/kytnacode/inventure/internal/auth/rbac"
	"github.com/kytnacode/inventure/internal/user"
	"github.com/kytnacode/inventure/transaction"
	"gorm.io/gorm"
)

const (
	// MasterRoleName is the name of the tenant master role.
	MasterRoleName = "master"

	// DefaultStorageName is the default name for retail's root place.
	DefaultStorageName = "Storage"
)

// ServiceAdapters contains transaction-scoped repositories used by [Service].
type ServiceAdapters struct {
	// TenantRepo implements tenant operations.
	TenantRepo tenantRepository

	// RoleCreator implements role operations.
	RoleCreator rbac.RoleCreator

	// Assignator implements role assignation.
	Assignator roleAssignator
}

// ServiceProvider is a transaction provider for [Service].
type ServiceProvider = transaction.Provider[*ServiceAdapters]

// NewServiceProvider creates a new [ServiceProvider].
func NewServiceProvider(db *gorm.DB) ServiceProvider {
	return transaction.NewGormProvider(db, func(tx *gorm.DB) *ServiceAdapters {
		return &ServiceAdapters{
			TenantRepo:  NewRepository(tx),
			Assignator:  user.NewRepository(tx),
			RoleCreator: rbac.NewRepository(tx),
		}
	})
}

// tenantRepository abstracts away tenant's persistence layer.
//
// MUST be safe for concurrent use.
type tenantRepository interface {
	CreateFullTenant(
		ctx context.Context,
		data *TenantData,
		retails []Data,
		userIDs []uuid.UUID,
	) (id uuid.UUID, err error)
}

// roleAssignator assign roles to a user.
type roleAssignator interface {
	AssingRoles(ctx context.Context, userID uuid.UUID, roleIDs ...uuid.UUID) error
}

type repository interface {
	CreateRetail(ctx context.Context, data *Data) (uuid.UUID, error)
}

// Service handles high-level domain logic for tenants.
//
// Safe for concurrent use.
type Service struct {
	provider ServiceProvider
	repo     repository
}

// NewService creates a new [Service].
func NewService(provider ServiceProvider, repo repository) *Service {
	return &Service{
		provider: provider,
		repo:     repo,
	}
}

// CreateDefaultTenant creates the default tenant for new users.
func (s *Service) CreateDefaultTenant(
	ctx context.Context,
	creator user.User,
) (id uuid.UUID, err error) {
	data := TenantData{
		Name: fmt.Sprintf("%v's Company", creator.Name),
	}

	retails := []Data{
		{
			Name: "My Retail",
			Storage: &PlaceData{
				Name: "Storage",
			},
		},
	}

	err = s.provider.Transact(ctx, func(adapters *ServiceAdapters) error {
		repo := adapters.TenantRepo
		roleRepo := adapters.RoleCreator
		userRepo := adapters.Assignator

		id, err = repo.CreateFullTenant(ctx, &data, retails, []uuid.UUID{creator.ID})
		if err != nil {
			return err
		}

		res := rbac.NewResource(EntityTenants, id, nil)

		masterRoleID, err := rbac.RoleBuilder(roleRepo).
			Name(MasterRoleName).BelongsTo(res).
			On(res).Allow(PermRoot).Build(ctx)
		if err != nil {
			return err
		}

		return userRepo.AssingRoles(ctx, creator.ID, masterRoleID)
	})
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("could not create default tenant: %w", err)
	}

	return id, nil
}

// CreateEmptyRetail creates a new empty retail with an empty root storage on the given tenant.
func (s *Service) CreateEmptyRetail(
	ctx context.Context,
	name string,
	tenantID uuid.UUID,
) (uuid.UUID, error) {
	return s.repo.CreateRetail(ctx, &Data{
		Name:     name,
		TenantID: tenantID,
		Storage: &PlaceData{
			Name: DefaultStorageName,
		},
	})
}
