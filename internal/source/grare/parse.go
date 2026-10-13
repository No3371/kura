// Package grare adapts GALLERY RARE's online store (g-rare.com) to Kura's
// listing model: list-page and product-page parsers plus a sweeper that reads
// the first page of one brand's bag category.
package grare

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"github.com/No3371/kura/internal/listing"
)

const SourceName = "grare"

// ListPage is the parsed result of one category list page.
type ListPage struct {
	Total int // total hits in the category, as displayed by the site
	Cards []listing.Listing
}

// Complete reports whether the page holds every listing of its category.
func (p ListPage) Complete() bool { return len(p.Cards) >= p.Total }

var (
	digitsRe  = regexp.MustCompile(`[\d,]+`)
	productID = regexp.MustCompile(`^[0-9A-Za-z]+$`)
	codeRe    = regexp.MustCompile(`^[0-9A-Z][0-9A-Z-]{4,}$`)
	upperRe   = regexp.MustCompile(`^[A-Z]{4,}$`)
)

// ParseList parses a category list page. base resolves relative links.
//
// The result count is read from inside the product list: category pages also
// carry a sub-category block with its own count (e.g. "4件中" for four
// sub-categories). A page without the product list's count, a count without a
// number, or a count below the cards on the page is an error, not a short
// page: a misread count makes the sweep look complete, which would mark every
// other listing of the brand gone.
func ParseList(r io.Reader, base *url.URL) (ListPage, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return ListPage{}, err
	}
	total := doc.Find(".fs-c-productList .fs-c-listControl__status__total").First()
	if total.Length() == 0 {
		return ListPage{}, errors.New("grare: no result count on list page")
	}
	if !digitsRe.MatchString(total.Text()) {
		return ListPage{}, fmt.Errorf("grare: result count %q has no number", cleanText(total.Text()))
	}
	page := ListPage{Total: firstInt(total.Text())}

	doc.Find("article.fs-c-productListItem").Each(func(_ int, s *goquery.Selection) {
		href, ok := s.Find(".fs-c-productName a").First().Attr("href")
		if !ok {
			return
		}
		u, err := base.Parse(href)
		if err != nil {
			return
		}
		u.RawQuery, u.Fragment = "", ""
		// Product ids are the last path segment: mostly the shop's numeric
		// 商品番号, sometimes a lower-case model code (".../m46203"). Category
		// slugs carry hyphens, so they never pass.
		id := path.Base(u.Path)
		if !productID.MatchString(id) {
			return
		}
		var l listing.Listing
		l.Source = SourceName
		l.SourceID = id
		l.URL = u.String()
		l.PriceJPY = int64(firstInt(s.Find(".fs-c-productPrice--selling .fs-c-price__value").First().Text()))
		if l.PriceJPY == 0 {
			return
		}
		title := cleanText(s.Find(".fs-c-productName__name").First().Text())
		l.Title = title
		if src, ok := s.Find("img.fs-c-productListItem__image__image").First().Attr("data-layzr"); ok {
			if iu, err := base.Parse(src); err == nil {
				l.ImageURL = iu.String()
			}
		}
		s.Find(".fs-c-productMark__mark.label_top .fs-c-productMark__label").EachWithBreak(func(_ int, m *goquery.Selection) bool {
			g, ok := Scheme.Grade(cleanText(m.Text()))
			if ok {
				l.Grade, l.Rank = g.Code, g.Rank
			}
			return !ok
		})
		s.Find(".fs-c-productMark__mark.label_bottom .fs-c-productMark__label").Each(func(_ int, m *goquery.Selection) {
			if cleanText(m.Text()) == "NEW" {
				l.Badges = append(l.Badges, "NEW")
			}
		})
		applyTitle(&l, title)
		page.Cards = append(page.Cards, l)
	})
	if page.Total < len(page.Cards) {
		return ListPage{}, fmt.Errorf("grare: result count %d below the %d cards on the page", page.Total, len(page.Cards))
	}
	return page, nil
}

var genders = map[string]bool{"メンズ": true, "レディース": true, "ユニセックス": true}

// applyTitle splits a GALLERY RARE title such as
//
//	ルイヴィトン ハンドバッグ モノグラム スピーディ・バンドリエール20 M46222 LOUIS VUITTON 2way
//
// into brand / model name / category / gender / model code. Titles run
// {brand} {category?} {model…} {colour/金具…} {code?} {BRAND EN} {extras}; the
// model name ends at the first token that starts the colour, code or English
// brand part. HERMES titles carry no code, so their colour words (before a
// "/") end the name. Misreads stay possible (a colour word without "/" joins
// the name); the model code, when present, is what groups listings.
func applyTitle(l *listing.Listing, title string) {
	toks := strings.Fields(title)
	if len(toks) == 0 {
		return
	}
	if b, ok := listing.BrandByName(toks[0]); ok {
		l.Brand = b.Name
		toks = toks[1:]
	}
	l.Category = "bag"
	for _, t := range toks {
		if k := listing.CategoryFromJA(t); k != "" && k != "bag" {
			l.Category = k
			break
		}
	}
	for _, t := range toks {
		if codeRe.MatchString(t) && strings.ContainsAny(t, "0123456789") && !brandModel(t) {
			l.ModelCode = t
			break
		}
	}
	var g, name []string
	stopped := false
	for _, t := range toks {
		if genders[t] {
			g = append(g, t)
			continue
		}
		if t == "セール品" {
			l.Badges = append(l.Badges, "SALE")
		}
		if stopped || listing.CategoryFromJA(t) != "" {
			continue
		}
		// Before any name token, a "/" token is the material, not a colour
		// (カーフレザー/モノグラム スピーディ…), and an English brand token is
		// the model's own name (シャネル ハンドバッグ CHANEL 22 ミニ…): ending
		// the name on either would leave it empty.
		_, isBrand := listing.BrandByName(t)
		lead := len(name) > 0
		if (lead && (isBrand || upperRe.MatchString(t) || strings.Contains(t, "/"))) ||
			(l.ModelCode != "" && t == l.ModelCode) || t == "セール品" {
			stopped = true
			continue
		}
		name = append(name, t)
	}
	l.Gender = strings.Join(g, "/")
	l.ModelName = strings.Join(name, " ")
}

// brandModel reports a code-shaped token that is a brand's model name, such
// as CHANEL22 or CHANEL25 (the CHANEL 22 / 25 bags), not a model code.
func brandModel(t string) bool {
	for _, b := range listing.Brands {
		if strings.HasPrefix(t, strings.ReplaceAll(b.Name, " ", "")) {
			return true
		}
	}
	return false
}

// ParseDetail parses a product page's spec table
// (<table class="detailSpec"><tr><th>型番</th><td>…</td></tr>…). sold reports
// the shop's out-of-stock notice (.fs-c-productNotice--outOfStock): the
// product template answers it by revealing 「表記価格は最終販売した時点での価格です。」
// and a 商品お探しサービス button, so a sold item's page stays up rather than
// answering 404. Rejected: treating only a 404 as gone, which reads such a
// page as a live listing. The template's own scripts name the class in their
// text; only an element carrying it counts.
func ParseDetail(r io.Reader) (d listing.Detail, sold bool, err error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return listing.Detail{}, false, err
	}
	if doc.Find(".fs-c-productNotice--outOfStock").Length() > 0 {
		return listing.Detail{}, true, nil
	}
	specs := map[string]string{}
	doc.Find("table.detailSpec tr").Each(func(_ int, tr *goquery.Selection) {
		th := tr.Find("th").First()
		if th.Length() == 0 {
			return
		}
		// <br> separates lines; keep them as " / " so sizes stay readable.
		var lines []string
		tr.Find("td").First().Contents().Each(func(_ int, n *goquery.Selection) {
			if ln := cleanText(n.Text()); ln != "" {
				lines = append(lines, ln)
			}
		})
		k := cleanText(th.Text())
		if _, dup := specs[k]; !dup {
			specs[k] = strings.Join(lines, " / ")
		}
	})
	if len(specs) == 0 {
		return listing.Detail{}, false, errors.New("grare: no spec table on product page")
	}
	return listing.Detail{
		ModelCode:   specs["型番"],
		Color:       specs["色"],
		Material:    specs["素材"],
		Size:        specs["サイズ"],
		Accessories: specs["付属品"],
		Notes:       specs["状態"],
	}, false, nil
}

// cleanText and firstInt repeat brandoff's unexported one-liners; a shared
// helper package for two one-liners was rejected as more surface than it saves.
func cleanText(s string) string { return strings.Join(strings.Fields(s), " ") }

func firstInt(s string) int {
	m := digitsRe.FindString(s)
	n, _ := strconv.Atoi(strings.ReplaceAll(m, ",", ""))
	return n
}
