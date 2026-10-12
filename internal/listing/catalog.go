package listing

import (
	"slices"
	"strings"
)

// Category is a canonical bag type with its labels.
type Category struct {
	Key string
	ZH  string
	JA  []string // source spellings that map to this category
}

// Categories is ordered so that more specific spellings are tried first.
var Categories = []Category{
	{Key: "shoulder", ZH: "肩背包", JA: []string{"ショルダーバッグ"}},
	{Key: "tote", ZH: "托特包", JA: []string{"トートバッグ"}},
	{Key: "handbag", ZH: "手提包", JA: []string{"ハンドバッグ"}},
	{Key: "2way", ZH: "兩用包", JA: []string{"2WAYバッグ", "2wayバッグ", "２WAYバッグ"}},
	{Key: "crossbody", ZH: "斜背包", JA: []string{"クロスボディバッグ"}},
	{Key: "backpack", ZH: "後背包", JA: []string{"リュックサック", "リュック", "バックパック"}},
	{Key: "boston", ZH: "波士頓包", JA: []string{"ボストンバッグ"}},
	{Key: "clutch", ZH: "手拿包", JA: []string{"クラッチバッグ", "セカンドバッグ"}},
	{Key: "belt", ZH: "腰包／胸包", JA: []string{"ウエストバッグ", "ボディバッグ", "ウエストポーチ"}},
	{Key: "business", ZH: "公事包", JA: []string{"ビジネスバッグ", "ブリーフケース"}},
	{Key: "travel", ZH: "旅行包", JA: []string{"キャリーバッグ", "トラベルバッグ", "スーツケース"}},
	{Key: "pouch", ZH: "收納包", JA: []string{"ポーチ"}},
	{Key: "vanity", ZH: "化妝包", JA: []string{"バニティバッグ"}},
	{Key: "basket", ZH: "籃子包", JA: []string{"かごバッグ", "カゴバッグ"}},
	{Key: "bag", ZH: "包款", JA: []string{"バッグ"}},
}

// CategoryFromJA maps a Japanese category token to a canonical key, or "".
func CategoryFromJA(token string) string {
	for _, c := range Categories {
		if slices.Contains(c.JA, token) {
			return c.Key
		}
	}
	return ""
}

// IsCategory reports whether key is a canonical category key.
func IsCategory(key string) bool {
	return slices.ContainsFunc(Categories, func(c Category) bool { return c.Key == key })
}

// CategoryZH returns the Chinese label for a key ("包款" if unknown).
func CategoryZH(key string) string {
	for _, c := range Categories {
		if c.Key == key {
			return c.ZH
		}
	}
	return "包款"
}

// Brand is a canonical brand with the spellings sources use for it.
type Brand struct {
	Name string   // canonical, upper-case English
	ZH   string   // Traditional Chinese name used in Taiwan
	JA   []string // katakana spellings seen in titles / search params
}

var Brands = []Brand{
	{Name: "HERMES", ZH: "愛馬仕", JA: []string{"エルメス"}},
	{Name: "CHANEL", ZH: "香奈兒", JA: []string{"シャネル"}},
	{Name: "LOUIS VUITTON", ZH: "路易威登", JA: []string{"ルイ・ヴィトン", "ルイヴィトン"}},
	{Name: "CELINE", ZH: "思琳", JA: []string{"セリーヌ"}},
	{Name: "LOEWE", ZH: "羅威", JA: []string{"ロエベ"}},
	{Name: "GOYARD", ZH: "戈雅", JA: []string{"ゴヤール"}},
	{Name: "DIOR", ZH: "迪奧", JA: []string{"ディオール", "クリスチャン・ディオール"}},
	{Name: "GUCCI", ZH: "古馳", JA: []string{"グッチ"}},
	{Name: "PRADA", ZH: "普拉達", JA: []string{"プラダ"}},
	{Name: "BOTTEGA VENETA", ZH: "寶緹嘉", JA: []string{"ボッテガ・ヴェネタ", "ボッテガヴェネタ"}},
	{Name: "FENDI", ZH: "芬迪", JA: []string{"フェンディ"}},
	{Name: "SAINT LAURENT", ZH: "聖羅蘭", JA: []string{"サンローラン", "イヴ・サンローラン"}},
	{Name: "MIU MIU", ZH: "繆繆", JA: []string{"ミュウミュウ"}},
}

// BrandByName finds a brand by canonical name, Chinese name or any katakana
// spelling, case-insensitively.
func BrandByName(s string) (Brand, bool) {
	s = strings.TrimSpace(s)
	for _, b := range Brands {
		if strings.EqualFold(b.Name, s) || b.ZH == s || slices.Contains(b.JA, s) {
			return b, true
		}
	}
	return Brand{}, false
}
