package simplellm

import "encoding/json"

// ── Chat ──────────────────────────────────────────────────────────────────────

// Message represents a single message in a chat conversation.
type Message struct {
	// Role is one of "system", "user", "assistant", or "tool".
	Role string `json:"role"`
	// Content is the message text. May be nil for tool-call assistant turns.
	Content *string `json:"content"`
	// Name is an optional participant name.
	Name string `json:"name,omitempty"`
	// ToolCalls are tool invocations in an assistant message.
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
	// ToolCallID ties a tool result back to a ToolCall.
	ToolCallID string `json:"tool_call_id,omitempty"`
}

// ToolCall is a function call issued by the model.
type ToolCall struct {
	// ID is the call identifier.
	ID string `json:"id"`
	// Type is always "function".
	Type string `json:"type"`
	// Function holds the name and JSON-encoded arguments.
	Function ToolCallFunction `json:"function"`
}

// ToolCallFunction holds the name and serialised arguments of a tool call.
type ToolCallFunction struct {
	// Name is the function name.
	Name string `json:"name"`
	// Arguments is a JSON-encoded string of the function arguments.
	Arguments string `json:"arguments"`
}

// Tool describes a function the model may call.
type Tool struct {
	// Type is always "function".
	Type string `json:"type"`
	// Function is the function descriptor.
	Function ToolFunction `json:"function"`
}

// ToolFunction is the descriptor for a callable function.
type ToolFunction struct {
	// Name is the function name.
	Name string `json:"name"`
	// Description is an optional human-readable description.
	Description string `json:"description,omitempty"`
	// Parameters is the JSON Schema for the function arguments.
	Parameters any `json:"parameters,omitempty"`
}

// Stop represents the stop sequence(s) for a chat completion.
// It marshals as a JSON string when a single value is set, or a JSON array
// when multiple values are set.
type Stop struct {
	values []string
}

// StopString returns a Stop with a single string value.
func StopString(s string) Stop { return Stop{values: []string{s}} }

// StopStrings returns a Stop with multiple string values.
func StopStrings(ss []string) Stop { return Stop{values: ss} }

// MarshalJSON implements json.Marshaler.
func (s Stop) MarshalJSON() ([]byte, error) {
	switch len(s.values) {
	case 0:
		return []byte("null"), nil
	case 1:
		return json.Marshal(s.values[0])
	default:
		return json.Marshal(s.values)
	}
}

// IsZero reports whether s has no values.
func (s Stop) IsZero() bool { return len(s.values) == 0 }

// compile-time: Stop implements json.Marshaler
var _ json.Marshaler = Stop{}

// ChatRequest is the request body for POST /v1/chat/completions.
type ChatRequest struct {
	// Model is the model identifier.
	Model string `json:"model"`
	// Messages is the conversation history.
	Messages []Message `json:"messages"`
	// Temperature controls randomness (0–2).
	Temperature *float64 `json:"temperature,omitempty"`
	// TopP controls nucleus sampling.
	TopP *float64 `json:"top_p,omitempty"`
	// MaxTokens caps the output length.
	MaxTokens *int `json:"max_tokens,omitempty"`
	// Stream enables server-sent events streaming.
	Stream bool `json:"stream,omitempty"`
	// Stop is one or more sequences where the model stops generating.
	Stop *Stop `json:"stop,omitempty"`
	// Tools lists functions the model may call.
	Tools []Tool `json:"tools,omitempty"`
	// ToolChoice controls whether/how the model calls tools.
	ToolChoice any `json:"tool_choice,omitempty"`
}

// ChatCompletion is the non-streaming response from POST /v1/chat/completions.
type ChatCompletion struct {
	// ID is the unique completion identifier.
	ID string `json:"id"`
	// Object is always "chat.completion".
	Object string `json:"object"`
	// Created is a Unix timestamp.
	Created int64 `json:"created"`
	// Model is the model that generated the response.
	Model string `json:"model"`
	// Choices are the generated alternatives.
	Choices []ChatChoice `json:"choices"`
	// Usage holds token counts, if available.
	Usage *Usage `json:"usage,omitempty"`
}

// ChatChoice is one generated alternative in a ChatCompletion.
type ChatChoice struct {
	// Index is the zero-based choice index.
	Index int `json:"index"`
	// Message is the generated message.
	Message Message `json:"message"`
	// FinishReason explains why generation stopped.
	FinishReason *string `json:"finish_reason"`
}

// Usage holds token counts for a completion.
type Usage struct {
	// PromptTokens is the number of tokens in the prompt.
	PromptTokens int `json:"prompt_tokens"`
	// CompletionTokens is the number of tokens in the completion.
	CompletionTokens int `json:"completion_tokens"`
	// TotalTokens is PromptTokens + CompletionTokens.
	TotalTokens int `json:"total_tokens"`
}

// ChatCompletionChunk is a single SSE event in a streaming chat completion.
type ChatCompletionChunk struct {
	// ID is the unique completion identifier (same across all chunks).
	ID string `json:"id"`
	// Object is always "chat.completion.chunk".
	Object string `json:"object"`
	// Created is a Unix timestamp.
	Created int64 `json:"created"`
	// Model is the model that generated the response.
	Model string `json:"model"`
	// Choices are the partial alternatives.
	Choices []ChunkChoice `json:"choices"`
}

// ChunkChoice is one partial alternative in a ChatCompletionChunk.
type ChunkChoice struct {
	// Index is the zero-based choice index.
	Index int `json:"index"`
	// Delta is the incremental message fragment.
	Delta Message `json:"delta"`
	// FinishReason is non-nil on the final chunk for this choice.
	FinishReason *string `json:"finish_reason"`
}

// ── Models ────────────────────────────────────────────────────────────────────

// Model describes an available model.
type Model struct {
	// ID is the model identifier.
	ID string `json:"id"`
	// Object is always "model".
	Object string `json:"object"`
	// Created is a Unix timestamp.
	Created int64 `json:"created"`
	// OwnedBy is the model owner.
	OwnedBy string `json:"owned_by"`
}

// ModelList is the response from GET /v1/models.
type ModelList struct {
	// Object is always "list".
	Object string `json:"object"`
	// Data is the list of available models.
	Data []Model `json:"data"`
}

// ── Audio ─────────────────────────────────────────────────────────────────────

// TranscribeRequest is the multipart request for POST /v1/audio/transcriptions.
type TranscribeRequest struct {
	// File is the audio content to transcribe.
	File []byte
	// Filename is the name of the audio file (used for MIME sniffing).
	Filename string
	// Model is the transcription model to use.
	Model string
	// Language is the BCP-47 language code (e.g. "en", "de").
	Language string
	// Prompt is an optional context/spelling hint.
	Prompt string
	// ResponseFormat is "json", "text", "srt", "verbose_json", or "vtt".
	ResponseFormat string
	// Temperature controls sampling randomness.
	Temperature *float64
}

// Transcription is the response from POST /v1/audio/transcriptions.
type Transcription struct {
	// Text is the transcribed text.
	Text string `json:"text"`
}

// SpeechRequest is the JSON body for POST /v1/audio/speech.
type SpeechRequest struct {
	// Input is the text to synthesize (max 4096 chars).
	Input string `json:"input"`
	// Model is the TTS model to use.
	Model string `json:"model,omitempty"`
	// Voice is the voice identifier.
	Voice string `json:"voice,omitempty"`
	// ResponseFormat is "mp3", "opus", "aac", "flac", "wav", or "pcm".
	ResponseFormat string `json:"response_format,omitempty"`
	// Speed is the playback speed (0.25–4.0).
	Speed *float64 `json:"speed,omitempty"`
}

// ── Images ────────────────────────────────────────────────────────────────────

// ImageRequest is the JSON body for POST /v1/images/generations.
type ImageRequest struct {
	// Prompt describes the image to generate.
	Prompt string `json:"prompt"`
	// Model is the image generation model.
	Model string `json:"model,omitempty"`
	// N is the number of images to generate.
	N *int `json:"n,omitempty"`
	// Size is the image dimensions, e.g. "1024x1024".
	Size string `json:"size,omitempty"`
	// ResponseFormat is "url" or "b64_json".
	ResponseFormat string `json:"response_format,omitempty"`
}

// ImageResponse is the response from POST /v1/images/generations.
type ImageResponse struct {
	// Created is a Unix timestamp.
	Created int64 `json:"created"`
	// Data is the list of generated images.
	Data []ImageData `json:"data"`
}

// ImageData holds one generated image.
type ImageData struct {
	// URL is the image URL (when response_format is "url").
	URL string `json:"url,omitempty"`
	// B64JSON is the base64-encoded image (when response_format is "b64_json").
	B64JSON string `json:"b64_json,omitempty"`
	// RevisedPrompt is the prompt after revision (if the model revised it).
	RevisedPrompt string `json:"revised_prompt,omitempty"`
}

// ── Account Usage ─────────────────────────────────────────────────────────────

// AccountUsage is the response from GET /v1/usage.
type AccountUsage struct {
	// Balance is the current credit balance.
	Balance *float64 `json:"balance,omitempty"`
	// TotalSpent is the total credits spent.
	TotalSpent *float64 `json:"total_spent,omitempty"`
	// TotalRequests is the total number of API requests made.
	TotalRequests *int64 `json:"total_requests,omitempty"`
}

// ── API Keys ──────────────────────────────────────────────────────────────────

// APIKeyInfo describes an API key.
type APIKeyInfo struct {
	// ID is the key's unique identifier.
	ID string `json:"id"`
	// Name is the human-readable name.
	Name string `json:"name"`
	// Prefix is the visible key prefix.
	Prefix string `json:"prefix"`
	// UserID is the owning user's identifier.
	UserID string `json:"user_id"`
	// CreatedAt is the creation timestamp.
	CreatedAt string `json:"created_at"`
	// LastUsedAt is the last-use timestamp (nil if never used).
	LastUsedAt *string `json:"last_used_at,omitempty"`
	// BalanceEnabled indicates whether per-key balance enforcement is active.
	BalanceEnabled bool `json:"balance_enabled"`
	// BalanceSC is the current per-key balance in service credits.
	BalanceSC int64 `json:"balance_sc"`
	// BalanceLimitSC is the spending limit in service credits.
	BalanceLimitSC int64 `json:"balance_limit_sc"`
	// RateLimitRPM is the per-minute request limit (nil = unlimited).
	RateLimitRPM *int `json:"rate_limit_rpm,omitempty"`
	// RateLimitTPM is the per-minute token limit (nil = unlimited).
	RateLimitTPM *int `json:"rate_limit_tpm,omitempty"`
}

// APIKeyList is the response from GET /v1/keys.
type APIKeyList struct {
	// APIKeys is the list of API keys.
	APIKeys []APIKeyInfo `json:"api_keys"`
}

// APIKeyUsage holds aggregate usage statistics for an API key.
type APIKeyUsage struct {
	// TotalRequests is the total number of requests made with this key.
	TotalRequests int64 `json:"total_requests"`
	// TotalTokensIn is the total number of input tokens.
	TotalTokensIn int64 `json:"total_tokens_in"`
	// TotalTokensOut is the total number of output tokens.
	TotalTokensOut int64 `json:"total_tokens_out"`
	// TotalCostSC is the total cost in service credits.
	TotalCostSC int64 `json:"total_cost_sc"`
}

// APIKeyDailyUsage holds usage statistics for one day.
type APIKeyDailyUsage struct {
	// Date is the calendar date (YYYY-MM-DD).
	Date string `json:"date"`
	// Requests is the number of requests on this date.
	Requests int64 `json:"requests"`
	// TokensIn is the number of input tokens on this date.
	TokensIn int64 `json:"tokens_in"`
	// TokensOut is the number of output tokens on this date.
	TokensOut int64 `json:"tokens_out"`
	// CostSC is the cost in service credits on this date.
	CostSC int64 `json:"cost_sc"`
	// Model is the model used (may be aggregated).
	Model string `json:"model"`
}

// APIKeyDailyUsageList is the response from daily-usage endpoints.
type APIKeyDailyUsageList struct {
	// Usage is the per-day breakdown.
	Usage []APIKeyDailyUsage `json:"usage"`
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// Ptr returns a pointer to v. Useful for setting optional fields inline:
//
//	simplellm.Ptr(0.7)  // *float64
//	simplellm.Ptr(1024) // *int
func Ptr[T any](v T) *T { return &v }
