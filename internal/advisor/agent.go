// Package advisor implements the "AI 找包顧問": a Claude tool-use agent that
// answers bag-hunting questions using only Kura's own listing data, plus an
// offline rule-based fallback for running the demo without an API key.
package advisor

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/shared/constant"

	"github.com/No3371/kura/internal/radar"
)

// Event is streamed to the browser over SSE.
type Event struct {
	Type     string       `json:"type"` // text | tool_call | tool_result | listings | error | done
	Text     string       `json:"text,omitempty"`
	Tool     string       `json:"tool,omitempty"`
	Input    string       `json:"input,omitempty"`
	Summary  string       `json:"summary,omitempty"`
	IsError  bool         `json:"is_error,omitempty"`
	Listings []radar.Item `json:"listings,omitempty"`
}

// Advisor answers one user message in a session, streaming events.
type Advisor interface {
	Chat(ctx context.Context, sessionID, message string, emit func(Event)) error
	Mode() string
}

const (
	maxToolRounds = 8
	maxSessions   = 500
)

type Claude struct {
	Client anthropic.Client
	Model  string
	Tools  *Tools
	Log    *slog.Logger

	mu       sync.Mutex
	sessions map[string]*session
	system   string
	tools    []anthropic.BetaToolUnionParam
}

type session struct {
	mu       sync.Mutex // one in-flight turn per session
	messages []anthropic.BetaMessageParam
	used     time.Time
}

func NewClaude(client anthropic.Client, model string, tools *Tools, log *slog.Logger) *Claude {
	return &Claude{Client: client, Model: model, Tools: tools, Log: log,
		sessions: map[string]*session{}, system: buildSystemPrompt(), tools: toolDefs()}
}

func (c *Claude) Mode() string { return "claude:" + c.Model }

func (c *Claude) session(id string) *session {
	c.mu.Lock()
	defer c.mu.Unlock()
	s, ok := c.sessions[id]
	if !ok {
		if len(c.sessions) >= maxSessions {
			c.evictOldest()
		}
		s = &session{}
		c.sessions[id] = s
	}
	s.used = time.Now()
	return s
}

func (c *Claude) evictOldest() {
	var oldest string
	var at time.Time
	for id, s := range c.sessions {
		if oldest == "" || s.used.Before(at) {
			oldest, at = id, s.used
		}
	}
	delete(c.sessions, oldest)
}

func (c *Claude) params(messages []anthropic.BetaMessageParam) anthropic.BetaMessageNewParams {
	return anthropic.BetaMessageNewParams{
		Model:     anthropic.Model(c.Model),
		MaxTokens: 64000,
		System:    []anthropic.BetaTextBlockParam{{Text: c.system, CacheControl: anthropic.NewBetaCacheControlEphemeralParam()}},
		Tools:     c.tools,
		Messages:  messages,
		// Chat with several tool hops: medium effort keeps replies quick
		// without skimping on the search strategy.
		OutputConfig: anthropic.BetaOutputConfigParam{Effort: anthropic.BetaOutputConfigEffortMedium},
		// If a safety classifier declines, let the API re-serve the turn on a
		// fallback model chosen by refusal category instead of failing.
		Fallbacks: anthropic.BetaFallbacksParamUnion{OfDefault: constant.ValueOf[constant.Default]()},
		Betas:     []anthropic.AnthropicBeta{anthropic.AnthropicBetaServerSideFallback2026_07_01},
	}
}

func (c *Claude) Chat(ctx context.Context, sessionID, message string, emit func(Event)) error {
	s := c.session(sessionID)
	if !s.mu.TryLock() {
		return errors.New("這個對話還在回覆上一則訊息，請稍候")
	}
	defer s.mu.Unlock()

	// History is append-only (thinking blocks must be replayed unchanged), and
	// a failed turn is rolled back so the next one starts from a valid state.
	checkpoint := len(s.messages)
	s.messages = append(s.messages, anthropic.NewBetaUserMessage(anthropic.NewBetaTextBlock(message)))
	err := c.run(ctx, s, emit)
	if err != nil {
		s.messages = s.messages[:checkpoint]
	}
	return err
}

func (c *Claude) run(ctx context.Context, s *session, emit func(Event)) error {
	for round := 0; round <= maxToolRounds; round++ {
		msg, err := c.stream(ctx, s.messages, emit)
		if err != nil {
			return err
		}
		s.messages = append(s.messages, msg.ToParam())

		switch msg.StopReason {
		case anthropic.BetaStopReasonToolUse:
			// fall through to tool execution below
		case anthropic.BetaStopReasonRefusal:
			return errors.New("這個請求無法處理，請換個方式描述您想找的包款")
		case anthropic.BetaStopReasonMaxTokens:
			// Tool input may be truncated; never execute it.
			return errors.New("回覆過長被截斷，請把問題縮小一些再試一次")
		default:
			c.Log.Info("advisor turn", "rounds", round, "in", msg.Usage.InputTokens, "out", msg.Usage.OutputTokens,
				"cache_read", msg.Usage.CacheReadInputTokens)
			return nil
		}

		var results []anthropic.BetaContentBlockParamUnion
		for _, block := range msg.Content {
			tu, ok := block.AsAny().(anthropic.BetaToolUseBlock)
			if !ok {
				continue
			}
			raw := tu.JSON.Input.Raw()
			emit(Event{Type: "tool_call", Tool: tu.Name, Input: raw})
			res, err := c.Tools.Run(tu.Name, raw)
			if err != nil {
				emit(Event{Type: "tool_result", Tool: tu.Name, Summary: err.Error(), IsError: true})
				results = append(results, anthropic.NewBetaToolResultBlock(tu.ID, err.Error(), true))
				continue
			}
			emit(Event{Type: "tool_result", Tool: tu.Name, Summary: res.Summary})
			if len(res.Listings) > 0 {
				emit(Event{Type: "listings", Listings: res.Listings})
			}
			results = append(results, anthropic.NewBetaToolResultBlock(tu.ID, res.JSON, false))
		}
		// All results go back in one user turn so parallel calls stay parallel.
		s.messages = append(s.messages, anthropic.NewBetaUserMessage(results...))
	}
	return fmt.Errorf("查詢步驟過多（超過 %d 輪），請把問題拆小一點", maxToolRounds)
}

// stream runs one model call, forwarding text deltas as they arrive.
func (c *Claude) stream(ctx context.Context, messages []anthropic.BetaMessageParam, emit func(Event)) (anthropic.BetaMessage, error) {
	stream := c.Client.Beta.Messages.NewStreaming(ctx, c.params(messages))
	defer stream.Close()
	var msg anthropic.BetaMessage
	for stream.Next() {
		ev := stream.Current()
		if err := msg.Accumulate(ev); err != nil {
			return msg, err
		}
		if d, ok := ev.AsAny().(anthropic.BetaRawContentBlockDeltaEvent); ok {
			if t, ok := d.Delta.AsAny().(anthropic.BetaTextDelta); ok {
				emit(Event{Type: "text", Text: t.Text})
			}
		}
	}
	if err := stream.Err(); err != nil {
		return msg, describeAPIError(err)
	}
	return msg, nil
}

func describeAPIError(err error) error {
	if apiErr, ok := errors.AsType[*anthropic.Error](err); ok {
		switch apiErr.StatusCode {
		case 401, 403:
			return fmt.Errorf("Claude API 驗證失敗，請確認 ANTHROPIC_API_KEY（%w）", err)
		case 429:
			return fmt.Errorf("Claude API 請求過於頻繁，請稍後再試（%w）", err)
		case 529:
			return fmt.Errorf("Claude API 目前負載過高，請稍後再試（%w）", err)
		}
	}
	return err
}
