package jmap

// Email represents a single email message (RFC 8621 "Email" object).
// It contains immutable content (headers, body) and mutable per‑mailbox metadata.
type Email struct {
	// Id is the server‑assigned identifier for the email.
	// Omit when constructing a create payload.
	Id *string `json:"id,omitempty"`

	// BlobId is the server‑assigned identifier for the raw RFC 5322 message blob.
	// Omit when constructing a create payload.
	BlobId *string `json:"blobId,omitempty"`

	// ThreadId is the server‑assigned identifier for the thread this email belongs to.
	// Omit when constructing a create payload.
	ThreadId *string `json:"threadId,omitempty"`

	// MailboxIds is a required map of mailbox identifiers to a boolean indicating presence.
	MailboxIds map[string]bool `json:"mailboxIds"`

	// Keywords is a map of keyword strings to a boolean indicating presence.
	// Defaults to an empty map.
	Keywords map[string]bool `json:"keywords,omitempty"`

	// Size is the server‑assigned size of the email in octets.
	// Omit when constructing a create payload.
	Size *int `json:"size,omitempty"`

	// ReceivedAt is the server‑assigned UTC timestamp when the email was received.
	// Omit when constructing a create payload.
	ReceivedAt *string `json:"receivedAt,omitempty"`

	// MessageId holds the values of the Message‑Id header, if any.
	MessageId []string `json:"messageId,omitempty"`

	// InReplyTo holds the values of the In‑Reply‑To header, if any.
	InReplyTo []string `json:"inReplyTo,omitempty"`

	// References holds the values of the References header, if any.
	References []string `json:"references,omitempty"`

	// Sender holds the sender addresses.
	Sender []EmailAddress `json:"sender,omitempty"`

	// From holds the from addresses.
	From []EmailAddress `json:"from,omitempty"`

	// To holds the primary recipient addresses.
	To []EmailAddress `json:"to,omitempty"`

	// Cc holds the carbon‑copy recipient addresses.
	Cc []EmailAddress `json:"cc,omitempty"`

	// Bcc holds the blind‑carbon‑copy recipient addresses.
	Bcc []EmailAddress `json:"bcc,omitempty"`

	// ReplyTo holds the reply‑to addresses.
	ReplyTo []EmailAddress `json:"replyTo,omitempty"`

	// Subject is the email subject.
	Subject *string `json:"subject,omitempty"`

	// SentAt is the UTC timestamp when the email was sent.
	SentAt *string `json:"sentAt,omitempty"`

	// BodyStructure is the server‑assigned parsed body structure.
	BodyStructure *EmailBodyPart `json:"bodyStructure,omitempty"`

	// BodyValues maps body part identifiers to their values.
	BodyValues map[string]EmailBodyValue `json:"bodyValues,omitempty"`

	// TextBody contains the plain‑text body parts.
	TextBody []EmailBodyPart `json:"textBody,omitempty"`

	// HtmlBody contains the HTML body parts.
	HtmlBody []EmailBodyPart `json:"htmlBody,omitempty"`

	// Attachments contains the attachment body parts.
	Attachments []EmailBodyPart `json:"attachments,omitempty"`

	// HasAttachment indicates whether the email has any attachments.
	HasAttachment *bool `json:"hasAttachment,omitempty"`

	// Preview is a short preview of the email content.
	Preview *string `json:"preview,omitempty"`
}

// EmailBodyValue represents the value of a body part in an Email.
type EmailBodyValue struct {
	// Value is the raw string value of the body part.
	Value string `json:"value"`

	// IsEncodingProblem indicates whether there was a problem decoding the body part.
	IsEncodingProblem bool `json:"isEncodingProblem"`

	// IsTruncated indicates whether the body part value was truncated.
	IsTruncated bool `json:"isTruncated"`
}
