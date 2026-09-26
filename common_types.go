package jmap

import "fmt"

// SetError represents a per‑id error returned in the notCreated, notUpdated or
// notDestroyed objects of a JMAP Set response.
type SetError struct {
	// Type is a machine‑readable error identifier, e.g. "invalidProperties",
	// "notFound", "forbidden", "tooLarge".
	Type string `json:"type"`

	// Description is a human‑readable description of the error, or nil if not
	// provided.
	Description *string `json:"description,omitempty"`

	// Properties lists the property names that caused the error, or nil if not
	// applicable.
	Properties *[]string `json:"properties,omitempty"`
}

// MethodError represents a top‑level error response for a JMAP method call.
type MethodError struct {
	// Type is a machine‑readable error identifier, e.g. "unknownMethod",
	// "invalidArguments", "accountNotFound", "serverFail".
	Type string `json:"type"`

	// Description is a human‑readable description of the error, or nil if not
	// provided.
	Description *string `json:"description,omitempty"`
}

// ResultReference is a back‑reference used inside a request argument to point
// at a value produced by an earlier method call in the same request.
type ResultReference struct {
	// ResultOf is the name of the earlier method call whose result is being
	// referenced.
	ResultOf string `json:"resultOf"`

	// Name is the name of the property within that method's result to use.
	Name string `json:"name"`

	// Path is a JSON Pointer into the referenced result.
	Path string `json:"path"`
}

// JmapError is the common base type for all JMAP‑related errors.
type JmapError struct{}

// ProtocolError represents an error returned by the JMAP server in a method
// response's "error" invocation.
type ProtocolError struct {
	JmapError
	// Type is the protocol‑level error identifier.
	Type string `json:"type"`
	// Description is an optional human‑readable description.
	Description *string `json:"description,omitempty"`
}

// Error implements the error interface for ProtocolError.
func (e *ProtocolError) Error() string {
	if e.Description != nil && *e.Description != "" {
		return fmt.Sprintf("%s: %s", e.Type, *e.Description)
	}
	return e.Type
}

// NetworkError represents a failure in the transport layer (e.g., network
// connectivity, HTTP status errors, JSON decoding problems).
type NetworkError struct {
	JmapError
	// Err is the underlying error that caused the network failure.
	Err error `json:"-"`
}

// Error implements the error interface for NetworkError.
func (e *NetworkError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return "network error"
}

// Unwrap returns the underlying error, enabling errors.Is / errors.As.
func (e *NetworkError) Unwrap() error {
	return e.Err
}

// PatchObject (inline documentation only): a JSON object where each key is a JSON
// Pointer (RFC 6901) relative to the object being patched, and each value is either
// the new value to set at that path or null to remove the path. Used as the `update`
// argument shape in every object type's "set" method (e.g. Mailbox/set, Email/set).
// This type has no Go representation; callers construct it as a map[string]interface{}
// when needed.
