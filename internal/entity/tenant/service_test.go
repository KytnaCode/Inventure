package tenant_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/kytnacode/inventure/internal/entity"
	"github.com/kytnacode/inventure/internal/entity/tenant"
	"github.com/kytnacode/inventure/internal/entity/user"
)

type testTenantRepo struct {
	tenantID uuid.UUID
	err      error
}

func (r *testTenantRepo) CreateFullTenant(
	_ context.Context,
	_ *tenant.Data,
	_ []tenant.RetailData,
	_ []uuid.UUID,
) (id uuid.UUID, err error) {
	if r.err != nil {
		return uuid.UUID{}, r.err
	}

	return r.tenantID, nil
}

func newService(repo *testTenantRepo) *tenant.Service {
	return tenant.NewService(repo)
}

func TestService_CreateDefaultTenantShouldCreateTenant(t *testing.T) {
	t.Parallel()

	expected := uuid.New()

	service := newService(&testTenantRepo{
		tenantID: expected,
	})

	creator := user.User{
		ID:    uuid.New(),
		Name:  "Amity Blight",
		Email: "amity.blight@gmail.com",
		On:    entity.NewReference("tenant", uuid.New(), nil),
	}

	got, err := service.CreateDefaultTenant(t.Context(), creator)
	if err != nil {
		t.Fatalf("expected a nil error: %v", err)
	}

	if got.String() != expected.String() {
		t.Errorf("expected tenant id to be '%v': got '%v'", expected, got)
	}
}

func TestService_CreateDefaultTenantShouldWrapRepositoryError(t *testing.T) {
	t.Parallel()

	ErrExpected := errors.New("expected")

	service := newService(&testTenantRepo{
		err: ErrExpected,
	})

	creator := user.User{
		ID:    uuid.New(),
		Name:  "Amity Blight",
		Email: "amity.blight@gmail.com",
		On:    entity.NewReference("tenant", uuid.New(), nil),
	}

	got, err := service.CreateDefaultTenant(t.Context(), creator)
	if got.String() != (uuid.UUID{}).String() {
		t.Errorf("expected got ID to be the zero UUID: got '%v'", got)
	}

	if err == nil {
		t.Fatalf("expected a non-nil error: %v", err)
	}

	if !errors.Is(err, ErrExpected) {
		t.Errorf("expected error to wrap repository error: got '%v'", err)
	}
}
