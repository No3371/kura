package judge

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// fakeJev serves one canned status + body and records the last request.
type fakeJev struct {
	status int
	body   string
	delay  time.Duration
	path   string
	auth   string
	req    map[string]any
}

func (f *fakeJev) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	raw, _ := io.ReadAll(r.Body)
	f.path, f.auth = r.URL.Path, r.Header.Get("Authorization")
	json.Unmarshal(raw, &f.req)
	if f.delay > 0 {
		select {
		case <-time.After(f.delay):
		case <-r.Context().Done():
			return
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(f.status)
	io.WriteString(w, f.body)
}

func newClient(t *testing.T, f *fakeJev) *Client {
	t.Helper()
	srv := httptest.NewServer(f)
	t.Cleanup(srv.Close)
	return &Client{Key: "test-key", Model: "jev-1.14.0", BaseURL: srv.URL, HTTP: srv.Client()}
}

var testQuestions = map[string]Question{
	"intent": {Type: "choice", Instructions: "What?", Criteria: map[string]string{"find_bag": "a", "off_topic": "b"}},
	"spam":   {Type: "noul", Instructions: "Spam?", Criteria: map[string]string{"true": "t", "false": "f"}},
}

func TestAskDecodesChoiceAndNoul(t *testing.T) {
	f := &fakeJev{status: 200, body: `{"model":"jev-1.14.0","answers":{
		"intent":{"type":"choice","choice":"find_bag","confidence":0.92,"probabilities":{"find_bag":0.95,"off_topic":0.05}},
		"spam":{"type":"noul","noul":0.03}},"usage":{"input_tokens":1342,"output_tokens":0}}`}
	c := newClient(t, f)
	res, err := c.Ask(context.Background(), map[string]any{"message": "愛馬仕"}, testQuestions)
	if err != nil {
		t.Fatal(err)
	}
	if res.Model != "jev-1.14.0" || res.Usage.InputTokens != 1342 {
		t.Errorf("model/usage = %q %+v", res.Model, res.Usage)
	}
	in := res.Answers["intent"]
	if in.Choice != "find_bag" || in.Confidence == nil || *in.Confidence != 0.92 || in.Probabilities["off_topic"] != 0.05 || in.Noul != nil {
		t.Errorf("intent answer = %+v", in)
	}
	sp := res.Answers["spam"]
	if sp.Noul == nil || *sp.Noul != 0.03 || sp.Confidence != nil {
		t.Errorf("noul answer = %+v (want noul 0.03, no confidence)", sp)
	}

	// Request shape: path, bearer key, pinned model, Choice options under criteria.
	if f.path != "/v1/systemone" || f.auth != "Bearer test-key" {
		t.Errorf("path %q auth %q", f.path, f.auth)
	}
	if f.req["model"] != "jev-1.14.0" {
		t.Errorf("model sent = %v", f.req["model"])
	}
	q := f.req["questions"].(map[string]any)["intent"].(map[string]any)
	if q["type"] != "choice" || q["criteria"].(map[string]any)["off_topic"] != "b" || q["options"] != nil {
		t.Errorf("intent question sent = %v", q)
	}
	if f.req["state"].(map[string]any)["message"] != "愛馬仕" {
		t.Errorf("state sent = %v", f.req["state"])
	}
}

func TestAskMapsErrors(t *testing.T) {
	cases := []struct {
		name              string
		status            int
		body              string
		wantType, wantMsg string
		rateLimited, auth bool
	}{
		{"auth", 403, `{"detail":{"error_type":"authentication_error","message":"Invalid API key"}}`,
			"authentication_error", "Invalid API key", false, true},
		{"rate limit", 429, `{"detail":{"error_type":"rate_limit_error","message":"slow down"}}`,
			"rate_limit_error", "slow down", true, false},
		{"string detail", 429, `{"detail":"Too Many Requests"}`, "", "Too Many Requests", true, false},
		{"server", 502, `<html>bad gateway</html>`, "", "", false, false},
		{"server json", 500, `{"detail":{"error_type":"internal_error","message":"boom"}}`, "internal_error", "boom", false, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := newClient(t, &fakeJev{status: tc.status, body: tc.body})
			_, err := c.Ask(context.Background(), "x", testQuestions)
			apiErr, ok := errors.AsType[*APIError](err)
			if !ok {
				t.Fatalf("err = %v, want *APIError", err)
			}
			if apiErr.Status != tc.status || apiErr.Type != tc.wantType || apiErr.Message != tc.wantMsg ||
				apiErr.RateLimited() != tc.rateLimited || apiErr.Auth() != tc.auth {
				t.Errorf("got %+v (rate %v auth %v)", apiErr, apiErr.RateLimited(), apiErr.Auth())
			}
		})
	}
}

func TestAskErrorNeverCarriesKey(t *testing.T) {
	c := newClient(t, &fakeJev{status: 403, body: `{"detail":{"error_type":"authentication_error","message":"bad"}}`})
	_, err := c.Ask(context.Background(), "x", testQuestions)
	if err == nil {
		t.Fatal("want error")
	}
	if s := err.Error(); s == "" || strings.Contains(s, "test-key") {
		t.Errorf("error text %q", s)
	}
}

func TestAskDecodeErrorsAreWrapped(t *testing.T) {
	for _, body := range []string{`not json`, `{"model":"jev-1.14.0"}`} {
		c := newClient(t, &fakeJev{status: 200, body: body})
		_, err := c.Ask(context.Background(), "x", testQuestions)
		if err == nil {
			t.Errorf("body %q: want error", body)
			continue
		}
		if _, ok := errors.AsType[*APIError](err); ok {
			t.Errorf("body %q: decode failure reported as APIError: %v", body, err)
		}
	}
}

func TestAskRespectsContextDeadline(t *testing.T) {
	c := newClient(t, &fakeJev{status: 200, body: `{"answers":{}}`, delay: 2 * time.Second})
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	start := time.Now()
	_, err := c.Ask(ctx, "x", testQuestions)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err = %v, want deadline exceeded", err)
	}
	if took := time.Since(start); took > time.Second {
		t.Errorf("Ask took %v after a 50ms deadline", took)
	}
}
