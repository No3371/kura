package advisor

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
)

// OpenAI runs the same advisor on the OpenAI Responses API: the same system
// prompt, tools and tool runner as Claude, so only the model differs. It
// speaks plain HTTP rather than pulling in an SDK; BaseURL lets it target any
// Responses-compatible endpoint.
type OpenAI struct {
	APIKey  string
	BaseURL string // e.g. https://api.openai.com/v1
	Model   string
	// Effort is sent as reasoning.effort when set; models without reasoning
	// reject the field, so it is opt-in.
	Effort    string
	MaxTokens int
	Tools     *Tools
	Log       *slog.Logger
	HTTP      *http.Client

	// History is kept here as raw input items, not on OpenAI's side (store is
	// off): each response's output items are replayed verbatim, reasoning
	// included, as the API requires for tool calls.
	sessions sessionStore[json.RawMessage]
	system   string
	tools    []oaTool
}

// oaTool is a Responses function tool, flat rather than nested under
// "function" as in Chat Completions.
type oaTool struct {
	Type        string         `json:"type"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

// oaItem is the part of an output item the loop reads.
type oaItem struct {
	Type      string `json:"type"`
	CallID    string `json:"call_id"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

func NewOpenAI(apiKey, baseURL, model, effort string, tools *Tools, log *slog.Logger) *OpenAI {
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	o := &OpenAI{APIKey: apiKey, BaseURL: strings.TrimRight(baseURL, "/"), Model: model, Effort: effort,
		MaxTokens: 16384, Tools: tools, Log: log, HTTP: http.DefaultClient, system: buildSystemPrompt()}
	for _, t := range toolSpecs() {
		params := map[string]any{"type": "object", "properties": t.Properties}
		if len(t.Required) > 0 {
			params["required"] = t.Required
		}
		o.tools = append(o.tools, oaTool{Type: "function", Name: t.Name, Description: t.Description, Parameters: params})
	}
	return o
}

func (o *OpenAI) Mode() string { return "openai:" + o.Model }

func (o *OpenAI) Chat(ctx context.Context, sessionID, message string, emit func(Event)) error {
	s := o.sessions.get(sessionID)
	if !s.mu.TryLock() {
		return shopperErr("這個對話還在回覆上一則訊息，請稍候", nil)
	}
	defer s.mu.Unlock()

	// As Claude: a failed turn is rolled back so the next starts clean.
	checkpoint := len(s.messages)
	s.messages = append(s.messages, oaJSON(map[string]string{"role": "user", "content": message}))
	err := o.run(ctx, s, emit)
	if err == nil {
		// As Claude: a turn finished from an already-received reply after
		// its context was cancelled still counts as cancelled.
		err = ctx.Err()
	}
	if err != nil {
		s.messages = s.messages[:checkpoint]
		return err
	}
	s.commit(checkpoint)
	return nil
}

func (o *OpenAI) run(ctx context.Context, s *conversation[json.RawMessage], emit func(Event)) error {
	for round := 0; round <= maxToolRounds; round++ {
		output, err := o.stream(ctx, s.messages, emit)
		if err != nil {
			return err
		}
		var calls []oaItem
		for _, raw := range output {
			var it oaItem
			if err := json.Unmarshal(raw, &it); err != nil {
				return fmt.Errorf("OpenAI output item: %w", err)
			}
			s.messages = append(s.messages, raw)
			if it.Type == "function_call" {
				calls = append(calls, it)
			}
		}
		if len(calls) == 0 {
			o.Log.Info("advisor turn", "rounds", round)
			return nil
		}
		for _, c := range calls {
			raw := c.Arguments
			if strings.TrimSpace(raw) == "" {
				raw = "{}"
			}
			emit(Event{Type: "tool_call", Tool: c.Name, Input: raw})
			out := ""
			res, err := o.Tools.Run(c.Name, raw)
			if err != nil {
				emit(Event{Type: "tool_result", Tool: c.Name, Summary: err.Error(), IsError: true})
				out = "error: " + err.Error()
			} else {
				emit(Event{Type: "tool_result", Tool: c.Name, Summary: res.Summary})
				if len(res.Listings) > 0 {
					emit(Event{Type: "listings", Listings: res.Listings})
				}
				out = res.JSON
			}
			s.messages = append(s.messages, oaJSON(map[string]string{"type": "function_call_output", "call_id": c.CallID, "output": out}))
		}
	}
	return shopperErr(fmt.Sprintf("查詢步驟過多（超過 %d 輪），請把問題拆小一點", maxToolRounds), nil)
}

func oaJSON(v any) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}

// oaEvent is the part of a Responses stream event the client reads.
type oaEvent struct {
	Type     string `json:"type"`
	Delta    string `json:"delta"`
	Code     string `json:"code"`    // type "error"
	Message  string `json:"message"` // type "error"
	Response struct {
		Output []json.RawMessage `json:"output"`
		Error  *struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
		IncompleteDetails *struct {
			Reason string `json:"reason"`
		} `json:"incomplete_details"`
		Usage *struct {
			InputTokens        int `json:"input_tokens"`
			OutputTokens       int `json:"output_tokens"`
			InputTokensDetails struct {
				CachedTokens int `json:"cached_tokens"`
			} `json:"input_tokens_details"`
		} `json:"usage"`
	} `json:"response"`
}

// stream runs one model call, forwarding text deltas as they arrive, and
// returns the response's output items.
func (o *OpenAI) stream(ctx context.Context, input []json.RawMessage, emit func(Event)) ([]json.RawMessage, error) {
	req := map[string]any{
		"model": o.Model, "instructions": o.system, "input": input, "tools": o.tools,
		"stream": true, "store": false, "max_output_tokens": o.MaxTokens,
		// With store off, reasoning items only replay with their encrypted content.
		"include": []string{"reasoning.encrypted_content"},
	}
	if o.Effort != "" {
		req["reasoning"] = map[string]string{"effort": o.Effort}
	}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	hr, err := http.NewRequestWithContext(ctx, http.MethodPost, o.BaseURL+"/responses", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	hr.Header.Set("Content-Type", "application/json")
	hr.Header.Set("Authorization", "Bearer "+o.APIKey)
	resp, err := o.HTTP.Do(hr)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, describeOpenAIError(resp)
	}

	refused := false
	sc := bufio.NewScanner(resp.Body)
	sc.Buffer(make([]byte, 64*1024), 16*1024*1024)
	for sc.Scan() {
		data, ok := strings.CutPrefix(sc.Text(), "data:")
		if !ok {
			continue
		}
		var ev oaEvent
		if err := json.Unmarshal([]byte(strings.TrimSpace(data)), &ev); err != nil {
			return nil, fmt.Errorf("OpenAI stream event: %w", err)
		}
		r := ev.Response
		switch ev.Type {
		case "response.output_text.delta":
			emit(Event{Type: "text", Text: ev.Delta})
		case "response.refusal.delta":
			refused = true
		case "response.completed":
			if u := r.Usage; u != nil {
				o.Log.Info("advisor call", "in", u.InputTokens, "out", u.OutputTokens, "cached", u.InputTokensDetails.CachedTokens)
			}
			if refused {
				return nil, shopperErr("這個請求無法處理，請換個方式描述您想找的包款", nil)
			}
			return r.Output, nil
		case "response.incomplete":
			// Tool arguments may be truncated; never execute them.
			reason := ""
			if r.IncompleteDetails != nil {
				reason = r.IncompleteDetails.Reason
			}
			if reason == "content_filter" {
				return nil, shopperErr("這個請求無法處理，請換個方式描述您想找的包款", nil)
			}
			return nil, shopperErr("回覆過長被截斷，請把問題縮小一些再試一次", fmt.Errorf("OpenAI response incomplete: %s", reason))
		case "response.failed":
			err := errors.New("OpenAI response failed")
			if e := r.Error; e != nil {
				err = fmt.Errorf("OpenAI response failed: %s: %s", e.Code, e.Message)
				if e.Code == "rate_limit_exceeded" {
					return nil, shopperErr("目前詢問的人較多，請稍後再試", err)
				}
			}
			return nil, err
		case "error":
			return nil, fmt.Errorf("OpenAI stream error: %s: %s", ev.Code, ev.Message)
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return nil, errors.New("OpenAI stream ended before response.completed")
}

func describeOpenAIError(resp *http.Response) error {
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	var e struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	detail := strings.TrimSpace(string(b))
	if json.Unmarshal(b, &e) == nil && e.Error.Message != "" {
		detail = e.Error.Message
	}
	err := fmt.Errorf("OpenAI API %d: %s", resp.StatusCode, detail)
	switch {
	case resp.StatusCode == 401 || resp.StatusCode == 403:
		return shopperErr("顧問暫時無法使用，請稍後再試", fmt.Errorf("OpenAI API 驗證失敗，請確認 OPENAI_API_KEY: %w", err))
	case resp.StatusCode == 429:
		return shopperErr("目前詢問的人較多，請稍後再試", err)
	case resp.StatusCode >= 500:
		return shopperErr("顧問目前忙碌中，請稍後再試", err)
	}
	return err
}
