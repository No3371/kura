package advisor

import (
	"context"
	"fmt"
	"log/slog"
	"net/http/httptest"
	"strings"
	"testing"
)

// oaSSE renders Chat Completions stream chunks, ending with [DONE].
func oaSSE(chunks ...string) string {
	var b strings.Builder
	for _, c := range chunks {
		fmt.Fprintf(&b, "data: %s\n\n", c)
	}
	b.WriteString("data: [DONE]\n\n")
	return b.String()
}

// oaToolTurn streams one search_listings call with its arguments split in
// two fragments, as the API does.
func oaToolTurn(args string) string {
	half := len(args) / 2
	return oaSSE(
		`{"choices":[{"index":0,"delta":{"role":"assistant","tool_calls":[{"index":0,"id":"call_1","type":"function","function":{"name":"search_listings","arguments":""}}]}}]}`,
		`{"choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"function":{"arguments":`+jsonString(args[:half])+`}}]}}]}`,
		`{"choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"function":{"arguments":`+jsonString(args[half:])+`}}]}}]}`,
		`{"choices":[{"index":0,"delta":{},"finish_reason":"tool_calls"}]}`,
		`{"choices":[],"usage":{"prompt_tokens":100,"completion_tokens":20}}`,
	)
}

func oaTextTurn(parts ...string) string {
	var chunks []string
	for _, p := range parts {
		chunks = append(chunks, `{"choices":[{"index":0,"delta":{"content":`+jsonString(p)+`}}]}`)
	}
	chunks = append(chunks, `{"choices":[{"index":0,"delta":{},"finish_reason":"stop"}]}`)
	return oaSSE(chunks...)
}

func newTestOpenAI(t *testing.T, api *fakeAPI) *OpenAI {
	srv := httptest.NewServer(api)
	t.Cleanup(srv.Close)
	return NewOpenAI("test-key", srv.URL+"/v1", "gpt-test", "", testTools(t), slog.New(slog.DiscardHandler))
}

func TestOpenAILoopRunsToolsAndStreams(t *testing.T) {
	api := &fakeAPI{scripts: []string{
		oaToolTurn(`{"brand":"HERMES","max_price_twd":55000,"sort":"deal"}`),
		oaTextTurn("推薦這個 ", "[[brandoff:1003]]"),
	}}
	o := newTestOpenAI(t, api)
	if o.Mode() != "openai:gpt-test" {
		t.Errorf("mode = %q", o.Mode())
	}

	var events []Event
	if err := o.Chat(context.Background(), "s1", "愛馬仕 6 萬以內", func(e Event) { events = append(events, e) }); err != nil {
		t.Fatal(err)
	}
	var types []string
	for _, e := range events {
		types = append(types, e.Type)
	}
	if got := strings.Join(types, ","); got != "tool_call,tool_result,listings,text,text" {
		t.Fatalf("event sequence = %s", got)
	}
	if in := events[0].Input; in != `{"brand":"HERMES","max_price_twd":55000,"sort":"deal"}` {
		t.Errorf("reassembled arguments = %s", in)
	}
	if l := events[2].Listings; len(l) != 1 || l[0].ID != "brandoff:1003" || l[0].Fair == nil {
		t.Errorf("listings event = %+v", l)
	}

	if len(api.requests) != 2 {
		t.Fatalf("API called %d times, want 2", len(api.requests))
	}
	if h := api.headers[0].Get("Authorization"); h != "Bearer test-key" {
		t.Errorf("auth header = %q", h)
	}
	first := api.requests[0]
	if first["model"] != "gpt-test" || first["stream"] != true {
		t.Errorf("model/stream = %v / %v", first["model"], first["stream"])
	}
	if _, ok := first["reasoning_effort"]; ok {
		t.Error("reasoning_effort sent although Effort is unset")
	}
	if sys := first["messages"].([]any)[0].(map[string]any); sys["role"] != "system" || sys["content"] != buildSystemPrompt() {
		t.Errorf("first message = %v", sys)
	}
	tools := first["tools"].([]any)
	if len(tools) != len(toolSpecs()) {
		t.Errorf("%d tools, want %d", len(tools), len(toolSpecs()))
	}
	if fn := tools[0].(map[string]any)["function"].(map[string]any); fn["name"] != "search_listings" ||
		fn["parameters"].(map[string]any)["type"] != "object" {
		t.Errorf("tool def = %v", fn)
	}

	// Second request: system, user, assistant(tool_calls), tool.
	msgs := api.requests[1]["messages"].([]any)
	if len(msgs) != 4 {
		t.Fatalf("second request has %d messages", len(msgs))
	}
	call := msgs[2].(map[string]any)["tool_calls"].([]any)[0].(map[string]any)
	if call["id"] != "call_1" || call["type"] != "function" {
		t.Errorf("assistant tool call = %v", call)
	}
	res := msgs[3].(map[string]any)
	if res["role"] != "tool" || res["tool_call_id"] != "call_1" || !strings.Contains(fmt.Sprint(res["content"]), "brandoff:1003") {
		t.Errorf("tool message = %v", res)
	}

	// History (without the system prompt): user, assistant, tool, assistant.
	if n := len(o.sessions.m["s1"].messages); n != 4 {
		t.Errorf("session history = %d messages, want 4", n)
	}
}

func TestOpenAIReturnsToolErrorsToModel(t *testing.T) {
	api := &fakeAPI{scripts: []string{
		oaToolTurn(`{"brand":"NOT A BRAND"}`),
		oaTextTurn("抱歉，請問是哪個品牌？"),
	}}
	o := newTestOpenAI(t, api)
	var sawErr bool
	err := o.Chat(context.Background(), "s", "x", func(e Event) { sawErr = sawErr || (e.Type == "tool_result" && e.IsError) })
	if err != nil || !sawErr {
		t.Fatalf("err=%v sawErr=%v", err, sawErr)
	}
	res := api.requests[1]["messages"].([]any)[3].(map[string]any)
	if !strings.Contains(fmt.Sprint(res["content"]), "unknown brand") {
		t.Errorf("tool message = %v", res)
	}
}

func TestOpenAIRollsBackFailedTurn(t *testing.T) {
	api := &fakeAPI{} // every call fails with 400
	o := newTestOpenAI(t, api)
	err := o.Chat(context.Background(), "s", "hello", func(Event) {})
	if err == nil || !strings.Contains(err.Error(), "400") {
		t.Fatalf("err = %v, want the API's 400", err)
	}
	if n := len(o.sessions.m["s"].messages); n != 0 {
		t.Errorf("history after failed turn = %d messages, want 0", n)
	}
}

func TestOpenAITruncatedToolCallNotRun(t *testing.T) {
	api := &fakeAPI{scripts: []string{oaSSE(
		`{"choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"id":"call_1","type":"function","function":{"name":"search_listings","arguments":"{\"brand\":\"HER"}}]}}]}`,
		`{"choices":[{"index":0,"delta":{},"finish_reason":"length"}]}`,
	)}}
	o := newTestOpenAI(t, api)
	var ran bool
	err := o.Chat(context.Background(), "s", "x", func(e Event) { ran = ran || e.Type == "tool_call" })
	if err == nil || ran {
		t.Fatalf("err=%v ran=%v: a truncated call must not run", err, ran)
	}
}
