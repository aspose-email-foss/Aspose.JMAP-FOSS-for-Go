package jmap

// Thread represents an ordered list of Email ids that make up a conversation.
type Thread struct {
	// Id is the server-assigned identifier of the thread.
	Id *string `json:"id,omitempty"`

	// EmailIds is the ordered list of Email ids that make up the conversation.
	EmailIds *[]string `json:"emailIds,omitempty"`
}
