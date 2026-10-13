package allu

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/No3371/kura/internal/crawl"
	"github.com/No3371/kura/internal/listing"
)

// baseURL is a variable, not a constant, only so tests can point the sweeper
// at an httptest server; nothing else changes it. Rejected: a per-Source base
// field, which every caller would have to leave empty to get the real site.
var baseURL = "https://allu-official.com"

// bagCategory is ALLU's top-level バッグ category code.
const bagCategory = "c-58"

// maxListPages caps list pages per brand per sweep. LOUIS VUITTON alone has
// about 49,000 ALLU公式 bag rows (over 1,600 pages) and HERMES about 22,000
// (over 700), sold rows included; ALLU's terms bar requests repeated
// more than necessary (利用規約 第18条 (12)). Newest first, ten pages keep
// new arrivals and recent price changes. Rejected: -max-pages alone — its
// default 0 means every page.
const maxListPages = 10

// brandCodes are ALLU's market brand codes for Kura's configured brands.
var brandCodes = map[string]string{
	"HERMES":        "b-4",
	"CHANEL":        "b-8",
	"LOUIS VUITTON": "b-21",
	"CELINE":        "b-81",
	"LOEWE":         "b-88",
	"GOYARD":        "b-422",
}

// Source sweeps ALLU公式 bags of each configured brand. Targets are canonical
// brand names (see listing.Brands).
type Source struct {
	Fetcher *crawl.Fetcher
	Brands  []string
}

var _ crawl.Source = (*Source)(nil)

func (s *Source) Name() string { return SourceName }

// Targets keeps the configured brands that have an ALLU brand code, in order.
func (s *Source) Targets() []string {
	var out []string
	for _, b := range s.Brands {
		if _, ok := brandCodes[b]; ok {
			out = append(out, b)
		}
	}
	return out
}

func listURL(code string, page int) string {
	return baseURL + "/jp/ja/market/" + code + "/" + bagCategory + "/?seller=1&page=" + strconv.Itoa(page)
}

// Sweep reads ALLU公式 bags of target newest first (ALLU's default order),
// filtered by ?seller=1 and again by ParseList, and stops after maxListPages
// pages (or maxPages, when lower). Consequences: a brand with more ALLU公式
// bags than the cap never completes a sweep, so sweeps never mark its items
// gone and none count as seen appearing; sold items surface through sold rows
// in the pages read (ListPage.Sold) and through Detail, which the crawler runs
// once per listing and then repeats for listings that left the pages read
// (crawl.Crawler.RecheckAfter); price changes are seen only while an item is
// within the first pages. Rejected: reading every page (see maxListPages) and
// ALLU's JSON API (Kura reads only the pages a browser is served).
func (s *Source) Sweep(ctx context.Context, target string, maxPages int, visit func(int, []listing.Listing, []string) error) (bool, error) {
	code, ok := brandCodes[target]
	if !ok {
		return false, fmt.Errorf("allu: no brand code for %q", target)
	}
	limit := maxListPages
	if maxPages > 0 && maxPages < limit {
		limit = maxPages
	}
	for page := 1; page <= limit; page++ {
		u := listURL(code, page)
		resp, err := s.Fetcher.Get(ctx, u)
		if err != nil {
			return false, err
		}
		if resp.Status != http.StatusOK {
			return false, fmt.Errorf("allu: list %s: status %d", u, resp.Status)
		}
		lp, err := ParseList(bytes.NewReader(resp.Body))
		if err != nil {
			return false, err
		}
		for i := range lp.Cards {
			if lp.Cards[i].Brand == "" {
				lp.Cards[i].Brand = target
			}
		}
		if err := visit(page, lp.Cards, lp.Sold); err != nil {
			return false, err
		}
		if lp.Complete() {
			return true, nil
		}
		// An empty page must not read as the end: the sweep would then
		// count as complete and mark unseen listings gone.
		if lp.Rows == 0 {
			return false, fmt.Errorf("allu: empty page %d before total %d", page, lp.Total)
		}
	}
	return false, nil
}

// Detail fetches a product page. A removed product answers 404; a sold one
// keeps its page and carries the sold button type (ParseDetail).
func (s *Source) Detail(ctx context.Context, u string) (listing.Detail, bool, error) {
	resp, err := s.Fetcher.Get(ctx, u)
	if err != nil {
		return listing.Detail{}, false, err
	}
	if resp.Status == http.StatusNotFound {
		return listing.Detail{}, true, nil
	}
	if resp.Status != http.StatusOK {
		return listing.Detail{}, false, fmt.Errorf("allu: product %s: status %d", u, resp.Status)
	}
	return ParseDetail(bytes.NewReader(resp.Body))
}
