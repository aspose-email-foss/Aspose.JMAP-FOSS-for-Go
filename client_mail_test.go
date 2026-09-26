package jmap

import (
	"encoding/json"
	"errors"
	"testing"
)

// ClientMailFakeTransport is a test‑only Transport that returns a canned HTTPResponse
// and records the last HTTPRequest it received.
type ClientMailFakeTransport struct {
	respStatus int
	respBody   []byte
	lastReq    *HTTPRequest
}

func (f *ClientMailFakeTransport) Send(req *HTTPRequest) (*HTTPResponse, error) {
	f.lastReq = req
	return &HTTPResponse{
		StatusCode: f.respStatus,
		Headers:    map[string]string{},
		Body:       f.respBody,
	}, nil
}

// helper to build a minimal JMAP response envelope JSON.
func clientMailBuildEnvelope(methodName string, args interface{}) []byte {
	env := map[string]interface{}{
		"methodResponses": []interface{}{
			[]interface{}{methodName, args, "c1"},
		},
	}
	b, _ := json.Marshal(env)
	return b
}

// Test ListMailboxes normal flow and request verification.
func TestClientMail_ListMailboxes(t *testing.T) {
	// Prepare a successful response payload.
	respArgs := map[string]interface{}{
		"accountId": "u1",
		"state":     "1",
		"list": []interface{}{
			map[string]interface{}{
				"id":   "mb1",
				"name": "Inbox",
			},
		},
		"notFound": []interface{}{},
	}
	fake := &ClientMailFakeTransport{
		respStatus: 200,
		respBody:   clientMailBuildEnvelope("Mailbox/get", respArgs),
	}
	client := &JmapClient{
		opts:    ClientOptions{Username: "u", Password: "p"},
		transport: fake,
		apiURL:  "https://example.com/jmap",
	}

	res, err := client.ListMailboxes("u1")
	if err != nil {
		t.Fatalf("ListMailboxes returned error: %v", err)
	}
	if res.AccountId != "u1" || res.State != "1" || len(res.List) != 1 {
		t.Fatalf("unexpected ListMailboxes result: %+v", res)
	}
	if res.List[0].Id != "mb1" || res.List[0].Name != "Inbox" {
		t.Fatalf("unexpected mailbox data: %+v", res.List[0])
	}

	// Verify the request that was sent.
	if fake.lastReq == nil {
		t.Fatalf("transport did not receive a request")
	}
	if fake.lastReq.Method != "POST" {
		t.Errorf("expected POST method, got %s", fake.lastReq.Method)
	}
	if fake.lastReq.URL != client.apiURL {
		t.Errorf("expected URL %s, got %s", client.apiURL, fake.lastReq.URL)
	}
	var reqEnv struct {
		Using       []string        `json:"using"`
		MethodCalls []json.RawMessage `json:"methodCalls"`
	}
	if err := json.Unmarshal(fake.lastReq.Body, &reqEnv); err != nil {
		t.Fatalf("cannot unmarshal request body: %v", err)
	}
	if len(reqEnv.MethodCalls) != 1 {
		t.Fatalf("expected one method call, got %d", len(reqEnv.MethodCalls))
	}
	var inv Invocation
	if err := json.Unmarshal(reqEnv.MethodCalls[0], &inv); err != nil {
		t.Fatalf("cannot unmarshal invocation: %v", err)
	}
	if inv.Name != "Mailbox/get" {
		t.Errorf("expected method name Mailbox/get, got %s", inv.Name)
	}
	if inv.Arguments["accountId"] != "u1" {
		t.Errorf("expected accountId u1, got %v", inv.Arguments["accountId"])
	}
}

// Test GetMailbox without optional properties (edge case: properties omitted).
func TestClientMail_GetMailbox_NoProperties(t *testing.T) {
	respArgs := map[string]interface{}{
		"accountId": "u1",
		"state":     "2",
		"list": []interface{}{
			map[string]interface{}{
				"id":   "mb2",
				"name": "Sent",
			},
		},
		"notFound": []interface{}{},
	}
	fake := &ClientMailFakeTransport{
		respStatus: 200,
		respBody:   clientMailBuildEnvelope("Mailbox/get", respArgs),
	}
	client := &JmapClient{
		opts:    ClientOptions{Username: "u", Password: "p"},
		transport: fake,
		apiURL:  "https://example.com/jmap",
	}

	_, err := client.GetMailbox("u1", []string{"mb2"}, nil)
	if err != nil {
		t.Fatalf("GetMailbox returned error: %v", err)
	}
	// Ensure the request payload does NOT contain a "properties" field.
	var reqEnv struct {
		MethodCalls []json.RawMessage `json:"methodCalls"`
	}
	if err := json.Unmarshal(fake.lastReq.Body, &reqEnv); err != nil {
		t.Fatalf("cannot unmarshal request body: %v", err)
	}
	var inv Invocation
	if err := json.Unmarshal(reqEnv.MethodCalls[0], &inv); err != nil {
		t.Fatalf("cannot unmarshal invocation: %v", err)
	}
	if _, ok := inv.Arguments["properties"]; ok {
		t.Errorf("properties field should be omitted when empty")
	}
}

// Test CreateMailbox merges server‑assigned fields (Id) into the original create map.
func TestClientMail_CreateMailbox_Merge(t *testing.T) {
	// Per RFC 8620 section 5.3, the server's "created" entry is partial but always
	// includes the real server-assigned id (distinct from "new1", the client-chosen
	// CREATION id used only to correlate this entry back to the request) - this
	// fixture intentionally uses a different id than the creation key so a test that
	// wrongly copies the map key into the merged object's Id (instead of using the
	// id the server actually returned) fails loudly.
	respArgs := map[string]interface{}{
		"accountId": "u1",
		"newState":  "3",
		"created": map[string]interface{}{
			"new1": map[string]interface{}{
				"id":   "mb-server-assigned-42",
				"name": "Drafts",
			},
		},
	}
	fake := &ClientMailFakeTransport{
		respStatus: 200,
		respBody:   clientMailBuildEnvelope("Mailbox/set", respArgs),
	}
	client := &JmapClient{
		opts:    ClientOptions{Username: "u", Password: "p"},
		transport: fake,
		apiURL:  "https://example.com/jmap",
	}

	createMap := map[string]Mailbox{
		"new1": {Name: "Drafts"},
	}
	res, err := client.CreateMailbox("u1", createMap)
	if err != nil {
		t.Fatalf("CreateMailbox returned error: %v", err)
	}
	created, ok := res.Created["new1"]
	if !ok {
		t.Fatalf("created map missing key new1")
	}
	if created.Id != "mb-server-assigned-42" {
		t.Errorf("expected merged Id 'mb-server-assigned-42', got %q", created.Id)
	}
	if created.Name != "Drafts" {
		t.Errorf("expected Name 'Drafts', got %q", created.Name)
	}
}

// Test that a protocol‑level error is returned as *ProtocolError.
func TestClientMail_ListMailboxes_ProtocolError(t *testing.T) {
	errArgs := map[string]interface{}{
		"type":        "accountNotFound",
		"description": "Account missing",
	}
	fake := &ClientMailFakeTransport{
		respStatus: 200,
		respBody:   clientMailBuildEnvelope("error", errArgs),
	}
	client := &JmapClient{
		opts:    ClientOptions{Username: "u", Password: "p"},
		transport: fake,
		apiURL:  "https://example.com/jmap",
	}

	_, err := client.ListMailboxes("u1")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	var perr *ProtocolError
	if !errors.As(err, &perr) {
		t.Fatalf("expected ProtocolError, got %T", err)
	}
	if perr.Type != "accountNotFound" {
		t.Errorf("expected error type accountNotFound, got %s", perr.Type)
	}
	if perr.Description == nil || *perr.Description != "Account missing" {
		t.Errorf("unexpected description: %v", perr.Description)
	}
}

// TestClientMail_MoveMessage_TypedSignature is a regression test: MoveMessage previously
// took a raw map[string]map[string]bool wire-protocol patch, leaking the JMAP PatchObject
// shape (and the "mailboxIds" wrapper key) to callers, unlike Python/TypeScript/Rust which
// expose typed (accountId, ids, destinationMailboxId) parameters. Verifies both the typed
// signature compiles and that it builds the correct wire-format update map.
func TestClientMail_MoveMessage_TypedSignature(t *testing.T) {
	respArgs := map[string]interface{}{
		"accountId": "u1",
		"newState":  "2",
		"updated": map[string]interface{}{
			"e1": map[string]interface{}{},
		},
	}
	fake := &ClientMailFakeTransport{
		respStatus: 200,
		respBody:   clientMailBuildEnvelope("Email/set", respArgs),
	}
	client := &JmapClient{
		opts:      ClientOptions{Username: "u", Password: "p"},
		transport: fake,
		apiURL:    "https://example.com/jmap",
	}

	_, err := client.MoveMessage("u1", []string{"e1", "e2"}, "mb-target")
	if err != nil {
		t.Fatalf("MoveMessage returned error: %v", err)
	}

	var reqEnv struct {
		MethodCalls []json.RawMessage `json:"methodCalls"`
	}
	if err := json.Unmarshal(fake.lastReq.Body, &reqEnv); err != nil {
		t.Fatalf("cannot unmarshal request body: %v", err)
	}
	var inv Invocation
	if err := json.Unmarshal(reqEnv.MethodCalls[0], &inv); err != nil {
		t.Fatalf("cannot unmarshal invocation: %v", err)
	}
	update, ok := inv.Arguments["update"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected update to be a map, got %T", inv.Arguments["update"])
	}
	for _, id := range []string{"e1", "e2"} {
		patch, ok := update[id].(map[string]interface{})
		if !ok {
			t.Fatalf("expected patch for %s to be a map, got %v", id, update[id])
		}
		mailboxIds, ok := patch["mailboxIds"].(map[string]interface{})
		if !ok {
			t.Fatalf("expected mailboxIds map for %s, got %v", id, patch["mailboxIds"])
		}
		if v, ok := mailboxIds["mb-target"].(bool); !ok || !v {
			t.Errorf("expected mailboxIds[%q][mb-target] = true, got %v", id, mailboxIds["mb-target"])
		}
	}
}

// TestClientMail_SetMessageKeyword_TypedSignature is a regression test: same class of fix
// as MoveMessage - SetMessageKeyword now takes typed (accountId, ids, keyword, value)
// parameters instead of a raw wire-protocol map.
func TestClientMail_SetMessageKeyword_TypedSignature(t *testing.T) {
	respArgs := map[string]interface{}{
		"accountId": "u1",
		"newState":  "3",
		"updated": map[string]interface{}{
			"e1": map[string]interface{}{},
		},
	}
	fake := &ClientMailFakeTransport{
		respStatus: 200,
		respBody:   clientMailBuildEnvelope("Email/set", respArgs),
	}
	client := &JmapClient{
		opts:      ClientOptions{Username: "u", Password: "p"},
		transport: fake,
		apiURL:    "https://example.com/jmap",
	}

	_, err := client.SetMessageKeyword("u1", []string{"e1"}, "$seen", true)
	if err != nil {
		t.Fatalf("SetMessageKeyword returned error: %v", err)
	}

	var reqEnv struct {
		MethodCalls []json.RawMessage `json:"methodCalls"`
	}
	if err := json.Unmarshal(fake.lastReq.Body, &reqEnv); err != nil {
		t.Fatalf("cannot unmarshal request body: %v", err)
	}
	var inv Invocation
	if err := json.Unmarshal(reqEnv.MethodCalls[0], &inv); err != nil {
		t.Fatalf("cannot unmarshal invocation: %v", err)
	}
	update, ok := inv.Arguments["update"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected update to be a map, got %T", inv.Arguments["update"])
	}
	patch, ok := update["e1"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected patch for e1 to be a map, got %v", update["e1"])
	}
	keywords, ok := patch["keywords"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected keywords map, got %v", patch["keywords"])
	}
	if v, ok := keywords["$seen"].(bool); !ok || !v {
		t.Errorf("expected keywords[$seen] = true, got %v", keywords["$seen"])
	}
}
