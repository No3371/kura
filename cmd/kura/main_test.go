package main

import (
	"strings"
	"testing"

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
