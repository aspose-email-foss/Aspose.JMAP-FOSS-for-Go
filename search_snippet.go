package jmap

// SearchSnippet represents highlighted subject/preview snippet for an Email id matched by a filter's text search.
type SearchSnippet struct {
	EmailId string  `json:"emailId"`           // EmailId is the identifier of the Email.
	Subject *string `json:"subject,omitempty"` // Subject may contain <mark></mark> tags around matches.
	Preview *string `json:"preview,omitempty"` // Preview is a snippet of the email body.
}
