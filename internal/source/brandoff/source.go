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

const (
	baseURL  = "https://www.brandoff-store.com"
	pageSize = 90 // the largest page size the site offers
)

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

func (s *Source) Sweep(ctx context.Context, target string, maxPages int, visit func(int, []listing.Listing) error) (bool, error) {
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
		if err := visit(page, bags); err != nil {
			return false, err
		}
		if len(lp.Cards) < pageSize || offset+len(lp.Cards) >= total {
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
	d, err := ParseDetail(bytes.NewReader(resp.Body))
	return d, false, err
}
