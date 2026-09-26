package jmap

// JmapResponseEnvelope represents the JSON body returned from the JMAP API URL.
// It contains the method responses, optional created IDs map, and the session state.
type JmapResponseEnvelope struct {
	// MethodResponses is a list of method invocation results.
	MethodResponses []Invocation `json:"methodResponses"`

	// CreatedIds is a map from client‑supplied IDs to server‑assigned IDs.
	// It is nullable; a nil value indicates the property was omitted or null.
	CreatedIds *map[string]string `json:"createdIds,omitempty"`

	// SessionState is the current session state string.
	SessionState string `json:"sessionState"`
}
