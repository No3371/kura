package brandoff

import "github.com/No3371/kura/internal/listing"

// Scheme is BRAND OFF's condition grading. Labels and Japanese definitions are
// quoted from the store's FAQ ("商品は中古品ですか？新品ですか？"); the shop's
// appraisers assign them.
var Scheme = listing.Scheme{
	Source: SourceName,
	Name:   "BRAND OFF 商品ランク",
	Shop:   "BRAND OFF",
	Graded: "店家鑑定",
	Ref:    "https://www.brandoff-store.com/Page/faq.aspx",
	Grades: []listing.Grade{
		{Code: "新品", Label: "新品", Short: "新品", Rank: listing.RankNew,
			DescJA: "新品未使用品です",
			DescZH: "全新未使用品"},
		{Code: "未使用品", Label: "未使用品", Short: "未使用", Rank: listing.RankUnused,
			DescJA: "未使用品にキズなどが少なく全体的に綺麗な状態",
			DescZH: "未使用，幾乎沒有刮痕，整體狀態漂亮"},
		{Code: "ランクA", Label: "ランクA", Short: "A", Rank: listing.RankA,
			DescJA: "多少のキズ・汚れ・若干の使用感等あり",
			DescZH: "有少許刮痕、髒污與輕微使用感"},
		{Code: "ランクB", Label: "ランクB", Short: "B", Rank: listing.RankB,
			DescJA: "全体的にキズ・汚れ・使用感等あり",
			DescZH: "整體有刮痕、髒污與使用感"},
		{Code: "ランクC", Label: "ランクC", Short: "C", Rank: listing.RankC,
			DescJA: "かなりのキズ・汚れ・使用感等あり",
			DescZH: "有相當多的刮痕、髒污與使用感"},
	},
}

func init() { listing.RegisterScheme(Scheme) }

// The site shows the grade only as a numbered icon
// (/Contents/ProductImages/ProductIconN.gif, empty alt text).
var iconGrade = map[string]string{"3": "新品", "4": "未使用品", "5": "ランクA", "6": "ランクB", "7": "ランクC"}
