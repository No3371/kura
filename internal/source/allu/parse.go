// Package allu adapts ALLU's market (allu-official.com) to Kura's listing
// model: ALLU公式 stock only, read from the server-rendered page payload
// (Nuxt's __NUXT_DATA__) of list and product pages, plus a capped
// newest-first sweeper.
package allu

import (
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/No3371/kura/internal/listing"
)

const SourceName = "allu"

// official is the seller name ALLU's own stock carries (app_customer.name).
const official = "ALLU公式"

// jRank is ALLU's lowest grade (使用困難), kept in Scheme but never stored.
const jRank = "J"

// soldStatus is a list row's sale_status for a sold item. Evidence: the list
// fixture's one row with stock 0 has sale_status 2 (the other 29: stock 1,
// sale_status 1), and that item's product page (detail_sold.html, item
// 2117814) carries soldButton. Only stock 0 with this status is read as sold.
// Rejected: any status other than 1 as sold — one value seen, and a misread
// active item would be marked gone.
const soldStatus = 2

// apiItem is one row of a market list payload.
type apiItem struct {
	ID              int64  `json:"id"`
	Name            string `json:"name"`
	Model           string `json:"model"`
	ItemNumber      string `json:"item_number"`
	Color           string `json:"color"`
	DisplayPriceMin int64  `json:"display_price_min"`
	DisplayPriceMax int64  `json:"display_price_max"`
	Image           struct {
		URL string `json:"url"`
	} `json:"image"`
	AppCustomer struct {
		Name string `json:"name"`
	} `json:"app_customer"`
	IsAlluPartner  bool   `json:"is_allu_partner"`
	TopCategory    string `json:"item_category_top_name"`
	MiddleCategory string `json:"item_category_middle_name"`
	BrandName      string `json:"item_category_name"`
	Conditions     string `json:"conditions"`
	Stock          int    `json:"stock"`
	SaleStatus     int    `json:"sale_status"`
	IsNew          bool   `json:"is_new"`
	IsSale         bool   `json:"is_sale"`
}

type searchItems struct {
	TotalCount *int      `json:"total_count"`
	Page       int       `json:"page"`
	Limit      int       `json:"limit"`
	Items      []apiItem `json:"items"`
}

// ListPage is the parsed result of one market list page.
type ListPage struct {
	Total int // total_count: rows across all pages, sold rows included
	Page  int
	Limit int // rows per page
	Rows  int // rows on this page before filtering
	Cards []listing.Listing
	Sold  []string // source ids of ALLU公式 rows shown sold (soldStatus)
}

// Complete reports whether this page is the last one.
func (p ListPage) Complete() bool { return p.Page*p.Limit >= p.Total }

// ParseList parses a market list page. A payload without total_count, page or
// limit is an error, not a short page: a misread count that reads as complete
// would mark the brand's other listings gone after two sweeps.
func ParseList(r io.Reader) (ListPage, error) {
	root, err := payload(r)
	if err != nil {
		return ListPage{}, err
	}
	si, _, err := node[searchItems](root, "market:searchItems:")
	if err != nil {
		return ListPage{}, err
	}
	if si.TotalCount == nil || si.Limit <= 0 || si.Page <= 0 {
		return ListPage{}, errors.New("allu: list payload without total_count/page/limit")
	}
	p := ListPage{Total: *si.TotalCount, Page: si.Page, Limit: si.Limit, Rows: len(si.Items)}
	var sellerOK int
	for _, it := range si.Items {
		// Partner and user listings are not ALLU's own appraisal. The ?seller=1
		// query is not trusted alone. Partner-owned consigned items that ALLU
		// lists itself (利用規約 第四章, inspected only after purchase when not
		// yet in ALLU's hands) carry the ALLU公式 seller and no list-row flag, so
		// they pass this check; only the product payload marks them
		// (is_consignment, item_type_key), and ParseDetail does not read it.
		if it.AppCustomer.Name != official || it.IsAlluPartner {
			continue
		}
		sellerOK++
		// Sold rows stay in ALLU's list; storing them would show sold bags as
		// on sale. Their ids go to Sold instead, which marks a listing Kura
		// already holds gone.
		if it.Stock <= 0 || it.SaleStatus != 1 {
			if it.Stock <= 0 && it.SaleStatus == soldStatus {
				p.Sold = append(p.Sold, strconv.FormatInt(it.ID, 10))
			}
			continue
		}
		// The list URL already asks for the バッグ category (c-58); the row's
		// own top category is checked too because that request scope is a
		// query Kura cannot verify from the response, and a wallet or
		// accessory stored as a bag would join a bag's comparables. Rejected:
		// trusting the URL alone.
		if it.TopCategory != "バッグ" {
			continue
		}
		if it.Conditions == jRank {
			continue
		}
		g, ok := Scheme.Grade(it.Conditions)
		if !ok {
			continue
		}
		// A price range (variations) has no single price to compare.
		if it.DisplayPriceMin <= 0 || it.DisplayPriceMin != it.DisplayPriceMax {
			continue
		}
		title := cleanText(it.Name)
		l := listing.Listing{
			Source:    SourceName,
			SourceID:  strconv.FormatInt(it.ID, 10),
			URL:       itemURL(it.ID),
			ImageURL:  it.Image.URL,
			Title:     title,
			ModelName: cleanText(it.Model),
			ModelCode: modelCode(it.ItemNumber, title),
			Category:  category(it.MiddleCategory),
			Grade:     g.Code,
			Rank:      g.Rank,
			PriceJPY:  it.DisplayPriceMin,
		}
		if b, ok := listing.BrandByName(it.BrandName); ok {
			l.Brand = b.Name
		}
		if it.IsNew {
			l.Badges = append(l.Badges, "NEW")
		}
		if it.IsSale {
			l.Badges = append(l.Badges, "SALE")
		}
		p.Cards = append(p.Cards, l)
	}
	// Every row of a ?seller=1 page is ALLU公式. A page whose rows all fail the
	// seller check means ALLU changed a value Kura matches on (seller name,
	// partner flag); returning it as an empty result would leave the brand
	// stale without a signal, and a brand small enough to complete a sweep
	// would have every listing marked gone. Rejected: dropping the rows
	// silently, as for the other filters, whose all-dropped pages are normal
	// (a page of sold rows).
	if p.Rows > 0 && sellerOK == 0 {
		return ListPage{}, fmt.Errorf("allu: page %d: none of %d rows has seller %s (seller values changed?)", p.Page, p.Rows, official)
	}
	return p, nil
}

func itemURL(id int64) string {
	return baseURL + "/jp/ja/market/items/" + strconv.FormatInt(id, 10) + "/"
}

// category maps ALLU's middle category (ショルダーバッグ,
// ウエストポーチ・ボディバッグ, バックパック/リュック, …) to a canonical key:
// the first part with a specific match wins, else "bag".
func category(middle string) string {
	for _, part := range strings.FieldsFunc(middle, func(r rune) bool { return r == '/' || r == '・' }) {
		if k := listing.CategoryFromJA(strings.TrimSpace(part)); k != "" && k != "bag" {
			return k
		}
	}
	return "bag"
}

var codeRe = regexp.MustCompile(`^[0-9A-Z][0-9A-Z-]{4,}$`)

// modelCode prefers ALLU's 型番 field (item_number), else the first
// code-shaped title token with a digit. HERMES titles end with a stamp such
// as G:2026年, which the pattern rejects.
func modelCode(itemNumber, title string) string {
	if c := strings.ToUpper(cleanText(itemNumber)); c != "" {
		return c
	}
	for _, t := range strings.Fields(title) {
		if codeRe.MatchString(t) && strings.ContainsAny(t, "0123456789") {
			return t
		}
	}
	return ""
}

// apiAsset is the product payload's item.
type apiAsset struct {
	ItemNumber     string `json:"item_number"`
	Color          string `json:"color"`
	Material       string `json:"material"`
	Size           string `json:"size"`
	Accessories    string `json:"accessories"`
	ListingComment string `json:"listing_comment"`
	ButtonType     int    `json:"button_type"`
}

// soldButton is the product payload's button_type on a sold item. Evidence:
// two product pages of one list page (testdata), an in-stock row
// (stock 1, sale_status 1) with button_type 1 and a sold row (stock 0,
// sale_status 2) with button_type 6; the page renders its buy button from it
// client-side, so the HTML carries no sold text. Other values are read as
// not sold. Rejected: any value other than 1 as sold — values seen on one
// pair only, and a misread active item would be marked gone.
const soldButton = 6

// ParseDetail parses a product page's payload. sold reports button_type 6
// (soldButton); a sold item's page stays up (200) rather than answering 404.
// A page that answers 200 without the item node is an error, not gone: an
// error page is not evidence of a sale.
func ParseDetail(r io.Reader) (d listing.Detail, sold bool, err error) {
	root, err := payload(r)
	if err != nil {
		return listing.Detail{}, false, err
	}
	a, _, err := node[apiAsset](root, "market:asset:")
	if err != nil {
		return listing.Detail{}, false, err
	}
	if a.ButtonType == soldButton {
		return listing.Detail{}, true, nil
	}
	return listing.Detail{
		ModelCode:   modelCode(a.ItemNumber, ""),
		Color:       cleanText(a.Color),
		Material:    cleanText(a.Material),
		Size:        cleanText(a.Size),
		Accessories: cleanText(a.Accessories),
		Notes:       cleanText(a.ListingComment),
	}, false, nil
}

// cleanText repeats brandoff's and grare's unexported one-liner; a shared
// helper package for it was rejected as more surface than it saves.
func cleanText(s string) string { return strings.Join(strings.Fields(s), " ") }
