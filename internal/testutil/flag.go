package testutil

import "testing"

// Integration marks current test as an integration test.
func Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
}
