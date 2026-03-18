package ibkr

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

// Client handles communication with the IBKR Client Portal Gateway.
type Client struct {
	baseURL       string
	httpClient    *http.Client
	limiter       *rate.Limiter
	sessionCookie string // optional: injected as Cookie header on every request
}

// NewClient creates a new IBKR Gateway client without a session cookie.
func NewClient(baseURL string) *Client {
	return NewClientWithCookie(baseURL, "")
}

// NewClientWithCookie creates a new IBKR Gateway client, injecting a browser session cookie
// on every request. The cookie should be the full raw value from browser DevTools, e.g.:
//
//	"x-sess-uuid=0.12345678.1234567890.abcdef01"
//
// To obtain it: log in at https://127.0.0.1:5000, open DevTools (F12), go to
// Application → Cookies → https://127.0.0.1, copy the x-sess-uuid value.
// Store in .env as SESSION_COOKIE=x-sess-uuid=... (.env is gitignored).
func NewClientWithCookie(baseURL, sessionCookie string) *Client {
	// IBKR limits to 10 requests per second. We use 9 to be safe.
	limiter := rate.NewLimiter(rate.Limit(9.0), 1)

	// Gateway uses a self-signed cert, so we disable verification.
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	return &Client{
		baseURL:       baseURL,
		sessionCookie: sessionCookie,
		httpClient: &http.Client{
			Timeout:   30 * time.Second,
			Transport: tr,
		},
		limiter: limiter,
	}
}

// Do executes an HTTP request, respecting the rate limit and handling 429 backoffs.
// If a session cookie is configured it is injected into every outgoing request.
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	if c.sessionCookie != "" {
		req.Header.Set("Cookie", c.sessionCookie)
	}

	const maxRetries = 3
	var resp *http.Response
	var err error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if err := c.limiter.Wait(req.Context()); err != nil {
			return nil, fmt.Errorf("rate limiter wait: %w", err)
		}

		resp, err = c.httpClient.Do(req)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			resp.Body.Close()
			if attempt == maxRetries {
				break
			}
			backoff := time.Duration(1<<attempt) * time.Second
			select {
			case <-time.After(backoff):
				continue
			case <-req.Context().Done():
				return nil, req.Context().Err()
			}
		}

		break
	}

	return resp, err
}

// Get is a convenience method for GET requests.
func (c *Client) Get(ctx context.Context, endpoint string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+endpoint, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

// Post is a convenience method for POST requests with a JSON payload.
func (c *Client) Post(ctx context.Context, endpoint string, body any) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal body: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+endpoint, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return c.Do(req)
}

// PostEmpty is a convenience method for POST requests without a body.
func (c *Client) PostEmpty(ctx context.Context, endpoint string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+endpoint, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}
