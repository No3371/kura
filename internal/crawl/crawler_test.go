package crawl

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/No3371/kura/internal/listing"
	"github.com/No3371/kura/internal/store"
)

// fakeSource serves fixed list pages per target and records detail fetches.
type fakeSource struct {
	name    string
	targets []string
	pages   map[string][][]listing.Listing
	fail    map[string]bool     // target -> Sweep errors after its first page
	sold    map[string][]string // target -> source ids its first page shows sold
	capped  map[string]bool     // target -> Sweep stops before the last page
	goneURL map[string]bool     // product pages that say sold
	details []string
}

func (f *fakeSource) Name() string      { return f.name }
func (f *fakeSource) Targets() []string { return f.targets }

func (f *fakeSource) Sweep(ctx context.Context, target string, maxPages int, visit func(page int, items []listing.Listing, sold []string) error) (bool, error) {
	for i, items := range f.pages[target] {
		var sold []string
		if i == 0 {
			sold = f.sold[target]
		}
		if err := visit(i+1, items, sold); err != nil {
			return false, err
		}
		if f.fail[target] {
			return false, errors.New("list page 2: 503")
		}
	}
	return !f.capped[target], nil
}

func (f *fakeSource) Detail(ctx context.Context, url string) (listing.Detail, bool, error) {
	f.details = append(f.details, url)
	if f.goneURL[url] {
		return listing.Detail{}, true, nil
	}
	return listing.Detail{Color: "c"}, false, nil
}

// bags returns n listings of source on one list page each, ids prefixed by target.
func bags(source, target string, n int) [][]listing.Listing {
	var pages [][]listing.Listing
	for i := range n {
		id := fmt.Sprintf("%s-%d", target, i)
		pages = append(pages, []listing.Listing{{Source: source, SourceID: id, URL: "https://" + source + "/" + id,
			Brand: target, Title: id, Rank: listing.RankA, PriceJPY: 100000}})
	}
	return pages
}

func newCrawler(t *testing.T, budget int, sources ...Source) (*Crawler, *store.Store) {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "kura.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	now := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	return &Crawler{Store: st, Sources: sources, Log: slog.New(slog.NewTextHandler(io.Discard, nil)),
		DetailBudget: budget, Now: func() time.Time { return now }}, st
}

func TestRunAllSweepsEverySource(t *testing.T) {
	ctx := context.Background()
	fa := &fakeSource{name: "fa", targets: []string{"HERMES", "CHANEL"},
		pages: map[string][][]listing.Listing{"HERMES": bags("fa", "HERMES", 2), "CHANEL": bags("fa", "CHANEL", 2)},
		fail:  map[string]bool{"CHANEL": true}}
	fb := &fakeSource{name: "fb", targets: []string{"HERMES"},
		pages: map[string][][]listing.Listing{"HERMES": bags("fb", "HERMES", 2)}}
	c, st := newCrawler(t, 10, fa, fb)

	if err := c.RunAll(ctx); err != nil {
		t.Fatal(err)
	}

	runs, err := st.RecentRuns(ctx, 10)
	if err != nil {
		t.Fatal(err)
	}
	type key struct{ source, target string }
	got := map[key]store.Run{}
	for _, r := range runs {
		got[key{r.Source, r.Target}] = r
	}
	if len(runs) != 3 || len(got) != 3 {
		t.Fatalf("runs = %+v, want one per source and target", runs)
	}
	if r := got[key{"fa", "HERMES"}]; !r.Complete || r.Error != "" {
		t.Errorf("fa HERMES run = %+v, want complete", r)
	}
	if r := got[key{"fa", "CHANEL"}]; r.Complete || r.Error == "" {
		t.Errorf("fa CHANEL run = %+v, want incomplete with error", r)
	}
	if r := got[key{"fb", "HERMES"}]; !r.Complete || r.Error != "" {
		t.Errorf("fb HERMES run = %+v, want complete after fa's failure", r)
	}

	rows, err := st.AllRows(ctx)
	if err != nil {
		t.Fatal(err)
	}
	perSource := map[string]int{}
	for _, r := range rows {
		perSource[strings.SplitN(r.ID, ":", 2)[0]]++
	}
	if perSource["fa"] == 0 || perSource["fb"] == 0 {
		t.Errorf("listings per source = %v, want both", perSource)
	}

	for _, f := range []*fakeSource{fa, fb} {
		if len(f.details) == 0 {
			t.Errorf("%s fetched no details", f.name)
		}
		for _, u := range f.details {
			if !strings.HasPrefix(u, "https://"+f.name+"/") {
				t.Errorf("%s fetched another source's page %q", f.name, u)
			}
		}
	}

	on, err := st.Onboarded(ctx)
	if err != nil {
		t.Fatal(err)
	}
	_, faH := on[store.Pair{Source: "fa", Target: "HERMES"}]
	_, fbH := on[store.Pair{Source: "fb", Target: "HERMES"}]
	_, faC := on[store.Pair{Source: "fa", Target: "CHANEL"}]
	if !faH || !fbH || faC {
		t.Errorf("onboarded = %v, want fa/HERMES and fb/HERMES only", on)
	}
}

func TestDetailBudgetPerSource(t *testing.T) {
	fa := &fakeSource{name: "fa", targets: []string{"HERMES"},
		pages: map[string][][]listing.Listing{"HERMES": bags("fa", "HERMES", 3)}}
	fb := &fakeSource{name: "fb", targets: []string{"HERMES"},
		pages: map[string][][]listing.Listing{"HERMES": bags("fb", "HERMES", 3)}}
	c, _ := newCrawler(t, 2, fa, fb)

	if err := c.RunAll(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(fa.details) != 2 || len(fb.details) != 2 {
		t.Errorf("detail fetches fa=%d fb=%d, want 2 each", len(fa.details), len(fb.details))
	}
}

// row returns the stored row of listing id, or fails the test.
func row(t *testing.T, st *store.Store, id string) store.Row {
	t.Helper()
	rows, err := st.AllRows(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range rows {
		if r.ID == id {
			return r
		}
	}
	t.Fatalf("no row %s", id)
	return store.Row{}
}

func latestRun(t *testing.T, st *store.Store) store.Run {
	t.Helper()
	runs, err := st.RecentRuns(context.Background(), 1)
	if err != nil || len(runs) != 1 {
		t.Fatalf("runs = %v, %v", runs, err)
	}
	return runs[0]
}

func TestListSoldRowsMarkGone(t *testing.T) {
	ctx := context.Background()
	fa := &fakeSource{name: "fa", targets: []string{"HERMES"},
		pages:  map[string][][]listing.Listing{"HERMES": bags("fa", "HERMES", 2)},
		capped: map[string]bool{"HERMES": true}}
	c, st := newCrawler(t, 0, fa)
	t0 := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	now := t0
	c.Now = func() time.Time { return now }
	if err := c.RunAll(ctx); err != nil {
		t.Fatal(err)
	}

	fa.pages["HERMES"] = fa.pages["HERMES"][:1]
	fa.sold = map[string][]string{"HERMES": {"HERMES-1", "nosuch"}}
	now = t0.Add(time.Hour)
	if err := c.RunAll(ctx); err != nil {
		t.Fatal(err)
	}
	if r := row(t, st, "fa:HERMES-1"); r.Status != "gone" || !r.GoneAt.Equal(t0.Add(time.Hour)) {
		t.Errorf("HERMES-1 = %s gone at %v, want gone at t0+1h", r.Status, r.GoneAt)
	}
	if r := row(t, st, "fa:HERMES-0"); r.Status != "active" {
		t.Errorf("HERMES-0 = %s, want active", r.Status)
	}
	rows, _ := st.AllRows(ctx)
	if len(rows) != 2 {
		t.Errorf("rows = %d, want 2 (a sold id Kura does not hold inserts nothing)", len(rows))
	}
	if r := latestRun(t, st); r.Gone != 1 {
		t.Errorf("run gone = %d, want 1", r.Gone)
	}

	now = t0.Add(2 * time.Hour) // the sold row is still on the page
	if err := c.RunAll(ctx); err != nil {
		t.Fatal(err)
	}
	if r := row(t, st, "fa:HERMES-1"); !r.GoneAt.Equal(t0.Add(time.Hour)) {
		t.Errorf("HERMES-1 gone at %v after a repeated signal, want t0+1h", r.GoneAt)
	}
	if r := latestRun(t, st); r.Gone != 0 {
		t.Errorf("run gone = %d on a repeated signal, want 0", r.Gone)
	}
}

// TestListSoldRowsAddToSweepMissing: a complete sweep's SweepMissing count is
// added to the list-row marks, not written over them.
func TestListSoldRowsAddToSweepMissing(t *testing.T) {
	ctx := context.Background()
	fa := &fakeSource{name: "fa", targets: []string{"HERMES"},
		pages: map[string][][]listing.Listing{"HERMES": bags("fa", "HERMES", 3)}}
	c, st := newCrawler(t, 0, fa)
	t0 := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	now := t0
	c.Now = func() time.Time { return now }
	if err := c.RunAll(ctx); err != nil { // all three seen
		t.Fatal(err)
	}
	fa.pages["HERMES"] = fa.pages["HERMES"][:1]
	for i, sold := range [][]string{nil, {"HERMES-1"}} { // HERMES-1 and -2 missed twice
		now = t0.Add(time.Duration(i+1) * time.Hour)
		fa.sold = map[string][]string{"HERMES": sold}
		if err := c.RunAll(ctx); err != nil {
			t.Fatal(err)
		}
	}
	// Last run: HERMES-1 by its sold row, HERMES-2 by its second missed sweep.
	if r := latestRun(t, st); r.Gone != 2 || !r.Complete {
		t.Errorf("run = %+v, want complete with gone 2", r)
	}
}

func TestRecheckIncompleteTargetsOnly(t *testing.T) {
	ctx := context.Background()
	fa := &fakeSource{name: "fa", targets: []string{"HERMES", "CHANEL"},
		pages:  map[string][][]listing.Listing{"HERMES": bags("fa", "HERMES", 2), "CHANEL": bags("fa", "CHANEL", 1)},
		capped: map[string]bool{"HERMES": true}}
	c, st := newCrawler(t, 10, fa)
	c.RecheckAfter = 24 * time.Hour
	t0 := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	now := t0
	c.Now = func() time.Time { return now }
	if err := c.RunAll(ctx); err != nil {
		t.Fatal(err)
	}
	if len(fa.details) != 3 {
		t.Fatalf("first crawl details = %v, want 3 first fetches", fa.details)
	}

	fa.pages = map[string][][]listing.Listing{"HERMES": bags("fa", "HERMES", 1)}
	fa.goneURL = map[string]bool{"https://fa/HERMES-1": true}
	fa.details = nil
	now = t0.Add(25 * time.Hour)
	if err := c.RunAll(ctx); err != nil {
		t.Fatal(err)
	}
	// HERMES-0 was on this crawl's list page; CHANEL's sweep completed, so its
	// missing item is SweepMissing's.
	if want := []string{"https://fa/HERMES-1"}; !slices.Equal(fa.details, want) {
		t.Errorf("re-checks = %v, want %v", fa.details, want)
	}
	if r := row(t, st, "fa:HERMES-1"); r.Status != "gone" {
		t.Errorf("HERMES-1 = %s after a sold re-check, want gone", r.Status)
	}
}

func TestRecheckCadenceAndPriority(t *testing.T) {
	ctx := context.Background()
	fa := &fakeSource{name: "fa", targets: []string{"HERMES"},
		pages:  map[string][][]listing.Listing{"HERMES": bags("fa", "HERMES", 2)},
		capped: map[string]bool{"HERMES": true}}
	c, _ := newCrawler(t, 2, fa)
	c.RecheckAfter = 24 * time.Hour
	t0 := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	now := t0
	c.Now = func() time.Time { return now }
	crawl := func(at time.Duration, pages [][]listing.Listing) []string {
		t.Helper()
		fa.pages = map[string][][]listing.Listing{"HERMES": pages}
		fa.details = nil
		now = t0.Add(at)
		if err := c.RunAll(ctx); err != nil {
			t.Fatal(err)
		}
		return fa.details
	}
	if got := crawl(0, bags("fa", "HERMES", 2)); len(got) != 2 {
		t.Fatalf("t0 details = %v, want 2 first fetches", got)
	}
	if got := crawl(time.Hour, nil); len(got) != 0 {
		t.Errorf("t0+1h details = %v, want none (read 1 h ago)", got)
	}
	newOne := bags("fa", "NEW", 1)
	if got, want := crawl(25*time.Hour, newOne), []string{"https://fa/NEW-0", "https://fa/HERMES-0"}; !slices.Equal(got, want) {
		t.Errorf("t0+25h details = %v, want %v (first fetch first, then oldest read)", got, want)
	}
	if got, want := crawl(26*time.Hour, nil), []string{"https://fa/HERMES-1"}; !slices.Equal(got, want) {
		t.Errorf("t0+26h details = %v, want %v (HERMES-0 read 1 h ago)", got, want)
	}
	c.RecheckAfter = 0
	if got := crawl(60*time.Hour, nil); len(got) != 0 {
		t.Errorf("RecheckAfter 0: details = %v, want none", got)
	}
}
