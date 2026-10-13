package web

import (
	"net/http"

	"github.com/No3371/kura/internal/radar"
)

// 每日一品 (Home's hero) is one real listing per Taipei calendar day: the
// first request of the day asks the radar for its candidate (best condition,
// then lowest price, radar.DailyCandidate) and the store keeps it for the rest
// of the day. A pick that sells meanwhile is still the day's pick and comes
// back with its non-active status, so the page marks it 已找到新主人 rather than
// swapping in another listing.

type dailyPick struct {
	Day  string      `json:"day"`
	Item *radar.Item `json:"item"` // null when nothing in sale qualifies
}

func (s *Server) daily(r *http.Request) (any, int, error) {
	ctx := r.Context()
	day := s.now().In(taipei).Format("2006-01-02")
	id, ok, err := s.Store.DailyPick(ctx, day)
	if err != nil {
		return nil, 0, err
	}
	if !ok {
		c, found := s.Radar.DailyCandidate()
		if !found {
			return dailyPick{Day: day}, 0, nil
		}
		if id, err = s.Store.SetDailyPick(ctx, day, c.ID, s.now()); err != nil {
			return nil, 0, err
		}
	}
	it, _, found := s.Radar.Get(id)
	if !found {
		return dailyPick{Day: day}, 0, nil
	}
	it.History = nil
	return dailyPick{Day: day, Item: &it}, 0, nil
}
