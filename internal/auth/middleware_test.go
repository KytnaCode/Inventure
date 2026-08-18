package auth_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/kytnacode/inventure/internal/auth"
	"github.com/kytnacode/inventure/internal/testutil"
)

func newSessionManager() *scs.SessionManager {
	m := scs.New()
	m.Lifetime = time.Minute

	return m
}

func TestRequireAuthShouldReturnUnauthorizedResponse(t *testing.T) {
	t.Parallel()

	const (
		expectedStatusCode = http.StatusUnauthorized
	)

	m := newSessionManager()

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/protected", nil)

	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {})

	m.LoadAndSave(auth.RequireAuth(m)(handler)).ServeHTTP(w, req)

	res := w.Result()

	if got := res.StatusCode; got != expectedStatusCode {
		testutil.PrintStatusCode(t, got, expectedStatusCode)

		t.Fail()
	}
}

func TestRequireAuthShouldPassNextHandler(t *testing.T) {
	t.Parallel()

	const expectedStatusCode = http.StatusOK

	m := newSessionManager()

	mux := http.NewServeMux()

	mux.HandleFunc("POST /login", func(_ http.ResponseWriter, r *http.Request) {
		m.Put(r.Context(), auth.KeySessionData, &auth.Session{
			ID: "my-super-real-id",
		})
	})

	handler := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {})

	mux.Handle("GET /", auth.RequireAuth(m)(handler))

	client, serv := testutil.NewCookieTLSServer(t, m.LoadAndSave(mux))
	defer serv.Close()

	loginEndpoint := testutil.JoinURL(t, serv.URL, "/login")

	_, err := client.Post(loginEndpoint, "", nil)
	if err != nil {
		t.Fatalf("could not send request to login endpoint: %v", err)
	}

	res, err := client.Get(serv.URL)
	if err != nil {
		t.Fatalf("could not send request to protected endpoint: %v", err)
	}

	if got := res.StatusCode; got != expectedStatusCode {
		testutil.PrintStatusCode(t, got, expectedStatusCode)

		t.Fail()
	}
}
