package jmap

// Envelope represents the SMTP MAIL FROM / RCPT TO envelope for a submission,
// distinct from the message's own From/To headers.
type Envelope struct {
	// MailFrom is the SMTP MAIL FROM address.
	MailFrom Address `json:"mailFrom"`

	// RcptTo is the list of SMTP RCPT TO addresses.
	RcptTo []Address `json:"rcptTo"`
}

// Address represents an email address used in an envelope, together with optional
// SMTP parameters.
type Address struct {
	// Email is the email address string.
	Email string `json:"email"`

	// Parameters are optional SMTP parameters for the address, such as {"RET":"HDRS"}.
	// The map values may be nil to represent a null value, and the entire map may be nil.
	Parameters map[string]*string `json:"parameters,omitempty"`
}
