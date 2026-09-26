package jmap

import (
	"encoding/json"
	"reflect"
	"testing"
)

// EnvelopeStrPtr returns a pointer to the given string.
// Unique to this test file to avoid name collisions.
func EnvelopeStrPtr(s string) *string { return &s }

func TestEnvelopeJSONRoundTrip(t *testing.T) {
	orig := Envelope{
		MailFrom: Address{
			Email: "sender@example.com",
			Parameters: map[string]*string{
				"RET": EnvelopeStrPtr("HDRS"),
			},
		},
		RcptTo: []Address{
			{
				Email:      "rcpt1@example.com",
				Parameters: nil,
			},
			{
				Email: "rcpt2@example.com",
				Parameters: map[string]*string{
					"SIZE": EnvelopeStrPtr("12345"),
				},
			},
		},
	}

	data, err := json.Marshal(orig)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded Envelope
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if !reflect.DeepEqual(orig, decoded) {
		t.Errorf("round‑trip mismatch.\nOriginal: %#v\nDecoded:  %#v", orig, decoded)
	}
}

func TestEnvelopeJSONDeserializationEdge(t *testing.T) {
	const payload = `{
		"mailFrom": {
			"email": "sender@example.com",
			"parameters": {"RET": null}
		},
		"rcptTo": [
			{
				"email": "rcpt@example.com"
			}
		]
	}`

	var env Envelope
	if err := json.Unmarshal([]byte(payload), &env); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	// Verify required fields are present.
	if env.MailFrom.Email != "sender@example.com" {
		t.Errorf("unexpected MailFrom.Email: got %q, want %q", env.MailFrom.Email, "sender@example.com")
	}
	if len(env.RcptTo) != 1 || env.RcptTo[0].Email != "rcpt@example.com" {
		t.Errorf("unexpected RcptTo: %#v", env.RcptTo)
	}

	// Parameters map should exist with a key whose value is nil.
	if env.MailFrom.Parameters == nil {
		t.Fatalf("MailFrom.Parameters map is nil, expected map with key")
	}
	val, ok := env.MailFrom.Parameters["RET"]
	if !ok {
		t.Fatalf("MailFrom.Parameters missing key \"RET\"")
	}
	if val != nil {
		t.Errorf("MailFrom.Parameters[\"RET\"] expected nil, got %v", *val)
	}
}
