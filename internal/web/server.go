// Package web serves the JSON API, the advisor SSE stream and the embedded UI.
package web

import (
	"context"
	"embed"
	"encoding/json"
	"io/fs"
	"log/slog"
	"mime"
	"net/http"
	"path"
	"slices"
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
	mux.Handle("GET /", spaFallback(sub))

	mux.HandleFunc("GET /api/overview", s.api(s.overview))
	mux.HandleFunc("GET /api/listings", s.api(s.listings))
	mux.HandleFunc("GET /api/listings/{id}", s.api(s.listing))
	mux.HandleFunc("GET /api/drops", s.api(s.drops))
	mux.HandleFunc("GET /api/models", s.api(s.models))
	mux.HandleFunc("GET /api/models/{brand}/{key}", s.api(s.model))
	mux.HandleFunc("GET /api/member", s.api(s.member))
	mux.HandleFunc("GET /api/wanted", s.api(s.wanted))
	mux.HandleFunc("POST /api/crawl", s.api(s.crawl))
	mux.HandleFunc("POST /api/chat", s.chat)
	return mux
}

func init() {
	// On Windows the mime package reads the registry, which can map .js to
	// text/plain; browsers refuse to run a module script served that way.
	mime.AddExtensionType(".js", "text/javascript; charset=utf-8")
	mime.AddExtensionType(".css", "text/css; charset=utf-8")
}

// spaFallback serves embedded files as they are and answers any other GET with
// index.html, so client routes (/browse, /listing/{id}) survive a reload or a
// pasted link. Missing assets stay 404, since index.html there would hand HTML
// to fetch(), <script> or <link>: /api/ paths, anything under an embedded
// directory (js/, dev/, legacy/), and names with an asset extension.
//
// "Has any extension" is not the asset test: model codes carry dots
// (191502BZJ.04LU), and a route keyed by one would render through in-app
// navigation yet 404 on reload. Directories are served only when they hold an
// index.html; any other directory would answer with a file listing.
func spaFallback(root fs.FS) http.Handler {
	files := http.FileServerFS(root)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if name == "" {
			name = "."
		}
		if fi, err := fs.Stat(root, name); err == nil {
			if fi.IsDir() && !exists(root, path.Join(name, "index.html")) {
				http.NotFound(w, r)
				return
			}
			files.ServeHTTP(w, r)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/api/") || underAssetDir(root, name) ||
			assetExt[strings.ToLower(path.Ext(name))] {
			http.NotFound(w, r)
			return
		}
		http.ServeFileFS(w, r, root, "index.html")
	})
}

var assetExt = map[string]bool{
	".js": true, ".mjs": true, ".css": true, ".map": true, ".json": true, ".html": true,
	".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".webp": true, ".svg": true, ".ico": true,
	".woff": true, ".woff2": true, ".ttf": true, ".txt": true,
}

func exists(root fs.FS, name string) bool {
	_, err := fs.Stat(root, name)
	return err == nil
}

// underAssetDir reports whether name sits below a top-level embedded directory.
func underAssetDir(root fs.FS, name string) bool {
	dir, _, nested := strings.Cut(name, "/")
	if !nested {
		return false
	}
	fi, err := fs.Stat(root, dir)
	return err == nil && fi.IsDir()
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
		// The radar's own thresholds, for the pages that explain them (行情怎麼算,
		// the detail page's 行情估價 prose): stated from here, never typed into copy.
		"method": map[string]any{
			"deal_min_pct": radar.DealMinPct, "anomaly_pct": radar.AnomalyPct,
			"min_comps": radar.MinComps, "min_sold_history_days": int(radar.MinSoldHistory / (24 * time.Hour)),
		},
		"schemes": listing.Schemes(), // each shop's grade ladder, verbatim
	}, 0, nil
}

func (s *Server) listings(r *http.Request) (any, int, error) {
	q := r.URL.Query()
	f := radar.Filter{
		Brand: q.Get("brand"), Category: q.Get("category"), ModelCode: q.Get("model_code"),
		MinRank: q.Get("min_rank"), MinPriceTWD: intParam(r, "min_twd"), MaxPriceTWD: intParam(r, "max_twd"),
		OnlyDeals: q.Get("deals") == "1", IncludeGone: q.Get("gone") == "1",
		Sort: q.Get("sort"), Offset: int(intParam(r, "offset")), Limit: int(intParam(r, "limit")),
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
	res := map[string]any{"item": it, "model": stats}
	if sc, ok := listing.SchemeFor(it.Source); ok {
		res["scheme"] = sc // the shop's full grade ladder for the 鑑定卡
	}
	return res, 0, nil
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
	// seekers is seeded demand (wanted.go), apart from the stats so the
	// advisor, which reads ModelStats, never takes it for data.
	res := map[string]any{"model": stats, "items": items, "seekers": s.modelSeekers(stats)}
	// One ladder per shop in the group: grades are only comparable within a shop.
	// Starts as [], not nil: a nil slice encodes as null, and distHTML's
	// `schemes = []` default replaces only undefined, so null would throw on
	// schemes.filter while rendering the model and listing pages.
	schemes := []listing.Scheme{}
	seen := map[string]bool{}
	for _, it := range items {
		if seen[it.Source] {
			continue
		}
		seen[it.Source] = true
		if sc, ok := listing.SchemeFor(it.Source); ok {
			schemes = append(schemes, sc)
		}
	}
	slices.SortFunc(schemes, func(a, b listing.Scheme) int { return strings.Compare(a.Source, b.Source) })
	res["schemes"] = schemes
	return res, 0, nil
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
