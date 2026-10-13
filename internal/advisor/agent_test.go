package advisor

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"

	"github.com/No3371/kura/internal/fx"
	"github.com/No3371/kura/internal/listing"
	"github.com/No3371/kura/internal/radar"
	"github.com/No3371/kura/internal/store"
)

// fakeAPI replays scripted SSE responses for POST /v1/messages and records
// every request body, so the agent loop can be tested without an API key.
type fakeAPI struct {
	mu       sync.Mutex
	scripts  []string
	requests []map[string]any
	headers  []http.Header
}

func (f *fakeAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	var req map[string]any
	json.Unmarshal(body, &req)
	f.mu.Lock()
	i := len(f.requests)
	f.requests = append(f.requests, req)
	f.headers = append(f.headers, r.Header.Clone())
	f.mu.Unlock()
	if i >= len(f.scripts) {
		http.Error(w, `{"type":"error","error":{"type":"invalid_request_error","message":"unexpected call"}}`, 400)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	io.WriteString(w, f.scripts[i])
}

func sse(events ...string) string {
	var b strings.Builder
	for _, e := range events {
		var probe struct{ Type string }
		json.Unmarshal([]byte(e), &probe)
		fmt.Fprintf(&b, "event: %s\ndata: %s\n\n", probe.Type, e)
	}
	return b.String()
}

const msgStart = `{"type":"message_start","message":{"id":"msg_%d","type":"message","role":"assistant","model":"claude-opus-5-5","content":[],"stop_reason":null,"stop_sequence":null,"usage":{"input_tokens":100,"output_tokens":1}}}`

func toolTurn(input string) string {
	return sse(
		fmt.Sprintf(msgStart, 1),
		`{"type":"content_block_start","index":0,"content_block":{"type":"tool_use","id":"toolu_1","name":"search_listings","input":{}}}`,
		`{"type":"content_block_delta","index":0,"delta":{"type":"input_json_delta","partial_json":`+jsonString(input)+`}}`,
		`{"type":"content_block_stop","index":0}`,
		`{"type":"message_delta","delta":{"stop_reason":"tool_use","stop_sequence":null},"usage":{"output_tokens":20}}`,
		`{"type":"message_stop"}`,
	)
}

func textTurn(text string) string {
	return sse(
		fmt.Sprintf(msgStart, 2),
		`{"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}`,
		`{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":`+jsonString(text)+`}}`,
		`{"type":"content_block_stop","index":0}`,
		`{"type":"message_delta","delta":{"stop_reason":"end_turn","stop_sequence":null},"usage":{"output_tokens":30}}`,
		`{"type":"message_stop"}`,
	)
}

func jsonString(s string) string { b, _ := json.Marshal(s); return string(b) }

func testTools(t *testing.T) *Tools {
	t.Helper()
	ctx := context.Background()
	st, err := store.Open(filepath.Join(t.TempDir(), "kura.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	for i, p := range []int64{300000, 320000, 340000, 250000} {
		l := listing.Listing{Source: "brandoff", SourceID: fmt.Sprint(1000 + i), URL: "https://example.test",
			Brand: "HERMES", Title: "エルメス ピコタンロックPM", ModelName: "ピコタンロックPM", ModelCode: "056289CK",
			Category: "handbag", Rank: listing.RankA, PriceJPY: p}
		if _, err := st.RecordSeen(ctx, "HERMES", l, time.Now()); err != nil {
			t.Fatal(err)
		}
	}
	conv := &fx.Converter{Store: st, Fallback: 0.2}
	eng := &radar.Engine{Store: st, FX: conv}
	if err := eng.Refresh(ctx); err != nil {
		t.Fatal(err)
	}
	return &Tools{Radar: eng, FX: conv}
}

func newTestClaude(t *testing.T, api *fakeAPI) *Claude {
	srv := httptest.NewServer(api)
	t.Cleanup(srv.Close)
	client := anthropic.NewClient(option.WithAPIKey("test"), option.WithBaseURL(srv.URL), option.WithMaxRetries(0))
	return NewClaude(client, "claude-opus-5-5", testTools(t), slog.New(slog.DiscardHandler))
}

func TestAgentLoopRunsToolsAndStreams(t *testing.T) {
	api := &fakeAPI{scripts: []string{
		toolTurn(`{"brand":"HERMES","max_price_twd":55000,"sort":"deal"}`),
		textTurn("推薦這個 [[brandoff:1003]]"),
	}}
	c := newTestClaude(t, api)

	var events []Event
	if err := c.Chat(context.Background(), "s1", "愛馬仕 6 萬以內", func(e Event) { events = append(events, e) }); err != nil {
		t.Fatal(err)
	}

	var types []string
	for _, e := range events {
		types = append(types, e.Type)
	}
	if got := strings.Join(types, ","); got != "tool_call,tool_result,listings,text" {
		t.Fatalf("event sequence = %s", got)
	}
	if l := events[2].Listings; len(l) != 1 || l[0].ID != "brandoff:1003" || l[0].Fair == nil {
		t.Errorf("listings event = %+v", l)
	}

	if len(api.requests) != 2 {
		t.Fatalf("API called %d times, want 2", len(api.requests))
	}
	first := api.requests[0]
	if first["model"] != "claude-opus-5-5" || first["fallbacks"] != "default" {
		t.Errorf("model/fallbacks = %v / %v", first["model"], first["fallbacks"])
	}
	if !strings.Contains(api.headers[0].Get("Anthropic-Beta"), "server-side-fallback-2026-07-01") {
		t.Errorf("beta header = %q", api.headers[0].Get("Anthropic-Beta"))
	}
	if sys := first["system"].([]any)[0].(map[string]any); sys["cache_control"] == nil {
		t.Error("system prompt is not marked for caching")
	}
	if tool := first["tools"].([]any)[0].(map[string]any); tool["eager_input_streaming"] != true {
		t.Errorf("tool def = %v", tool)
	}

	// Second request: user, assistant(tool_use), user(tool_result).
	msgs := api.requests[1]["messages"].([]any)
	if len(msgs) != 3 {
		t.Fatalf("second request has %d messages", len(msgs))
	}
	res := msgs[2].(map[string]any)["content"].([]any)[0].(map[string]any)
	if res["type"] != "tool_result" || res["tool_use_id"] != "toolu_1" || res["is_error"] == true {
		t.Errorf("tool_result = %v", res)
	}
	if !strings.Contains(fmt.Sprint(res["content"]), "brandoff:1003") {
		t.Errorf("tool_result content lacks the matching listing: %v", res["content"])
	}

	// History persists for the next turn: 4 messages after a completed turn.
	if n := len(c.sessions.m["s1"].messages); n != 4 {
		t.Errorf("session history = %d messages, want 4", n)
	}
}

func TestAgentReturnsToolErrorsToModel(t *testing.T) {
	api := &fakeAPI{scripts: []string{
		toolTurn(`{"brand":"NOT A BRAND"}`),
		textTurn("抱歉，請問是哪個品牌？"),
	}}
	c := newTestClaude(t, api)
	var sawErr bool
	err := c.Chat(context.Background(), "s", "x", func(e Event) { sawErr = sawErr || (e.Type == "tool_result" && e.IsError) })
	if err != nil || !sawErr {
		t.Fatalf("err=%v sawErr=%v", err, sawErr)
	}
	res := api.requests[1]["messages"].([]any)[2].(map[string]any)["content"].([]any)[0].(map[string]any)
	if res["is_error"] != true || !strings.Contains(fmt.Sprint(res["content"]), "unknown brand") {
		t.Errorf("tool_result = %v", res)
	}
}

func TestAgentRollsBackFailedTurn(t *testing.T) {
	api := &fakeAPI{} // every call fails with 400
	c := newTestClaude(t, api)
	if err := c.Chat(context.Background(), "s", "hello", func(Event) {}); err == nil {
		t.Fatal("expected an error")
	}
	if n := len(c.sessions.m["s"].messages); n != 0 {
		t.Errorf("history after failed turn = %d messages, want 0", n)
	}
}

func TestPromptStatesThresholds(t *testing.T) {
	listing.RegisterScheme(listing.Scheme{Source: "ax", Shop: "店甲"})
	listing.RegisterScheme(listing.Scheme{Source: "ay", Shop: "店乙"})
	p := buildSystemPrompt()
	for _, shop := range []string{"店甲", "店乙"} {
		if !strings.Contains(p, shop) {
			t.Errorf("system prompt does not name registered shop %s", shop)
		}
	}
	if want := "低超過 " + pctText(radar.AnomalyPct) + "%"; !strings.Contains(p, want) {
		t.Errorf("system prompt lacks %q", want)
	}
	if strings.Contains(p, "{{") {
		t.Error("system prompt has an unreplaced placeholder")
	}
	var desc string
	for _, d := range toolDefs() {
		if d.OfTool != nil && d.OfTool.Name == "search_listings" {
			props, _ := d.OfTool.InputSchema.Properties.(map[string]any)
			od, _ := props["only_deals"].(map[string]any)
			desc, _ = od["description"].(string)
		}
	}
	if want := "at least " + pctText(radar.DealMinPct) + "%"; !strings.Contains(desc, want) {
		t.Errorf("only_deals description %q lacks %q", desc, want)
	}
}
