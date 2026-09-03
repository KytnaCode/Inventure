package retail_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/kytnacode/inventure/internal/auth/rbac"
	"github.com/kytnacode/inventure/internal/retail"
	"github.com/kytnacode/inventure/internal/user"
	"github.com/kytnacode/inventure/transaction"
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

type testRoleCreator struct {
	err error
}

func (r *testRoleCreator) CreateRole(
	_ context.Context,
	_ *rbac.RoleData,
	_ ...rbac.AccessData,
) (id uuid.UUID, err error) {
	if r.err != nil {
		return uuid.UUID{}, r.err
	}

	return uuid.New(), nil
}

type roleAssignator struct {
	err error
}

func (r *roleAssignator) AssingRoles(_ context.Context, _ uuid.UUID, _ ...uuid.UUID) error {
	return r.err
}

func newService(
	repo *testTenantRepo,
	roleAssignator *roleAssignator,
	roleCreator *testRoleCreator,
) *retail.Service {
	provider := transaction.NewTestProvider(func() *retail.ServiceAdapters {
		return &retail.ServiceAdapters{
			TenantRepo:  repo,
			RoleCreator: roleCreator,
			Assignator:  roleAssignator,
		}
	})

	return retail.NewService(provider)
}

func TestService_CreateDefaultTenantShouldCreateTenant(t *testing.T) {
	t.Parallel()

	expected := uuid.New()

	service := newService(&testTenantRepo{
		tenantID: expected,
	}, &roleAssignator{}, &testRoleCreator{})

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
	}, &roleAssignator{}, &testRoleCreator{})

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

func TestService_CreateDefaultTenantShouldWrapRoleAsignatorError(t *testing.T) {
	t.Parallel()

	ErrExpected := errors.New("expected")

	service := newService(&testTenantRepo{}, &roleAssignator{err: ErrExpected}, &testRoleCreator{})

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
		t.Errorf("expected error to wrap assignator error: got '%v'", err)
	}
}

func TestService_CreateDefaultTenantShouldWrapRoleRepositoryError(t *testing.T) {
	t.Parallel()

	ErrExpected := errors.New("expected")

	service := newService(&testTenantRepo{}, &roleAssignator{}, &testRoleCreator{err: ErrExpected})

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
