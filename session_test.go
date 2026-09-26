package jmap

import (
	"encoding/json"
	"reflect"
	"testing"
)

// SessionFakeTransport is a test‑only Transport that records the request
// and returns a pre‑configured response.
type SessionFakeTransport struct {
	Req  *HTTPRequest
	Resp *HTTPResponse
	Err  error
}

// Send implements the Transport interface.
func (f *SessionFakeTransport) Send(req *HTTPRequest) (*HTTPResponse, error) {
	f.Req = req
	return f.Resp, f.Err
}

// TestSessionSerialization verifies that a Session value round‑trips
// through JSON (encoding/json) without loss.
func TestSessionSerialization(t *testing.T) {
	orig := Session{
		// Numeric/slice values are typed to match what json.Unmarshal produces into a
		// map[string]any (numbers always decode as float64, slices as []any) so the
		// round-trip comparison below via reflect.DeepEqual is apples-to-apples.
		Capabilities: map[string]any{
			"urn:ietf:params:jmap:core": map[string]any{
				"maxSizeUpload":         float64(10),
				"maxConcurrentUpload":   float64(2),
				"maxSizeRequest":        float64(20),
				"maxConcurrentRequests": float64(3),
				"maxCallsInRequest":     float64(4),
				"maxObjectsInGet":       float64(5),
				"maxObjectsInSet":       float64(6),
				"collationAlgorithms":   []any{"i;unicode-casemap"},
			},
		},
		Accounts: map[string]Account{
			"acc1": {
				Name:       "Test Account",
				IsPersonal: true,
				IsReadOnly: false,
				AccountCapabilities: map[string]any{
					"urn:ietf:params:jmap:mail": map[string]any{},
				},
			},
		},
		PrimaryAccounts: map[string]string{
			"urn:ietf:params:jmap:mail": "acc1",
		},
		Username:        "user@example.test",
		ApiUrl:          "https://example.com/jmap/",
		DownloadUrl:     "https://example.com/download/{accountId}/{blobId}/{type}/{name}",
		UploadUrl:       "https://example.com/upload/{accountId}",
		EventSourceUrl:  "https://example.com/events",
		State:           "state123",
		origin:          "https://example.com",
	}

	data, err := json.Marshal(&orig)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded Session
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// origin is not part of the wire format; it should be zero after unmarshal.
	if decoded.origin != "" {
		t.Errorf("expected origin to be empty after unmarshal, got %q", decoded.origin)
	}

	// Compare all exported fields.
	if !reflect.DeepEqual(orig.Capabilities, decoded.Capabilities) {
		t.Errorf("Capabilities mismatch: got %+v, want %+v", decoded.Capabilities, orig.Capabilities)
	}
	if !reflect.DeepEqual(orig.Accounts, decoded.Accounts) {
		t.Errorf("Accounts mismatch: got %+v, want %+v", decoded.Accounts, orig.Accounts)
	}
	if !reflect.DeepEqual(orig.PrimaryAccounts, decoded.PrimaryAccounts) {
		t.Errorf("PrimaryAccounts mismatch: got %+v, want %+v", decoded.PrimaryAccounts, orig.PrimaryAccounts)
	}
	if orig.Username != decoded.Username {
		t.Errorf("Username mismatch: got %q, want %q", decoded.Username, orig.Username)
	}
	if orig.ApiUrl != decoded.ApiUrl {
		t.Errorf("ApiUrl mismatch: got %q, want %q", decoded.ApiUrl, orig.ApiUrl)
	}
	if orig.DownloadUrl != decoded.DownloadUrl {
		t.Errorf("DownloadUrl mismatch: got %q, want %q", decoded.DownloadUrl, orig.DownloadUrl)
	}
	if orig.UploadUrl != decoded.UploadUrl {
		t.Errorf("UploadUrl mismatch: got %q, want %q", decoded.UploadUrl, orig.UploadUrl)
	}
	if orig.EventSourceUrl != decoded.EventSourceUrl {
		t.Errorf("EventSourceUrl mismatch: got %q, want %q", decoded.EventSourceUrl, orig.EventSourceUrl)
	}
	if orig.State != decoded.State {
		t.Errorf("State mismatch: got %q, want %q", decoded.State, orig.State)
	}
}

// TestConnectSuccess verifies that Connect fetches the session, resolves
// relative URLs, caches the result, and issues the correct HTTP request.
func TestConnectSuccess(t *testing.T) {
	const sessionURL = "https://example.com/.well-known/jmap"

	// Session JSON with relative URLs.
	sessionJSON := `{
		"capabilities": {
			"urn:ietf:params:jmap:core": {
				"maxSizeUpload": 10,
				"maxConcurrentUpload": 2,
				"maxSizeRequest": 20,
				"maxConcurrentRequests": 3,
				"maxCallsInRequest": 4,
				"maxObjectsInGet": 5,
				"maxObjectsInSet": 6,
				"collationAlgorithms": ["i;unicode-casemap"]
			}
		},
		"accounts": {},
		"primaryAccounts": {},
		"username": "user@example.test",
		"apiUrl": "/jmap/",
		"downloadUrl": "/download/{accountId}/{blobId}/{type}/{name}",
		"uploadUrl": "/upload/{accountId}",
		"eventSourceUrl": "/events",
		"state": "state123"
	}`

	fake := &SessionFakeTransport{
		Resp: &HTTPResponse{
			StatusCode: 200,
			Body:       []byte(sessionJSON),
		},
	}

	// Build a client with the fake transport.
	opts := ClientOptions{
		SessionURL: sessionURL,
		Username:   "user@example.test",
		Password:   "secret",
	}
	client := NewClient(opts)
	client.transport = fake

	// First call – should hit the transport.
	sess, err := client.Connect()
	if err != nil {
		t.Fatalf("Connect returned error: %v", err)
	}
	if sess == nil {
		t.Fatalf("Connect returned nil session")
	}

	// Verify request details.
	if fake.Req == nil {
		t.Fatalf("Transport did not receive a request")
	}
	if fake.Req.Method != "GET" {
		t.Errorf("expected method GET, got %s", fake.Req.Method)
	}
	if fake.Req.URL != sessionURL {
		t.Errorf("expected URL %s, got %s", sessionURL, fake.Req.URL)
	}
	if got, want := fake.Req.Headers["Accept"], "application/json"; got != want {
		t.Errorf("expected Accept header %q, got %q", want, got)
	}

	// Verify URL resolution.
	if sess.ApiUrl != "https://example.com/jmap/" {
		t.Errorf("ApiUrl resolved incorrectly: got %q", sess.ApiUrl)
	}
	if sess.DownloadUrl != "https://example.com/download/{accountId}/{blobId}/{type}/{name}" {
		t.Errorf("DownloadUrl resolved incorrectly: got %q", sess.DownloadUrl)
	}
	if sess.UploadUrl != "https://example.com/upload/{accountId}" {
		t.Errorf("UploadUrl resolved incorrectly: got %q", sess.UploadUrl)
	}
	if sess.EventSourceUrl != "https://example.com/events" {
		t.Errorf("EventSourceUrl resolved incorrectly: got %q", sess.EventSourceUrl)
	}

	// Second call - must re-fetch (not return a cached session): the session's `state`
	// field (RFC 8620 section 2) exists precisely because session data can change
	// server-side, and every other language target in this project (Python, TypeScript,
	// Rust) re-fetches on every Connect() call too.
	fake.Req = nil // clear to detect the new request
	sess2, err := client.Connect()
	if err != nil {
		t.Fatalf("second Connect returned error: %v", err)
	}
	if sess2 == nil {
		t.Fatalf("second Connect returned nil session")
	}
	if fake.Req == nil {
		t.Errorf("expected a fresh request on second Connect, but transport was not called")
	}
}

// TestConnectErrorStatus verifies that a non‑200 HTTP status results in an error.
func TestConnectErrorStatus(t *testing.T) {
	const sessionURL = "https://example.com/.well-known/jmap"

	fake := &SessionFakeTransport{
		Resp: &HTTPResponse{
			StatusCode: 500,
			Body:       []byte(`{}`),
		},
	}

	opts := ClientOptions{
		SessionURL: sessionURL,
		Username:   "user@example.test",
		Password:   "secret",
	}
	client := NewClient(opts)
	client.transport = fake

	_, err := client.Connect()
	if err == nil {
		t.Fatalf("expected error for non‑200 status, got nil")
	}
}
