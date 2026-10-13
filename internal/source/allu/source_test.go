package allu

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/No3371/kura/internal/crawl"
	"github.com/No3371/kura/internal/listing"
)

func TestTargets(t *testing.T) {
	got := (&Source{Brands: []string{"HERMES", "DIOR", "GOYARD"}}).Targets()
	if !reflect.DeepEqual(got, []string{"HERMES", "GOYARD"}) {
		t.Errorf("Targets = %v", got)
	}
}

func TestListURL(t *testing.T) {
	if got, want := listURL("b-21", 2), "https://allu-official.com/jp/ja/market/b-21/c-58/?seller=1&page=2"; got != want {
		t.Errorf("listURL = %q, want %q", got, want)
	}
}

// TestSweepCap serves the page-1 fixture (total far above one page) for every
// list request, so only the page cap ends the sweep.
func TestSweepCap(t *testing.T) {
	body, err := os.ReadFile("testdata/list_lv.html")
	if err != nil {
		t.Fatal(err)
	}
	var lists atomic.Int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/jp/ja/market/b-21/c-58/") {
			http.NotFound(w, r)
			return
		}
		lists.Add(1)
		w.Write(body)
	}))
	defer srv.Close()
	old := baseURL
	baseURL = srv.URL
	defer func() { baseURL = old }()

	for _, c := range []struct{ maxPages, want int }{{2, 2}, {0, maxListPages}, {50, maxListPages}} {
		lists.Store(0)
		src := &Source{Fetcher: crawl.NewFetcher(0), Brands: []string{"LOUIS VUITTON"}}
		var pages []int
		complete, err := src.Sweep(context.Background(), "LOUIS VUITTON", c.maxPages, func(p int, _ []listing.Listing, _ []string) error {
			pages = append(pages, p)
			return nil
		})
		if err != nil || complete {
			t.Fatalf("maxPages %d: complete %v err %v; want incomplete, no error", c.maxPages, complete, err)
		}
		if int(lists.Load()) != c.want || len(pages) != c.want || pages[len(pages)-1] != c.want {
			t.Errorf("maxPages %d: %d requests, visited %v; want %d", c.maxPages, lists.Load(), pages, c.want)
		}
	}
}

// serveBase points baseURL at an httptest server for the test's duration.
func serveBase(t *testing.T, h http.HandlerFunc) {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	old := baseURL
	baseURL = srv.URL
	t.Cleanup(func() { baseURL = old })
}

func TestSweepComplete(t *testing.T) {
	serveBase(t, func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/jp/ja/market/b-4/c-58/") {
			http.NotFound(w, r)
			return
		}
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		sold := row(100+page, func(r map[string]any) { r["stock"], r["sale_status"] = 0, 2 })
		io.WriteString(w, listPayload(map[string]any{
			"total_count": 4, "page": page, "limit": 2, "items": []any{row(page, nil), sold},
		}))
	})
	var pages []int
	src := &Source{Fetcher: crawl.NewFetcher(0)}
	complete, err := src.Sweep(context.Background(), "HERMES", 0, func(p int, cards []listing.Listing, sold []string) error {
		if len(cards) != 1 || cards[0].SourceID != strconv.Itoa(p) {
			t.Errorf("page %d: cards %+v", p, cards)
		}
		if want := []string{strconv.Itoa(100 + p)}; !reflect.DeepEqual(sold, want) {
			t.Errorf("page %d: sold %v, want %v", p, sold, want)
		}
		pages = append(pages, p)
		return nil
	})
	if err != nil || !complete || !reflect.DeepEqual(pages, []int{1, 2}) {
		t.Errorf("complete %v err %v pages %v; want complete after pages 1, 2", complete, err, pages)
	}
}

func TestSweepEmptyPage(t *testing.T) {
	serveBase(t, func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/jp/ja/market/b-4/c-58/") {
			http.NotFound(w, r)
			return
		}
		io.WriteString(w, listPayload(map[string]any{"total_count": 60, "page": 1, "limit": 30, "items": []any{}}))
	})
	src := &Source{Fetcher: crawl.NewFetcher(0)}
	complete, err := src.Sweep(context.Background(), "HERMES", 0, func(int, []listing.Listing, []string) error { return nil })
	if err == nil || complete || !strings.Contains(err.Error(), "empty page") {
		t.Errorf("complete %v err %v; want an empty-page error, not complete", complete, err)
	}
}

func TestDetail(t *testing.T) {
	bag, err := os.ReadFile("testdata/detail_bag.html")
	if err != nil {
		t.Fatal(err)
	}
	sold, err := os.ReadFile("testdata/detail_sold.html")
	if err != nil {
		t.Fatal(err)
	}
	serveBase(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/p/bag":
			w.Write(bag)
		case "/p/sold":
			w.Write(sold)
		case "/p/403":
			w.WriteHeader(http.StatusForbidden)
		case "/p/500":
			w.WriteHeader(http.StatusInternalServerError)
		case "/p/202":
			w.WriteHeader(http.StatusAccepted)
			w.Write(bag)
		default:
			http.NotFound(w, r)
		}
	})
	f := crawl.NewFetcher(0)
	f.Retries = 0 // a 5xx would otherwise back off for seconds
	src := &Source{Fetcher: f}
	for _, c := range []struct {
		path      string
		gone, err bool
	}{
		{"/p/bag", false, false},
		{"/p/sold", true, false},
		{"/p/404", true, false},
		{"/p/403", false, true},
		{"/p/202", false, true},
		{"/p/500", false, true}, // last: the fetcher schedules its 2 s backoff even with no retry left
	} {
		d, gone, err := src.Detail(context.Background(), baseURL+c.path)
		if gone != c.gone || (err != nil) != c.err {
			t.Errorf("%s: gone %v err %v; want gone %v, error %v", c.path, gone, err, c.gone, c.err)
		}
		if c.path == "/p/bag" && d.ModelCode != "M15242" {
			t.Errorf("%s: detail %+v", c.path, d)
		}
	}
}

func TestSweepUnknownBrand(t *testing.T) {
	if _, err := (&Source{}).Sweep(context.Background(), "DIOR", 1, nil); err == nil {
		t.Fatal("want error for a brand without code")
	}
}
