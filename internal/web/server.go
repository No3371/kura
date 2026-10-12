// Package web serves the JSON API, the advisor SSE stream and the embedded UI.
package web

import (
	"context"
	"embed"
	"encoding/json"
	"io/fs"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/No3371/kura/internal/advisor"
	"github.com/No3371/kura/internal/fx"
	"github.com/No3371/kura/internal/listing"
	"github.com/No3371/kura/internal/radar"
	"github.com/No3371/kura/internal/store"
)

//go:embed static
var static embed.FS

type Server struct {
	Store   *store.Store
	Radar   *radar.Engine
	FX      *fx.Converter
	Advisor advisor.Advisor
	Log     *slog.Logger
	// Crawl, if set, lets POST /api/crawl start a crawl in the background.
	Crawl    func(context.Context) error
	crawling atomic.Bool
	BaseCtx  context.Context
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	sub, _ := fs.Sub(static, "static")
	mux.Handle("GET /", http.FileServerFS(sub))

	mux.HandleFunc("GET /api/overview", s.api(s.overview))
	mux.HandleFunc("GET /api/listings", s.api(s.listings))
	mux.HandleFunc("GET /api/listings/{id}", s.api(s.listing))
	mux.HandleFunc("GET /api/drops", s.api(s.drops))
	mux.HandleFunc("GET /api/models", s.api(s.models))
	mux.HandleFunc("GET /api/models/{brand}/{key}", s.api(s.model))
	mux.HandleFunc("POST /api/crawl", s.api(s.crawl))
	mux.HandleFunc("POST /api/chat", s.chat)
	return mux
}

type apiFunc func(r *http.Request) (any, int, error)

// api refreshes the radar index (a cheap version check) and writes JSON.
func (s *Server) api(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := s.Radar.Refresh(r.Context()); err != nil {
			s.Log.Error("radar refresh", "err", err)
		}
		v, code, err := f(r)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if err != nil {
			if code == 0 {
				code = http.StatusInternalServerError
			}
			w.WriteHeader(code)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		if code != 0 {
			w.WriteHeader(code)
		}
		json.NewEncoder(w).Encode(v)
	}
}

type httpError struct{ msg string }

func (e httpError) Error() string { return e.msg }

func intParam(r *http.Request, k string) int64 {
	n, _ := strconv.ParseInt(r.URL.Query().Get(k), 10, 64)
	return n
}

func canonicalBrand(s string) string {
	if b, ok := listing.BrandByName(s); ok {
		return b.Name
	}
	return ""
}

func (s *Server) overview(r *http.Request) (any, int, error) {
	runs, err := s.Store.RecentRuns(r.Context(), 12)
	if err != nil {
		return nil, 0, err
	}
	var brands []map[string]string
	for _, b := range listing.Brands {
		brands = append(brands, map[string]string{"name": b.Name, "zh": b.ZH})
	}
	var cats []map[string]string
	for _, c := range listing.Categories {
		cats = append(cats, map[string]string{"key": c.Key, "zh": c.ZH})
	}
	return map[string]any{
		"overview": s.Radar.Overview(), "fx": s.FX.Current(), "advisor": s.Advisor.Mode(),
		"runs": runs, "crawling": s.crawling.Load(), "can_crawl": s.Crawl != nil,
		"brands": brands, "categories": cats,
	}, 0, nil
}

func (s *Server) listings(r *http.Request) (any, int, error) {
	q := r.URL.Query()
	f := radar.Filter{
		Brand: q.Get("brand"), Category: q.Get("category"), ModelCode: q.Get("model_code"),
		MinRank: q.Get("min_rank"), MinPriceTWD: intParam(r, "min_twd"), MaxPriceTWD: intParam(r, "max_twd"),
		OnlyDeals: q.Get("deals") == "1", IncludeGone: q.Get("gone") == "1",
		Sort: q.Get("sort"), Limit: int(intParam(r, "limit")),
	}
	if kw := strings.TrimSpace(q.Get("q")); kw != "" {
		f.Keywords = strings.Fields(kw)
	}
	items, total := s.Radar.Search(f)
	return map[string]any{"total": total, "items": items}, 0, nil
}

func (s *Server) listing(r *http.Request) (any, int, error) {
	it, stats, ok := s.Radar.Get(r.PathValue("id"))
	if !ok {
		return nil, http.StatusNotFound, httpError{"listing not found"}
	}
	return map[string]any{"item": it, "model": stats}, 0, nil
}

func (s *Server) drops(r *http.Request) (any, int, error) {
	return s.Radar.Drops(canonicalBrand(r.URL.Query().Get("brand")), int(intParam(r, "limit"))), 0, nil
}

func (s *Server) models(r *http.Request) (any, int, error) {
	limit := int(intParam(r, "limit"))
	if limit == 0 {
		limit = 40
	}
	return s.Radar.Models(canonicalBrand(r.URL.Query().Get("brand")), limit), 0, nil
}

func (s *Server) model(r *http.Request) (any, int, error) {
	stats, items, ok := s.Radar.Model(canonicalBrand(r.PathValue("brand")), r.PathValue("key"))
	if !ok {
		return nil, http.StatusNotFound, httpError{"model not found"}
	}
	return map[string]any{"model": stats, "items": items}, 0, nil
}

func (s *Server) crawl(r *http.Request) (any, int, error) {
	if s.Crawl == nil {
		return nil, http.StatusForbidden, httpError{"crawl trigger disabled"}
	}
	if !s.crawling.CompareAndSwap(false, true) {
		return nil, http.StatusConflict, httpError{"a crawl is already running"}
	}
	go func() {
		defer s.crawling.Store(false)
		if err := s.Crawl(s.BaseCtx); err != nil {
			s.Log.Error("crawl", "err", err)
		}
	}()
	return map[string]bool{"started": true}, http.StatusAccepted, nil
}

// RunCrawl runs a crawl unless one is in progress (used by the scheduler).
func (s *Server) RunCrawl(ctx context.Context) {
	if s.Crawl == nil || !s.crawling.CompareAndSwap(false, true) {
		return
	}
	defer s.crawling.Store(false)
	if err := s.Crawl(ctx); err != nil {
		s.Log.Error("crawl", "err", err)
	}
}

// chat streams advisor events as Server-Sent Events.
func (s *Server) chat(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Session string `json:"session"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10)).Decode(&req); err != nil ||
		strings.TrimSpace(req.Message) == "" || req.Session == "" {
		http.Error(w, `{"error":"session and message are required"}`, http.StatusBadRequest)
		return
	}
	if err := s.Radar.Refresh(r.Context()); err != nil {
		s.Log.Error("radar refresh", "err", err)
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Accel-Buffering", "no")

	emit := func(ev advisor.Event) {
		b, _ := json.Marshal(ev)
		w.Write([]byte("data: "))
		w.Write(b)
		w.Write([]byte("\n\n"))
		flusher.Flush()
	}
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Minute)
	defer cancel()
	start := time.Now()
	if err := s.Advisor.Chat(ctx, req.Session, req.Message, emit); err != nil {
		s.Log.Warn("advisor", "err", err)
		emit(advisor.Event{Type: "error", Text: err.Error()})
	}
	emit(advisor.Event{Type: "done"})
	s.Log.Info("chat", "mode", s.Advisor.Mode(), "took", time.Since(start).Round(time.Millisecond))
}
