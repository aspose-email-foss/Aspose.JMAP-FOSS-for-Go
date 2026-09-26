package jmap

// EmailAddressGroup represents a group of email addresses as used in From/To headers.
// The group syntax is e.g. "Team: a@x, b@x;".
type EmailAddressGroup struct {
	// Name is the optional display name of the group. Null is represented by a nil pointer.
	Name *string `json:"name,omitempty"`
	// Addresses is the list of email addresses belonging to the group.
	Addresses []EmailAddress `json:"addresses"`
}
