package jmap

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sync/atomic"
)

// ClientOptions configures a JmapClient.
type ClientOptions struct {
	// SessionURL is the well-known JMAP session endpoint.
	SessionURL string
	// Username and Password are used for HTTP Basic authentication.
	Username string
	Password string
	// BearerToken, if non-empty, is used for OAuth 2.0 Bearer authentication
	// (RFC 6750) instead of HTTP Basic authentication. When set, it takes
	// precedence over Username/Password.
	BearerToken string
	// Transport, if non-nil, overrides the default HTTPTransport.
	Transport Transport
}

// UploadResponse models the JSON response from the blob upload endpoint.
type UploadResponse struct {
	AccountId string `json:"accountId"`
	BlobId    string `json:"blobId"`
	Type      string `json:"type"`
	Size      int    `json:"size"`
}

// JmapClient provides methods for the core JMAP module.
type JmapClient struct {
	opts           ClientOptions
	transport      Transport
	session        *Session
	apiURL         string
	uploadURL      string
	downloadURL    string
	eventSourceURL string
}

// NewClient creates a new JmapClient using the supplied options.
// If opts.Transport is nil, a default HTTPTransport is used.
func NewClient(opts ClientOptions) *JmapClient {
	tr := opts.Transport
	if tr == nil {
		tr = NewHTTPTransport(nil)
	}
	return &JmapClient{
		opts:      opts,
		transport: tr,
	}
}

// Connect fetches the Session resource, resolves relative URLs, and stores the
// result. Unlike a naive cache, it always re-fetches on every call - the session
// state can change server-side (RFC 8620 section 2's `state` field exists
// precisely because it does), and every other language target in this project
// (Python, TypeScript, Rust) re-fetches on each Connect() call too; a Go-only
// cache-and-short-circuit was an inconsistent idempotency contract for a
// same-named method across language bindings.
func (c *JmapClient) Connect() (*Session, error) {
	req := &HTTPRequest{
		Method: "GET",
		URL:    c.opts.SessionURL,
		Headers: map[string]string{
			"Accept":        "application/json",
			"Authorization": c.authHeader(),
		},
	}
	resp, err := c.transport.Send(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &NetworkError{Err: errors.New(describeHTTPError("failed to fetch session", resp.StatusCode, resp.Body))}
	}

	var sess Session
	if err := json.Unmarshal(resp.Body, &sess); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session JSON: %w", err)
	}

	base, err := url.Parse(c.opts.SessionURL)
	if err != nil {
		return nil, fmt.Errorf("invalid session URL %q: %w", c.opts.SessionURL, err)
	}
	sess.origin = base.Scheme + "://" + base.Host
	sess.ApiUrl = resolveURL(base, sess.ApiUrl)
	sess.DownloadUrl = resolveURL(base, sess.DownloadUrl)
	sess.UploadUrl = resolveURL(base, sess.UploadUrl)
	sess.EventSourceUrl = resolveURL(base, sess.EventSourceUrl)

	c.apiURL = sess.ApiUrl
	c.uploadURL = sess.UploadUrl
	c.downloadURL = sess.DownloadUrl
	c.eventSourceURL = sess.EventSourceUrl
	c.session = &sess
	return c.session, nil
}

// Echo calls the Core/echo method and returns the echoed arguments.
func (c *JmapClient) Echo(args map[string]interface{}) (map[string]interface{}, error) {
	call := Invocation{
		Name:         "Core/echo",
		Arguments:    args,
		MethodCallId: nextCallID(),
	}
	env, err := c.SendRequest([]Invocation{call}, []string{"urn:ietf:params:jmap:core"})
	if err != nil {
		return nil, err
	}
	if len(env.MethodResponses) == 0 {
		return nil, fmt.Errorf("no methodResponses in echo reply")
	}
	return env.MethodResponses[0].Arguments, nil
}

// UploadBlob uploads raw data and returns the server-generated blob information.
func (c *JmapClient) UploadBlob(accountId string, data []byte, contentType string) (*UploadResponse, error) {
	if c.uploadURL == "" {
		return nil, fmt.Errorf("upload URL not set; call Connect first")
	}
	target := strings.ReplaceAll(c.uploadURL, "{accountId}", url.PathEscape(accountId))

	req := &HTTPRequest{
		Method: "POST",
		URL:    target,
		Headers: map[string]string{
			"Authorization": c.authHeader(),
			"Content-Type":  contentType,
		},
		Body: data,
	}
	resp, err := c.transport.Send(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &NetworkError{Err: errors.New(describeHTTPError("blob upload failed", resp.StatusCode, resp.Body))}
	}
	var up UploadResponse
	if err := json.Unmarshal(resp.Body, &up); err != nil {
		return nil, fmt.Errorf("failed to unmarshal upload response: %w", err)
	}
	return &up, nil
}

// DownloadBlob fetches a previously uploaded or server-generated blob and
// returns its raw bytes and Content-Type.
func (c *JmapClient) DownloadBlob(accountId, blobId, typ, name string) ([]byte, string, error) {
	if c.downloadURL == "" {
		return nil, "", fmt.Errorf("download URL not set; call Connect first")
	}
	target := c.downloadURL
	target = strings.ReplaceAll(target, "{accountId}", url.PathEscape(accountId))
	target = strings.ReplaceAll(target, "{blobId}", url.PathEscape(blobId))
	target = strings.ReplaceAll(target, "{type}", url.PathEscape(typ))
	target = strings.ReplaceAll(target, "{name}", url.PathEscape(name))

	req := &HTTPRequest{
		Method: "GET",
		URL:    target,
		Headers: map[string]string{
			"Authorization": c.authHeader(),
		},
	}
	resp, err := c.transport.Send(req)
	if err != nil {
		return nil, "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, "", &NetworkError{Err: errors.New(describeHTTPError("blob download failed", resp.StatusCode, resp.Body))}
	}
	return resp.Body, resp.Headers["Content-Type"], nil
}

// SendRequest builds a JMAP request envelope, sends it via the transport,
// parses the response envelope, and returns a *ProtocolError if any method
// response is a top-level "error" Invocation.
//
// Exported so callers can batch multiple method calls into a single HTTP round
// trip (RFC 8620 section 3.7): pass more than one entry in calls, and use a
// ResultReference as a later call's argument value (under a key prefixed with
// "#", e.g. map[string]interface{}{"#ids": ResultReference{ResultOf: "c1", Name:
// "Email/query", Path: "/ids"}}) to reference an earlier call's response
// without a second request.
func (c *JmapClient) SendRequest(calls []Invocation, using []string) (*JmapResponseEnvelope, error) {
	if c.apiURL == "" {
		return nil, fmt.Errorf("API URL not set; call Connect first")
	}
	env := JmapRequestEnvelope{
		Using:       using,
		MethodCalls: calls,
	}
	body, err := json.Marshal(env)
	if err != nil {
		return nil, err
	}
	req := &HTTPRequest{
		Method: "POST",
		URL:    c.apiURL,
		Headers: map[string]string{
			"Authorization": c.authHeader(),
			"Content-Type":  "application/json",
			"Accept":        "application/json",
		},
		Body: body,
	}
	resp, err := c.transport.Send(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &NetworkError{Err: errors.New(describeHTTPError("JMAP request failed", resp.StatusCode, resp.Body))}
	}
	var respEnv JmapResponseEnvelope
	if err := json.Unmarshal(resp.Body, &respEnv); err != nil {
		return nil, err
	}
	for _, mr := range respEnv.MethodResponses {
		if mr.Name == "error" {
			typ, _ := mr.Arguments["type"].(string)
			desc, hasDesc := mr.Arguments["description"].(string)
			pe := &ProtocolError{Type: typ}
			if hasDesc {
				pe.Description = &desc
			}
			return nil, pe
		}
	}
	return &respEnv, nil
}

// basicAuth creates an HTTP Basic authentication header value.
func basicAuth(user, pass string) string {
	cred := user + ":" + pass
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(cred))
}

// authHeader returns the value for the Authorization header. If a
// BearerToken is configured, it is used (RFC 6750); otherwise it falls back
// to HTTP Basic authentication using Username/Password.
func (c *JmapClient) authHeader() string {
	if c.opts.BearerToken != "" {
		return "Bearer " + c.opts.BearerToken
	}
	return basicAuth(c.opts.Username, c.opts.Password)
}

const maxErrorBodyExcerpt = 2000

// describeHTTPError builds a non-2xx error message that includes an excerpt of the
// response body. A real JMAP server's error response commonly carries an RFC 8620 section
// 3.6.1 "problem details" JSON object (type/title/detail) explaining exactly what went
// wrong - discarding the body entirely throws that information away, leaving only the bare
// status code to debug from.
func describeHTTPError(prefix string, statusCode int, body []byte) string {
	text := strings.TrimSpace(string(body))
	if text == "" {
		return fmt.Sprintf("%s: unexpected HTTP status %d", prefix, statusCode)
	}
	if len(text) > maxErrorBodyExcerpt {
		text = text[:maxErrorBodyExcerpt] + "..."
	}
	return fmt.Sprintf("%s: unexpected HTTP status %d: %s", prefix, statusCode, text)
}

// resolveURL resolves raw (which may be relative) against base and returns an absolute URL
// string. JMAP session URLs (uploadUrl/downloadUrl) are URI TEMPLATES containing literal
// "{" / "}" placeholder characters, which are not valid per strict RFC 3986 syntax - a full
// url.Parse + ResolveReference round-trip percent-encodes them (producing "%7BaccountId%7D"),
// corrupting the template. Handle the two shapes RFC 8620 actually specifies (already-absolute,
// and root-relative "/...") with plain string logic, which preserves the template characters
// verbatim; fall back to the general resolver only for the rare non-root-relative case.
func resolveURL(base *url.URL, raw string) string {
	if raw == "" {
		return raw
	}
	if strings.Contains(raw, "://") {
		return raw
	}
	if strings.HasPrefix(raw, "/") {
		return base.Scheme + "://" + base.Host + raw
	}
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	return base.ResolveReference(u).String()
}

// callCounter backs nextCallID; atomic so it is safe for concurrent use.
var callCounter int64

// nextCallID produces a simple unique method-call identifier.
func nextCallID() string {
	return "c" + strconv.FormatInt(atomic.AddInt64(&callCounter, 1), 10)
}
