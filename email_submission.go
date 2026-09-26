package jmap

// EmailSubmission represents one attempt to submit an Email for delivery.
// Creating an EmailSubmission dispatches the message via the server's outbound MTA.
type EmailSubmission struct {
	// Id is the server-assigned identifier of the EmailSubmission.
	// Omitted when constructing a create payload.
	Id *string `json:"id,omitempty"`

	// IdentityId must reference an existing Identity. Required.
	IdentityId string `json:"identityId"`

	// EmailId is the identifier of the Email to send. Required.
	EmailId string `json:"emailId"`

	// ThreadId is the server-assigned thread identifier.
	// Omitted when constructing a create payload.
	ThreadId *string `json:"threadId,omitempty"`

	// Envelope contains the SMTP envelope information.
	// Null (or omitted) lets the server derive it from the Email's headers.
	Envelope *Envelope `json:"envelope,omitempty"`

	// SendAt is the server-assigned UTC date/time when the message will be sent.
	// Omitted when constructing a create payload.
	SendAt *string `json:"sendAt,omitempty"`

	// UndoStatus is the server-assigned status of the submission (pending, final, canceled).
	// Omitted when constructing a create payload.
	UndoStatus *string `json:"undoStatus,omitempty"`

	// DeliveryStatus maps recipient identifiers to their delivery status.
	// Null (or omitted) when not yet available.
	DeliveryStatus *map[string]DeliveryStatus `json:"deliveryStatus,omitempty"`

	// DsnBlobIds are server-assigned blob identifiers for DSN (Delivery Status Notification) messages.
	// Omitted when constructing a create payload.
	DsnBlobIds []string `json:"dsnBlobIds,omitempty"`

	// MdnBlobIds are server-assigned blob identifiers for MDN (Message Disposition Notification) messages.
	// Omitted when constructing a create payload.
	MdnBlobIds []string `json:"mdnBlobIds,omitempty"`
}
