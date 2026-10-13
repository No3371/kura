package allu

import "github.com/No3371/kura/internal/listing"

// Scheme is ALLU's condition grading. Labels and Japanese definitions are
// quoted from ALLU's 状態ランク一覧 (FAQ); ALLU grades its own stock. Codes are
// the rank letters the market payload carries (conditions). The FAQ's notes on
// what a rank covers (the item body, not its box, dust bag or guarantee card)
// belong to no single grade and are not part of any definition here.
var Scheme = listing.Scheme{
	Source: SourceName,
	Name:   "ALLU 状態ランク",
	Shop:   "ALLU",
	Graded: "店家鑑定",
	Ref:    "https://faq.allu-official.com/ja/articles/condition-rank-015f73d9",
	Grades: []listing.Grade{
		{Code: "N", Label: "N：未使用", Short: "N", Rank: listing.RankNew,
			DescJA: "プライスタグや保護シールなどが付属しており、未使用と判断できる大変状態の良い商品",
			DescZH: "附有價格吊牌或保護貼等、可判斷為未使用，狀態極佳的商品"},
		{Code: "S", Label: "S：新古品", Short: "S", Rank: listing.RankUnused,
			DescJA: "着用もしくは使用された形跡のない大変状態の良い商品",
			DescZH: "沒有穿戴或使用痕跡、狀態極佳的商品"},
		{Code: "SA", Label: "SA：非常にきれい", Short: "SA", Rank: listing.RankSA,
			DescJA: "数回の使用の形跡が見られるが、新品と遜色のない状態の良い商品",
			DescZH: "可見數次使用痕跡，但狀態良好、不遜於新品的商品"},
		{Code: "A", Label: "A：きれい", Short: "A", Rank: listing.RankA,
			DescJA: "少々の使用感が見られるが、状態の良い商品",
			DescZH: "可見少許使用感，但狀態良好的商品"},
		{Code: "AB", Label: "AB：ややきれい", Short: "AB", Rank: listing.RankAB,
			DescJA: "目立たない汚れや、小傷などが見られる商品",
			DescZH: "可見不明顯的髒污或細小刮痕等的商品"},
		{Code: "B", Label: "B：使用感あり", Short: "B", Rank: listing.RankB,
			DescJA: "使用による傷や汚れが見られる商品",
			DescZH: "可見使用造成的刮痕或髒污的商品"},
		{Code: "BC", Label: "BC：劣化あり", Short: "BC", Rank: listing.RankBC,
			DescJA: "傷や汚れにより、外観に影響が出ている商品",
			DescZH: "刮痕或髒污已影響外觀的商品"},
		{Code: "C", Label: "C：非常に劣化あり", Short: "C", Rank: listing.RankC,
			DescJA: "通常使用にも支障が出るレベルの商品",
			DescZH: "劣化程度已影響正常使用的商品"},
		// J stays in the ladder because it is the shop's scale, but ParseList
		// stores no J listing, so its Rank is only a placeholder bridge.
		// Rejected: storing J bridged to RankC — a 使用困難 bag priced against
		// C comparables reads as a below-market deal (a false 異常低價).
		// Rejected: a new Rank below C — it would change rankFactor, the
		// advisor's min_rank enum and every scheme's bridge for one grade.
		{Code: jRank, Label: "J：使用困難", Short: "J", Rank: listing.RankC,
			DescJA: "状態の劣化や損傷により使用が困難な商品",
			DescZH: "因狀態劣化或損傷而難以使用的商品"},
	},
}

func init() { listing.RegisterScheme(Scheme) }
