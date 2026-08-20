package rbac_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kytnacode/inventure/internal/auth/rbac"
	"github.com/kytnacode/inventure/internal/testutil"
	"github.com/kytnacode/inventure/pkg/api"
)

type testEngine struct {
	role       *rbac.Role
	resID      string
	shouldAuth bool
	authError  error
}

func (e *testEngine) RoleFrom(_ *http.Request) (*rbac.Role, error) {
	if e.role == nil {
		return nil, fmt.Errorf("example error")
	}

	return e.role, nil
}

func (e *testEngine) ResourceIDFrom(_ *http.Request) (string, error) {
	if e.resID == "" {
		return "", fmt.Errorf("example error")
	}

	return e.resID, nil
}

func (e *testEngine) Authorize(
	_ context.Context,
	_ *rbac.Role,
	_ rbac.Resource,
	_ ...rbac.Perm,
) (bool, error) {
	if e.authError != nil {
		return false, e.authError
	}

	return e.shouldAuth, nil
}

var testPerm rbac.Perm = "test-perm"

func newMiddleware(eng rbac.Engine) *rbac.Middleware {
	return rbac.NewMiddleware(eng)
}

func TestRequirePermsShouldReturnUnauthorizedResponse(t *testing.T) {
	t.Parallel()

	const (
		expectedStatusCode     = http.StatusUnauthorized
		expectedResponseStatus = api.StatusError
	)

	m := newMiddleware(&testEngine{})

	handler := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {})

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/resource", nil)

	w := httptest.NewRecorder()

	m.RequirePerms("test", testPerm)(handler).ServeHTTP(w, req)

	res := w.Result()

	if got := res.StatusCode; got != expectedStatusCode {
		testutil.PrintStatusCode(t, got, expectedStatusCode)

		t.Fail()
	}

	resp := new(api.Response)

	testutil.DecodeBody(t, res.Body, resp)

	if got := resp.Status; got != string(expectedResponseStatus) {
		t.Errorf("expected status '%v': got '%v'", string(expectedResponseStatus), got)
	}
}

func TestRequirePermsShouldReturnErrorOnMissingResourceID(t *testing.T) {
	t.Parallel()

	const (
		expectedStatusCode     = http.StatusInternalServerError
		expectedResponseStatus = api.StatusError
	)

	m := newMiddleware(&testEngine{role: &rbac.Role{Allow: []rbac.Perm{testPerm}}})

	handler := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {})

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/resource", nil)

	w := httptest.NewRecorder()

	m.RequirePerms("test", testPerm)(handler).ServeHTTP(w, req)

	res := w.Result()

	if got := res.StatusCode; got != expectedStatusCode {
		testutil.PrintStatusCode(t, got, expectedStatusCode)

		t.Fail()
	}

	resp := new(api.Response)

	testutil.DecodeBody(t, res.Body, resp)

	if got := resp.Status; got != string(expectedResponseStatus) {
		t.Errorf("expected status '%v': got '%v'", string(expectedResponseStatus), got)
	}
}

func TestRequirePermsShouldReturnErrorOnFailedAuthorization(t *testing.T) {
	t.Parallel()

	const (
		expectedStatusCode     = http.StatusInternalServerError
		expectedResponseStatus = api.StatusError
	)

	m := newMiddleware(&testEngine{
		role:      &rbac.Role{Allow: []rbac.Perm{testPerm}},
		resID:     "real-id",
		authError: fmt.Errorf("random error"),
	})

	handler := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {})

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/resource", nil)

	w := httptest.NewRecorder()

	m.RequirePerms("test", testPerm)(handler).ServeHTTP(w, req)

	res := w.Result()

	if got := res.StatusCode; got != expectedStatusCode {
		testutil.PrintStatusCode(t, got, expectedStatusCode)

		t.Fail()
	}

	resp := new(api.Response)

	testutil.DecodeBody(t, res.Body, resp)

	if got := resp.Status; got != string(expectedResponseStatus) {
		t.Errorf("expected status '%v': got '%v'", string(expectedResponseStatus), got)
	}
}

func TestRequirePermsShouldReturnForbidden(t *testing.T) {
	t.Parallel()

	const (
		expectedStatusCode     = http.StatusForbidden
		expectedResponseStatus = api.StatusError
	)

	m := newMiddleware(&testEngine{
		role:       &rbac.Role{Allow: []rbac.Perm{testPerm}},
		resID:      "real-id",
		shouldAuth: false,
	})

	handler := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {})

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/resource", nil)

	w := httptest.NewRecorder()

	m.RequirePerms("test", testPerm)(handler).ServeHTTP(w, req)

	res := w.Result()

	if got := res.StatusCode; got != expectedStatusCode {
		testutil.PrintStatusCode(t, got, expectedStatusCode)

		t.Fail()
	}

	resp := new(api.Response)

	testutil.DecodeBody(t, res.Body, resp)

	if got := resp.Status; got != string(expectedResponseStatus) {
		t.Errorf("expected status '%v': got '%v'", string(expectedResponseStatus), got)
	}
}

func TestRequirePermsShouldAllowRequest(t *testing.T) {
	t.Parallel()

	const expectedStatusCode = http.StatusOK

	m := newMiddleware(&testEngine{
		role:       &rbac.Role{Allow: []rbac.Perm{testPerm}},
		resID:      "real-id",
		shouldAuth: true,
	})

	handler := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {})

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/resource", nil)

	w := httptest.NewRecorder()

	m.RequirePerms("test", testPerm)(handler).ServeHTTP(w, req)

	res := w.Result()

	if got := res.StatusCode; got != expectedStatusCode {
		testutil.PrintStatusCode(t, got, expectedStatusCode)

		t.Fail()
	}
}
