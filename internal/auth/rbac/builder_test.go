package rbac_test

import (
	"cmp"
	"context"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/kytnacode/inventure/internal/auth/rbac"
)

type testRepository struct {
	err error
}

func (r *testRepository) CreateRole(
	_ context.Context,
	_ *rbac.RoleData,
	_ ...rbac.AccessData,
) (id uuid.UUID, err error) {
	if r.err != nil {
		return uuid.UUID{}, r.err
	}

	return uuid.New(), nil
}

func joinPerms[T ~string](xs []T) string {
	slices.Sort(xs)

	var str strings.Builder

	first := true

	for _, v := range xs {
		if !first {
			_, _ = str.WriteString(":")
		}

		_, _ = str.WriteString(string(v))

		first = false
	}

	return str.String()
}

func TestBuilderShouldRequireRoleName(t *testing.T) {
	t.Parallel()

	res := rbac.NewResource("tenant", uuid.New(), nil)

	id, err := rbac.RoleBuilder(&testRepository{}).
		// Name("admin").
		BelongsTo(res).
		On(res).Allow("tenant-mod", "tenant-del").
		Build(t.Context())
	if id.String() != (uuid.UUID{}).String() {
		t.Errorf("expected an empty ID: got '%v'", id)
	}

	if err == nil {
		t.Fatal("expected a non-nil error")
	}

	if !errors.Is(err, rbac.ErrMissingRoleData) {
		t.Errorf("expected error to be ErrMissingRoleData: got '%v'", err)
	}
}

func TestBuilderShouldRequireRoleResource(t *testing.T) {
	t.Parallel()

	res := rbac.NewResource("tenant", uuid.New(), nil)

	id, err := rbac.RoleBuilder(&testRepository{}).
		Name("admin").
		// BelongsTo(res).
		On(res).Allow("tenant-mod", "tenant-del").
		Build(t.Context())
	if id.String() != (uuid.UUID{}).String() {
		t.Errorf("expected an empty ID: got '%v'", id)
	}

	if err == nil {
		t.Fatal("expected a non-nil error")
	}

	if !errors.Is(err, rbac.ErrMissingRoleData) {
		t.Errorf("expected error to be ErrMissingRoleData: got '%v'", err)
	}
}

func TestBuilderShouldCreateRoleWithoutPermissions(t *testing.T) {
	t.Parallel()

	res := rbac.NewResource("tenant", uuid.New(), nil)

	id, err := rbac.RoleBuilder(&testRepository{}).
		Name("admin").
		BelongsTo(res).
		Build(t.Context())
	if err != nil {
		t.Error("expected a nil error")
	}

	if id.String() == (uuid.UUID{}).String() {
		t.Error("expected id to be not empty")
	}
}

func TestBuilderShouldCreateRoleWithPermissions(t *testing.T) {
	t.Parallel()

	res := rbac.NewResource("tenant", uuid.New(), nil)
	merchant1 := rbac.NewResource("retail", uuid.New(), nil)
	merchant2 := rbac.NewResource("retail", uuid.New(), nil)

	id, err := rbac.RoleBuilder(&testRepository{}).
		Name("admin").
		BelongsTo(res).
		On(res).Allow("tenant-mod").Allow("tenant-del").
		On(merchant1).Allow("item-read").
		On(merchant2).Allow("item-read", "iteam-del", "item-add").
		Build(t.Context())
	if err != nil {
		t.Error("expected a nil error")
	}

	if id.String() == (uuid.UUID{}).String() {
		t.Error("expected id to be not empty")
	}
}

func TestBuilderShouldCreateRoleWithRemovedPermissions(t *testing.T) {
	t.Parallel()

	res := rbac.NewResource("tenant", uuid.New(), nil)
	merchant1 := rbac.NewResource("retail", uuid.New(), nil)
	merchant2 := rbac.NewResource("retail", uuid.New(), nil)

	id, err := rbac.RoleBuilder(&testRepository{}).
		Name("admin").
		BelongsTo(res).
		On(res).Allow("tenant-mod").Allow("tenant-del").
		On(merchant1).Allow("item-read").
		On(merchant2).Allow("item-read", "iteam-del", "item-add").
		Remove(merchant1).
		Build(t.Context())
	if err != nil {
		t.Error("expected a nil error")
	}

	if id.String() == (uuid.UUID{}).String() {
		t.Error("expected id to be not empty")
	}
}

func TestBuilderSqlite3ShouldCreateRoleWithoutPermissions(t *testing.T) {
	t.Parallel()

	repo, _ := newRepo(t)

	res := rbac.NewResource("tenant", uuid.New(), nil)

	const (
		expectedName        = "admin"
		expectedAccessesNum = 0
	)

	id, err := rbac.RoleBuilder(repo).
		Name(expectedName).
		BelongsTo(res).
		Build(t.Context())
	if err != nil {
		t.Error("expected a nil error")
	}

	if id.String() == (uuid.UUID{}).String() {
		t.Error("expected id to be not empty")
	}

	roles, err := repo.GetRoles(t.Context(), id)
	if err != nil {
		t.Fatalf("could not get inserted roles: %v", err)
	}

	if len(roles) != 1 {
		t.Fatalf("expected only one role: got %v roles", len(roles))
	}

	got := roles[0]

	if got.Name != expectedName {
		t.Errorf("expected role name to be '%v': got '%v'", expectedName, got.Name)
	}

	if !got.On.Equal(res) {
		t.Errorf("expected role resource to be '%v': got '%v'", res, got.On)
	}

	if len(got.Accesses) != expectedAccessesNum {
		t.Errorf("expected to have %v accesses: got %v", expectedAccessesNum, len(got.Accesses))
	}
}

func TestBuilderSqlite3ShouldCreateRoleWithPermissions(t *testing.T) {
	t.Parallel()

	repo, _ := newRepo(t)

	res := rbac.NewResource("tenant", uuid.New(), nil)
	merchant1 := rbac.NewResource("retail", uuid.New(), nil)
	merchant2 := rbac.NewResource("retail", uuid.New(), nil)

	const (
		expectedName        = "admin"
		expectedAccessesNum = 3
	)

	var (
		expectedTenantPerms    = []rbac.Perm{"tenant-mod", "tenant-del", "user-add"}
		expectedMerchant1Perms = []rbac.Perm{"item-read"}
		expectedMerchant2Perms = []rbac.Perm{"item-read", "item-del", "item-add"}
		expectedAcceses        = []rbac.Access{
			{
				On:    res,
				Perms: expectedTenantPerms,
			},
			{
				On:    merchant1,
				Perms: expectedMerchant1Perms,
			},
			{
				On:    merchant2,
				Perms: expectedMerchant2Perms,
			},
		}
	)

	id, err := rbac.RoleBuilder(repo).
		Name(expectedName).
		BelongsTo(res).
		On(res).Allow(expectedTenantPerms[0]).Allow(expectedTenantPerms[1:]...).
		On(merchant1).Allow(expectedMerchant1Perms...).
		On(merchant2).Allow(expectedMerchant2Perms...).
		Build(t.Context())
	if err != nil {
		t.Error("expected a nil error")
	}

	if id.String() == (uuid.UUID{}).String() {
		t.Error("expected id to be not empty")
	}

	roles, err := repo.GetRoles(t.Context(), id)
	if err != nil {
		t.Fatalf("could not get inserted roles: %v", err)
	}

	if len(roles) != 1 {
		t.Fatalf("expected only one role: got %v roles", len(roles))
	}

	got := roles[0]

	if got.Name != expectedName {
		t.Errorf("expected role name to be '%v': got '%v'", expectedName, got.Name)
	}

	if !got.On.Equal(res) {
		t.Errorf("expected role resource to be '%v': got '%v'", res, got.On)
	}

	if len(got.Accesses) != expectedAccessesNum {
		t.Errorf("expected to have %v accesses: got %v", expectedAccessesNum, len(got.Accesses))
	}

	sortByPerms := func(a, b rbac.Access) int {
		return cmp.Compare(joinPerms(a.Perms), joinPerms(b.Perms))
	}

	slices.SortFunc(got.Accesses, sortByPerms)
	slices.SortFunc(expectedAcceses, sortByPerms)

	for i, gotAccess := range got.Accesses {
		expected := expectedAcceses[i]

		if !gotAccess.On.Equal(expected.On) {
			t.Errorf("expected access resource to be '%v': got '%v'", &expected.On, &gotAccess.On)
		}

		slices.Sort(gotAccess.Perms)
		slices.Sort(expected.Perms)

		if !slices.Equal(gotAccess.Perms, expected.Perms) {
			t.Errorf("expected accesses to be '%v': got '%v'", expected.Perms, gotAccess.Perms)
		}
	}
}

func TestBuilderSqlite3ShouldCreateRoleWithRemovedPermissions(t *testing.T) {
	t.Parallel()

	repo, _ := newRepo(t)

	res := rbac.NewResource("tenant", uuid.New(), nil)
	merchant1 := rbac.NewResource("retail", uuid.New(), nil)
	merchant2 := rbac.NewResource("retail", uuid.New(), nil)

	const (
		expectedName        = "admin"
		expectedAccessesNum = 2
	)

	var (
		expectedTenantPerms = []rbac.Perm{"tenant-mod", "tenant-del", "user-add"}
		// expectedMerchant1Perms = []rbac.Perm{"item-read"}
		expectedMerchant2Perms = []rbac.Perm{"item-read", "item-del", "item-add"}
		expectedAcceses        = []rbac.Access{
			{
				On:    res,
				Perms: expectedTenantPerms,
			},
			// Removed.
			// {
			// 	On: merchant1,
			// 	Perms: expectedMerchant1Perms,
			// },
			{
				On:    merchant2,
				Perms: expectedMerchant2Perms,
			},
		}
	)

	id, err := rbac.RoleBuilder(repo).
		Name("admin").
		BelongsTo(res).
		On(res).Allow(expectedTenantPerms[0]).Allow(expectedTenantPerms[1:]...).
		On(merchant1).Allow("item-read").
		On(merchant2).Allow(expectedMerchant2Perms...).
		Remove(merchant1).
		Build(t.Context())
	if err != nil {
		t.Error("expected a nil error")
	}

	if id.String() == (uuid.UUID{}).String() {
		t.Error("expected id to be not empty")
	}

	roles, err := repo.GetRoles(t.Context(), id)
	if err != nil {
		t.Fatalf("could not get inserted roles: %v", err)
	}

	if len(roles) != 1 {
		t.Fatalf("expected only one role: got %v roles", len(roles))
	}

	got := roles[0]

	if got.Name != expectedName {
		t.Errorf("expected role name to be '%v': got '%v'", expectedName, got.Name)
	}

	if !got.On.Equal(res) {
		t.Errorf("expected role resource to be '%v': got '%v'", res, got.On)
	}

	if len(got.Accesses) != expectedAccessesNum {
		t.Errorf("expected to have %v accesses: got %v", expectedAccessesNum, len(got.Accesses))
	}

	sortByPerms := func(a, b rbac.Access) int {
		return cmp.Compare(joinPerms(a.Perms), joinPerms(b.Perms))
	}

	slices.SortFunc(got.Accesses, sortByPerms)
	slices.SortFunc(expectedAcceses, sortByPerms)

	for i, gotAccess := range got.Accesses {
		expected := expectedAcceses[i]

		if !gotAccess.On.Equal(expected.On) {
			t.Errorf("expected access resource to be '%v': got '%v'", &expected.On, &gotAccess.On)
		}

		slices.Sort(gotAccess.Perms)
		slices.Sort(expected.Perms)

		if !slices.Equal(gotAccess.Perms, expected.Perms) {
			t.Errorf("expected accesses to be '%v': got '%v'", expected.Perms, gotAccess.Perms)
		}
	}
}
