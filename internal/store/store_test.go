package store

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"github.com/No3371/kura/internal/listing"
)

func open(t *testing.T) *Store {
	t.Helper()
	s, err := Open(filepath.Join(t.TempDir(), "kura.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func item(id string, price int64) listing.Listing {
	return listing.Listing{Source: "src", SourceID: id, URL: "u", Brand: "HERMES", Title: id,
		ModelCode: "c1", Rank: listing.RankA, PriceJPY: price}
}

func TestPriceHistoryAndGoneLifecycle(t *testing.T) {
	ctx := context.Background()
	s := open(t)
	day := func(n int) time.Time { return time.Date(2026, 10, n, 0, 0, 0, 0, time.UTC) }

	// Day 1: two items seen.
	for _, l := range []listing.Listing{item("a", 100), item("b", 200)} {
		if res, err := s.RecordSeen(ctx, "HERMES", l, day(1)); err != nil || !res.New {
			t.Fatalf("first sight: %+v %v", res, err)
		}
	}
	// Day 2: a drops in price, b is missing (first miss: not yet gone).
	res, err := s.RecordSeen(ctx, "HERMES", item("a", 90), day(2))
	if err != nil || res.New || res.OldPrice != 100 {
		t.Fatalf("price change: %+v %v", res, err)
	}
	if n, _ := s.SweepMissing(ctx, "src", "HERMES", day(2), day(2)); n != 0 {
		t.Fatalf("gone after one miss = %d, want 0", n)
	}
	// Day 3: b missing again -> gone. Same price for a -> no new point.
	if _, err := s.RecordSeen(ctx, "HERMES", item("a", 90), day(3)); err != nil {
		t.Fatal(err)
	}
	if n, _ := s.SweepMissing(ctx, "src", "HERMES", day(3), day(3)); n != 1 {
		t.Fatalf("gone after two misses = %d, want 1", n)
	}

	rows, err := s.AllRows(ctx)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]Row{}
	for _, r := range rows {
		got[r.SourceID] = r
	}
	if a := got["a"]; a.Status != "active" || len(a.PriceHistory) != 2 || a.PriceHistory[1].PriceJPY != 90 {
		t.Errorf("a = %+v", a)
	}
	if b := got["b"]; b.Status != "gone" || !b.GoneAt.Equal(day(3)) {
		t.Errorf("b = %+v", b)
	}

	// Day 4: b reappears (relisted / sweep artefact) -> active again.
	if _, err := s.RecordSeen(ctx, "HERMES", item("b", 200), day(4)); err != nil {
		t.Fatal(err)
	}
	rows, _ = s.AllRows(ctx)
	for _, r := range rows {
		if r.SourceID == "b" && (r.Status != "active" || !r.GoneAt.IsZero()) {
			t.Errorf("b after reappearing = %+v", r)
		}
	}
}

func TestPendingDetailsPriority(t *testing.T) {
	ctx := context.Background()
	s := open(t)
	now := time.Now()
	add := func(target, id, code string) {
		l := item(id, 100)
		l.ModelCode = code
		if _, err := s.RecordSeen(ctx, target, l, now); err != nil {
			t.Fatal(err)
		}
	}
	add("LV", "lv1", "M1")
	add("LV", "lv2", "M2")
	add("LV", "lv3", "M3")
	add("HERMES", "h1", "")
	add("CHANEL", "c1", "A1")
	todo, err := s.PendingDetails(ctx, "src", 3)
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, d := range todo {
		got = append(got, d.ID)
	}
	// Code-less first, then one per target before a second LV item.
	if len(got) != 3 || got[0] != "src:h1" || got[1] == got[2] ||
		(got[1] != "src:c1" && got[2] != "src:c1") {
		t.Errorf("order = %v", got)
	}
}

func TestMigrateAddsGradeAndBackfills(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "old.db")
	// A database from before per-source grades: no grade column.
	old, err := sql.Open("sqlite", "file:"+path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := old.Exec(`CREATE TABLE listings (id TEXT PRIMARY KEY, source TEXT NOT NULL, source_id TEXT NOT NULL,
		target TEXT NOT NULL, url TEXT NOT NULL, image_url TEXT NOT NULL DEFAULT '', brand TEXT NOT NULL, title TEXT NOT NULL,
		model_name TEXT NOT NULL DEFAULT '', model_code TEXT NOT NULL DEFAULT '', group_key TEXT NOT NULL DEFAULT '',
		category TEXT NOT NULL DEFAULT 'bag', gender TEXT NOT NULL DEFAULT '', rank TEXT NOT NULL DEFAULT '',
		badges TEXT NOT NULL DEFAULT '', price_jpy INTEGER NOT NULL, was_price_jpy INTEGER NOT NULL DEFAULT 0,
		color TEXT NOT NULL DEFAULT '', material TEXT NOT NULL DEFAULT '', size TEXT NOT NULL DEFAULT '',
		accessories TEXT NOT NULL DEFAULT '', notes TEXT NOT NULL DEFAULT '', status TEXT NOT NULL DEFAULT 'active',
		first_seen INTEGER NOT NULL, last_seen INTEGER NOT NULL, gone_at INTEGER, missed_sweeps INTEGER NOT NULL DEFAULT 0,
		detail_at INTEGER);
		INSERT INTO listings (id, source, source_id, target, url, brand, title, rank, price_jpy, first_seen, last_seen)
		VALUES ('src:1','src','1','HERMES','u','HERMES','x','A',100,1,1), ('src:2','src','2','HERMES','u','HERMES','y','B',90,1,1);`); err != nil {
		t.Fatal(err)
	}
	old.Close()

	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	scheme := listing.Scheme{Source: "src", Grades: []listing.Grade{
		{Code: "ランクA", Rank: listing.RankA},
		{Code: "ランクB", Rank: listing.RankB}, {Code: "ランクB+", Rank: listing.RankB}, // ambiguous: left alone
	}}
	n, err := s.BackfillGrades(ctx, scheme)
	if err != nil || n != 1 {
		t.Fatalf("backfilled %d rows, err %v; want 1", n, err)
	}
	rows, _ := s.AllRows(ctx)
	got := map[string]string{}
	for _, r := range rows {
		got[r.SourceID] = r.Grade
	}
	if got["1"] != "ランクA" || got["2"] != "" {
		t.Errorf("grades = %v", got)
	}
}

func TestSaveDetailOverridesModelCode(t *testing.T) {
	ctx := context.Background()
	s := open(t)
	l := item("a", 100)
	l.ModelCode = ""
	l.ModelName = "バーキン30"
	if _, err := s.RecordSeen(ctx, "HERMES", l, time.Now()); err != nil {
		t.Fatal(err)
	}
	if todo, _ := s.PendingDetails(ctx, "src", 10); len(todo) != 1 {
		t.Fatalf("pending = %v", todo)
	}
	if err := s.SaveDetail(ctx, l.ID(), listing.Detail{ModelCode: "h041344ck", Color: "Gold"}, time.Now()); err != nil {
		t.Fatal(err)
	}
	rows, _ := s.AllRows(ctx)
	if r := rows[0]; r.ModelCode != "h041344ck" || r.Color != "Gold" || !r.HasDetail {
		t.Errorf("row = %+v", r)
	}
	if todo, _ := s.PendingDetails(ctx, "src", 10); len(todo) != 0 {
		t.Errorf("still pending after SaveDetail: %v", todo)
	}
}
