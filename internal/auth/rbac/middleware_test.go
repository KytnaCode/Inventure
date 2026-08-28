package rbac_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/kytnacode/inventure/api"
	"github.com/kytnacode/inventure/internal/auth/rbac"
	"github.com/kytnacode/inventure/internal/testutil"
	"github.com/kytnacode/inventure/internal/web"
)

type testRoleGetter struct {
	roles []rbac.Role
	err   error
}

func (r *testRoleGetter) GetRoles(_ context.Context, _ ...uuid.UUID) ([]rbac.Role, error) {
	if r.err != nil {
		return nil, r.err
	}

	return r.roles, nil
}

func newMiddleware(
	resources map[string]map[string]rbac.Resource,
	roleGetter *testRoleGetter,
	extractor api.Extractor[*rbac.AuthData],
) *rbac.Middleware {
	resolver := testResolver{
		resources: resources,
	}

	eng := rbac.NewEngine(&resolver)

	m := rbac.NewMiddleware(eng, roleGetter, extractor)

	return m
}

func TestMiddleware_RequirePermsShouldReturnInternalServerErrorOnExtractorError(t *testing.T) {
	t.Parallel()

	const (
		expectedStatus         = http.StatusInternalServerError
		expectedResponseStatus = api.StatusError
	)

	res := rbac.NewResource("tenant", uuid.New(), nil)

	perms := []rbac.Perm{"user-read", "user-add"}

	resources := map[string]map[string]rbac.Resource{
		res.Typ: {
			res.ID.String(): res,
		},
	}

	adminRole := rbac.Role{
		ID:       uuid.New(),
		Name:     "admin",
		On:       res,
		Accesses: make([]rbac.Access, 0, 1),
	}

	access := rbac.Access{
		ID:    uuid.New(),
		On:    res,
		Role:  &adminRole,
		Perms: perms,
	}

	adminRole.Accesses = append(adminRole.Accesses, access)

	roleGetter := &testRoleGetter{
		roles: []rbac.Role{
			adminRole,
		},
	}

	errorExtractor := func(_ context.Context) (*rbac.AuthData, error) {
		return nil, errors.New("expected error")
	}

	m := newMiddleware(resources, roleGetter, errorExtractor)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/protected", nil)

	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {})

	m.RequirePerms(perms...)(handler).ServeHTTP(w, req)

	result := w.Result()

	if got := result.StatusCode; got != expectedStatus {
		testutil.PrintStatusCode(t, got, expectedStatus)

		t.Fail()
	}

	resp := new(api.Response)

	testutil.DecodeBody(t, result.Body, resp)

	if got := resp.Status; got != string(expectedResponseStatus) {
		t.Errorf("expected response status to be '%v': got '%v'", string(expectedResponseStatus), got)
	}
}

func TestMiddleware_RequirePermsShouldReturnInternalServerErrorOnRepositoryError(t *testing.T) {
	t.Parallel()

	const (
		expectedStatus         = http.StatusInternalServerError
		expectedResponseStatus = api.StatusError
	)

	res := rbac.NewResource("tenant", uuid.New(), nil)

	perms := []rbac.Perm{"user-read", "user-add"}

	resources := map[string]map[string]rbac.Resource{
		res.Typ: {
			res.ID.String(): res,
		},
	}

	adminRole := rbac.Role{
		ID:       uuid.New(),
		Name:     "admin",
		On:       res,
		Accesses: make([]rbac.Access, 0, 1),
	}

	access := rbac.Access{
		ID:    uuid.New(),
		On:    res,
		Role:  &adminRole,
		Perms: perms,
	}

	adminRole.Accesses = append(adminRole.Accesses, access)

	roleGetter := &testRoleGetter{
		err: errors.New("repository error"),
	}

	errorExtractor := func(_ context.Context) (*rbac.AuthData, error) {
		return &rbac.AuthData{
			RoleIDs:  []uuid.UUID{adminRole.ID},
			Resource: res,
		}, nil
	}

	m := newMiddleware(resources, roleGetter, errorExtractor)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/protected", nil)

	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {})

	m.RequirePerms(perms...)(handler).ServeHTTP(w, req)

	result := w.Result()

	if got := result.StatusCode; got != expectedStatus {
		testutil.PrintStatusCode(t, got, expectedStatus)

		t.Fail()
	}

	resp := new(api.Response)

	testutil.DecodeBody(t, result.Body, resp)

	if got := resp.Status; got != string(expectedResponseStatus) {
		t.Errorf("expected response status to be '%v': got '%v'", string(expectedResponseStatus), got)
	}
}

func TestMiddleware_RequirePermsShouldReturnForbiddenOnMissingPermissions(t *testing.T) {
	t.Parallel()

	var (
		expectedStatus         = http.StatusForbidden
		expectedResponseStatus = api.StatusError
		expectedCode           = web.CodeMissingPerms
	)

	res := rbac.NewResource("tenant", uuid.New(), nil)

	perms := []rbac.Perm{"user-read", "user-add"}

	resources := map[string]map[string]rbac.Resource{
		res.Typ: {
			res.ID.String(): res,
		},
	}

	adminRole := rbac.Role{
		ID:       uuid.New(),
		Name:     "admin",
		On:       res,
		Accesses: make([]rbac.Access, 0, 1),
	}

	access := rbac.Access{
		ID:   uuid.New(),
		On:   res,
		Role: &adminRole,
		// Perms: perms, // No perms
	}

	adminRole.Accesses = append(adminRole.Accesses, access)

	roleGetter := &testRoleGetter{
		roles: []rbac.Role{adminRole},
	}

	errorExtractor := func(_ context.Context) (*rbac.AuthData, error) {
		return &rbac.AuthData{
			RoleIDs:  []uuid.UUID{adminRole.ID},
			Resource: res,
		}, nil
	}

	m := newMiddleware(resources, roleGetter, errorExtractor)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/protected", nil)

	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {})

	m.RequirePerms(perms...)(handler).ServeHTTP(w, req)

	result := w.Result()

	if got := result.StatusCode; got != expectedStatus {
		testutil.PrintStatusCode(t, got, expectedStatus)

		t.Fail()
	}

	resp := new(api.Response)

	testutil.DecodeBody(t, result.Body, resp)

	if got := resp.Status; got != string(expectedResponseStatus) {
		t.Errorf("expected response status to be '%v': got '%v'", string(expectedResponseStatus), got)
	}

	if resp.Code == nil {
		t.Fatal("expected a non-nil code error")
	}

	if got := *resp.Code; got != *expectedCode {
		t.Errorf("expected error code to be '%v': got '%v'", *expectedCode, got)
	}
}

func TestMiddleware_RequirePermsShouldPassOnCorrectPermissions(t *testing.T) {
	t.Parallel()

	expectedStatus := http.StatusOK

	res := rbac.NewResource("tenant", uuid.New(), nil)

	perms := []rbac.Perm{"user-read", "user-add"}

	resources := map[string]map[string]rbac.Resource{
		res.Typ: {
			res.ID.String(): res,
		},
	}

	adminRole := rbac.Role{
		ID:       uuid.New(),
		Name:     "admin",
		On:       res,
		Accesses: make([]rbac.Access, 0, 1),
	}

	access := rbac.Access{
		ID:    uuid.New(),
		On:    res,
		Role:  &adminRole,
		Perms: perms,
	}

	adminRole.Accesses = append(adminRole.Accesses, access)

	roleGetter := &testRoleGetter{
		roles: []rbac.Role{adminRole},
	}

	errorExtractor := func(_ context.Context) (*rbac.AuthData, error) {
		return &rbac.AuthData{
			RoleIDs:  []uuid.UUID{adminRole.ID},
			Resource: res,
		}, nil
	}

	m := newMiddleware(resources, roleGetter, errorExtractor)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/protected", nil)

	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {})

	m.RequirePerms(perms...)(handler).ServeHTTP(w, req)

	result := w.Result()

	if got := result.StatusCode; got != expectedStatus {
		testutil.PrintStatusCode(t, got, expectedStatus)

		t.Fail()
	}
}
