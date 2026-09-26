package jmap

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"testing"
)

// TransportFakeTransport is a fake http.RoundTripper used to capture the
// request sent by HTTPTransport and to return a canned response or error.
type TransportFakeTransport struct {
	// Resp is the response that RoundTrip will return.
	Resp *http.Response
	// Err is the error that RoundTrip will return.
	Err error
	// CapturedReq stores the request received by RoundTrip.
	CapturedReq *http.Request
}

// RoundTrip implements http.RoundTripper.
func (f *TransportFakeTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	f.CapturedReq = req
	if f.Err != nil {
		return nil, f.Err
	}
	return f.Resp, nil
}

// TestHTTPTransportSendSuccess verifies that HTTPTransport correctly translates
// an HTTPRequest into an *http.Request, sends it via the underlying client, and
// converts the *http.Response back into an HTTPResponse.
func TestHTTPTransportSendSuccess(t *testing.T) {
	// Prepare a fake response.
	bodyContent := []byte(`{"ok":true}`)
	fakeResp := &http.Response{
		StatusCode: 200,
		Header: http.Header{
			"Content-Type": []string{"application/json"},
			"X-Custom":     []string{"value"},
		},
		Body: io.NopCloser(bytes.NewReader(bodyContent)),
	}
	fakeTransport := &TransportFakeTransport{Resp: fakeResp}
	client := &http.Client{Transport: fakeTransport}
	transport := NewHTTPTransport(client)

	// Build the request to be sent.
	req := &HTTPRequest{
		Method: "POST",
		URL:    "https://example.com/jmap",
		Headers: map[string]string{
			"Accept":       "application/json",
			"Authorization": "Bearer token",
		},
		Body: []byte(`{"methodCalls":[]}`),
	}

	// Execute.
	resp, err := transport.Send(req)
	if err != nil {
		t.Fatalf("unexpected error from Send: %v", err)
	}
	if resp == nil {
		t.Fatalf("expected non-nil response")
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
	if ct := resp.Headers["Content-Type"]; ct != "application/json" {
		t.Errorf("expected Content-Type header 'application/json', got %q", ct)
	}
	if custom := resp.Headers["X-Custom"]; custom != "value" {
		t.Errorf("expected X-Custom header 'value', got %q", custom)
	}
	if !bytes.Equal(resp.Body, bodyContent) {
		t.Errorf("response body mismatch: got %s, want %s", string(resp.Body), string(bodyContent))
	}

	// Verify the request that was actually sent.
	captured := fakeTransport.CapturedReq
	if captured == nil {
		t.Fatalf("no request captured by fake transport")
	}
	if captured.Method != req.Method {
		t.Errorf("method mismatch: got %s, want %s", captured.Method, req.Method)
	}
	if captured.URL.String() != req.URL {
		t.Errorf("URL mismatch: got %s, want %s", captured.URL.String(), req.URL)
	}
	for k, v := range req.Headers {
		if got := captured.Header.Get(k); got != v {
			t.Errorf("header %s mismatch: got %s, want %s", k, got, v)
		}
	}
	sentBody, err := io.ReadAll(captured.Body)
	if err != nil {
		t.Fatalf("reading captured request body: %v", err)
	}
	if !bytes.Equal(sentBody, req.Body) {
		t.Errorf("request body mismatch: got %s, want %s", string(sentBody), string(req.Body))
	}
}

// TestHTTPTransportSendNilRequest ensures that sending a nil HTTPRequest
// results in a NetworkError.
func TestHTTPTransportSendNilRequest(t *testing.T) {
	transport := NewHTTPTransport(nil) // nil client defaults to http.DefaultClient
	_, err := transport.Send(nil)
	if err == nil {
		t.Fatalf("expected error for nil request, got nil")
	}
	var netErr *NetworkError
	if !errors.As(err, &netErr) {
		t.Fatalf("expected error of type *NetworkError, got %T", err)
	}
}

// TestHTTPTransportSendTransportError verifies that an error returned by the
// underlying http.Client is wrapped in a NetworkError.
func TestHTTPTransportSendTransportError(t *testing.T) {
	expectedErr := errors.New("transport failure")
	fakeTransport := &TransportFakeTransport{Err: expectedErr}
	client := &http.Client{Transport: fakeTransport}
	transport := NewHTTPTransport(client)

	req := &HTTPRequest{
		Method:  "GET",
		URL:     "https://example.com/fail",
		Headers: map[string]string{},
		Body:    nil,
	}
	_, err := transport.Send(req)
	if err == nil {
		t.Fatalf("expected error from Send, got nil")
	}
	var netErr *NetworkError
	if !errors.As(err, &netErr) {
		t.Fatalf("expected *NetworkError, got %T", err)
	}
	// http.Client.Do wraps RoundTripper errors in a *url.Error before returning them;
	// errors.Is follows that wrapper's Unwrap() chain back to the exact original error.
	if netErr.Err == nil || !errors.Is(netErr.Err, expectedErr) {
		t.Errorf("NetworkError.Err mismatch: got %v, want (wrapping) %v", netErr.Err, expectedErr)
	}
}
