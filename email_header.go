package jmap

// EmailHeader represents an email header with name and value.
type EmailHeader struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
