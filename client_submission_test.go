package jmap

import (
	"encoding/json"
	"errors"
	"testing"
)

// ClientSubmissionFakeTransport is a test double for the Transport interface.
// It records the request it receives and returns a pre‑configured response.
type ClientSubmissionFakeTransport struct {
	Req        *HTTPRequest
	RespBody   []byte
	RespStatus int
	Err        error
}

func (f *ClientSubmissionFakeTransport) Send(req *HTTPRequest) (*HTTPResponse, error) {
	f.Req = req
	if f.Err != nil {
		return nil, f.Err
	}
	return &HTTPResponse{
		StatusCode: f.RespStatus,
		Headers:    map[string]string{},
		Body:       f.RespBody,
	}, nil
}

// testRequestEnvelope decodes the outgoing request body; MethodCalls reuses the real
// Invocation type since its custom (Un)MarshalJSON matches the actual wire tuple shape.
type testRequestEnvelope struct {
	Using       []string    `json:"using"`
	MethodCalls []Invocation `json:"methodCalls"`
}

// TestSendSuccess verifies that Send creates an EmailSubmission and merges the
// server‑provided fields correctly, and that the generated request contains the
// expected method name and arguments.
func TestSendSuccess(t *testing.T) {
	// Prepare a canned successful response.
	created := map[string]interface{}{
		"id":         "s1",
		"sendAt":     "2026-08-18T10:00:00Z",
		"undoStatus": "final",
	}
	respEnv := JmapResponseEnvelope{
		MethodResponses: []Invocation{
			{
				Name: "EmailSubmission/set",
				Arguments: map[string]interface{}{
					"accountId": "u1",
					"newState":  "1",
					"created": map[string]interface{}{
						"c1": created,
					},
				},
				MethodCallId: "c1",
			},
		},
	}
	respBody, _ := json.Marshal(respEnv)

	fake := &ClientSubmissionFakeTransport{
		RespBody:   respBody,
		RespStatus: 200,
	}
	opts := ClientOptions{
		Username: "u",
		Password: "p",
		Transport: fake,
	}
	client := NewClient(opts)
	client.apiURL = "https://example.com/jmap"

	sub := EmailSubmission{
		IdentityId: "id1",
		EmailId:    "e1",
	}
	got, err := client.Send("u1", sub)
	if err != nil {
		t.Fatalf("Send returned error: %v", err)
	}
	if got.Id == nil || *got.Id != "s1" {
		t.Errorf("Send returned Id = %v, want %q", got.Id, "s1")
	}
	if got.SendAt == nil || *got.SendAt != "2026-08-18T10:00:00Z" {
		t.Errorf("Send returned SendAt = %v, want %q", got.SendAt, "2026-08-18T10:00:00Z")
	}
	if got.UndoStatus == nil || *got.UndoStatus != "final" {
		t.Errorf("Send returned UndoStatus = %v, want %q", got.UndoStatus, "final")
	}

	// Verify the request payload.
	if fake.Req == nil {
		t.Fatalf("transport did not receive a request")
	}
	var reqEnv testRequestEnvelope
	if err := json.Unmarshal(fake.Req.Body, &reqEnv); err != nil {
		t.Fatalf("failed to unmarshal request body: %v", err)
	}
	if len(reqEnv.Using) != 1 || reqEnv.Using[0] != "urn:ietf:params:jmap:submission" {
		t.Errorf("request Using = %v, want [%q]", reqEnv.Using, "urn:ietf:params:jmap:submission")
	}
	if len(reqEnv.MethodCalls) != 1 {
		t.Fatalf("expected 1 method call, got %d", len(reqEnv.MethodCalls))
	}
	inv := reqEnv.MethodCalls[0]
	if inv.Name != "EmailSubmission/set" {
		t.Errorf("Invocation Name = %q, want %q", inv.Name, "EmailSubmission/set")
	}
	if inv.Arguments["accountId"] != "u1" {
		t.Errorf("Invocation accountId = %v, want %q", inv.Arguments["accountId"], "u1")
	}
	createMap, ok := inv.Arguments["create"].(map[string]interface{})
	if !ok {
		t.Fatalf("Invocation create argument has unexpected type")
	}
	if _, ok := createMap["c1"]; !ok {
		t.Errorf("create map missing key \"c1\"")
	}
}

// TestCancelSendSuccess ensures that CancelSend does not return an error when the
// server responds without an explicit error invocation.
func TestCancelSendSuccess(t *testing.T) {
	respEnv := JmapResponseEnvelope{
		MethodResponses: []Invocation{
			{
				Name: "EmailSubmission/set",
				Arguments: map[string]interface{}{
					"accountId": "u1",
					"newState":  "2",
				},
				MethodCallId: "c1",
			},
		},
	}
	respBody, _ := json.Marshal(respEnv)

	fake := &ClientSubmissionFakeTransport{
		RespBody:   respBody,
		RespStatus: 200,
	}
	client := NewClient(ClientOptions{
		Username: "u",
		Password: "p",
		Transport: fake,
	})
	client.apiURL = "https://example.com/jmap"

	if err := client.CancelSend("u1", "sub123"); err != nil {
		t.Fatalf("CancelSend returned error: %v", err)
	}
	if fake.Req == nil {
		t.Fatalf("transport did not receive a request")
	}
	var reqEnv testRequestEnvelope
	if err := json.Unmarshal(fake.Req.Body, &reqEnv); err != nil {
		t.Fatalf("failed to unmarshal request body: %v", err)
	}
	if len(reqEnv.MethodCalls) != 1 || reqEnv.MethodCalls[0].Name != "EmailSubmission/set" {
		t.Errorf("unexpected method call in request: %+v", reqEnv.MethodCalls)
	}
	// Regression test: a PatchObject key is a JSON Pointer (RFC 6901) relative to the
	// object being patched - a bare top-level property name has no leading slash.
	// "/undoStatus" instead points at a property literally named the empty string, which
	// a real JMAP server rejects/ignores, silently breaking cancel-send.
	update, _ := reqEnv.MethodCalls[0].Arguments["update"].(map[string]interface{})
	patch, _ := update["sub123"].(map[string]interface{})
	if _, ok := patch["undoStatus"]; !ok {
		t.Errorf("expected patch to have key \"undoStatus\" (no leading slash), got: %+v", patch)
	}
	if _, ok := patch["/undoStatus"]; ok {
		t.Errorf("patch must not use \"/undoStatus\" as a key - a PatchObject key is a JSON "+
			"Pointer relative to the patched object, so a leading slash points at a "+
			"differently-named property, not \"undoStatus\" itself: %+v", patch)
	}
}

// TestListSubmissionsEmpty verifies that ListSubmissions correctly returns an
// empty slice when the server returns an empty list.
func TestListSubmissionsEmpty(t *testing.T) {
	respEnv := JmapResponseEnvelope{
		MethodResponses: []Invocation{
			{
				Name: "EmailSubmission/get",
				Arguments: map[string]interface{}{
					"accountId": "u1",
					"list":      []interface{}{},
				},
				MethodCallId: "c1",
			},
		},
	}
	respBody, _ := json.Marshal(respEnv)

	fake := &ClientSubmissionFakeTransport{
		RespBody:   respBody,
		RespStatus: 200,
	}
	client := NewClient(ClientOptions{
		Username: "u",
		Password: "p",
		Transport: fake,
	})
	client.apiURL = "https://example.com/jmap"

	got, err := client.ListSubmissions("u1", nil, nil)
	if err != nil {
		t.Fatalf("ListSubmissions returned error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %d elements", len(got))
	}
}

// TestSendProtocolError ensures that a protocol‑level error invocation is
// surfaced as a *ProtocolError.
func TestSendProtocolError(t *testing.T) {
	respEnv := JmapResponseEnvelope{
		MethodResponses: []Invocation{
			{
				Name: "error",
				Arguments: map[string]interface{}{
					"type":        "unknownMethod",
					"description": "method not found",
				},
				MethodCallId: "c1",
			},
		},
	}
	respBody, _ := json.Marshal(respEnv)

	fake := &ClientSubmissionFakeTransport{
		RespBody:   respBody,
		RespStatus: 200,
	}
	client := NewClient(ClientOptions{
		Username: "u",
		Password: "p",
		Transport: fake,
	})
	client.apiURL = "https://example.com/jmap"

	_, err := client.Send("u1", EmailSubmission{IdentityId: "id1", EmailId: "e1"})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	var perr *ProtocolError
	if !errors.As(err, &perr) {
		t.Fatalf("expected ProtocolError, got %T: %v", err, err)
	}
	if perr.Type != "unknownMethod" {
		t.Errorf("ProtocolError Type = %q, want %q", perr.Type, "unknownMethod")
	}
	if perr.Description == nil || *perr.Description != "method not found" {
		t.Errorf("ProtocolError Description = %v, want %q", perr.Description, "method not found")
	}
}
