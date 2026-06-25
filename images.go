package simplellm

import "context"

// GenerateImage generates one or more images from the given prompt.
func (c *Client) GenerateImage(ctx context.Context, req ImageRequest) (*ImageResponse, error) {
	var out ImageResponse
	if err := c.do(ctx, "POST", "/v1/images/generations", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
