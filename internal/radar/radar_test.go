package radar

import (
	"slices"
	"testing"
	"time"

	"github.com/No3371/kura/internal/fx"
	"github.com/No3371/kura/internal/listing"
	"github.com/No3371/kura/internal/store"
)

var t0 = time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)

func init() {
	// Test source "t" grades like a shop would: its own labels, one per rank.
	var gs []listing.Grade
	for _, r := range listing.Ranks() {
		gs = append(gs, listing.Grade{Code: "等級" + string(r), Label: "等級" + string(r), DescZH: "測試", Rank: r})
	}
	listing.RegisterScheme(listing.Scheme{Source: "t", Name: "テスト", Shop: "テスト店", Graded: "店家鑑定", Grades: gs})
}

func row(id string, code string, rank listing.Rank, price int64, firstDay int, prices ...int64) store.Row {
	r := store.Row{ID: "t:" + id, Status: "active", FirstSeen: t0.AddDate(0, 0, firstDay)}
	r.Listing = listing.Listing{Source: "t", SourceID: id, Brand: "LOUIS VUITTON", ModelName: "ネヴァーフルMM",
		ModelCode: code, Category: "tote", Grade: "等級" + string(rank), Rank: rank, PriceJPY: price}
	at := r.FirstSeen
	for _, p := range append(prices, price) {
		r.PriceHistory = append(r.PriceHistory, store.PricePoint{At: at, PriceJPY: p})
		at = at.Add(24 * time.Hour)
	}
	return r
}

func engine(rows ...store.Row) *Engine {
	e := &Engine{FX: &fx.Converter{Fallback: 0.2}, Now: func() time.Time { return t0.AddDate(0, 1, 0) }}
	e.build(rows, "test")
	return e
}

func TestSameRankValuation(t *testing.T) {
	e := engine(
		row("a", "M40156", listing.RankA, 100000, 0),
		row("b", "M40156", listing.RankA, 120000, 0),
		row("c", "M40156", listing.RankA, 140000, 0),
		row("d", "M40156", listing.RankA, 90000, 0), // the deal
	)
	it, _, _ := e.Get("t:d")
	if it.Fair == nil || it.Fair.Method != "same-grade" || it.Fair.FairJPY != 120000 || it.Fair.Comps != 3 {
		t.Fatalf("fair = %+v", it.Fair)
	}
	if got := it.Fair.DealPct; got < 0.24 || got > 0.26 {
		t.Errorf("DealPct = %v, want 0.25", got)
	}
	if it.PriceTWD != 18000 {
		t.Errorf("PriceTWD = %d", it.PriceTWD)
	}
}

func TestRankAdjustedValuation(t *testing.T) {
	// Only B-rank comparables; an A-rank copy should be valued at
	// median(B / 0.85) * 1.0.
	e := engine(
		row("b1", "M40156", listing.RankB, 85000, 0),
		row("b2", "M40156", listing.RankB, 85000, 0),
		row("b3", "M40156", listing.RankB, 85000, 0),
		row("a", "M40156", listing.RankA, 100000, 0),
	)
	it, _, _ := e.Get("t:a")
	if it.Fair == nil || it.Fair.Method != "rank-adjusted" || it.Fair.FairJPY != 100000 || it.Fair.Confidence != "low" {
		t.Fatalf("fair = %+v", it.Fair)
	}
}

func TestDispersedGroupIsLowConfidence(t *testing.T) {
	// Same name, wildly different prices (e.g. canvas vs leather variants).
	var rows []store.Row
	for i, p := range []int64{60000, 90000, 250000, 300000, 600000, 700000} {
		rows = append(rows, row(string(rune('a'+i)), "", listing.RankA, p, 0))
	}
	rows = append(rows, row("x", "", listing.RankA, 100000, 0))
	e := engine(rows...)
	it, _, _ := e.Get("t:x")
	if it.Fair == nil || it.Fair.Confidence != "low" || it.Fair.Dispersion < 1 {
		t.Fatalf("fair = %+v", it.Fair)
	}
	if _, total := e.Search(Filter{OnlyDeals: true}); total != 0 {
		t.Errorf("low-confidence 'deals' leaked into OnlyDeals: %d", total)
	}
}

func TestHalfPriceIsAnomalyNotDeal(t *testing.T) {
	e := engine(
		row("a", "M51130", listing.RankB, 80000, 0),
		row("b", "M51130", listing.RankB, 80000, 0),
		row("c", "M51130", listing.RankB, 85000, 0),
		row("cheap", "M51130", listing.RankB, 22000, 0),
		row("fair", "M51130", listing.RankB, 68000, 0),
	)
	it, _, _ := e.Get("t:cheap")
	if it.Fair == nil || !it.Fair.Anomaly {
		t.Fatalf("fair = %+v", it.Fair)
	}
	got, _ := e.Search(Filter{Sort: "deal", Limit: 1})
	if got[0].ID != "t:fair" {
		t.Errorf("top deal = %s, want the plausible 15%% deal over the anomaly", got[0].ID)
	}
	if _, total := e.Search(Filter{OnlyDeals: true}); total != 1 {
		t.Errorf("OnlyDeals total = %d, want 1 (anomaly excluded)", total)
	}
}

func TestSameGradeMeansSameSource(t *testing.T) {
	// Another shop also says "等級A", but its A isn't necessarily ours: its
	// listings may only count through the rank bridge, not as same-grade comps.
	other := func(id string, price int64) store.Row {
		r := row(id, "M40156", listing.RankA, price, 0)
		r.ID, r.Source = "u:"+id, "u"
		return r
	}
	e := engine(
		other("a", 100000), other("b", 100000), other("c", 100000),
		row("x", "M40156", listing.RankA, 90000, 0),
	)
	it, _, _ := e.Get("t:x")
	if it.Fair == nil || it.Fair.Method != "rank-adjusted" {
		t.Fatalf("fair = %+v, want rank-adjusted across sources", it.Fair)
	}
	if it.Condition.Label != "等級A" || it.ConditionBy != "店家鑑定" {
		t.Errorf("condition = %+v by %q", it.Condition, it.ConditionBy)
	}
}

func TestTooFewComps(t *testing.T) {
	e := engine(row("a", "X1", listing.RankA, 1, 0), row("b", "X1", listing.RankA, 2, 0))
	if it, _, _ := e.Get("t:a"); it.Fair != nil {
		t.Errorf("expected no valuation with 1 comparable, got %+v", it.Fair)
	}
}

func TestDropsAndStats(t *testing.T) {
	sold := row("s", "M40156", listing.RankA, 110000, 5)
	sold.Status = "gone"
	sold.GoneAt = sold.FirstSeen.AddDate(0, 0, 4)
	e := engine(
		row("a", "M40156", listing.RankA, 100000, 0, 120000), // dropped from 120k
		row("b", "M40156", listing.RankB, 80000, 0),
		sold,
	)
	drops := e.Drops("", 10)
	if len(drops) != 1 || drops[0].ID != "t:a" || drops[0].Drop.FromJPY != 120000 {
		t.Fatalf("drops = %+v", drops)
	}
	if d := drops[0].Drop; d.FromTWD != 24000 || d.ToTWD != 20000 {
		t.Errorf("drop TWD = %d → %d, want 24000 → 20000", d.FromTWD, d.ToTWD)
	}
	if drops[0].SourceName != "テスト店" {
		t.Errorf("SourceName = %q", drops[0].SourceName)
	}
	s, items, ok := e.Model("LOUIS VUITTON", "M40156")
	if !ok || len(items) != 3 || items[2].Status != "gone" {
		t.Fatalf("model items = %+v", items)
	}
	if s.Active != 2 || s.Sold != 1 || s.MedianDaysToSell == nil || *s.MedianDaysToSell != 4 {
		t.Errorf("stats = %+v", s)
	}
	if len(s.ByGrade) != 2 || s.ByGrade[0].Grade.Label != "等級A" || s.ByGrade[1].Grade.Label != "等級B" {
		t.Errorf("ByGrade = %+v", s.ByGrade)
	}
	if g := s.ByGrade; g[0].MedianTWD != 20000 || g[0].MinTWD != 20000 || g[1].MaxTWD != 16000 {
		t.Errorf("ByGrade TWD = %+v", g)
	}
	it, _, _ := e.Get("t:a")
	if h := it.History; len(h) != 2 || h[0].PriceTWD != 24000 || h[1].PriceTWD != 20000 {
		t.Errorf("History = %+v", h)
	}
}

func TestIsDeal(t *testing.T) {
	cases := []struct {
		name string
		fair *Valuation
		want bool
	}{
		{"no valuation", nil, false},
		{"5% high", &Valuation{DealPct: 0.05, Confidence: "high"}, true},
		{"4.9% high", &Valuation{DealPct: 0.049, Confidence: "high"}, false},
		{"20% medium", &Valuation{DealPct: 0.2, Confidence: "medium"}, true},
		{"20% low", &Valuation{DealPct: 0.2, Confidence: "low"}, false},
		{"anomaly", &Valuation{DealPct: 0.6, Confidence: "high", Anomaly: true}, false},
	}
	for _, c := range cases {
		if got := IsDeal(&Item{Fair: c.fair}); got != c.want {
			t.Errorf("%s: IsDeal = %v, want %v", c.name, got, c.want)
		}
	}
}

func TestGone(t *testing.T) {
	gone := func(id string, day int) store.Row {
		r := row(id, "M40156", listing.RankA, 100000, 0)
		r.Status, r.GoneAt = "gone", t0.AddDate(0, 0, day)
		return r
	}
	e := engine(gone("g3", 3), row("a", "M40156", listing.RankA, 100000, 0), gone("g6", 6))
	got := e.Gone(10)
	if len(got) != 2 || got[0].ID != "t:g6" || got[1].ID != "t:g3" {
		t.Fatalf("Gone(10) = %+v, want [t:g6 t:g3]", got)
	}
	if got := e.Gone(1); len(got) != 1 || got[0].ID != "t:g6" {
		t.Errorf("Gone(1) = %+v", got)
	}
}

func TestSearchFilters(t *testing.T) {
	e := engine(
		row("a", "M40156", listing.RankA, 100000, 0),
		row("b", "M40156", listing.RankB, 50000, 0),
		row("c", "N41605", listing.RankUnused, 300000, 0),
	)
	got, total := e.Search(Filter{Brand: "路易威登", MinRank: "A", MaxPriceTWD: 25000, Sort: "price_asc"})
	if total != 1 || got[0].ID != "t:a" {
		t.Errorf("search = %d %+v", total, got)
	}
	if _, total := e.Search(Filter{Keywords: []string{"n41605"}}); total != 1 {
		t.Errorf("keyword search total = %d", total)
	}
	all, _ := e.Search(Filter{Sort: "price_asc"})
	page, total := e.Search(Filter{Sort: "price_asc", Offset: 1, Limit: 1})
	if total != 3 || len(page) != 1 || page[0].ID != all[1].ID {
		t.Errorf("offset page = %d %+v, want %s", total, page, all[1].ID)
	}
	if past, total := e.Search(Filter{Offset: 5}); total != 3 || len(past) != 0 {
		t.Errorf("offset past end = %d %+v", total, past)
	}
	if o := e.Overview(); o.ByCategory["tote"] != 3 {
		t.Errorf("ByCategory = %+v", o.ByCategory)
	}
}

func TestNewestOrder(t *testing.T) {
	at := func(id string, after time.Duration) store.Row {
		r := row(id, "M40156", listing.RankA, 100000, 0)
		r.FirstSeen = t0.Add(after)
		return r
	}
	// The initial crawl pages newest first: i1 sits on an earlier page than
	// i2, so it is the newer listing. w1 and w2 appeared after it.
	initial := []store.Row{at("i2", time.Hour), at("i1", 0), at("i1b", 0)}
	gone := at("w3", 60*time.Hour)
	gone.Status, gone.GoneAt = "gone", gone.FirstSeen.Add(time.Hour)
	e := engine(append(initial, at("w1", 30*time.Hour), at("w2", 50*time.Hour), gone)...)

	ids := func(items []Item) (out []string) {
		for _, it := range items {
			out = append(out, it.ID)
		}
		return out
	}
	got, _ := e.Search(Filter{Sort: "newest"})
	if want := []string{"t:w2", "t:w1", "t:i1", "t:i1b", "t:i2"}; !slices.Equal(ids(got), want) {
		t.Errorf("newest = %v, want %v", ids(got), want)
	}
	_, items, _ := e.Model("LOUIS VUITTON", "M40156")
	if n := e.NewestActive(items); n == nil || n.ID != "t:w2" {
		t.Errorf("NewestActive = %+v, want t:w2 (gone t:w3 skipped)", n)
	}

	// First-sweep data only: the shop's newest is the first crawled.
	e = engine(initial...)
	_, items, _ = e.Model("LOUIS VUITTON", "M40156")
	if n := e.NewestActive(items); n == nil || n.ID != "t:i1" {
		t.Errorf("first sweep NewestActive = %+v, want t:i1", n)
	}
	if got, _ := e.Search(Filter{Sort: "newest", Limit: 1}); len(got) != 1 || got[0].ID != "t:i1" {
		t.Errorf("first sweep newest = %v, want [t:i1]", ids(got))
	}
}

func TestPhotoPending(t *testing.T) {
	cases := map[string]bool{
		"": true,
		"https://x/Contents/ProductImages/NowPrinting_L.jpg": true,
		"https://x/2101220285869_L.jpg":                      false,
	}
	for url, want := range cases {
		r := row("p", "M40156", listing.RankA, 100000, 0)
		r.ImageURL = url
		it, _, _ := engine(r).Get("t:p")
		if it.PhotoPending != want {
			t.Errorf("PhotoPending(%q) = %v, want %v", url, it.PhotoPending, want)
		}
	}
}

func TestSoldStatsReady(t *testing.T) {
	// One listing already gone in every case: sold rows alone do not open the
	// gate, only the span from first to last observation does.
	at := func(span time.Duration) *Engine {
		a := row("a", "M40156", listing.RankA, 100000, 0)
		a.LastSeen = a.FirstSeen.Add(span)
		g := row("g", "M40156", listing.RankA, 100000, 0)
		g.Status, g.GoneAt = "gone", g.FirstSeen.Add(24*time.Hour)
		return engine(a, g)
	}
	for _, c := range []struct {
		span time.Duration
		want bool
	}{{MinSoldHistory - time.Hour, false}, {MinSoldHistory, true}, {MinSoldHistory + 24*time.Hour, true}} {
		e := at(c.span)
		if got := e.SoldStatsReady(); got != c.want {
			t.Errorf("span %v: SoldStatsReady = %v, want %v", c.span, got, c.want)
		}
		if o := e.Overview(); o.SoldStatsReady != c.want || o.Sold != 1 {
			t.Errorf("span %v: Overview = %+v", c.span, o)
		}
	}
	// A gone time past every last_seen still counts as an observation.
	g := row("g", "M40156", listing.RankA, 100000, 0)
	g.Status, g.GoneAt = "gone", g.FirstSeen.Add(MinSoldHistory)
	if !engine(g).SoldStatsReady() {
		t.Error("gone_at at MinSoldHistory: SoldStatsReady = false")
	}
	if (&Engine{}).SoldStatsReady() {
		t.Error("empty index: SoldStatsReady = true")
	}
}
