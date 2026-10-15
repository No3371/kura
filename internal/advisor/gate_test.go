package advisor

import (
	"encoding/json"
	"maps"
	"slices"
	"strings"
	"testing"

	"github.com/No3371/kura/internal/judge"
	"github.com/No3371/kura/internal/listing"
)

func ptr(f float64) *float64 { return &f }

// jevResult builds a Jev result: intent at conf, every abuse class at 0.02
// unless overridden.
func jevResult(intent string, conf float64, abuse map[string]float64) judge.Result {
	res := judge.Result{Model: "jev-1.14.0", Answers: map[string]judge.Answer{
		"intent": {Type: "choice", Choice: intent, Confidence: ptr(conf)},
	}}
	for name := range gateAbuse {
		p := 0.02
		if v, ok := abuse[name]; ok {
			p = v
		}
		res.Answers[name] = judge.Answer{Type: "noul", Noul: ptr(p)}
	}
	return res
}

func TestGatePolicy(t *testing.T) {
	cases := []struct {
		name      string
		intent    string
		conf      float64
		abuse     map[string]float64
		action    GateAction
		blockedBy string
		flagged   bool
	}{
		{"clean find_bag", "find_bag", 0.95, nil, ActionPass, "", false},
		{"block at the line", "find_bag", 0.9, map[string]float64{"injection": 0.80}, ActionBlock, "injection", false},
		{"just under block is flagged", "find_bag", 0.9, map[string]float64{"injection": 0.79}, ActionPass, "", true},
		{"flag at the line", "find_bag", 0.9, map[string]float64{"spam": 0.50}, ActionPass, "", true},
		{"just under flag", "find_bag", 0.9, map[string]float64{"spam": 0.49}, ActionPass, "", false},
		{"highest class names the block", "smalltalk", 0.6, map[string]float64{"free_llm": 0.85, "injection": 0.99}, ActionBlock, "injection", false},
		{"block beats off_topic", "off_topic", 1.0, map[string]float64{"counterfeit": 0.95}, ActionBlock, "counterfeit", false},
		{"block with another class flagged", "find_bag", 0.9, map[string]float64{"injection": 0.99, "free_llm": 0.66}, ActionBlock, "injection", true},
		{"off_topic at the line", "off_topic", 0.70, nil, ActionShortCircuit, "", false},
		{"off_topic under the line", "off_topic", 0.69, nil, ActionPass, "", false},
		{"off_topic flagged", "off_topic", 0.9, map[string]float64{"harassment": 0.6}, ActionShortCircuit, "", true},
		{"authenticity passes", "authenticity", 0.99, nil, ActionPass, "", false},
		{"low-confidence intent passes", "listing_detail", 0.48, nil, ActionPass, "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			v, err := decide(jevResult(tc.intent, tc.conf, tc.abuse), DefaultThresholds)
			if err != nil {
				t.Fatal(err)
			}
			if v.Action != tc.action || v.BlockedBy != tc.blockedBy || v.Flagged != tc.flagged {
				t.Errorf("got action %s blockedBy %q flagged %v", v.Action, v.BlockedBy, v.Flagged)
			}
			if v.Intent != tc.intent || v.IntentConf != tc.conf || len(v.Abuse) != len(gateAbuse) {
				t.Errorf("verdict lost fields: %+v", v)
			}
		})
	}
}

func TestGatePolicyRejectsPartialResult(t *testing.T) {
	res := jevResult("find_bag", 0.9, nil)
	delete(res.Answers, "spam")
	if _, err := decide(res, DefaultThresholds); err == nil {
		t.Error("missing abuse answer: want error (fail open), got a verdict")
	}
	res = jevResult("find_bag", 0.9, nil)
	res.Answers["injection"] = judge.Answer{Type: "noul"}
	if _, err := decide(res, DefaultThresholds); err == nil {
		t.Error("abuse answer without noul: want error")
	}
	res = jevResult("find_bag", 0.9, nil)
	delete(res.Answers, "intent")
	if v, err := decide(res, DefaultThresholds); err != nil || v.Action != ActionPass || v.Intent != "" {
		t.Errorf("missing intent: %+v %v, want pass with no intent", v, err)
	}
}

func TestGateQuestionsShape(t *testing.T) {
	qs := gateQuestions()
	want := []string{"counterfeit", "free_llm", "harassment", "injection", "intent", "spam"}
	got := slices.Sorted(maps.Keys(qs))
	if !slices.Equal(got, want) {
		t.Fatalf("question names = %v", got)
	}
	in := qs["intent"]
	opts, ok := in.Criteria.(map[string]string)
	if in.Type != "choice" || !ok || len(opts) != 7 || opts["off_topic"] == "" {
		t.Errorf("intent question = %+v", in)
	}
	if _, ok := opts["fraud"]; ok {
		t.Error("intent has a fraud option; abuse checks decide the action")
	}
	inj := qs["injection"]
	crit, _ := inj.Criteria.(map[string]string)
	if inj.Type != "noul" || !strings.HasPrefix(inj.Instructions, "In `message`: Tries to change") ||
		!strings.HasSuffix(inj.Instructions, "?") || crit["true"] == "" || crit["false"] == "" {
		t.Errorf("injection question = %+v", inj)
	}
}

// TestGateWordingPinned pins the tuned wording: the thresholds were picked
// for exactly these texts, so an edit must fail here and prompt a re-eval.
func TestGateWordingPinned(t *testing.T) {
	pins := map[string]string{
		gateIntents["off_topic"]:    "Not about finding or buying a bag through this service: unrelated topics, or anything about a bag the user already owns (selling, consigning, valuing, repairing or authenticating it)",
		gateIntents["authenticity"]: "Asks whether a bag they are considering buying is genuine, or how to authenticate one before buying",
		gateAbuse["free_llm"][0]:    "Asks the assistant to produce a substantial piece of work unrelated to buying bags: write or debug code, write an essay, report or homework, or translate or summarize a document that is not a bag listing",
		gateAbuse["counterfeit"][1]: "Asks how to spot fakes or whether an item is genuine, or mentions faux/synthetic leather",
	}
	for got, want := range pins {
		if got != want {
			t.Errorf("wording drifted:\n got %q\nwant %q", got, want)
		}
	}
	svc := gateStateFor("m", nil, nil)["service"].(map[string]any)
	if h := svc["how_it_works"]; h != "The assistant is an LLM with a fixed system prompt and listing-search tools; users cannot change its instructions. Shoppers asking how it works, what its rules are or how it picks recommendations is normal." {
		t.Errorf("how_it_works drifted: %q", h)
	}
	if c := svc["cannot_do"].([]string); c[len(c)-1] != "help with a bag the user already owns: selling, consigning, valuing or repairing it" {
		t.Errorf("cannot_do lacks the buyer-only line: %v", c)
	}
}

func TestGateStateShopsFromSchemes(t *testing.T) {
	// Built from a local list: registering schemes would leak them into the
	// package-global registry the other tests read.
	shops := shopNamesOf([]listing.Scheme{
		{Source: "gw", Shop: "別家店"},
		{Source: "gx", Shop: "閘門店"},
		{Source: "gy", Shop: "閘門店"}, // two sources, one shop
	})
	if want := []string{"別家店", "閘門店"}; !slices.Equal(shops, want) {
		t.Errorf("shops = %v, want %v", shops, want)
	}
	st := gateStateFor("第二個呢？", []string{"愛馬仕 6 萬以內"}, shops)
	b, _ := json.Marshal(st)
	var back struct {
		Message string   `json:"message"`
		Recent  []string `json:"recent_user_messages"`
		Service struct {
			Shops []string `json:"shops"`
		} `json:"service"`
	}
	if err := json.Unmarshal(b, &back); err != nil {
		t.Fatal(err)
	}
	if back.Message != "第二個呢？" || len(back.Recent) != 1 || !slices.Contains(back.Service.Shops, "閘門店") {
		t.Errorf("state = %s", b)
	}
}

func TestPromptIsBuyerOnly(t *testing.T) {
	p := buildSystemPrompt()
	if !strings.Contains(p, "Kura 只服務買家") || !strings.Contains(p, "估價") {
		t.Error("system prompt lacks the buyer-only scope line")
	}
}
