package simplellm

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Error represents an API error returned by the SimpleLLM API.
// It implements the error interface.
type Error struct {
	// Message is the human-readable error description.
	Message string
	// Status is the HTTP status code.
	Status int
	// Code is the machine-readable error code, if provided.
	Code string
}

// Error implements the error interface.
func (e *Error) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("simplellm: %s (status %d, code %s)", e.Message, e.Status, e.Code)
	}
	return fmt.Sprintf("simplellm: %s (status %d)", e.Message, e.Status)
}

// parseError reads the response body and returns an *Error.
// It tries two JSON shapes:
//
//	{"error":{"message":"...","code":"..."}}
//	{"message":"...","code":"..."}
//
// Falls back to HTTP status text if neither parses.
func parseError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)

	// Shape 1: {"error":{"message":"...","code":"..."}}
	var shape1 struct {
		Error struct {
			Message string `json:"message"`
			Code    string `json:"code"`
		} `json:"error"`
	}
	if json.Unmarshal(body, &shape1) == nil && shape1.Error.Message != "" {
		return &Error{
			Message: shape1.Error.Message,
			Status:  resp.StatusCode,
			Code:    shape1.Error.Code,
		}
	}

	// Shape 2: {"message":"...","code":"..."}
	var shape2 struct {
		Message string `json:"message"`
		Code    string `json:"code"`
	}
	if json.Unmarshal(body, &shape2) == nil && shape2.Message != "" {
		return &Error{
			Message: shape2.Message,
			Status:  resp.StatusCode,
			Code:    shape2.Code,
		}
	}

	// Fallback
	return &Error{
		Message: http.StatusText(resp.StatusCode),
		Status:  resp.StatusCode,
	}
}
