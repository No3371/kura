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
	"sort"
	"strings"
)

// OpenAI runs the same advisor on the OpenAI Chat Completions API: the same
// system prompt, tools and tool runner as Claude, so only the model differs.
// It speaks plain HTTP rather than pulling in an SDK; BaseURL lets it target
// any Chat Completions-compatible endpoint.
type OpenAI struct {
	APIKey  string
	BaseURL string // e.g. https://api.openai.com/v1
	Model   string
	// Effort is sent as reasoning_effort when set; models without reasoning
	// reject the field, so it is opt-in.
	Effort    string
	MaxTokens int
	Tools     *Tools
	Log       *slog.Logger
	HTTP      *http.Client

	sessions sessionStore[oaMessage]
	system   string
	tools    []oaTool
}

type oaMessage struct {
	Role       string       `json:"role"`
	Content    string       `json:"content"`
	ToolCalls  []oaToolCall `json:"tool_calls,omitempty"`
	ToolCallID string       `json:"tool_call_id,omitempty"`
}

type oaToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

type oaTool struct {
	Type     string `json:"type"`
	Function struct {
		Name        string         `json:"name"`
		Description string         `json:"description"`
		Parameters  map[string]any `json:"parameters"`
	} `json:"function"`
}

func NewOpenAI(apiKey, baseURL, model, effort string, tools *Tools, log *slog.Logger) *OpenAI {
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	o := &OpenAI{APIKey: apiKey, BaseURL: strings.TrimRight(baseURL, "/"), Model: model, Effort: effort,
		MaxTokens: 16384, Tools: tools, Log: log, HTTP: http.DefaultClient, system: buildSystemPrompt()}
	for _, t := range toolSpecs() {
		var ot oaTool
		ot.Type = "function"
		ot.Function.Name, ot.Function.Description = t.Name, t.Description
		params := map[string]any{"type": "object", "properties": t.Properties}
		if len(t.Required) > 0 {
			params["required"] = t.Required
		}
		ot.Function.Parameters = params
		o.tools = append(o.tools, ot)
	}
	return o
}

func (o *OpenAI) Mode() string { return "openai:" + o.Model }

func (o *OpenAI) Chat(ctx context.Context, sessionID, message string, emit func(Event)) error {
	s := o.sessions.get(sessionID)
	if !s.mu.TryLock() {
		return errors.New("這個對話還在回覆上一則訊息，請稍候")
	}
	defer s.mu.Unlock()

	// As Claude: a failed turn is rolled back so the next starts clean.
	checkpoint := len(s.messages)
	s.messages = append(s.messages, oaMessage{Role: "user", Content: message})
	err := o.run(ctx, s, emit)
	if err != nil {
		s.messages = s.messages[:checkpoint]
	}
	return err
}

func (o *OpenAI) run(ctx context.Context, s *conversation[oaMessage], emit func(Event)) error {
	for round := 0; round <= maxToolRounds; round++ {
		msg, finish, err := o.stream(ctx, s.messages, emit)
		if err != nil {
			return err
		}
		switch finish {
		case "length":
			// Tool arguments may be truncated; never execute them.
			return errors.New("回覆過長被截斷，請把問題縮小一些再試一次")
		case "content_filter":
			return errors.New("這個請求無法處理，請換個方式描述您想找的包款")
		}
		s.messages = append(s.messages, msg)
		// Some compatible endpoints end a tool turn with "stop", so the calls
		// themselves decide whether to continue.
		if len(msg.ToolCalls) == 0 {
			o.Log.Info("advisor turn", "rounds", round)
			return nil
		}
		for _, tc := range msg.ToolCalls {
			name, raw := tc.Function.Name, tc.Function.Arguments
			if strings.TrimSpace(raw) == "" {
				raw = "{}"
			}
			emit(Event{Type: "tool_call", Tool: name, Input: raw})
			res, err := o.Tools.Run(name, raw)
			if err != nil {
				emit(Event{Type: "tool_result", Tool: name, Summary: err.Error(), IsError: true})
				s.messages = append(s.messages, oaMessage{Role: "tool", ToolCallID: tc.ID, Content: "error: " + err.Error()})
				continue
			}
			emit(Event{Type: "tool_result", Tool: name, Summary: res.Summary})
			if len(res.Listings) > 0 {
				emit(Event{Type: "listings", Listings: res.Listings})
			}
			s.messages = append(s.messages, oaMessage{Role: "tool", ToolCallID: tc.ID, Content: res.JSON})
		}
	}
	return fmt.Errorf("查詢步驟過多（超過 %d 輪），請把問題拆小一點", maxToolRounds)
}

type oaChunk struct {
	Choices []struct {
		Delta struct {
			Content   string `json:"content"`
			Refusal   string `json:"refusal"`
			ToolCalls []struct {
				Index    int    `json:"index"`
				ID       string `json:"id"`
				Function struct {
					Name      string `json:"name"`
					Arguments string `json:"arguments"`
				} `json:"function"`
			} `json:"tool_calls"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage *struct {
		PromptTokens        int `json:"prompt_tokens"`
		CompletionTokens    int `json:"completion_tokens"`
		PromptTokensDetails struct {
			CachedTokens int `json:"cached_tokens"`
		} `json:"prompt_tokens_details"`
	} `json:"usage"`
}

// stream runs one model call, forwarding text deltas as they arrive, and
// returns the assembled assistant message with its finish reason.
func (o *OpenAI) stream(ctx context.Context, history []oaMessage, emit func(Event)) (oaMessage, string, error) {
	msgs := append([]oaMessage{{Role: "system", Content: o.system}}, history...)
	req := map[string]any{
		"model": o.Model, "messages": msgs, "tools": o.tools, "stream": true,
		"stream_options":        map[string]any{"include_usage": true},
		"max_completion_tokens": o.MaxTokens,
	}
	if o.Effort != "" {
		req["reasoning_effort"] = o.Effort
	}
	body, err := json.Marshal(req)
	if err != nil {
		return oaMessage{}, "", err
	}
	hr, err := http.NewRequestWithContext(ctx, http.MethodPost, o.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return oaMessage{}, "", err
	}
	hr.Header.Set("Content-Type", "application/json")
	hr.Header.Set("Authorization", "Bearer "+o.APIKey)
	resp, err := o.HTTP.Do(hr)
	if err != nil {
		return oaMessage{}, "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return oaMessage{}, "", describeOpenAIError(resp)
	}

	msg := oaMessage{Role: "assistant"}
	var text, refusal strings.Builder
	calls := map[int]*oaToolCall{}
	var finish string
	sc := bufio.NewScanner(resp.Body)
	sc.Buffer(make([]byte, 64*1024), 4*1024*1024)
	for sc.Scan() {
		data, ok := strings.CutPrefix(sc.Text(), "data:")
		if !ok {
			continue
		}
		data = strings.TrimSpace(data)
		if data == "[DONE]" {
			break
		}
		var ch oaChunk
		if err := json.Unmarshal([]byte(data), &ch); err != nil {
			return msg, "", fmt.Errorf("OpenAI 串流格式錯誤：%w", err)
		}
		if u := ch.Usage; u != nil {
			o.Log.Info("advisor call", "in", u.PromptTokens, "out", u.CompletionTokens, "cached", u.PromptTokensDetails.CachedTokens)
		}
		for _, c := range ch.Choices {
			if c.Delta.Content != "" {
				text.WriteString(c.Delta.Content)
				emit(Event{Type: "text", Text: c.Delta.Content})
			}
			refusal.WriteString(c.Delta.Refusal)
			// A call arrives in pieces keyed by index: id and name first,
			// then the arguments JSON in fragments.
			for _, d := range c.Delta.ToolCalls {
				tc := calls[d.Index]
				if tc == nil {
					tc = &oaToolCall{Type: "function"}
					calls[d.Index] = tc
				}
				if d.ID != "" {
					tc.ID = d.ID
				}
				tc.Function.Name += d.Function.Name
				tc.Function.Arguments += d.Function.Arguments
			}
			if c.FinishReason != "" {
				finish = c.FinishReason
			}
		}
	}
	if err := sc.Err(); err != nil {
		return msg, "", err
	}
	if refusal.Len() > 0 {
		return msg, "", errors.New("這個請求無法處理，請換個方式描述您想找的包款")
	}
	msg.Content = text.String()
	idx := make([]int, 0, len(calls))
	for i := range calls {
		idx = append(idx, i)
	}
	sort.Ints(idx)
	for _, i := range idx {
		msg.ToolCalls = append(msg.ToolCalls, *calls[i])
	}
	return msg, finish, nil
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
		return fmt.Errorf("OpenAI API 驗證失敗，請確認 OPENAI_API_KEY（%w）", err)
	case resp.StatusCode == 429:
		return fmt.Errorf("OpenAI API 請求過於頻繁或額度不足，請稍後再試（%w）", err)
	case resp.StatusCode >= 500:
		return fmt.Errorf("OpenAI API 暫時無法使用，請稍後再試（%w）", err)
	}
	return err
}
