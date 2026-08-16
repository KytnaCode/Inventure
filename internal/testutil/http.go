package testutil

import (
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"testing"
)

// NewCookieTLSServer creates a new TLS test server and a cookie enabled HTTP client.
func NewCookieTLSServer(t *testing.T, h http.Handler) (*http.Client, *httptest.Server) {
	t.Helper()

	serv := httptest.NewTLSServer(h)

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("could not create cookie jar: %v", err)
	}

	client := serv.Client()
	client.Jar = jar

	return client, serv
}

// JoinURL joins a base URL and a list of elements, test fails on error.
func JoinURL(t *testing.T, base string, elements ...string) string {
	t.Helper()

	res, err := url.JoinPath(base, elements...)
	if err != nil {
		t.Fatalf("could not join url paths: %v", err)
	}

	return res
}
