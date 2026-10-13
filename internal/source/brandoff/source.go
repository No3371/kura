package brandoff

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/No3371/kura/internal/crawl"
	"github.com/No3371/kura/internal/listing"
)

const pageSize = 90 // the largest page size the site offers

// baseURL is a variable, not a constant, only so tests can point the sweeper
// at a local server; nothing else assigns it.
var baseURL = "https://www.brandoff-store.com"

// Source sweeps BRAND OFF brand catalogues. Targets are canonical brand names
// (see listing.Brands); the site's search parameter uses the katakana name.
type Source struct {
	Fetcher *crawl.Fetcher
	Brands  []string
}

var _ crawl.Source = (*Source)(nil)

// searchNames are the brand filter values the site uses where they differ from
// the dotted spelling in its titles (taken from the site's brand menu).
var searchNames = map[string]string{
	"LOUIS VUITTON":  "ルイヴィトン",
	"BOTTEGA VENETA": "ボッテガヴェネタ",
}

func (s *Source) Name() string      { return SourceName }
func (s *Source) Targets() []string { return s.Brands }

func listURL(brandJA string, offset int) string {
	q := url.Values{}
	q.Set("s8[]", brandJA)
	q.Set("limit", strconv.Itoa(pageSize))
	q.Set("sort", "Number1") // newest first
	if offset > 0 {
		q.Set("o", strconv.Itoa(offset))
	}
	return baseURL + "/Form/Product/ProductList.aspx?" + q.Encode()
}

// Sweep pages through target's catalogue, newest first, until a short page or
// the site's hit count ends it. Tiles the page marks SOLDOUT (ListPage.Sold)
// go to visit as sold ids, bags only — only bags are ever held. A page's row
// count includes its sold tiles: counting only the cards would read a full
// page holding sold tiles as the last page, report the sweep complete early,
// and let store.SweepMissing count every listing past that page as missed.
func (s *Source) Sweep(ctx context.Context, target string, maxPages int, visit func(int, []listing.Listing, []string) error) (bool, error) {
	brand, ok := listing.BrandByName(target)
	if !ok {
		return false, fmt.Errorf("brandoff: unknown brand %q", target)
	}
	query := brand.JA[0]
	if q, ok := searchNames[brand.Name]; ok {
		query = q
	}

	total := -1
	for page := 1; maxPages == 0 || page <= maxPages; page++ {
		offset := (page - 1) * pageSize
		resp, err := s.Fetcher.Get(ctx, listURL(query, offset))
		if err != nil {
			return false, err
		}
		if resp.Status != http.StatusOK {
			return false, fmt.Errorf("brandoff: list page %d: status %d", page, resp.Status)
		}
		lp, err := ParseList(bytes.NewReader(resp.Body), resp.URL)
		if err != nil {
			return false, err
		}
		if total < 0 {
			total = lp.Total
		}
		var bags []listing.Listing
		for _, c := range lp.Cards {
			if !c.IsBag() {
				continue
			}
			if c.Brand == "" {
				c.Brand = brand.Name
			}
			bags = append(bags, c.Listing)
		}
		var sold []string
		for _, c := range lp.Sold {
			if c.IsBag() {
				sold = append(sold, c.SourceID)
			}
		}
		if err := visit(page, bags, sold); err != nil {
			return false, err
		}
		rows := len(lp.Cards) + len(lp.Sold)
		if rows < pageSize || offset+rows >= total {
			return true, nil
		}
	}
	return false, nil
}

func (s *Source) Detail(ctx context.Context, u string) (listing.Detail, bool, error) {
	resp, err := s.Fetcher.Get(ctx, u)
	if err != nil {
		return listing.Detail{}, false, err
	}
	// Removed products 404 or bounce to the error page.
	if resp.Status == http.StatusNotFound || strings.Contains(resp.URL.Path, "/Form/Error.aspx") {
		return listing.Detail{}, true, nil
	}
	d, err := ParseDetail(bytes.NewReader(resp.Body), resp.URL)
	return d, false, err
}
