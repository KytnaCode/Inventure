package testutil

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"
)

// DecodeBody decodes JSON data from response body and then closes it. Body is decoded as
// using [json.Decoder.DisallowUnknownFields]. If an error occurs test fails immediately.
func DecodeBody(t *testing.T, body io.ReadCloser, data any) {
	t.Helper()

	defer func() {
		if err := body.Close(); err != nil {
			t.Fatalf("could not close response body: %v", err)
		}
	}()

	dec := json.NewDecoder(body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(data); err != nil {
		t.Fatalf("could not decode body: %v", err)
	}
}

// EncodeBody encodes data into a JSON buffer. If an error occurs test fails immediately.
func EncodeBody(t *testing.T, data any) *bytes.Buffer {
	t.Helper()

	var buf bytes.Buffer

	if err := json.NewEncoder(&buf).Encode(data); err != nil {
		t.Fatalf("could not encode body: %v", err)
	}

	return &buf
}
