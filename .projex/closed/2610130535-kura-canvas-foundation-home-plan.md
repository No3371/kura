# Kura canvas foundation + Home (guest)

> **Sprint:** 2610130524-kura-interactive-mock-sprint.md — iteration 1
> **Status:** Complete
> **Author:** Claude Opus 5.5 (planner, sprint subagent)
> **Parent:** 2610130524-kura-interactive-mock-sprint.md
> **Source:** Sprint derive, iteration 1 (sprint nav § Current Position, commit `54e6f3d`)
> **Related Projex:** 2610130524-kura-interactive-mock-sprint.md | 2610130449-kura-mock-def.md (data classes, canvas precedence) | 2610130420-kura-roadmap-nav.md (Phase 3 tokens + IA milestones) | 2610130552-kura-canvas-foundation-home-patch.md (execution)
> **Execution:** [PATCHED] Full plan (Steps 1–7, S1–S11) completed via 2610130552-kura-canvas-foundation-home-patch.md — commits `0e472c0`, `169953a`; nothing remains open
> **Worktree:** Yes — existing sprint worktree `kura/.projexwt/2610130524-kura-interactive-mock`, branch `projex/sprint/2610130524-kura-interactive-mock`; execute via patch-projex plan-execution mode, no new branch

---

## Summary

Replace the legacy single-page UI at `/` with the canvas (pinned 1791839497-46ec) foundation — tokens, fonts, light/dark, 1280/375 responsive shell, guest header + footer, listing card in every canvas state incl. 「尚無行情估價」 — and the first canvas screen, Home (行情雷達), on real BRAND OFF data. Every later iteration (browse, detail, model, advisor, member, footer pages) builds on this shell, router and card.

**Scope:** `internal/web` (server routing + new static UI), `internal/radar` + `internal/listing` + `internal/source/brandoff` (three small data fields the card needs). Legacy UI kept reachable at `/legacy/` until later iterations replace advisor/detail.
**Estimated Changes:** ~16 files — 4 Go edited, 1 Go test added, 1 Go test extended, 3 legacy files moved, 7 static files added.

---

## Objective

### Problem / Gap / Need
No canvas screen exists in the running app (legacy tabs + drawer UI; `docs/design-handoff/kura-tokens.css` unused). Home is the first screen a reviewer lands on; nothing else is reachable without it.

### Success Criteria
- [x] **S1 Shell/routing:** `GET /` serves the new canvas shell; extensionless non-API paths (`/browse`, `/listing/x`, `/about`) serve the same shell (client route); missing assets and unknown `/api/*` → 404; `.js` served with a JavaScript MIME type; legacy UI works at `/legacy/`. Proven by `internal/web/server_test.go`.
- [x] **S2 Tokens:** `static/kura.css` carries every `--k-*` token of `docs/design-handoff/kura-tokens.css` (same values) + `--k-chart-mark` (canvas Tokens sheet: light `#7A6B58`, dark `#857868`); dark = OS `prefers-color-scheme` unless `data-theme="light"`; `data-theme="dark|light"` forces either, also on any sub-element; `?theme=dark|light` URL param sets it on `<html>`.
- [x] **S3 Fonts:** the one Google Fonts `<link>` from `kura-tokens.css` loads; display/num/ja/ui/hand families applied per token roles.
- [x] **S4 Guest header** matches canvas `Main.dc.html` (guest): 蔵 seal + `Kura` wordmark → `/`; desktop search pill (placeholder 搜尋型號、編號，可用日文) GET → `/browse?q=`; mobile search icon → `/browse`; 登入 (both) + 加入會員 (desktop only).
- [x] **S5 Footer** matches canvas: seal + wordmark, 6 links (關於 Kura / 行情怎麼算 / 常見問題 / 聯絡我們 / 隱私權政策 / 服務條款), footnote 「價格未含國際運費與進口稅費 · 真偽與商品狀況以賣家說明為準 · 匯率 1 JPY = {rate 4dp} TWD（{yyyy/mm/dd}）」 with the **live** `/api/overview` rate, `© 2026 Kura`.
- [x] **S6 Card** renders all states with canvas `Card.dc.html` look + copy: near | below (high/medium conf) | low | anomaly | above | none (尚無行情估價) | drop | sold | photo pending; real photo in mat; track toggle; shop link-out. `/dev/cards.html` shows all 9 states in light + dark rows + one row-layout card.
- [x] **S7 Home (data):** 最新上架 = 5 newest active listings; 價格最美 = top 3 of the `deals=1&sort=deal` query, link 「看全部價格最美 {total} 件 →」 where total = that same response's `total`; 大家都在找 absent (canvas `trends: hidden` state — see Assumptions A1).
- [x] **S8 Home (empty):** with zero active listings → canvas empty state 「還在準備中。」/「店裡正在整理，新的包很快就會上架。」, no listing sections.
- [x] **S9 Responsive:** 1280 → canvas desktop values (40px gutter, 5-col grids); 375 → canvas mobile values (16px gutter, horizontal card carousels); no horizontal page scroll at 375 (`document.documentElement.scrollWidth <= innerWidth`); interactive targets ≥24px; visible focus ring.
- [x] **S10 Data classes:** every number on Home traces to a real API field (table in Verification); no sample/fixture value on `/`; no real-derived or mocked surface shown.
- [x] **S11** `go vet ./...` + `go test ./...` green; no projex references in any source comment (SKILL.md § Source Hygiene).

### Out of Scope
- Browse, detail, model pages; footer page content (links target final routes, render the router's not-found view for now)
- Advisor dock/panel (canvas 找包顧問 dock) — advisor iteration; legacy advisor stays at `/legacy/#advisor`
- Member header, notifications, tracking persistence, 登入/加入會員 behaviour (sprint Open Question: member-state entry)
- 大家都在找 shelf (mocked demand + real-derived trend/speed) — A1
- Theme toggle UI (canvas designs none), header status line + 近期降價 (def Canvas gaps, undecided)
- Removing the legacy UI; multi-provider provenance (sprint Deferred)

---

## Context

### Current State
- `internal/web/server.go` `Handler`: `mux.Handle("GET /", http.FileServerFS(sub))` over embedded `static/` (`index.html` | `app.js` | `style.css`, legacy tabbed UI; `app.js` calls absolute `/api/...`, HTML references `style.css`/`app.js` relatively → works unchanged from `/legacy/`).
- API (unchanged): `GET /api/overview` → `{overview:{active,sold,models,valued,deals,drops,by_brand,since,built_at}, fx:{Base,Quote,Provider,Rate,AsOf}, advisor, runs, crawling, can_crawl, brands, categories}`; `GET /api/listings?sort=newest|deal|price_asc|price_desc&deals=1&limit=N` → `{total, items:[Item]}`.
- `radar.Item` JSON: `id url image_url source brand brand_zh title model_name model_code group_key category category_zh rank condition{code,label,desc_ja,desc_zh,rank} condition_scheme condition_by price_jpy price_twd status first_seen days_listed fair{fair_jpy,fair_twd,deal_pct,method,comps,confidence,dispersion,anomaly}? drop{from_jpy,to_jpy,pct,at}?`. No shop display name, no TWD for drop prices, no photo-pending flag.
- `radar.Search` `OnlyDeals` predicate: `fair != nil && deal_pct >= 0.05 && confidence != low && !anomaly` — same predicate the advisor's `only_deals` tool param uses (`internal/advisor/tools.go`) → one source for the 價格最美 count.
- Data (`kura.db`): 15,367 active brandoff listings, one sweep; 26 have `image_url` `.../NowPrinting_L.jpg` (real pending state); 0 drops, 0 sold (fixture-only states); FX `0.201837` as of 2026-10-12 (open.er-api.com).
- Canvas home (`project/Main.dc.html`, guest, `page: home`): header → [大家都在找, only when `trends: sample`] → 最新上架 (5 cards) → 價格最美 (3 cards) → footer → advisor dock. With `trends: hidden` the 最新上架 section takes first-section padding and drops its brass top rule.

### Key Files

| File | Role | Change Summary |
|------|------|----------------|
| `internal/web/server.go` | HTTP routes | SPA fallback handler; MIME init |
| `internal/web/server_test.go` | new | static/route tests |
| `internal/web/static/legacy/{index.html,app.js,style.css}` | legacy UI | moved from `static/` |
| `internal/web/static/index.html` | new shell | head, header, `<main id="main">`, footer |
| `internal/web/static/kura.css` | new | tokens + base + header/footer + card + home + responsive |
| `internal/web/static/js/app.js` | new | theme param, router, link interception, footer FX |
| `internal/web/static/js/util.js` | new | `api`, `esc`, number/date format |
| `internal/web/static/js/card.js` | new | valuation state + card markup + track toggle |
| `internal/web/static/js/home.js` | new | Home view |
| `internal/web/static/dev/cards.html` + `dev/cards.js` | new | card-state fixture page |
| `internal/radar/radar.go` | listing view | `Item.SourceName`, `Item.PhotoPending`, `Drop.FromTWD/ToTWD` |
| `internal/radar/radar_test.go` | tests | cover the three fields |
| `internal/listing/condition.go` | grading schemes | `Scheme.Shop` |
| `internal/source/brandoff/condition.go` | BRAND OFF scheme | `Shop: "BRAND OFF"` |

### Dependencies
- **Requires:** nothing beyond sprint branch HEAD `54e6f3d`.
- **Blocks:** browse / detail / model / advisor / member iterations (reuse shell, router, card, tokens).

### Constraints
- HANDOFF §3.6: plain HTML/CSS/JS embedded in the Go binary, no build step, no framework; Google Fonts OK → native ES modules.
- Canvas governs look/IA/copy/states; def governs data semantics (def § Canvas precedence). Card copy = `Card.dc.html` verbatim.
- Source Hygiene: no projex IDs/filenames in comments; non-obvious decisions carry rationale naming the rejected alternative.
- All commits on the sprint branch in the worktree; never `S:\Repos\portfolio\kura` (master).

### Assumptions
- **A1 — 大家都在找 not built.** Canvas designs exactly two states: `trends: sample` (sample values, "for layout only") and `trends: hidden` (section absent). Its tile mixes mocked demand (「這週 N 人在找」, ranking) with real-derived trend/speed, which have no history (one sweep). Ranking definition is open (def L3 "Home sections … definitions of each ranking?"). Sprint rule: open question blocks → defer, don't decide. → Home renders the canvas `trends: hidden` state. Orchestrator should log the shelf under sprint § Deferred (blocker: def L3 ranking definition + crawl history).
- **A2 — Guest track button** rendered per canvas (Card shows it for guests), in-memory toggle only, no persistence (Phase 4 / member entry decides semantics).
- **A3 — 登入 / 加入會員** keep canvas `href="#"` (member-state entry is a sprint Open Question).
- **A4 — Footer routes:** `/about` `/methodology` `/faq` `/contact` `/privacy` `/terms` (slugs are this plan's choice; pages built later).
- **A5 — Valuation thresholds** reuse legacy `app.js` `valBadge` rule: ±3% band = 接近行情.
- **A6 — Photos** hotlinked from BRAND OFF as legacy does (roadmap image-policy question unchanged).
- **A7 — Windows MIME:** `mime` reads the registry on Windows and may type `.js` as `text/plain`, which blocks module scripts → explicit `AddExtensionType` (verify in S1 test).

### Impact Analysis
- **Direct:** `/` UI replaced; legacy URL moves to `/legacy/`.
- **Adjacent:** `radar.Item`/`Drop` JSON gain fields (additive; advisor builds its own tool payload from `Item` fields → unaffected).
- **Downstream:** none outside the repo.

---

## Implementation

### Overview
Server first (fallback + MIME + legacy move, tested), then data fields (tested), then the static UI bottom-up: CSS → shell/router → card → home → fixture page, then the browser verification matrix.

### Step 1: Server routing, MIME, legacy move

**Objective:** S1. **Confidence:** High. **Depends on:** None.
**Do-Projex:** Encouraged

**Files:** `internal/web/server.go`, `internal/web/server_test.go` (new), `internal/web/static/{index.html,app.js,style.css}` → `internal/web/static/legacy/` (`move-n-stage`).

**Changes:**

```go
// Before (Handler):
	sub, _ := fs.Sub(static, "static")
	mux.Handle("GET /", http.FileServerFS(sub))

// After:
	sub, _ := fs.Sub(static, "static")
	mux.Handle("GET /", spaFallback(sub))
```

Add (imports `mime`, `path`):

```go
func init() {
	// On Windows the mime package reads the registry, which can map .js to
	// text/plain; browsers refuse to run a module script served that way.
	mime.AddExtensionType(".js", "text/javascript; charset=utf-8")
	mime.AddExtensionType(".css", "text/css; charset=utf-8")
}

// spaFallback serves embedded files as they are and answers any other
// extensionless GET with index.html, so client routes (/browse, /listing/{id})
// survive a reload or a pasted link. /api/ paths and missing asset files stay
// 404: answering them with index.html would hand HTML to fetch() or <script>.
func spaFallback(root fs.FS) http.Handler {
	files := http.FileServerFS(root)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if name == "" {
			name = "."
		}
		if _, err := fs.Stat(root, name); err == nil {
			files.ServeHTTP(w, r)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/api/") || path.Ext(name) != "" {
			http.NotFound(w, r)
			return
		}
		http.ServeFileFS(w, r, root, "index.html")
	})
}
```

`server_test.go` — `h := (&Server{}).Handler()` (static routes touch no deps), table over `httptest.NewRecorder`:

| Path | Code | Content-Type contains | Body contains |
|---|---|---|---|
| `/` | 200 | `text/html` | `id="main"` |
| `/browse` | 200 | `text/html` | `id="main"` |
| `/listing/brandoff:1` | 200 | `text/html` | `id="main"` |
| `/js/app.js` | 200 | `javascript` | — |
| `/kura.css` | 200 | `text/css` | `--k-paper` |
| `/legacy/` | 200 | `text/html` | `AI 找包顧問` |
| `/js/missing.js` | 404 | — | — |
| `/api/nope` | 404 | — | — |

**Rationale:** history-API routes need server fallback; serving index.html for everything (common SPA shortcut) breaks asset 404s and API errors. Legacy kept (not deleted) so the working advisor stays reachable mid-sprint.

**Verification:** `go test ./internal/web/` (passes once Step 4 adds `index.html`, `kura.css`, `js/app.js` — run after Step 4; compile check now: `go vet ./internal/web/`).

**If this fails:** revert `server.go`; `move-n-stage` back the legacy files.

---

### Step 2: Card data fields

**Objective:** Shop display name, drop prices in TWD (server FX, one source), photo-pending flag. **Confidence:** High. **Depends on:** None.
**Do-Projex:** Encouraged

**Files:** `internal/listing/condition.go`, `internal/source/brandoff/condition.go`, `internal/radar/radar.go`, `internal/radar/radar_test.go`.

**Changes:**

```go
// listing.Scheme — add after Name:
	Shop   string  `json:"shop"`   // the source's display name, e.g. "BRAND OFF"

// brandoff.Scheme — add after Name:
	Shop:   "BRAND OFF",

// radar.Drop — add:
	FromTWD int64 `json:"from_twd"`
	ToTWD   int64 `json:"to_twd"`

// radar.Item — add after Source:
	SourceName string `json:"source_name"` // the shop's display name
// and after ImageURL:
	// PhotoPending: the source has no product photo yet.
	PhotoPending bool `json:"photo_pending,omitempty"`

// toItem: set SourceName: scheme.Shop, PhotoPending: photoPending(r.ImageURL);
// Drop literal gains FromTWD: e.FX.TWD(prev.PriceJPY), ToTWD: e.FX.TWD(last.PriceJPY)

// photoPending reports a missing product photo. BRAND OFF lists items before
// their studio shots exist and points the image at .../NowPrinting_L.jpg.
func photoPending(url string) bool {
	return url == "" || strings.Contains(url, "/NowPrinting")
}
```

Tests: test scheme in `radar_test.go` `init` gains `Shop: "テスト店"`; `TestDropsAndStats` asserts `drops[0].Drop.FromTWD == 24000 && ToTWD == 20000` (FX fallback 0.2) and `drops[0].SourceName == "テスト店"`; new `TestPhotoPending` builds rows with `ImageURL` `""`, `https://x/Contents/ProductImages/NowPrinting_L.jpg`, `https://x/2101220285869_L.jpg` → `true,true,false`.

**Rationale:** client must not re-convert JPY (def: one source per stat; server already rounds TWD via `fx.Converter.TWD`). Placeholder detection lives with the code that already knows the source's quirks' effect on `Item`; no new package.

**Verification:** `go test ./internal/radar/ ./internal/listing/... ./internal/source/... ./internal/store/`.

**If this fails:** revert the four files.

---

### Step 3: `kura.css` — tokens, base, shell, card, home

**Objective:** S2, S3 (CSS side), S9 layout values. **Confidence:** High. **Depends on:** None.

**Files:** `internal/web/static/kura.css` (new).

**Changes — sections in order:**

1. **Tokens.** Copy `docs/design-handoff/kura-tokens.css` `:root` block + both dark blocks with values unchanged, with these selector changes so a sub-tree can force a theme (fixture page, later artboard checks):
   ```css
   :root, [data-theme="light"] { color-scheme: light; /* all light tokens */ --k-chart-mark: #7A6B58; }
   @media (prefers-color-scheme: dark) { :root:not([data-theme="light"]) { color-scheme: dark; /* dark tokens */ --k-chart-mark: #857868; } }
   [data-theme="dark"] { color-scheme: dark; /* dark tokens */ --k-chart-mark: #857868; }
   ```
   Type/space/radius tokens stay on `:root` only.
2. **Base** (tokens file + additions): `body { margin:0; background: var(--k-paper); color: var(--k-ink); font-family: var(--k-font-ui); -webkit-font-smoothing: antialiased }`, `[data-theme] { background: var(--k-paper); color: var(--k-ink) }`, `:focus-visible` rule from tokens, `.k-sr` visually-hidden utility, `.k-link:hover { color: var(--k-accent-ink); text-decoration: underline; text-underline-offset: 3px }`, `.k-wrap { max-width: 1280px; margin: 0 auto; padding: 0 40px; box-sizing: border-box }`.
3. **Tokens-file components** verbatim: `.k-price`, `.k-badge`, `.k-badge--below|low|anomaly|near|above|none`, `.k-grade`, `.k-grade--seller`, `.k-note`.
4. **Header** (canvas desktop): `display:flex; flex-wrap:wrap; align-items:center; justify-content:space-between; gap:16px 32px; padding:22px 0; border-bottom:1px solid var(--k-rule)`. Logo: seal 36×36 accent bg, on-accent `蔵` font-ja 22px, radius 3px, `box-shadow: inset 0 0 0 2px var(--k-accent), inset 0 0 0 3px var(--k-on-accent)`; wordmark font-num 26px, letter-spacing .06em, gap 12px. Search pill: `flex:1 1 280px; max-width:440px; height:40px; padding:0 14px; border:1px solid var(--k-rule-strong); border-radius:999px; background:var(--k-surface)`, 16px magnifier (ink-3), borderless transparent input 14px. 登入: 40px tall, `padding:0 12px`, 14px ink. 加入會員: 40px, `padding:0 16px`, 14px on-accent on accent, radius 2px.
5. **Footer:** `margin-top:72px; background:var(--k-paper-sunk); border-top:1px solid var(--k-brass)`; inner `.k-wrap` `padding-top:40px; padding-bottom:104px; display:flex; flex-direction:column; gap:28px`; top row flex wrap space-between gap `24px 48px`; seal 30×30 / 18px, wordmark 22px; nav `display:grid; grid-template-columns:repeat(3, minmax(0,max-content)); gap:4px 40px; font-size:14px`, links ink-2, `min-height:32px`; bottom row `border-top:1px solid var(--k-rule); padding-top:16px; flex wrap space-between; gap:6px 24px; font-size:12px; line-height:1.7; color:var(--k-ink-3)`.
6. **Home sections** (class names used by Steps 4–6): `.k-sec__head` row `display:flex; flex-wrap:wrap; justify-content:space-between; align-items:baseline; gap:6px 16px; margin-bottom:22px`; `.k-sec h2` font-display 400, 30px, line-height 1.3, margin 0; `.k-sec__more` link 13px ink-2, `min-height:24px; display:inline-flex; align-items:center`. `.k-sec--first` (最新上架 while trends hidden) `padding:44px 0 64px`, no rule. `.k-sec--ruled` (價格最美) `padding:0 0 64px`, its `.k-sec__head` `padding-top:28px; border-top:1px solid var(--k-brass)`. Card row `.k-row { display:grid; grid-template-columns:repeat(5, minmax(0,1fr)); gap:28px }`. `.k-empty` and `.k-notfound`: `padding:96px 0 72px; max-width:620px`; h1 font-display 400 40px/1.3 margin `0 0 16px`; p 16px/1.8 ink-2. Header classes: `.k-header`, `.k-logo`, `.k-seal`, `.k-wordmark`, `.k-search`, `.k-search-icon`, `.k-guest`, `.k-join`; footer: `.k-footer`.
7. **Card** (canvas `Card.dc.html`, stack layout; row modifier for later browse-mobile):
   ```css
   .k-card { display:flex; flex-direction:column; gap:14px; min-width:0; font-family:var(--k-font-ui); color:var(--k-ink) }
   .k-card--row { display:grid; grid-template-columns:116px minmax(0,1fr); gap:16px; align-items:start }
   .k-card__media { position:relative }
   .k-card__mat { aspect-ratio:1/1; box-sizing:border-box; padding:10%; background:var(--k-mat); border:1px solid var(--k-rule);
     box-shadow:inset 0 0 0 6px var(--k-mat), inset 0 0 0 7px var(--k-mat-line);
     display:flex; flex-direction:column; align-items:center; justify-content:center; gap:8px; color:var(--k-mat-ink) }
   .k-card__mat img { width:100%; height:100%; object-fit:contain; mix-blend-mode:multiply }
   .k-card.is-sold .k-card__mat { opacity:.6 }
   .k-card__pending { font-size:12px; color:#7D7262; letter-spacing:.04em }
   .k-card__track { position:absolute; top:10px; right:10px; width:34px; height:34px; padding:0; border-radius:50%;
     border:1px solid var(--k-rule); background:var(--k-surface); color:var(--k-ink-2); display:flex; align-items:center; justify-content:center; cursor:pointer }
   .k-card__track[aria-pressed="true"] { color:var(--k-accent-ink) }
   .k-card__track[aria-pressed="true"] svg { fill:currentColor }
   .k-card__sold { position:absolute; top:16px; left:-6px; transform:rotate(-4deg); background:var(--k-paper); border:1px solid var(--k-brass);
     padding:5px 12px 5px 10px; display:flex; align-items:center; gap:8px; box-shadow:var(--k-shadow); font-family:var(--k-font-hand); font-size:15px; color:var(--k-ink) }
   .k-card__sold::before { content:""; width:7px; height:7px; border-radius:50%; border:1px solid var(--k-brass) }
   .k-card__body { display:flex; flex-direction:column; gap:8px; min-width:0 }
   .k-card__brand { font-family:var(--k-font-num); font-size:12px; letter-spacing:.16em; color:var(--k-ink-3) }
   .k-card__model { font-family:var(--k-font-ja); font-size:19px; line-height:1.35; color:var(--k-ink); text-decoration:none; font-weight:600 }
   .k-card--row .k-card__model { font-size:16px }
   .k-card__grade { display:flex; gap:8px; align-items:baseline; min-width:0 }
   .k-card__grade .k-grade { color:var(--k-ink) }
   .k-card__gradezh { font-size:13px; color:var(--k-ink-2); overflow:hidden; text-overflow:ellipsis; white-space:nowrap; min-width:0 }
   .k-card__price { display:flex; align-items:center; gap:10px; flex-wrap:wrap; margin-top:2px }
   .k-card__price .k-price { font-size:16px; line-height:1.2; font-weight:400 }
   .k-card--row .k-card__price .k-price { font-size:15px }
   .k-card.is-sold .k-price { color:var(--k-ink-3) }
   .k-card__lastsold { font-size:12px; color:var(--k-ink-3) }
   .k-card .k-badge { font-size:12px; padding:0 6px; line-height:20px; gap:5px }
   .k-card .k-badge--near, .k-card .k-badge--above, .k-card .k-badge--none { padding:0 }
   .k-card__drop { display:flex; align-items:center; gap:6px; font-size:12px; color:var(--k-ink-2); font-variant-numeric:tabular-nums }
   .k-card__drop s { color:var(--k-ink-3) }
   .k-card__shop { align-self:flex-start; display:inline-flex; align-items:center; gap:4px; min-height:24px; font-size:12px; color:var(--k-ink-3); text-decoration:none; letter-spacing:.04em }
   ```
8. **Responsive.** `@media (max-width: 1023px) { .k-row { grid-template-columns: repeat(3, minmax(0,1fr)) } }`. `@media (max-width: 767px)` (canvas mobile): `.k-wrap { padding: 0 16px }`; header `flex-wrap:nowrap; gap:4px; padding:14px 0 0` (canvas `headerPad`), hide `.k-search` + `.k-join`, show `.k-search-icon` (40×40, `margin-left:auto`, ink-2, 20px icon; hidden ≥768); h2 24px; 最新上架 `padding:24px 0 40px`; 價格最美 `padding:0 0 40px`; `.k-row { display:flex; gap:16px; overflow-x:auto; scrollbar-width:none }` with `.k-row > * { flex:0 0 200px; min-width:0 }`; footer `padding-bottom:96px`, nav 2 columns; empty-state h1 stays 40px (canvas has no mobile empty artboard; check no overflow at 375).

**Rationale:** values transcribed from canvas `Main.dc.html` `renderVals` (`padX`, `headerPad`, `freshPad`, `sectionPad`, `h2Size`, `freshRow`, `freshItem`, `footPadB`, `footCols`) and `Card.dc.html`; 768–1023 has no artboard → 3 columns keeps cards readable.

**Verification:** visual, Step 7.

**If this fails:** file is new — fix forward or delete.

---

### Step 4: Shell, router, util

**Objective:** S1 (client side), S2 theme param, S3, S4, S5. **Confidence:** High. **Depends on:** Step 1, Step 3.

**Files:** `internal/web/static/index.html`, `internal/web/static/js/util.js`, `internal/web/static/js/app.js` (all new).

**Changes:**

- `index.html`: `<!doctype html><html lang="zh-Hant-TW">`; head: charset, viewport, `<title>Kura 蔵 · 行情雷達</title>`, legacy favicon data-URI, inline theme script **before** CSS:
  `<script>try{var t=new URLSearchParams(location.search).get("theme");if(t==="light"||t==="dark")document.documentElement.dataset.theme=t}catch(e){}</script>`,
  Google Fonts `<link>` (URL from `kura-tokens.css` header comment), `<link rel="stylesheet" href="/kura.css">`, `<script type="module" src="/js/app.js"></script>`. All asset URLs absolute (routes are nested).
  Body: `<div class="k-wrap">` → `<header class="k-header">` (logo `<a href="/" aria-label="Kura 蔵 首頁">`; `<form class="k-search" role="search" action="/browse" method="get">` with sr-only label 搜尋 + `<input type="search" name="q" placeholder="搜尋型號、編號，可用日文">`; `<a class="k-search-icon" href="/browse" aria-label="搜尋">`; `<div class="k-guest">` with `<a href="#" class="k-link">登入</a>` + `<a href="#" class="k-join">加入會員</a>`) → `<main id="main"></main>` → close wrap; `<footer class="k-footer">` (logo → `/`; `<nav aria-label="頁尾">` links per A4 order 關於 Kura `/about`, 行情怎麼算 `/methodology`, 常見問題 `/faq`, 聯絡我們 `/contact`, 隱私權政策 `/privacy`, 服務條款 `/terms`; bottom row `<span>價格未含國際運費與進口稅費 · 真偽與商品狀況以賣家說明為準<span id="fx"></span></span><span>© 2026 Kura</span>`). SVG paths copied from canvas header/footer.
- `util.js` (exports): `esc(s)` (HTML-escape `& < > " '`); `num(n)` → `Intl.NumberFormat("en-US")` (canvas `toLocaleString('en-US')`); `twd(n)` → `"NT$" + num(Math.round(n))`; `pct(f)` → `Math.round(f*100)`; `ymd(iso)` → `yyyy/mm/dd` in `Asia/Taipei`; `api(path)` → fetch JSON, throw `Error(body.error || statusText)` on non-OK.
- `app.js` (module):
  - `routes = [[/^\/$/, renderHome]]` (`import { renderHome } from "./home.js"` — file created in Step 6; later iterations append routes); unmatched → `renderNotFound(main)`: `<section class="k-notfound">` h1 「找不到這個頁面」 + `<a href="/" class="k-link">回到首頁</a>` (font-display; padding as empty state).
  - `render()` → `const ov = await overview()` (memoised `/api/overview` promise) → matched view `(main, ov)`; on error → `<p>` 「暫時無法載入資料，請稍後再試。」 (ink-2).
  - Link interception: document `click` on `a[href^="/"]` with no `target`/`download`, button 0, no modifier keys, path not starting `/legacy` `/dev` `/api` → `preventDefault`, `history.pushState`, `render()`, `scrollTo(0,0)`. `popstate` → `render()`.
  - Footer FX: `#fx` text = ` · 匯率 1 JPY = ${ov.fx.Rate.toFixed(4)} TWD（${ov.fx.Provider === "fallback" ? "預設匯率" : ymd(ov.fx.AsOf)}）`.
  - `document.title` per view (Home: `Kura 蔵 · 行情雷達`).

**Rationale:** history-API SPA (vs Go `html/template` pages) so the advisor panel (later iteration) persists across navigation as the canvas dock implies; ES modules = no build step. Header/footer static markup → no layout flash; only FX is dynamic.

**Verification:** `go test ./internal/web/` (Step 1 table) passes; page loads with no console errors.

**If this fails:** delete the new files; `/` then 404s until fixed (legacy unaffected).

---

### Step 5: Card module

**Objective:** S6. **Confidence:** High. **Depends on:** Step 2, Step 3, Step 4.
**Verify-Projex:** Encouraged

**Files:** `internal/web/static/js/card.js` (new).

**Changes:**

```js
import { esc, twd, pct } from "./util.js";

const tracked = new Set(); // in-memory only; persistence belongs to the member layer

// Card valuation state. ±3% around fair value reads as 接近行情.
export function valuation(it) {
  if (it.status !== "active") return { state: "sold" };
  const f = it.fair;
  if (!f) return { state: "none" };
  if (f.anomaly) return { state: "anomaly" };
  if (f.deal_pct >= 0.03) return { state: f.confidence === "low" ? "low" : "below", pct: pct(f.deal_pct) };
  if (f.deal_pct <= -0.03) return { state: "above", pct: pct(-f.deal_pct) };
  return { state: "near" };
}

const BADGE = {
  below: (v) => `<span class="k-badge k-badge--below"><svg width="8" height="8" viewBox="0 0 10 10" aria-hidden="true"><path d="M1 2.5h8L5 8.5z" fill="currentColor"/></svg>低於行情 ${v.pct}%</span>`,
  low: () => `<span class="k-badge k-badge--low">可能低於行情 · 低信心</span>`,
  anomaly: () => `<span class="k-badge k-badge--anomaly"><svg width="11" height="11" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.6" aria-hidden="true"><path d="M8 2 1.5 13.5h13z"/><path d="M8 6.5v3.3M8 11.6v.4"/></svg>異常低價 · 先確認狀況</span>`,
  near: () => `<span class="k-badge k-badge--near">接近行情</span>`,
  above: (v) => `<span class="k-badge k-badge--above">高於行情 ${v.pct}%</span>`,
  none: () => `<span class="k-badge k-badge--none">尚無行情估價</span>`,
  sold: () => "",
};

export function cardHTML(it, { layout = "stack" } = {}) { /* returns <article class="k-card [k-card--row] [is-sold]" data-id>; track button reflects `tracked` */ }

// One document-level listener, installed at module load: re-rendered views
// would stack duplicate listeners (double toggle) if each render bound its own.
document.addEventListener("click", (e) => {
  /* closest("[data-track]") → toggle `tracked`, aria-pressed, aria-label + title 追蹤這件 ↔ 追蹤中，點一下取消 */
});
```

`cardHTML` markup (canvas `Card.dc.html` order):
1. `.k-card__media` → `.k-card__mat`: `photo_pending` → dashed bag SVG (44px, `stroke-dasharray="3 3"`, canvas paths) + `<span class="k-card__pending">店家照片準備中</span>`; else `<img src=image_url alt="{model}" loading="lazy">`. Not sold → track `<button type="button" class="k-card__track" data-track="{id}" aria-pressed aria-label title>` with heart SVG (canvas path, `fill="none"`). Sold → `<div class="k-card__sold">已找到新主人</div>`.
2. `.k-card__body`: brand `<span class="k-card__brand">{brand}</span>`; `<a class="k-card__model k-link" href="/listing/{encodeURIComponent(id)}">{model_name || title}</a>`; grade row only if `condition.label`: `<span class="k-grade">{label}</span><span class="k-card__gradezh">{desc_zh}</span>`; price row: sold → `<span class="k-card__lastsold">最後售價</span>`, `<span class="k-price">{twd(price_twd)}</span>`, `BADGE[state]`; drop line if `drop && !sold`: trend SVG + `已降價，原價` + `<s>{twd(drop.from_twd)}</s>`; shop link `<a class="k-card__shop k-link" href="{url}" target="_blank" rel="noopener noreferrer">{source_name || source}<span class="k-sr">（另開視窗）</span>{arrow svg}</a>`.
All interpolated strings through `esc`.

**Rationale:** card copy/structure = canvas; state derivation client-side from server fields (no new server classification — advisor already reasons from the same `deal_pct/confidence/anomaly`).

**Verification:** Step 6 fixture page shows each state; spot-check real cards on `/` against `/api/listings` JSON.

**If this fails:** delete file.

---

### Step 6: Home view + card-state fixture page

**Objective:** S7, S8, S6 fixture. **Confidence:** High. **Depends on:** Step 5.

**Files:** `internal/web/static/js/home.js`, `internal/web/static/dev/cards.html`, `internal/web/static/dev/cards.js` (new); `app.js` imports `renderHome`.

**Changes:**

- `home.js` `export async function renderHome(main, ov)`:
  - `ov.overview.active === 0` → empty state (`<section class="k-empty"><h1>還在準備中。</h1><p>店裡正在整理，新的包很快就會上架。</p></section>`), return.
  - `const [fresh, best] = await Promise.all([api("/api/listings?sort=newest&limit=5"), api("/api/listings?deals=1&sort=deal&limit=3")])`.
  - Markup: `<h1 class="k-sr">Kura 行情雷達</h1>` (canvas h1 lives in the hidden 大家都在找 section); `<section aria-labelledby="new-h" class="k-sec k-sec--first">` → `.k-sec__head` (h2#new-h `最新上架` + `<a class="k-sec__more k-link" href="/browse?sort=newest">看全部最新上架 →</a>`) + `<div role="list" class="k-row">` of `<div role="listitem">{cardHTML}</div>`; then, if `best.total > 0`, `<section aria-labelledby="deal-h" class="k-sec k-sec--ruled">` → head (h2#deal-h `價格最美` + `<a class="k-sec__more k-link" href="/browse">看全部價格最美 ${num(best.total)} 件 →</a>`) + row. `best.total === 0` → section omitted (canvas has no zero-deal state).
- `dev/cards.html`: standalone page (same head as shell: fonts, `/kura.css`, theme script) titled `Kura 商品卡狀態`; two `<section data-theme="light|dark">` rows, each h2 「商品卡 · 日間|夜間」 + caption 「狀態示意：資料取自範例商品；降價、售出、照片準備中為示意，不是真實紀錄」 (canvas `CardStates.dc.html`), grid `repeat(auto-fill, minmax(200px, 1fr))` gap 32px of labelled cards; plus one `layout: "row"` card at `max-width: 375px`.
- `dev/cards.js`: fixtures in `radar.Item` JSON shape from HANDOFF §5 real sample listings + photo URLs (ピコタンロックPM / マトラッセ / ネヴァーフルMM / ガーデンパーティPM / アルマPM). Labels + states: 預設 · 接近行情 (`deal_pct: 0.01`) | 低於行情 · 高信心 (`0.31`, high) | 可能低於行情 · 低信心 (`0.47`, low) | 異常低價 (`anomaly: true`) | 高於行情 (`-0.12`) | 尚無行情估價 (`fair` absent) | 降價 (`0.32`, `drop:{from_twd:27854,to_twd:26239}`) | 已找到新主人 (`status:"gone"`) | 照片準備中 (`photo_pending:true`).

**Rationale:** drop/sold have no real instance (one sweep) → fixture page is the only way to verify those states; separate `/dev/` page keeps samples off `/` (def: no sample values in product surfaces), labelled with the canvas's own disclaimer.

**Verification:** Step 7 matrix.

**If this fails:** delete files.

---

### Step 7: End-to-end browser verification

**Objective:** S4–S10 evidence. **Confidence:** Medium (font loading / hotlinked photos depend on network). **Depends on:** Steps 1–6.
**Verify-Projex:** Encouraged

**Procedure:**
1. `go vet ./...` + `go test ./...`.
2. Background: `go run ./cmd/kura serve -db kura.db -addr 127.0.0.1:8090` (worktree root). Empty DB: `go run ./cmd/kura serve -db <scratchpad>/empty.db -addr 127.0.0.1:8091` (path outside the worktree; auto-migrates; delete after).
3. Browser pane matrix — `/` at 1280×900 and 375×812, each `?theme=light` and `?theme=dark` (also once via colorScheme emulation without the param); `/dev/cards.html` at 1280; `/` on :8091 at 1280 light + 375 dark; `/legacy/` once; `/browse` + footer link → not-found view (client route, no full reload for footer clicks).
4. Per view: screenshot; `read_console_messages` onlyErrors empty; at 375 `document.documentElement.scrollWidth <= innerWidth`; Tab through header → focus ring visible.
5. Stop servers; remove `empty.db*`; worktree `git status` clean except committed work.

---

## Verification Plan

### Automated Checks
- [x] `go vet ./...` clean
- [x] `go test ./...` green (incl. new `internal/web` routing test, radar field tests)

### Manual Verification
- [x] Home 1280/375 × light/dark matches canvas `Main.dc.html` (`trends: hidden`, guest) / `HomeMobile` / `HomeDark` (guest header instead of member) layout values
- [x] `/dev/cards.html`: 9 states × 2 themes match `CardStates.dc.html` + Tokens sheet 尚無行情估價; track toggle flips aria-pressed + colour
- [x] Empty DB home = `HomeEmpty.dc.html`
- [x] `/legacy/` loads tiles + advisor tab
- [x] Source hygiene: `grep -rnE "[0-9]{10}-[a-z0-9-]+-(plan|sprint|def|nav)" internal cmd` → no hits

### Acceptance Criteria Validation

| Criterion | How to Verify | Expected Result |
|---|---|---|
| S1 | `go test ./internal/web/` | table passes |
| S2/S3 | computed style of `body` bg in light/dark; Network shows fonts CSS 200 | `#F5F0E6` / `#1B1815`; Noto Serif TC etc. applied |
| S4/S5 | screenshots vs canvas header/footer; footer FX text vs `/api/overview` `fx` | `匯率 1 JPY = 0.2018 TWD（2026/10/12）` |
| S6 | `/dev/cards.html` + ≥1 real NowPrinting card reachable via `/api/listings` (`photo_pending`) | all states render with canvas copy |
| S7 | compare 5 + 3 card ids/prices/badges to `/api/listings?sort=newest&limit=5` and `?deals=1&sort=deal&limit=3`; count = `total` | identical |
| S8 | :8091 screenshot | empty-state copy only |
| S9 | scrollWidth check at 375; 5 cols at 1280 | no overflow; 5 columns |
| S10 | trace table below | every number ← API field |
| S11 | vet/test/grep | green / none |

**S10 trace (Home):** card price ← `price_twd` (real) | badge % ← `fair.deal_pct` (real — radar valuation of current listings) | grade ← `condition.label/desc_zh` (real) | shop ← `source_name` (real) | 價格最美 count ← `/api/listings?deals=1` `total` (real; same predicate as advisor `only_deals`) | footer FX ← `/api/overview` `fx` (real). Real-derived: none shown. Mocked: none shown.

---

## Rollback Plan

1. `git -C <work-root> log --oneline 54e6f3d..HEAD` → identify this plan's commits.
2. `git -C <work-root> revert <sha>…` newest first (sprint branch only); legacy UI returns to `/`.

---

## Notes

### Split Decision
No split — single scope (`internal/` code, one `.projex` folder), within size budget (~470 lines / ~38 KB; 7 steps — heuristic needs both). Steps tightly coupled: one shell, one card, one view, each consuming the previous.

### Risks
- Google Fonts unreachable → fallback stacks in tokens; not a failure.
- BRAND OFF hotlink blocked → broken images; legacy has the same exposure (A6).
- Windows MIME (A7) — covered by S1 test.
- `/dev/cards.html` reachable in a hosted build → harmless, labelled; Phase 6 may strip it.

### Open Questions
None blocking execution. Decisions this plan made that the orchestrator may want logged in the sprint nav: A1 (大家都在找 → Deferred with blocker), A2–A4.
