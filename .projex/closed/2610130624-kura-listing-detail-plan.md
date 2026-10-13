# Kura listing detail page (商品頁)

> **Sprint:** 2610130524-kura-interactive-mock-sprint.md — iteration 2
> **Status:** Complete
> **Author:** Claude Opus 5.5 (planner, sprint subagent)
> **Parent:** 2610130524-kura-interactive-mock-sprint.md
> **Source:** Sprint derive, iteration 2 (sprint nav § Current Position + Revision Log, commit `5ad2212`)
> **Related Projex:** 2610130524-kura-interactive-mock-sprint.md | 2610130449-kura-mock-def.md (data classes, canvas precedence, Canvas gaps) | 2610130420-kura-roadmap-nav.md (Phase 3 New IA) | 2610130535-kura-canvas-foundation-home-plan.md (shell, router, card this builds on) | 2610130602-kura-canvas-foundation-home-audit.md (F1/F6 — not touched here) | 2610130634-kura-listing-detail-patch.md (execution record)
> **Execution:** all steps + criteria S1–S11 completed via 2610130634-kura-listing-detail-patch.md [PATCHED] — commits `766218d`, `b12c30a`; deviations listed there
> **Worktree:** Yes — existing sprint worktree `kura/.projexwt/2610130524-kura-interactive-mock`, branch `projex/sprint/2610130524-kura-interactive-mock`; execute via patch-projex plan-execution mode, no new branch

---

## Summary

Build the canvas 商品頁 (pinned 1791839497-46ec, artboards `Detail` desktop light + `DetailMobileDark`) at `/listing/{id}` on real BRAND OFF data: breadcrumb, photo, info + CTA + track, 鑑定卡 with the shop's grade ladder, 商品資料, 行情估價 prose, 同型號價格分布 (chart + table, real in-sale data), 售價變化. Closes Home's dead-end card click and builds the distribution component the model page (next) reuses.

**Scope:** `internal/web` (route, detail + distribution views, CSS, listing API field), `internal/radar` + `internal/listing` + `internal/source/brandoff` (server-side TWD for grade stats + price history, grade short label), `internal/advisor` (consume the same TWD fields). One `.projex` scope.
**Estimated Changes:** 14 files — 7 Go (incl. `radar_test.go` extended, `server_test.go` new test), 5 static edited (`util.js`, `app.js`, `card.js`, `kura.css`, legacy `app.js` key rename), 2 static added (`detail.js`, `dist.js`).

---

## Objective

### Problem / Gap / Need
Router has one route (`/`); every Home card links to `/listing/{id}` → not-found view. Canvas detail is the next unblocked flow (sprint nav blocking map: Browse waits on F1; Detail unblocked; Model after Detail).

### Success Criteria
- [x] **S1 Route:** `/listing/{id}` renders the detail view via in-app click (Home card) and direct load/reload; title `{model} · {brand} · Kura 蔵`; unknown id or malformed encoding → router not-found view + title 「找不到這個頁面 · Kura 蔵」; other API errors → existing error line.
- [x] **S2 Top block** matches canvas `Main.dc.html` detail (values in Step 4): breadcrumb 首頁 / `{BRAND} {model}` (→ model route, A1) / `{code} · {grade}`; square mat photo (or 店家照片準備中 pending state), no thumbs strip (1 image per listing); brand + 中文 brand, h1 model, `{類型} · {code}`; NT$ price + ¥ price; valuation badge (card states) + 「行情 {fair} · {信心}」 link → `#fair`; CTA 「到 {shop} 商品頁」 (new tab); 追蹤這件 ↔ 追蹤中 toggle kept in sync with any card of the same listing; fine print.
- [x] **S3 鑑定卡:** 鑑定卡 | shop; grade label; zh meaning; 「ja definition」; ladder = shop scheme grades best-first, short labels (A3), this grade highlighted, `aria-label` with full labels; 店家鑑定 + scheme name; 店鑑 stamp when graded by the shop. Parts with no data omitted (A11).
- [x] **S4 商品資料:** rows in canvas order — 型號編號 / 類型 / 顏色 / 材質 / 尺寸 / 付屬品 / 店家備註 / 上架 — only rows with shop data; 店家備註 original text only; 上架 `{yyyy/mm/dd} · 在架第 {days_listed+1} 天`.
- [x] **S5 行情估價:** prose per state (Step 5 table) from `fair` fields; dl 方法 / 樣本 / 信心 / 行情怎麼算 (→ `/methodology`); no-estimate copy when `fair` absent; section absent for sold listings.
- [x] **S6 同型號價格分布:** from `/api/models/{brand}/{group_key}` items (real, never canvas sample); one column per shop grade best-first; filled = in-sale, ring = sold, accent = this listing + 「此件」 label; median tick per grade = `by_grade[].median_twd`; y gridlines in 萬; hover tooltip `{grade} · {status}` + price; 以表格檢視 ↔ 以圖表檢視 toggles a table (等級 / 店家定義 / 在售 / 中位數 / 最低 / 最高) from `by_grade` (TWD from server). Section absent when the listing has no model group.
- [x] **S7 售價變化:** 1 price point (all real data) → canvas no-change strip (`{m/d} 上架` · 今天 · price) + 「上架以來還沒有調價。店家調價時，會畫在這條線上。」; ≥2 points → step line (A6), proven on a scratch DB copy.
- [x] **S8 Sold listing** (A5, scratch DB copy): 最後售價 + 已找到新主人, no badge/track/行情估價, CTA kept.
- [x] **S9 Responsive/themes:** 1280 (canvas desktop values) and 375 (canvas mobile values) × light/dark; no horizontal page scroll at 375 (table scrolls inside its own box); interactive targets ≥24px; visible focus ring.
- [x] **S10 One source per stat:** grade-stat and price-history TWD computed once in `radar`; advisor `model_price_stats` + `get_listing` emit those same fields; detail does no JPY→TWD math; page renders no `sell_through`, `sold`, or `median_days_to_sell` value (real-derived, no history — sprint nav hazard).
- [x] **S11** `go vet ./...` + `go test ./...` green (new/extended tests in Steps 1–2); no projex references in source comments.

### Out of Scope
- Model page (`/model/...` route): breadcrumb link targets it and renders not-found until the model iteration
- Photo gallery / thumbs strip (1 `image_url` per listing; roadmap gallery milestone + def L3 gallery question)
- Translation of 店家備註 (no source), multi-provider provenance (sprint Deferred), member header / notifications, advisor dock, 行情怎麼算 page content, Browse (F1)
- Any change to radar valuation, deal predicate, anomaly cutoff (F1/F6 open)

---

## Context

### Current State
- `GET /api/listings/{id}` (`server.go` `listing`) → `{item, model}`: `item` = full `radar.Item` incl. `history` (`[]store.PricePoint`, JSON keys `At`/`PriceJPY`, no TWD); `model` = `ModelStats` or null (no group key). `ByGrade[]` carries `median_jpy/min_jpy/max_jpy` only. Unknown id → 404 `{"error":"listing not found"}`.
- `GET /api/models/{brand}/{key}` → `{model, items}` — every item of the group (active first, then gone), full fields. Largest group CHANEL `name:マトラッセ` = 409 items ≈ 490 KB.
- Grade scheme (`listing.Scheme`: `source name shop graded ref grades[]`) not exposed by any endpoint; `Item.condition` = one `listing.Grade{code,label,desc_ja,desc_zh,rank}`.
- Advisor converts the same JPY values itself: `tools.go` `modelStats` → `t.FX.TWD(g.MedianJPY|MinJPY|MaxJPY)`; `getListing` → `t.FX.TWD(p.PriceJPY)`.
- Client (iteration 1): `app.js` routes `[[/^\/$/, renderHome, title]]`, views `(el, ov)`, title set before render; `util.js` `api()` throws `Error(body.error)` without status; `card.js` exports `valuation`, `cardHTML`; track Set + document listener private to `card.js`.
- Legacy UI (`/legacy/`) reads `history[].At` / `.PriceJPY` in `historyChart`.
- Data (worktree `kura.db`, plan-time probe): 15,367 active, 0 gone, 1 price point each; grades ランクA 9,321 / ランクB 5,170 / 未使用品 596 / 新品 235 / ランクC 44 / empty 1; 3,757 single-listing groups; specs on 158, notes on 15. All valuation states present (same-grade high/medium/low, rank-adjusted medium/low, anomaly, none).
- Canvas detail (`project/Main.dc.html`, `page: detail`): crumb → `detailTop` (photo column | info column) → `max-width:880px` column: 商品資料 → 行情估價 → 同型號價格分布 (shared with model page, `kuraChart`) → 售價變化. Artboard title 「分布圖 = 示意數據」 → sample `KURA_DIST` never shipped.

### Key Files

| File | Role | Change Summary |
|------|------|----------------|
| `internal/listing/condition.go` | grade type | `Grade.Short` |
| `internal/source/brandoff/condition.go` | BRAND OFF scheme | short labels per canvas ladder |
| `internal/radar/radar.go` | listing index | `PricePoint` w/ TWD; `GradeStat` TWD fields |
| `internal/radar/radar_test.go` | tests | assert new TWD fields |
| `internal/advisor/tools.go` | advisor tools | read radar TWD fields |
| `internal/web/static/legacy/app.js` | legacy UI | history key rename |
| `internal/web/server.go` | API | listing response gains `scheme` |
| `internal/web/server_test.go` | tests | `TestListingAPI` |
| `internal/web/static/js/util.js` | helpers | `api` error status; `jpy`, `md` |
| `internal/web/static/js/app.js` | router | path params, view-returned title, 404 → not-found; detail route |
| `internal/web/static/js/card.js` | card | export `badgeHTML`, `isTracked`, icons; text-style track buttons |
| `internal/web/static/kura.css` | styles | `[hidden]`; detail, 鑑定卡, distribution, history, mobile |
| `internal/web/static/js/detail.js` | new | detail view |
| `internal/web/static/js/dist.js` | new | distribution chart + table + interactions |

### Dependencies
- **Requires:** sprint branch HEAD `5ad2212` (iteration 1 shell/router/card/`spaFallback`).
- **Blocks:** model page iteration (reuses `dist.js`, model route URL A1).

### Constraints
- HANDOFF §3.6: no build step / framework → native ES modules; charts hand-drawn (canvas uses positioned HTML; history step line uses inline SVG).
- Canvas governs look/IA/copy/states; def governs data semantics. Detail artboards are `user: member`; guest header (Home's) stays (sprint nav).
- Client never converts JPY → TWD (iteration-1 rule; one source per stat).
- Source Hygiene: no projex IDs/filenames in comments; non-obvious decisions name the rejected alternative.
- No crawls; commits on the sprint branch in the worktree only.

### Assumptions
- **A1 — Model route URL** `/model/{encodeURIComponent(brand)}/{encodeURIComponent(group_key)}`, mirroring `/api/models/{brand}/{key}` (group keys need the brand; `name:` keys contain `:`). `spaFallback` serves the shell for it (no asset dir `model/`, no asset ext). Model iteration adopts it or edits the one breadcrumb href.
- **A2 — Guest header** on detail (canvas artboard shows member; member entry is a sprint Open Question).
- **A3 — Ladder short labels** = canvas copy (新品 / 未使用 / A / B / C) via new `Grade.Short` set by the BRAND OFF adapter; empty → `label`. Full shop labels stay on the 鑑定卡 label, the ladder `aria-label`, chart axis and table, so no grade is relabelled.
- **A4 — Plan-authored copy** (canvas draws only same-grade/below/high): rank-adjusted sentence, near/above sentences, generic low-confidence line, 信心 dd 中/低, 方法 「跨等級換算」. All other strings verbatim from canvas, HANDOFF §5, or legacy UI. Orchestrator: log as copy for canvas pass 2 review.
- **A5 — Sold detail state** undesigned on canvas → built only from card's designed sold vocabulary (已找到新主人, 最後售價, faded mat). No real instance (0 gone). Orchestrator: add "detail sold state" to def § Canvas gaps.
- **A6 — Multi-point 售價變化** undesigned on canvas (draws the no-change state only) → step line per HANDOFF §4.B/§6 in the strip's own vocabulary (2px `--k-chart-mark` line, filled dots, ring = today). No real instance. Orchestrator: add to def § Canvas gaps.
- **A7 — 上架 = `first_seen`** (Kura's first observation). Every listing from the initial sweep reads 10/13 · 在架第 1 天 though shops listed many earlier — canvas copy kept; ties to def L4 「Day-1 appearance」 (open). Flag, don't fix.
- **A8 — Chart dots hover-only** (spans, not buttons): plot is `role="img"` with a summary label (children presentational), and a 400-dot group would otherwise add 400 tab stops; keyboard/AT path = summary + 以表格檢視 (HANDOFF §3.7). Canvas renders buttons; look unchanged.
- **A9 — Two medians, two labels:** prose 「中位數 NT$X」 = `fair.fair_twd` (comparables exclude this listing); chart tick/table 「中位數（在售）」 = `by_grade.median_twd` (in-sale incl. this listing). Both server-computed; may differ by design.
- **A10 — Payload:** detail fetches the full model group (≤ ~490 KB) — accepted locally; hosting/model iteration may slim `Model` items.
- **A11 — Missing grade** (1 listing, `brandoff:2600041253809`, empty code): 鑑定卡 omitted when `condition.label` empty; ladder omitted when code not in scheme; points with empty code excluded from distribution (sprint Deferred: "condition state for grade w/o published definition").

### Impact Analysis
- **Direct:** `/listing/{id}` view; listing API adds `scheme`; `history[]` JSON keys become `at`/`price_jpy`/`price_twd`; `by_grade[]` gains `*_twd`; every `condition` gains `short`.
- **Adjacent:** advisor tool JSON (same values, now from radar fields); legacy `historyChart` (key rename); card track listener (shared with detail button).
- **Downstream:** none outside the repo.

---

## Implementation

### Overview
Server data first (TWD fields, short labels, scheme in API; tested), then client plumbing (router params/404, util, card exports), CSS, detail view, distribution component, then the browser matrix incl. a scratch DB copy for the two states real data lacks.

### Step 1: Server-side TWD for grade stats + history; grade short label; advisor + legacy follow [PATCHED]

**Objective:** S10, S3 data. **Confidence:** High. **Depends on:** None.
**Do-Projex:** Encouraged

**Files:** `internal/listing/condition.go`, `internal/source/brandoff/condition.go`, `internal/radar/radar.go`, `internal/radar/radar_test.go`, `internal/advisor/tools.go`, `internal/web/static/legacy/app.js`.

**Changes:**

```go
// listing.Grade — add after Label:
	// Short is a compact label for tight grade ladders (e.g. "A" for ランクA);
	// empty means use Label. Never shown without the full label nearby.
	Short string `json:"short,omitempty"`

// brandoff.Scheme grades — add Short: 新品→"新品", 未使用品→"未使用", ランクA→"A", ランクB→"B", ランクC→"C"

// radar — new type above Item:
// PricePoint is one observed asking price, converted once with the index's FX
// rate so every consumer shows the same TWD figure.
type PricePoint struct {
	At       time.Time `json:"at"`
	PriceJPY int64     `json:"price_jpy"`
	PriceTWD int64     `json:"price_twd"`
}

// Item.History: []store.PricePoint → []PricePoint (tag unchanged: `json:"history,omitempty"`)
// toItem: replace `History: r.PriceHistory,` with a loop after the literal:
	for _, p := range r.PriceHistory {
		it.History = append(it.History, PricePoint{At: p.At, PriceJPY: p.PriceJPY, PriceTWD: e.FX.TWD(p.PriceJPY)})
	}

// GradeStat — add:
	MedianTWD int64 `json:"median_twd"`
	MinTWD    int64 `json:"min_twd"`
	MaxTWD    int64 `json:"max_twd"`
// stats(): build the JPY values first, then
	med, lo, hi := int64(median(ps)), int64(slices.Min(ps)), int64(slices.Max(ps))
	s.ByGrade = append(s.ByGrade, GradeStat{Source: k.source, Grade: grades[k], Count: len(ps),
		MedianJPY: med, MinJPY: lo, MaxJPY: hi,
		MedianTWD: e.FX.TWD(med), MinTWD: e.FX.TWD(lo), MaxTWD: e.FX.TWD(hi)})
```

`advisor/tools.go`: `modelStats` grade loop → `gradeTWD{g.Source, g.Grade.Label, g.Grade.DescZH, g.Count, g.MedianTWD, g.MinTWD, g.MaxTWD}`; `getListing` → `pt{p.At.Format(time.DateOnly), p.PriceTWD}`. (`Drop` still reads `r.PriceHistory` — unchanged.)

`legacy/app.js` `historyChart` + no-change line: `hist[0].At` → `hist[0].at`; object literal `{ At: …, PriceJPY: hist[hist.length - 1].PriceJPY, now: true }` → `{ at: …, price_jpy: hist[hist.length - 1].price_jpy, now: true }`; `new Date(p.At)` → `new Date(p.at)`; `p.PriceJPY` → `p.price_jpy`; `jpy(src.PriceJPY)` → `jpy(src.price_jpy)`.

`radar_test.go` `TestDropsAndStats` gains:
```go
	if g := s.ByGrade; g[0].MedianTWD != 20000 || g[0].MinTWD != 20000 || g[1].MaxTWD != 16000 {
		t.Errorf("ByGrade TWD = %+v", g)
	}
	it, _, _ := e.Get("t:a")
	if h := it.History; len(h) != 2 || h[0].PriceTWD != 24000 || h[1].PriceTWD != 20000 {
		t.Errorf("History = %+v", h)
	}
```

**Rationale:** radar already owns TWD for `price_twd`, `fair_twd`, `drop.*_twd`; extending it keeps UI and advisor on one computation (def: one source per stat). Snake-case keys over keeping `At`/`PriceJPY`: matches every other API field; legacy is the only reader.

**Verification:** `go vet ./...`; `go test ./internal/radar/ ./internal/advisor/ ./internal/listing/... ./internal/source/...`; `/legacy/` detail drawer 售價變化 renders (Step 7).

**If this fails:** revert the six files.

---

### Step 2: Listing API returns the grade scheme [PATCHED]

**Objective:** S3 ladder data, S1 404 path. **Confidence:** High. **Depends on:** Step 1.
**Do-Projex:** Encouraged

**Files:** `internal/web/server.go`, `internal/web/server_test.go`.

**Changes:**

```go
// Before (listing):
	return map[string]any{"item": it, "model": stats}, 0, nil
// After:
	res := map[string]any{"item": it, "model": stats}
	if sc, ok := listing.SchemeFor(it.Source); ok {
		res["scheme"] = sc // the shop's full grade ladder for the 鑑定卡
	}
	return res, 0, nil
```

`server_test.go` — `TestListingAPI`: register `listing.Scheme{Source: "wt", Name: "テスト", Shop: "テスト店", Graded: "店家鑑定", Grades: []listing.Grade{{Code: "A", Label: "ランクA", Short: "A", Rank: listing.RankA}}}`; `store.Open(filepath.Join(t.TempDir(), "kura.db"))` (Close via `t.Cleanup`); `RecordSeen(ctx, "HERMES", listing.Listing{Source: "wt", SourceID: "1", URL: "u", Brand: "HERMES", Title: "t", ModelCode: "c1", Grade: "A", Rank: listing.RankA, PriceJPY: 100000}, time.Now())`; `fxc := &fx.Converter{Fallback: 0.2}`; `h := (&Server{Store: st, Radar: &radar.Engine{Store: st, FX: fxc}, FX: fxc, Log: slog.New(slog.NewTextHandler(io.Discard, nil))}).Handler()`.
- `GET /api/listings/wt:1` → 200; decoded body: `scheme.shop == "テスト店"`, `scheme.grades[0].short == "A"`, `item.history[0].price_twd == 20000`, `model.by_grade[0].median_twd == 20000`.
- `GET /api/listings/wt:nope` → 404.

**Rationale:** scheme is per source and static; one field on the one response that needs it beats repeating the ladder on every `Item`.

**Verification:** `go test ./internal/web/`.

**If this fails:** revert both files.

---

### Step 3: Client plumbing — util, router, card exports [PATCHED]

**Objective:** S1, S2 track sync. **Confidence:** High. **Depends on:** None (Step 5 consumes).

**Files:** `internal/web/static/js/util.js`, `internal/web/static/js/app.js`, `internal/web/static/js/card.js`.

**Changes:**
- `util.js`:
  ```js
  export const jpy = (n) => "¥" + num(n);
  const MD = new Intl.DateTimeFormat("en-US", { timeZone: "Asia/Taipei", month: "numeric", day: "numeric" });
  // m/d in Taiwan time, e.g. 10/13.
  export const md = (iso) => MD.format(new Date(iso));
  // api(): on !r.ok → const err = new Error(body.error || r.statusText); err.status = r.status; throw err;
  ```
- `app.js`:
  - `import { renderDetail } from "./detail.js";` routes += `[/^\/listing\/([^/]+)$/, renderDetail, "商品頁 · Kura 蔵"]`.
  - Route contract comment updated: view `(el, ov, ...pathParams)` may return a document title.
  - `render()`: find route with `exec` (keep match); `let title = route ? route[2] : NOT_FOUND_TITLE` (`const NOT_FOUND_TITLE = "找不到這個頁面 · Kura 蔵"`); `title = (await route[1](out, ov, ...m.slice(1))) || title`; catch: `err.status === 404` → `renderNotFound(out); title = NOT_FOUND_TITLE;` else existing warn + error line; set `document.title = title` **inside** the `mine === seq` guard (a stale view must not retitle a newer page).
- `card.js`:
  - `export const isTracked = (id) => tracked.has(id);` `export const badgeHTML = (v) => BADGE[v.state](v);` (`cardHTML` uses it); export `PENDING_ICON`, `HEART`, `OUT_ICON`.
  - Listener update loop: `const text = b.querySelector("[data-track-text]"); if (text) text.textContent = on ? "追蹤中" : "追蹤這件"; else { aria-label + title as now }` — text-labelled buttons keep their visible text as the accessible name.

**Rationale:** path params + returned title keep views stateless and race-safe under the existing `seq` guard; 404 mapped centrally so later views (model, footer pages) reuse it. One tracked Set → detail and card toggles stay in step (canvas: same listing, same state).

**Verification:** Home still renders (`/`), card track toggles; `/listing/x` before Step 5 → module import error expected — run Step 3 + 5 together if checking in browser.

**If this fails:** revert the three files.

---

### Step 4: `kura.css` — detail, 鑑定卡, distribution, history [PATCHED]

**Objective:** S2–S9 look. **Confidence:** High. **Depends on:** None.

**Files:** `internal/web/static/kura.css`.

**Changes** (values from canvas `Main.dc.html` detail markup + `renderVals`: `crumbPad`, `detailTop`, `detailH1`, `heroIcon`, `detailGap`, `specCols`, `chartPad`):
1. Base: `[hidden] { display: none !important; }` (component rules set `display`, which would beat the attribute).
2. **Crumb** `.k-crumb` `padding:28px 0 10px; font-size:13px; color:ink-3; display:flex; flex-wrap:wrap; gap:6px`; links ink-3, no underline, `min-height:24px; display:inline-flex; align-items:center`; `.k-crumb__model` font-ja.
3. **Top** `.k-detail__top` `grid; columns minmax(0,1.1fr) minmax(0,1fr); gap:56px; align-items:start`. `.k-detail__mat` = card mat with `box-shadow: inset 0 0 0 10px var(--k-mat), inset 0 0 0 11px var(--k-mat-line)`, `padding:10%`, flex column centred gap 12px, color mat-ink; `img` contain + multiply; `svg` 200×200; `.k-detail__mat.is-sold { opacity:.6 }`.
4. **Info** `.k-detail__info` flex column gap 20px min-width 0. `.k-detail__id` column gap 6px; `.k-detail__brand` font-num 13px `.18em` ink-2 (inner `span` font-ui `.04em` ink-3); `h1` font-ja 600 32px/1.3 margin 0 `overflow-wrap:anywhere`; `.k-detail__sub` 13px ink-3. `.k-detail__pricing` column gap 8px; `.k-detail__price` baseline gap 12px wrap, `.k-price` 26px; `.k-detail__jpy` 13px ink-3 tabular; `.k-detail__val` center gap 10px wrap, `.k-badge` padding `2px 8px`; `.k-detail__fairlink` 13px ink-2 no underline, `min-height:24px; inline-flex`; `.k-detail__lastsold` 13px ink-3; `.k-detail__sold` font-hand 15px inline-flex gap 8px + `::before` 7px brass ring (as `.k-card__sold::before`).
5. **Actions** `.k-detail__actions` flex gap 10px wrap. `.k-cta` `flex:1 1 220px; inline-flex; center; gap:8px; height:48px; padding:0 20px; box-sizing:border-box; background:accent; color:on-accent; no underline; radius 2px; 15px`. `.k-dtrack` `height:48px; padding:0 18px; inline-flex; gap:8px; font:inherit; 15px; color:ink; background:transparent; border:1px solid ink-2; radius 2px; cursor:pointer`, `svg` 18×18; `[aria-pressed="true"]` → color accent-ink, bg accent-wash, border accent, `svg { fill: currentColor }`. `.k-detail__fine` `margin:-8px 0 0; 12px/1.7; ink-3`.
6. **鑑定卡** `.k-tag` relative `padding-top:18px; max-width:420px`; `.k-tag__card` relative surface, `1px solid rule-strong`, shadow, `padding:26px 22px 18px`, column gap 10px; `.k-tag__hole` abs `top:9px; left:50%; 10×10; margin-left:-5px; radius 50%; 1px solid brass; bg paper`; `.k-tag__head` space-between 11px `.18em` ink-3, last span font-num `.14em`; `.k-tag__label` font-ja 32px/1.1 600; `.k-tag__zh` 15px/1.6; `.k-tag__ja` font-ja 13px/1.6 ink-3. `.k-ladder` flex gap 4px margin-top 4px; `> span` flex 1, column gap 4px, font-ja 11px ink-3, `::before { content:""; height:3px; background:rule }`; `> .is-this` ink 600, `::before` ink. `.k-tag__foot` `border-top:1px dashed rule-strong; margin-top:6px; padding-top:12px; flex space-between center gap 12px`; `.k-tag__by` column gap 2px (`strong` 14px 700, `span` 12px ink-3); `.k-tag__stamp` 44×44 `flex:none; box-sizing:border-box; radius 50%; border:1.5px solid accent-ink; box-shadow: inset 0 0 0 2px var(--k-surface), inset 0 0 0 3px var(--k-accent-ink); color:accent-ink; flex center; font-ja 13px; line-height:1.05; writing-mode:vertical-rl; transform:rotate(-8deg); opacity:.9`.
7. **Lower column** `.k-detail__col` `max-width:880px; padding-top:64px`; `.k-dsec` `padding:0 0 48px`; `.k-dsec--last` `padding:0 0 8px`; `.k-dsec > h2` `margin:0 0 16px; padding-top:24px; border-top:1px solid brass; font-display 400 22px`. `.k-specs` dl `margin:0; grid; repeat(2,minmax(0,1fr)); column-gap:40px`; row `div` `grid; 96px minmax(0,1fr); gap:12px; padding:11px 0; border-bottom:1px solid rule; 14px/1.6`; `dt` ink-3; `dd` margin 0 `overflow-wrap:anywhere`. `.k-fair p` `margin:0 0 18px; 16px/1.85; max-width:40em`; `strong` font-num 18px 500; `.k-fair dl` `margin:0; flex wrap; gap:10px 32px; 13px`; rows flex gap 8px; `dt` ink-3; `dd` margin 0; link ink-2 `min-height:24px; inline-flex; align-items:center`.
8. **Distribution** `.k-dist__head` `flex wrap space-between baseline; gap:8px 16px; padding-top:24px; border-top:1px solid brass; margin-bottom:14px`, `h2` margin 0, font-display 400 22px. `.k-dist__toggle` `font:inherit; 13px; ink-2; transparent; 1px solid rule-strong; radius 2px; height:32px; padding:0 12px; cursor:pointer`. `.k-legend` `flex wrap; gap:6px 20px; 12px; ink-2; margin-bottom:16px`, items inline-flex gap 6px; `i` block border-box: `.sale` 10px circle chart-mark; `.sold` 10px circle `2px solid chart-mark`; `.this` 12px circle accent-ink; `.median` 18×2 ink. `.k-dist__plot` `relative; height:300px; margin-left:44px`; `.k-dist__grid` abs `left:0; right:0; height:1px; rule`; `.k-dist__y` abs `left:-44px; width:38px; translateY(-50%); right-aligned; 12px; ink-3; tabular`; `.k-dist__median` abs `width:34px; margin-left:-17px; height:2px; ink`; `.k-dot` abs `24×24; margin:-12px 0 0 -12px; flex center`, `::after { content:""; 10×10; radius 50%; box-sizing:border-box; background:chart-mark; box-shadow:0 0 0 2px var(--k-paper) }`; `.k-dot--sold::after` `border:2px solid chart-mark; background:paper`; `.k-dot--this { z-index:1 }` `::after` 14×14 accent-ink; `.k-dist__this` abs `translate(14px,-50%); 12px; 700; accent-ink; nowrap; pointer-events:none`; `.k-dist__tip` abs `translate(-50%, calc(-100% - 14px)); surface; 1px solid rule-strong; shadow; padding:6px 10px; 12px/1.5; ink; nowrap; pointer-events:none; z-index:2` (`.ja` font-ja; `.k-price` 15px 400). `.k-dist__axis` `grid; margin-left:44px; margin-top:10px; border-top:1px solid ink-3`; cells center `padding-top:8px; column gap 2px`; `b` font-ja 13px ink-2 400; `.is-this b` ink 700; `small` 11px ink-3. `.k-dist__table` `overflow-x:auto`; `table` `width:100%; collapse; 13px; min-width:560px`; `thead th` left 400 ink-3 `padding:8px 12px 8px 0; border-bottom:1px solid ink-3`; `tbody th` left 600 font-ja `padding:10px 12px 10px 0; border-bottom:1px solid rule; nowrap`; `td` `padding:10px 12px 10px 0; border-bottom:1px solid rule; tabular`; `.def` ink-2; `.num` right; last column `padding-right:0`.
9. **History** `.k-hist` `relative; height:56px; margin:0 8px`; `.k-hist__line` abs `left:0; right:0; top:27px; height:2px; chart-mark`; `.k-hist__start` abs `left:0; top:22px; 12×12; margin-left:-6px; radius 50%; chart-mark; box-shadow:0 0 0 2px paper`; `.k-hist__now` abs `right:0; top:22px; 12×12; margin-right:-6px; radius 50%; border:2px solid chart-mark; paper; border-box`; `.k-hist__label` abs `top:0; 12px; ink-3` (`--end` → `right:0`); `.k-hist__price` abs `left:50%; top:36px; translateX(-50%); font-num 15px; tabular`; `.k-hist__note` `margin:18px 0 0; 14px; ink-2`. Steps variant `.k-hist--steps { height:140px }`: `.k-hist__plot` abs `left:0; right:0; top:22px; bottom:26px`, `svg` 100%×100% `overflow:visible`; `.k-hist__pt` abs `12×12; margin:-6px 0 0 -6px; radius 50%; chart-mark; box-shadow:0 0 0 2px paper`; `.k-hist__pt--now` ring as `.k-hist__now`; `.k-hist__ptlabel` abs `translate(-50%, calc(-100% - 10px)); font-num 13px; tabular; nowrap`; labels at `bottom:0` instead of `top:0`.
10. **Mobile** (inside existing `@media (max-width: 767px)`): `.k-crumb { padding:16px 0 8px }`; `.k-detail__top { display:flex; flex-direction:column; gap:24px }`; `.k-detail__info h1 { font-size:26px }`; `.k-detail__mat svg { width:120px; height:120px }`; `.k-detail__col { padding-top:40px }`; `.k-specs { grid-template-columns:1fr }`.

**Rationale:** transcribed canvas values; class names consumed by Steps 5–6.

**Verification:** visual, Step 7.

**If this fails:** remove the added blocks.

---

### Step 5: Detail view [PATCHED]

**Objective:** S1–S5, S7, S8. **Confidence:** High (S7 multi-point: Medium). **Depends on:** Steps 2–4.
**Verify-Projex:** Encouraged

**Files:** `internal/web/static/js/detail.js` (new).

**Changes:** `export async function renderDetail(main, ov, rawId)`:
1. `let id; try { id = decodeURIComponent(rawId) } catch { throw Object.assign(new Error("bad listing id"), { status: 404 }) }`.
2. `const { item: it, model, scheme } = await api(\`/api/listings/${encodeURIComponent(id)}\`)`; if `model && it.group_key` → `group = await api(\`/api/models/${encodeURIComponent(it.brand)}/${encodeURIComponent(it.group_key)}\`).catch(() => null)` (distribution is secondary; its failure must not blank the page).
3. `main.innerHTML` = crumb + `.k-detail__top` + `.k-detail__col` (商品資料, 行情估價 unless sold, `distHTML(...)` from Step 6, 售價變化). Return `` `${name} · ${it.brand} · Kura 蔵` ``. All interpolations through `esc`. `name = it.model_name || it.title`; `sold = it.status !== "active"`; `v = valuation(it)`.

**Markup (canvas order):**
- Crumb `<nav aria-label="位置" class="k-crumb">`: `首頁` → `/`; `/`; if `group_key`: `<a class="k-link k-crumb__model" href="/model/{enc brand}/{enc group_key}">{brand} {name}</a>` (A1) else plain span; tail span `[model_code, condition.label].filter(Boolean).join(" · ")` when non-empty. Separators `<span aria-hidden="true">/</span>`.
- Photo column: `<div class="k-detail__mat{sold ? " is-sold" : ""}">` → `photo_pending` ? `PENDING_ICON` + `<span class="k-card__pending">店家照片準備中</span>` : `<img src alt="{name} 商品照片（店家白底照）">`.
- Info column `.k-detail__info`:
  - `.k-detail__id`: `<span class="k-detail__brand">{brand} <span>{brand_zh}</span></span>`, `<h1>{name}</h1>`, `<span class="k-detail__sub">{[category_zh, model_code].filter(Boolean).join(" · ")}</span>`.
  - `.k-detail__pricing`: price row (`sold` → `<span class="k-detail__lastsold">最後售價</span>` first) `<span class="k-price">{twd(price_twd)}</span><span class="k-detail__jpy">{jpy(price_jpy)}</span>`; val row: sold → `<span class="k-detail__sold">已找到新主人</span>`; else `badgeHTML(v)` + (fair ? `<a href="#fair" class="k-detail__fairlink k-link">行情 {twd(fair_twd)} · {CONF[confidence]}</a>` : ""), `CONF = {high:"高信心", medium:"中信心", low:"低信心"}`.
  - `.k-detail__actions`: `<a class="k-cta" href="{url}" target="_blank" rel="noopener noreferrer">到 {source_name} 商品頁<span class="k-sr">（另開視窗）</span>{OUT_ICON}</a>` + (not sold) `<button type="button" class="k-dtrack" data-track="{id}" aria-pressed="{isTracked}">{HEART}<span data-track-text>{追蹤中|追蹤這件}</span></button>`.
  - `<p class="k-detail__fine">` not sold: `追蹤後，降價或售出時會通知你。<br>Kura 不販售商品。購買、付款與寄送都在 {source_name} 完成。`; sold: second sentence only.
  - 鑑定卡 when `condition.label` (A11): `.k-tag > .k-tag__card` → `.k-tag__hole` (aria-hidden), `.k-tag__head` `<span>鑑定卡</span><span>{source_name}</span>`, `.k-tag__label`, `.k-tag__zh` (if desc_zh), `.k-tag__ja` 「{desc_ja}」 (if desc_ja), ladder if `scheme.grades` contains `condition.code`: `<div class="k-ladder" role="img" aria-label="店家等級：{labels joined 、, this one + （此件）}">` with `<span class="is-this?">{short || label}</span>` per grade; foot `.k-tag__by` `<strong>{condition_by}</strong><span>{condition_scheme}</span>` + stamp `<span class="k-tag__stamp" aria-hidden="true">店鑑</span>` only when `condition_by === "店家鑑定"`.
- `.k-detail__col`:
  - 商品資料 `<section class="k-dsec" aria-labelledby="spec-h"><h2 id="spec-h">商品資料</h2><dl class="k-specs">` rows `[["型號編號", model_code], ["類型", category_zh], ["顏色", color], ["材質", material], ["尺寸", size], ["付屬品", accessories], ["店家備註", notes], ["上架", sold ? ymd(first_seen) : \`${ymd(first_seen)} · 在架第 ${days_listed + 1} 天\`]]` filtered on truthy value → `<div><dt>k</dt><dd>v</dd></div>`.
  - 行情估價 (not sold) `<section id="fair" class="k-dsec k-fair" aria-labelledby="fair-h">` → prose `<p>` + dl per table below.
  - `distHTML({ stats: group.model, items: group.items, scheme, thisId: it.id })` when `group`.
  - 售價變化 `<section class="k-dsec k-dsec--last" aria-labelledby="hist-h"><h2 id="hist-h">售價變化</h2>{historyHTML(it)}</section>`.

**行情估價 prose** (`f = it.fair`, `g = condition.label`, `v` = card `valuation(it)`):

| Part | Condition | Copy |
|---|---|---|
| none | `!f` | `<p>同型號的比較樣本不足 3 件，暫不估計行情。</p>` (HANDOFF §5); dl = 行情怎麼算 link only |
| sentence 1 | `method === "same-grade"` | `同一家店、同樣是 {g} 的 {name}，目前有 {comps} 件可以比較，中位數是 <strong>{twd(fair_twd)}</strong>。` (canvas) |
| sentence 1 | `rank-adjusted` | `同一家店、同樣是 {g} 的 {name} 不到 3 件，所以把同型號各等級的 {comps} 件換算到 {g} 來比較，估計行情是 <strong>{twd(fair_twd)}</strong>。` (A4) |
| sentence 2 | `v.state === "anomaly"` | `價格比同型號同成色的行情低一半以上，通常代表有標題沒寫的瑕疵、缺件或版本差異，購買前請務必確認商品頁的照片與備註。` (HANDOFF §5) — no saving |
| sentence 2 | `below` \| `low` | `這件比它低 {v.pct}%，約省 {twd(fair_twd - price_twd)}。` (canvas) |
| sentence 2 | `above` | `這件比它高 {v.pct}%。` (legacy) |
| sentence 2 | `near` | `這件的售價接近行情。` (A4) |
| sentence 3 | `confidence === "low"` | `dispersion > 0.4` → `同型號的價格非常分散（可能混有不同皮革、尺寸或限定版），這個估計僅供參考。` (HANDOFF §5); else `這個估計僅供參考。` (A4) |

dl: `方法` → `{ "same-grade": "同店同等級中位數", "rank-adjusted": "跨等級換算" }`; `樣本` → `{comps} 件`; `信心` → `{ high: "高 · 同等級價格集中", medium: "中", low: "低" }`; `<a href="/methodology" class="k-link">行情怎麼算</a>`.

**`historyHTML(it)`** (`h = it.history ?? []`):
- `h.length < 2` (canvas): `<div class="k-hist"><span class="k-hist__line" aria-hidden="true"></span><span class="k-hist__start" aria-hidden="true"></span><span class="k-hist__now" aria-hidden="true"></span><span class="k-hist__label">{md(h[0]?.at ?? first_seen)} 上架</span><span class="k-hist__label k-hist__label--end">{sold ? md(gone_at) : "今天"}</span><span class="k-hist__price">{twd(h[0]?.price_twd ?? price_twd)}</span></div><p class="k-hist__note">上架以來還沒有調價。店家調價時，會畫在這條線上。</p>`.
- `≥ 2` (A6): `t0 = Date(h[0].at)`, `t1 = sold ? Date(gone_at) : Date.now()` (`t1 <= t0` → `t0 + 1`); `x(t) = (t - t0) / (t1 - t0) * 100`; price range `lo..hi` of `h[].price_twd` padded 20% (`|| hi * 0.05`), `y(v) = (hi - v) / (hi - lo) * 100`. `.k-hist.k-hist--steps` → `.k-hist__plot` holding `<svg viewBox="0 0 100 100" preserveAspectRatio="none" aria-hidden="true"><path d="M{x0} {y0} H{x1} V{y1} … H100" fill="none" stroke="var(--k-chart-mark)" stroke-width="2" vector-effect="non-scaling-stroke"/></svg>`, one `.k-hist__pt` per point (`left:{x}%; top:{y}%`, `title="{ymd(at)} {twd}"`) + `.k-hist__ptlabel` (twd) at the same spot, `.k-hist__pt--now` at `left:100%; top:{y(last)}%`; labels `{md(h[0].at)} 上架` / 今天; `<ul class="k-sr">` one `<li>{ymd(at)} {twd(price_twd)}</li>` per point (chart text alternative). No note paragraph.

Import: `api, esc, twd, jpy, ymd, md` (util), `valuation, badgeHTML, isTracked, PENDING_ICON, HEART, OUT_ICON` (card), `distHTML` (dist).

**Rationale:** one view module per route (home pattern); prose built only from `fair` fields + server TWD (S10); sold/multi-point built from existing designed vocabulary (A5/A6) instead of new visuals.

**Verification:** Step 7 matrix rows D1–D12.

**If this fails:** delete `detail.js`, drop the route line in `app.js`.

---

### Step 6: Distribution component [PATCHED]

**Objective:** S6. **Confidence:** High. **Depends on:** Steps 2, 4.
**Verify-Projex:** Encouraged

**Files:** `internal/web/static/js/dist.js` (new).

**Changes:** `export function distHTML({ stats, items, scheme, thisId })` → section HTML or `""`. Reusable by the model page (`thisId` null → no 此件 legend/label, no column highlight).

```js
const cols = (scheme?.grades ?? []).map((grade) => ({ grade }));
for (const it of items) {
  const c = it.condition;
  if (c?.code && !cols.some((x) => x.grade.code === c.code)) cols.push({ grade: c }); // grades outside the published ladder, appended
}
for (const col of cols) {
  col.stat = stats.by_grade?.find((s) => s.grade.code === col.grade.code && (!scheme || s.source === scheme.source));
  col.pts = items.filter((it) => it.condition?.code === col.grade.code).sort((a, b) => a.price_twd - b.price_twd);
}
const vals = cols.flatMap((c) => [...c.pts.map((p) => p.price_twd), ...(c.stat ? [c.stat.median_twd] : [])]);
if (!vals.length) return "";
let lo = Math.min(...vals), hi = Math.max(...vals);
const pad = (hi - lo) * 0.08 || hi * 0.1;
lo = Math.max(0, lo - pad); hi += pad;
const top = (v) => (((hi - v) / (hi - lo)) * 100).toFixed(2) + "%";
const OFFS = [0, -13, 13, -26, 26]; // canvas jitter within a grade column, px
```
Copy `niceTicks(lo, hi, 4)` and `shortTWD` (`≥ 10000` → `(v/10000).toFixed(v % 10000 ? 1 : 0) + " 萬"`, else `num(v)`) from `legacy/app.js`; keep ticks with `lo < t < hi`.

Markup:
```html
<section class="k-dsec k-dist" aria-labelledby="dist-h">
  <div class="k-dist__head"><h2 id="dist-h">同型號價格分布</h2><button type="button" class="k-dist__toggle" data-dist-toggle>以表格檢視</button></div>
  <div class="k-legend" aria-hidden="true"><span><i class="sale"></i>在售</span><span><i class="sold"></i>已售出</span>{thisId: <span><i class="this"></i>此件</span>}<span><i class="median"></i>中位數（在售）</span></div>
  <div class="k-dist__chart">
    <div class="k-dist__plot" role="img" aria-label="{summary}">
      {per tick: <div class="k-dist__grid" style="top:{top(t)}"></div><span class="k-dist__y" style="top:{top(t)}">{shortTWD(t)}</span>}
      {per col with stat: <div class="k-dist__median" style="left:{cx}%;top:{top(stat.median_twd)}"></div>}
      {per col, per pt i: <span class="k-dot k-dot--{this|sale|sold}" style="left:calc({cx}% + {OFFS[i % 5]}px);top:{top(price_twd)}" data-grade="{label}" data-status="{此件 · 在售|此件 · 已售出|在售|已售出}" data-price="{price_twd}"></span>}
      {this pt: <span class="k-dist__this" style="left:…;top:…">此件</span>}
      <div class="k-dist__tip" role="status" hidden></div>
    </div>
    <div class="k-dist__axis" aria-hidden="true" style="grid-template-columns:repeat({cols.length},minmax(0,1fr))">
      {per col: <div class="{is-this?}"><b>{label}</b><small>{stat ? `在售 ${stat.count}` : "—"}</small></div>}
    </div>
  </div>
  <div class="k-dist__table" hidden><table>…等級 / 店家定義 / 在售 / 中位數 / 最低 / 最高; per col: th label, td.def desc_zh, td.num count|—, twd(median_twd)|—, twd(min_twd)|—, twd(max_twd)|—</table></div>
</section>
```
`cx = (index + 0.5) / cols.length * 100`. Summary: `` `${stats.model_name} 各店家等級的價格分布，在售 ${stats.active} 件` `` + (this col stat ? `；{label} 中位數 {twd(median_twd)}` : "") + (this pt ? `，此件 {twd(price_twd)}` : "") + `。可切換為表格檢視。` — no sold count (S10).

Interactions — document-level listeners installed once at module load (same reason as `card.js`: re-rendered views would stack listeners):
- `click` on `[data-dist-toggle]` → within `closest(".k-dist")` flip `hidden` on `.k-dist__chart` / `.k-dist__table`, button text 以表格檢視 ↔ 以圖表檢視, hide tip.
- `mouseover` on `.k-dot` → plot's `.k-dist__tip` innerHTML `<span class="ja">{esc grade}</span> · {esc status}<br><span class="k-price">{twd(price)}</span>`, `style.left/top` = dot's, `hidden = false`; `mouseout` on `.k-dot` → `hidden = true`.

**Rationale:** canvas `kuraChart` geometry with real points; medians/table from server `by_grade` (one computation, S10) rather than recomputed in JS as the canvas sample does; dots hover-only per A8.

**Verification:** Step 7 D1, D4, D6, D10; table numbers = `/api/models/…` `by_grade[].*_twd`.

**If this fails:** delete `dist.js`; detail renders without the section (remove the `distHTML` call).

---

### Step 7: End-to-end browser verification [PATCHED]

**Objective:** S1–S11 evidence. **Confidence:** Medium (hotlinked photos/fonts depend on network). **Depends on:** Steps 1–6.
**Verify-Projex:** Encouraged

**Procedure:**
1. `go vet ./...` + `go test ./...`; hygiene grep `grep -rnE "[0-9]{10}-[a-z0-9-]+-(plan|sprint|def|nav|audit|patch)" internal cmd` → none.
2. Background `go run ./cmd/kura serve -db kura.db -addr 127.0.0.1:8090` (worktree root; no crawl flags).
3. Scratch copy for A5/A6: `sqlite3 -readonly kura.db ".backup '<scratchpad>/mut.db'"`; in `mut.db`: `UPDATE listings SET status='gone', gone_at=<first_seen + 2 days> WHERE id='brandoff:2101220298944'`; for `brandoff:2101220216016` insert a `price_points` row at `<first observed_at + 86400>` with `price_jpy` 10% lower and set `listings.price_jpy` to it (check column types in `internal/store/schema.sql` first). Serve on `127.0.0.1:8092`.
4. Matrix (browser pane; ids from plan-time probe — re-pick via `/api/listings` if absent):

| # | Case | Id / how | Check |
|---|---|---|---|
| D1 | same-grade, high | `brandoff:2101220216016` | full page; prose canvas copy; tick/table = `by_grade` TWD |
| D2 | same-grade medium / low | `2101220298944` / `2101220156220` | 信心 中 / 低 + low line |
| D3 | rank-adjusted medium / low | `2101219298115` / `2101219445359` | A4 sentence, 跨等級換算 |
| D4 | anomaly | `2101218625240` | anomaly badge + line, no 約省 |
| D5 | no estimate | `2101219439341` | none copy; chart still if group |
| D6 | biggest group | any id from `/api/models/CHANEL/name%3Aマトラッセ` | 409 dots render, tooltip, page responsive |
| D7 | photo pending | `2104102531481` | pending mat |
| D8 | notes/specs | `2100301207912` | rows present, original notes |
| D9 | empty grade | `2600041253809` | no 鑑定卡, crumb tail = code only, no crash |
| D10 | Home → card click | `/` first card | in-app nav, title, back button returns to Home |
| D11 | unknown / malformed | `/listing/brandoff%3Anope`, `/listing/%E0%A4%A` | not-found view + title |
| D12 | sold, multi-point | :8092 `2101220298944`, `2101220216016` | S8 state; step line + sr list |

   D1 at 1280 + 375 × `?theme=light|dark` (and once via colorScheme emulation); D4, D6, D12 at 1280 light + 375 dark; rest 1280 light. Track toggle on D1 → back to Home → same card shows tracked (in-memory, same session). `#fair` link scrolls without re-render. `/legacy/` detail drawer 售價變化 still renders (Step 1 key rename).
5. Per view: screenshot vs canvas `Detail` / `DetailMobileDark`; `read_console_messages` onlyErrors empty; at 375 `document.documentElement.scrollWidth <= innerWidth`; Tab through crumb → CTA → track → toggle: focus ring visible; page text has no 售出率 / 天找到新主人 / 已售出 N 件 strings (S10).
6. Stop servers; delete `<scratchpad>/mut.db*`; worktree `git status` clean except committed work.

---

## Verification Plan

### Automated Checks
- [x] `go vet ./...` clean
- [x] `go test ./...` green incl. `TestListingAPI`, extended `TestDropsAndStats`

### Manual Verification
- [x] Step 7 matrix D1–D12
- [x] Home unchanged (cards, shelves, footer FX)
- [x] Source hygiene grep → none

### Acceptance Criteria Validation

| Criterion | How to Verify | Expected Result |
|---|---|---|
| S1 | D10, D11; `document.title` | detail title; not-found view for bad ids |
| S2 | D1 screenshots vs canvas; track sync | canvas layout/values; toggle state shared with card |
| S3 | D1 鑑定卡; ladder aria-label; D9 | ladder 新品 未使用 A B C, this bold; parts omitted when absent |
| S4 | D1, D8 | only rows with data, canvas labels/order |
| S5 | D1–D5 prose vs table in Step 5 | exact copy per state; numbers = `fair` fields |
| S6 | D1, D6 dots/tick/table vs `/api/models/…` | counts, medians, min/max equal server `*_twd` |
| S7 | D1 strip; D12 step line | canvas no-change copy; steps at observed prices |
| S8 | D12 sold | 最後售價 + 已找到新主人; no badge/track/行情估價 |
| S9 | 375 scrollWidth; focus; both themes | no overflow; ring visible; tokens switch |
| S10 | advisor `go test`; code review of `tools.go`; page text grep | radar fields used; no real-derived values rendered |
| S11 | vet/test/grep | green / none |

**Data-class trace (detail):** price / ¥ / photo / grade + definitions / specs / notes / 上架 date ← listing fields (real) | badge, prose, 行情 link ← `fair` (real — radar valuation of current listings) | ladder ← `scheme` (real — shop's published scheme) | dots ← model group items' `price_twd`/`status` (real) | medians, table ← `by_grade` (real — current in-sale listings) | 在售 N ← `stats.active` (real) | 售價變化 ← `history[].price_twd` (real). Real-derived (sell-through, sold count, days-to-sell): none shown. Mocked: none (track toggle = in-memory UI state, no number).

---

## Rollback Plan

1. `git -C <work-root> log --oneline 5ad2212..HEAD` → this plan's commits.
2. `git -C <work-root> revert <sha>…` newest first (sprint branch only); `/listing/{id}` returns to the not-found view.

---

## Notes

### Split Decision
No split — single scope (`internal/` code, one `.projex` folder), within size budget (~475 lines / ~47 KB; 7 steps — heuristic needs both). Steps tightly coupled: data fields → API → client plumbing → CSS → view → component → verification.

### Risks
- 409-dot columns overlap heavily (canvas jitter only ±26px): legible as density, exact values via table — accept; model iteration may revisit.
- Model group payload ~490 KB (A10) — slow on poor networks; local mock fine.
- `history` JSON key rename breaks any external consumer — none known (legacy updated in Step 1).
- BRAND OFF hotlink / fonts unreachable → same exposure as Home.

### Open Questions
None blocking execution. For the orchestrator to log (sprint nav / def), not decided here:
- A4 plan-authored copy → canvas pass 2 copy review.
- A5 detail sold state, A6 multi-point 售價變化 → def § Canvas gaps.
- A7 上架 = first observation (initial-sweep listings all read 在架第 1 天) → def L4 「Day-1 appearance」.
- A1 model route URL → model iteration.
