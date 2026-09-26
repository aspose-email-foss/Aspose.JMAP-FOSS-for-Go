package jmap

import (
	"encoding/json"
	"reflect"
	"testing"
)

// IdentityPtrString returns a pointer to the given string.
// Used for constructing expected Identity values with pointer fields.
func IdentityPtrString(s string) *string { return &s }

// IdentityPtrBool returns a pointer to the given bool.
func IdentityPtrBool(b bool) *bool { return &b }

// IdentityPtrEmailSlice returns a pointer to the given slice of EmailAddress.
func IdentityPtrEmailSlice(v []EmailAddress) *[]EmailAddress { return &v }

func TestIdentityMarshalRoundTrip(t *testing.T) {
	orig := Identity{
		Id:            "id123",
		Name:          "John Doe",
		Email:         "john@example.com",
		ReplyTo:       IdentityPtrEmailSlice([]EmailAddress{{Email: "reply@example.com"}}),
		Bcc:           IdentityPtrEmailSlice([]EmailAddress{{Email: "bcc@example.com"}}),
		TextSignature: "Best regards",
		HtmlSignature: "<p>Best regards</p>",
		MayDelete:     IdentityPtrBool(true),
	}

	data, err := json.Marshal(orig)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded Identity
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if !reflect.DeepEqual(orig, decoded) {
		t.Errorf("Round‑trip mismatch.\nOriginal: %#v\nDecoded:  %#v", orig, decoded)
	}
}

func TestIdentityUnmarshalWithNulls(t *testing.T) {
	const payload = `{
		"id": "id456",
		"name": "Alice",
		"email": "alice@example.com",
		"replyTo": null,
		"bcc": null,
		"textSignature": "",
		"htmlSignature": "",
		"mayDelete": null
	}`

	var id Identity
	if err := json.Unmarshal([]byte(payload), &id); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if id.Id != "id456" {
		t.Errorf("Id mismatch: got %q, want %q", id.Id, "id456")
	}
	if id.Name != "Alice" {
		t.Errorf("Name mismatch: got %q, want %q", id.Name, "Alice")
	}
	if id.Email != "alice@example.com" {
		t.Errorf("Email mismatch: got %q, want %q", id.Email, "alice@example.com")
	}
	if id.ReplyTo != nil {
		t.Errorf("ReplyTo expected nil, got %v", id.ReplyTo)
	}
	if id.Bcc != nil {
		t.Errorf("Bcc expected nil, got %v", id.Bcc)
	}
	if id.TextSignature != "" {
		t.Errorf("TextSignature expected empty, got %q", id.TextSignature)
	}
	if id.HtmlSignature != "" {
		t.Errorf("HtmlSignature expected empty, got %q", id.HtmlSignature)
	}
	if id.MayDelete != nil {
		t.Errorf("MayDelete expected nil, got %v", id.MayDelete)
	}
}
