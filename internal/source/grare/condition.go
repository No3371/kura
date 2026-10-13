package grare

import "github.com/No3371/kura/internal/listing"

// Scheme is GALLERY RARE's condition grading. Labels and Japanese definitions
// are quoted from the shop's 商品ランクガイド; the shop's buyers set the rank
// and its product managers check it before listing. Codes are the labels as
// list pages show them (未使用, not BRAND OFF's 未使用品).
var Scheme = listing.Scheme{
	Source: SourceName,
	Name:   "GALLERY RARE 商品ランク",
	Shop:   "GALLERY RARE",
	Graded: "店家鑑定",
	Ref:    "https://www.g-rare.com/f/condition",
	Grades: []listing.Grade{
		{Code: "新品", Label: "新品", Short: "新品", Rank: listing.RankNew,
			DescJA: "傷や使用感がなく、商品の製造より2年以内の商品。",
			DescZH: "無傷、無使用感，且製造日起 2 年內的商品。"},
		{Code: "未使用", Label: "未使用", Short: "未使用", Rank: listing.RankUnused,
			DescJA: "傷や使用感はほぼないが、下記のいずれか、または複数に該当する商品。①製造より2年以上経過しているか、もしくは製造年が不明。②保存や、展示の際軽微な状態変化が見られる。",
			DescZH: "幾乎無傷、無使用感，但符合下列一項以上的商品：①製造已逾 2 年或製造年份不明；②保存或展示時出現輕微狀態變化。"},
		// 中古S bridges to RankSA, not RankUnused: that rank stands for unused
		// stock, and 中古S is a used item by the shop's own definition. A new
		// Rank between them was rejected: it would change rankFactor, the
		// advisor's min_rank enum and every other scheme's bridge for one shop.
		// 中古S and 中古SA sharing RankSA is harmless; Rank is only the
		// cross-shop bridge, and the ladder keeps the shop's own order.
		{Code: "中古S", Label: "中古S", Short: "S", Rank: listing.RankSA,
			DescJA: "わずかしか使用の形跡が見られない中古品。",
			DescZH: "僅有極少使用痕跡的二手品。"},
		{Code: "中古SA", Label: "中古SA", Short: "SA", Rank: listing.RankSA,
			DescJA: "ジュエリー・時計など磨き上げを施したものできれいになった商品。",
			DescZH: "珠寶、手錶等經拋光處理而變得漂亮的商品。"},
		{Code: "中古A", Label: "中古A", Short: "A", Rank: listing.RankA,
			DescJA: "傷などの使用の形跡が見られるが、コンディションの良い商品。",
			DescZH: "可見刮痕等使用痕跡，但狀態良好的商品。"},
		{Code: "中古AB", Label: "中古AB", Short: "AB", Rank: listing.RankAB,
			DescJA: "日常的に使用していた形跡が見られ、目立つ傷等もあるがコンディションの良い商品。",
			DescZH: "可見日常使用痕跡，也有明顯刮痕等，但狀態良好的商品。"},
		{Code: "中古B", Label: "中古B", Short: "B", Rank: listing.RankB,
			DescJA: "日常的に使用していた形跡が見られ、目立つ傷等もある中古商品。",
			DescZH: "可見日常使用痕跡，也有明顯刮痕等的二手商品。"},
		{Code: "中古BC", Label: "中古BC", Short: "BC", Rank: listing.RankBC,
			DescJA: "日常的に使用していた形跡が見られ、目立つ傷が多くみられる中古商品。",
			DescZH: "可見日常使用痕跡，且有許多明顯刮痕的二手商品。"},
		{Code: "中古C", Label: "中古C", Short: "C", Rank: listing.RankC,
			DescJA: "非常に使用感や傷があり劣化の目立つ中古商品。遠くから見ても劣化のわかる商品。",
			DescZH: "使用感與刮痕非常明顯、劣化顯著的二手商品，遠看也看得出劣化。"},
	},
}

func init() { listing.RegisterScheme(Scheme) }
