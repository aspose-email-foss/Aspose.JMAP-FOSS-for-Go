package jmap

// DeliveryStatus represents per-recipient delivery outcome, keyed by recipient email address on EmailSubmission.deliveryStatus.
type DeliveryStatus struct {
	// SmtpReply is the SMTP reply string.
	SmtpReply string `json:"smtpReply"`

	// Delivered is the delivery status: "queued", "yes", "no", or "unknown".
	Delivered string `json:"delivered"`

	// Displayed is the display status: "unknown" or "yes".
	Displayed string `json:"displayed"`
}
