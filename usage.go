package simplellm

import "context"

// Usage returns account-level balance and usage statistics.
func (c *Client) Usage(ctx context.Context) (*AccountUsage, error) {
	var out AccountUsage
	if err := c.do(ctx, "GET", "/v1/usage", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
