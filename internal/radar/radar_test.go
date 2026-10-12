package radar

import (
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
	listing.RegisterScheme(listing.Scheme{Source: "t", Name: "テスト", Graded: "店家鑑定", Grades: gs})
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
}
