package advisor

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"

	"github.com/No3371/kura/internal/judge"
)

// fakeJev stands in for the Jev API: respond picks the status and body for
// each request, and every request's state is recorded. No test reaches the
// real API.
type fakeJev struct {
	mu      sync.Mutex
	states  []map[string]any
	delay   time.Duration
	respond func(state map[string]any) (int, string)
}

func (f *fakeJev) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req struct {
		State map[string]any `json:"state"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	f.mu.Lock()
	f.states = append(f.states, req.State)
	respond := f.respond
	f.mu.Unlock()
	if f.delay > 0 {
		select {
		case <-time.After(f.delay):
		case <-r.Context().Done():
			return
		}
	}
	status, body := respond(req.State)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	io.WriteString(w, body)
}

func (f *fakeJev) calls() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.states)
}

// jevSays answers every request with one verdict.
func jevSays(intent string, conf float64, abuse map[string]float64) func(map[string]any) (int, string) {
	b, _ := json.Marshal(jevResult(intent, conf, abuse))
	return func(map[string]any) (int, string) { return 200, string(b) }
}

// syncBuffer is a log sink safe for the handler's concurrent writes.
type syncBuffer struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (s *syncBuffer) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.Write(p)
}

func (s *syncBuffer) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.String()
}

func newGate(t *testing.T, inner Advisor, jev *fakeJev, mode GateMode) (*Gated, *syncBuffer) {
	t.Helper()
	srv := httptest.NewServer(jev)
	t.Cleanup(srv.Close)
	logs := &syncBuffer{}
	j := &judge.Client{Key: "test", Model: "jev-1.14.0", BaseURL: srv.URL, HTTP: srv.Client()}
	return NewGated(inner, j, mode, 2*time.Second, slog.New(slog.NewTextHandler(logs, nil))), logs
}

// scriptAdvisor emits a fixed event list, then returns err (or ctx's error
// if the turn was cancelled meanwhile).
type scriptAdvisor struct {
	events []Event
	err    error
	calls  atomic.Int32
}

func (a *scriptAdvisor) Mode() string { return "script" }

func (a *scriptAdvisor) Chat(ctx context.Context, _, _ string, emit func(Event)) error {
	a.calls.Add(1)
	for _, e := range a.events {
		emit(e)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return a.err
}

var innerEvents = []Event{
	{Type: "tool_call", Tool: "search_listings", Input: "{}"},
	{Type: "tool_result", Tool: "search_listings", Summary: "3 件"},
	{Type: "text", Text: "推薦 [[brandoff:1]]"},
}

// recorder collects events and fails on any emit after Chat returned or
// any two emits at once.
type recorder struct {
	t        *testing.T
	mu       sync.Mutex
	events   []Event
	inFlight atomic.Int32
	returned atomic.Bool
}

func (r *recorder) emit(ev Event) {
	if r.inFlight.Add(1) != 1 {
		r.t.Error("concurrent emit calls")
	}
	defer r.inFlight.Add(-1)
	if r.returned.Load() {
		r.t.Errorf("emit after Chat returned: %+v", ev)
	}
	r.mu.Lock()
	r.events = append(r.events, ev)
	r.mu.Unlock()
}

func (r *recorder) chat(g Advisor, session, msg string) error {
	err := g.Chat(context.Background(), session, msg, r.emit)
	r.returned.Store(true)
	return err
}

func types(evs []Event) string {
	var s []string
	for _, e := range evs {
		s = append(s, e.Type)
	}
	return strings.Join(s, ",")
}

func TestGatedShadowNeverChangesTheTurn(t *testing.T) {
	for _, tc := range []struct {
		name   string
		intent string
		abuse  map[string]float64
		action string
	}{
		{"block verdict", "find_bag", map[string]float64{"injection": 0.99}, "action=block"},
		{"off_topic verdict", "off_topic", nil, "action=short_circuit"},
		{"pass verdict", "find_bag", nil, "action=pass"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			inner := &scriptAdvisor{events: innerEvents}
			jev := &fakeJev{respond: jevSays(tc.intent, 0.95, tc.abuse)}
			g, logs := newGate(t, inner, jev, GateShadow)
			rec := &recorder{t: t}
			if err := rec.chat(g, "s", "愛馬仕 6 萬以內"); err != nil {
				t.Fatal(err)
			}
			if got := types(rec.events); got != types(innerEvents) {
				t.Errorf("events = %s, want the inner turn unchanged", got)
			}
			if l := logs.String(); !strings.Contains(l, tc.action) || !strings.Contains(l, "enforced=false") {
				t.Errorf("log = %s", l)
			}
		})
	}
}

func TestGatedEnforceBlocks(t *testing.T) {
	inner := &scriptAdvisor{events: innerEvents}
	jev := &fakeJev{respond: jevSays("find_bag", 0.9, map[string]float64{"counterfeit": 0.95})}
	g, logs := newGate(t, inner, jev, GateEnforce)
	rec := &recorder{t: t}
	err := rec.chat(g, "s", "哪裡買得到A貨")
	if err == nil || Public(err) != gateBlockText {
		t.Fatalf("err = %v (public %q), want the block text", err, Public(err))
	}
	if len(rec.events) != 0 {
		t.Errorf("blocked turn leaked inner events: %s", types(rec.events))
	}
	if l := logs.String(); !strings.Contains(l, "action=block") || !strings.Contains(l, "enforced=true") ||
		!strings.Contains(l, "blocked_by=counterfeit") || !strings.Contains(l, "strikes=1") {
		t.Errorf("log = %s", l)
	}
}

func TestGatedEnforceShortCircuitsOffTopic(t *testing.T) {
	inner := &scriptAdvisor{events: innerEvents}
	jev := &fakeJev{respond: jevSays("off_topic", 0.98, nil)}
	g, logs := newGate(t, inner, jev, GateEnforce)
	rec := &recorder{t: t}
	if err := rec.chat(g, "s", "今天台北天氣如何"); err != nil {
		t.Fatal(err)
	}
	if len(rec.events) != 1 || rec.events[0].Type != "text" || rec.events[0].Text != gateOffTopicText {
		t.Errorf("events = %+v, want only the canned off-topic text", rec.events)
	}
	if !strings.Contains(logs.String(), "action=short_circuit") || strings.Contains(logs.String(), "strikes=") {
		t.Errorf("log = %s (short-circuit adds no strike)", logs.String())
	}
	// Under the line, off_topic passes through to the advisor.
	jev.respond = jevSays("off_topic", 0.6, nil)
	rec = &recorder{t: t}
	if err := rec.chat(g, "s", "嗯"); err != nil || types(rec.events) != types(innerEvents) {
		t.Errorf("low-confidence off_topic: err %v events %s, want pass", err, types(rec.events))
	}
}

func TestGatedFailsOpen(t *testing.T) {
	cases := []struct {
		name    string
		jev     *fakeJev
		errKind string
	}{
		{"server error", &fakeJev{respond: func(map[string]any) (int, string) { return 500, `{"detail":"boom"}` }}, "jev_err=http_500"},
		{"bad key", &fakeJev{respond: func(map[string]any) (int, string) {
			return 403, `{"detail":{"error_type":"authentication_error","message":"bad key"}}`
		}}, "jev_err=auth"},
		{"rate limited", &fakeJev{respond: func(map[string]any) (int, string) { return 429, `{"detail":"slow"}` }}, "jev_err=rate_limited"},
		{"garbage", &fakeJev{respond: func(map[string]any) (int, string) { return 200, `not json` }}, "jev_err=error"},
		{"partial answers", &fakeJev{respond: func(map[string]any) (int, string) {
			return 200, `{"model":"jev-1.14.0","answers":{"intent":{"type":"choice","choice":"find_bag","confidence":0.9}}}`
		}}, "jev_err=error"},
		{"timeout", &fakeJev{delay: 2 * time.Second, respond: jevSays("find_bag", 0.9, map[string]float64{"injection": 0.99})}, "jev_err=timeout"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			inner := &scriptAdvisor{events: innerEvents}
			g, logs := newGate(t, inner, tc.jev, GateEnforce)
			g.Budget = 50 * time.Millisecond
			rec := &recorder{t: t}
			start := time.Now()
			if err := rec.chat(g, "s", "愛馬仕"); err != nil {
				t.Fatal(err)
			}
			if took := time.Since(start); took > time.Second {
				t.Errorf("turn took %v with a 50ms budget", took)
			}
			if types(rec.events) != types(innerEvents) {
				t.Errorf("events = %s, want the inner turn", types(rec.events))
			}
			if l := logs.String(); !strings.Contains(l, "action=fail_open") || !strings.Contains(l, tc.errKind) {
				t.Errorf("log = %s, want fail_open + %s", l, tc.errKind)
			}
		})
	}
}

func TestGatedPassesInnerError(t *testing.T) {
	inner := &scriptAdvisor{events: innerEvents[:1], err: shopperErr("顧問暫時無法使用，請稍後再試", nil)}
	g, _ := newGate(t, inner, &fakeJev{respond: jevSays("find_bag", 0.9, nil)}, GateEnforce)
	rec := &recorder{t: t}
	if err := rec.chat(g, "s", "愛馬仕"); err == nil || Public(err) != "顧問暫時無法使用，請稍後再試" {
		t.Errorf("err = %v, want the inner error unchanged", err)
	}
}

// burstAdvisor emits many events as fast as it can.
type burstAdvisor struct{ n int }

func (b *burstAdvisor) Mode() string { return "burst" }
func (b *burstAdvisor) Chat(_ context.Context, _, _ string, emit func(Event)) error {
	for i := range b.n {
		emit(Event{Type: "text", Text: fmt.Sprint(i)})
	}
	return nil
}

func TestGatedKeepsEventOrder(t *testing.T) {
	for _, mode := range []GateMode{GateShadow, GateEnforce} {
		jev := &fakeJev{delay: 30 * time.Millisecond, respond: jevSays("find_bag", 0.9, nil)}
		g, _ := newGate(t, &burstAdvisor{n: 200}, jev, mode)
		for turn := range 3 {
			rec := &recorder{t: t}
			if err := rec.chat(g, fmt.Sprint("s", turn), "愛馬仕"); err != nil {
				t.Fatal(err)
			}
			if len(rec.events) != 200 {
				t.Fatalf("%v: %d events, want 200", mode, len(rec.events))
			}
			for i, e := range rec.events {
				if e.Text != fmt.Sprint(i) {
					t.Fatalf("%v: event %d = %q, out of order", mode, i, e.Text)
				}
			}
		}
	}
}

func TestGatedConcurrentSessions(t *testing.T) {
	jev := &fakeJev{respond: func(st map[string]any) (int, string) {
		abuse := map[string]float64{}
		if strings.Contains(st["message"].(string), "bad") {
			abuse["injection"] = 0.99
		}
		b, _ := json.Marshal(jevResult("find_bag", 0.9, abuse))
		return 200, string(b)
	}}
	g, _ := newGate(t, &burstAdvisor{n: 20}, jev, GateEnforce)
	var wg sync.WaitGroup
	for i := range 16 {
		wg.Go(func() {
			msg := "good"
			if i%2 == 1 {
				msg = "bad"
			}
			rec := &recorder{t: t}
			err := rec.chat(g, fmt.Sprint("s", i), msg)
			if msg == "bad" && (err == nil || len(rec.events) != 0) {
				t.Errorf("session %d: bad turn passed (err %v, %d events)", i, err, len(rec.events))
			}
			if msg == "good" && (err != nil || len(rec.events) != 20) {
				t.Errorf("session %d: good turn failed (err %v, %d events)", i, err, len(rec.events))
			}
		})
	}
	wg.Wait()
}

func TestGatedBlockHook(t *testing.T) {
	abusive := jevSays("find_bag", 0.9, map[string]float64{"injection": 0.99})
	for _, tc := range []struct {
		name    string
		mode    GateMode
		respond func(map[string]any) (int, string)
		want    int32
	}{
		{"enforce block", GateEnforce, abusive, 1},
		{"shadow block", GateShadow, abusive, 0},
		{"enforce pass", GateEnforce, jevSays("find_bag", 0.9, nil), 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, _ := newGate(t, &scriptAdvisor{events: innerEvents}, &fakeJev{respond: tc.respond}, tc.mode)
			var hits atomic.Int32
			ctx := WithBlockHook(context.Background(), func() time.Duration { hits.Add(1); return 0 })
			err := g.Chat(ctx, "s", "x", func(Event) {})
			if hits.Load() != tc.want {
				t.Errorf("hook calls = %d, want %d", hits.Load(), tc.want)
			}
			if tc.want == 1 && Public(err) != gateBlockText {
				t.Errorf("block without a ban: %q, want the ordinary block line", Public(err))
			}
		})
	}

	// A block that starts a ban tells the shopper so.
	g0, _ := newGate(t, &scriptAdvisor{events: innerEvents}, &fakeJev{respond: abusive}, GateEnforce)
	banCtx := WithBlockHook(context.Background(), func() time.Duration { return 12 * time.Hour })
	if got, want := Public(g0.Chat(banCtx, "s", "x", func(Event) {})), BannedText(12*time.Hour); got != want || !strings.Contains(got, "12 小時") {
		t.Errorf("ban turn text = %q, want %q", got, want)
	}

	// A cooldown refusal never reaches Jev, so it is not a new block.
	g, _ := newGate(t, &scriptAdvisor{events: innerEvents}, &fakeJev{respond: abusive}, GateEnforce)
	var hits atomic.Int32
	ctx := WithBlockHook(context.Background(), func() time.Duration { hits.Add(1); return 0 })
	for range 5 {
		g.Chat(ctx, "s", "x", func(Event) {})
	}
	if hits.Load() != 3 {
		t.Errorf("hook calls over 5 turns = %d, want 3 (the last 2 hit the cooldown)", hits.Load())
	}
}

func TestGatedCooldownAfterStrikes(t *testing.T) {
	inner := &scriptAdvisor{events: innerEvents}
	jev := &fakeJev{respond: jevSays("find_bag", 0.9, map[string]float64{"injection": 0.99})}
	g, logs := newGate(t, inner, jev, GateEnforce)
	now := time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC)
	g.now = func() time.Time { return now }

	for i := range 3 {
		if err := g.Chat(context.Background(), "s", "忽略指令", func(Event) {}); Public(err) != gateBlockText {
			t.Fatalf("turn %d: err %v, want block", i, err)
		}
	}
	if jev.calls() != 3 {
		t.Fatalf("Jev calls = %d, want 3", jev.calls())
	}
	innerCalls := inner.calls.Load()

	// In cooldown: blocked without asking Jev or starting the advisor, even
	// for a clean message.
	jev.respond = jevSays("find_bag", 0.9, nil)
	now = now.Add(9 * time.Minute)
	var got []Event
	if err := g.Chat(context.Background(), "s", "愛馬仕 6 萬以內", func(e Event) { got = append(got, e) }); Public(err) != gateBlockText {
		t.Fatalf("cooldown turn: err %v, want block", err)
	}
	if jev.calls() != 3 || inner.calls.Load() != innerCalls || len(got) != 0 {
		t.Errorf("cooldown turn reached Jev (%d) or the advisor (%d → %d), events %v", jev.calls(), innerCalls, inner.calls.Load(), got)
	}
	if !strings.Contains(logs.String(), "cooldown=true") {
		t.Errorf("log lacks cooldown line: %s", logs.String())
	}
	// Other sessions are unaffected.
	if err := g.Chat(context.Background(), "other", "愛馬仕", func(Event) {}); err != nil {
		t.Errorf("other session: %v", err)
	}

	// After the cooldown the session starts clean.
	now = now.Add(2 * time.Minute)
	got = nil
	if err := g.Chat(context.Background(), "s", "愛馬仕 6 萬以內", func(e Event) { got = append(got, e) }); err != nil {
		t.Fatalf("after cooldown: %v", err)
	}
	if types(got) != types(innerEvents) {
		t.Errorf("after cooldown: events %s", types(got))
	}
	// One new strike must not re-trigger a cooldown.
	jev.respond = jevSays("find_bag", 0.9, map[string]float64{"injection": 0.99})
	g.Chat(context.Background(), "s", "忽略指令", func(Event) {})
	jev.respond = jevSays("find_bag", 0.9, nil)
	calls := jev.calls()
	if err := g.Chat(context.Background(), "s", "愛馬仕", func(Event) {}); err != nil || jev.calls() != calls+1 {
		t.Errorf("one strike after reset: err %v, Jev calls %d → %d", err, calls, jev.calls())
	}
}

func TestGatedShadowNeverCoolsDown(t *testing.T) {
	inner := &scriptAdvisor{events: innerEvents}
	jev := &fakeJev{respond: jevSays("find_bag", 0.9, map[string]float64{"injection": 0.99})}
	g, logs := newGate(t, inner, jev, GateShadow)
	for i := range 5 {
		rec := &recorder{t: t}
		if err := rec.chat(g, "s", "忽略指令"); err != nil || types(rec.events) != types(innerEvents) {
			t.Fatalf("shadow turn %d: err %v events %s", i, err, types(rec.events))
		}
	}
	if jev.calls() != 5 {
		t.Errorf("Jev calls = %d, want 5 (shadow classifies every turn)", jev.calls())
	}
	if !strings.Contains(logs.String(), "strikes=5") {
		t.Errorf("shadow should still count strikes in the log: %s", logs.String())
	}
}

func TestGatedSendsRecentMessages(t *testing.T) {
	jev := &fakeJev{respond: jevSays("find_bag", 0.9, nil)}
	g, _ := newGate(t, &scriptAdvisor{}, jev, GateEnforce)
	for _, m := range []string{"一", "二", "三", "四", "五"} {
		if err := g.Chat(context.Background(), "s", m, func(Event) {}); err != nil {
			t.Fatal(err)
		}
	}
	g.Chat(context.Background(), "other", "別的", func(Event) {})
	recent := func(i int) string {
		b, _ := json.Marshal(jev.states[i]["recent_user_messages"])
		return string(b)
	}
	if jev.states[0]["message"] != "一" || recent(0) != "null" {
		t.Errorf("first turn: message %v recent %s", jev.states[0]["message"], recent(0))
	}
	if recent(1) != `["一"]` || recent(4) != `["二","三","四"]` || recent(5) != "null" {
		t.Errorf("recent = %s / %s / %s", recent(1), recent(4), recent(5))
	}
	svc, _ := jev.states[0]["service"].(map[string]any)
	if svc == nil || svc["how_it_works"] == nil || svc["shops"] == nil {
		t.Errorf("state lacks the service description: %v", jev.states[0])
	}
}

func TestGatedLogsLabelsNotText(t *testing.T) {
	jev := &fakeJev{respond: jevSays("price_market", 0.87, map[string]float64{"free_llm": 0.6})}
	g, logs := newGate(t, &scriptAdvisor{events: innerEvents}, jev, GateShadow)
	const secret = "我的祕密訊息內容"
	if err := g.Chat(context.Background(), "s", secret, func(Event) {}); err != nil {
		t.Fatal(err)
	}
	l := logs.String()
	if strings.Contains(l, secret) {
		t.Errorf("log carries the message text: %s", l)
	}
	for _, want := range []string{"msg=\"advisor gate\"", "mode=shadow", "action=pass", "intent=price_market", "intent_conf=0.87",
		"flagged=true", "p_free_llm=0.6", "p_injection=0.02", "model=jev-1.14.0", "jev_ms="} {
		if !strings.Contains(l, want) {
			t.Errorf("log lacks %s: %s", want, l)
		}
	}
}

func TestGatedMode(t *testing.T) {
	g := NewGated(&Mock{}, &judge.Client{BaseURL: "http://127.0.0.1:1"}, GateShadow, 0, slog.New(slog.DiscardHandler))
	if g.Mode() != "mock+jev:shadow" {
		t.Errorf("Mode = %q", g.Mode())
	}
	if BaseMode(g) != "mock" || BaseMode(&Mock{}) != "mock" {
		t.Errorf("BaseMode = %q, want the inner mode", BaseMode(g))
	}
}

// stallAPI is a Claude API fake whose stall-marked requests stream the
// start of a text reply and then hang until the client goes away, so a turn
// is still in flight when the gate decides.
type stallAPI struct {
	mu       sync.Mutex
	scripts  []string // "" = stall
	requests []map[string]any
}

func (s *stallAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	var req map[string]any
	json.Unmarshal(body, &req)
	s.mu.Lock()
	i := len(s.requests)
	s.requests = append(s.requests, req)
	s.mu.Unlock()
	w.Header().Set("Content-Type", "text/event-stream")
	if i >= len(s.scripts) {
		http.Error(w, `{"type":"error","error":{"type":"invalid_request_error","message":"unexpected call"}}`, 400)
		return
	}
	if s.scripts[i] != "" {
		io.WriteString(w, s.scripts[i])
		return
	}
	io.WriteString(w, sse(fmt.Sprintf(msgStart, 9),
		`{"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}`,
		`{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"這是"}}`))
	w.(http.Flusher).Flush()
	<-r.Context().Done()
}

// overrideBackend is one real backend behind the gate, with views of its
// session history and of the history it sent with its second request.
type overrideBackend struct {
	inner   Advisor
	history func() int
	sent    func() []any
}

func claudeBackend(t *testing.T, first string) overrideBackend {
	api := &stallAPI{scripts: []string{first, textTurn("好的")}}
	srv := httptest.NewServer(api)
	t.Cleanup(srv.Close)
	client := anthropic.NewClient(option.WithAPIKey("test"), option.WithBaseURL(srv.URL), option.WithMaxRetries(0))
	c := NewClaude(client, "claude-opus-5-5", testTools(t), slog.New(slog.DiscardHandler))
	return overrideBackend{c,
		func() int { return len(c.sessions.m["s"].messages) },
		func() []any {
			api.mu.Lock()
			defer api.mu.Unlock()
			msgs, _ := api.requests[1]["messages"].([]any)
			return msgs
		}}
}

func openAIBackend(t *testing.T) overrideBackend {
	api := &fakeAPI{scripts: []string{oaTextTurn("這是", "一則完整回覆"), oaTextTurn("好的")}}
	o := newTestOpenAI(t, api, "")
	return overrideBackend{o,
		func() int { return len(o.sessions.m["s"].messages) },
		func() []any {
			api.mu.Lock()
			defer api.mu.Unlock()
			in, _ := api.requests[1]["input"].([]any)
			return in
		}}
}

// A blocked or short-circuited turn must leave no backend history, both when
// it is still streaming at the verdict and when the backend already holds the
// whole reply: cancelling the latter does not stop it finishing, so the
// backend has to treat the cancelled turn as failed.
func TestGatedOverrideLeavesNoHistory(t *testing.T) {
	backends := []struct {
		name string
		make func(*testing.T) overrideBackend
	}{
		{"claude_streaming", func(t *testing.T) overrideBackend { return claudeBackend(t, "") }},
		{"claude_full_reply", func(t *testing.T) overrideBackend { return claudeBackend(t, textTurn("這是一則完整回覆")) }},
		{"openai_full_reply", openAIBackend},
	}
	verdicts := []struct {
		name   string
		intent string
		abuse  map[string]float64
	}{
		{"block", "find_bag", map[string]float64{"injection": 0.99}},
		{"short_circuit", "off_topic", nil},
	}
	for _, be := range backends {
		for _, tc := range verdicts {
			t.Run(be.name+"/"+tc.name, func(t *testing.T) {
				b := be.make(t)
				jev := &fakeJev{delay: 50 * time.Millisecond, respond: jevSays(tc.intent, 0.95, tc.abuse)}
				g, _ := newGate(t, b.inner, jev, GateEnforce)

				rec := &recorder{t: t}
				rec.chat(g, "s", "第一則")
				for _, e := range rec.events {
					if e.Type == "text" && e.Text != gateOffTopicText {
						t.Errorf("the overridden turn reached the shopper: %q", e.Text)
					}
				}
				if n := b.history(); n != 0 {
					t.Fatalf("history after %s = %d messages, want 0", tc.name, n)
				}

				// The next turn passes and the backend sends only it.
				jev.respond = jevSays("find_bag", 0.95, nil)
				rec = &recorder{t: t}
				if err := rec.chat(g, "s", "第二則"); err != nil {
					t.Fatal(err)
				}
				if sent := b.sent(); len(sent) != 1 || !strings.Contains(fmt.Sprint(sent[0]), "第二則") {
					t.Errorf("second request history = %v, want only the second turn", sent)
				}
			})
		}
	}
}

func TestGateErrKind(t *testing.T) {
	cases := map[string]error{
		"timeout":      fmt.Errorf("jev: %w", context.DeadlineExceeded),
		"canceled":     context.Canceled,
		"auth":         &judge.APIError{Status: 403},
		"rate_limited": &judge.APIError{Status: 429},
		"http_503":     &judge.APIError{Status: 503},
		"error":        errors.New("jev: decode response"),
	}
	for want, err := range cases {
		if got := gateErrKind(err); got != want {
			t.Errorf("gateErrKind(%v) = %q, want %q", err, got, want)
		}
	}
}
