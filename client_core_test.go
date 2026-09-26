package jmap

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

// ClientCoreFakeTransport is a minimal Transport implementation for tests.
// It records the last request it received and returns a pre‑configured response.
type ClientCoreFakeTransport struct {
	LastRequest *HTTPRequest
	Response    *HTTPResponse
	Err         error
}

func (f *ClientCoreFakeTransport) Send(req *HTTPRequest) (*HTTPResponse, error) {
	f.LastRequest = req
	if f.Err != nil {
		return nil, f.Err
	}
	return f.Response, nil
}

// helper to create a basic auth header matching the client implementation.
func clientCoreBasicAuth(user, pass string) string {
	cred := user + ":" + pass
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(cred))
}

// Test Connect with absolute and relative URLs.
func TestClientCoreConnect(t *testing.T) {
	const sessionURL = "https://example.com/.well-known/jmap"
	const username = "alice"
	const password = "secret"

	// Session JSON with a relative upload URL to test resolution.
	sessionJSON := `{
		"apiUrl": "https://example.com/jmap",
		"uploadUrl": "/upload/{accountId}",
		"downloadUrl": "/download/{accountId}/{blobId}/{type}/{name}",
		"eventSourceUrl": "/events"
	}`

	fake := &ClientCoreFakeTransport{
		Response: &HTTPResponse{
			StatusCode: 200,
			Headers:    map[string]string{"Content-Type": "application/json"},
			Body:       []byte(sessionJSON),
		},
	}

	opts := ClientOptions{
		SessionURL: sessionURL,
		Username:   username,
		Password:   password,
		Transport:  fake,
	}
	client := NewClient(opts)

	sess, err := client.Connect()
	if err != nil {
		t.Fatalf("Connect returned error: %v", err)
	}
	if sess == nil {
		t.Fatalf("Connect returned nil session")
	}
	// Verify that the request used basic auth.
	if got := fake.LastRequest.Headers["Authorization"]; got != clientCoreBasicAuth(username, password) {
		t.Errorf("Authorization header = %q, want %q", got, clientCoreBasicAuth(username, password))
	}
	// Verify URL resolution.
	expectedUpload := "https://example.com/upload/{accountId}"
	if client.uploadURL != expectedUpload {
		t.Errorf("uploadURL = %q, want %q", client.uploadURL, expectedUpload)
	}
	expectedDownload := "https://example.com/download/{accountId}/{blobId}/{type}/{name}"
	if client.downloadURL != expectedDownload {
		t.Errorf("downloadURL = %q, want %q", client.downloadURL, expectedDownload)
	}
}

// TestClientCoreConnectBearerToken verifies that when ClientOptions.BearerToken is set, the
// outgoing Authorization header uses the "Bearer <token>" scheme (RFC 6750) instead of HTTP
// Basic auth, even though Username/Password are also populated (BearerToken must take
// precedence).
func TestClientCoreConnectBearerToken(t *testing.T) {
	const sessionURL = "https://example.com/.well-known/jmap"
	const token = "oauth-access-token-123"

	sessionJSON := `{
		"apiUrl": "https://example.com/jmap",
		"uploadUrl": "/upload/{accountId}",
		"downloadUrl": "/download/{accountId}/{blobId}/{type}/{name}",
		"eventSourceUrl": "/events"
	}`

	fake := &ClientCoreFakeTransport{
		Response: &HTTPResponse{
			StatusCode: 200,
			Headers:    map[string]string{"Content-Type": "application/json"},
			Body:       []byte(sessionJSON),
		},
	}

	opts := ClientOptions{
		SessionURL: sessionURL,
		// Populate Username/Password too, to confirm BearerToken wins when both are set.
		Username:    "alice",
		Password:    "secret",
		BearerToken: token,
		Transport:   fake,
	}
	client := NewClient(opts)

	sess, err := client.Connect()
	if err != nil {
		t.Fatalf("Connect returned error: %v", err)
	}
	if sess == nil {
		t.Fatalf("Connect returned nil session")
	}

	want := "Bearer " + token
	if got := fake.LastRequest.Headers["Authorization"]; got != want {
		t.Errorf("Authorization header = %q, want %q", got, want)
	}
}

// countingTransport records how many times Send was called; used to verify Connect()
// re-fetches the session on every call rather than caching-and-short-circuiting.
type countingTransport struct {
	Response *HTTPResponse
	Calls    int
}

func (f *countingTransport) Send(req *HTTPRequest) (*HTTPResponse, error) {
	f.Calls++
	return f.Response, nil
}

// TestClientCoreConnectAlwaysRefetches is a regression test: a prior version cached the
// Session after the first Connect() call and short-circuited every subsequent call,
// returning the stale session without ever contacting the server again - inconsistent with
// Python/TypeScript/Rust, which all re-fetch on every Connect() call. The session's `state`
// field (RFC 8620 section 2) exists precisely because session data can change server-side.
func TestClientCoreConnectAlwaysRefetches(t *testing.T) {
	sessionJSON := `{
		"apiUrl": "https://example.com/jmap",
		"uploadUrl": "/upload/{accountId}",
		"downloadUrl": "/download/{accountId}/{blobId}/{type}/{name}",
		"eventSourceUrl": "/events"
	}`
	fake := &countingTransport{
		Response: &HTTPResponse{
			StatusCode: 200,
			Headers:    map[string]string{"Content-Type": "application/json"},
			Body:       []byte(sessionJSON),
		},
	}
	client := NewClient(ClientOptions{
		SessionURL: "https://example.com/.well-known/jmap",
		Username:   "alice",
		Password:   "secret",
		Transport:  fake,
	})

	if _, err := client.Connect(); err != nil {
		t.Fatalf("first Connect returned error: %v", err)
	}
	if _, err := client.Connect(); err != nil {
		t.Fatalf("second Connect returned error: %v", err)
	}

	if fake.Calls != 2 {
		t.Errorf("transport.Send called %d times across two Connect() calls, want 2 (no caching)", fake.Calls)
	}
}

// TestClientCoreConnect_NonOKStatusIncludesBody is a regression test: a prior version
// discarded the response body entirely on a non-2xx status, keeping only the bare status
// code. Real JMAP servers commonly return an RFC 8620 section 3.6.1 "problem details" JSON
// body (type/title/detail) explaining exactly what went wrong - silently dropping it makes
// debugging a failed request much harder than necessary.
func TestClientCoreConnect_NonOKStatusIncludesBody(t *testing.T) {
	fake := &ClientCoreFakeTransport{
		Response: &HTTPResponse{
			StatusCode: 400,
			Headers:    map[string]string{"Content-Type": "application/problem+json"},
			Body:       []byte(`{"type":"urn:ietf:params:jmap:error:notJSON","detail":"bad request body"}`),
		},
	}
	client := NewClient(ClientOptions{
		SessionURL: "https://example.com/.well-known/jmap",
		Username:   "u",
		Password:   "p",
		Transport:  fake,
	})

	_, err := client.Connect()
	if err == nil {
		t.Fatalf("expected an error, got nil")
	}
	if !strings.Contains(err.Error(), "bad request body") {
		t.Errorf("expected error to include the response body, got: %v", err)
	}
}

// Test successful Echo call.
func TestClientCoreEchoSuccess(t *testing.T) {
	const apiURL = "https://example.com/jmap"
	const username = "bob"
	const password = "pwd"

	// Prepare the response envelope JSON.
	respEnv := JmapResponseEnvelope{
		MethodResponses: []Invocation{
			{
				Name:        "Core/echo",
				Arguments:   map[string]interface{}{"hello": true, "high": float64(5)},
				MethodCallId: "c1",
			},
		},
	}
	respBody, _ := json.Marshal(respEnv)

	fake := &ClientCoreFakeTransport{
		Response: &HTTPResponse{
			StatusCode: 200,
			Headers:    map[string]string{"Content-Type": "application/json"},
			Body:       respBody,
		},
	}

	opts := ClientOptions{
		SessionURL: "https://example.com/.well-known/jmap",
		Username:   username,
		Password:   password,
		Transport:  fake,
	}
	client := NewClient(opts)
	// Manually set apiURL to avoid a real Connect call.
	client.apiURL = apiURL

	args := map[string]interface{}{"hello": true, "high": 5}
	got, err := client.Echo(args)
	if err != nil {
		t.Fatalf("Echo returned error: %v", err)
	}
	if got["hello"] != true || got["high"] != float64(5) {
		t.Errorf("Echo returned %v, want %v", got, args)
	}
	// Verify request payload.
	if fake.LastRequest == nil {
		t.Fatalf("No request was sent")
	}
	if fake.LastRequest.Method != "POST" {
		t.Errorf("Echo request method = %s, want POST", fake.LastRequest.Method)
	}
	if fake.LastRequest.URL != apiURL {
		t.Errorf("Echo request URL = %s, want %s", fake.LastRequest.URL, apiURL)
	}
	var sentEnv JmapRequestEnvelope
	if err := json.Unmarshal(fake.LastRequest.Body, &sentEnv); err != nil {
		t.Fatalf("Failed to unmarshal request body: %v", err)
	}
	if len(sentEnv.MethodCalls) != 1 {
		t.Fatalf("Expected 1 method call, got %d", len(sentEnv.MethodCalls))
	}
	if sentEnv.MethodCalls[0].Name != "Core/echo" {
		t.Errorf("Method name = %s, want Core/echo", sentEnv.MethodCalls[0].Name)
	}
	if sentEnv.MethodCalls[0].Arguments["hello"] != true {
		t.Errorf("Argument hello = %v, want true", sentEnv.MethodCalls[0].Arguments["hello"])
	}
}

// Test Echo handling of a protocol error response.
func TestClientCoreEchoProtocolError(t *testing.T) {
	const apiURL = "https://example.com/jmap"
	const username = "carol"
	const password = "pwd"

	// Error response envelope.
	errResp := JmapResponseEnvelope{
		MethodResponses: []Invocation{
			{
				Name: "error",
				Arguments: map[string]interface{}{
					"type":        "unknownMethod",
					"description": "Method not supported",
				},
				MethodCallId: "c1",
			},
		},
	}
	body, _ := json.Marshal(errResp)

	fake := &ClientCoreFakeTransport{
		Response: &HTTPResponse{
			StatusCode: 200,
			Headers:    map[string]string{"Content-Type": "application/json"},
			Body:       body,
		},
	}

	opts := ClientOptions{
		SessionURL: "https://example.com/.well-known/jmap",
		Username:   username,
		Password:   password,
		Transport:  fake,
	}
	client := NewClient(opts)
	client.apiURL = apiURL

	_, err := client.Echo(map[string]interface{}{})
	if err == nil {
		t.Fatalf("Expected error, got nil")
	}
	var pErr *ProtocolError
	if !errors.As(err, &pErr) {
		t.Fatalf("Error is not ProtocolError: %v", err)
	}
	if pErr.Type != "unknownMethod" || pErr.Description == nil || *pErr.Description != "Method not supported" {
		t.Errorf("ProtocolError = %+v, want type=unknownMethod description=Method not supported", pErr)
	}
}

// Test UploadBlob uploads data and parses the response.
func TestClientCoreUploadBlob(t *testing.T) {
	const uploadURL = "https://example.com/upload/{accountId}"
	const username = "dave"
	const password = "pwd"

	upResp := UploadResponse{
		AccountId: "account123",
		BlobId:    "blob456",
		Type:      "text/plain",
		Size:      12,
	}
	respBody, _ := json.Marshal(upResp)

	fake := &ClientCoreFakeTransport{
		Response: &HTTPResponse{
			StatusCode: 201,
			Headers:    map[string]string{"Content-Type": "application/json"},
			Body:       respBody,
		},
	}

	opts := ClientOptions{
		SessionURL: "https://example.com/.well-known/jmap",
		Username:   username,
		Password:   password,
		Transport:  fake,
	}
	client := NewClient(opts)
	client.uploadURL = uploadURL

	data := []byte("hello world")
	got, err := client.UploadBlob("acct", data, "text/plain")
	if err != nil {
		t.Fatalf("UploadBlob error: %v", err)
	}
	if got.BlobId != upResp.BlobId || got.Size != upResp.Size {
		t.Errorf("UploadBlob returned %+v, want %+v", got, upResp)
	}
	// Verify request URL substitution.
	expectedURL := "https://example.com/upload/acct"
	if fake.LastRequest.URL != expectedURL {
		t.Errorf("UploadBlob URL = %s, want %s", fake.LastRequest.URL, expectedURL)
	}
	if fake.LastRequest.Method != "POST" {
		t.Errorf("UploadBlob method = %s, want POST", fake.LastRequest.Method)
	}
	if ct := fake.LastRequest.Headers["Content-Type"]; ct != "text/plain" {
		t.Errorf("UploadBlob Content-Type = %s, want text/plain", ct)
	}
}

// Test DownloadBlob retrieves raw bytes and content type.
func TestClientCoreDownloadBlob(t *testing.T) {
	const downloadURL = "https://example.com/download/{accountId}/{blobId}/{type}/{name}"
	const username = "eve"
	const password = "pwd"

	fake := &ClientCoreFakeTransport{
		Response: &HTTPResponse{
			StatusCode: 200,
			Headers:    map[string]string{"Content-Type": "application/pdf"},
			Body:       []byte("%PDF-1.4..."),
		},
	}

	opts := ClientOptions{
		SessionURL: "https://example.com/.well-known/jmap",
		Username:   username,
		Password:   password,
		Transport:  fake,
	}
	client := NewClient(opts)
	client.downloadURL = downloadURL

	data, ct, err := client.DownloadBlob("acct", "blob123", "application/pdf", "file.pdf")
	if err != nil {
		t.Fatalf("DownloadBlob error: %v", err)
	}
	if string(data) != "%PDF-1.4..." {
		t.Errorf("DownloadBlob data = %s, want %s", string(data), "%PDF-1.4...")
	}
	if ct != "application/pdf" {
		t.Errorf("DownloadBlob content type = %s, want application/pdf", ct)
	}
	// Verify URL substitution.
	expected := "https://example.com/download/acct/blob123/application%2Fpdf/file.pdf"
	if fake.LastRequest.URL != expected {
		t.Errorf("DownloadBlob URL = %s, want %s", fake.LastRequest.URL, expected)
	}
	if fake.LastRequest.Method != "GET" {
		t.Errorf("DownloadBlob method = %s, want GET", fake.LastRequest.Method)
	}
}

// Regression test: SendRequest must be exported (not the unexported sendRequest it replaced),
// so callers can batch multiple method calls - chained via a ResultReference back-reference
// (RFC 8620 section 3.7) - into a single HTTP round trip instead of one request per call.
func TestClientCoreSendRequestBatchesMultipleCallsWithResultReference(t *testing.T) {
	const apiURL = "https://example.com/jmap"

	respEnv := JmapResponseEnvelope{
		MethodResponses: []Invocation{
			{Name: "Core/echo", Arguments: map[string]interface{}{"hello": "world"}, MethodCallId: "c1"},
			{Name: "Core/echo", Arguments: map[string]interface{}{"foo": "world"}, MethodCallId: "c2"},
		},
	}
	respBody, _ := json.Marshal(respEnv)

	fake := &ClientCoreFakeTransport{
		Response: &HTTPResponse{
			StatusCode: 200,
			Headers:    map[string]string{"Content-Type": "application/json"},
			Body:       respBody,
		},
	}

	opts := ClientOptions{
		SessionURL: "https://example.com/.well-known/jmap",
		Username:   "bob",
		Password:   "pwd",
		Transport:  fake,
	}
	client := NewClient(opts)
	client.apiURL = apiURL

	first := Invocation{Name: "Core/echo", Arguments: map[string]interface{}{"hello": "world"}, MethodCallId: "c1"}
	second := Invocation{
		Name: "Core/echo",
		Arguments: map[string]interface{}{
			"#foo": ResultReference{ResultOf: "c1", Name: "Core/echo", Path: "/hello"},
		},
		MethodCallId: "c2",
	}

	env, err := client.SendRequest([]Invocation{first, second}, []string{"urn:ietf:params:jmap:core"})
	if err != nil {
		t.Fatalf("SendRequest returned error: %v", err)
	}
	if len(env.MethodResponses) != 2 {
		t.Fatalf("Expected 2 method responses, got %d", len(env.MethodResponses))
	}
	if env.MethodResponses[0].Arguments["hello"] != "world" {
		t.Errorf("MethodResponses[0].Arguments[hello] = %v, want world", env.MethodResponses[0].Arguments["hello"])
	}
	if env.MethodResponses[1].Arguments["foo"] != "world" {
		t.Errorf("MethodResponses[1].Arguments[foo] = %v, want world", env.MethodResponses[1].Arguments["foo"])
	}

	var sentEnv JmapRequestEnvelope
	if err := json.Unmarshal(fake.LastRequest.Body, &sentEnv); err != nil {
		t.Fatalf("Failed to unmarshal request body: %v", err)
	}
	if len(sentEnv.MethodCalls) != 2 {
		t.Fatalf("Expected 2 method calls in one request, got %d", len(sentEnv.MethodCalls))
	}
	refArg, ok := sentEnv.MethodCalls[1].Arguments["#foo"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected #foo argument to be an object, got %T", sentEnv.MethodCalls[1].Arguments["#foo"])
	}
	if refArg["resultOf"] != "c1" || refArg["name"] != "Core/echo" || refArg["path"] != "/hello" {
		t.Errorf("#foo ResultReference = %v, want resultOf=c1 name=Core/echo path=/hello", refArg)
	}
}
