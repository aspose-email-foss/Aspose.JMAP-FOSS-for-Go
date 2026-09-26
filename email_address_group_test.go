package jmap

import (
	"encoding/json"
	"reflect"
	"testing"
)

// EmailAddressGroupStrPtr returns a pointer to the given string.
// Helper name is prefixed to avoid collisions with other test files.
func EmailAddressGroupStrPtr(s string) *string {
	return &s
}

// TestEmailAddressGroupMarshal verifies that a populated EmailAddressGroup
// marshals to the expected JSON representation.
func TestEmailAddressGroupMarshal(t *testing.T) {
	group := EmailAddressGroup{
		Name: EmailAddressGroupStrPtr("Team"),
		Addresses: []EmailAddress{
			{Email: "a@x"},
			{Email: "b@x"},
		},
	}
	data, err := json.Marshal(group)
	if err != nil {
		t.Fatalf("unexpected error during marshal: %v", err)
	}
	expected := `{"name":"Team","addresses":[{"email":"a@x"},{"email":"b@x"}]}`
	if string(data) != expected {
		t.Errorf("marshal output mismatch.\nExpected: %s\nGot:      %s", expected, string(data))
	}
}

// TestEmailAddressGroupUnmarshal verifies that JSON with all required fields
// unmarshals into an EmailAddressGroup with the correct values.
func TestEmailAddressGroupUnmarshal(t *testing.T) {
	jsonInput := `{"name":"Team","addresses":[{"email":"a@x"},{"email":"b@x"}]}`
	var group EmailAddressGroup
	if err := json.Unmarshal([]byte(jsonInput), &group); err != nil {
		t.Fatalf("unexpected error during unmarshal: %v", err)
	}
	if group.Name == nil || *group.Name != "Team" {
		t.Errorf("expected Name to be 'Team', got %+v", group.Name)
	}
	if len(group.Addresses) != 2 {
		t.Fatalf("expected 2 addresses, got %d", len(group.Addresses))
	}
	if group.Addresses[0].Email != "a@x" || group.Addresses[1].Email != "b@x" {
		t.Errorf("addresses mismatch: %+v", group.Addresses)
	}
}

// TestEmailAddressGroupUnmarshalNullName verifies that a JSON payload with a
// null or omitted name results in a nil Name pointer.
func TestEmailAddressGroupUnmarshalNullName(t *testing.T) {
	// Case 1: name omitted
	jsonOmitted := `{"addresses":[{"email":"a@x"}]}`
	var grpOmitted EmailAddressGroup
	if err := json.Unmarshal([]byte(jsonOmitted), &grpOmitted); err != nil {
		t.Fatalf("unexpected error during unmarshal (omitted name): %v", err)
	}
	if grpOmitted.Name != nil {
		t.Errorf("expected Name to be nil when omitted, got %+v", grpOmitted.Name)
	}
	if len(grpOmitted.Addresses) != 1 || grpOmitted.Addresses[0].Email != "a@x" {
		t.Errorf("unexpected addresses when name omitted: %+v", grpOmitted.Addresses)
	}

	// Case 2: name explicitly null
	jsonNull := `{"name":null,"addresses":[{"email":"b@x"}]}`
	var grpNull EmailAddressGroup
	if err := json.Unmarshal([]byte(jsonNull), &grpNull); err != nil {
		t.Fatalf("unexpected error during unmarshal (null name): %v", err)
	}
	if grpNull.Name != nil {
		t.Errorf("expected Name to be nil when null, got %+v", grpNull.Name)
	}
	if len(grpNull.Addresses) != 1 || grpNull.Addresses[0].Email != "b@x" {
		t.Errorf("unexpected addresses when name null: %+v", grpNull.Addresses)
	}
}

// TestEmailAddressGroupRoundTrip ensures that marshaling then unmarshaling
// yields an equivalent struct (including nil Name handling).
func TestEmailAddressGroupRoundTrip(t *testing.T) {
	original := EmailAddressGroup{
		Name: nil,
		Addresses: []EmailAddress{
			{Email: "c@x"},
		},
	}
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	var decoded EmailAddressGroup
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if !reflect.DeepEqual(original, decoded) {
		t.Errorf("round-trip mismatch.\nOriginal: %+v\nDecoded:  %+v", original, decoded)
	}
}
