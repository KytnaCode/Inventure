package routes_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path"
	"testing"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/alexedwards/scs/v2/memstore"
	"github.com/kytnacode/inventure/internal/auth/routes"
	"github.com/kytnacode/inventure/internal/auth/session"
	"github.com/kytnacode/inventure/internal/testutil"
	userrepository "github.com/kytnacode/inventure/internal/user/repository"
	"github.com/kytnacode/inventure/internal/web"
	"github.com/kytnacode/inventure/pkg/api"
	"github.com/kytnacode/inventure/pkg/passhash"
	"github.com/kytnacode/inventure/pkg/validation"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

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

	conf := &routes.Config{
		Validator:             v,
		SessionManager:        sessionManager,
		RequestLimit:          10,
		TimeWindow:            time.Minute,
		UserRepo:              userRepo,
		LoginAttemptLimit:     5,
		LoginAttempTimeWindow: time.Minute,
	}

	return routes.New(conf), g, &session
}

func TestRoutes_SignUpShouldStoreUser(t *testing.T) {
	t.Parallel()

	ro, g, sessionManager := newRoutes(t)

	data := routes.SignUpData{
		Name:     "my user name",
		Email:    "my-user@email.com",
		Password: "abcde",
	}

	body := testutil.EncodeBody(t, data)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/v1/auth/login", body)

	w := httptest.NewRecorder()

	sessionManager.LoadAndSave(http.HandlerFunc(ro.SignUp)).ServeHTTP(w, req)

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

	body := testutil.EncodeBody(t, data)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/v1/auth/signup", body)

	w := httptest.NewRecorder()

	ro, _, sessionManager := newRoutes(t)

	sessionManager.LoadAndSave(http.HandlerFunc(ro.SignUp)).ServeHTTP(w, req)

	res := w.Result()

	if got := res.StatusCode; got != expectedStatusCode {
		testutil.PrintStatusCode(t, got, expectedStatusCode)

		t.Fail()
	}

	resp := new(api.Response)

	testutil.DecodeBody(t, res.Body, resp)

	if got := resp.Status; got != string(api.StatusFail) {
		t.Fatalf("expected a fail response: got 'status: %v'", got)
	}
}

func TestRoutes_SignUpShouldStoreSessionData(t *testing.T) {
	t.Parallel()

	ro, _, sessionManager := newRoutes(t)

	body := testutil.EncodeBody(t, routes.SignUpData{
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

func TestRoutes_SignInShouldReturnUserNotFound(t *testing.T) {
	t.Parallel()

	const (
		expectedStatusCode     = http.StatusNotFound
		expectedResponseStatus = api.StatusError
	)

	expectedErrorCode := web.CodeUserNotFound

	ro, _, sessionManager := newRoutes(t)

	body := testutil.EncodeBody(t, routes.SignInData{
		Email:    "non-existing@email.com",
		Password: "random-password123",
	})

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/v1/auth/signin", body)

	w := httptest.NewRecorder()

	sessionManager.LoadAndSave(http.HandlerFunc(ro.SignIn)).ServeHTTP(w, req)

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

	if resp.Code == nil {
		t.Fatal("expected a non-nil error code")
	}

	code := *resp.Code

	if code != *expectedErrorCode {
		t.Errorf("expected error code to be '%v': got '%v'", *expectedErrorCode, code)
	}
}

func TestRoutes_SignInShouldReturnNoPasswordAuthError(t *testing.T) {
	t.Parallel()

	const (
		expectedStatusCode     = http.StatusBadRequest
		expectedResponseStatus = api.StatusError
	)

	expectedErrorCode := web.CodeNoPasswordAuth

	const userEmail = "my-real@email.com"

	ro, g, sessionManager := newRoutes(t)

	u := &userrepository.User{
		Email: userEmail,
		Name:  "username",
		// No password hash.
		// PasswordHash:
	}

	if err := g.Create(t.Context(), u); err != nil {
		t.Fatalf("could not create test user: %v", err)
	}

	body := testutil.EncodeBody(t, routes.SignInData{
		Email:    userEmail,
		Password: "random-password",
	})

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/v1/auth/signin", body)

	w := httptest.NewRecorder()

	sessionManager.LoadAndSave(http.HandlerFunc(ro.SignIn)).ServeHTTP(w, req)

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

	if resp.Code == nil {
		t.Fatal("expected a non-nil error code")
	}

	code := *resp.Code

	if code != *expectedErrorCode {
		t.Errorf("expected error code to be '%v': got '%v'", *expectedErrorCode, code)
	}
}

func TestRoutes_SignInShouldReturnWrongCredentialsError(t *testing.T) {
	t.Parallel()

	const (
		expectedStatusCode     = http.StatusUnauthorized
		expectedResponseStatus = api.StatusError
	)

	expectedErrorCode := web.CodeWrongCredentials

	const (
		userEmail    = "diana.cavendish@email.com"
		userPassword = "iloveakko"
	)

	ro, g, sessionManager := newRoutes(t)

	otherPass := passhash.Hash([]byte("other-password"))

	u := &userrepository.User{
		Email:        userEmail,
		PasswordHash: &otherPass,
	}

	if err := g.Create(t.Context(), u); err != nil {
		t.Fatalf("could not create test user: %v", err)
	}

	body := testutil.EncodeBody(t, routes.SignInData{
		Email:    userEmail,
		Password: userPassword,
	})

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/v1/auth/signin", body)

	w := httptest.NewRecorder()

	sessionManager.LoadAndSave(http.HandlerFunc(ro.SignIn)).ServeHTTP(w, req)

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

	if resp.Code == nil {
		t.Fatal("expected a non-nil error code")
	}

	code := *resp.Code

	if code != *expectedErrorCode {
		t.Errorf("expected error code to be '%v': got '%v'", code, *expectedErrorCode)
	}
}

func TestRoutes_SignInShouldStoreUserInSession(t *testing.T) {
	t.Parallel()

	const expectedStatusCode = http.StatusOK

	const (
		userEmail = "amity.blight@toh.com"
		userPass  = "ilovemygf"
	)

	ro, g, sessionManager := newRoutes(t)

	passwordHash := passhash.Hash([]byte(userPass))

	u := &userrepository.User{
		Email:        userEmail,
		PasswordHash: &passwordHash,
		Name:         "Amity Blight",
	}

	if err := g.Create(t.Context(), u); err != nil {
		t.Fatalf("could not create test user: %v", err)
	}

	body := testutil.EncodeBody(t, routes.SignInData{
		Email:    userEmail,
		Password: userPass,
	})

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/v1/auth/signin", body)

	w := httptest.NewRecorder()

	sessionManager.LoadAndSave(http.HandlerFunc(ro.SignIn)).ServeHTTP(w, req)

	res := w.Result()

	if got := res.StatusCode; got != expectedStatusCode {
		testutil.PrintStatusCode(t, got, expectedStatusCode)

		t.Fail()
	}

	var found bool

	err := sessionManager.Iterate(t.Context(), func(sessCtx context.Context) error {
		raw := sessionManager.Get(sessCtx, session.KeySessionData)

		data, ok := raw.(*session.Session)
		if !ok {
			t.Errorf("wrong session data type: got: '%T: %v'", raw, raw)
		}

		if data == nil {
			t.Fatalf("expected non-nil session data")
		}

		if data.ID != u.ID.String() {
			t.Fatalf("wrong user id: expected '%v' got '%v'", u.ID, data.ID)
		}

		found = true

		return nil
	})
	if err != nil {
		t.Fatalf("could not iterate for sessions: %v", err)
	}

	if !found {
		t.Error("no sessions found")
	}
}
