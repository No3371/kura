package web

import (
	"context"
	"encoding/json"
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

func TestDailyPick(t *testing.T) {
	listing.RegisterScheme(listing.Scheme{Source: "dt", Name: "テスト", Shop: "テスト店", Graded: "店家鑑定",
		Grades: []listing.Grade{
			{Code: "新品", Label: "新品", Rank: listing.RankNew},
			{Code: "A", Label: "ランクA", Rank: listing.RankA},
		}})
	st, err := store.Open(filepath.Join(t.TempDir(), "kura.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	ctx := context.Background()
	t0 := time.Date(2026, 10, 14, 1, 0, 0, 0, time.UTC) // 09:00 in Taipei
	seen := func(id, grade string, rank listing.Rank, price int64, img string, at time.Time) {
		t.Helper()
		l := listing.Listing{Source: "dt", SourceID: id, URL: "u", Brand: "HERMES", Title: id, ModelName: "M",
			ModelCode: "c" + id, Grade: grade, Rank: rank, PriceJPY: price, ImageURL: img}
		if _, err := st.RecordSeen(ctx, "HERMES", l, at); err != nil {
			t.Fatal(err)
		}
	}
	photo := "https://img/x.jpg"
	seen("cheap-a", "A", listing.RankA, 10000, photo, t0)                    // cheaper, but worse condition
	seen("new-dear", "新品", listing.RankNew, 90000, photo, t0)                // best condition, dearer
	seen("new-late", "新品", listing.RankNew, 50000, photo, t0.Add(time.Hour)) // ties new-early on price, listed later
	seen("new-early", "新品", listing.RankNew, 50000, photo, t0)
	seen("new-nophoto", "新品", listing.RankNew, 20000, "https://img/NowPrinting_L.jpg", t0) // no shop photo yet

	now := t0.Add(2 * time.Hour)
	fxc := &fx.Converter{Fallback: 0.2}
	h := (&Server{Store: st, Radar: &radar.Engine{Store: st, FX: fxc, Now: func() time.Time { return now }}, FX: fxc,
		Log: slog.New(slog.NewTextHandler(io.Discard, nil))}).Handler()
	get := func() dailyPick {
		t.Helper()
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/daily", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("code %d: %s", rec.Code, rec.Body)
		}
		var p dailyPick
		if err := json.Unmarshal(rec.Body.Bytes(), &p); err != nil {
			t.Fatal(err)
		}
		return p
	}

	p := get()
	if p.Day != "2026-10-14" || p.Item == nil || p.Item.ID != "dt:new-early" || p.Item.Status != "active" {
		t.Fatalf("first pick = %+v", p)
	}

	// Sold during its day: still the day's pick, now marked sold.
	if _, err := st.MarkGone(ctx, "dt:new-early", now.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	now = now.Add(2 * time.Hour)
	if p = get(); p.Item == nil || p.Item.ID != "dt:new-early" || p.Item.Status == "active" {
		t.Fatalf("pick after sale = %+v", p.Item)
	}

	// The next Taipei day picks again from what is still in sale.
	now = time.Date(2026, 10, 14, 16, 30, 0, 0, time.UTC) // 00:30 on the 15th in Taipei
	if p = get(); p.Day != "2026-10-15" || p.Item == nil || p.Item.ID != "dt:new-late" {
		t.Fatalf("next day's pick = %+v", p)
	}
}
