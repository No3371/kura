// Package web serves the JSON API, the advisor SSE stream and the embedded UI.
package web

import (
	"cmp"
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"io/fs"
	"log/slog"
	"mime"
	"net/http"
	"net/url"
	"path"
	"slices"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
	"unicode/utf8"

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
	// Crawl runs one crawl, for the scheduler and, unless NoCrawlAPI is set,
	// for POST /api/crawl to start in the background.
	Crawl      func(context.Context) error
	NoCrawlAPI bool
	crawling   atomic.Bool
	BaseCtx    context.Context
	// ChatLimit rate-limits POST /api/chat per client IP; nil = no limit.
	ChatLimit *ChatLimiter
	// TrustProxy keys ChatLimit on the last X-Forwarded-For hop, for a
	// deploy behind one reverse proxy (Railway); off, it keys on RemoteAddr.
	TrustProxy bool
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	sub, _ := fs.Sub(static, "static")
	mux.Handle("GET /", spaFallback(sub, s.pageMeta))

	mux.HandleFunc("GET /api/overview", s.api(s.overview))
	mux.HandleFunc("GET /api/listings", s.api(s.listings))
	mux.HandleFunc("GET /api/listings/{id}", s.api(s.listing))
	mux.HandleFunc("GET /api/drops", s.api(s.drops))
	mux.HandleFunc("GET /api/models", s.api(s.models))
	mux.HandleFunc("GET /api/models/{brand}/{key}", s.api(s.model))
	mux.HandleFunc("GET /api/member", s.api(s.member))
	mux.HandleFunc("GET /api/wanted", s.api(s.wanted))
	mux.HandleFunc("GET /api/daily", s.api(s.daily))
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
// meta gives a client route's own link-preview tags (see serveIndex).
func spaFallback(root fs.FS, meta func(*http.Request) *pageMeta) http.Handler {
	files := http.FileServerFS(root)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if name == "" {
			name = "."
		}
		if name == "." {
			serveIndex(w, r, root, nil)
			return
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
		serveIndex(w, r, root, meta(r))
	})
}

// serveIndex writes index.html with __ORIGIN__ and __URL__ filled in from the
// request, since og:image and og:url must be absolute and the public host is
// only known per deploy. Railway's proxy terminates TLS and reports the
// scheme in X-Forwarded-Proto. A non-nil m replaces the site-wide preview tags
// with one listing's or model's, as link crawlers never run the SPA.
func serveIndex(w http.ResponseWriter, r *http.Request, root fs.FS, m *pageMeta) {
	b, err := fs.ReadFile(root, "index.html")
	if err != nil {
		http.NotFound(w, r)
		return
	}
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	if p := r.Header.Get("X-Forwarded-Proto"); p == "https" || p == "http" {
		scheme = p
	}
	origin := scheme + "://" + r.Host
	page := strings.NewReplacer(
		"__ORIGIN__", html.EscapeString(origin),
		"__URL__", html.EscapeString(origin+r.URL.EscapedPath()),
	).Replace(string(b))
	if m != nil {
		page = m.apply(page)
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	io.WriteString(w, page)
}

// pageMeta is a shared /listing or /model link's own preview.
type pageMeta struct {
	docTitle string // <title>, as the SPA sets it once rendered
	title    string // og:title
	desc     string
	// image is the shop's photo URL, hotlinked as the UI's <img> tags do; ""
	// keeps the site card (og.png) for items whose photo is still pending.
	image, imageAlt string
}

// apply rewrites index.html's site-wide tags. A shop photo is a small square
// of unknown size, so og.png's 1200x630 dimensions go and the Twitter card
// drops to the thumbnail layout.
func (m *pageMeta) apply(page string) string {
	e := html.EscapeString
	page = strings.Replace(page, "<title>Kura 蔵</title>", "<title>"+e(m.docTitle)+"</title>", 1)
	page = setMeta(page, `property="og:title"`, e(m.title))
	page = setMeta(page, `property="og:description"`, e(m.desc))
	if m.image != "" {
		page = setMeta(page, `property="og:image"`, e(m.image))
		page = setMeta(page, `property="og:image:alt"`, e(m.imageAlt))
		page = dropMeta(page, `property="og:image:width"`)
		page = dropMeta(page, `property="og:image:height"`)
		page = setMeta(page, `name="twitter:card"`, "summary")
	}
	return page
}

// metaTag finds the <meta> tag carrying attr, returning the bounds of the tag
// and of its content value.
func metaTag(page, attr string) (start, vStart, vEnd, end int, ok bool) {
	i := strings.Index(page, "<meta "+attr+` content="`)
	if i < 0 {
		return 0, 0, 0, 0, false
	}
	vStart = i + len("<meta "+attr+` content="`)
	n := strings.IndexByte(page[vStart:], '"')
	if n < 0 {
		return 0, 0, 0, 0, false
	}
	vEnd = vStart + n
	n = strings.IndexByte(page[vEnd:], '>')
	if n < 0 {
		return 0, 0, 0, 0, false
	}
	return i, vStart, vEnd, vEnd + n + 1, true
}

func setMeta(page, attr, value string) string {
	if _, vs, ve, _, ok := metaTag(page, attr); ok {
		return page[:vs] + value + page[ve:]
	}
	return page
}

func dropMeta(page, attr string) string {
	if s, _, _, e, ok := metaTag(page, attr); ok {
		return page[:s] + strings.TrimPrefix(page[e:], "\n")
	}
	return page
}

// pageMeta looks up the listing or model a client route names, with the
// SPA's own patterns (/listing/{id}, /model/{brand}/{key}, each segment
// URI-encoded). nil, for the site-wide tags, when the route is another one,
// the item is unknown, or the server runs without data (static-route tests).
func (s *Server) pageMeta(r *http.Request) *pageMeta {
	if s.Radar == nil || s.Radar.Store == nil {
		return nil
	}
	seg := strings.Split(strings.TrimPrefix(r.URL.EscapedPath(), "/"), "/")
	if !(len(seg) == 2 && seg[0] == "listing") && !(len(seg) == 3 && seg[0] == "model") {
		return nil
	}
	for i, p := range seg {
		v, err := url.PathUnescape(p)
		if err != nil || v == "" {
			return nil
		}
		seg[i] = v
	}
	if err := s.Radar.Refresh(r.Context()); err != nil {
		s.Log.Error("radar refresh", "err", err)
	}
	if seg[0] == "listing" {
		return s.listingMeta(seg[1])
	}
	return s.modelMeta(seg[1], seg[2])
}

func (s *Server) listingMeta(id string) *pageMeta {
	it, _, ok := s.Radar.Get(id)
	if !ok {
		return nil
	}
	name := cmp.Or(it.ModelName, it.Title)
	shop := cmp.Or(it.SourceName, it.Source)
	var desc []string
	if it.Status == "active" {
		desc = append(desc, twd(it.PriceTWD))
	} else {
		desc = append(desc, "已找到新主人", "最後售價 "+twd(it.PriceTWD))
	}
	if it.Condition.Label != "" {
		desc = append(desc, shop+" "+it.Condition.Label)
	} else {
		desc = append(desc, shop)
	}
	m := &pageMeta{
		docTitle: name + " · " + it.Brand + " · Kura 蔵",
		title:    it.Brand + " " + name,
		desc:     strings.Join(desc, " · "),
		imageAlt: name + " 商品照片",
	}
	// The detail page's first photo: the stored gallery's, else the list photo.
	if len(it.Images) > 0 {
		m.image = it.Images[0]
	} else if !it.PhotoPending {
		m.image = it.ImageURL
	}
	return m
}

func (s *Server) modelMeta(brand, key string) *pageMeta {
	st, items, ok := s.Radar.Model(canonicalBrand(brand), key)
	if !ok {
		return nil
	}
	name := cmp.Or(st.ModelName, st.ModelCode, key)
	m := &pageMeta{
		docTitle: name + " 型號 · " + st.Brand + " · Kura 蔵",
		title:    st.Brand + " " + name,
		imageAlt: name + " 型號照片",
	}
	// items lists active ones first, cheapest first; the photo is the first
	// real one, as on the model page (the stats' may be a placeholder).
	if len(items) > 0 && items[0].Status == "active" {
		m.desc = fmt.Sprintf("在售 %d 件，%s 起", st.Active, twd(items[0].PriceTWD))
	} else {
		m.desc = fmt.Sprintf("目前沒有在售 · 已售出 %d 件", st.Sold)
	}
	for _, it := range items {
		if !it.PhotoPending && it.ImageURL != "" {
			m.image = it.ImageURL
			break
		}
	}
	return m
}

// twd formats like the UI's twd(): NT$ and comma-grouped dollars.
func twd(n int64) string {
	s := strconv.FormatInt(max(n, 0), 10)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	return "NT$" + s
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
		"overview": s.Radar.Overview(), "fx": s.FX.Current(), "advisor": advisor.BaseMode(s.Advisor),
		"runs": runs, "crawling": s.crawling.Load(), "can_crawl": s.Crawl != nil && !s.NoCrawlAPI,
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
	if s.Crawl == nil || s.NoCrawlAPI {
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

// maxChatRunes caps one chat message. Shoppers' questions fit in a few
// hundred characters; a pasted dump is cut off here, before it can cost a
// classifier call or an advisor turn.
const maxChatRunes = 1000

// chatError writes a JSON error whose text the panel shows to the shopper.
func chatError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// chat streams advisor events as Server-Sent Events.
func (s *Server) chat(w http.ResponseWriter, r *http.Request) {
	if s.ChatLimit != nil && !s.ChatLimit.Allow(clientIP(r, s.TrustProxy)) {
		s.Log.Info("chat rate limited")
		chatError(w, http.StatusTooManyRequests, "目前詢問的人較多，請稍後再試")
		return
	}
	var req struct {
		Session string `json:"session"`
		Message string `json:"message"`
	}
	tooLong := fmt.Sprintf("訊息太長了，請精簡到 %d 字以內再試一次", maxChatRunes)
	err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10)).Decode(&req)
	if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
		// A huge paste trips the byte cap before the rune cap; the shopper
		// gets the same length message rather than the malformed-request one.
		chatError(w, http.StatusBadRequest, tooLong)
		return
	}
	if err != nil || strings.TrimSpace(req.Message) == "" || req.Session == "" {
		chatError(w, http.StatusBadRequest, "session and message are required")
		return
	}
	if utf8.RuneCountInString(req.Message) > maxChatRunes {
		chatError(w, http.StatusBadRequest, tooLong)
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
		emit(advisor.Event{Type: "error", Text: advisor.Public(err)})
	}
	emit(advisor.Event{Type: "done"})
	s.Log.Info("chat", "mode", s.Advisor.Mode(), "took", time.Since(start).Round(time.Millisecond))
}
