package jmap

import (
	"encoding/json"
	"errors"
)

// Invocation represents a single JMAP method call tuple
// [name, arguments, methodCallId].
//
// It is used inside the public MethodCalls/MethodResponses properties of
// JmapRequestEnvelope and JmapResponseEnvelope.
type Invocation struct {
	// Name is the JMAP method name (e.g. "Mailbox/get").
	Name string `json:"name"`

	// Arguments is a map of method arguments.
	Arguments map[string]interface{} `json:"arguments"`

	// MethodCallId is the client‑chosen identifier for this call.
	MethodCallId string `json:"methodCallId"`
}

// MarshalJSON implements the custom JSON encoding for Invocation as a three‑element array.
func (i Invocation) MarshalJSON() ([]byte, error) {
	return json.Marshal([]interface{}{i.Name, i.Arguments, i.MethodCallId})
}

// UnmarshalJSON implements the custom JSON decoding for Invocation from a three‑element array.
func (i *Invocation) UnmarshalJSON(data []byte) error {
	var raw []json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw) != 3 {
		return errors.New("invocation must be a 3‑element array")
	}
	if err := json.Unmarshal(raw[0], &i.Name); err != nil {
		return err
	}
	if err := json.Unmarshal(raw[1], &i.Arguments); err != nil {
		return err
	}
	if err := json.Unmarshal(raw[2], &i.MethodCallId); err != nil {
		return err
	}
	return nil
}
