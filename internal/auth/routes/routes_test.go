package routes_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path"
	"testing"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/alexedwards/scs/v2/memstore"
	"github.com/kytnacode/inventure/internal/auth/routes"
	"github.com/kytnacode/inventure/internal/auth/session"
	userrepository "github.com/kytnacode/inventure/internal/user/repository"
	"github.com/kytnacode/inventure/pkg/api"
	"github.com/kytnacode/inventure/pkg/validation"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const redirectLocation = "/"

func newRoutes(t *testing.T) (
	*routes.Routes,
	gorm.Interface[userrepository.User],
	*scs.SessionManager,
) {
	t.Helper()

	sessionManager := scs.New()
	sessionManager.Store = memstore.New()
	sessionManager.Lifetime = time.Minute * 10
	sessionManager.Cookie.HttpOnly = false
	sessionManager.Cookie.Secure = false

	v := validation.New()

	db, err := gorm.Open(sqlite.Open(path.Join(t.TempDir(), "app.db")))
	if err != nil {
		t.Fatalf("could not create a test database: %v", err)
	}

	if err := db.AutoMigrate(userrepository.User{}); err != nil {
		t.Fatalf("could not migrate test database: %v", err)
	}

	g := gorm.G[userrepository.User](db)

	userRepo := userrepository.New(g, v)

	session := *sessionManager

	return routes.New(userRepo, &session, v, redirectLocation), g, &session
}

func encodeData(t *testing.T, data any) *bytes.Buffer {
	t.Helper()

	var buf bytes.Buffer

	if err := json.NewEncoder(&buf).Encode(data); err != nil {
		t.Fatalf("could not encode body: %v", err)
	}

	return &buf
}

func TestRoutes_SignUpShouldStoreUser(t *testing.T) {
	t.Parallel()

	ro, g, sessionManager := newRoutes(t)

	data := routes.SignUpData{
		Name:     "my user name",
		Email:    "my-user@email.com",
		Password: "abcde",
	}

	body := encodeData(t, data)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/v1/auth/login", body)

	w := httptest.NewRecorder()

	sessionManager.LoadAndSave(http.HandlerFunc(ro.SignUp)).ServeHTTP(w, req)

	res := w.Result()

	if got := res.Header.Get("Location"); got != redirectLocation {
		t.Errorf("expected location to be '%v': got '%v'", redirectLocation, got)
	}

	got, err := g.Where("email = ?", data.Email).First(t.Context())
	if err != nil {
		t.Fatalf("could not get created user: %v", err)
	}

	if got.Name != data.Name {
		t.Errorf("expected name to be '%v': got '%v'", data.Name, got.Name)
	}
}

func TestRoutes_SignUpShouldValidateData(t *testing.T) {
	t.Parallel()

	const (
		expectedStatusCode = http.StatusBadRequest
	)

	data := routes.SignUpData{
		// Missing name.
		Email: "invalid-email",
		// Missing password.
	}

	body := encodeData(t, data)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/v1/auth/signup", body)

	w := httptest.NewRecorder()

	ro, _, sessionManager := newRoutes(t)

	sessionManager.LoadAndSave(http.HandlerFunc(ro.SignUp)).ServeHTTP(w, req)

	res := w.Result()
	defer func() {
		if err := res.Body.Close(); err != nil {
			t.Fatalf("could not close body: %v", err)
		}
	}()

	if got := res.StatusCode; got != expectedStatusCode {
		t.Errorf("expected status '%v': got '%v'", expectedStatusCode, got)
	}

	var resp api.Response

	dec := json.NewDecoder(res.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&resp); err != nil {
		t.Fatalf("could not decode response body: %v", err)
	}

	if got := resp.Status; got != string(api.StatusFail) {
		t.Fatalf("expected a fail response: got 'status: %v'", got)
	}
}

func TestRoutes_SignUpShouldStoreSessionData(t *testing.T) {
	t.Parallel()

	ro, _, sessionManager := newRoutes(t)

	body := encodeData(t, routes.SignUpData{
		Name:     "my valid user name",
		Email:    "my-valid@email.com",
		Password: "my-super-secret-and-valid-password",
	})

	ctx := t.Context()

	req := httptest.NewRequestWithContext(ctx, http.MethodPost, "/api/v1/auth/signup", body)

	w := httptest.NewRecorder()

	sessionManager.LoadAndSave(http.HandlerFunc(ro.SignUp)).ServeHTTP(w, req)

	var found bool

	err := sessionManager.Iterate(ctx, func(sessCtx context.Context) error {
		data, ok := sessionManager.Get(sessCtx, session.KeySessionData).(*session.Session)
		if !ok || data == nil {
			t.Fatal("could not find stored session data")
		}

		if data.ID == "" {
			t.Fatal("expected a non empty user id in session data")
		}

		found = true

		return nil
	})
	if err != nil {
		t.Fatalf("could not iterate for sessions: %v", err)
	}

	if !found {
		t.Fatal("no sessions found in store")
	}
}
