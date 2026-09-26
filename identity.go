package jmap

// Identity represents a sending identity (name/email/replyTo used as the From when submitting mail).
type Identity struct {
	Id            string           `json:"id,omitempty"`
	Name          string           `json:"name,omitempty"`
	Email         string           `json:"email"`
	ReplyTo       *[]EmailAddress  `json:"replyTo,omitempty"`
	Bcc           *[]EmailAddress  `json:"bcc,omitempty"`
	TextSignature string           `json:"textSignature,omitempty"`
	HtmlSignature string           `json:"htmlSignature,omitempty"`
	MayDelete     *bool            `json:"mayDelete,omitempty"`
}
