package jmap

import (
	"encoding/json"
	"reflect"
	"testing"
)

// TestDeliveryStatusJSONRoundTrip verifies that a DeliveryStatus struct marshals to the
// expected JSON representation and unmarshals back to an equivalent struct.
func TestDeliveryStatusJSONRoundTrip(t *testing.T) {
	orig := DeliveryStatus{
		SmtpReply: "250 2.0.0 OK",
		Delivered: "yes",
		Displayed: "yes",
	}
	data, err := json.Marshal(orig)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}
	const expected = `{"smtpReply":"250 2.0.0 OK","delivered":"yes","displayed":"yes"}`
	if string(data) != expected {
		t.Fatalf("unexpected JSON output.\nGot:  %s\nWant: %s", string(data), expected)
	}

	var decoded DeliveryStatus
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}
	if !reflect.DeepEqual(orig, decoded) {
		t.Fatalf("round‑trip mismatch.\nOrig: %+v\nDecoded: %+v", orig, decoded)
	}
}

// TestDeliveryStatusJSONUnmarshalPartial ensures that missing optional fields (none are
// optional in this model) result in zero values without error, and that unknown fields are ignored.
func TestDeliveryStatusJSONUnmarshalPartial(t *testing.T) {
	const payload = `{
		"smtpReply": "550 5.1.1 User unknown",
		"delivered": "no",
		"extraField": "should be ignored"
	}`
	var ds DeliveryStatus
	if err := json.Unmarshal([]byte(payload), &ds); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}
	if ds.SmtpReply != "550 5.1.1 User unknown" {
		t.Errorf("SmtpReply mismatch: got %q, want %q", ds.SmtpReply, "550 5.1.1 User unknown")
	}
	if ds.Delivered != "no" {
		t.Errorf("Delivered mismatch: got %q, want %q", ds.Delivered, "no")
	}
	// Displayed is required but absent; it should be the zero value (empty string).
	if ds.Displayed != "" {
		t.Errorf("Displayed expected zero value, got %q", ds.Displayed)
	}
}
