package jmap

import (
	"bytes"
	"errors"
	"io"
	"net/http"
)

// HTTPRequest represents a raw HTTP request for the Transport seam.
type HTTPRequest struct {
	Method  string
	URL     string
	Headers map[string]string
	Body    []byte
}

// HTTPResponse represents a raw HTTP response returned by the Transport.
type HTTPResponse struct {
	StatusCode int
	Headers    map[string]string
	Body       []byte
}

// Transport is the seam that abstracts the underlying HTTP transport.
// Implementations must send the given HTTPRequest and return an HTTPResponse
// or an error.
type Transport interface {
	// Send performs the HTTP request described by req and returns the response.
	Send(req *HTTPRequest) (*HTTPResponse, error)
}

// HTTPTransport is the default implementation of Transport that uses
// net/http.Client to perform real network calls.
type HTTPTransport struct {
	client *http.Client
}

// NewHTTPTransport creates a new HTTPTransport using the provided http.Client.
// If client is nil, http.DefaultClient is used.
func NewHTTPTransport(client *http.Client) *HTTPTransport {
	if client == nil {
		client = http.DefaultClient
	}
	return &HTTPTransport{client: client}
}

// Send implements the Transport interface using the standard net/http package.
// It translates the high-level HTTPRequest into an *http.Request, executes it,
// and converts the result into an HTTPResponse.
func (t *HTTPTransport) Send(req *HTTPRequest) (*HTTPResponse, error) {
	if req == nil {
		return nil, &NetworkError{Err: errors.New("nil request")}
	}

	httpReq, err := http.NewRequest(req.Method, req.URL, bytes.NewReader(req.Body))
	if err != nil {
		return nil, &NetworkError{Err: err}
	}
	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}

	resp, err := t.client.Do(httpReq)
	if err != nil {
		return nil, &NetworkError{Err: err}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &NetworkError{Err: err}
	}

	// Convert response headers to a simple map[string]string (first value only).
	hdrs := make(map[string]string, len(resp.Header))
	for k, vv := range resp.Header {
		if len(vv) > 0 {
			hdrs[k] = vv[0]
		}
	}

	return &HTTPResponse{
		StatusCode: resp.StatusCode,
		Headers:    hdrs,
		Body:       body,
	}, nil
}
