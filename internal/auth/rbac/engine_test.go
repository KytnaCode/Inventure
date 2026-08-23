package rbac_test

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/kytnacode/inventure/internal/auth/rbac"
)

type testResolver struct {
	resources map[string]map[string]rbac.Resource
}

func (r *testResolver) IsSameOrAncestor(_ context.Context, a, b rbac.Resource) (bool, error) {
	if a.Equal(b) {
		return true, nil
	}

	if b.Parent == nil {
		return false, nil
	}

	res := b

	for {
		res, ok := r.resources[res.Parent.Typ][res.Parent.ID]
		if !ok {
			return false, nil
		}

		if a.Equal(res) {
			return true, nil
		}
	}
}

func newEngine(resources map[string]map[string]rbac.Resource) *rbac.Engine {
	return rbac.NewEngine(&testResolver{
		resources: resources,
	})
}

func TestEngine_AuthorizeShouldAuthorizeNoRequiredPermissions(t *testing.T) {
	t.Parallel()

	eng := newEngine(map[string]map[string]rbac.Resource{})

	err := eng.Authorize(t.Context(), []rbac.Role{}, rbac.NewResource("project", "real-id", nil))
	if err != nil {
		t.Errorf("expected a nil error: got '%v'", err)
	}
}

func TestEngine_AuthorizeShouldReturnAnErrorOnNoApplicableRoles(t *testing.T) {
	t.Parallel()

	res := rbac.NewResource("merchant", "real-id", nil)

	const perms rbac.Perm = "item-read"

	resources := map[string]map[string]rbac.Resource{
		res.Typ: {
			res.ID: res,
		},
	}

	role := rbac.Role{
		ID:   "real-role-id",
		Name: "admin",
		On:   rbac.NewResource("non", "applicable", nil),
	}

	accesses := []rbac.Access{
		{
			ID:    "real-access-id",
			Role:  &role,
			On:    res,
			Perms: []rbac.Perm{perms},
		},
	}

	role.Accesses = accesses

	eng := newEngine(resources)

	err := eng.Authorize(t.Context(), []rbac.Role{role}, res, perms)
	if err == nil {
		t.Fatalf("expected a non-nil error: %v", err)
	}

	accessErr, ok := errors.AsType[*rbac.AccessError](err)
	if !ok {
		t.Fatalf("expected error to be an access error: got '%v'", err)
	}

	if !slices.Equal([]rbac.Perm{perms}, accessErr.MissingPermissions) {
		t.Fatalf(
			"expected missing permissions to be '%v': got '%v'",
			[]rbac.Perm{perms},
			accessErr.MissingPermissions,
		)
	}
}

func TestEngine_AuthorizeShouldReutrnAnErrorOnMissingPermissions(t *testing.T) {
	t.Parallel()

	res := rbac.NewResource("merchant", "real-id", nil)

	const missingPerm rbac.Perm = "item-del"

	requiredPermissions := []rbac.Perm{
		missingPerm,
		"item-read",
	}

	resources := map[string]map[string]rbac.Resource{
		res.Typ: {
			res.ID: res,
		},
	}

	role := rbac.Role{
		ID:   "real-role-id",
		Name: "admin",
		On:   res,
	}

	accesses := []rbac.Access{
		{
			ID:    "real-access-id",
			Role:  &role,
			On:    res,
			Perms: []rbac.Perm{"item-add", "item-read"},
		},
	}

	role.Accesses = accesses

	eng := newEngine(resources)

	err := eng.Authorize(t.Context(), []rbac.Role{role}, res, requiredPermissions...)
	if err == nil {
		t.Fatal("expected a non nil error")
	}

	accessErr, ok := errors.AsType[*rbac.AccessError](err)
	if !ok {
		t.Fatalf("expected error to be an access error: got '%v'", err)
	}

	if !slices.Equal([]rbac.Perm{missingPerm}, accessErr.MissingPermissions) {
		t.Errorf(
			"expected missing permissions to be '%v': got '%v'",
			[]rbac.Perm{missingPerm},
			accessErr.MissingPermissions,
		)
	}
}

func TestEngine_AuthorizeShouldNotReturnErrorOnDirectlyRequiredPermissions(t *testing.T) {
	t.Parallel()

	res := rbac.NewResource("merchant", "real-id", nil)

	requiredPermissions := []rbac.Perm{
		"item-add",
		"item-del",
		"item-read",
	}

	resources := map[string]map[string]rbac.Resource{
		res.Typ: {
			res.ID: res,
		},
	}

	role := rbac.Role{
		ID:   "real-role-id",
		Name: "admin",
		On:   res,
	}

	accesses := []rbac.Access{
		{
			ID:    "real-access-id",
			Role:  &role,
			On:    res,
			Perms: requiredPermissions,
		},
	}

	role.Accesses = accesses

	eng := newEngine(resources)

	err := eng.Authorize(t.Context(), []rbac.Role{role}, res, requiredPermissions...)
	if err != nil {
		t.Fatalf("expected a nil error: %v", err)
	}
}

func TestEngine_AuthorizeShouldNotReturnErrorWhenMultipleRolesHaveRequiredPerms(t *testing.T) {
	t.Parallel()

	res := rbac.NewResource("merchant", "real-id", nil)

	roleAPerms := []rbac.Perm{
		"item-read",
	}

	roleBPerms := []rbac.Perm{
		"item-add",
		"item-del",
	}

	requiredPermissions := slices.Concat(roleAPerms, roleBPerms)

	resources := map[string]map[string]rbac.Resource{
		res.Typ: {
			res.ID: res,
		},
	}

	roleA := rbac.Role{
		ID:   "role-a-id",
		Name: "role-a",
		On:   res,
	}

	roleAAcceses := []rbac.Access{
		{
			ID:    "access-a-id",
			On:    res,
			Role:  &roleA,
			Perms: roleAPerms,
		},
	}

	roleA.Accesses = roleAAcceses

	roleB := rbac.Role{
		ID:   "role-b-id",
		Name: "role-b",
		On:   res,
	}

	roleBAcceses := []rbac.Access{
		{
			ID:    "access-b-id",
			On:    res,
			Role:  &roleB,
			Perms: roleBPerms,
		},
	}

	roleB.Accesses = roleBAcceses

	eng := newEngine(resources)

	err := eng.Authorize(t.Context(), []rbac.Role{roleA, roleB}, res, requiredPermissions...)
	if err != nil {
		t.Errorf("expected a nil error: %v", err)
	}
}

func TestEngine_AuthorizeShouldInheritPermissions(t *testing.T) {
	t.Parallel()

	resParent := rbac.NewResource("tenant", "real-tenant-id", nil)
	resChild := rbac.NewResource("merchant", "real-merchant-id", &resParent)

	requiredPermissions := []rbac.Perm{
		"item-add",
		"item-del",
		"item-read",
	}

	resources := map[string]map[string]rbac.Resource{
		resParent.Typ: {
			resParent.ID: resParent,
		},
		resChild.Typ: {
			resChild.ID: resChild,
		},
	}

	role := rbac.Role{
		ID:   "real-role-id",
		Name: "admin",
		On:   resParent,
	}

	accesses := []rbac.Access{
		{
			ID:    "real-access-id",
			Role:  &role,
			On:    resParent,
			Perms: requiredPermissions,
		},
	}

	role.Accesses = accesses

	eng := newEngine(resources)

	err := eng.Authorize(t.Context(), []rbac.Role{role}, resChild, requiredPermissions...)
	if err != nil {
		t.Errorf("expected a nil error: got '%v'", err)
	}
}

func TestEngine_AuthorizeShouldReturnErrorWhenNoAccess(t *testing.T) {
	t.Parallel()

	res := rbac.NewResource("merchant", "real-id", nil)

	requiredPermissions := []rbac.Perm{
		"item-add",
		"item-del",
		"item-read",
	}

	resources := map[string]map[string]rbac.Resource{
		res.Typ: {
			res.ID: res,
		},
	}

	role := rbac.Role{
		ID:   "real-role-id",
		Name: "admin",
		On:   res,
	}

	role.Accesses = make([]rbac.Access, 0)

	eng := newEngine(resources)

	err := eng.Authorize(t.Context(), []rbac.Role{role}, res, requiredPermissions...)
	if err == nil {
		t.Fatalf("expected a non-nil error: %v", err)
	}

	accessErr, ok := errors.AsType[*rbac.AccessError](err)
	if !ok {
		t.Fatalf("expected error to be an access error: %v", err)
	}

	slices.Sort(requiredPermissions)
	slices.Sort(accessErr.MissingPermissions)

	if !slices.Equal(requiredPermissions, accessErr.MissingPermissions) {
		t.Errorf(
			"expected missing permissions to be '%v': got '%v'",
			requiredPermissions,
			accessErr.MissingPermissions,
		)
	}
}
