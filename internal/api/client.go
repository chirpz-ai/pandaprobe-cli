// Package api is the HTTP transport layer for the PandaProbe backend. All
// network access goes through Client; command code never builds requests
// directly.
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/uuid"

	"github.com/chirpz-ai/pandaprobe-cli/internal/config"
	"github.com/chirpz-ai/pandaprobe-cli/internal/output"
	"github.com/chirpz-ai/pandaprobe-cli/internal/version"
)

// Client talks to the PandaProbe REST API.
type Client struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
	project    string
	bearer     string // reserved for a future `auth login` flow
	userAgent  string
	debug      bool
	debugOut   io.Writer
	newReqID   func() string
}

// Option customizes a Client.
type Option func(*Client)

// WithHTTPClient overrides the underlying *http.Client (used in tests).
func WithHTTPClient(h *http.Client) Option { return func(c *Client) { c.httpClient = h } }

// WithRequestID overrides the request-ID generator (used in tests for
// determinism).
func WithRequestID(fn func() string) Option { return func(c *Client) { c.newReqID = fn } }

// WithDebugOutput overrides where debug logs are written (defaults to stderr at
// construction).
func WithDebugOutput(w io.Writer) Option { return func(c *Client) { c.debugOut = w } }

// WithBearerToken configures bearer-token auth instead of the API key. Reserved
// for future control-plane commands; when set, the Authorization header is sent
// and X-API-Key is omitted.
func WithBearerToken(tok string) Option { return func(c *Client) { c.bearer = tok } }

// New builds a Client from resolved config.
func New(cfg *config.Config, debugOut io.Writer, opts ...Option) *Client {
	c := &Client{
		httpClient: &http.Client{Timeout: cfg.Timeout},
		baseURL:    strings.TrimRight(cfg.Endpoint, "/"),
		apiKey:     cfg.APIKey,
		project:    cfg.ProjectName,
		userAgent:  version.UserAgent(),
		debug:      cfg.Debug,
		debugOut:   debugOut,
		newReqID:   uuid.NewString,
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// Do performs a request and returns the raw response body on success, or an
// *APIError on a non-2xx response. Network/transport failures return a plain
// error (mapped to a general exit code).
func (c *Client) Do(ctx context.Context, method, path string, query url.Values, body any) (json.RawMessage, error) {
	u := c.baseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	var bodyReader io.Reader
	var bodyBytes []byte
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("encode request body: %w", err)
		}
		bodyBytes = b
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, u, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	reqID := c.newReqID()
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("X-Request-ID", reqID)
	if c.bearer != "" {
		req.Header.Set("Authorization", "Bearer "+c.bearer)
	} else {
		req.Header.Set("X-API-Key", c.apiKey)
	}
	req.Header.Set("X-Project-Name", c.project)
	if bodyBytes != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	c.logRequest(req, bodyBytes, reqID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request to %s %s failed: %w", method, path, err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	c.logResponse(resp.StatusCode, respBody, reqID)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, Parse(resp.StatusCode, respBody, reqID)
	}
	return respBody, nil
}

func (c *Client) logRequest(req *http.Request, body []byte, reqID string) {
	if !c.debug || c.debugOut == nil {
		return
	}
	_, _ = fmt.Fprintf(c.debugOut, "[debug] req %s %s (request_id=%s)\n", req.Method, req.URL.String(), reqID)
	for k, vals := range req.Header {
		v := strings.Join(vals, ",")
		if strings.EqualFold(k, "X-API-Key") || strings.EqualFold(k, "Authorization") {
			v = output.MaskSecret(v)
		}
		_, _ = fmt.Fprintf(c.debugOut, "[debug]   %s: %s\n", k, v)
	}
	if len(body) > 0 {
		_, _ = fmt.Fprintf(c.debugOut, "[debug]   body: %s\n", string(body))
	}
}

func (c *Client) logResponse(status int, body []byte, reqID string) {
	if !c.debug || c.debugOut == nil {
		return
	}
	_, _ = fmt.Fprintf(c.debugOut, "[debug] resp %d (request_id=%s) %d bytes\n", status, reqID, len(body))
}

// doDecode performs a request and decodes the JSON response into T.
func doDecode[T any](ctx context.Context, c *Client, method, path string, query url.Values, body any) (*T, error) {
	raw, err := c.Do(ctx, method, path, query, body)
	if err != nil {
		return nil, err
	}
	var out T
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &out, nil
}
