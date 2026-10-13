package brandoff

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"sort"
	"testing"

	"github.com/No3371/kura/internal/crawl"
	"github.com/No3371/kura/internal/listing"
)

// TestSweepSoldTiles serves the LV fixture (90 tiles, 7 SOLDOUT, thousands of
// hits) as page 1 and the past-the-end fixture as page 2. The sweep must read
// page 1 as full, pass its sold bag ids, and complete only on page 2.
func TestSweepSoldTiles(t *testing.T) {
	full, err := os.ReadFile("testdata/list_lv.html")
	if err != nil {
		t.Fatal(err)
	}
	empty, err := os.ReadFile("testdata/list_empty.html")
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/Form/Product/ProductList.aspx" {
			http.NotFound(w, r)
			return
		}
		if r.URL.Query().Get("o") == "" {
			w.Write(full)
			return
		}
		w.Write(empty)
	}))
	defer srv.Close()
	old := baseURL
	baseURL = srv.URL
	defer func() { baseURL = old }()

	src := &Source{Fetcher: crawl.NewFetcher(0)}
	var pages []int
	var sold []string
	complete, err := src.Sweep(context.Background(), "LOUIS VUITTON", 0, func(p int, bags []listing.Listing, s []string) error {
		pages = append(pages, p)
		sold = append(sold, s...)
		for _, l := range bags {
			for _, id := range s {
				if l.SourceID == id {
					t.Errorf("page %d: %s both listed and sold", p, id)
				}
			}
		}
		return nil
	})
	if err != nil || !complete || !reflect.DeepEqual(pages, []int{1, 2}) {
		t.Fatalf("complete %v err %v pages %v; want complete after pages 1, 2", complete, err, pages)
	}
	sort.Strings(sold)
	want := []string{"2100301259737", "2100301266889", "2101220306854", "2101220320522", "2107601195543", "2118400254263"}
	if !reflect.DeepEqual(sold, want) {
		t.Errorf("sold %v, want the fixture's 6 SOLDOUT bags %v", sold, want)
	}
}
