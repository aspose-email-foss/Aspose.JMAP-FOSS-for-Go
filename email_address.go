package jmap

// EmailAddress represents one address in a header such as From, To, or Cc.
// It follows the JMAP EmailAddress object definition.
type EmailAddress struct {
	// Name is the display name of the address. It may be null.
	Name *string `json:"name,omitempty"`
	// Email is the email address. This field is required.
	Email string `json:"email"`
}
