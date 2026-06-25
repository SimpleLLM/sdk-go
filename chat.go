package simplellm

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// Chat sends a non-streaming chat completion request and returns the result.
func (c *Client) Chat(ctx context.Context, req ChatRequest) (*ChatCompletion, error) {
	req.Stream = false
	var out ChatCompletion
	if err := c.do(ctx, "POST", "/v1/chat/completions", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ChatStream sends a streaming chat completion request and returns a [*ChatStream]
// that the caller iterates with [ChatStream.Recv].
// The caller must call [ChatStream.Close] when done, even on error.
func (c *Client) ChatStream(ctx context.Context, req ChatRequest) (*ChatStream, error) {
	req.Stream = true
	resp, err := c.doStream(ctx, "POST", "/v1/chat/completions", req)
	if err != nil {
		return nil, err
	}
	return &ChatStream{
		body:    resp.Body,
		scanner: bufio.NewScanner(resp.Body),
	}, nil
}

// ChatStream is an iterator over server-sent events from a streaming chat
// completion. Call [ChatStream.Recv] to get the next chunk.
// Call [ChatStream.Close] when done, even if Recv returned an error.
type ChatStream struct {
	body    io.ReadCloser
	scanner *bufio.Scanner
}

// Recv returns the next chunk from the stream.
// Returns (nil, [io.EOF]) when the stream ends normally (data: [DONE]).
// Returns a non-nil error on parse errors or network errors.
func (s *ChatStream) Recv() (*ChatCompletionChunk, error) {
	for s.scanner.Scan() {
		line := s.scanner.Text()

		// SSE: only handle "data: " lines; skip comments and event: lines.
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		payload := strings.TrimPrefix(line, "data: ")
		if payload == "[DONE]" {
			return nil, io.EOF
		}

		var chunk ChatCompletionChunk
		if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
			return nil, fmt.Errorf("simplellm: parse chunk: %w", err)
		}
		return &chunk, nil
	}

	if err := s.scanner.Err(); err != nil {
		return nil, fmt.Errorf("simplellm: read stream: %w", err)
	}
	return nil, io.EOF
}

// Close closes the underlying HTTP response body.
func (s *ChatStream) Close() error {
	return s.body.Close()
}
