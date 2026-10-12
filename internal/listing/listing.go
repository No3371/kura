// Package listing holds the source-independent shape of a resale listing and the
// normalisation tables (condition ranks, bag categories, brands) shared by every
// source adapter, the radar and the advisor.
package listing

import (
	"strings"
	"time"
)

// Rank is Kura's internal condition scale. Every source keeps its own grading
// system (see Scheme); each grade maps onto a Rank so listings graded by
// different sources can still be filtered and roughly compared.
type Rank string

const (
	RankNew    Rank = "N"  // 新品
	RankUnused Rank = "S"  // 未使用品
	RankSA     Rank = "SA" // 中古SA
	RankA      Rank = "A"
	RankAB     Rank = "AB"
	RankB      Rank = "B"
	RankBC     Rank = "BC"
	RankC      Rank = "C"
)

var rankOrder = []Rank{RankNew, RankUnused, RankSA, RankA, RankAB, RankB, RankBC, RankC}

// rankFactor is the expected price of a rank relative to an A-rank copy of the
// same model. It lets the radar compare listings across ranks when a model has
// too few same-rank comparables. The numbers are deliberately coarse; the radar
// reports lower confidence whenever it has to lean on them.
var rankFactor = map[Rank]float64{
	RankNew: 1.25, RankUnused: 1.15, RankSA: 1.07, RankA: 1.0,
	RankAB: 0.93, RankB: 0.85, RankBC: 0.78, RankC: 0.70,
}

// Score orders ranks from best (7) to worst (0); -1 for unknown.
func (r Rank) Score() int {
	for i, x := range rankOrder {
		if x == r {
			return len(rankOrder) - 1 - i
		}
	}
	return -1
}

// Factor returns the rank's price multiplier relative to A (1.0 if unknown).
func (r Rank) Factor() float64 {
	if f, ok := rankFactor[r]; ok {
		return f
	}
	return 1.0
}

// ParseRank accepts a normalised rank string ("A", "s", "SA") and reports
// whether it is known.
func ParseRank(s string) (Rank, bool) {
	r := Rank(strings.ToUpper(strings.TrimSpace(s)))
	return r, r.Score() >= 0
}

// Ranks returns every rank, best first.
func Ranks() []Rank { return append([]Rank(nil), rankOrder...) }

// Listing is one item for sale as observed on a source at crawl time.
type Listing struct {
	Source    string // e.g. "brandoff"
	SourceID  string // the source's own product id
	URL       string
	ImageURL  string
	Brand     string // canonical English brand, e.g. "LOUIS VUITTON"
	Title     string // original Japanese title
	ModelName string // e.g. "ネヴァーフルPM"
	ModelCode string // e.g. "N41359"; empty when the source gives none
	Category  string // canonical category key, see Categories
	Gender    string
	// Grade is the source's own condition grade code (see Scheme); Rank is
	// derived from it for internal comparisons only.
	Grade    string
	Rank     Rank
	PriceJPY int64
	// WasPriceJPY is the struck-through price when the source shows a markdown.
	WasPriceJPY int64
	Badges      []string // source badges such as "NEW", "SALE"
}

// ID is the globally unique key used in storage: "<source>:<source id>".
func (l Listing) ID() string { return l.Source + ":" + l.SourceID }

// GroupKey identifies "the same bag" for price comparison: the model code when
// present (most reliable), otherwise the normalised model name.
func (l Listing) GroupKey() string { return GroupKey(l.ModelCode, l.ModelName) }

// GroupKey is the free-function form used where only the columns are at hand.
func GroupKey(code, name string) string {
	if code = strings.ToUpper(strings.TrimSpace(code)); code != "" {
		return code
	}
	name = strings.Join(strings.Fields(name), "")
	if name == "" {
		return ""
	}
	return "name:" + name
}

// Detail is the extra information only available on a product page.
type Detail struct {
	ModelCode   string
	Color       string
	Material    string
	Size        string
	Accessories string
	Notes       string
	FetchedAt   time.Time
}
