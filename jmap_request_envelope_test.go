package jmap

import (
	"encoding/json"
	"reflect"
	"testing"
)

// JmapRequestEnvelopeMapPtr is a helper that returns a pointer to a map[string]string.
// It is uniquely named for this test file to avoid collisions with other test helpers.
func JmapRequestEnvelopeMapPtr(m map[string]string) *map[string]string {
	return &m
}

// TestJmapRequestEnvelopeMarshalWithCreatedIds verifies that a non‑nil CreatedIds
// field is marshaled as a JSON object with the expected contents.
func TestJmapRequestEnvelopeMarshalWithCreatedIds(t *testing.T) {
	env := JmapRequestEnvelope{
		Using: []string{"urn:ietf:params:jmap:core"},
		MethodCalls: []Invocation{
			{
				Name: "Mailbox/get",
				Arguments: map[string]interface{}{
					"accountId": "a1",
				},
				MethodCallId: "c1",
			},
		},
		CreatedIds: JmapRequestEnvelopeMapPtr(map[string]string{
			"client1": "server1",
			"client2": "server2",
		}),
	}

	data, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("unexpected error marshaling envelope: %v", err)
	}

	// Decode back into a generic map to avoid ordering issues.
	var got map[string]interface{}
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unexpected error unmarshaling marshaled JSON: %v", err)
	}

	// Verify required fields.
	if using, ok := got["using"]; !ok {
		t.Errorf("missing 'using' field")
	} else {
		exp := []interface{}{"urn:ietf:params:jmap:core"}
		if !reflect.DeepEqual(using, exp) {
			t.Errorf("'using' mismatch: got %v, want %v", using, exp)
		}
	}

	if mc, ok := got["methodCalls"]; !ok {
		t.Errorf("missing 'methodCalls' field")
	} else {
		// Expected methodCalls is a slice with one three‑element array.
		exp := []interface{}{
			[]interface{}{
				"Mailbox/get",
				map[string]interface{}{"accountId": "a1"},
				"c1",
			},
		}
		if !reflect.DeepEqual(mc, exp) {
			t.Errorf("'methodCalls' mismatch: got %v, want %v", mc, exp)
		}
	}

	if cid, ok := got["createdIds"]; !ok {
		t.Errorf("missing 'createdIds' field")
	} else {
		exp := map[string]interface{}{
			"client1": "server1",
			"client2": "server2",
		}
		if !reflect.DeepEqual(cid, exp) {
			t.Errorf("'createdIds' mismatch: got %v, want %v", cid, exp)
		}
	}
}

// TestJmapRequestEnvelopeMarshalNilCreatedIds verifies that a nil CreatedIds
// pointer is marshaled as JSON null.
func TestJmapRequestEnvelopeMarshalNilCreatedIds(t *testing.T) {
	env := JmapRequestEnvelope{
		Using: []string{"urn:ietf:params:jmap:core"},
		MethodCalls: []Invocation{
			{
				Name: "Identity/get",
				Arguments: map[string]interface{}{
					"accountId": "a2",
				},
				MethodCallId: "c2",
			},
		},
		CreatedIds: nil, // explicit nil
	}

	data, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("unexpected error marshaling envelope: %v", err)
	}

	var got map[string]interface{}
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unexpected error unmarshaling JSON: %v", err)
	}

	if cid, ok := got["createdIds"]; !ok {
		t.Errorf("missing 'createdIds' field")
	} else if cid != nil {
		t.Errorf("'createdIds' expected to be null, got %v", cid)
	}
}

// TestJmapRequestEnvelopeUnmarshalNullCreatedIds verifies that unmarshaling a
// JSON payload with "createdIds":null results in a nil CreatedIds pointer.
func TestJmapRequestEnvelopeUnmarshalNullCreatedIds(t *testing.T) {
	const payload = `{
		"using": ["urn:ietf:params:jmap:core"],
		"methodCalls": [
			["Email/query", {"accountId":"a3","filter":{}}, "c3"]
		],
		"createdIds": null
	}`

	var env JmapRequestEnvelope
	if err := json.Unmarshal([]byte(payload), &env); err != nil {
		t.Fatalf("unexpected error unmarshaling envelope: %v", err)
	}

	if env.CreatedIds != nil {
		t.Errorf("expected CreatedIds to be nil, got %v", *env.CreatedIds)
	}

	// Verify other fields round‑trip correctly.
	if len(env.Using) != 1 || env.Using[0] != "urn:ietf:params:jmap:core" {
		t.Errorf("unexpected Using field: %v", env.Using)
	}
	if len(env.MethodCalls) != 1 {
		t.Fatalf("expected one method call, got %d", len(env.MethodCalls))
	}
	call := env.MethodCalls[0]
	if call.Name != "Email/query" {
		t.Errorf("unexpected method name: %s", call.Name)
	}
	if call.MethodCallId != "c3" {
		t.Errorf("unexpected methodCallId: %s", call.MethodCallId)
	}
	if acct, ok := call.Arguments["accountId"]; !ok || acct != "a3" {
		t.Errorf("unexpected arguments: %v", call.Arguments)
	}
}
