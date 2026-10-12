package advisor

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/No3371/kura/internal/fx"
	"github.com/No3371/kura/internal/listing"
	"github.com/No3371/kura/internal/radar"
)

// Tools executes advisor tool calls against the radar. It is shared by the
// Claude agent and the offline mock so both see exactly the same data.
type Tools struct {
	Radar *radar.Engine
	FX    *fx.Converter
}

// ToolResult is what a tool returns: JSON for the model, and listings for the UI.
type ToolResult struct {
	JSON     string
	Listings []radar.Item
	Summary  string // one line for the UI's activity log
}

type searchInput struct {
	Brand       string   `json:"brand"`
	Category    string   `json:"category"`
	Keywords    []string `json:"keywords"`
	ModelCode   string   `json:"model_code"`
	MinRank     string   `json:"min_rank"`
	MinPriceTWD int64    `json:"min_price_twd"`
	MaxPriceTWD int64    `json:"max_price_twd"`
	OnlyDeals   bool     `json:"only_deals"`
	Sort        string   `json:"sort"`
	Limit       int      `json:"limit"`
}

type modelInput struct {
	Brand     string `json:"brand"`
	ModelCode string `json:"model_code"`
	GroupKey  string `json:"group_key"`
}

type listingInput struct {
	ID string `json:"id"`
}

type overviewInput struct {
	Brand string `json:"brand"`
}

// compact is the per-listing shape sent to the model: enough to reason about,
// small enough that ten of them stay cheap.
type compact struct {
	ID         string    `json:"id"`
	Brand      string    `json:"brand"`
	Model      string    `json:"model"`
	ModelCode  string    `json:"model_code,omitempty"`
	Category   string    `json:"category"`
	Condition  condition `json:"condition"`
	PriceTWD   int64     `json:"price_twd"`
	PriceJPY   int64     `json:"price_jpy"`
	FairTWD    int64     `json:"fair_twd,omitempty"`
	DealPct    *float64  `json:"deal_pct,omitempty"`
	Confidence string    `json:"fair_confidence,omitempty"`
	Comps      int       `json:"comps,omitempty"`
	Dispersion *float64  `json:"comps_dispersion,omitempty"`
	Anomaly    bool      `json:"price_anomaly,omitempty"`
	Color      string    `json:"color,omitempty"`
	Material   string    `json:"material,omitempty"`
	Size       string    `json:"size,omitempty"`
	Accessory  string    `json:"accessories,omitempty"`
	DaysListed int       `json:"days_listed"`
	DroppedPct *float64  `json:"recent_price_drop_pct,omitempty"`
	Status     string    `json:"status,omitempty"`
}

// condition is the source's grade as the model sees it: the label the shop
// uses, the shop's own definition, and who graded it.
type condition struct {
	Grade      string `json:"grade"`
	Definition string `json:"definition_ja"`
	Meaning    string `json:"meaning_zh"`
	GradedBy   string `json:"graded_by"`
}

func toCondition(it radar.Item) condition {
	return condition{it.Condition.Label, it.Condition.DescJA, it.Condition.DescZH, it.ConditionScheme + "（" + it.ConditionBy + "）"}
}

func round2(f float64) *float64 { v := math.Round(f*100) / 100; return &v }

func toCompact(it radar.Item) compact {
	c := compact{ID: it.ID, Brand: it.Brand, Model: it.ModelName, ModelCode: it.ModelCode,
		Category: it.CategoryZH, Condition: toCondition(it), PriceTWD: it.PriceTWD, PriceJPY: it.PriceJPY,
		Color: it.Color, Material: it.Material, Size: it.Size, Accessory: it.Accessories, DaysListed: it.DaysListed}
	if it.Fair != nil {
		c.FairTWD, c.DealPct, c.Confidence, c.Comps = it.Fair.FairTWD, round2(it.Fair.DealPct), it.Fair.Confidence, it.Fair.Comps
		c.Dispersion = round2(it.Fair.Dispersion)
		c.Anomaly = it.Fair.Anomaly
	}
	if it.Drop != nil {
		c.DroppedPct = round2(it.Drop.Pct)
	}
	if it.Status != "active" {
		c.Status = "sold"
	}
	return c
}

func (t *Tools) fxNote() map[string]any {
	r := t.FX.Current()
	asOf := "n/a"
	if !r.AsOf.IsZero() {
		asOf = r.AsOf.Format(time.DateOnly)
	}
	return map[string]any{"jpy_to_twd": r.Rate, "as_of": asOf, "provider": r.Provider}
}

func marshal(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

// ErrInvalidInput marks tool input that failed validation; the agent returns it
// to the model as an error tool_result so it can correct itself.
type ErrInvalidInput struct{ msg string }

func (e ErrInvalidInput) Error() string { return e.msg }

func decode(raw string, v any) error {
	dec := json.NewDecoder(strings.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return ErrInvalidInput{"INVALID_JSON: " + err.Error()}
	}
	return nil
}

func checkBrand(b string) error {
	if b == "" {
		return nil
	}
	if _, ok := listing.BrandByName(b); !ok {
		var names []string
		for _, x := range listing.Brands {
			names = append(names, x.Name)
		}
		return ErrInvalidInput{fmt.Sprintf("unknown brand %q; known: %s", b, strings.Join(names, ", "))}
	}
	return nil
}

// Run dispatches a tool call by name.
func (t *Tools) Run(name, rawInput string) (ToolResult, error) {
	switch name {
	case "search_listings":
		var in searchInput
		if err := decode(rawInput, &in); err != nil {
			return ToolResult{}, err
		}
		return t.search(in)
	case "model_price_stats":
		var in modelInput
		if err := decode(rawInput, &in); err != nil {
			return ToolResult{}, err
		}
		return t.modelStats(in)
	case "get_listing":
		var in listingInput
		if err := decode(rawInput, &in); err != nil {
			return ToolResult{}, err
		}
		return t.getListing(in)
	case "market_overview":
		var in overviewInput
		if err := decode(rawInput, &in); err != nil {
			return ToolResult{}, err
		}
		return t.overview(in)
	}
	return ToolResult{}, ErrInvalidInput{"unknown tool " + name}
}

func (t *Tools) search(in searchInput) (ToolResult, error) {
	if err := checkBrand(in.Brand); err != nil {
		return ToolResult{}, err
	}
	if in.Category != "" && !listing.IsCategory(in.Category) {
		return ToolResult{}, ErrInvalidInput{"unknown category " + in.Category}
	}
	if in.MinRank != "" {
		if _, ok := listing.ParseRank(in.MinRank); !ok {
			return ToolResult{}, ErrInvalidInput{"min_rank must be one of N,S,SA,A,AB,B,BC,C"}
		}
	}
	if in.Limit <= 0 || in.Limit > 10 {
		in.Limit = 8
	}
	items, total := t.Radar.Search(radar.Filter{
		Brand: in.Brand, Category: in.Category, Keywords: in.Keywords, ModelCode: in.ModelCode,
		MinRank: in.MinRank, MinPriceTWD: in.MinPriceTWD, MaxPriceTWD: in.MaxPriceTWD,
		OnlyDeals: in.OnlyDeals, Sort: in.Sort, Limit: in.Limit,
	})
	out := make([]compact, len(items))
	for i, it := range items {
		out[i] = toCompact(it)
	}
	return ToolResult{
		JSON:     marshal(map[string]any{"total_matches": total, "returned": len(out), "fx": t.fxNote(), "listings": out}),
		Listings: items,
		Summary:  fmt.Sprintf("找到 %d 件符合條件的商品，回傳 %d 件", total, len(out)),
	}, nil
}

func (t *Tools) modelStats(in modelInput) (ToolResult, error) {
	b, ok := listing.BrandByName(in.Brand)
	if !ok {
		return ToolResult{}, checkBrand(in.Brand)
	}
	key := in.GroupKey
	if key == "" {
		key = listing.GroupKey(in.ModelCode, "")
	}
	if key == "" {
		return ToolResult{}, ErrInvalidInput{"model_code or group_key is required"}
	}
	stats, items, ok := t.Radar.Model(b.Name, key)
	if !ok {
		return ToolResult{JSON: marshal(map[string]any{"found": false,
			"hint": "no listings with that model code; try search_listings with keywords to find the right code"}),
			Summary: "查無此型號"}, nil
	}
	type gradeTWD struct {
		Source    string `json:"source"`
		Grade     string `json:"grade"`
		Meaning   string `json:"meaning_zh"`
		Count     int    `json:"active_count"`
		MedianTWD int64  `json:"median_twd"`
		MinTWD    int64  `json:"min_twd"`
		MaxTWD    int64  `json:"max_twd"`
	}
	var grades []gradeTWD
	for _, g := range stats.ByGrade {
		grades = append(grades, gradeTWD{g.Source, g.Grade.Label, g.Grade.DescZH, g.Count,
			t.FX.TWD(g.MedianJPY), t.FX.TWD(g.MinJPY), t.FX.TWD(g.MaxJPY)})
	}
	var cheapest []compact
	var cards []radar.Item
	for _, it := range items {
		if it.Status == "active" && len(cheapest) < 5 {
			cheapest = append(cheapest, toCompact(it))
			cards = append(cards, it)
		}
	}
	res := map[string]any{
		"brand": stats.Brand, "model": stats.ModelName, "model_code": stats.ModelCode, "group_key": stats.GroupKey,
		"active_listings": stats.Active, "sold_observed": stats.Sold,
		"sell_through": round2(stats.SellThrough), "median_days_to_sell": stats.MedianDaysToSell,
		"median_twd": t.FX.TWD(stats.MedianJPY), "by_condition_grade": grades, "cheapest_active": cheapest, "fx": t.fxNote(),
	}
	return ToolResult{JSON: marshal(res), Listings: cards,
		Summary: fmt.Sprintf("%s：在售 %d 件、已售出 %d 件", stats.ModelName, stats.Active, stats.Sold)}, nil
}

func (t *Tools) getListing(in listingInput) (ToolResult, error) {
	it, stats, ok := t.Radar.Get(in.ID)
	if !ok {
		return ToolResult{}, ErrInvalidInput{"no listing with id " + in.ID}
	}
	type pt struct {
		Date     string `json:"date"`
		PriceTWD int64  `json:"price_twd"`
	}
	var hist []pt
	for _, p := range it.History {
		hist = append(hist, pt{p.At.Format(time.DateOnly), t.FX.TWD(p.PriceJPY)})
	}
	res := map[string]any{"listing": toCompact(it), "title_ja": it.Title, "notes": it.Notes,
		"price_history": hist, "url": it.URL, "fx": t.fxNote()}
	if stats != nil {
		res["model_active_listings"] = stats.Active
		res["model_sold_observed"] = stats.Sold
	}
	return ToolResult{JSON: marshal(res), Listings: []radar.Item{it}, Summary: "讀取商品 " + it.ModelName}, nil
}

func (t *Tools) overview(in overviewInput) (ToolResult, error) {
	brand := ""
	if in.Brand != "" {
		b, ok := listing.BrandByName(in.Brand)
		if !ok {
			return ToolResult{}, checkBrand(in.Brand)
		}
		brand = b.Name
	}
	type model struct {
		Model     string `json:"model"`
		Brand     string `json:"brand"`
		GroupKey  string `json:"group_key"`
		Category  string `json:"category"`
		Active    int    `json:"active"`
		Sold      int    `json:"sold_observed"`
		MedianTWD int64  `json:"median_twd"`
		MinTWD    int64  `json:"min_twd"`
	}
	var models []model
	for _, m := range t.Radar.Models(brand, 15) {
		models = append(models, model{m.ModelName, m.Brand, m.GroupKey, listing.CategoryZH(m.Category),
			m.Active, m.Sold, t.FX.TWD(m.MedianJPY), t.FX.TWD(m.MinJPY)})
	}
	o := t.Radar.Overview()
	res := map[string]any{"active_listings": o.Active, "sold_observed": o.Sold, "by_brand": o.ByBrand,
		"tracking_since": o.Since.Format(time.DateOnly), "top_models": models, "fx": t.fxNote()}
	return ToolResult{JSON: marshal(res), Summary: fmt.Sprintf("市場概況：%d 件在售", o.Active)}, nil
}
