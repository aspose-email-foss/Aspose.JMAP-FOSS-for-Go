package jmap

import (
	"encoding/json"
	"errors"
	"testing"
)

// CommonTypesStrPtr returns a pointer to the given string.
func CommonTypesStrPtr(s string) *string { return &s }

// CommonTypesErrPtr returns a pointer to the given error.
func CommonTypesErrPtr(err error) error { return err }

func TestSetErrorJSONSerialization(t *testing.T) {
	// Full struct with all optional fields set.
	se := SetError{
		Type:        "invalidProperties",
		Description: CommonTypesStrPtr("invalid value"),
		Properties:  &[]string{"propA", "propB"},
	}
	data, err := json.Marshal(se)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	expected := `{"type":"invalidProperties","description":"invalid value","properties":["propA","propB"]}`
	if string(data) != expected {
		t.Fatalf("Unexpected JSON. Got %s, want %s", data, expected)
	}

	// Unmarshal back and verify fields.
	var se2 SetError
	if err := json.Unmarshal(data, &se2); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if se2.Type != se.Type {
		t.Errorf("Type mismatch: got %s, want %s", se2.Type, se.Type)
	}
	if se2.Description == nil || *se2.Description != *se.Description {
		t.Errorf("Description mismatch: got %v, want %v", se2.Description, se.Description)
	}
	if se2.Properties == nil || len(*se2.Properties) != 2 ||
		(*se2.Properties)[0] != "propA" || (*se2.Properties)[1] != "propB" {
		t.Errorf("Properties mismatch: got %v, want %v", se2.Properties, se.Properties)
	}
}

func TestSetErrorJSONOmitEmpty(t *testing.T) {
	// Struct with optional fields nil; they should be omitted.
	se := SetError{
		Type: "notFound",
	}
	data, err := json.Marshal(se)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	expected := `{"type":"notFound"}`
	if string(data) != expected {
		t.Fatalf("Unexpected JSON for omitempty fields. Got %s, want %s", data, expected)
	}
}

func TestMethodErrorErrorString(t *testing.T) {
	// Without description.
	me1 := &ProtocolError{Type: "unknownMethod"}
	if got, want := me1.Error(), "unknownMethod"; got != want {
		t.Errorf("Error string without description mismatch: got %q, want %q", got, want)
	}

	// With description.
	desc := "method not supported"
	me2 := &ProtocolError{Type: "unknownMethod", Description: &desc}
	if got, want := me2.Error(), "unknownMethod: method not supported"; got != want {
		t.Errorf("Error string with description mismatch: got %q, want %q", got, want)
	}
}

func TestNetworkErrorErrorAndUnwrap(t *testing.T) {
	underlying := errors.New("connection refused")
	ne := &NetworkError{Err: underlying}

	if got, want := ne.Error(), underlying.Error(); got != want {
		t.Errorf("NetworkError.Error mismatch: got %q, want %q", got, want)
	}

	if unwrapped := errors.Unwrap(ne); unwrapped != underlying {
		t.Errorf("NetworkError.Unwrap returned %v, want %v", unwrapped, underlying)
	}
}

func TestResultReferenceJSON(t *testing.T) {
	rr := ResultReference{
		ResultOf: "getMailboxes",
		Name:     "list",
		Path:     "/ids/*",
	}
	data, err := json.Marshal(rr)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	expected := `{"resultOf":"getMailboxes","name":"list","path":"/ids/*"}`
	if string(data) != expected {
		t.Fatalf("Unexpected JSON for ResultReference. Got %s, want %s", data, expected)
	}

	var rr2 ResultReference
	if err := json.Unmarshal(data, &rr2); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if rr2 != rr {
		t.Errorf("ResultReference mismatch after round‑trip. Got %+v, want %+v", rr2, rr)
	}
}
