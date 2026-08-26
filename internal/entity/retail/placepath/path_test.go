package placepath_test

import (
	"slices"
	"testing"

	"github.com/kytnacode/inventure/internal/entity/retail/placepath"
)

func TestTrimLeftPath(t *testing.T) {
	t.Parallel()

	type test struct {
		input, expected string
	}

	testData := []test{
		{
			input:    "/",
			expected: "/",
		},
		{
			input:    "/id",
			expected: "/",
		},
		{
			input:    "/one/two",
			expected: "/two",
		},
		{
			input:    "/a/b/c/d/e",
			expected: "/b/c/d/e",
		},
	}

	for _, data := range testData {
		t.Run(data.input, func(t *testing.T) {
			if got := placepath.TrimLeftPath(data.input); got != data.expected {
				t.Errorf("expected result to be '%v': got '%v'", data.expected, got)
			}
		})
	}
}

func TestCutPrefix(t *testing.T) {
	t.Parallel()

	type test struct {
		input, prefix, expected string
	}

	testData := []test{
		{
			input:    "/",
			prefix:   "/",
			expected: "/",
		},
		{
			input:    "/id",
			prefix:   "/id",
			expected: "/",
		},
		{
			input:    "/id",
			prefix:   "/",
			expected: "/id",
		},
		{
			input:    "/one/two",
			prefix:   "/one",
			expected: "/two",
		},
		{
			input:    "/a/b/c/d/e",
			prefix:   "/a/b/c",
			expected: "/d/e",
		},
	}

	for _, data := range testData {
		t.Run(data.input, func(t *testing.T) {
			if got := placepath.CutPrefix(data.input, data.prefix); got != data.expected {
				t.Errorf("expected result to be '%v': got '%v'", data.expected, got)
			}
		})
	}
}

func TestComponents(t *testing.T) {
	t.Parallel()

	type test struct {
		input    string
		expected []string
	}

	testData := []test{
		{
			input:    "/",
			expected: []string{},
		},
		{
			input:    "/id",
			expected: []string{"id"},
		},
		{
			input:    "/one/two",
			expected: []string{"one", "two"},
		},
		{
			input:    "/a/b/c/d/e",
			expected: []string{"a", "b", "c", "d", "e"},
		},
	}

	for _, data := range testData {
		t.Run(data.input, func(t *testing.T) {
			if got := placepath.Components(data.input); !slices.Equal(got, data.expected) {
				t.Errorf("expected result to be '%v': got '%v'", data.expected, got)
			}
		})
	}
}
