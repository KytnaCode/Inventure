package testutil

import (
	"net/http"
	"testing"
)

// PrintStatusCode logs got and expected status code and text as returned by [http.StatusText].
func PrintStatusCode(t *testing.T, got, expected int) {
	t.Helper()

	t.Logf(
		"expected status '%v: %v': got '%v: %v'",
		expected,
		http.StatusText(expected),
		got,
		http.StatusText(got),
	)
}
