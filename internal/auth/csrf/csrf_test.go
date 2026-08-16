package csrf_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/kytnacode/inventure/internal/auth/csrf"
	"github.com/kytnacode/inventure/internal/testutil"
	"github.com/kytnacode/inventure/internal/web"
	"github.com/kytnacode/inventure/pkg/api"
)

func newSessionManager() *scs.SessionManager {
	m := scs.New()
	m.Lifetime = time.Hour
	m.Cookie.Path = "/"
	m.Cookie.HttpOnly = false
	m.Cookie.SameSite = http.SameSiteLaxMode
	m.Cookie.Secure = true

	return m
}

func TestRequireCSRFShouldReturnMissingSessionToken(t *testing.T) {
	t.Parallel()

	const (
		expectedStatusCode     = http.StatusForbidden
		expectedResponseStatus = api.StatusError
	)

	expectedErrorCode := web.CodeMissingCSRFTokenSession

	m := newSessionManager()

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/state", nil)

	w := httptest.NewRecorder()

	m.LoadAndSave(csrf.RequireCSRF(m)(handler)).ServeHTTP(w, req)

	res := w.Result()

	if got := res.StatusCode; got != expectedStatusCode {
		testutil.PrintStatusCode(t, got, expectedStatusCode)

		t.Fail()
	}

	resp := new(api.Response)

	testutil.DecodeBody(t, res.Body, resp)

	if got := resp.Status; got != string(expectedResponseStatus) {
		t.Errorf(
			"expected response with status '%v': got '%v'",
			string(expectedResponseStatus),
			resp.Status,
		)
	}

	if resp.Code == nil {
		t.Fatal("expected a non-nil error code")
	}

	code := *resp.Code

	if code != *expectedErrorCode {
		t.Errorf("expected code '%v': got '%v'", *expectedErrorCode, code)
	}
}

func TestRequireCSRFShouldReturnErrorOnMissingTokenHeader(t *testing.T) {
	t.Parallel()

	const (
		expectedStatusCode     = http.StatusForbidden
		expectedResponseStatus = api.StatusError
	)

	m := newSessionManager()

	mux := http.NewServeMux()

	mux.HandleFunc("GET /csrf", func(_ http.ResponseWriter, r *http.Request) {
		csrf.InjectToken(m, r)
	})

	// Automatically return 200 status.
	handler := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {})

	mux.Handle("POST /", csrf.RequireCSRF(m)(handler))

	client, serv := testutil.NewCookieTLSServer(t, m.LoadAndSave(mux))
	defer serv.Close()

	csrfURL := testutil.JoinURL(t, serv.URL, "/csrf")

	_, err := client.Get(csrfURL)
	if err != nil {
		t.Fatalf("could not make csrf request to test server: %v", err)
	}

	res, err := client.Post(serv.URL, "", nil)
	if err != nil {
		t.Fatalf("could not make request to test server: %v", err)
	}

	if got := res.StatusCode; got != expectedStatusCode {
		testutil.PrintStatusCode(t, got, expectedStatusCode)

		t.Fail()
	}

	resp := new(api.Response)

	testutil.DecodeBody(t, res.Body, resp)

	if got := resp.Status; got != string(expectedResponseStatus) {
		t.Errorf("expected status '%v': got '%v'", string(expectedResponseStatus), got)
	}

	if resp.Code != nil {
		t.Errorf("expected a nil error code: got '%v'", *resp.Code)
	}
}
