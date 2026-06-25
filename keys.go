package simplellm

import (
	"context"
	"fmt"
)

// ListKeys returns all API keys associated with the authenticated account.
func (c *Client) ListKeys(ctx context.Context) (*APIKeyList, error) {
	var out APIKeyList
	if err := c.do(ctx, "GET", "/v1/keys", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CurrentKey returns details about the API key used to authenticate this request.
func (c *Client) CurrentKey(ctx context.Context) (*APIKeyInfo, error) {
	var out APIKeyInfo
	if err := c.do(ctx, "GET", "/v1/keys/current", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CurrentKeyUsage returns aggregate usage statistics for the current API key.
func (c *Client) CurrentKeyUsage(ctx context.Context) (*APIKeyUsage, error) {
	var out APIKeyUsage
	if err := c.do(ctx, "GET", "/v1/keys/current/usage", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CurrentKeyDailyUsage returns per-day usage statistics for the current API key.
func (c *Client) CurrentKeyDailyUsage(ctx context.Context) (*APIKeyDailyUsageList, error) {
	var out APIKeyDailyUsageList
	if err := c.do(ctx, "GET", "/v1/keys/current/usage/daily", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// KeyUsage returns aggregate usage statistics for the API key with the given ID.
func (c *Client) KeyUsage(ctx context.Context, id string) (*APIKeyUsage, error) {
	var out APIKeyUsage
	if err := c.do(ctx, "GET", fmt.Sprintf("/v1/keys/%s/usage", id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// KeyDailyUsage returns per-day usage statistics for the API key with the given ID.
func (c *Client) KeyDailyUsage(ctx context.Context, id string) (*APIKeyDailyUsageList, error) {
	var out APIKeyDailyUsageList
	if err := c.do(ctx, "GET", fmt.Sprintf("/v1/keys/%s/usage/daily", id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
