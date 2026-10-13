package advisor

import (
	"context"
	"fmt"
	"log/slog"
	"net/http/httptest"
	"strings"
	"testing"
)

const oaReasoning = `{"id":"rs_1","type":"reasoning","summary":[],"encrypted_content":"gAAA-opaque"}`

// oaCompleted ends a Responses stream with the given output items.
func oaCompleted(items ...string) string {
	return `{"type":"response.completed","response":{"id":"resp_1","status":"completed","output":[` + strings.Join(items, ",") +
		`],"usage":{"input_tokens":100,"output_tokens":20,"input_tokens_details":{"cached_tokens":64}}}}`
}

// oaToolTurn streams one search_listings call after a reasoning item, with
// its arguments split in two deltas, as the API does.
func oaToolTurn(args string) string {
	half := len(args) / 2
	call := `{"id":"fc_1","type":"function_call","status":"completed","call_id":"call_1","name":"search_listings","arguments":` + jsonString(args) + `}`
	return sse(
		`{"type":"response.created","response":{"id":"resp_1","status":"in_progress","output":[]}}`,
		`{"type":"response.output_item.added","output_index":1,"item":{"id":"fc_1","type":"function_call","call_id":"call_1","name":"search_listings","arguments":""}}`,
		`{"type":"response.function_call_arguments.delta","item_id":"fc_1","output_index":1,"delta":`+jsonString(args[:half])+`}`,
		`{"type":"response.function_call_arguments.delta","item_id":"fc_1","output_index":1,"delta":`+jsonString(args[half:])+`}`,
		oaCompleted(oaReasoning, call),
	)
}

func oaTextTurn(parts ...string) string {
	var events []string
	for _, p := range parts {
		events = append(events, `{"type":"response.output_text.delta","item_id":"msg_1","output_index":0,"content_index":0,"delta":`+jsonString(p)+`}`)
	}
	msg := `{"id":"msg_1","type":"message","role":"assistant","status":"completed","content":[{"type":"output_text","text":` +
		jsonString(strings.Join(parts, "")) + `,"annotations":[]}]}`
	events = append(events, oaCompleted(msg))
	return sse(events...)
}

func newTestOpenAI(t *testing.T, api *fakeAPI, effort string) *OpenAI {
	srv := httptest.NewServer(api)
	t.Cleanup(srv.Close)
	return NewOpenAI("test-key", srv.URL+"/v1", "gpt-test", effort, testTools(t), slog.New(slog.DiscardHandler))
}

func TestOpenAILoopRunsToolsAndStreams(t *testing.T) {
	api := &fakeAPI{scripts: []string{
		oaToolTurn(`{"brand":"HERMES","max_price_twd":55000,"sort":"deal"}`),
		oaTextTurn("推薦這個 ", "[[brandoff:1003]]"),
	}}
	o := newTestOpenAI(t, api, "")
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
		t.Errorf("tool arguments = %s", in)
	}
	if l := events[2].Listings; len(l) != 1 || l[0].ID != "brandoff:1003" || l[0].Fair == nil {
		t.Errorf("listings event = %+v", l)
	}

	if len(api.requests) != 2 {
		t.Fatalf("API called %d times, want 2", len(api.requests))
	}
	if api.paths[0] != "/v1/responses" {
		t.Errorf("path = %q, want /v1/responses", api.paths[0])
	}
	if h := api.headers[0].Get("Authorization"); h != "Bearer test-key" {
		t.Errorf("auth header = %q", h)
	}
	first := api.requests[0]
	if first["model"] != "gpt-test" || first["stream"] != true || first["store"] != false {
		t.Errorf("model/stream/store = %v / %v / %v", first["model"], first["stream"], first["store"])
	}
	if first["instructions"] != buildSystemPrompt() {
		t.Error("instructions are not the system prompt")
	}
	if inc := fmt.Sprint(first["include"]); inc != "[reasoning.encrypted_content]" {
		t.Errorf("include = %s", inc)
	}
	if _, ok := first["reasoning"]; ok {
		t.Error("reasoning sent although Effort is unset")
	}
	if in := first["input"].([]any); len(in) != 1 || in[0].(map[string]any)["role"] != "user" {
		t.Errorf("first input = %v", in)
	}
	tools := first["tools"].([]any)
	if len(tools) != len(toolSpecs()) {
		t.Errorf("%d tools, want %d", len(tools), len(toolSpecs()))
	}
	if fn := tools[0].(map[string]any); fn["type"] != "function" || fn["name"] != "search_listings" ||
		fn["parameters"].(map[string]any)["type"] != "object" {
		t.Errorf("tool def = %v", fn)
	}

	// Second request: user, reasoning and call replayed as output, then the result.
	in := api.requests[1]["input"].([]any)
	if len(in) != 4 {
		t.Fatalf("second request has %d input items", len(in))
	}
	if r := in[1].(map[string]any); r["type"] != "reasoning" || r["encrypted_content"] != "gAAA-opaque" {
		t.Errorf("reasoning item not replayed verbatim: %v", r)
	}
	if c := in[2].(map[string]any); c["type"] != "function_call" || c["call_id"] != "call_1" {
		t.Errorf("function call item = %v", c)
	}
	res := in[3].(map[string]any)
	if res["type"] != "function_call_output" || res["call_id"] != "call_1" || !strings.Contains(fmt.Sprint(res["output"]), "brandoff:1003") {
		t.Errorf("function call output = %v", res)
	}

	// History: user, reasoning, call, call output, assistant message.
	if n := len(o.sessions.m["s1"].messages); n != 5 {
		t.Errorf("session history = %d items, want 5", n)
	}
}

func TestOpenAISendsEffort(t *testing.T) {
	api := &fakeAPI{scripts: []string{oaTextTurn("好")}}
	o := newTestOpenAI(t, api, "medium")
	if err := o.Chat(context.Background(), "s", "x", func(Event) {}); err != nil {
		t.Fatal(err)
	}
	if r := api.requests[0]["reasoning"]; fmt.Sprint(r) != "map[effort:medium]" {
		t.Errorf("reasoning = %v", r)
	}
}

func TestOpenAIReturnsToolErrorsToModel(t *testing.T) {
	api := &fakeAPI{scripts: []string{
		oaToolTurn(`{"brand":"NOT A BRAND"}`),
		oaTextTurn("抱歉，請問是哪個品牌？"),
	}}
	o := newTestOpenAI(t, api, "")
	var sawErr bool
	err := o.Chat(context.Background(), "s", "x", func(e Event) { sawErr = sawErr || (e.Type == "tool_result" && e.IsError) })
	if err != nil || !sawErr {
		t.Fatalf("err=%v sawErr=%v", err, sawErr)
	}
	res := api.requests[1]["input"].([]any)[3].(map[string]any)
	if !strings.Contains(fmt.Sprint(res["output"]), "unknown brand") {
		t.Errorf("function call output = %v", res)
	}
}

func TestOpenAIRollsBackFailedTurn(t *testing.T) {
	api := &fakeAPI{} // every call fails with 400
	o := newTestOpenAI(t, api, "")
	err := o.Chat(context.Background(), "s", "hello", func(Event) {})
	if err == nil || !strings.Contains(err.Error(), "400") {
		t.Fatalf("err = %v, want the API's 400", err)
	}
	if n := len(o.sessions.m["s"].messages); n != 0 {
		t.Errorf("history after failed turn = %d items, want 0", n)
	}
}

func TestOpenAIFailedResponseRollsBack(t *testing.T) {
	api := &fakeAPI{scripts: []string{
		oaToolTurn(`{"brand":"HERMES"}`),
		sse(`{"type":"response.failed","response":{"id":"resp_2","status":"failed","output":[],"error":{"code":"server_error","message":"boom"}}}`),
	}}
	o := newTestOpenAI(t, api, "")
	err := o.Chat(context.Background(), "s", "x", func(Event) {})
	if err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("err = %v, want the failed response's message", err)
	}
	if Public(err) != "顧問暫時無法回覆，請稍後再試" {
		t.Errorf("shopper text = %q", Public(err))
	}
	if n := len(o.sessions.m["s"].messages); n != 0 {
		t.Errorf("history after failed turn = %d items, want 0", n)
	}
}

func TestOpenAITruncatedToolCallNotRun(t *testing.T) {
	api := &fakeAPI{scripts: []string{sse(
		`{"type":"response.function_call_arguments.delta","item_id":"fc_1","output_index":0,"delta":"{\"brand\":\"HER"}`,
		`{"type":"response.incomplete","response":{"id":"resp_1","status":"incomplete","incomplete_details":{"reason":"max_output_tokens"},`+
			`"output":[{"id":"fc_1","type":"function_call","status":"incomplete","call_id":"call_1","name":"search_listings","arguments":"{\"brand\":\"HER"}]}}`,
	)}}
	o := newTestOpenAI(t, api, "")
	var ran bool
	err := o.Chat(context.Background(), "s", "x", func(e Event) { ran = ran || e.Type == "tool_call" })
	if err == nil || ran {
		t.Fatalf("err=%v ran=%v: a truncated call must not run", err, ran)
	}
}

func TestOpenAIStreamWithoutCompletionFails(t *testing.T) {
	api := &fakeAPI{scripts: []string{sse(`{"type":"response.output_text.delta","delta":"半"}`)}}
	o := newTestOpenAI(t, api, "")
	if err := o.Chat(context.Background(), "s", "x", func(Event) {}); err == nil {
		t.Fatal("a stream cut before response.completed must fail the turn")
	}
}
