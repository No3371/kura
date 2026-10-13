package grare

import (
	"bytes"
	"context"
	"fmt"
	"net/http"

	"github.com/No3371/kura/internal/crawl"
	"github.com/No3371/kura/internal/listing"
)

const baseURL = "https://www.g-rare.com"

// bagPaths: each brand's bag category. Brands without one (GOYARD lists
// products directly under its brand page, mixed with wallets) are not targeted.
var bagPaths = map[string]string{
	"HERMES":        "/c/fashion/hermes/hermes-bag",
	"CHANEL":        "/c/fashion/chanel/chanel-bag",
	"LOUIS VUITTON": "/c/fashion/louisvuitton/louisvuitton-bag",
	"CELINE":        "/c/fashion/celine/celine-bag",
	"LOEWE":         "/c/fashion/loewe/loewe-bag",
}

// Source sweeps GALLERY RARE brand bag categories. Targets are canonical brand
// names (see listing.Brands).
type Source struct {
	Fetcher *crawl.Fetcher
	Brands  []string
}

var _ crawl.Source = (*Source)(nil)

func (s *Source) Name() string { return SourceName }

// Targets keeps the configured brands that have a bag category, in order.
func (s *Source) Targets() []string {
	var out []string
	for _, b := range s.Brands {
		if _, ok := bagPaths[b]; ok {
			out = append(out, b)
		}
	}
	return out
}

// Sweep reads page 1 of target's bag category only. Pagination and sort exist
// only as query strings, which Kura leaves alone (the shop's robots.txt bars
// them for the bots it names). Rejected: ?page= paging (outside that limit)
// and sitemap enumeration (one product-page fetch per listing per crawl).
// Consequences: order is the shop's default, not newest first as
// crawl.Source.Sweep asks; a brand with more bags than one page never
// completes a sweep, so sweeps never mark its items gone and none count as
// seen appearing. Sold items surface only through Detail: the crawler reads
// each listing's product page once, usually in the crawl that first sees it,
// then re-reads it for listings that left page 1 (crawl.Crawler.RecheckAfter).
// List pages show no sold rows, so visit gets nil. Price changes are seen only
// while an item is on page 1.
// maxPages is ignored: there is only ever one page.
func (s *Source) Sweep(ctx context.Context, target string, maxPages int, visit func(int, []listing.Listing, []string) error) (bool, error) {
	p, ok := bagPaths[target]
	if !ok {
		return false, fmt.Errorf("grare: no bag category for %q", target)
	}
	resp, err := s.Fetcher.Get(ctx, baseURL+p)
	if err != nil {
		return false, err
	}
	if resp.Status != http.StatusOK {
		return false, fmt.Errorf("grare: list %s: status %d", p, resp.Status)
	}
	lp, err := ParseList(bytes.NewReader(resp.Body), resp.URL)
	if err != nil {
		return false, err
	}
	for i := range lp.Cards {
		if lp.Cards[i].Brand == "" {
			lp.Cards[i].Brand = target
		}
	}
	if err := visit(1, lp.Cards, nil); err != nil {
		return false, err
	}
	return lp.Complete(), nil
}

// Detail fetches a product page. A removed product answers 404; a sold one
// keeps its page and shows the shop's out-of-stock notice (ParseDetail).
func (s *Source) Detail(ctx context.Context, u string) (listing.Detail, bool, error) {
	resp, err := s.Fetcher.Get(ctx, u)
	if err != nil {
		return listing.Detail{}, false, err
	}
	if resp.Status == http.StatusNotFound {
		return listing.Detail{}, true, nil
	}
	return ParseDetail(bytes.NewReader(resp.Body))
}
