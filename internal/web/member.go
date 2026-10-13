package web

import (
	"net/http"
	"sort"
	"time"

	"github.com/No3371/kura/internal/radar"
)

// The demo member is the one signed-in visitor the design draws. Kura has no
// users, so what she tracks — and so which events reach her — is picked here
// from live listings, and every notification carries the same Item the
// listing page shows. 降價 and 已找到新主人 are a price cut or sale the crawl
// observed. 新上架 is the model's newest active listing in the radar's
// 最新上架 order: the top of the latest scan that found any of its listings.
// Rejected: canned sample notifications — a price cut or sale invented for a
// real, active listing would contradict that listing's own page.

// memberBrands are the brands the design's sample notifications track.
var memberBrands = []string{"LOUIS VUITTON", "HERMES", "CHANEL"}

const (
	memberModels   = 2 // the design's 追蹤清單（4）: two models and two listings
	memberListings = 2
)

type memberEvent struct {
	Type string     `json:"type"` // drop | new | gone
	Key  string     `json:"key"`  // the watchlist entry it is about
	At   time.Time  `json:"at"`
	Item radar.Item `json:"item"`
}

type memberView struct {
	Name      string        `json:"name"`
	Initial   string        `json:"initial"`
	Watchlist []string      `json:"watchlist"`
	Events    []memberEvent `json:"events"`
}

// modelKey is the model page's 追蹤 key (model.js trackKey); listing keys are
// listing ids, so one watchlist holds both kinds.
func modelKey(brand, group string) string { return "model:" + brand + "/" + group }

func (s *Server) member(r *http.Request) (any, int, error) {
	v := memberView{Name: "林小姐", Initial: "林", Watchlist: []string{}, Events: []memberEvent{}}

	// Models: per brand, the first of its most-listed models whose newest
	// listing is a deal, so the 新上架 notification has a below-market figure
	// to state, as the design's sample does.
	var models int
	for _, b := range memberBrands {
		if models == memberModels {
			break
		}
		for _, m := range s.Radar.Models(b, 10) {
			_, items, ok := s.Radar.Model(b, m.GroupKey)
			if !ok {
				continue
			}
			newest := s.Radar.NewestActive(items)
			if newest == nil || !radar.IsDeal(newest) {
				continue
			}
			key := modelKey(b, m.GroupKey)
			v.Watchlist = append(v.Watchlist, key)
			v.Events = append(v.Events, memberEvent{Type: "new", Key: key, At: newest.FirstSeen, Item: bare(*newest)})
			models++
			break
		}
	}

	// Listings: real price cuts and sales first, so 降價 and 已找到新主人
	// appear whenever the crawl has observed one; then each brand's top deal.
	var cands []string
	for _, it := range s.Radar.Drops("", memberListings) {
		cands = append(cands, it.ID)
	}
	for _, it := range s.Radar.Gone(memberListings) {
		cands = append(cands, it.ID)
	}
	for _, b := range memberBrands {
		top, _ := s.Radar.Search(radar.Filter{Brand: b, OnlyDeals: true, Limit: 1})
		for _, it := range top {
			cands = append(cands, it.ID)
		}
	}
	seen := map[string]bool{}
	var listings int
	for _, id := range cands {
		if listings == memberListings {
			break
		}
		if seen[id] {
			continue
		}
		seen[id] = true
		it, _, ok := s.Radar.Get(id)
		if !ok {
			continue
		}
		v.Watchlist = append(v.Watchlist, id)
		listings++
		switch {
		case it.Status == "active" && it.Drop != nil:
			v.Events = append(v.Events, memberEvent{Type: "drop", Key: id, At: it.Drop.At, Item: bare(it)})
		case it.Status != "active" && it.GoneAt != nil:
			v.Events = append(v.Events, memberEvent{Type: "gone", Key: id, At: *it.GoneAt, Item: bare(it)})
		}
	}

	sort.SliceStable(v.Events, func(i, j int) bool {
		if !v.Events[i].At.Equal(v.Events[j].At) {
			return v.Events[i].At.After(v.Events[j].At)
		}
		return v.Events[i].Item.ID < v.Events[j].Item.ID
	})
	return v, 0, nil
}

// bare drops the price history: the feed needs the listing's current figures,
// and its page fetches the full history itself.
func bare(it radar.Item) radar.Item {
	it.History = nil
	return it
}
