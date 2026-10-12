// Package radar turns stored listings into market signals: a fair-value
// estimate and deal score per listing, price drops, and per-model statistics
// (price by condition, sell-through, time to sell).
//
// Fair value is the median of comparable listings — the same brand and model
// (model code when known) — preferring the same condition grade *as the same
// source assigned it*. When a model has too few same-grade comparables, prices
// are bridged through Kura's internal rank scale (listing.Rank.Factor) and
// confidence drops.
package radar

import (
	"context"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/No3371/kura/internal/fx"
	"github.com/No3371/kura/internal/listing"
	"github.com/No3371/kura/internal/store"
)

const (
	minComps = 3
	// Gone listings older than this stop counting as comparables.
	compWindow = 180 * 24 * time.Hour
)

type Valuation struct {
	FairJPY    int64   `json:"fair_jpy"`
	FairTWD    int64   `json:"fair_twd"`
	DealPct    float64 `json:"deal_pct"` // >0: asking below fair value
	Method     string  `json:"method"`   // same-grade | rank-adjusted
	Comps      int     `json:"comps"`
	Confidence string  `json:"confidence"` // high | medium | low
	Dispersion float64 `json:"dispersion"` // IQR / median of the comparables
	Anomaly    bool    `json:"anomaly"`    // implausibly far below fair value
}

type Drop struct {
	FromJPY int64     `json:"from_jpy"`
	ToJPY   int64     `json:"to_jpy"`
	Pct     float64   `json:"pct"`
	At      time.Time `json:"at"`
}

// Item is the API/agent view of a listing.
type Item struct {
	ID         string `json:"id"`
	URL        string `json:"url"`
	ImageURL   string `json:"image_url"`
	Source     string `json:"source"`
	Brand      string `json:"brand"`
	BrandZH    string `json:"brand_zh"`
	Title      string `json:"title"`
	ModelName  string `json:"model_name"`
	ModelCode  string `json:"model_code,omitempty"`
	GroupKey   string `json:"group_key"`
	Category   string `json:"category"`
	CategoryZH string `json:"category_zh"`
	Rank       string `json:"rank"` // internal scale, for filtering only
	// Condition is the source's own grade with its definition text.
	Condition listing.Grade `json:"condition"`
	// ConditionScheme names the grading system and who applies it.
	ConditionScheme string             `json:"condition_scheme"`
	ConditionBy     string             `json:"condition_by"`
	PriceJPY        int64              `json:"price_jpy"`
	PriceTWD        int64              `json:"price_twd"`
	WasPriceJPY     int64              `json:"was_price_jpy,omitempty"`
	Status          string             `json:"status"`
	FirstSeen       time.Time          `json:"first_seen"`
	GoneAt          *time.Time         `json:"gone_at,omitempty"`
	DaysListed      int                `json:"days_listed"`
	Color           string             `json:"color,omitempty"`
	Material        string             `json:"material,omitempty"`
	Size            string             `json:"size,omitempty"`
	Accessories     string             `json:"accessories,omitempty"`
	Notes           string             `json:"notes,omitempty"`
	Fair            *Valuation         `json:"fair,omitempty"`
	Drop            *Drop              `json:"drop,omitempty"`
	History         []store.PricePoint `json:"history,omitempty"`
}

// GradeStat is price by one source grade.
type GradeStat struct {
	Source    string        `json:"source"`
	Grade     listing.Grade `json:"grade"`
	Count     int           `json:"count"`
	MedianJPY int64         `json:"median_jpy"`
	MinJPY    int64         `json:"min_jpy"`
	MaxJPY    int64         `json:"max_jpy"`
}

type ModelStats struct {
	Brand     string      `json:"brand"`
	GroupKey  string      `json:"group_key"`
	ModelName string      `json:"model_name"`
	ModelCode string      `json:"model_code,omitempty"`
	Category  string      `json:"category"`
	ImageURL  string      `json:"image_url"`
	Active    int         `json:"active"`
	Sold      int         `json:"sold"`
	MedianJPY int64       `json:"median_jpy"`
	MinJPY    int64       `json:"min_jpy"`
	ByGrade   []GradeStat `json:"by_grade"`
	// SellThrough is sold / (sold + active) over everything observed.
	SellThrough float64 `json:"sell_through"`
	// MedianDaysToSell only counts items Kura saw being listed (first seen
	// after the initial crawl), since older items' listing date is unknown.
	MedianDaysToSell *float64 `json:"median_days_to_sell,omitempty"`
}

type Engine struct {
	Store *store.Store
	FX    *fx.Converter
	Now   func() time.Time

	mu      sync.RWMutex
	version string
	items   []*Item
	byID    map[string]*Item
	groups  map[string][]*Item // brand|group_key -> items
	builtAt time.Time
	epoch   time.Time // earliest first_seen = when crawling began
}

func (e *Engine) now() time.Time {
	if e.Now != nil {
		return e.Now()
	}
	return time.Now()
}

// Refresh rebuilds the index if the store changed since the last build.
func (e *Engine) Refresh(ctx context.Context) error {
	v, err := e.Store.Version(ctx)
	if err != nil {
		return err
	}
	e.mu.RLock()
	same := v == e.version && e.items != nil
	e.mu.RUnlock()
	if same {
		return nil
	}
	rows, err := e.Store.AllRows(ctx)
	if err != nil {
		return err
	}
	e.build(rows, v)
	return nil
}

func groupID(brand, key string) string { return brand + "|" + key }

func (e *Engine) build(rows []store.Row, version string) {
	now := e.now()
	items := make([]*Item, 0, len(rows))
	byID := make(map[string]*Item, len(rows))
	groups := map[string][]*Item{}
	var epoch time.Time
	for i := range rows {
		r := &rows[i]
		it := e.toItem(r, now)
		items = append(items, it)
		byID[it.ID] = it
		if it.GroupKey != "" {
			g := groupID(it.Brand, it.GroupKey)
			groups[g] = append(groups[g], it)
		}
		if epoch.IsZero() || r.FirstSeen.Before(epoch) {
			epoch = r.FirstSeen
		}
	}
	for _, it := range items {
		if it.Status == "active" && it.GroupKey != "" {
			it.Fair = e.value(it, groups[groupID(it.Brand, it.GroupKey)], now)
		}
	}
	e.mu.Lock()
	e.items, e.byID, e.groups, e.version, e.builtAt, e.epoch = items, byID, groups, version, now, epoch
	e.mu.Unlock()
}

func (e *Engine) toItem(r *store.Row, now time.Time) *Item {
	b, _ := listing.BrandByName(r.Brand)
	scheme, _ := listing.SchemeFor(r.Source)
	it := &Item{
		ID: r.ID, URL: r.URL, ImageURL: r.ImageURL, Source: r.Source,
		Brand: r.Brand, BrandZH: b.ZH, Title: r.Title,
		ModelName: r.ModelName, ModelCode: r.ModelCode, GroupKey: listing.GroupKey(r.ModelCode, r.ModelName),
		Category: r.Category, CategoryZH: listing.CategoryZH(r.Category),
		Rank: string(r.Rank), Condition: listing.GradeFor(r.Source, r.Grade),
		ConditionScheme: scheme.Name, ConditionBy: scheme.Graded,
		PriceJPY: r.PriceJPY, PriceTWD: e.FX.TWD(r.PriceJPY), WasPriceJPY: r.WasPriceJPY,
		Status: r.Status, FirstSeen: r.FirstSeen,
		Color: r.Color, Material: r.Material, Size: r.Size, Accessories: r.Accessories, Notes: r.Notes,
		History: r.PriceHistory,
	}
	end := now
	if !r.GoneAt.IsZero() {
		g := r.GoneAt
		it.GoneAt = &g
		end = g
	}
	it.DaysListed = int(end.Sub(r.FirstSeen).Hours() / 24)
	if h := r.PriceHistory; len(h) >= 2 {
		prev, last := h[len(h)-2], h[len(h)-1]
		if last.PriceJPY < prev.PriceJPY {
			it.Drop = &Drop{FromJPY: prev.PriceJPY, ToJPY: last.PriceJPY, At: last.At,
				Pct: float64(prev.PriceJPY-last.PriceJPY) / float64(prev.PriceJPY)}
		}
	}
	return it
}

// value estimates fair price for it from its group.
func (e *Engine) value(it *Item, group []*Item, now time.Time) *Valuation {
	rank := listing.Rank(it.Rank)
	var same, adjusted []float64
	for _, c := range group {
		if c == it || (c.GoneAt != nil && now.Sub(*c.GoneAt) > compWindow) {
			continue
		}
		p := float64(c.PriceJPY)
		cr := listing.Rank(c.Rank)
		// "Same condition" means the same grader said the same thing: grades
		// from different sources are only comparable through the rank bridge.
		if it.Condition.Code != "" && c.Source == it.Source && c.Condition.Code == it.Condition.Code {
			same = append(same, p)
		}
		if cr.Score() >= 0 {
			adjusted = append(adjusted, p/cr.Factor())
		}
	}
	v := &Valuation{}
	var comps []float64
	level := 0 // 0 low, 1 medium, 2 high
	switch {
	case len(same) >= minComps:
		comps = same
		v.FairJPY, v.Method, v.Comps = int64(median(same)), "same-grade", len(same)
		level = 1
		if len(same) >= 6 {
			level = 2
		}
	case len(adjusted) >= minComps && rank.Score() >= 0:
		comps = adjusted
		v.FairJPY, v.Method, v.Comps = int64(median(adjusted)*rank.Factor()), "rank-adjusted", len(adjusted)
		if len(adjusted) >= 6 {
			level = 1
		}
	default:
		return nil
	}
	// A wide spread means the "same model" group mixes things the title can't
	// tell apart (leathers, sizes, limited editions), so its median says little.
	// A well-defined model at one rank typically has IQR/median of 0.1–0.3.
	v.Dispersion = dispersion(comps)
	switch {
	case v.Dispersion > 0.6:
		level = 0
	case v.Dispersion > 0.4:
		level--
	}
	// Without a model code the group is only "same name": a line like
	// マトラッセ or a Birkin size spans leathers and hardware that move the price
	// more than condition does. Show the estimate, but never vouch for it.
	if strings.HasPrefix(it.GroupKey, "name:") {
		level = 0
	}
	v.Confidence = []string{"low", "low", "medium", "high"}[max(level, -1)+1]
	v.FairTWD = e.FX.TWD(v.FairJPY)
	v.DealPct = float64(v.FairJPY-it.PriceJPY) / float64(v.FairJPY)
	// Half price against a solid comparable set is rarely a bargain; it's
	// usually damage or missing parts the title doesn't mention.
	v.Anomaly = v.DealPct > anomalyPct
	return v
}

const anomalyPct = 0.5

// dispersion is the interquartile range relative to the median.
func dispersion(xs []float64) float64 {
	s := slices.Clone(xs)
	slices.Sort(s)
	q := func(p float64) float64 { return s[int(p*float64(len(s)-1)+0.5)] }
	m := median(s)
	if m == 0 {
		return 0
	}
	return (q(0.75) - q(0.25)) / m
}

// confidenceWeight discounts deal scores the radar is unsure about when ranking.
var confidenceWeight = map[string]float64{"high": 1, "medium": 0.7, "low": 0.35}

func median(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	s := slices.Clone(xs)
	slices.Sort(s)
	m := len(s) / 2
	if len(s)%2 == 1 {
		return s[m]
	}
	return (s[m-1] + s[m]) / 2
}

// Filter selects listings. Zero values mean "any".
type Filter struct {
	Brand       string   `json:"brand,omitempty"`
	Category    string   `json:"category,omitempty"`
	Keywords    []string `json:"keywords,omitempty"` // any-match over title/model/code/color/material
	ModelCode   string   `json:"model_code,omitempty"`
	MinRank     string   `json:"min_rank,omitempty"`
	MinPriceTWD int64    `json:"min_price_twd,omitempty"`
	MaxPriceTWD int64    `json:"max_price_twd,omitempty"`
	OnlyDeals   bool     `json:"only_deals,omitempty"`
	IncludeGone bool     `json:"include_gone,omitempty"`
	Sort        string   `json:"sort,omitempty"` // deal | price_asc | price_desc | newest
	Limit       int      `json:"limit,omitempty"`
}

// Search returns copies of matching items (without history, to keep payloads small).
func (e *Engine) Search(f Filter) (items []Item, total int) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	minRank := -1
	if r, ok := listing.ParseRank(f.MinRank); ok {
		minRank = r.Score()
	}
	brand := ""
	if b, ok := listing.BrandByName(f.Brand); ok {
		brand = b.Name
	}
	var kws []string
	for _, k := range f.Keywords {
		if k = strings.ToLower(strings.TrimSpace(k)); k != "" {
			kws = append(kws, k)
		}
	}
	var out []*Item
	for _, it := range e.items {
		if !f.IncludeGone && it.Status != "active" {
			continue
		}
		if (f.Brand != "" && it.Brand != brand) ||
			(f.Category != "" && it.Category != f.Category) ||
			(f.ModelCode != "" && !strings.EqualFold(it.ModelCode, f.ModelCode)) ||
			(minRank >= 0 && listing.Rank(it.Rank).Score() < minRank) ||
			(f.MinPriceTWD > 0 && it.PriceTWD < f.MinPriceTWD) ||
			(f.MaxPriceTWD > 0 && it.PriceTWD > f.MaxPriceTWD) ||
			(f.OnlyDeals && (it.Fair == nil || it.Fair.DealPct < 0.05 || it.Fair.Confidence == "low" || it.Fair.Anomaly)) {
			continue
		}
		if len(kws) > 0 && !matchAny(it, kws) {
			continue
		}
		out = append(out, it)
	}
	sortItems(out, f.Sort)
	total = len(out)
	limit := f.Limit
	if limit <= 0 || limit > 100 {
		limit = 30
	}
	if len(out) > limit {
		out = out[:limit]
	}
	items = make([]Item, len(out))
	for i, it := range out {
		items[i] = *it
		items[i].History = nil
	}
	return items, total
}

func matchAny(it *Item, kws []string) bool {
	hay := strings.ToLower(strings.Join([]string{it.Title, it.ModelName, it.ModelCode, it.Color, it.Material}, " "))
	for _, k := range kws {
		if strings.Contains(hay, k) {
			return true
		}
	}
	return false
}

// dealScore ranks listings by how far below fair value they are, discounted by
// confidence so a reliable 15% beats a shaky 50%.
func dealScore(it *Item) float64 {
	if it.Fair == nil {
		return -1
	}
	if it.Fair.Anomaly {
		return 0 // rank with at-market listings, never above a plausible deal
	}
	return it.Fair.DealPct * confidenceWeight[it.Fair.Confidence]
}

func sortItems(xs []*Item, by string) {
	switch by {
	case "price_asc":
		sort.SliceStable(xs, func(i, j int) bool { return xs[i].PriceJPY < xs[j].PriceJPY })
	case "price_desc":
		sort.SliceStable(xs, func(i, j int) bool { return xs[i].PriceJPY > xs[j].PriceJPY })
	case "newest":
		sort.SliceStable(xs, func(i, j int) bool { return xs[i].FirstSeen.After(xs[j].FirstSeen) })
	default: // deal
		sort.SliceStable(xs, func(i, j int) bool {
			di, dj := dealScore(xs[i]), dealScore(xs[j])
			if di != dj {
				return di > dj
			}
			return xs[i].PriceJPY < xs[j].PriceJPY
		})
	}
}

// Get returns one item with full history and its group's stats.
func (e *Engine) Get(id string) (Item, *ModelStats, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	it, ok := e.byID[id]
	if !ok {
		return Item{}, nil, false
	}
	var ms *ModelStats
	if g := e.groups[groupID(it.Brand, it.GroupKey)]; it.GroupKey != "" {
		s := e.stats(g)
		ms = &s
	}
	return *it, ms, true
}

// Drops returns active listings whose latest observation was a price cut.
func (e *Engine) Drops(brand string, limit int) []Item {
	e.mu.RLock()
	defer e.mu.RUnlock()
	var out []*Item
	for _, it := range e.items {
		if it.Status == "active" && it.Drop != nil && (brand == "" || it.Brand == brand) {
			out = append(out, it)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Drop.At.After(out[j].Drop.At) })
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	res := make([]Item, len(out))
	for i, it := range out {
		res[i] = *it
	}
	return res
}

// Models returns per-model stats for a brand ("" = all), most-listed first.
func (e *Engine) Models(brand string, limit int) []ModelStats {
	e.mu.RLock()
	defer e.mu.RUnlock()
	var out []ModelStats
	for _, g := range e.groups {
		if brand != "" && g[0].Brand != brand {
			continue
		}
		out = append(out, e.stats(g))
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Active+out[i].Sold != out[j].Active+out[j].Sold {
			return out[i].Active+out[i].Sold > out[j].Active+out[j].Sold
		}
		return out[i].GroupKey < out[j].GroupKey
	})
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out
}

// Model returns stats and every item (active first) for one model.
func (e *Engine) Model(brand, key string) (ModelStats, []Item, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	g, ok := e.groups[groupID(brand, key)]
	if !ok {
		return ModelStats{}, nil, false
	}
	items := make([]Item, len(g))
	for i, it := range g {
		items[i] = *it
	}
	sort.SliceStable(items, func(i, j int) bool {
		if (items[i].Status == "active") != (items[j].Status == "active") {
			return items[i].Status == "active"
		}
		return items[i].PriceJPY < items[j].PriceJPY
	})
	return e.stats(g), items, true
}

func (e *Engine) stats(g []*Item) ModelStats {
	first := g[0]
	s := ModelStats{Brand: first.Brand, GroupKey: first.GroupKey, ModelName: first.ModelName,
		ModelCode: first.ModelCode, Category: first.Category, ImageURL: first.ImageURL}
	type gradeKey struct{ source, code string }
	byGrade := map[gradeKey][]float64{}
	grades := map[gradeKey]listing.Grade{}
	var all, days []float64
	for _, it := range g {
		if it.Status == "active" {
			s.Active++
			all = append(all, float64(it.PriceJPY))
			k := gradeKey{it.Source, it.Condition.Code}
			byGrade[k] = append(byGrade[k], float64(it.PriceJPY))
			grades[k] = it.Condition
			if s.ImageURL == "" {
				s.ImageURL = it.ImageURL
			}
			if len(it.ModelName) > len(s.ModelName) {
				s.ModelName = it.ModelName
			}
		} else {
			s.Sold++
			// Only items we watched appear can tell us how long selling took.
			if it.GoneAt != nil && it.FirstSeen.After(e.epoch.Add(24*time.Hour)) {
				days = append(days, it.GoneAt.Sub(it.FirstSeen).Hours()/24)
			}
		}
	}
	if len(all) > 0 {
		s.MedianJPY = int64(median(all))
		s.MinJPY = int64(slices.Min(all))
	}
	if n := s.Active + s.Sold; n > 0 {
		s.SellThrough = float64(s.Sold) / float64(n)
	}
	if len(days) > 0 {
		m := median(days)
		s.MedianDaysToSell = &m
	}
	for k, ps := range byGrade {
		s.ByGrade = append(s.ByGrade, GradeStat{Source: k.source, Grade: grades[k], Count: len(ps),
			MedianJPY: int64(median(ps)), MinJPY: int64(slices.Min(ps)), MaxJPY: int64(slices.Max(ps))})
	}
	// Best condition first, then by source for a stable order.
	sort.Slice(s.ByGrade, func(i, j int) bool {
		a, b := s.ByGrade[i], s.ByGrade[j]
		if sa, sb := a.Grade.Rank.Score(), b.Grade.Rank.Score(); sa != sb {
			return sa > sb
		}
		return a.Source+a.Grade.Code < b.Source+b.Grade.Code
	})
	return s
}

type Overview struct {
	Active  int            `json:"active"`
	Sold    int            `json:"sold"`
	Models  int            `json:"models"`
	Valued  int            `json:"valued"`
	Deals   int            `json:"deals"`
	Drops   int            `json:"drops"`
	ByBrand map[string]int `json:"by_brand"`
	Since   time.Time      `json:"since"`
	BuiltAt time.Time      `json:"built_at"`
}

func (e *Engine) Overview() Overview {
	e.mu.RLock()
	defer e.mu.RUnlock()
	o := Overview{ByBrand: map[string]int{}, Models: len(e.groups), Since: e.epoch, BuiltAt: e.builtAt}
	for _, it := range e.items {
		if it.Status != "active" {
			o.Sold++
			continue
		}
		o.Active++
		o.ByBrand[it.Brand]++
		if it.Fair != nil {
			o.Valued++
			if it.Fair.DealPct >= 0.1 && it.Fair.Confidence != "low" && !it.Fair.Anomaly {
				o.Deals++
			}
		}
		if it.Drop != nil {
			o.Drops++
		}
	}
	return o
}
