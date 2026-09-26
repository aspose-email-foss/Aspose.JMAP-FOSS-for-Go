package jmap

import (
	"encoding/json"
	"reflect"
	"testing"
)

// EmailSubmissionStrPtr returns a pointer to the given string.
// Unique helper for this test file.
func EmailSubmissionStrPtr(s string) *string { return &s }

// EmailSubmissionMapPtr returns a pointer to the given map.
// Unique helper for this test file.
func EmailSubmissionMapPtr(m map[string]DeliveryStatus) *map[string]DeliveryStatus { return &m }

func TestEmailSubmissionMarshalFull(t *testing.T) {
	sub := EmailSubmission{
		Id:          EmailSubmissionStrPtr("sub123"),
		IdentityId:  "ident456",
		EmailId:     "email789",
		ThreadId:    EmailSubmissionStrPtr("thread321"),
		Envelope: &Envelope{
			MailFrom: Address{
				Email: "sender@example.com",
			},
			RcptTo: []Address{
				{Email: "rcpt1@example.com"},
				{Email: "rcpt2@example.com"},
			},
		},
		SendAt:       EmailSubmissionStrPtr("2023-10-01T12:34:56Z"),
		UndoStatus:   EmailSubmissionStrPtr("pending"),
		DsnBlobIds:   []string{"dsn1", "dsn2"},
		MdnBlobIds:   []string{"mdn1"},
		DeliveryStatus: EmailSubmissionMapPtr(map[string]DeliveryStatus{
			"rcpt1@example.com": {
				SmtpReply: "250 OK",
				Delivered: "yes",
				Displayed: "yes",
			},
			"rcpt2@example.com": {
				SmtpReply: "250 OK",
				Delivered: "yes",
				Displayed: "unknown",
			},
		}),
	}

	data, err := json.Marshal(sub)
	if err != nil {
		t.Fatalf("unexpected marshal error: %v", err)
	}

	var got map[string]interface{}
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unexpected unmarshal of marshaled data: %v", err)
	}

	// Build expected map structure.
	expected := map[string]interface{}{
		"id":         "sub123",
		"identityId": "ident456",
		"emailId":    "email789",
		"threadId":   "thread321",
		"envelope": map[string]interface{}{
			"mailFrom": map[string]interface{}{
				"email": "sender@example.com",
			},
			"rcptTo": []interface{}{
				map[string]interface{}{"email": "rcpt1@example.com"},
				map[string]interface{}{"email": "rcpt2@example.com"},
			},
		},
		"sendAt":   "2023-10-01T12:34:56Z",
		"undoStatus": "pending",
		"dsnBlobIds": []interface{}{"dsn1", "dsn2"},
		"mdnBlobIds": []interface{}{"mdn1"},
		"deliveryStatus": map[string]interface{}{
			"rcpt1@example.com": map[string]interface{}{
				"smtpReply": "250 OK",
				"delivered": "yes",
				"displayed": "yes",
			},
			"rcpt2@example.com": map[string]interface{}{
				"smtpReply": "250 OK",
				"delivered": "yes",
				"displayed": "unknown",
			},
		},
	}

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("marshal output mismatch.\nGot:  %#v\nWant: %#v", got, expected)
	}
}

func TestEmailSubmissionUnmarshalPartial(t *testing.T) {
	// JSON with only required fields and explicit null for deliveryStatus.
	const payload = `{
		"identityId": "identABC",
		"emailId": "emailXYZ",
		"deliveryStatus": null
	}`

	var sub EmailSubmission
	if err := json.Unmarshal([]byte(payload), &sub); err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}

	if sub.Id != nil {
		t.Errorf("expected Id nil, got %v", *sub.Id)
	}
	if sub.ThreadId != nil {
		t.Errorf("expected ThreadId nil, got %v", *sub.ThreadId)
	}
	if sub.Envelope != nil {
		t.Errorf("expected Envelope nil, got %+v", sub.Envelope)
	}
	if sub.SendAt != nil {
		t.Errorf("expected SendAt nil, got %v", *sub.SendAt)
	}
	if sub.UndoStatus != nil {
		t.Errorf("expected UndoStatus nil, got %v", *sub.UndoStatus)
	}
	if sub.DeliveryStatus != nil {
		t.Errorf("expected DeliveryStatus nil, got %+v", *sub.DeliveryStatus)
	}
	if sub.DsnBlobIds != nil {
		t.Errorf("expected DsnBlobIds nil, got %+v", sub.DsnBlobIds)
	}
	if sub.MdnBlobIds != nil {
		t.Errorf("expected MdnBlobIds nil, got %+v", sub.MdnBlobIds)
	}
	if sub.IdentityId != "identABC" {
		t.Errorf("unexpected IdentityId: %s", sub.IdentityId)
	}
	if sub.EmailId != "emailXYZ" {
		t.Errorf("unexpected EmailId: %s", sub.EmailId)
	}
}
