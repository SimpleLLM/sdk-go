package simplellm

import "context"

// Models returns the list of models available to the authenticated account.
func (c *Client) Models(ctx context.Context) (*ModelList, error) {
	var out ModelList
	if err := c.do(ctx, "GET", "/v1/models", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
