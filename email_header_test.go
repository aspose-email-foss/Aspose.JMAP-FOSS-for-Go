package jmap

import (
	"encoding/json"
	"testing"
)

// TestEmailHeaderJSONMarshal verifies that a normal EmailHeader marshals to the expected JSON.
func TestEmailHeaderJSONMarshal(t *testing.T) {
	h := EmailHeader{
		Name:  "Subject",
		Value: "Hello World",
	}
	data, err := json.Marshal(h)
	if err != nil {
		t.Fatalf("unexpected error during marshal: %v", err)
	}
	const expected = `{"name":"Subject","value":"Hello World"}`
	if string(data) != expected {
		t.Errorf("marshal output mismatch.\nExpected: %s\nGot:      %s", expected, string(data))
	}
}

// TestEmailHeaderJSONUnmarshal verifies that a normal JSON payload unmarshals into an EmailHeader.
func TestEmailHeaderJSONUnmarshal(t *testing.T) {
	const payload = `{"name":"From","value":"alice@example.com"}`
	var h EmailHeader
	if err := json.Unmarshal([]byte(payload), &h); err != nil {
		t.Fatalf("unexpected error during unmarshal: %v", err)
	}
	if h.Name != "From" {
		t.Errorf("unexpected Name field: want %q, got %q", "From", h.Name)
	}
	if h.Value != "alice@example.com" {
		t.Errorf("unexpected Value field: want %q, got %q", "alice@example.com", h.Value)
	}
}

// TestEmailHeaderJSONUnmarshalMissingValue checks that missing optional fields (none in this case)
// result in zero values without error.
func TestEmailHeaderJSONUnmarshalMissingValue(t *testing.T) {
	const payload = `{"name":"X-Custom-Header"}`
	var h EmailHeader
	if err := json.Unmarshal([]byte(payload), &h); err != nil {
		t.Fatalf("unexpected error during unmarshal with missing value: %v", err)
	}
	if h.Name != "X-Custom-Header" {
		t.Errorf("unexpected Name field: want %q, got %q", "X-Custom-Header", h.Name)
	}
	if h.Value != "" {
		t.Errorf("expected empty Value for missing field, got %q", h.Value)
	}
}

// TestEmailHeaderJSONUnmarshalNullName documents standard encoding/json behavior: unmarshalling
// JSON null into a plain (non-pointer) string field is a documented no-op (the field is left at
// its zero value), not an error - per model_conventions, EmailHeader relies entirely on
// encoding/json's struct-tag-driven decoding with no hand-written UnmarshalJSON, so it cannot
// (and should not) special-case null itself.
func TestEmailHeaderJSONUnmarshalNullName(t *testing.T) {
	const payload = `{"name":null,"value":"something"}`
	var h EmailHeader
	if err := json.Unmarshal([]byte(payload), &h); err != nil {
		t.Fatalf("unexpected error unmarshalling null into a plain string field: %v", err)
	}
	if h.Name != "" {
		t.Errorf("expected Name to remain the zero value after a null field, got %q", h.Name)
	}
}
