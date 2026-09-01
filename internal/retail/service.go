package retail

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/kytnacode/inventure/internal/user"
)

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

// Service handles high-level domain logic for tenants.
//
// Safe for concurrent use.
type Service struct {
	repo tenantRepository
}

// NewService creates a new [Service].
func NewService(repo tenantRepository) *Service {
	return &Service{
		repo: repo,
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

	id, err = s.repo.CreateFullTenant(ctx, &data, retails, []uuid.UUID{creator.ID})
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("could not create default tenant: %w", err)
	}

	return id, nil
}
