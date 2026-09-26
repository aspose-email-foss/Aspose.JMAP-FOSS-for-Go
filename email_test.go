package jmap

import (
	"encoding/json"
	"reflect"
	"testing"
)

// EmailStrPtr returns a pointer to the given string.
func EmailStrPtr(s string) *string { return &s }

// EmailBoolPtr returns a pointer to the given bool.
func EmailBoolPtr(b bool) *bool { return &b }

// EmailIntPtr returns a pointer to the given int.
func EmailIntPtr(i int) *int { return &i }

// TestEmailSerializationNormal verifies that a fully populated Email struct
// marshals to the expected JSON and unmarshals back to an equivalent value.
func TestEmailSerializationNormal(t *testing.T) {
	orig := Email{
		Id:           EmailStrPtr("email123"),
		BlobId:       EmailStrPtr("blob456"),
		ThreadId:     EmailStrPtr("thread789"),
		MailboxIds:   map[string]bool{"mailboxA": true, "mailboxB": false},
		Keywords:     map[string]bool{"$seen": true, "$draft": false},
		Size:         EmailIntPtr(1024),
		ReceivedAt:   EmailStrPtr("2023-01-02T15:04:05Z"),
		MessageId:    []string{"<msgid@example.com>"},
		InReplyTo:    []string{"<prev@example.com>"},
		References:   []string{"<ref1@example.com>", "<ref2@example.com>"},
		Sender:       []EmailAddress{{Email: "sender@example.com", Name: EmailStrPtr("Sender")}},
		From:         []EmailAddress{{Email: "from@example.com"}},
		To:           []EmailAddress{{Email: "to@example.com"}},
		Cc:           []EmailAddress{{Email: "cc@example.com"}},
		Bcc:          []EmailAddress{{Email: "bcc@example.com"}},
		ReplyTo:      []EmailAddress{{Email: "reply@example.com"}},
		Subject:      EmailStrPtr("Test Subject"),
		SentAt:       EmailStrPtr("2023-01-01T12:00:00Z"),
		BodyStructure: &EmailBodyPart{
			PartID: EmailStrPtr("0"),
			Size:   512,
			Headers: []EmailHeader{
				{Name: "Content-Type", Value: "text/plain"},
			},
			Type: "text/plain",
			Language: []string{"en"},
			SubParts: nil,
		},
		BodyValues: map[string]EmailBodyValue{
			"0": {
				Value:             "Hello, world!",
				IsEncodingProblem: false,
				IsTruncated:       false,
			},
		},
		TextBody: []EmailBodyPart{
			{
				PartID: EmailStrPtr("0"),
				Size:   512,
				Headers: []EmailHeader{
					{Name: "Content-Type", Value: "text/plain"},
				},
				Type: "text/plain",
				Language: []string{"en"},
			},
		},
		HtmlBody: []EmailBodyPart{
			{
				PartID: EmailStrPtr("1"),
				Size:   1024,
				Headers: []EmailHeader{
					{Name: "Content-Type", Value: "text/html"},
				},
				Type: "text/html",
				Language: []string{"en"},
			},
		},
		Attachments: []EmailBodyPart{
			{
				PartID: EmailStrPtr("2"),
				Size:   2048,
				Headers: []EmailHeader{
					{Name: "Content-Type", Value: "application/pdf"},
				},
				Type: "application/pdf",
				Language: []string{},
			},
		},
		HasAttachment: EmailBoolPtr(true),
		Preview:       EmailStrPtr("Hello, world!"),
	}

	data, err := json.Marshal(orig)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded Email
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if !reflect.DeepEqual(orig, decoded) {
		t.Errorf("decoded Email differs from original.\nOriginal: %#v\nDecoded:  %#v", orig, decoded)
	}
}

// TestEmailSerializationEdge verifies that optional fields omitted (nil pointers,
// empty maps) are correctly omitted from JSON and that unmarshalling restores the
// zero values without errors.
func TestEmailSerializationEdge(t *testing.T) {
	orig := Email{
		MailboxIds: map[string]bool{"mailboxX": true},
		// Keywords omitted (should be omitted from JSON)
		// All pointer fields left nil
		// Slice fields left nil (should encode as null when present, omitted otherwise)
	}

	data, err := json.Marshal(orig)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	// Ensure that omitted optional fields are not present in the JSON output.
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal to raw map failed: %v", err)
	}
	if _, ok := raw["keywords"]; ok {
		t.Errorf("keywords field should be omitted when empty")
	}
	if _, ok := raw["id"]; ok {
		t.Errorf("id field should be omitted when nil")
	}
	if _, ok := raw["size"]; ok {
		t.Errorf("size field should be omitted when nil")
	}
	if _, ok := raw["subject"]; ok {
		t.Errorf("subject field should be omitted when nil")
	}

	// Unmarshal back and verify required field is preserved and optional fields are nil/zero.
	var decoded Email
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Id != nil {
		t.Errorf("expected Id to be nil, got %v", *decoded.Id)
	}
	if decoded.Size != nil {
		t.Errorf("expected Size to be nil, got %v", *decoded.Size)
	}
	if decoded.Subject != nil {
		t.Errorf("expected Subject to be nil, got %v", *decoded.Subject)
	}
	if len(decoded.MailboxIds) != 1 || !decoded.MailboxIds["mailboxX"] {
		t.Errorf("MailboxIds not preserved correctly: %#v", decoded.MailboxIds)
	}
	if decoded.Keywords != nil && len(decoded.Keywords) != 0 {
		t.Errorf("expected Keywords to be nil or empty, got %#v", decoded.Keywords)
	}
}

// TestEmailBodyValueSerialization verifies JSON (de)serialization of EmailBodyValue.
func TestEmailBodyValueSerialization(t *testing.T) {
	orig := EmailBodyValue{
		Value:             "sample body",
		IsEncodingProblem: true,
		IsTruncated:       false,
	}
	data, err := json.Marshal(orig)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var decoded EmailBodyValue
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if !reflect.DeepEqual(orig, decoded) {
		t.Errorf("decoded EmailBodyValue differs from original.\nOriginal: %#v\nDecoded:  %#v", orig, decoded)
	}
}
