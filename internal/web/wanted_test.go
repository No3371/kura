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
	"testing"
	"time"

	"github.com/No3371/kura/internal/fx"
	"github.com/No3371/kura/internal/listing"
	"github.com/No3371/kura/internal/radar"
	"github.com/No3371/kura/internal/store"
)

func TestSeekers(t *testing.T) {
	mon := time.Date(2026, 10, 12, 0, 30, 0, 0, taipei) // Monday, ISO week 42
	sun := time.Date(2026, 10, 18, 23, 30, 0, 0, taipei)
	if a, b := seekers("HERMES", "g", 20, mon), seekers("HERMES", "g", 20, sun); a != b {
		t.Errorf("same week: %d vs %d", a, b)
	}
	// The week turns at Taipei midnight, not UTC: Sunday 23:30 in Taipei is
	// Sunday 15:30 UTC, still week 42.
	if a, b := seekers("HERMES", "g", 20, sun), seekers("HERMES", "g", 20, sun.UTC()); a != b {
		t.Errorf("zone: %d vs %d", a, b)
	}
	// Over many models, a week reshuffles at least one and the weekly factor
	// stays inside its bounds.
	changed := false
	for i := range 50 {
		g := fmt.Sprint("g", i)
		a, b := seekers("CHANEL", g, 16, mon), seekers("CHANEL", g, 16, mon.AddDate(0, 0, 7))
		changed = changed || a != b
		// base = 6 + 4*sqrt(16) = 22; factor 0.55–1.45 → 12–32
		if a < 12 || a > 32 {
			t.Errorf("%s: %d outside 12–32", g, a)
		}
	}
	if !changed {
		t.Error("no model changed across weeks")
	}
}

func TestWantedAPI(t *testing.T) {
	listing.RegisterScheme(listing.Scheme{Source: "wt", Name: "テスト", Shop: "テスト店", Graded: "店家鑑定",
		Grades: []listing.Grade{{Code: "A", Label: "ランクA", Short: "A", Rank: listing.RankA}}})
	st, err := store.Open(filepath.Join(t.TempDir(), "kura.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	ctx := context.Background()
	t0 := time.Now().Add(-30 * time.Hour).Truncate(time.Second)
	seen := func(code, id string, price int64) {
		t.Helper()
		l := listing.Listing{Source: "wt", SourceID: id, URL: "u", Brand: "LOUIS VUITTON", Title: "t" + id,
			ModelName: "M" + code, ModelCode: code, Grade: "A", Rank: listing.RankA, PriceJPY: price,
			ImageURL: "https://img/" + id + ".jpg"}
		if _, err := st.RecordSeen(ctx, "LOUIS VUITTON", l, t0); err != nil {
			t.Fatal(err)
		}
	}
	// Seven models with one to seven listings; the sold-out one never ranks.
	for m := 1; m <= 7; m++ {
		for i := range m {
			seen(fmt.Sprint("m", m), fmt.Sprintf("m%d-%d", m, i), int64(100000+1000*i))
		}
	}
	if _, err := st.MarkGone(ctx, "wt:m1-0", t0.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}

	fxc := &fx.Converter{Fallback: 0.2}
	h := (&Server{Store: st, Radar: &radar.Engine{Store: st, FX: fxc}, FX: fxc,
		Log: slog.New(slog.NewTextHandler(io.Discard, nil))}).Handler()
	get := func(path string, v any) {
		t.Helper()
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("%s: code %d: %s", path, rec.Code, rec.Body)
		}
		if err := json.Unmarshal(rec.Body.Bytes(), v); err != nil {
			t.Fatal(err)
		}
	}

	var shelf []wantedModel
	get("/api/wanted", &shelf)
	if len(shelf) != wantedShelf {
		t.Fatalf("shelf has %d models, want %d", len(shelf), wantedShelf)
	}
	for i, w := range shelf {
		if w.Rank != i+1 || (i > 0 && w.Seekers > shelf[i-1].Seekers) {
			t.Errorf("tile %d: rank %d, seekers %d after %d", i, w.Rank, w.Seekers, shelf[max(i-1, 0)].Seekers)
		}
		if w.GroupKey == "" || w.Active == 0 || w.ModelCode == "m1" {
			t.Errorf("tile %d = %+v: want a model in sale", i, w)
		}
		if w.MedianDaysToSell != nil {
			t.Errorf("tile %d states days to sell before history allows", i)
		}
		if w.ImageURL == "" {
			t.Errorf("tile %d has no photo", i)
		}
		// Real figures and the demand number match the model page.
		var mp struct {
			Model   radar.ModelStats `json:"model"`
			Seekers int              `json:"seekers"`
		}
		get("/api/models/LOUIS%20VUITTON/"+w.GroupKey, &mp)
		if mp.Seekers != w.Seekers || mp.Model.Active != w.Active || fxc.TWD(mp.Model.MedianJPY) != w.MedianTWD {
			t.Errorf("tile %d = %+v, model page %+v seekers %d", i, w, mp.Model, mp.Seekers)
		}
	}
}
