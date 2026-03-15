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
	baseURL    string
	httpClient *http.Client
	limiter    *rate.Limiter
}

// NewClient creates a new IBKR Gateway client.
// By default, the Gateway runs on https://localhost:5000 and uses a self-signed cert.
func NewClient(baseURL string) *Client {
	// IBKR limits to 10 requests per second. We use 9 to be safe.
	// We allow a burst of 1 to ensure strict pacing.
	limiter := rate.NewLimiter(rate.Limit(9.0), 1)

	// Gateway uses a self-signed cert by default, so we must disable verification.
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout:   30 * time.Second,
			Transport: tr,
		},
		limiter: limiter,
	}
}

// Do executes an HTTP request, respecting the rate limit and handling 429 backoffs.
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	const maxRetries = 3
	var resp *http.Response
	var err error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Wait for the rate limiter.
		if err := c.limiter.Wait(req.Context()); err != nil {
			return nil, fmt.Errorf("rate limiter wait: %w", err)
		}

		resp, err = c.httpClient.Do(req)
		if err != nil {
			return nil, err
		}

		// If 429 Too Many Requests, back off and retry.
		if resp.StatusCode == http.StatusTooManyRequests {
			resp.Body.Close()
			if attempt == maxRetries {
				break
			}
			backoff := time.Duration(1<<attempt) * time.Second // 1s, 2s, 4s
			select {
			case <-time.After(backoff):
				continue
			case <-req.Context().Done():
				return nil, req.Context().Err()
			}
		}

		// Success or other failure
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

// Post is a convenience method for POST requests with JSON payload.
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

// PostForm is a convenience method for POST requests without a body (or empty body text).
// Sometimes IBKR wants form or empty post.
func (c *Client) PostEmpty(ctx context.Context, endpoint string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+endpoint, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}
