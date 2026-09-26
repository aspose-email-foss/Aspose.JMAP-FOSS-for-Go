package jmap

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

// TestInvocationMarshalJSON verifies that MarshalJSON produces the expected three‑element array.
func TestInvocationMarshalJSON(t *testing.T) {
	inv := Invocation{
		Name:         "Mailbox/get",
		Arguments:    map[string]interface{}{"accountId": "user-123", "ids": []interface{}{"id1", "id2"}},
		MethodCallId: "call-001",
	}
	data, err := json.Marshal(inv)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	expected := `["Mailbox/get",{"accountId":"user-123","ids":["id1","id2"]},"call-001"]`
	// json.Marshal may emit map keys in any order; normalize both strings for comparison.
	compact := func(s string) string {
		var buf bytes.Buffer
		if err := json.Compact(&buf, []byte(s)); err != nil {
			return s
		}
		return buf.String()
	}
	if got, want := compact(string(data)), compact(expected); got != want {
		t.Errorf("MarshalJSON = %s, want %s", got, want)
	}
}

// TestInvocationUnmarshalJSON verifies that a valid three‑element array is decoded correctly.
func TestInvocationUnmarshalJSON(t *testing.T) {
	payload := `["Email/set",{"accountId":"user-456","create":{"msg1":{"subject":"Hello"}}},"call-002"]`
	var inv Invocation
	if err := json.Unmarshal([]byte(payload), &inv); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if inv.Name != "Email/set" {
		t.Errorf("Name = %s, want %s", inv.Name, "Email/set")
	}
	if inv.MethodCallId != "call-002" {
		t.Errorf("MethodCallId = %s, want %s", inv.MethodCallId, "call-002")
	}
	// Verify a nested argument.
	if acct, ok := inv.Arguments["accountId"]; !ok || acct != "user-456" {
		t.Errorf("Arguments[accountId] = %v, want %v", acct, "user-456")
	}
	create, ok := inv.Arguments["create"].(map[string]interface{})
	if !ok {
		t.Fatalf("Arguments[create] missing or wrong type")
	}
	if _, ok := create["msg1"]; !ok {
		t.Errorf("create map missing key 'msg1'")
	}
}

// TestInvocationUnmarshalInvalidLength ensures that malformed arrays return an error.
func TestInvocationUnmarshalInvalidLength(t *testing.T) {
	invalid := `["onlyOneElement"]`
	var inv Invocation
	err := json.Unmarshal([]byte(invalid), &inv)
	if err == nil {
		t.Fatalf("Expected error for invalid array length, got nil")
	}
	expectedMsg := "invocation must be a 3‑element array"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Error message = %q, want to contain %q", err.Error(), expectedMsg)
	}
}
