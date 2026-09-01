package retail_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/kytnacode/inventure/internal/auth/rbac"
	"github.com/kytnacode/inventure/internal/retail"
	"github.com/kytnacode/inventure/internal/user"
)

type testTenantRepo struct {
	tenantID uuid.UUID
	err      error
}

func (r *testTenantRepo) CreateFullTenant(
	_ context.Context,
	_ *retail.TenantData,
	_ []retail.Data,
	_ []uuid.UUID,
) (id uuid.UUID, err error) {
	if r.err != nil {
		return uuid.UUID{}, r.err
	}

	return r.tenantID, nil
}

func newService(repo *testTenantRepo) *retail.Service {
	return retail.NewService(repo)
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
		On:    rbac.NewResource("tenant", uuid.New(), nil),
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
		On:    rbac.NewResource("tenant", uuid.New(), nil),
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
