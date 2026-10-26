package web

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
	"sync/atomic"
	"testing"
	"time"

	"github.com/No3371/kura/internal/advisor"
	"github.com/No3371/kura/internal/fx"
	"github.com/No3371/kura/internal/judge"
	"github.com/No3371/kura/internal/radar"
	"github.com/No3371/kura/internal/store"
)

func testServer(t *testing.T, adv advisor.Advisor) *Server {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "kura.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	fxc := &fx.Converter{Fallback: 0.2}
	return &Server{Store: st, Radar: &radar.Engine{Store: st, FX: fxc}, FX: fxc, Advisor: adv,
		Log: slog.New(slog.NewTextHandler(io.Discard, nil))}
}

// The UI shows its demo banner for advisor == "mock", so wrapping the mock
// in the gate must not change what the overview reports.
func TestOverviewReportsInnerAdvisorMode(t *testing.T) {
	gated := advisor.NewGated(&advisor.Mock{}, &judge.Client{BaseURL: "http://127.0.0.1:1"}, advisor.GateShadow, 0, slog.New(slog.DiscardHandler))
	h := testServer(t, gated).Handler()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/overview", nil))
	var got struct {
		Advisor string `json:"advisor"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Advisor != "mock" {
		t.Errorf("overview advisor = %q, want mock", got.Advisor)
	}
}

// countingAdvisor answers every turn with one text event.
type countingAdvisor struct{ calls atomic.Int32 }

func (a *countingAdvisor) Mode() string { return "counting" }
func (a *countingAdvisor) Chat(_ context.Context, _, _ string, emit func(advisor.Event)) error {
	a.calls.Add(1)
	emit(advisor.Event{Type: "text", Text: "好"})
	return nil
}

func postChat(h http.Handler, remote, xff, message string) *httptest.ResponseRecorder {
	body, _ := json.Marshal(map[string]string{"session": "s", "message": message})
	req := httptest.NewRequest(http.MethodPost, "/api/chat", strings.NewReader(string(body)))
	req.RemoteAddr = remote
	if xff != "" {
		req.Header.Set("X-Forwarded-For", xff)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func errorText(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()
	var got struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("error body %q: %v", rec.Body, err)
	}
	return got.Error
}

func TestChatRateLimit(t *testing.T) {
	adv := &countingAdvisor{}
	s := testServer(t, adv)
	s.ChatLimit = NewChatLimiter(6, 3)
	h := s.Handler()
	for i := range 3 {
		if rec := postChat(h, "203.0.113.7:5000", "", "愛馬仕"); rec.Code != 200 {
			t.Fatalf("turn %d within burst: %d %s", i, rec.Code, rec.Body)
		}
	}
	rec := postChat(h, "203.0.113.7:5001", "", "愛馬仕") // same IP, new port
	if rec.Code != http.StatusTooManyRequests || errorText(t, rec) != "目前詢問的人較多，請稍後再試" {
		t.Fatalf("over the limit: %d %s", rec.Code, rec.Body)
	}
	if adv.calls.Load() != 3 {
		t.Errorf("advisor calls = %d, want 3 (a limited turn never reaches it)", adv.calls.Load())
	}
	if rec := postChat(h, "198.51.100.2:5000", "", "愛馬仕"); rec.Code != 200 {
		t.Errorf("another client: %d", rec.Code)
	}
}

func TestChatLimiterRefills(t *testing.T) {
	l := NewChatLimiter(6, 3)
	now := time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC)
	l.now = func() time.Time { return now }
	for i := range 3 {
		if !l.Allow("a") {
			t.Fatalf("burst turn %d refused", i)
		}
	}
	if l.Allow("a") {
		t.Fatal("4th turn in the same instant allowed")
	}
	now = now.Add(9 * time.Second)
	if l.Allow("a") {
		t.Error("allowed before one token refilled (6/min = one per 10s)")
	}
	now = now.Add(2 * time.Second)
	if !l.Allow("a") {
		t.Error("refused after a token refilled")
	}
	now = now.Add(time.Hour)
	for i := range 3 {
		if !l.Allow("a") {
			t.Errorf("after an idle hour, burst turn %d refused", i)
		}
	}
	if l.Allow("a") {
		t.Error("refill exceeded the burst")
	}
}

func TestChatLimiterSweepsIdleClients(t *testing.T) {
	l := NewChatLimiter(6, 3)
	now := time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC)
	l.now = func() time.Time { return now }
	for i := range sweepAt {
		l.Allow(fmt.Sprint("c", i))
	}
	now = now.Add(2 * time.Minute)
	l.Allow("new")
	if n := len(l.clients); n != 1 {
		t.Errorf("clients after sweep = %d, want 1", n)
	}
}

func TestChatClientIP(t *testing.T) {
	cases := []struct {
		trust  bool
		remote string
		xff    []string
		want   string
	}{
		{false, "203.0.113.7:5000", nil, "203.0.113.7"},
		{false, "203.0.113.7:5000", []string{"1.2.3.4"}, "203.0.113.7"}, // header ignored without a proxy
		{true, "10.0.0.1:5000", []string{"1.2.3.4"}, "1.2.3.4"},
		{true, "10.0.0.1:5000", []string{"9.9.9.9, 1.2.3.4"}, "1.2.3.4"}, // spoofed first hop ignored
		{true, "10.0.0.1:5000", []string{"9.9.9.9", "5.6.7.8"}, "5.6.7.8"},
		{true, "10.0.0.1:5000", nil, "10.0.0.1"},
		{true, "[2001:db8::1]:443", []string{" "}, "2001:db8::/64"},
		// IPv6 keys on the /64 a subscriber can rotate inside; IPv4 stays exact.
		{false, "[2001:db8:0:7:aaaa::1]:443", nil, "2001:db8:0:7::/64"},
		{true, "10.0.0.1:5000", []string{"2001:db8:0:7:bbbb::2"}, "2001:db8:0:7::/64"},
		{false, "[::ffff:203.0.113.7]:443", nil, "::ffff:203.0.113.7"},
	}
	for _, tc := range cases {
		r := httptest.NewRequest(http.MethodPost, "/api/chat", nil)
		r.RemoteAddr = tc.remote
		for _, v := range tc.xff {
			r.Header.Add("X-Forwarded-For", v)
		}
		if got := clientIP(r, tc.trust); got != tc.want {
			t.Errorf("clientIP(trust=%v, %s, %v) = %q, want %q", tc.trust, tc.remote, tc.xff, got, tc.want)
		}
	}
}

func TestChatLimitBehindProxy(t *testing.T) {
	s := testServer(t, &countingAdvisor{})
	s.ChatLimit, s.TrustProxy = NewChatLimiter(6, 1), true
	h := s.Handler()
	// One proxy address, two shoppers: separate buckets.
	if rec := postChat(h, "10.0.0.1:1", "", "a"); rec.Code != 200 {
		t.Fatal(rec.Code)
	}
	if rec := postChat(h, "10.0.0.1:1", "1.2.3.4", "a"); rec.Code != 200 {
		t.Fatalf("second shopper behind the proxy: %d", rec.Code)
	}
	if rec := postChat(h, "10.0.0.1:1", "6.6.6.6, 1.2.3.4", "a"); rec.Code != http.StatusTooManyRequests {
		t.Errorf("spoofed first hop escaped the bucket: %d", rec.Code)
	}
}

func TestChatLimitIPv6Prefix(t *testing.T) {
	s := testServer(t, &countingAdvisor{})
	s.ChatLimit = NewChatLimiter(6, 1)
	h := s.Handler()
	if rec := postChat(h, "[2001:db8:0:7::1]:1", "", "a"); rec.Code != 200 {
		t.Fatal(rec.Code)
	}
	if rec := postChat(h, "[2001:db8:0:7::2]:1", "", "a"); rec.Code != http.StatusTooManyRequests {
		t.Errorf("a new address in the same /64 got a fresh bucket: %d", rec.Code)
	}
	if rec := postChat(h, "[2001:db8:0:8::1]:1", "", "a"); rec.Code != 200 {
		t.Errorf("another /64: %d", rec.Code)
	}
}

func TestChatDailyBudget(t *testing.T) {
	adv := &countingAdvisor{}
	s := testServer(t, adv)
	s.Budget = NewTurnBudget(2)
	h := s.Handler()
	// A malformed request never spends the budget.
	if rec := postChat(h, "203.0.113.7:1", "", " "); rec.Code != http.StatusBadRequest {
		t.Fatal(rec.Code)
	}
	for i, ip := range []string{"203.0.113.7:1", "198.51.100.2:1"} {
		if rec := postChat(h, ip, "", "a"); rec.Code != 200 {
			t.Fatalf("turn %d within budget: %d", i, rec.Code)
		}
	}
	rec := postChat(h, "192.0.2.9:1", "", "a") // a third client: the cap is shared
	if rec.Code != http.StatusServiceUnavailable || !strings.Contains(errorText(t, rec), "明天") {
		t.Fatalf("over the daily budget: %d %s", rec.Code, rec.Body)
	}
	if adv.calls.Load() != 2 {
		t.Errorf("advisor calls = %d, want 2", adv.calls.Load())
	}
}

func TestTurnBudgetResetsAtTaiwanMidnight(t *testing.T) {
	b := NewTurnBudget(1)
	now := time.Date(2026, 10, 16, 15, 59, 0, 0, time.UTC) // 23:59 in Taipei
	b.now = func() time.Time { return now }
	if !b.Take() || b.Take() {
		t.Fatal("want one turn, then none")
	}
	now = now.Add(2 * time.Minute) // 00:01 the next Taipei day
	if !b.Take() {
		t.Error("budget did not reset at Taiwan midnight")
	}
}

func TestChatRuneCap(t *testing.T) {
	adv := &countingAdvisor{}
	h := testServer(t, adv).Handler()
	if rec := postChat(h, "203.0.113.7:1", "", strings.Repeat("包", maxChatRunes)); rec.Code != 200 {
		t.Fatalf("message at the cap: %d %s", rec.Code, rec.Body)
	}
	rec := postChat(h, "203.0.113.7:1", "", strings.Repeat("包", maxChatRunes+1))
	if rec.Code != http.StatusBadRequest || !strings.Contains(errorText(t, rec), "1000 字以內") {
		t.Fatalf("message over the cap: %d %s", rec.Code, rec.Body)
	}
	// A paste past the 16 KB body cap gets the same shopper-facing line.
	rec = postChat(h, "203.0.113.7:1", "", strings.Repeat("包", 20000))
	if rec.Code != http.StatusBadRequest || !strings.Contains(errorText(t, rec), "1000 字以內") {
		t.Fatalf("body over the byte cap: %d %s", rec.Code, rec.Body)
	}
	if adv.calls.Load() != 1 {
		t.Errorf("advisor calls = %d, want 1 (an over-long message never reaches it)", adv.calls.Load())
	}
}

func TestChatNoLimiterByDefault(t *testing.T) {
	h := testServer(t, &countingAdvisor{}).Handler()
	for i := range 20 {
		if rec := postChat(h, "203.0.113.7:1", "", "a"); rec.Code != 200 {
			t.Fatalf("turn %d without a limiter: %d", i, rec.Code)
		}
	}
}
