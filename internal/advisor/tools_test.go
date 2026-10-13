package advisor

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/No3371/kura/internal/listing"
)

func TestSoldStatsGatedUntilHistory(t *testing.T) {
	tools := testTools(t) // 4 HERMES 056289CK listings, all first seen now → history span 0
	run := func(name, in string) (map[string]any, string) {
		t.Helper()
		res, err := tools.Run(name, in)
		if err != nil {
			t.Fatal(err)
		}
		var m map[string]any
		if err := json.Unmarshal([]byte(res.JSON), &m); err != nil {
			t.Fatal(err)
		}
		return m, res.Summary
	}
	const statsIn = `{"brand":"HERMES","model_code":"056289CK"}`
	const listingIn = `{"id":"brandoff:1000"}`

	// Not ready: no sold figure anywhere, marker present, summary silent on sales.
	m, sum := run("model_price_stats", statsIn)
	for _, k := range []string{"sold_observed", "sell_through", "median_days_to_sell"} {
		if _, ok := m[k]; ok {
			t.Errorf("model_price_stats: %s present before history is long enough", k)
		}
	}
	if m["sold_stats"] != soldUnknown {
		t.Errorf("model_price_stats: sold_stats = %v", m["sold_stats"])
	}
	if strings.Contains(sum, "已售出") {
		t.Errorf("model_price_stats summary states a sold count: %q", sum)
	}

	m, _ = run("get_listing", listingIn)
	if _, ok := m["model_sold_observed"]; ok {
		t.Error("get_listing: model_sold_observed present")
	}
	if m["sold_stats"] != soldUnknown {
		t.Errorf("get_listing: sold_stats = %v", m["sold_stats"])
	}

	m, _ = run("market_overview", `{}`)
	if _, ok := m["sold_observed"]; ok {
		t.Error("market_overview: sold_observed present")
	}
	if m["sold_stats"] != soldUnknown {
		t.Errorf("market_overview: sold_stats = %v", m["sold_stats"])
	}
	top := m["top_models"].([]any)[0].(map[string]any)
	if _, ok := top["sold_observed"]; ok {
		t.Errorf("market_overview: top_models[0] = %v", top)
	}

	// Ready: a listing first seen 15 days ago, gone now → span ≥ MinSoldHistory, 1 sold.
	ctx := context.Background()
	st := tools.Radar.Store
	old := listing.Listing{Source: "brandoff", SourceID: "2000", URL: "https://example.test", Brand: "HERMES",
		Title: "エルメス ピコタンロックPM", ModelName: "ピコタンロックPM", ModelCode: "056289CK",
		Category: "handbag", Rank: listing.RankA, PriceJPY: 310000}
	if _, err := st.RecordSeen(ctx, "HERMES", old, time.Now().Add(-15*24*time.Hour)); err != nil {
		t.Fatal(err)
	}
	if _, err := st.MarkGone(ctx, "brandoff:2000", time.Now()); err != nil {
		t.Fatal(err)
	}
	if err := tools.Radar.Refresh(ctx); err != nil {
		t.Fatal(err)
	}

	m, sum = run("model_price_stats", statsIn)
	if m["sold_observed"] != float64(1) {
		t.Errorf("model_price_stats: sold_observed = %v, want 1", m["sold_observed"])
	}
	if _, ok := m["sell_through"]; !ok {
		t.Error("model_price_stats: sell_through missing once ready")
	}
	if _, ok := m["sold_stats"]; ok {
		t.Error("model_price_stats: sold_stats marker present once ready")
	}
	if !strings.Contains(sum, "已售出 1 件") {
		t.Errorf("model_price_stats summary = %q", sum)
	}

	m, _ = run("get_listing", listingIn)
	if m["model_sold_observed"] != float64(1) {
		t.Errorf("get_listing: model_sold_observed = %v, want 1", m["model_sold_observed"])
	}

	m, _ = run("market_overview", `{}`)
	if m["sold_observed"] != float64(1) {
		t.Errorf("market_overview: sold_observed = %v, want 1", m["sold_observed"])
	}
	top = m["top_models"].([]any)[0].(map[string]any)
	if top["sold_observed"] != float64(1) {
		t.Errorf("market_overview: top_models[0].sold_observed = %v, want 1", top["sold_observed"])
	}
}
