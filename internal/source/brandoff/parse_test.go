package brandoff

import (
	"net/url"
	"os"
	"testing"

	"github.com/No3371/kura/internal/listing"
)

var testBase, _ = url.Parse("https://www.brandoff-store.com/Form/Product/ProductList.aspx")

func parseFixture(t *testing.T, name string) ListPage {
	t.Helper()
	f, err := os.Open("testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	p, err := ParseList(f, testBase)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestParseList(t *testing.T) {
	p := parseFixture(t, "list_lv.html")
	if p.Total < 1000 {
		t.Errorf("Total = %d, want the site-wide hit count (thousands)", p.Total)
	}
	if len(p.Cards) != 90 {
		t.Fatalf("got %d cards, want 90", len(p.Cards))
	}
	var bags int
	for _, c := range p.Cards {
		if c.SourceID == "" || c.PriceJPY <= 0 || c.Title == "" || c.ImageURL == "" {
			t.Errorf("incomplete card: %+v", c)
		}
		if g, ok := Scheme.Grade(c.Grade); !ok || g.Rank != c.Rank {
			t.Errorf("card %s: grade %q rank %q not consistent with the scheme", c.SourceID, c.Grade, c.Rank)
		}
		if c.IsBag() {
			bags++
		}
	}
	if bags == 0 {
		t.Error("expected some bag cards in the LV fixture")
	}

	var nf *Card
	for i := range p.Cards {
		if p.Cards[i].SourceID == "2101220315016" {
			nf = &p.Cards[i]
		}
	}
	if nf == nil {
		t.Fatal("Neverfull PM (pid 2101220315016) not found in fixture")
	}
	want := listing.Listing{
		Source: "brandoff", SourceID: "2101220315016",
		Brand: "LOUIS VUITTON", ModelName: "ネヴァーフルPM", ModelCode: "N41359",
		Category: "tote", Gender: "レディース", PriceJPY: 180000,
	}
	got := nf.Listing
	if got.Brand != want.Brand || got.ModelName != want.ModelName || got.ModelCode != want.ModelCode ||
		got.Category != want.Category || got.Gender != want.Gender || got.PriceJPY != want.PriceJPY {
		t.Errorf("Neverfull parsed as %+v", got)
	}
	// Icon 6 is ランクB; icon 1 on the same tile is the NEW badge.
	if nf.Grade != "ランクB" || nf.Rank != listing.RankB || len(nf.Badges) != 1 || nf.Badges[0] != "NEW" {
		t.Errorf("Neverfull grade=%q rank=%q badges=%v", nf.Grade, nf.Rank, nf.Badges)
	}
	if nf.CategoryID != "100104" || !nf.IsBag() {
		t.Errorf("CategoryID = %q", nf.CategoryID)
	}
	if got.URL != "https://www.brandoff-store.com/Form/Product/ProductDetail.aspx?cat=100104&pid=2101220315016&shop=0" {
		t.Errorf("URL = %s", got.URL)
	}
}

func TestParseListPastEnd(t *testing.T) {
	if p := parseFixture(t, "list_empty.html"); len(p.Cards) != 0 {
		t.Errorf("got %d cards past the last page, want 0", len(p.Cards))
	}
}

func TestApplyTitle(t *testing.T) {
	cases := []struct {
		title                        string
		brand, model, code, cat, sex string
	}{
		{"ルイ・ヴィトン キーポル55  ボストンバッグ バッグ メンズ レディース M41424",
			"LOUIS VUITTON", "キーポル55", "M41424", "boston", "メンズ/レディース"},
		{"エルメス バーキン30 ハンドバッグ バッグ レディース 【中古】",
			"HERMES", "バーキン30", "", "handbag", "レディース"},
		{"シャネル マトラッセ チェーンショルダー バッグ レディース A01112",
			"CHANEL", "マトラッセ チェーンショルダー", "A01112", "bag", "レディース"},
		{"ルイ・ヴィトン ジッピー ウォレット  ラウンド長財布 財布 レディース M82501",
			"LOUIS VUITTON", "ジッピー ウォレット ラウンド長財布 財布 レディース M82501", "", "bag", ""},
	}
	for _, c := range cases {
		var l listing.Listing
		applyTitle(&l, c.title)
		if l.Brand != c.brand || l.ModelName != c.model || l.ModelCode != c.code || l.Category != c.cat || l.Gender != c.sex {
			t.Errorf("applyTitle(%q)\n got brand=%q model=%q code=%q cat=%q gender=%q", c.title, l.Brand, l.ModelName, l.ModelCode, l.Category, l.Gender)
		}
	}
}

func TestParseDetail(t *testing.T) {
	f, err := os.Open("testdata/detail_bag.html")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	d, err := ParseDetail(f)
	if err != nil {
		t.Fatal(err)
	}
	if d.ModelCode != "N41359" || d.Color != "エベヌ/ブラウン系" || d.Accessories != "なし" {
		t.Errorf("detail = %+v", d)
	}
	if d.Size != "W約28cm x H約22cm x D約12.5cm / ハンドルの長さ：約44cm" {
		t.Errorf("Size = %q", d.Size)
	}
	if d.Material == "" || d.Notes == "" {
		t.Errorf("missing material/notes: %+v", d)
	}
}
