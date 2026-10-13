package web

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/No3371/kura/internal/advisor"
	"github.com/No3371/kura/internal/fx"
	"github.com/No3371/kura/internal/listing"
	"github.com/No3371/kura/internal/radar"
	"github.com/No3371/kura/internal/store"
)

func TestStaticRoutes(t *testing.T) {
	// Static routes never reach the store, radar or advisor.
	h := (&Server{}).Handler()
	cases := []struct {
		path, ctype, body string
		code              int
	}{
		{"/", "text/html", `id="main"`, 200},
		{"/browse", "text/html", `id="main"`, 200},
		{"/listing/brandoff:1", "text/html", `id="main"`, 200},
		{"/", "text/html", "找包顧問", 200},
		{"/js/app.js", "javascript", "", 200},
		{"/js/advisor.js", "javascript", "initAdvisor", 200},
		{"/js/member.js", "javascript", "signIn", 200},
		{"/js/notify.js", "javascript", "initNotify", 200},
		{"/js/pages.js", "javascript", "renderMethodology", 200},
		{"/methodology", "text/html", `id="main"`, 200},
		{"/about", "text/html", `id="main"`, 200},
		{"/", "text/html", "k-member", 200},
		{"/kura.css", "text/css", "--k-paper", 200},
		{"/legacy/", "text/html", "AI 找包顧問", 200},
		{"/model/191502BZJ.04LU", "text/html", `id="main"`, 200},
		{"/model/CHANEL/name%3A%E3%83%9E%E3%83%88%E3%83%A9%E3%83%83%E3%82%BB", "text/html", `id="main"`, 200},
		{"/model/HERMES/name%3A24%2F24", "text/html", `id="main"`, 200},
		{"/dev/cards.html", "text/html", "", 200},
		{"/js/missing.js", "", "", 404},
		{"/js/missing", "", "", 404},
		{"/missing.png", "", "", 404},
		{"/js/", "", "", 404},
		{"/dev/", "", "", 404},
		{"/legacy", "", "", 301},
		{"/api/nope", "", "", 404},
	}
	for _, c := range cases {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, c.path, nil))
		if rec.Code != c.code {
			t.Errorf("%s: code %d, want %d", c.path, rec.Code, c.code)
			continue
		}
		if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, c.ctype) {
			t.Errorf("%s: Content-Type %q, want it to contain %q", c.path, ct, c.ctype)
		}
		if !strings.Contains(rec.Body.String(), c.body) {
			t.Errorf("%s: body lacks %q", c.path, c.body)
		}
	}
}

func TestOverviewMethod(t *testing.T) {
	listing.RegisterScheme(listing.Scheme{Source: "wo", Name: "テスト", Shop: "テスト店", Graded: "店家鑑定",
		Grades: []listing.Grade{{Code: "A", Label: "ランクA", Short: "A", Rank: listing.RankA}}})
	st, err := store.Open(filepath.Join(t.TempDir(), "kura.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	fxc := &fx.Converter{Fallback: 0.2}
	h := (&Server{Store: st, Radar: &radar.Engine{Store: st, FX: fxc}, FX: fxc, Advisor: &advisor.Mock{},
		Log: slog.New(slog.NewTextHandler(io.Discard, nil))}).Handler()

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/overview", nil))
	if rec.Code != 200 {
		t.Fatalf("overview: code %d: %s", rec.Code, rec.Body)
	}
	var got struct {
		Method struct {
			DealMinPct         float64 `json:"deal_min_pct"`
			AnomalyPct         float64 `json:"anomaly_pct"`
			MinComps           int     `json:"min_comps"`
			MinSoldHistoryDays int     `json:"min_sold_history_days"`
		} `json:"method"`
		Schemes []listing.Scheme `json:"schemes"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	m := got.Method
	if m.DealMinPct != radar.DealMinPct || m.AnomalyPct != radar.AnomalyPct || m.MinComps != radar.MinComps ||
		m.MinSoldHistoryDays != int(radar.MinSoldHistory/(24*time.Hour)) {
		t.Errorf("method = %+v, want the radar's constants", m)
	}
	// The scheme registry is process-wide (other tests register theirs), so
	// assert membership rather than the whole list.
	found := false
	for _, s := range got.Schemes {
		found = found || (s.Source == "wo" && s.Shop == "テスト店" && len(s.Grades) == 1)
	}
	if !found {
		t.Errorf("schemes lack source wo: %+v", got.Schemes)
	}
}

func TestListingAPI(t *testing.T) {
	listing.RegisterScheme(listing.Scheme{Source: "wt", Name: "テスト", Shop: "テスト店", Graded: "店家鑑定",
		Grades: []listing.Grade{{Code: "A", Label: "ランクA", Short: "A", Rank: listing.RankA}}})
	st, err := store.Open(filepath.Join(t.TempDir(), "kura.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	l := listing.Listing{Source: "wt", SourceID: "1", URL: "u", Brand: "HERMES", Title: "t",
		ModelCode: "c1", Grade: "A", Rank: listing.RankA, PriceJPY: 100000}
	if _, err := st.RecordSeen(context.Background(), "HERMES", l, time.Now()); err != nil {
		t.Fatal(err)
	}
	l2 := l
	l2.SourceID, l2.ModelCode = "2", "c/2"
	if _, err := st.RecordSeen(context.Background(), "HERMES", l2, time.Now()); err != nil {
		t.Fatal(err)
	}
	fxc := &fx.Converter{Fallback: 0.2}
	h := (&Server{Store: st, Radar: &radar.Engine{Store: st, FX: fxc}, FX: fxc,
		Log: slog.New(slog.NewTextHandler(io.Discard, nil))}).Handler()

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/listings/wt:1", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("code %d: %s", rec.Code, rec.Body)
	}
	var body struct {
		Item   radar.Item        `json:"item"`
		Model  *radar.ModelStats `json:"model"`
		Scheme *listing.Scheme   `json:"scheme"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Scheme == nil || body.Scheme.Shop != "テスト店" || body.Scheme.Grades[0].Short != "A" {
		t.Errorf("scheme = %+v", body.Scheme)
	}
	if h := body.Item.History; len(h) != 1 || h[0].PriceTWD != 20000 {
		t.Errorf("history = %+v", h)
	}
	if body.Model == nil || len(body.Model.ByGrade) != 1 || body.Model.ByGrade[0].MedianTWD != 20000 {
		t.Errorf("model = %+v", body.Model)
	}

	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/listings/wt:nope", nil))
	if rec.Code != http.StatusNotFound {
		t.Errorf("unknown id: code %d, want 404", rec.Code)
	}

	getList := func(qs string) (total int, ids []string) {
		t.Helper()
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/listings?"+qs, nil))
		var lb struct {
			Total int          `json:"total"`
			Items []radar.Item `json:"items"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &lb); err != nil {
			t.Fatal(err)
		}
		for _, it := range lb.Items {
			ids = append(ids, it.ID)
		}
		return lb.Total, ids
	}
	_, first := getList("sort=newest&limit=1")
	if total, second := getList("sort=newest&limit=1&offset=1"); total != 2 || len(second) != 1 || len(first) != 1 || second[0] == first[0] {
		t.Errorf("offset 1: total %d, pages %v / %v", total, first, second)
	}
	if total, past := getList("offset=9"); total != 2 || len(past) != 0 {
		t.Errorf("offset past end: total %d, items %v", total, past)
	}

	type modelBody struct {
		Model  radar.ModelStats `json:"model"`
		Items  []radar.Item     `json:"items"`
		Scheme *listing.Scheme  `json:"scheme"`
	}
	getModel := func(path string) (int, modelBody) {
		t.Helper()
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		var mb modelBody
		if rec.Code == http.StatusOK {
			if err := json.Unmarshal(rec.Body.Bytes(), &mb); err != nil {
				t.Fatal(err)
			}
		}
		return rec.Code, mb
	}
	if code, mb := getModel("/api/models/HERMES/C1"); code != http.StatusOK {
		t.Errorf("model C1: code %d, want 200", code)
	} else if mb.Scheme == nil || mb.Scheme.Shop != "テスト店" || len(mb.Items) != 1 || mb.Model.Active != 1 {
		t.Errorf("model C1: scheme %+v, %d items, active %d", mb.Scheme, len(mb.Items), mb.Model.Active)
	}
	// An escaped "/" stays inside the key rather than splitting the path.
	if code, mb := getModel("/api/models/HERMES/C%2F2"); code != http.StatusOK {
		t.Errorf("model C/2: code %d, want 200", code)
	} else if len(mb.Items) != 1 || mb.Items[0].ID != "wt:2" {
		t.Errorf("model C/2: items %+v", mb.Items)
	}
	if code, _ := getModel("/api/models/HERMES/NOPE"); code != http.StatusNotFound {
		t.Errorf("unknown model: code %d, want 404", code)
	}
}

func TestMemberAPI(t *testing.T) {
	listing.RegisterScheme(listing.Scheme{Source: "wm", Name: "テスト", Shop: "テスト店", Graded: "店家鑑定",
		Grades: []listing.Grade{{Code: "A", Label: "ランクA", Short: "A", Rank: listing.RankA}}})
	st, err := store.Open(filepath.Join(t.TempDir(), "kura.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	ctx := context.Background()
	// t0 is the initial crawl; listings first seen a day or more later are
	// ones Kura watched appear.
	t0 := time.Now().Add(-30 * time.Hour).Truncate(time.Second)
	seenAs := func(brand, id, code string, price int64, at time.Time) {
		t.Helper()
		l := listing.Listing{Source: "wm", SourceID: id, URL: "u", Brand: brand, Title: "t" + id,
			ModelName: "アルマPM", ModelCode: code, Grade: "A", Rank: listing.RankA, PriceJPY: price}
		if _, err := st.RecordSeen(ctx, brand, l, at); err != nil {
			t.Fatal(err)
		}
	}
	seen := func(id, code string, price int64, at time.Time) { seenAs("LOUIS VUITTON", id, code, price, at) }
	// m1: three comparables at market, then one well below it that appeared
	// after the initial crawl.
	seen("c1", "m1", 118000, t0)
	seen("c2", "m1", 120000, t0)
	seen("c3", "m1", 122000, t0)
	seen("new", "m1", 90000, t0.Add(25*time.Hour))
	// m2: one listing observed at 120k, then cut to 100k.
	seen("cut", "m2", 120000, t0)
	seen("cut", "m2", 100000, t0.Add(26*time.Hour))
	// h1: initial crawl only, paged newest first — hnew is on the first page
	// and is the shop's newest; hold, crawled later, is older stock.
	hermes := func(id string, price int64, at time.Time) { seenAs("HERMES", id, "h1", price, at) }
	hermes("hnew", 90000, t0)
	hermes("h1", 118000, t0.Add(10*time.Minute))
	hermes("h2", 120000, t0.Add(10*time.Minute))
	hermes("h3", 122000, t0.Add(10*time.Minute))
	hermes("hold", 95000, t0.Add(20*time.Minute))

	fxc := &fx.Converter{Fallback: 0.2}
	h := (&Server{Store: st, Radar: &radar.Engine{Store: st, FX: fxc}, FX: fxc,
		Log: slog.New(slog.NewTextHandler(io.Discard, nil))}).Handler()

	type event struct {
		Type string     `json:"type"`
		Key  string     `json:"key"`
		At   time.Time  `json:"at"`
		Item radar.Item `json:"item"`
	}
	var body struct {
		Name      string   `json:"name"`
		Initial   string   `json:"initial"`
		Watchlist []string `json:"watchlist"`
		Events    []event  `json:"events"`
	}
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
	var models []radar.ModelStats
	get("/api/models?brand=LOUIS%20VUITTON", &models)
	m1Key := ""
	for _, m := range models {
		if m.ModelCode == "m1" {
			m1Key = "model:LOUIS VUITTON/" + m.GroupKey
		}
	}
	if m1Key == "" {
		t.Fatalf("models = %+v, no m1", models)
	}
	get("/api/models?brand=HERMES", &models)
	if len(models) != 1 || models[0].ModelCode != "h1" {
		t.Fatalf("HERMES models = %+v", models)
	}
	h1Key := "model:HERMES/" + models[0].GroupKey
	// Every figure an event carries must be the one the listing page shows.
	sameAsListing := func(ev event) {
		t.Helper()
		var lb struct {
			Item radar.Item `json:"item"`
		}
		get("/api/listings/"+ev.Item.ID, &lb)
		a, b := ev.Item, lb.Item
		if a.PriceTWD != b.PriceTWD || a.Status != b.Status ||
			!reflect.DeepEqual(a.Fair, b.Fair) || !reflect.DeepEqual(a.Drop, b.Drop) || len(a.History) != 0 {
			t.Errorf("%s event item differs from /api/listings/%s:\n%+v\n%+v", ev.Type, a.ID, a, b)
		}
	}
	types := func() (out []string) {
		for _, ev := range body.Events {
			out = append(out, ev.Type+" "+ev.Key)
		}
		return out
	}

	get("/api/member", &body)
	if body.Name != "林小姐" || body.Initial != "林" {
		t.Errorf("persona = %q %q", body.Name, body.Initial)
	}
	// The top deal (wm:new) is watched too, but a listing that has neither
	// dropped nor gone has nothing to notify.
	want := []string{m1Key, h1Key, "wm:cut", "wm:new"}
	if !reflect.DeepEqual(body.Watchlist, want) {
		t.Errorf("watchlist = %v, want %v", body.Watchlist, want)
	}
	if got, want := types(), []string{"drop wm:cut", "new " + m1Key, "new " + h1Key}; !reflect.DeepEqual(got, want) {
		t.Fatalf("events = %v, want %v (newest first)", got, want)
	}
	// 新上架 is the newest listing: the one Kura watched appear, else the
	// shop's newest from the initial crawl, the first crawled.
	for i, id := range map[int]string{1: "wm:new", 2: "wm:hnew"} {
		if ev := body.Events[i]; ev.Item.ID != id || !radar.IsDeal(&ev.Item) || !ev.At.Equal(ev.Item.FirstSeen) {
			t.Errorf("new event %d = %s %+v, want %s", i, ev.Item.ID, ev, id)
		}
	}
	if d := body.Events[0].Item.Drop; d == nil || d.FromTWD != 24000 || d.ToTWD != 20000 {
		t.Errorf("drop = %+v", d)
	}
	for _, ev := range body.Events {
		sameAsListing(ev)
	}

	// Once the cut listing sells, it is a 已找到新主人 instead of a 降價.
	if err := st.MarkGone(ctx, "wm:cut", t0.Add(27*time.Hour)); err != nil {
		t.Fatal(err)
	}
	body.Watchlist, body.Events = nil, nil
	get("/api/member", &body)
	if got, want := types(), []string{"gone wm:cut", "new " + m1Key, "new " + h1Key}; !reflect.DeepEqual(got, want) {
		t.Fatalf("after sale: events = %v, want %v", got, want)
	}
	for _, ev := range body.Events {
		sameAsListing(ev)
	}

	// No listings: an empty watchlist and feed, as JSON arrays.
	empty, err := store.Open(filepath.Join(t.TempDir(), "empty.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { empty.Close() })
	h = (&Server{Store: empty, Radar: &radar.Engine{Store: empty, FX: fxc}, FX: fxc,
		Log: slog.New(slog.NewTextHandler(io.Discard, nil))}).Handler()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/member", nil))
	if s := rec.Body.String(); !strings.Contains(s, `"watchlist":[]`) || !strings.Contains(s, `"events":[]`) {
		t.Errorf("empty store: %s", s)
	}
}
