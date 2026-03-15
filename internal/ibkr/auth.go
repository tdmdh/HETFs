package ibkr

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
)

// AuthStatus represents the response from /v1/api/iserver/auth/status
type AuthStatus struct {
	Authenticated bool   `json:"authenticated"`
	Competing     bool   `json:"competing"`
	Connected     bool   `json:"connected"`
	Message       string `json:"message"`
	Prompts       []any  `json:"prompts"`
}

// CheckAuthStatus calls the status endpoint to verify the gateway is authenticated and connected.
func (c *Client) CheckAuthStatus(ctx context.Context) (*AuthStatus, error) {
	// Important: ibapi sometimes prefers POST for status according to some docs,
	// but the official spec and gateway trace usually uses GET or POST.
	// We'll use POST as it is safer for IBKR's quirky API.
	resp, err := c.PostEmpty(ctx, "/v1/api/iserver/auth/status")
	if err != nil {
		return nil, fmt.Errorf("auth status request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	var status AuthStatus
	if err := json.Unmarshal(body, &status); err != nil {
		return nil, fmt.Errorf("parse json: %w", err)
	}

	return &status, nil
}

// Tickle ping the gateway to keep the session alive.
func (c *Client) Tickle(ctx context.Context) error {
	resp, err := c.PostEmpty(ctx, "/v1/api/tickle")
	if err != nil {
		return fmt.Errorf("tickle request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		// Log or handle, but tickle might return 401 if session dropped.
		return fmt.Errorf("tickle unexpected status code: %d", resp.StatusCode)
	}
	return nil
}
