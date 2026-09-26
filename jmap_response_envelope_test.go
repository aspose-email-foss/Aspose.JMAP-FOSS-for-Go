package jmap

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

// TestJmapResponseEnvelopeSerialization verifies that a populated JmapResponseEnvelope
// marshals to the expected JSON structure.
func TestJmapResponseEnvelopeSerialization(t *testing.T) {
	env := JmapResponseEnvelope{
		MethodResponses: []Invocation{
			{
				Name: "Mailbox/get",
				Arguments: map[string]interface{}{
					"accountId": "user@example.test",
					"ids":       []string{"id1"},
				},
				MethodCallId: "c1",
			},
		},
		CreatedIds: &map[string]string{
			"clientId": "serverId",
		},
		SessionState: "abc123",
	}

	data, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("unexpected error marshaling JmapResponseEnvelope: %v", err)
	}

	// Decode back into a generic map for easy field inspection.
	var got map[string]json.RawMessage
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unexpected error unmarshaling marshaled JSON: %v", err)
	}

	// methodResponses must be present and decode to the expected array.
	if _, ok := got["methodResponses"]; !ok {
		t.Fatalf("expected methodResponses field in JSON")
	}
	var mr []json.RawMessage
	if err := json.Unmarshal(got["methodResponses"], &mr); err != nil {
		t.Fatalf("failed to unmarshal methodResponses: %v", err)
	}
	if len(mr) != 1 {
		t.Fatalf("expected 1 method response, got %d", len(mr))
	}
	var inv Invocation
	if err := json.Unmarshal(mr[0], &inv); err != nil {
		t.Fatalf("failed to unmarshal Invocation: %v", err)
	}
	if inv.Name != "Mailbox/get" || inv.MethodCallId != "c1" {
		t.Errorf("unexpected Invocation fields: %+v", inv)
	}
	if acct, ok := inv.Arguments["accountId"]; !ok || acct != "user@example.test" {
		t.Errorf("unexpected or missing accountId argument: %v", inv.Arguments)
	}

	// createdIds must be present and match the map.
	if raw, ok := got["createdIds"]; !ok {
		t.Fatalf("expected createdIds field in JSON")
	} else {
		var cm map[string]string
		if err := json.Unmarshal(raw, &cm); err != nil {
			t.Fatalf("failed to unmarshal createdIds: %v", err)
		}
		expected := map[string]string{"clientId": "serverId"}
		if !reflect.DeepEqual(cm, expected) {
			t.Errorf("createdIds mismatch: got %v, want %v", cm, expected)
		}
	}

	// sessionState must be present and correct.
	if raw, ok := got["sessionState"]; !ok {
		t.Fatalf("expected sessionState field in JSON")
	} else {
		var ss string
		if err := json.Unmarshal(raw, &ss); err != nil {
			t.Fatalf("failed to unmarshal sessionState: %v", err)
		}
		if ss != "abc123" {
			t.Errorf("sessionState mismatch: got %q, want %q", ss, "abc123")
		}
	}
}

// TestJmapResponseEnvelopeDeserializationWithNilCreatedIds ensures that a response
// envelope with a null or omitted createdIds field results in a nil pointer.
func TestJmapResponseEnvelopeDeserializationWithNilCreatedIds(t *testing.T) {
	// JSON with createdIds explicitly set to null.
	jsonWithNull := `{
		"methodResponses": [
			["Identity/get", {"accountId":"user@example.test"}, "c2"]
		],
		"createdIds": null,
		"sessionState": "def456"
	}`

	var envNull JmapResponseEnvelope
	if err := json.Unmarshal([]byte(jsonWithNull), &envNull); err != nil {
		t.Fatalf("unexpected error unmarshaling JSON with null createdIds: %v", err)
	}
	if envNull.CreatedIds != nil {
		t.Errorf("expected CreatedIds to be nil when JSON contains null, got %v", *envNull.CreatedIds)
	}
	if envNull.SessionState != "def456" {
		t.Errorf("unexpected SessionState: got %q, want %q", envNull.SessionState, "def456")
	}
	if len(envNull.MethodResponses) != 1 {
		t.Fatalf("expected 1 method response, got %d", len(envNull.MethodResponses))
	}
	if envNull.MethodResponses[0].Name != "Identity/get" {
		t.Errorf("unexpected method name: %q", envNull.MethodResponses[0].Name)
	}

	// JSON with createdIds omitted entirely.
	jsonOmitted := `{
		"methodResponses": [
			["Identity/get", {"accountId":"user@example.test"}, "c3"]
		],
		"sessionState": "ghi789"
	}`

	var envOmitted JmapResponseEnvelope
	if err := json.Unmarshal([]byte(jsonOmitted), &envOmitted); err != nil {
		t.Fatalf("unexpected error unmarshaling JSON without createdIds: %v", err)
	}
	if envOmitted.CreatedIds != nil {
		t.Errorf("expected CreatedIds to be nil when omitted, got %v", *envOmitted.CreatedIds)
	}
	if envOmitted.SessionState != "ghi789" {
		t.Errorf("unexpected SessionState: got %q, want %q", envOmitted.SessionState, "ghi789")
	}
}

// JmapResponseEnvelopeFakeTransport is a minimal Transport implementation used
// to verify that a client method correctly sends a request and receives a
// JmapResponseEnvelope. It records the request it receives and returns a
// pre‑configured response.
type JmapResponseEnvelopeFakeTransport struct {
	Received *HTTPRequest
	Reply    *HTTPResponse
	Err      error
}

func (f *JmapResponseEnvelopeFakeTransport) Send(req *HTTPRequest) (*HTTPResponse, error) {
	f.Received = req
	return f.Reply, f.Err
}

// TestJmapResponseEnvelopeClientIntegration demonstrates a simple client call
// that expects a JmapResponseEnvelope. It uses the fake transport to assert
// that the request payload contains the expected method call.
func TestJmapResponseEnvelopeClientIntegration(t *testing.T) {
	// Prepare a fake response JSON that the client will decode.
	respJSON := `{
		"methodResponses": [
			["Mailbox/get", {"accountId":"user@example.test","list":[]}, "c4"]
		],
		"sessionState": "jkl012"
	}`
	fakeResp := &HTTPResponse{
		StatusCode: 200,
		Body:       []byte(respJSON),
		Headers:    map[string]string{},
	}
	fakeTransport := &JmapResponseEnvelopeFakeTransport{
		Reply: fakeResp,
	}

	// Construct a client with the fake transport.
	opts := ClientOptions{
		SessionURL: "https://example.test/.well-known/jmap",
		Username:   "user@example.test",
		Password:   "secret",
	}
	client := NewClient(opts)
	// Replace the default transport with our fake, and set apiURL directly
	// since sendRequest requires Connect to have been called first.
	client.transport = fakeTransport
	client.apiURL = "https://example.test/api"

	calls := []Invocation{
		{Name: "Mailbox/get", Arguments: map[string]interface{}{"accountId": "user@example.test"}, MethodCallId: "c4"},
	}
	_, err := client.SendRequest(calls, []string{"urn:ietf:params:jmap:mail"})
	if err != nil {
		t.Fatalf("sendRequest returned error: %v", err)
	}

	// Verify that the request sent to the transport matches expectations.
	if fakeTransport.Received == nil {
		t.Fatalf("fake transport did not receive a request")
	}
	if fakeTransport.Received.Method != "POST" {
		t.Errorf("expected POST method, got %s", fakeTransport.Received.Method)
	}
	if !strings.Contains(string(fakeTransport.Received.Body), `"Mailbox/get"`) {
		t.Errorf("request body does not contain expected method name")
	}
}
