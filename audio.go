package simplellm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"strconv"
)

// Transcribe uploads an audio file and returns the transcription.
// Set req.File to the raw audio bytes and req.Filename to the original filename
// (used by the server for MIME-type detection).
func (c *Client) Transcribe(ctx context.Context, req TranscribeRequest) (*Transcription, error) {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)

	// Required: the audio file
	fw, err := mw.CreateFormFile("file", req.Filename)
	if err != nil {
		return nil, fmt.Errorf("simplellm: create form file: %w", err)
	}
	if _, err := fw.Write(req.File); err != nil {
		return nil, fmt.Errorf("simplellm: write file field: %w", err)
	}

	// Optional string fields — only sent when non-empty
	for field, val := range map[string]string{
		"model":           req.Model,
		"language":        req.Language,
		"prompt":          req.Prompt,
		"response_format": req.ResponseFormat,
	} {
		if val != "" {
			if err := mw.WriteField(field, val); err != nil {
				return nil, fmt.Errorf("simplellm: write field %s: %w", field, err)
			}
		}
	}

	if req.Temperature != nil {
		v := strconv.FormatFloat(*req.Temperature, 'f', -1, 64)
		if err := mw.WriteField("temperature", v); err != nil {
			return nil, fmt.Errorf("simplellm: write field temperature: %w", err)
		}
	}

	if err := mw.Close(); err != nil {
		return nil, fmt.Errorf("simplellm: close multipart writer: %w", err)
	}

	resp, err := c.doRawBody(ctx, "POST", "/v1/audio/transcriptions", mw.FormDataContentType(), &buf)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, parseError(resp)
	}

	var out Transcription
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("simplellm: decode transcription: %w", err)
	}
	return &out, nil
}

// Speech synthesizes the given text and returns the raw audio bytes.
// The audio format depends on req.ResponseFormat (default is mp3).
func (c *Client) Speech(ctx context.Context, req SpeechRequest) ([]byte, error) {
	b, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("simplellm: marshal request: %w", err)
	}

	httpReq, err := c.newRequest(ctx, "POST", "/v1/audio/speech", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "*/*")

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("simplellm: http: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, parseError(resp)
	}

	audio, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("simplellm: read audio: %w", err)
	}
	return audio, nil
}
