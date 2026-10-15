package main

import (
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/No3371/kura/internal/advisor"
	"github.com/No3371/kura/internal/crawl"
	"github.com/No3371/kura/internal/source/allu"
	"github.com/No3371/kura/internal/source/brandoff"
	"github.com/No3371/kura/internal/source/grare"
)

// named is a source that only answers Name; checkSources calls nothing else.
type named struct {
	crawl.Source
	name string
}

func (n named) Name() string { return n.name }

func TestCheckSources(t *testing.T) {
	for _, c := range []struct {
		name    string
		sources []crawl.Source
		wantErr string
	}{
		{"brandoff alone", []crawl.Source{&brandoff.Source{}}, ""},
		{"brandoff + grare", []crawl.Source{&brandoff.Source{}, &grare.Source{}}, ""},
		{"all three", []crawl.Source{&brandoff.Source{}, &grare.Source{}, &allu.Source{}}, ""},
		{"grare twice", []crawl.Source{&grare.Source{}, &grare.Source{}}, `duplicate source "grare"`},
		{"none", nil, ""},
		{"no scheme", []crawl.Source{&brandoff.Source{}, named{name: "nosuch"}}, `source "nosuch" has no registered condition scheme`},
		{"duplicate name", []crawl.Source{&brandoff.Source{}, named{name: brandoff.SourceName}}, `duplicate source "brandoff"`},
	} {
		err := checkSources(c.sources)
		switch {
		case c.wantErr == "" && err != nil:
			t.Errorf("%s: unexpected error %v", c.name, err)
		case c.wantErr != "" && (err == nil || !strings.Contains(err.Error(), c.wantErr)):
			t.Errorf("%s: error = %v, want %q", c.name, err, c.wantErr)
		}
	}
}

func TestGateAdvisor(t *testing.T) {
	const key = "test-not-a-real-key"
	cases := []struct {
		judge, key string
		budget     time.Duration
		wantMode   string // "" = error
	}{
		{"auto", key, 800 * time.Millisecond, "mock+jev:shadow"},
		{"auto", "", 800 * time.Millisecond, "mock"},
		{"off", key, 800 * time.Millisecond, "mock"},
		{"shadow", key, 800 * time.Millisecond, "mock+jev:shadow"},
		{"enforce", key, 800 * time.Millisecond, "mock+jev:enforce"},
		{"enforce", "", 800 * time.Millisecond, ""},
		{"shadow", "", 800 * time.Millisecond, ""},
		{"enforce", key, 0, ""},
		{"strict", key, 800 * time.Millisecond, ""},
	}
	for _, tc := range cases {
		c := config{judge: tc.judge, judgeModel: "jev-1.14.0", judgeBudget: tc.budget}
		adv, err := gateAdvisor(&advisor.Mock{}, c, tc.key, slog.New(slog.DiscardHandler))
		if tc.wantMode == "" {
			if err == nil {
				t.Errorf("-judge %s key=%v budget %v: want error, got %s", tc.judge, tc.key != "", tc.budget, adv.Mode())
			} else if strings.Contains(err.Error(), key) {
				t.Errorf("error text carries the key: %v", err)
			}
			continue
		}
		if err != nil {
			t.Errorf("-judge %s key=%v: %v", tc.judge, tc.key != "", err)
			continue
		}
		if adv.Mode() != tc.wantMode {
			t.Errorf("-judge %s key=%v: mode %q, want %q", tc.judge, tc.key != "", adv.Mode(), tc.wantMode)
		}
	}
}
