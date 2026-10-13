package store

import (
	"context"
	"database/sql"
	"path/filepath"
	"slices"
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

func TestMarkGoneOnlyActive(t *testing.T) {
	ctx := context.Background()
	s := open(t)
	day := func(n int) time.Time { return time.Date(2026, 10, n, 0, 0, 0, 0, time.UTC) }
	if _, err := s.RecordSeen(ctx, "HERMES", item("a", 100), day(1)); err != nil {
		t.Fatal(err)
	}
	if ok, err := s.MarkGone(ctx, "src:a", day(2)); err != nil || !ok {
		t.Fatalf("first MarkGone = %v, %v; want true", ok, err)
	}
	if ok, err := s.MarkGone(ctx, "src:a", day(3)); err != nil || ok {
		t.Fatalf("repeated MarkGone = %v, %v; want false", ok, err)
	}
	if ok, err := s.MarkGone(ctx, "src:nosuch", day(2)); err != nil || ok {
		t.Fatalf("MarkGone of an unknown id = %v, %v; want false, nil", ok, err)
	}
	rows, err := s.AllRows(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].Status != "gone" || !rows[0].GoneAt.Equal(day(2)) {
		t.Fatalf("rows = %+v, want one gone row with gone_at day 2", rows)
	}

	// Relisted: active again, and its product page is still to be read.
	if _, err := s.RecordSeen(ctx, "HERMES", item("a", 100), day(4)); err != nil {
		t.Fatal(err)
	}
	todo, err := s.PendingDetails(ctx, "src", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(todo) != 1 || todo[0].ID != "src:a" {
		t.Errorf("pending after relisting = %v, want src:a", todo)
	}
}

func TestRecheckDue(t *testing.T) {
	ctx := context.Background()
	s := open(t)
	day := func(n int) time.Time { return time.Date(2026, 10, n, 0, 0, 0, 0, time.UTC) }
	seen := func(target, id string, at time.Time) {
		t.Helper()
		l := item(id, 100)
		l.Brand = target
		if _, err := s.RecordSeen(ctx, target, l, at); err != nil {
			t.Fatal(err)
		}
	}
	read := func(id string, at time.Time) {
		t.Helper()
		if err := s.SaveDetail(ctx, "src:"+id, listing.Detail{Color: "c"}, at); err != nil {
			t.Fatal(err)
		}
	}
	seen("HERMES", "a", day(1))
	read("a", day(1))
	seen("HERMES", "b", day(1))
	read("b", day(2))
	seen("HERMES", "c", day(1)) // never read: PendingDetails'
	seen("HERMES", "d", day(1))
	read("d", day(1))
	seen("HERMES", "d", day(3)) // shown by a list page since seenBefore
	seen("CHANEL", "e", day(1))
	read("e", day(1))
	seen("HERMES", "f", day(1))
	read("f", day(1))
	if _, err := s.MarkGone(ctx, "src:f", day(2)); err != nil {
		t.Fatal(err)
	}
	seen("HERMES", "g", day(1))
	read("g", day(3)) // read too recently

	due := func(targets []string, limit int) []string {
		t.Helper()
		todo, err := s.RecheckDue(ctx, "src", targets, day(3), day(3), limit)
		if err != nil {
			t.Fatal(err)
		}
		var ids []string
		for _, d := range todo {
			ids = append(ids, d.ID)
		}
		return ids
	}
	for _, c := range []struct {
		targets []string
		limit   int
		want    []string
	}{
		{[]string{"HERMES"}, 10, []string{"src:a", "src:b"}},
		{[]string{"HERMES"}, 1, []string{"src:a"}},
		{[]string{"HERMES", "CHANEL"}, 10, []string{"src:a", "src:e", "src:b"}},
		{nil, 10, nil},
	} {
		if got := due(c.targets, c.limit); !slices.Equal(got, c.want) {
			t.Errorf("RecheckDue(%v, limit %d) = %v, want %v", c.targets, c.limit, got, c.want)
		}
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

func TestSetImages(t *testing.T) {
	ctx := context.Background()
	s := open(t)
	a, b := item("a", 100), item("b", 100)
	for _, l := range []listing.Listing{a, b} {
		if _, err := s.RecordSeen(ctx, "HERMES", l, time.Now()); err != nil {
			t.Fatal(err)
		}
	}
	images := func() map[string][]string {
		t.Helper()
		rows, err := s.AllRows(ctx)
		if err != nil {
			t.Fatal(err)
		}
		m := map[string][]string{}
		for _, r := range rows {
			m[r.ID] = r.Images
		}
		return m
	}
	version := func() string {
		t.Helper()
		v, err := s.Version(ctx)
		if err != nil {
			t.Fatal(err)
		}
		return v
	}

	v0 := version()
	if err := s.SetImages(ctx, a.ID(), []string{"p1", " ", "p2", "p1", "p3"}); err != nil {
		t.Fatal(err)
	}
	if got := images(); !slices.Equal(got[a.ID()], []string{"p1", "p2", "p3"}) || got[b.ID()] != nil {
		t.Errorf("after set: %v", got)
	}
	v1 := version()
	if v1 == v0 {
		t.Error("Version unchanged after SetImages")
	}

	// Same count, different photos: Version must still move.
	if err := s.SetImages(ctx, a.ID(), []string{"q1", "q2", "q3"}); err != nil {
		t.Fatal(err)
	}
	if got := images()[a.ID()]; !slices.Equal(got, []string{"q1", "q2", "q3"}) {
		t.Errorf("after replace: %v", got)
	}
	v2 := version()
	if v2 == v1 {
		t.Error("Version unchanged after same-size replace")
	}

	if err := s.SetImages(ctx, a.ID(), nil); err != nil {
		t.Fatal(err)
	}
	if got := images()[a.ID()]; got != nil {
		t.Errorf("after clear: %v", got)
	}
	if version() == v2 {
		t.Error("Version unchanged after clear")
	}
}

func TestSaveDetailImages(t *testing.T) {
	ctx := context.Background()
	s := open(t)
	l := item("a", 100)
	if _, err := s.RecordSeen(ctx, "HERMES", l, time.Now()); err != nil {
		t.Fatal(err)
	}
	stored := func() Row {
		t.Helper()
		rows, err := s.AllRows(ctx)
		if err != nil {
			t.Fatal(err)
		}
		for _, r := range rows {
			if r.ID == l.ID() {
				return r
			}
		}
		t.Fatalf("no row %s", l.ID())
		return Row{}
	}
	save := func(d listing.Detail) {
		t.Helper()
		if err := s.SaveDetail(ctx, l.ID(), d, time.Now()); err != nil {
			t.Fatal(err)
		}
	}

	save(listing.Detail{Color: "c", Images: []string{"p1", "p2"}})
	if r := stored(); !slices.Equal(r.Images, []string{"p1", "p2"}) || r.Color != "c" {
		t.Errorf("first read: images %v color %q", r.Images, r.Color)
	}
	// A read that yields no photos keeps the gallery but still saves fields.
	save(listing.Detail{Color: "d"})
	if r := stored(); !slices.Equal(r.Images, []string{"p1", "p2"}) || r.Color != "d" {
		t.Errorf("photo-less re-read: images %v color %q", r.Images, r.Color)
	}
	// A read with photos replaces it.
	save(listing.Detail{Color: "d", Images: []string{"q1"}})
	if r := stored(); !slices.Equal(r.Images, []string{"q1"}) {
		t.Errorf("re-read with photos: images %v", r.Images)
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

func TestOnboarded(t *testing.T) {
	ctx := context.Background()
	s := open(t)
	day := func(n int) time.Time { return time.Date(2026, 10, n, 0, 0, 0, 0, time.UTC) }
	sweep := func(source, target string, at time.Time, complete bool) {
		t.Helper()
		id, err := s.StartRun(ctx, source, target, at.Add(-time.Hour))
		if err != nil {
			t.Fatal(err)
		}
		if err := s.FinishRun(ctx, Run{ID: id, FinishedAt: at, Complete: complete}); err != nil {
			t.Fatal(err)
		}
	}
	onboarded := func() map[Pair]time.Time {
		t.Helper()
		m, err := s.Onboarded(ctx)
		if err != nil {
			t.Fatal(err)
		}
		return m
	}
	version := func() string {
		t.Helper()
		v, err := s.Version(ctx)
		if err != nil {
			t.Fatal(err)
		}
		return v
	}
	hermes := Pair{"src", "HERMES"}

	if m := onboarded(); len(m) != 0 {
		t.Fatalf("empty store onboarded = %v", m)
	}
	v0 := version()
	sweep("src", "HERMES", day(1), false) // capped or failed: saw part of the catalogue
	if m := onboarded(); len(m) != 0 {
		t.Fatalf("after incomplete sweep onboarded = %v", m)
	}
	if v := version(); v != v0 {
		t.Errorf("incomplete sweep changed version %q -> %q", v0, v)
	}
	sweep("src", "HERMES", day(2), true)
	if got := onboarded()[hermes]; !got.Equal(day(2)) {
		t.Errorf("onboarded HERMES = %v, want %v", got, day(2))
	}
	if v := version(); v == v0 {
		t.Errorf("complete sweep left version at %q", v)
	}
	sweep("src", "HERMES", day(3), true)
	if got := onboarded()[hermes]; !got.Equal(day(2)) {
		t.Errorf("onboarded HERMES after later sweep = %v, want first %v", got, day(2))
	}
	if _, err := s.StartRun(ctx, "src", "CHANEL", day(3)); err != nil { // process killed mid-sweep
		t.Fatal(err)
	}
	sweep("other", "HERMES", day(4), true)
	m := onboarded()
	if _, ok := m[Pair{"src", "CHANEL"}]; ok {
		t.Errorf("unfinished sweep counted as onboarding: %v", m)
	}
	if got := m[Pair{"other", "HERMES"}]; !got.Equal(day(4)) || len(m) != 2 {
		t.Errorf("onboarded = %v, want src/HERMES day 2 and other/HERMES day 4", m)
	}
}
