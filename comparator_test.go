package jmap

import (
	"encoding/json"
	"reflect"
	"testing"
)

// ComparatorBoolPtr returns a pointer to the given bool.
// Unique to this test file to avoid name collisions.
func ComparatorBoolPtr(v bool) *bool { return &v }

// ComparatorStringPtr returns a pointer to the given string.
// Unique to this test file to avoid name collisions.
func ComparatorStringPtr(v string) *string { return &v }

func TestComparatorSerializationFull(t *testing.T) {
	orig := Comparator{
		Property:    "subject",
		IsAscending: ComparatorBoolPtr(false),
		Collation:   ComparatorStringPtr("i;unicode"),
	}

	data, err := json.Marshal(orig)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Expected JSON must contain all three fields.
	var gotMap map[string]json.RawMessage
	if err := json.Unmarshal(data, &gotMap); err != nil {
		t.Fatalf("Unmarshal to map failed: %v", err)
	}
	if _, ok := gotMap["property"]; !ok {
		t.Errorf("missing 'property' field in JSON: %s", string(data))
	}
	if _, ok := gotMap["isAscending"]; !ok {
		t.Errorf("missing 'isAscending' field in JSON: %s", string(data))
	}
	if _, ok := gotMap["collation"]; !ok {
		t.Errorf("missing 'collation' field in JSON: %s", string(data))
	}

	// Round‑trip back to struct.
	var decoded Comparator
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal back to Comparator failed: %v", err)
	}
	if decoded.Property != orig.Property {
		t.Errorf("Property mismatch: got %q, want %q", decoded.Property, orig.Property)
	}
	if decoded.IsAscending == nil || *decoded.IsAscending != *orig.IsAscending {
		t.Errorf("IsAscending mismatch: got %v, want %v", decoded.IsAscending, orig.IsAscending)
	}
	if decoded.Collation == nil || *decoded.Collation != *orig.Collation {
		t.Errorf("Collation mismatch: got %v, want %v", decoded.Collation, orig.Collation)
	}
}

func TestComparatorSerializationPartial(t *testing.T) {
	// Only the required field is set; optional fields are nil.
	orig := Comparator{
		Property: "receivedAt",
		// IsAscending and Collation left nil.
	}

	data, err := json.Marshal(orig)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// The JSON must contain only the required field.
	var gotMap map[string]json.RawMessage
	if err := json.Unmarshal(data, &gotMap); err != nil {
		t.Fatalf("Unmarshal to map failed: %v", err)
	}
	if len(gotMap) != 1 {
		t.Errorf("Expected exactly 1 field in JSON, got %d: %v", len(gotMap), gotMap)
	}
	if _, ok := gotMap["property"]; !ok {
		t.Errorf("Missing required 'property' field in JSON: %s", string(data))
	}
	if _, ok := gotMap["isAscending"]; ok {
		t.Errorf("'isAscending' should be omitted when nil")
	}
	if _, ok := gotMap["collation"]; ok {
		t.Errorf("'collation' should be omitted when nil")
	}

	// Unmarshal back and verify fields.
	var decoded Comparator
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal back to Comparator failed: %v", err)
	}
	if decoded.Property != orig.Property {
		t.Errorf("Property mismatch: got %q, want %q", decoded.Property, orig.Property)
	}
	if decoded.IsAscending != nil {
		t.Errorf("IsAscending expected nil, got %v", *decoded.IsAscending)
	}
	if decoded.Collation != nil {
		t.Errorf("Collation expected nil, got %v", *decoded.Collation)
	}
}

// Ensure that a Comparator with a nil IsAscending defaults to true
// when interpreted by client code (not by JSON). This test only checks
// that the pointer is nil after unmarshalling.
func TestComparatorDeserializationMissingIsAscending(t *testing.T) {
	jsonStr := `{"property":"size"}`
	var comp Comparator
	if err := json.Unmarshal([]byte(jsonStr), &comp); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if comp.Property != "size" {
		t.Errorf("Property mismatch: got %q, want %q", comp.Property, "size")
	}
	if comp.IsAscending != nil {
		t.Errorf("IsAscending should be nil when omitted, got %v", *comp.IsAscending)
	}
	if comp.Collation != nil {
		t.Errorf("Collation should be nil when omitted, got %v", *comp.Collation)
	}
}

// Verify that two Comparator structs with identical content are equal via reflect.DeepEqual.
func TestComparatorDeepEqual(t *testing.T) {
	a := Comparator{
		Property:    "from",
		IsAscending: ComparatorBoolPtr(true),
		Collation:   ComparatorStringPtr("i;unicode"),
	}
	b := Comparator{
		Property:    "from",
		IsAscending: ComparatorBoolPtr(true),
		Collation:   ComparatorStringPtr("i;unicode"),
	}
	if !reflect.DeepEqual(a, b) {
		t.Errorf("Expected structs to be deeply equal: %+v vs %+v", a, b)
	}
}
