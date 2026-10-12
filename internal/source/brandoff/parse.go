// Package brandoff adapts BRAND OFF's online store (brandoff-store.com) to
// Kura's listing model: list-page and product-page parsers plus a sweeper that
// pages through one brand's catalogue.
package brandoff

import (
	"fmt"
	"io"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"github.com/No3371/kura/internal/listing"
)

const SourceName = "brandoff"

// bagCategoryPrefix: product links carry a category id; 100xxx are bags
// (100104 tote, 100105 shoulder, ...). The list endpoint ignores category
// filters when a brand filter is present, so we filter client-side.
const bagCategoryPrefix = "100"

// Card is one product tile on a list page.
type Card struct {
	listing.Listing
	CategoryID string
}

// IsBag reports whether the tile belongs to the bag category tree.
func (c Card) IsBag() bool { return strings.HasPrefix(c.CategoryID, bagCategoryPrefix) }

// ListPage is the parsed result of one ProductList.aspx page.
type ListPage struct {
	Total int // total hits for the query, as displayed by the site
	Cards []Card
}

// Badges share the numbered-icon scheme with grades (see condition.go).
var iconBadge = map[string]string{"1": "NEW", "2": "SALE", "8": "PRESALE"}

var (
	iconRe   = regexp.MustCompile(`ProductIcon(\d+)\.gif`)
	digitsRe = regexp.MustCompile(`[\d,]+`)
)

// ParseList parses a product list page. base resolves relative links.
func ParseList(r io.Reader, base *url.URL) (ListPage, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return ListPage{}, err
	}
	var page ListPage
	page.Total = firstInt(doc.Find(".productlist__itemcount").First().Text())

	doc.Find(".product__container .product__item").Each(func(_ int, s *goquery.Selection) {
		href, ok := s.Find(".product__item--image a").First().Attr("href")
		if !ok {
			return
		}
		u, err := base.Parse(href)
		if err != nil {
			return
		}
		pid := u.Query().Get("pid")
		if pid == "" {
			return
		}
		title := cleanText(s.Find(".product__item--name").Text())
		c := Card{CategoryID: u.Query().Get("cat")}
		c.Source = SourceName
		c.SourceID = pid
		c.URL = productURL(base, pid, c.CategoryID)
		c.Title = title
		if src, ok := s.Find("img#picture").Attr("src"); ok {
			if iu, err := base.Parse(src); err == nil {
				c.ImageURL = iu.String()
			}
		}
		s.Find("img").Each(func(_ int, img *goquery.Selection) {
			m := iconRe.FindStringSubmatch(img.AttrOr("src", ""))
			if m == nil {
				return
			}
			if code, ok := iconGrade[m[1]]; ok {
				g, _ := Scheme.Grade(code)
				c.Grade, c.Rank = g.Code, g.Rank
			} else if b, ok := iconBadge[m[1]]; ok {
				c.Badges = append(c.Badges, b)
			}
		})
		var prices []int64
		s.Find(".product__price--numeric").Each(func(_ int, p *goquery.Selection) {
			if v := firstInt(p.Text()); v > 0 {
				prices = append(prices, int64(v))
			}
		})
		c.PriceJPY, c.WasPriceJPY = splitPrices(prices)
		applyTitle(&c.Listing, title)
		page.Cards = append(page.Cards, c)
	})
	return page, nil
}

// productURL builds a canonical product URL so the same item always gets the
// same URL regardless of the query string it was discovered through.
func productURL(base *url.URL, pid, cat string) string {
	q := url.Values{"shop": {"0"}, "pid": {pid}}
	if cat != "" {
		q.Set("cat", cat)
	}
	u := *base
	u.Path = "/Form/Product/ProductDetail.aspx"
	u.RawQuery = q.Encode()
	return u.String()
}

// splitPrices treats the lowest shown price as the asking price and the highest
// (if different) as the struck-through pre-markdown price.
func splitPrices(ps []int64) (price, was int64) {
	for _, p := range ps {
		if price == 0 || p < price {
			price = p
		}
		if p > was {
			was = p
		}
	}
	if was == price {
		was = 0
	}
	return price, was
}

var genders = map[string]bool{"メンズ": true, "レディース": true, "ユニセックス": true}

// applyTitle splits a BRAND OFF title such as
//
//	ルイ・ヴィトン キーポル55  ボストンバッグ バッグ メンズ レディース M41424
//
// into brand / model name / sub-category / main category / gender / model code.
func applyTitle(l *listing.Listing, title string) {
	toks := strings.Fields(title)
	// Drop trailing condition notes like 【中古】.
	for len(toks) > 0 && strings.HasPrefix(toks[len(toks)-1], "【") {
		toks = toks[:len(toks)-1]
	}
	if len(toks) == 0 {
		return
	}
	start := 0
	if b, ok := listing.BrandByName(toks[0]); ok {
		l.Brand = b.Name
		start = 1
	}
	main := -1
	for i := start; i < len(toks); i++ {
		if toks[i] == "バッグ" {
			main = i
			break
		}
	}
	if main < 0 {
		// Not a bag title we understand; keep everything as the model name.
		l.ModelName = strings.Join(toks[start:], " ")
		l.Category = "bag"
		return
	}
	modelEnd := main
	l.Category = "bag"
	if main-1 >= start {
		if k := listing.CategoryFromJA(toks[main-1]); k != "" && k != "bag" {
			l.Category = k
			modelEnd = main - 1
		}
	}
	l.ModelName = strings.Join(toks[start:modelEnd], " ")
	var g, code []string
	for _, t := range toks[main+1:] {
		if genders[t] {
			g = append(g, t)
		} else {
			code = append(code, t)
		}
	}
	l.Gender = strings.Join(g, "/")
	l.ModelCode = strings.Join(code, " ")
}

// ParseDetail parses a product page's spec list (<dl><dt>商品品番</dt><dd>…</dd>…).
func ParseDetail(r io.Reader) (listing.Detail, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return listing.Detail{}, err
	}
	specs := map[string]string{}
	doc.Find("#tab1-content dl dt").Each(func(_ int, dt *goquery.Selection) {
		// <br> separates lines; keep them as " / " so sizes stay readable.
		var lines []string
		dt.NextFiltered("dd").Contents().Each(func(_ int, n *goquery.Selection) {
			if ln := cleanText(n.Text()); ln != "" {
				lines = append(lines, ln)
			}
		})
		specs[cleanText(dt.Text())] = strings.Join(lines, " / ")
	})
	if len(specs) == 0 {
		return listing.Detail{}, fmt.Errorf("brandoff: no spec list on product page")
	}
	acc := specs["付属品"]
	if i := strings.Index(acc, " / ※"); i >= 0 {
		acc = acc[:i] // drop the boilerplate disclaimer
	}
	return listing.Detail{
		ModelCode:   specs["商品品番"],
		Color:       specs["カラー"],
		Material:    specs["素材"],
		Size:        specs["サイズ"],
		Accessories: acc,
		Notes:       specs["備考"],
	}, nil
}

func cleanText(s string) string { return strings.Join(strings.Fields(s), " ") }

func firstInt(s string) int {
	m := digitsRe.FindString(s)
	n, _ := strconv.Atoi(strings.ReplaceAll(m, ",", ""))
	return n
}
