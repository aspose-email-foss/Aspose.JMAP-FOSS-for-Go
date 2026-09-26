package jmap

import (
	"encoding/json"
	"reflect"
	"testing"
)

// SearchSnippetPtrString returns a pointer to the given string.
// Helper is uniquely prefixed for this test file.
func SearchSnippetPtrString(s string) *string {
	return &s
}

// TestSearchSnippetJSONRoundTrip verifies that a fully populated SearchSnippet
// marshals to JSON and unmarshals back to an equivalent value.
func TestSearchSnippetJSONRoundTrip(t *testing.T) {
	original := SearchSnippet{
		EmailId: "email-123",
		Subject: SearchSnippetPtrString("<mark>Test</mark> Subject"),
		Preview: SearchSnippetPtrString("This is a <mark>preview</mark>."),
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("unexpected error marshaling SearchSnippet: %v", err)
	}

	var decoded SearchSnippet
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unexpected error unmarshaling SearchSnippet: %v", err)
	}

	if !reflect.DeepEqual(original, decoded) {
		t.Errorf("round‑trip mismatch.\noriginal: %+v\ndecoded:  %+v", original, decoded)
	}
}

// TestSearchSnippetJSONOmitEmpty verifies that optional fields are omitted
// from the JSON output when they are nil.
func TestSearchSnippetJSONOmitEmpty(t *testing.T) {
	// Only the required field is set; optional fields are nil.
	snippet := SearchSnippet{
		EmailId: "email-456",
		Subject: nil,
		Preview: nil,
	}

	data, err := json.Marshal(snippet)
	if err != nil {
		t.Fatalf("unexpected error marshaling SearchSnippet with nil fields: %v", err)
	}

	// The JSON should contain only the emailId property.
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unexpected error unmarshaling intermediate JSON: %v", err)
	}

	if len(raw) != 1 {
		t.Fatalf("expected exactly 1 JSON property, got %d: %v", len(raw), raw)
	}
	if _, ok := raw["emailId"]; !ok {
		t.Fatalf("expected JSON property 'emailId' to be present")
	}
	if _, ok := raw["subject"]; ok {
		t.Errorf("unexpected 'subject' property in JSON when Subject is nil")
	}
	if _, ok := raw["preview"]; ok {
		t.Errorf("unexpected 'preview' property in JSON when Preview is nil")
	}
}
