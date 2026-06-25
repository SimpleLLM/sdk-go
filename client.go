package simplellm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	defaultBaseURL = "https://api.simplellm.eu"
	defaultTimeout = 120 * time.Second
)

// Client is the SimpleLLM API client.
// Create one with [New].
type Client struct {
	apiKey  string
	baseURL string
	http    *http.Client
}

// Option is a functional option for [New].
type Option func(*Client)

// WithAPIKey sets the API key, overriding the SIMPLELLM_API_KEY env var.
func WithAPIKey(key string) Option {
	return func(c *Client) { c.apiKey = key }
}

// WithBaseURL sets the base URL, overriding the SIMPLELLM_BASE_URL env var
// and the default (https://api.simplellm.eu). Trailing slashes are stripped.
func WithBaseURL(u string) Option {
	return func(c *Client) { c.baseURL = strings.TrimRight(u, "/") }
}

// WithHTTPClient sets a custom *http.Client.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) { c.http = hc }
}

// WithTimeout sets the HTTP timeout, overriding the default (120s).
func WithTimeout(d time.Duration) Option {
	return func(c *Client) {
		if c.http == nil {
			c.http = &http.Client{}
		}
		c.http.Timeout = d
	}
}

// New creates a new Client. Options are applied in order; env vars are read
// before options so options can override them.
//
// Returns an error if no API key is available after applying all options.
func New(opts ...Option) (*Client, error) {
	baseURL := defaultBaseURL
	if v := os.Getenv("SIMPLELLM_BASE_URL"); v != "" {
		baseURL = strings.TrimRight(v, "/")
	}

	c := &Client{
		apiKey:  os.Getenv("SIMPLELLM_API_KEY"),
		baseURL: baseURL,
		http:    &http.Client{Timeout: defaultTimeout},
	}

	for _, o := range opts {
		o(c)
	}

	if c.apiKey == "" {
		return nil, fmt.Errorf("simplellm: no API key provided; set SIMPLELLM_API_KEY or use WithAPIKey")
	}

	return c, nil
}

// newRequest builds an *http.Request with Authorization and Accept headers set.
func (c *Client) newRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return nil, fmt.Errorf("simplellm: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Accept", "application/json")
	return req, nil
}

// do performs a JSON request and decodes the JSON response into out.
// body may be nil (for GET) or any value that will be JSON-encoded.
// out may be nil if the response body should be discarded.
func (c *Client) do(ctx context.Context, method, path string, body, out any) error {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("simplellm: marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := c.newRequest(ctx, method, path, bodyReader)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("simplellm: http: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return parseError(resp)
	}

	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return fmt.Errorf("simplellm: decode response: %w", err)
		}
	}
	return nil
}

// doStream performs a JSON request and returns the raw *http.Response for
// streaming (SSE). The caller must close resp.Body when done.
// Returns an error (including API errors) without a response to close on failure.
func (c *Client) doStream(ctx context.Context, method, path string, body any) (*http.Response, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("simplellm: marshal request: %w", err)
	}

	req, err := c.newRequest(ctx, method, path, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("simplellm: http: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		return nil, parseError(resp)
	}
	return resp, nil
}

// doRawBody performs a request with a pre-built body and explicit Content-Type.
// Used for multipart uploads. Returns the raw response; the caller must close
// resp.Body.
func (c *Client) doRawBody(ctx context.Context, method, path, contentType string, body io.Reader) (*http.Response, error) {
	req, err := c.newRequest(ctx, method, path, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("simplellm: http: %w", err)
	}
	return resp, nil
}
