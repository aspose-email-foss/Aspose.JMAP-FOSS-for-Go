package jmap

import (
	"encoding/json"
	"reflect"
	"testing"
)

// ThreadStrPtr returns a pointer to the given string.
// Unique helper for this test file.
func ThreadStrPtr(s string) *string { return &s }

// ThreadSlicePtr returns a pointer to the given []string.
// Unique helper for this test file.
func ThreadSlicePtr(ss []string) *[]string { return &ss }

func TestThreadJSONMarshal(t *testing.T) {
	thread := Thread{
		Id:       ThreadStrPtr("thread123"),
		EmailIds: ThreadSlicePtr([]string{"email1", "email2"}),
	}
	data, err := json.Marshal(thread)
	if err != nil {
		t.Fatalf("unexpected error marshaling Thread: %v", err)
	}
	expected := `{"id":"thread123","emailIds":["email1","email2"]}`
	if string(data) != expected {
		t.Fatalf("marshal output mismatch.\nExpected: %s\nGot:      %s", expected, string(data))
	}
}

func TestThreadJSONUnmarshal(t *testing.T) {
	payload := []byte(`{"id":"thread123","emailIds":["email1","email2"]}`)
	var thread Thread
	if err := json.Unmarshal(payload, &thread); err != nil {
		t.Fatalf("unexpected error unmarshaling Thread: %v", err)
	}
	if thread.Id == nil || *thread.Id != "thread123" {
		t.Fatalf("expected Id 'thread123', got %+v", thread.Id)
	}
	if thread.EmailIds == nil || !reflect.DeepEqual(*thread.EmailIds, []string{"email1", "email2"}) {
		t.Fatalf("expected EmailIds ['email1','email2'], got %+v", thread.EmailIds)
	}
}

func TestThreadJSONMarshalEmpty(t *testing.T) {
	// All optional fields omitted.
	thread := Thread{}
	data, err := json.Marshal(thread)
	if err != nil {
		t.Fatalf("unexpected error marshaling empty Thread: %v", err)
	}
	expected := `{}` // because of omitempty on all fields
	if string(data) != expected {
		t.Fatalf("marshal empty Thread mismatch.\nExpected: %s\nGot:      %s", expected, string(data))
	}
}

func TestThreadJSONUnmarshalEmpty(t *testing.T) {
	payload := []byte(`{}`)
	var thread Thread
	if err := json.Unmarshal(payload, &thread); err != nil {
		t.Fatalf("unexpected error unmarshaling empty Thread: %v", err)
	}
	if thread.Id != nil {
		t.Fatalf("expected Id to be nil, got %+v", thread.Id)
	}
	if thread.EmailIds != nil {
		t.Fatalf("expected EmailIds to be nil, got %+v", thread.EmailIds)
	}
}
