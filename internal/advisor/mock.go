package advisor

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/No3371/kura/internal/listing"
)

// Mock is a rule-based stand-in used when no API key is configured. It
// extracts brand / category / budget with simple patterns and runs the same
// search tool the Claude agent uses, so the UI and data path are identical.
type Mock struct{ Tools *Tools }

func (m *Mock) Mode() string { return "mock" }

var (
	// 8萬, 8万, 80000, 8w, 80k, NT$80,000
	budgetRe = regexp.MustCompile(`(?i)(\d+(?:\.\d+)?)\s*(萬|万|w|k|千)?`)
	lvRe     = regexp.MustCompile(`\blv\b`)
	zhCats   = map[string]string{
		"托特": "tote", "肩背": "shoulder", "後背": "backpack", "背包": "backpack", "手提": "handbag",
		"斜背": "crossbody", "兩用": "2way", "手拿": "clutch", "腰包": "belt", "胸包": "belt",
		"公事": "business", "旅行": "travel", "波士頓": "boston",
		"tote": "tote", "shoulder": "shoulder", "backpack": "backpack", "clutch": "clutch",
	}
	// Popular model names in Chinese/English -> katakana used in listing titles.
	zhModels = map[string]string{
		"neverfull": "ネヴァーフル", "speedy": "スピーディ", "alma": "アルマ", "birkin": "バーキン", "柏金": "バーキン",
		"kelly": "ケリー", "凱莉": "ケリー", "picotin": "ピコタン", "evelyne": "エブリン", "菱格": "マトラッセ",
		"matelasse": "マトラッセ", "luggage": "ラゲージ", "puzzle": "パズル", "onthego": "オンザゴー",
		"on the go": "オンザゴー", "pochette": "ポシェット", "boy": "ボーイ", "lady dior": "レディディオール",
	}
)

func parseBudgetTWD(msg string) int64 {
	if !strings.ContainsAny(msg, "0123456789") {
		return 0
	}
	var best int64
	for _, m := range budgetRe.FindAllStringSubmatch(msg, -1) {
		v, err := strconv.ParseFloat(m[1], 64)
		if err != nil {
			continue
		}
		switch strings.ToLower(m[2]) {
		case "萬", "万", "w":
			v *= 10000
		case "k", "千":
			v *= 1000
		}
		// Ignore small numbers that are sizes (25, 30, 35cm) rather than money.
		if v >= 3000 && int64(v) > best {
			best = int64(v)
		}
	}
	return best
}

func (m *Mock) Chat(_ context.Context, _ string, msg string, emit func(Event)) error {
	lower := strings.ToLower(msg)
	in := searchInput{Sort: "deal", Limit: 6}
	for _, b := range listing.Brands {
		names := append([]string{strings.ToLower(b.Name), b.ZH}, b.JA...)
		for _, n := range names {
			if strings.Contains(lower, strings.ToLower(n)) {
				in.Brand = b.Name
			}
		}
	}
	if lvRe.MatchString(lower) {
		in.Brand = "LOUIS VUITTON"
	}
	for k, v := range zhCats {
		if strings.Contains(lower, k) {
			in.Category = v
		}
	}
	for k, v := range zhModels {
		if strings.Contains(lower, k) {
			in.Keywords = append(in.Keywords, v)
		}
	}
	if strings.Contains(msg, "全新") || strings.Contains(msg, "未使用") {
		in.MinRank = "S"
	} else if strings.Contains(msg, "近新") || strings.Contains(msg, "狀況好") {
		in.MinRank = "A"
	}
	in.MaxPriceTWD = parseBudgetTWD(msg)

	raw := marshal(in)
	emit(Event{Type: "tool_call", Tool: "search_listings", Input: raw})
	res, err := m.Tools.Run("search_listings", raw)
	if err != nil {
		return err
	}
	emit(Event{Type: "tool_result", Tool: "search_listings", Summary: res.Summary})

	var b strings.Builder
	b.WriteString("（示範模式：未設定 ANTHROPIC_API_KEY，以下為規則式搜尋結果，未經 AI 分析。）\n\n")
	var cond []string
	if in.Brand != "" {
		cond = append(cond, in.Brand)
	}
	if in.Category != "" {
		cond = append(cond, listing.CategoryZH(in.Category))
	}
	if in.MaxPriceTWD > 0 {
		cond = append(cond, fmt.Sprintf("NT$%s 以內", comma(in.MaxPriceTWD)))
	}
	if len(cond) == 0 {
		cond = append(cond, "全部品牌")
	}
	if len(res.Listings) == 0 {
		fmt.Fprintf(&b, "依條件「%s」目前沒有找到在售商品，可以放寬預算或換個品牌試試。", strings.Join(cond, "、"))
	} else {
		fmt.Fprintf(&b, "依條件「%s」，依「低於行情幅度」排序的前 %d 件：\n\n", strings.Join(cond, "、"), len(res.Listings))
		for _, it := range res.Listings {
			fmt.Fprintf(&b, "- %s %s（%s：%s）NT$%s", it.ModelName, it.CategoryZH,
				it.Condition.Label, it.Condition.DescZH, comma(it.PriceTWD))
			if it.Fair != nil {
				fmt.Fprintf(&b, "，行情約 NT$%s", comma(it.Fair.FairTWD))
			}
			fmt.Fprintf(&b, " [[%s]]\n", it.ID)
		}
		r := m.Tools.FX.Current()
		fmt.Fprintf(&b, "\n匯率 1 JPY = %.4f TWD（%s），價格未含國際運費與進口稅費。", r.Rate, r.Provider)
	}
	emit(Event{Type: "text", Text: b.String()})
	if len(res.Listings) > 0 {
		emit(Event{Type: "listings", Listings: res.Listings})
	}
	return nil
}

func comma(n int64) string {
	s := strconv.FormatInt(n, 10)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	return s
}
