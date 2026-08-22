package sqltypes_test

import (
	"slices"
	"testing"

	"github.com/kytnacode/inventure/sqltypes"
)

func TestListShouldEncodeAsACommaSeparatedString(t *testing.T) {
	t.Parallel()

	const expected = "a,b,c"

	l := sqltypes.List{
		"a",
		"b",
		"c",
	}

	v, err := l.Value()
	if err != nil {
		t.Fatalf("could not encode list value: %v", err)
	}

	strValue, ok := v.(string)
	if !ok {
		t.Fatal("expected list to encode as a string")
	}

	if strValue != expected {
		t.Fatalf("expected encoded string to be '%v': got '%v'", expected, strValue)
	}
}

func TestListShouldDecodeAStringOfCommaSeparatedValues(t *testing.T) {
	t.Parallel()

	const sqlValue = "1,6,4,7"

	expected := sqltypes.List{
		"1",
		"6",
		"4",
		"7",
	}

	var got sqltypes.List

	err := got.Scan(sqlValue)
	if err != nil {
		t.Fatalf("could not scan value: %v", err)
	}

	if !slices.Equal(got, expected) {
		t.Fatalf("expected result to be '%v': got '%v'", expected, got)
	}
}
