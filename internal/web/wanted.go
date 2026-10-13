package web

import (
	"fmt"
	"hash/fnv"
	"math"
	"net/http"
	"sort"
	"time"

	"github.com/No3371/kura/internal/radar"
)

// 大家都在找 ranks models by how many people looked for them this week. Kura
// has no users, so that count is seeded: each model's number is a function of
// its brand, group and the Taipei calendar week, scaled by how much of the
// model Kura has seen listed (in sale plus sold), so popular models usually
// lead and the shelf reshuffles each week instead of freezing. The same
// function answers the model page's 「這週 N 人在找」, so a model reads the same
// number on both. Every other figure on a tile is real: the model's median
// price, its listings in sale, and its days to a new owner once crawl history
// allows. Rejected: the canvas's five sample models and counts — fixed names
// could miss the live catalogue, and their 行情 would contradict the model
// pages they link to.

const wantedShelf = 5 // the canvas draws one row of five

var taipei = time.FixedZone("Asia/Taipei", 8*60*60)

// seekers is a model's seeded 「這週 N 人在找」 for the week holding now.
// Base grows with the square root of listings seen, so a model with forty
// listings outranks one with four without drowning it; the weekly factor
// spans 0.55–1.45, enough to reorder neighbours.
func seekers(brand, group string, seen int, now time.Time) int {
	y, w := now.In(taipei).ISOWeek()
	h := fnv.New64a()
	fmt.Fprintf(h, "%s|%s|%d-W%02d", brand, group, y, w)
	f := 0.55 + 0.9*float64(h.Sum64()%1000)/1000
	n := int(math.Round((6 + 4*math.Sqrt(float64(seen))) * f))
	return max(n, 3)
}

func (s *Server) now() time.Time {
	if s.Radar.Now != nil {
		return s.Radar.Now()
	}
	return time.Now()
}

func (s *Server) modelSeekers(m radar.ModelStats) int {
	return seekers(m.Brand, m.GroupKey, m.Active+m.Sold, s.now())
}

type wantedModel struct {
	Rank      int    `json:"rank"`
	Brand     string `json:"brand"`
	GroupKey  string `json:"group_key"`
	ModelName string `json:"model_name"`
	ModelCode string `json:"model_code,omitempty"`
	// ImageURL is a real photo of the model from one of its listings, or
	// empty when every listing still waits on the shop's photo.
	ImageURL  string `json:"image_url,omitempty"`
	Seekers   int    `json:"seekers"`
	MedianTWD int64  `json:"median_twd"`
	Active    int    `json:"active"`
	// MedianDaysToSell is withheld until SoldStatsReady, the gate every
	// sold-derived figure shares.
	MedianDaysToSell *float64 `json:"median_days_to_sell,omitempty"`
}

// wanted is the shelf: models with at least one listing in sale, most
// seekers first, so every tile links to something a visitor can buy.
func (s *Server) wanted(r *http.Request) (any, int, error) {
	ready := s.Radar.SoldStatsReady()
	type cand struct {
		m radar.ModelStats
		n int
	}
	var cs []cand
	for _, m := range s.Radar.Models("", 0) {
		if m.Active > 0 {
			cs = append(cs, cand{m, s.modelSeekers(m)})
		}
	}
	sort.Slice(cs, func(i, j int) bool {
		if cs[i].n != cs[j].n {
			return cs[i].n > cs[j].n
		}
		return cs[i].m.Brand+cs[i].m.GroupKey < cs[j].m.Brand+cs[j].m.GroupKey
	})
	out := []wantedModel{}
	for i, c := range cs {
		if i == wantedShelf {
			break
		}
		m := c.m
		w := wantedModel{Rank: i + 1, Brand: m.Brand, GroupKey: m.GroupKey, ModelName: m.ModelName,
			ModelCode: m.ModelCode, Seekers: c.n, MedianTWD: s.FX.TWD(m.MedianJPY), Active: m.Active}
		if ready {
			w.MedianDaysToSell = m.MedianDaysToSell
		}
		// As the model page: the stats' image can be the shop's placeholder.
		if _, items, ok := s.Radar.Model(m.Brand, m.GroupKey); ok {
			for _, it := range items {
				if !it.PhotoPending && it.ImageURL != "" {
					w.ImageURL = it.ImageURL
					break
				}
			}
		}
		out = append(out, w)
	}
	return out, 0, nil
}
