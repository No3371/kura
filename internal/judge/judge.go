// Package judge is a minimal client for TypeSafe's Jev classifier
// (POST /v1/systemone). Jev answers typed questions about one state in a
// single call: Choice questions return a label with per-option
// probabilities and a confidence, Noul questions return P(yes).
//
// There is no Go SDK, so this speaks plain HTTP, as the OpenAI advisor does.
package judge

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// DefaultBaseURL is Jev's API origin; requests go to BaseURL + "/v1/systemone".
const DefaultBaseURL = "https://api.typesafe.ai"

// Client calls Jev. It sets no timeout of its own: the caller's context
// bounds every request, so one budget covers connect, send and decode.
type Client struct {
	Key     string
	Model   string // a versioned id such as jev-1.14.0; an alias would move tuned thresholds silently
	BaseURL string
	HTTP    *http.Client
}

// New returns a client for model on the public API.
func New(key, model string) *Client {
	return &Client{Key: key, Model: model, BaseURL: DefaultBaseURL, HTTP: http.DefaultClient}
}

// Question is one named question in a request. Type is "choice" or "noul".
// Criteria carries a Choice's options as {option: description}, or a Noul's
// {"true": …, "false": …}; Jev rejects Choice options under any other field.
type Question struct {
	Type         string `json:"type"`
	Instructions string `json:"instructions"`
	Criteria     any    `json:"criteria,omitempty"`
}

// Answer is Jev's answer to one question. A Choice answer fills Choice,
// Confidence and Probabilities; a Noul answer fills only Noul, because Jev
// returns no confidence for it (|2p − 1| is the caller's to compute).
type Answer struct {
	Type          string             `json:"type"`
	Choice        string             `json:"choice,omitempty"`
	Noul          *float64           `json:"noul,omitempty"`
	Confidence    *float64           `json:"confidence,omitempty"`
	Probabilities map[string]float64 `json:"probabilities,omitempty"`
}

type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// Result is one response: the versioned model that answered, and an answer
// per question name.
type Result struct {
	Model   string            `json:"model"`
	Answers map[string]Answer `json:"answers"`
	Usage   Usage             `json:"usage"`
}

// APIError is a non-2xx response. Jev reports errors as
// {"detail":{"error_type":…,"message":…}}; a 403 is a missing or bad key.
type APIError struct {
	Status  int
	Type    string // detail.error_type, e.g. authentication_error
	Message string
}

func (e *APIError) Error() string {
	var b strings.Builder
	fmt.Fprintf(&b, "jev: HTTP %d", e.Status)
	if e.Type != "" {
		b.WriteString(" " + e.Type)
	}
	if e.Message != "" {
		b.WriteString(": " + e.Message)
	}
	return b.String()
}

// RateLimited reports a 429, which the gate logs apart from other failures.
func (e *APIError) RateLimited() bool { return e.Status == http.StatusTooManyRequests }

// Auth reports a rejected key (Jev answers 403 for a missing or bad key; 401
// is accepted too in case that changes).
func (e *APIError) Auth() bool {
	return e.Status == http.StatusForbidden || e.Status == http.StatusUnauthorized
}

// Ask sends one request with every question against one state. state may be
// a string, a JSON object or a list of texts.
func (c *Client) Ask(ctx context.Context, state any, qs map[string]Question) (Result, error) {
	body, err := json.Marshal(map[string]any{"model": c.Model, "state": state, "questions": qs})
	if err != nil {
		return Result{}, fmt.Errorf("jev: encode request: %w", err)
	}
	base := c.BaseURL
	if base == "" {
		base = DefaultBaseURL
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(base, "/")+"/v1/systemone", bytes.NewReader(body))
	if err != nil {
		return Result{}, fmt.Errorf("jev: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.Key)
	req.Header.Set("Content-Type", "application/json")
	hc := c.HTTP
	if hc == nil {
		hc = http.DefaultClient
	}
	resp, err := hc.Do(req)
	if err != nil {
		return Result{}, fmt.Errorf("jev: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return Result{}, fmt.Errorf("jev: read response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return Result{}, decodeError(resp.StatusCode, raw)
	}
	var res Result
	if err := json.Unmarshal(raw, &res); err != nil {
		return Result{}, fmt.Errorf("jev: decode response: %w", err)
	}
	if res.Answers == nil {
		return Result{}, errors.New("jev: response has no answers")
	}
	return res, nil
}

// decodeError reads detail.error_type / detail.message. A string detail
// (the framework's default shape) or an unreadable body still yields an
// APIError carrying the status, so callers can branch on it.
func decodeError(status int, raw []byte) error {
	e := &APIError{Status: status}
	var body struct {
		Detail json.RawMessage `json:"detail"`
	}
	if json.Unmarshal(raw, &body) != nil || len(body.Detail) == 0 {
		return e
	}
	var d struct {
		ErrorType string `json:"error_type"`
		Message   string `json:"message"`
	}
	if json.Unmarshal(body.Detail, &d) == nil {
		e.Type, e.Message = d.ErrorType, d.Message
		return e
	}
	var s string
	if json.Unmarshal(body.Detail, &s) == nil {
		e.Message = s
	}
	return e
}
