package grare

import (
	"net/url"
	"os"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/No3371/kura/internal/listing"
)

var testBase, _ = url.Parse("https://www.g-rare.com/c/fashion/louisvuitton/louisvuitton-bag")

func TestParseList(t *testing.T) {
	f, err := os.Open("testdata/list_lv.html")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	p, err := ParseList(f, testBase)
	if err != nil {
		t.Fatal(err)
	}
	// 3729 is the product list's count; the sub-category block above it
	// shows its own "4件中", which must not be read as the total.
	if len(p.Cards) != 40 || p.Total != 3729 || p.Complete() {
		t.Fatalf("got %d cards of %d (complete %v), want 40 of 3729, incomplete", len(p.Cards), p.Total, p.Complete())
	}
	var coded int
	for _, c := range p.Cards {
		if c.SourceID == "" || c.Title == "" || c.PriceJPY <= 0 || c.Brand != "LOUIS VUITTON" {
			t.Errorf("incomplete card: %+v", c)
		}
		if strings.Contains(c.URL, "?") || !strings.HasSuffix(c.URL, "/"+c.SourceID) {
			t.Errorf("card %s: URL %q", c.SourceID, c.URL)
		}
		if !strings.HasPrefix(c.ImageURL, "https://") || strings.Contains(c.ImageURL, "loading.svg") {
			t.Errorf("card %s: ImageURL %q", c.SourceID, c.ImageURL)
		}
		if g, ok := Scheme.Grade(c.Grade); !ok || g.Rank != c.Rank {
			t.Errorf("card %s: grade %q rank %q not consistent with the scheme", c.SourceID, c.Grade, c.Rank)
		}
		if c.ModelCode != "" {
			coded++
		}
	}
	if coded == 0 {
		t.Error("no card has a model code")
	}
	want := listing.Listing{
		Source: "grare", SourceID: "11545330",
		URL:       "https://www.g-rare.com/c/fashion/louisvuitton/louisvuitton-bag/monogram-bag/monogram-other-bag/11545330",
		ImageURL:  "https://galleryrare.itembox.cloud/product/598/000000559868/000000559868-01.jpg?size=m&w=NDAw&t=20261013123042",
		Brand:     "LOUIS VUITTON",
		Title:     "ルイヴィトン ボディバッグ モノグラム・エクリプス アヴェニュー・スリングバッグPM M47137 ブラック 黒",
		ModelName: "モノグラム・エクリプス アヴェニュー・スリングバッグPM", ModelCode: "M47137",
		Category: "belt", Grade: "未使用", Rank: listing.RankUnused, PriceJPY: 311700, Badges: []string{"NEW"},
	}
	if !reflect.DeepEqual(p.Cards[0], want) {
		t.Errorf("first card\n got %+v\nwant %+v", p.Cards[0], want)
	}
	// A product whose URL ends in a model code rather than a 商品番号.
	if !slices.ContainsFunc(p.Cards, func(c listing.Listing) bool { return c.SourceID == "m46203" }) {
		t.Error("card m46203 missing")
	}
}

// listPage builds a minimal list page: count is the product list's count
// text, followed by n priced cards.
func listPage(count string, n int) string {
	var b strings.Builder
	b.WriteString(`<html><body><div class="fs-c-productList">`)
	if count != "" {
		b.WriteString(`<span class="fs-c-listControl__status__total">` + count + `</span>`)
	}
	for i := range n {
		id := strconv.Itoa(11500000 + i)
		b.WriteString(`<article class="fs-c-productListItem"><span class="fs-c-productName"><a href="/c/fashion/x/` + id +
			`"><span class="fs-c-productName__name">ルイヴィトン ハンドバッグ スピーディ M41108</span></a></span>` +
			`<span class="fs-c-productPrice--selling"><span class="fs-c-price__value">100,000</span></span></article>`)
	}
	b.WriteString(`</div></body></html>`)
	return b.String()
}

func TestParseListCount(t *testing.T) {
	for _, c := range []struct {
		name, count string
		cards       int
		ok          bool
	}{
		{"whole category on the page", "2件中", 2, true},
		{"empty category", "0件中", 0, true},
		{"no count element", "", 1, false},
		{"count without a number", "件中", 1, false},
		{"count below the cards", "1件中", 2, false},
	} {
		p, err := ParseList(strings.NewReader(listPage(c.count, c.cards)), testBase)
		switch {
		case c.ok && err != nil:
			t.Errorf("%s: %v", c.name, err)
		case c.ok && (len(p.Cards) != c.cards || !p.Complete()):
			t.Errorf("%s: %d cards of %d (complete %v), want %d, complete", c.name, len(p.Cards), p.Total, p.Complete(), c.cards)
		case !c.ok && err == nil:
			t.Errorf("%s: parsed without error (%d cards of %d) — a sweep would read it as complete", c.name, len(p.Cards), p.Total)
		}
	}
}

func TestApplyTitle(t *testing.T) {
	cases := []struct {
		title                   string
		brand, model, code, cat string
		badges                  []string
	}{
		{"ルイヴィトン ハンドバッグ モノグラム スピーディ・バンドリエール20 M46222 LOUIS VUITTON 2way",
			"LOUIS VUITTON", "モノグラム スピーディ・バンドリエール20", "M46222", "handbag", nil},
		{"エルメス バーキン30 ブルーニュイ/シルバー金具 トゴ Y刻印 HERMES Birkin ハンドバッグ",
			"HERMES", "バーキン30", "", "handbag", nil},
		{"エルメス ハンドバッグ ケリー28 内縫い エトゥープ/シルバー金具 エバーカラー G刻印 HERMES Kelly 2wayバッグ",
			"HERMES", "ケリー28 内縫い", "", "handbag", nil},
		{"シャネル チェーンショルダーバッグ マトラッセ ココマーク バニティ パテント AP2873 CHANEL 黒 セール品",
			"CHANEL", "チェーンショルダーバッグ マトラッセ ココマーク バニティ パテント", "AP2873", "bag", []string{"SALE"}},
		{"セリーヌ ハンドバッグ ベルトバッグ マイクロ レザー 189153 CELINE 2wayショルダ",
			"CELINE", "ベルトバッグ マイクロ レザー", "189153", "handbag", nil},
		{"ルイヴィトン ハンドバッグ カーフレザー/モノグラム スピーディ P9 バンドリエール25 M11561 2way",
			"LOUIS VUITTON", "カーフレザー/モノグラム スピーディ P9 バンドリエール25", "M11561", "handbag", nil},
		{"シャネル ハンドバッグ CHANEL 22 ミニ ココマーク シャイニーカーフレザー AS3980 CHANEL 2wayショルダーバッグ",
			"CHANEL", "CHANEL 22 ミニ ココマーク シャイニーカーフレザー", "AS3980", "handbag", nil},
		{"シャネル ハンドバッグ CHANEL22 ミニ ココマーク シャイニーカーフレザー AS3980 CHANEL 2wayショルダー",
			"CHANEL", "CHANEL22 ミニ ココマーク シャイニーカーフレザー", "AS3980", "handbag", nil},
	}
	for _, c := range cases {
		var l listing.Listing
		applyTitle(&l, c.title)
		if l.Brand != c.brand || l.ModelName != c.model || l.ModelCode != c.code || l.Category != c.cat || !slices.Equal(l.Badges, c.badges) {
			t.Errorf("applyTitle(%q)\n got brand=%q model=%q code=%q cat=%q badges=%v", c.title, l.Brand, l.ModelName, l.ModelCode, l.Category, l.Badges)
		}
	}
	var l listing.Listing
	applyTitle(&l, "ルイヴィトン リュック ダミエ・グラフィット ジョッシュ NV N40365 LOUIS VUITTON ヴィトン バックパック メンズ ブラック 黒")
	if l.Gender != "メンズ" || l.Category != "backpack" {
		t.Errorf("gender %q category %q", l.Gender, l.Category)
	}
}

func TestParseDetail(t *testing.T) {
	f, err := os.Open("testdata/detail_bag.html")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	// The saved page is a live listing whose scripts name the out-of-stock
	// class in their text; that must not read as sold.
	d, sold, err := ParseDetail(f)
	if err != nil {
		t.Fatal(err)
	}
	if sold {
		t.Error("live product page parsed as sold")
	}
	want := listing.Detail{
		ModelCode:   "M81944",
		Color:       "ブルー",
		Material:    "モノグラム・トリヨン",
		Size:        "W29cm×H21.5cm×D1cm / B5サイズまで対応",
		Accessories: "保存袋 ・ハンドストラップ",
		Notes:       "表面/スレ、汚れ / 内側/押し跡 / ストラップ/汚れ / 金具/スレ / におい/素材特有のにおい",
	}
	if d != want {
		t.Errorf("detail\n got %+v\nwant %+v", d, want)
	}
	if _, _, err := ParseDetail(strings.NewReader("<html><body></body></html>")); err == nil {
		t.Error("a page without the spec table parsed without error")
	}
}

// No sold page was captured; this one is the saved live page with the
// out-of-stock notice its own scripts look for put before the cart button.
// The notice's text and position are synthetic, its class is the shop's.
func TestParseDetailSold(t *testing.T) {
	b, err := os.ReadFile("testdata/detail_bag.html")
	if err != nil {
		t.Fatal(err)
	}
	const cart = `<div class="fs-c-productActionButton`
	if !strings.Contains(string(b), cart) {
		t.Fatal("fixture has no cart button block")
	}
	notice := `<p class="fs-c-productNotice fs-c-productNotice--outOfStock">SOLD OUT</p>`
	for name, page := range map[string]string{
		"sold product page":     strings.Replace(string(b), cart, notice+cart, 1),
		"notice, no spec table": "<html><body>" + notice + "</body></html>",
	} {
		_, sold, err := ParseDetail(strings.NewReader(page))
		if err != nil || !sold {
			t.Errorf("%s: sold %v, err %v; want sold", name, sold, err)
		}
	}
}

func TestListPageComplete(t *testing.T) {
	three := make([]listing.Listing, 3)
	forty := make([]listing.Listing, 40)
	for _, c := range []struct {
		p    ListPage
		want bool
	}{
		{ListPage{Total: 3, Cards: three}, true},
		{ListPage{Total: 41, Cards: forty}, false},
		{ListPage{}, true},
	} {
		if got := c.p.Complete(); got != c.want {
			t.Errorf("ListPage{Total: %d, %d cards}.Complete() = %v, want %v", c.p.Total, len(c.p.Cards), got, c.want)
		}
	}
}

func TestTargets(t *testing.T) {
	got := (&Source{Brands: []string{"HERMES", "GOYARD", "LOUIS VUITTON"}}).Targets()
	if !slices.Equal(got, []string{"HERMES", "LOUIS VUITTON"}) {
		t.Errorf("Targets = %v, want [HERMES LOUIS VUITTON] (GOYARD has no bag category)", got)
	}
}
