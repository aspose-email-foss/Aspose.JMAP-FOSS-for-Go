package jmap

import (
	"encoding/json"
	"testing"
)

// EmailAddressStrPtr returns a pointer to the given string.
// Unique to this test file to avoid name collisions.
func EmailAddressStrPtr(s string) *string {
	return &s
}

// TestEmailAddressMarshalWithName verifies that marshaling an EmailAddress
// with a non-nil Name includes both fields in the JSON output.
func TestEmailAddressMarshalWithName(t *testing.T) {
	addr := EmailAddress{
		Name:  EmailAddressStrPtr("Alice"),
		Email: "alice@example.com",
	}
	data, err := json.Marshal(addr)
	if err != nil {
		t.Fatalf("unexpected error marshaling EmailAddress: %v", err)
	}
	expected := `{"name":"Alice","email":"alice@example.com"}`
	if string(data) != expected {
		t.Errorf("unexpected JSON output.\nGot:  %s\nWant: %s", data, expected)
	}
}

// TestEmailAddressMarshalWithoutName verifies that marshaling an EmailAddress
// with a nil Name omits the name property (omitempty) while still including email.
func TestEmailAddressMarshalWithoutName(t *testing.T) {
	addr := EmailAddress{
		Name:  nil,
		Email: "bob@example.com",
	}
	data, err := json.Marshal(addr)
	if err != nil {
		t.Fatalf("unexpected error marshaling EmailAddress: %v", err)
	}
	expected := `{"email":"bob@example.com"}`
	if string(data) != expected {
		t.Errorf("unexpected JSON output for nil Name.\nGot:  %s\nWant: %s", data, expected)
	}
}

// TestEmailAddressUnmarshalRoundTrip checks that unmarshaling JSON with both
// fields and with a null name correctly populates the struct.
func TestEmailAddressUnmarshalRoundTrip(t *testing.T) {
	// Case with name present
	jsonWithName := `{"name":"Carol","email":"carol@example.com"}`
	var addr1 EmailAddress
	if err := json.Unmarshal([]byte(jsonWithName), &addr1); err != nil {
		t.Fatalf("unexpected error unmarshaling with name: %v", err)
	}
	if addr1.Name == nil || *addr1.Name != "Carol" {
		t.Errorf("expected Name to be 'Carol', got %+v", addr1.Name)
	}
	if addr1.Email != "carol@example.com" {
		t.Errorf("expected Email to be 'carol@example.com', got %s", addr1.Email)
	}

	// Case with name explicitly null
	jsonWithNullName := `{"name":null,"email":"dave@example.com"}`
	var addr2 EmailAddress
	if err := json.Unmarshal([]byte(jsonWithNullName), &addr2); err != nil {
		t.Fatalf("unexpected error unmarshaling with null name: %v", err)
	}
	if addr2.Name != nil {
		t.Errorf("expected Name to be nil for null JSON value, got %+v", addr2.Name)
	}
	if addr2.Email != "dave@example.com" {
		t.Errorf("expected Email to be 'dave@example.com', got %s", addr2.Email)
	}
}
