package jmap

import (
	"encoding/json"
	"reflect"
	"testing"
)

// EmailBodyPartPtrString returns a pointer to the given string.
// Unique to this test file to avoid name collisions.
func EmailBodyPartPtrString(s string) *string {
	return &s
}

// TestEmailBodyPartJSONRoundTrip verifies that a fully populated EmailBodyPart
// marshals to the expected JSON structure and unmarshals back to an equivalent value.
func TestEmailBodyPartJSONRoundTrip(t *testing.T) {
	// Build a complete EmailBodyPart instance.
	part := EmailBodyPart{
		PartID:      EmailBodyPartPtrString("0"),
		BlobID:      EmailBodyPartPtrString("blob123"),
		Size:        1024,
		Headers:     []EmailHeader{{Name: "Subject", Value: "Test"}},
		Name:        EmailBodyPartPtrString("example.txt"),
		Type:        "text/plain",
		Charset:     EmailBodyPartPtrString("utf-8"),
		Disposition: EmailBodyPartPtrString("attachment"),
		CID:         EmailBodyPartPtrString("<cid@example>"),
		Language:    []string{"en", "fr"},
		Location:    EmailBodyPartPtrString("us"),
		SubParts: []EmailBodyPart{
			{
				PartID: EmailBodyPartPtrString("0.1"),
				Size:   512,
				Headers: []EmailHeader{
					{Name: "Content-Type", Value: "image/png"},
				},
				Type: "image/png",
			},
		},
	}

	// Marshal to JSON.
	data, err := json.Marshal(part)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Unmarshal into a generic map for structural inspection.
	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal to map failed: %v", err)
	}

	// Verify required fields are present and have expected values.
	if got["size"] != float64(part.Size) {
		t.Errorf("size mismatch: got %v want %d", got["size"], part.Size)
	}
	if got["type"] != part.Type {
		t.Errorf("type mismatch: got %v want %s", got["type"], part.Type)
	}
	if _, ok := got["headers"]; !ok {
		t.Errorf("headers missing")
	}
	if _, ok := got["language"]; !ok {
		t.Errorf("language missing")
	}
	if _, ok := got["subParts"]; !ok {
		t.Errorf("subParts missing")
	}

	// Unmarshal back into a struct and compare with the original.
	var roundTrip EmailBodyPart
	if err := json.Unmarshal(data, &roundTrip); err != nil {
		t.Fatalf("Unmarshal back to struct failed: %v", err)
	}
	if !reflect.DeepEqual(part, roundTrip) {
		t.Errorf("Round‑trip mismatch:\noriginal: %+v\nroundTrip: %+v", part, roundTrip)
	}
}

// TestEmailBodyPartJSONOptionalFields checks handling of nil optional fields.
// Optional fields should be omitted, while slice fields without omitempty should appear as null.
func TestEmailBodyPartJSONOptionalFields(t *testing.T) {
	part := EmailBodyPart{
		Size:    2048,
		Headers: []EmailHeader{{Name: "From", Value: "alice@example.com"}},
		Type:    "text/html",
		// Optional fields left nil; Language and SubParts are nil slices.
	}

	data, err := json.Marshal(part)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal to map failed: %v", err)
	}

	// Optional pointer fields must be omitted.
	optionalFields := []string{"partId", "blobId", "name", "charset", "disposition", "cid", "location"}
	for _, f := range optionalFields {
		if _, present := got[f]; present {
			t.Errorf("optional field %s should be omitted, got %v", f, got[f])
		}
	}

	// Language and subParts should be present with null values.
	if v, ok := got["language"]; !ok {
		t.Errorf("language field missing")
	} else if v != nil {
		t.Errorf("language field should be null, got %v", v)
	}
	if v, ok := got["subParts"]; !ok {
		t.Errorf("subParts field missing")
	} else if v != nil {
		t.Errorf("subParts field should be null, got %v", v)
	}
}
