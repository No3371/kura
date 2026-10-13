package allu

import (
	"os"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/No3371/kura/internal/listing"
)

func open(t *testing.T, name string) *os.File {
	t.Helper()
	f, err := os.Open("testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { f.Close() })
	return f
}

var digits = regexp.MustCompile(`^[0-9]+$`)

func atoi(s string) int64 { n, _ := strconv.ParseInt(s, 10, 64); return n }

func TestParseList(t *testing.T) {
	p, err := ParseList(open(t, "list_lv.html"))
	if err != nil {
		t.Fatal(err)
	}
	if p.Total <= 30 || p.Page != 1 || p.Limit != 30 || p.Rows != 30 || p.Complete() {
		t.Fatalf("page = total %d page %d limit %d rows %d complete %v; want >30, 1, 30, 30, false",
			p.Total, p.Page, p.Limit, p.Rows, p.Complete())
	}

	// Count the dropped classes from the decoded fixture itself.
	root, err := payload(open(t, "list_lv.html"))
	if err != nil {
		t.Fatal(err)
	}
	si, _, err := node[searchItems](root, "market:searchItems:")
	if err != nil {
		t.Fatal(err)
	}
	var dropped int
	for _, it := range si.Items {
		_, graded := Scheme.Grade(it.Conditions)
		if it.AppCustomer.Name != official || it.IsAlluPartner || it.Stock <= 0 || it.SaleStatus != 1 ||
			it.TopCategory != "バッグ" || it.Conditions == jRank || !graded ||
			it.DisplayPriceMin <= 0 || it.DisplayPriceMin != it.DisplayPriceMax {
			dropped++
		}
	}
	if dropped == 0 {
		t.Error("fixture has no dropped row; the filter is untested on real data")
	}
	if len(p.Cards) != p.Rows-dropped {
		t.Errorf("cards = %d, want rows %d - dropped %d", len(p.Cards), p.Rows, dropped)
	}

	var coded int
	for _, c := range p.Cards {
		if !digits.MatchString(c.SourceID) || c.URL != itemURL(atoi(c.SourceID)) || strings.Contains(c.URL, "?") {
			t.Errorf("card %s: URL %q", c.SourceID, c.URL)
		}
		if c.Title == "" || c.PriceJPY <= 0 || c.Brand != "LOUIS VUITTON" || !strings.HasPrefix(c.ImageURL, "https://") {
			t.Errorf("incomplete card: %+v", c)
		}
		if g, ok := Scheme.Grade(c.Grade); !ok || g.Code == jRank || g.Rank != c.Rank {
			t.Errorf("card %s: grade %q rank %q", c.SourceID, c.Grade, c.Rank)
		}
		if !listing.IsCategory(c.Category) {
			t.Errorf("card %s: category %q", c.SourceID, c.Category)
		}
		if c.ModelCode != "" {
			coded++
		}
	}
	if coded == 0 {
		t.Error("no card has a model code")
	}

	want := listing.Listing{
		Source: "allu", SourceID: "2117835",
		URL:       "https://allu-official.com/jp/ja/market/items/2117835/",
		ImageURL:  "https://image.nanboya.com/ALLU/9469512/AP180835.JPG",
		Brand:     "LOUIS VUITTON",
		Title:     "ヴィトン ダミエ ヴァヴァンPM N40108 ブラウン RFID反応あり",
		ModelName: "ヴァヴァンPM", ModelCode: "N40108",
		Category: "shoulder", Grade: "AB", Rank: listing.RankAB, PriceJPY: 198000,
	}
	i := slices.IndexFunc(p.Cards, func(c listing.Listing) bool { return c.SourceID == want.SourceID })
	if i < 0 {
		t.Fatalf("card %s not parsed", want.SourceID)
	}
	if !reflect.DeepEqual(p.Cards[i], want) {
		t.Errorf("card =\n%+v\nwant\n%+v", p.Cards[i], want)
	}
	// 2117814 is sold (stock 0, sale_status 2) and must not be stored; it is
	// the item of detail_sold.html, whose product page carries soldButton.
	if slices.ContainsFunc(p.Cards, func(c listing.Listing) bool { return c.SourceID == "2117814" }) {
		t.Error("sold row 2117814 kept")
	}
	if !reflect.DeepEqual(p.Sold, []string{"2117814"}) {
		t.Errorf("sold = %v, want [2117814]", p.Sold)
	}
}

// devalue flattens v into a devalue array the way Nuxt serialises state:
// every container slot holds the index of its value.
func devalue(v any) []any {
	var arr []any
	var add func(v any) int
	add = func(v any) int {
		i := len(arr)
		arr = append(arr, nil)
		switch v := v.(type) {
		case map[string]any:
			m := map[string]any{}
			for k, x := range v {
				m[k] = add(x)
			}
			arr[i] = m
		case []any:
			s := make([]any, len(v))
			for k, x := range v {
				s[k] = add(x)
			}
			arr[i] = s
		default:
			arr[i] = v
		}
		return i
	}
	add(v)
	return arr
}

func row(id int, mut func(map[string]any)) map[string]any {
	r := map[string]any{
		"id": id, "name": "エルメス バーキン25", "model": "バーキン25", "item_number": nil, "color": "ブラック",
		"display_price_min": 2098800, "display_price_max": 2098800,
		"image":                     map[string]any{"url": "https://image.nanboya.com/x.JPG"},
		"app_customer":              map[string]any{"name": "ALLU公式"},
		"is_allu_partner":           false,
		"item_category_top_name":    "バッグ",
		"item_category_middle_name": "ハンドバッグ",
		"item_category_name":        "エルメス",
		"conditions":                "A", "stock": 1, "sale_status": 1, "is_new": true, "is_sale": false,
	}
	if mut != nil {
		mut(r)
	}
	return r
}

func listPayload(search map[string]any) string {
	return `<script type="application/json" id="__NUXT_DATA__">` + mustJSON(devalue(map[string]any{
		"data": map[string]any{"market:searchItems:/jp/ja/market/b-4/c-58/:1:anonymous?seller=1&page=1": search},
	})) + `</script>`
}

func TestParseListFilters(t *testing.T) {
	items := []any{
		row(1, nil),
		row(2, func(r map[string]any) { r["is_allu_partner"] = true }),
		row(3, func(r map[string]any) { r["app_customer"] = map[string]any{"name": "someuser"} }),
		row(4, func(r map[string]any) { r["stock"], r["sale_status"] = 0, 2 }),
		row(5, func(r map[string]any) { r["conditions"] = "J" }),
		row(6, func(r map[string]any) { r["display_price_max"] = 2200000 }),
		row(7, func(r map[string]any) { r["stock"], r["sale_status"] = 0, 3 }), // unseen status: not sold evidence
		row(8, func(r map[string]any) { r["is_allu_partner"], r["stock"], r["sale_status"] = true, 0, 2 }),
	}
	p, err := ParseList(strings.NewReader(listPayload(map[string]any{
		"total_count": 8, "page": 1, "limit": 30, "items": items,
	})))
	if err != nil {
		t.Fatal(err)
	}
	if p.Rows != 8 || len(p.Cards) != 1 || p.Cards[0].SourceID != "1" || !p.Complete() {
		t.Fatalf("rows %d cards %+v complete %v; want 8 rows, only id 1, complete", p.Rows, p.Cards, p.Complete())
	}
	if !reflect.DeepEqual(p.Sold, []string{"4"}) {
		t.Errorf("sold = %v, want [4] (ALLU公式, stock 0, sale_status 2 only)", p.Sold)
	}
	c := p.Cards[0]
	if c.Brand != "HERMES" || c.Category != "handbag" || c.ModelCode != "" || c.Grade != "A" ||
		!reflect.DeepEqual(c.Badges, []string{"NEW"}) {
		t.Errorf("card = %+v", c)
	}

	_, err = ParseList(strings.NewReader(listPayload(map[string]any{"page": 1, "limit": 30, "items": items})))
	if err == nil || !strings.Contains(err.Error(), "total_count") {
		t.Errorf("missing total_count: err = %v", err)
	}
}

// TestParseListSellerDrift: a page whose rows all fail the seller check is an
// error, while a page of rows dropped by the other filters is not.
func TestParseListSellerDrift(t *testing.T) {
	renamed := func(r map[string]any) { r["app_customer"] = map[string]any{"name": "ALLU OFFICIAL"} }
	partner := func(r map[string]any) { r["is_allu_partner"] = true }
	_, err := ParseList(strings.NewReader(listPayload(map[string]any{
		"total_count": 90, "page": 1, "limit": 30, "items": []any{row(1, renamed), row(2, partner)},
	})))
	if err == nil || !strings.Contains(err.Error(), "seller") {
		t.Errorf("all rows fail the seller check: err = %v, want a seller error", err)
	}

	sold := func(r map[string]any) { r["stock"], r["sale_status"] = 0, 2 }
	p, err := ParseList(strings.NewReader(listPayload(map[string]any{
		"total_count": 90, "page": 1, "limit": 30, "items": []any{row(1, sold), row(2, renamed)},
	})))
	if err != nil || p.Rows != 2 || len(p.Cards) != 0 {
		t.Errorf("sold ALLU公式 page: rows %d cards %d err %v; want 2 rows, no cards, no error", p.Rows, len(p.Cards), err)
	}

	p, err = ParseList(strings.NewReader(listPayload(map[string]any{
		"total_count": 0, "page": 1, "limit": 30, "items": []any{},
	})))
	if err != nil || p.Rows != 0 || !p.Complete() {
		t.Errorf("empty last page: %+v err %v; want no rows, complete, no error", p, err)
	}
}

func TestCategory(t *testing.T) {
	for in, want := range map[string]string{
		"ショルダーバッグ":        "shoulder",
		"ハンドバッグ":          "handbag",
		"トートバッグ":          "tote",
		"ウエストポーチ・ボディバッグ":  "belt",
		"バックパック/リュック":     "backpack",
		"リュックサック":         "backpack",
		"クラッチバッグ/セカンドバッグ": "clutch",
		"スーツケース/キャリーバッグ":  "travel",
		"ビジネスバッグ/ブリーフケース": "business",
		"その他": "bag",
		"":    "bag",
	} {
		if got := category(in); got != want {
			t.Errorf("category(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestModelCode(t *testing.T) {
	for _, c := range []struct{ num, title, want string }{
		{"m46222", "anything", "M46222"},
		{"", "ルイヴィトン ネヴァーフルMM モノグラム M40156", "M40156"},
		{"", "エルメス バーキン25 ルトゥルネ ヴォースイフト ブルージーン シルバー金具 G:2026年", ""},
	} {
		if got := modelCode(c.num, c.title); got != c.want {
			t.Errorf("modelCode(%q, %q) = %q, want %q", c.num, c.title, got, c.want)
		}
	}
}

func TestParseDetail(t *testing.T) {
	d, sold, err := ParseDetail(open(t, "detail_bag.html"))
	if err != nil {
		t.Fatal(err)
	}
	const img = "https://image.nanboya.com/ALLU/9480610/ZZ029445"
	images := []string{img + ".JPG"}
	for i := 1; i <= 15; i++ {
		images = append(images, img+"-"+strconv.Itoa(i)+".JPG")
	}
	// The fixture's material is null in ALLU's payload.
	want := listing.Detail{
		ModelCode:   "M15242",
		Color:       "イエロー",
		Size:        "サイズW : 33cm サイズH : 21.5cm サイズD : 18.5cm ショルダー高さ : 56.5cm 持ち手高さ : 11cm",
		Accessories: "カデナ,クロシェット,ショルダーストラップ,付属バッグ/ポーチ",
		Notes:       "少々の使用感が見られますが、状態の良いお品物です。内部に保管臭がございます。※中古品の為、記載の無いキズや汚れ等がある場合がございます。予めご了承下さい。",
		Images:      images,
	}
	if sold || !reflect.DeepEqual(d, want) {
		t.Errorf("detail = %+v sold %v\nwant %+v, not sold", d, sold, want)
	}
}

func TestParseDetailImageOrder(t *testing.T) {
	page := `<script type="application/json" id="__NUXT_DATA__">` + mustJSON(devalue(map[string]any{
		"data": map[string]any{"market:asset:1": map[string]any{
			"button_type": 1,
			"app_item_images": []any{
				map[string]any{"url": "https://image.nanboya.com/b.JPG", "order": 2},
				map[string]any{"url": " ", "order": 3},
				map[string]any{"url": "https://image.nanboya.com/a.JPG", "order": 1},
			},
		}},
	})) + `</script>`
	d, sold, err := ParseDetail(strings.NewReader(page))
	if err != nil || sold {
		t.Fatalf("sold %v err %v", sold, err)
	}
	if want := []string{"https://image.nanboya.com/a.JPG", "https://image.nanboya.com/b.JPG"}; !slices.Equal(d.Images, want) {
		t.Errorf("Images = %v, want %v", d.Images, want)
	}
}

func TestParseDetailSold(t *testing.T) {
	_, sold, err := ParseDetail(open(t, "detail_sold.html"))
	if err != nil || !sold {
		t.Fatalf("sold = %v, err = %v; want sold", sold, err)
	}
}

func TestParseDetailNoNode(t *testing.T) {
	_, sold, err := ParseDetail(strings.NewReader(listPayload(map[string]any{"total_count": 0})))
	if err == nil || sold {
		t.Fatalf("sold = %v, err = %v; want an error, not sold", sold, err)
	}
}

func TestListPageComplete(t *testing.T) {
	for _, c := range []struct {
		p    ListPage
		want bool
	}{
		{ListPage{Total: 60, Page: 2, Limit: 30}, true},
		{ListPage{Total: 61, Page: 2, Limit: 30}, false},
		{ListPage{Total: 0, Page: 1, Limit: 30}, true},
	} {
		if got := c.p.Complete(); got != c.want {
			t.Errorf("%+v.Complete() = %v, want %v", c.p, got, c.want)
		}
	}
}
