package jmap

import (
	"encoding/json"
	"reflect"
	"testing"
)

// MailboxStrPtr returns a pointer to the given string.
// Unique helper for this test file.
func MailboxStrPtr(s string) *string { return &s }

// TestMailboxMarshalUnmarshalRoundTrip verifies that a fully populated Mailbox
// marshals to JSON and unmarshals back to an equivalent struct.
func TestMailboxMarshalUnmarshalRoundTrip(t *testing.T) {
	orig := Mailbox{
		Id:            "mailbox123",
		Name:          "Inbox",
		ParentId:      MailboxStrPtr("parent123"),
		Role:          MailboxStrPtr("inbox"),
		SortOrder:     10,
		TotalEmails:   100,
		UnreadEmails:  5,
		TotalThreads:  80,
		UnreadThreads: 2,
		MyRights: &MailboxRights{
			MayReadItems:   true,
			MayAddItems:    false,
			MayRemoveItems: true,
			MaySetSeen:     true,
			MaySetKeywords: false,
			MayCreateChild: true,
			MayRename:      false,
			MayDelete:      true,
			MaySubmit:      false,
		},
		IsSubscribed: true,
	}

	data, err := json.Marshal(&orig)
	if err != nil {
		t.Fatalf("unexpected error marshaling Mailbox: %v", err)
	}

	var decoded Mailbox
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unexpected error unmarshaling Mailbox: %v", err)
	}

	if !reflect.DeepEqual(orig, decoded) {
		t.Errorf("round‑trip Mailbox mismatch.\norig:   %+v\ndecoded:%+v", orig, decoded)
	}
}

// TestMailboxUnmarshalOptionalNull checks that null values for optional fields
// are correctly interpreted as nil pointers.
func TestMailboxUnmarshalOptionalNull(t *testing.T) {
	const payload = `{
		"id": "mailbox123",
		"name": "Archive",
		"parentId": null,
		"role": null,
		"sortOrder": 0,
		"isSubscribed": false
	}`

	var mb Mailbox
	if err := json.Unmarshal([]byte(payload), &mb); err != nil {
		t.Fatalf("unexpected error unmarshaling Mailbox with nulls: %v", err)
	}

	if mb.Id != "mailbox123" {
		t.Errorf("expected Id 'mailbox123', got %q", mb.Id)
	}
	if mb.Name != "Archive" {
		t.Errorf("expected Name 'Archive', got %q", mb.Name)
	}
	if mb.ParentId != nil {
		t.Errorf("expected ParentId nil, got %v", *mb.ParentId)
	}
	if mb.Role != nil {
		t.Errorf("expected Role nil, got %v", *mb.Role)
	}
	if mb.MyRights != nil {
		t.Errorf("expected MyRights nil, got %+v", mb.MyRights)
	}
	if mb.SortOrder != 0 {
		t.Errorf("expected SortOrder 0, got %d", mb.SortOrder)
	}
	if mb.IsSubscribed != false {
		t.Errorf("expected IsSubscribed false, got %v", mb.IsSubscribed)
	}
}
