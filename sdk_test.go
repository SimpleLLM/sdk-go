package simplellm_test

import (
	"context"
	"encoding/json"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/SimpleLLM/sdk-go"
)

// newTestClient creates a Client pointed at the given httptest.Server.
func newTestClient(t *testing.T, srv *httptest.Server) *simplellm.Client {
	t.Helper()
	c, err := simplellm.New(
		simplellm.WithAPIKey("test-key"),
		simplellm.WithBaseURL(srv.URL),
	)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return c
}

// ── Chat non-stream ───────────────────────────────────────────────────────────

func TestChat(t *testing.T) {
	want := simplellm.ChatCompletion{
		ID:      "chatcmpl-1",
		Object:  "chat.completion",
		Created: 1700000000,
		Model:   "test-model",
		Choices: []simplellm.ChatChoice{
			{
				Index:   0,
				Message: simplellm.Message{Role: "assistant", Content: simplellm.Ptr("Hello!")},
			},
		},
		Usage: &simplellm.Usage{PromptTokens: 5, CompletionTokens: 3, TotalTokens: 8},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/v1/chat/completions" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("bad auth header: %s", r.Header.Get("Authorization"))
		}

		var req simplellm.ChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode request body: %v", err)
		}
		if req.Stream {
			t.Error("expected stream=false for Chat()")
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(want); err != nil {
			t.Errorf("encode response: %v", err)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.Chat(context.Background(), simplellm.ChatRequest{
		Model:    "test-model",
		Messages: []simplellm.Message{{Role: "user", Content: simplellm.Ptr("Hi")}},
	})
	if err != nil {
		t.Fatalf("Chat: %v", err)
	}

	if got.ID != want.ID {
		t.Errorf("ID: got %q, want %q", got.ID, want.ID)
	}
	if len(got.Choices) != 1 {
		t.Fatalf("choices len: got %d, want 1", len(got.Choices))
	}
	if got.Choices[0].Message.Content == nil || *got.Choices[0].Message.Content != "Hello!" {
		t.Errorf("content: got %v", got.Choices[0].Message.Content)
	}
	if got.Usage == nil || got.Usage.TotalTokens != 8 {
		t.Errorf("usage: %+v", got.Usage)
	}
}

// ── Chat stream ───────────────────────────────────────────────────────────────

func TestChatStream(t *testing.T) {
	chunks := []string{
		`data: {"id":"chatcmpl-2","object":"chat.completion.chunk","created":1700000001,"model":"test-model","choices":[{"index":0,"delta":{"role":"assistant","content":"He"},"finish_reason":null}]}`,
		`data: {"id":"chatcmpl-2","object":"chat.completion.chunk","created":1700000001,"model":"test-model","choices":[{"index":0,"delta":{"content":"llo"},"finish_reason":null}]}`,
		`data: [DONE]`,
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		flusher, ok := w.(http.Flusher)
		for _, chunk := range chunks {
			if _, err := io.WriteString(w, chunk+"\n\n"); err != nil {
				return
			}
			if ok {
				flusher.Flush()
			}
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	stream, err := c.ChatStream(context.Background(), simplellm.ChatRequest{
		Model:    "test-model",
		Messages: []simplellm.Message{{Role: "user", Content: simplellm.Ptr("Hi")}},
	})
	if err != nil {
		t.Fatalf("ChatStream: %v", err)
	}
	defer stream.Close()

	var collected []string
	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("Recv: %v", err)
		}
		if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != nil {
			collected = append(collected, *chunk.Choices[0].Delta.Content)
		}
	}

	joined := strings.Join(collected, "")
	if joined != "Hello" {
		t.Errorf("stream content: got %q, want %q", joined, "Hello")
	}
}

// ── Models ────────────────────────────────────────────────────────────────────

func TestModels(t *testing.T) {
	want := simplellm.ModelList{
		Object: "list",
		Data: []simplellm.Model{
			{ID: "model-1", Object: "model", Created: 1700000000, OwnedBy: "simplellm"},
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/v1/models" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(want); err != nil {
			t.Errorf("encode response: %v", err)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.Models(context.Background())
	if err != nil {
		t.Fatalf("Models: %v", err)
	}
	if len(got.Data) != 1 || got.Data[0].ID != "model-1" {
		t.Errorf("Models: got %+v", got)
	}
}

// ── Error handling ────────────────────────────────────────────────────────────

func TestErrorShape1(t *testing.T) {
	// {"error":{"message":"...","code":"..."}}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(401)
		if err := json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{"message": "invalid api key", "code": "unauthorized"},
		}); err != nil {
			t.Errorf("encode response: %v", err)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.Models(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	apiErr, ok := err.(*simplellm.Error)
	if !ok {
		t.Fatalf("expected *simplellm.Error, got %T: %v", err, err)
	}
	if apiErr.Status != 401 {
		t.Errorf("Status: got %d, want 401", apiErr.Status)
	}
	if apiErr.Message != "invalid api key" {
		t.Errorf("Message: got %q", apiErr.Message)
	}
	if apiErr.Code != "unauthorized" {
		t.Errorf("Code: got %q", apiErr.Code)
	}
}

func TestErrorShape2(t *testing.T) {
	// {"message":"...","code":"..."}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(429)
		if err := json.NewEncoder(w).Encode(map[string]any{
			"message": "rate limit exceeded",
			"code":    "rate_limited",
		}); err != nil {
			t.Errorf("encode response: %v", err)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.Models(context.Background())
	apiErr, ok := err.(*simplellm.Error)
	if !ok {
		t.Fatalf("expected *simplellm.Error, got %T", err)
	}
	if apiErr.Status != 429 {
		t.Errorf("Status: got %d, want 429", apiErr.Status)
	}
	if apiErr.Message != "rate limit exceeded" {
		t.Errorf("Message: got %q", apiErr.Message)
	}
	if apiErr.Code != "rate_limited" {
		t.Errorf("Code: got %q", apiErr.Code)
	}
}

// ── Multipart transcription ───────────────────────────────────────────────────

func TestTranscribeMultipart(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/v1/audio/transcriptions" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}

		ct := r.Header.Get("Content-Type")
		mediaType, params, err := mime.ParseMediaType(ct)
		if err != nil || mediaType != "multipart/form-data" {
			t.Errorf("expected multipart/form-data, got %q", ct)
			http.Error(w, "bad content type", http.StatusBadRequest)
			return
		}

		mr := multipart.NewReader(r.Body, params["boundary"])
		foundFile := false
		foundModel := false

		for {
			part, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Errorf("multipart parse: %v", err)
				break
			}
			switch part.FormName() {
			case "file":
				foundFile = true
				data, _ := io.ReadAll(part)
				if string(data) != "fakeaudio" {
					t.Errorf("file content: got %q, want %q", data, "fakeaudio")
				}
			case "model":
				foundModel = true
				data, _ := io.ReadAll(part)
				if string(data) != "whisper-large-v3" {
					t.Errorf("model field: got %q, want %q", data, "whisper-large-v3")
				}
			}
		}

		if !foundFile {
			t.Error("missing 'file' field in multipart body")
		}
		if !foundModel {
			t.Error("missing 'model' field in multipart body")
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(simplellm.Transcription{Text: "hello world"}); err != nil {
			t.Errorf("encode response: %v", err)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.Transcribe(context.Background(), simplellm.TranscribeRequest{
		File:     []byte("fakeaudio"),
		Filename: "audio.mp3",
		Model:    "whisper-large-v3",
	})
	if err != nil {
		t.Fatalf("Transcribe: %v", err)
	}
	if got.Text != "hello world" {
		t.Errorf("Text: got %q, want %q", got.Text, "hello world")
	}
}

// ── Constructor ───────────────────────────────────────────────────────────────

func TestNewNoAPIKey(t *testing.T) {
	t.Setenv("SIMPLELLM_API_KEY", "")
	_, err := simplellm.New()
	if err == nil {
		t.Fatal("expected error when no API key is set")
	}
}

func TestNewWithAPIKey(t *testing.T) {
	c, err := simplellm.New(simplellm.WithAPIKey("sk-test"))
	if err != nil {
		t.Fatalf("New with explicit key: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil client")
	}
}
