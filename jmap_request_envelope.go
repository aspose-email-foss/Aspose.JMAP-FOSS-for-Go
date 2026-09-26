package jmap

// JmapRequestEnvelope represents the JSON body posted to the JMAP session API URL.
// It includes the capabilities used, the method calls to invoke, and an optional
// mapping of client‑generated IDs to server‑assigned IDs.
type JmapRequestEnvelope struct {
	// Using is a required list of capability URNs this request depends on.
	// It must include "urn:ietf:params:jmap:core".
	Using []string `json:"using"`

	// MethodCalls is a required list of method invocations to be performed.
	MethodCalls []Invocation `json:"methodCalls"`

	// CreatedIds is an optional map from client‑generated IDs to server‑assigned IDs.
	// A nil pointer encodes as JSON null.
	CreatedIds *map[string]string `json:"createdIds"`
}
