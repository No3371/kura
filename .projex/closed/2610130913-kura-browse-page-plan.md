# Kura browse page (價格最美)

> **Sprint:** 2610130524-kura-interactive-mock-sprint.md — iteration 5
> **Status:** Complete
> **Author:** Claude Opus 5.5 (planner, sprint subagent)
> **Parent:** 2610130524-kura-interactive-mock-sprint.md
> **Source:** Sprint derive, iteration 5 (sprint nav § Current Position + Revision Log, commit `aa020a0`)
> **Related Projex:** 2610130928-kura-browse-page-patch.md (execution record) | 2610130952-kura-browse-audit-fixes-patch.md (audit F1–F5 fixes) | 2610130524-kura-interactive-mock-sprint.md | 2610130449-kura-mock-def.md (data classes, canvas precedence, Canvas deviations — grade filter via internal rank, Canvas gaps — dark × browse) | 2610130420-kura-roadmap-nav.md (Phase 3 New IA) | 2610130602-kura-canvas-foundation-home-audit.md (F1 / F6 open; F3 link contract `/browse?deals=1&sort=deal`) | 2610130535-kura-canvas-foundation-home-plan.md (shell, card incl. row modifier, empty state) | 2610130722-kura-model-page-plan.md (A9 看全部 in place — revisited here, A9 below)
> **Worktree:** Yes — existing sprint worktree `kura/.projexwt/2610130524-kura-interactive-mock`, branch `projex/sprint/2610130524-kura-interactive-mock`; execute via patch-projex plan-execution mode, no new branch

> Partial Execution: Steps 1–5 (all objectives) completed via 2610130928-kura-browse-page-patch.md — commits `2f15a43`, `265eea0`. Nothing remains open in this plan; small deviations listed in the patch's Notes.

---

## Summary

Build the canvas Browse page (pinned 1791839497-46ec, artboards `Browse` desktop light 1280 + `BrowseMobile` light 375; markup `project/Main.dc.html` `browseData` block) at `/browse` on real BRAND OFF data: crumb, h1 + count, 篩選與排序 form (desktop row; mobile keyword + 篩選與排序 + switch), card grid (desktop stack cards, mobile row cards), 「再看 N 件」 load-more. State lives in the URL query, so Home's two 看全部 links and the header search land on the right list. Closes 3 shell dead ends (Home 看全部 ×2, header search form + mobile icon) and the dark × browse combo.

**Scope:** `internal/radar` (Search `Offset`, Overview `ByCategory` + tests) and `internal/web` (`/api/listings` `offset` + test; router route + in-app header search; new `browse.js`; `kura.css` browse block; `home.js` empty-state export; `model.js` comment). One `.projex` scope (repo root).
**Estimated Changes:** 9 files — `radar.go`, `radar_test.go`, `server.go`, `server_test.go`, `app.js`, `home.js`, `model.js`, `kura.css` edited; `browse.js` added.

---

## Objective

### Problem / Gap / Need
Home 「看全部最新上架」 (`/browse?sort=newest`), 「看全部價格最美 N 件」 (`/browse?deals=1&sort=deal`), header search (`/browse?q=…`) and the mobile search icon (`/browse`) all render the not-found view: no `/browse` route. Canvas Browse is the last canvas flow not held by an open question whose answer changes the page (sprint nav blocking map). `/api/listings` has no offset, so load-more cannot page.

### Success Criteria
- [x] **S1 Route:** `/browse` with any query renders the browse view by in-app click (both Home links, mobile search icon), header search submit (in-app, no page reload) and direct load/reload; `/browse/x` → not-found. Title `{heading} · Kura 蔵`.
- [x] **S2 Head:** crumb `首頁 / {heading}` (current page `aria-current`); h1 = heading per A2; count `{num(total)} 件` = `/api/listings` `total` for the page's query. `/browse?deals=1&sort=deal`: count = Home 「看全部價格最美 N 件」 N (1,471 on the worktree DB) and first 3 cards = Home 價格最美 shelf ids, same order.
- [x] **S3 Desktop filters** (canvas order, labels, widths, placeholders): 品牌 (全部品牌 + brands with active listings, count desc, `{name}（{count}）` — A3) | 類型 (全部類型 + categories with active listings, count desc — A4) | 最低成色 (任何成色 / 未使用以上 / 輕微使用感以上 / 明顯使用感以上 → `min_rank` "" / S / A / B — A5) | 預算上限（NT$） placeholder 不限 → `max_twd` | 關鍵字 placeholder 「型號、編號，可用日文：ネヴァーフル」 → `q` | 排序 (低於行情優先 / 最新上架 / 價格 ↑ / 價格 ↓ → `deal` / `newest` / `price_asc` / `price_desc`) | 只看低於行情 `role="switch"` ↔ `deals=1`. Each committed change (select/switch change; text field change or Enter) rewrites the URL (`replaceState`), heading, title, count and results; reload restores the identical state and controls (A1).
- [x] **S4 Mobile (≤ 767):** canvas form — 關鍵字 (placeholder 「型號、編號，可用日文」, 44px) then a row of 「篩選與排序」 button + switch; the button (`aria-expanded`, `aria-controls`) reveals brand / 類型 / 最低成色 / 預算 / 排序 in place (A7); same state semantics as desktop; both forms always show the same state.
- [x] **S5 Results:** desktop grid `auto-fill minmax(240px, 1fr)`, gap 48 / 28; mobile list of row-layout cards, each cell `padding 20px 0` + bottom rule. First page 40 cards; 「再看 {num(total − shown)} 件」 appends the next 40 (no duplicate ids), moves focus to the first new card's model link, disappears when all are shown, absent when total ≤ 40. Zero matches → 「沒有符合條件的包。」 (A8). No active listings site-wide → canvas 「還在準備中。」 state.
- [x] **S6 Server:** `GET /api/listings?offset=n` skips n sorted matches (`total` unchanged; offset ≥ total → `items: []`); `/api/overview` `overview.by_category` = active count per category key. Advisor search unchanged.
- [x] **S7 Responsive/themes:** 1280 + 375 × light/dark (dark = token-driven, no artboard); no horizontal page scroll at 375; targets ≥ 24px; visible focus ring on every control; 0 console errors.
- [x] **S8 One source per stat / open questions untouched:** count = API `total`; brand/類型 counts from overview; page computes no figure; deal predicate (`deal_pct ≥ 0.05`, not low, not anomaly), `anomalyPct`, `dealScore` unchanged (F1, F6 stay open); no mocked value rendered.
- [x] **S9** `go vet ./...` + `go test ./...` green (new tests Step 1); hygiene grep clean.

### Out of Scope
- F1 threshold (5 % vs 10 %), F6 anomaly-boundary ordering — Browse consumes radar `OnlyDeals` + `deal` sort as-is
- Model-group filter on Browse; model page 看全部 → link (A9)
- Filter-sheet / dialog design for mobile, designed zero-result state, search-results heading → canvas pass 2 (logged, not designed here)
- Member header, notifications, persisted tracking (member-entry open question); footer pages incl. 行情怎麼算
- Router-wide focus after route change (iteration-4 carry-over, shell a11y)
- `min_price` (canvas has budget max only); sold/gone listings on Browse (`gone=1` not exposed)

---

## Context

### Current State
- **Server:** `GET /api/listings` (`server.go` `listings`) → `radar.Filter{Brand, Category, ModelCode, MinRank, MinPriceTWD, MaxPriceTWD, OnlyDeals, IncludeGone, Sort, Limit}` from `brand category model_code min_rank min_twd max_twd deals gone sort limit` + `q` → `Keywords` (whitespace split, any-match over title/model/code/color/material) → `{total, items}`. `Search` (`radar.go`): filters active items, `sortItems` (deal default | price_asc | price_desc | newest; stable over index order), `total = len(out)`, limit ≤ 0 or > 100 → 30. **No offset.** `OnlyDeals` = `Fair != nil && DealPct ≥ 0.05 && Confidence != "low" && !Anomaly` — the predicate Home's shelf/count and the advisor's `only_deals` use. Unknown `brand` → 0 matches; `min_rank` floors via `listing.Rank(...).Score()` (unranked listings excluded once a floor is set).
- **Overview** (`radar.Overview`): `active sold models valued deals drops by_brand since built_at sold_stats_ready`; `by_brand` counts active listings. No per-category counts. Handler adds `brands[{name,zh}]` (13, catalog order) and `categories[{key,zh}]` (15, catalog order).
- **Client:** `app.js` routes `/`, `/listing/{id}`, `/model/{brand}/{key}`; document click handler pushes in-app paths; `popstate` re-renders when path+query differ from the last render. `index.html` header `<form class="k-search" role="search" action="/browse" method="get">` (native GET → full reload, dropping the in-memory track set and advisor session) + mobile `<a class="k-search-icon" href="/browse">`. `home.js` links `/browse?sort=newest`, `/browse?deals=1&sort=deal`; renders the canvas empty state inline when `ov.overview.active === 0`. `card.js` `cardHTML(it, {layout})` — `"row"` adds `.k-card--row` (3 CSS rules); track listener document-level. `model.js` 看全部 reveals hidden cards in place; its comment says Browse "does not exist yet". `util.js` `api` throws with `.status`.
- **Data** (worktree `kura.db`, plan-time probe 2026-10-13 via local serve, read-only): 15,367 active, 0 sold. `deals=1` → 1,471 (overview `deals` ≥10 % → 1,020). Brands (= canvas `KURA_BRANDS` exactly): LOUIS VUITTON 6,030 · CHANEL 3,979 · HERMES 2,931 · CELINE 1,432 · LOEWE 897 · GOYARD 98. Categories active: shoulder 6,861 · handbag 4,741 · tote 2,073 · backpack 524 · bag 361 · boston 297 · belt 286 · clutch 126 · business 61 · travel 37 (2way, crossbody, pouch, vanity, basket: 0). Grades: ランクA 9,321 · ランクB 5,170 · 未使用品 596 · 新品 235 · ランクC 44 · unranked 1. Totals: `min_rank=S` 831 · `A` 10,152 · `B` 15,322; `brand=GOYARD` 98, `+deals=1` 0; `q=ネヴァーフル` 290, `+deals=1` 63; `max_twd=20000` 2,082; `category=travel` 37; `q=zzzzqqq` 0.
- **Canvas Browse** (`Main.dc.html` `browseData`; `renderVals`): crumb `首頁 / 價格最美` (padding `crumbPad` 28px 0 10px / mobile 16px 0 8px; 13px ink-3) → head `flex wrap baseline; gap 6px 14px; padding 0 0 24px`: h1 display 400 30px / mobile 24px, lh 1.3 + count 14px ink-3 tabular `1,020 件` → desktop `<form aria-label="篩選與排序">` `flex wrap; align-items flex-end; gap 12px; padding 0 0 32px`; labels `column; gap 6px; 12px ink-3`, widths 150 / 120 / 160 / 130 / `flex 1 1 180px` / 150; fields `.k-field` (Tokens: `14px ink; surface bg; 1px rule-strong; radius 2px; height 38px; padding 0 10px; width 100%`, placeholder ink-3); switch label `row; center; gap 8px; height 38px; 14px ink; nowrap`, checkbox `role="switch"` checked, 18×18, `accent-color: accent` → mobile form `column; gap 10px; padding 0 0 12px`: 關鍵字 (field height 44px) then row `gap 10px; center`: button `flex 1 1 0; height 44px; 14px ink; surface bg; 1px rule-strong; radius 2px; center; gap 8px` + filter icon `M2 4h12M4.5 8h7M7 12h2` 16px + 「篩選與排序」, switch label height 44px → grid `gridStyle` (desktop `grid; repeat(auto-fill, minmax(240px, 1fr)); gap 48px 28px`; mobile `flex column; gap 0`, cell `padding 20px 0; border-bottom 1px rule`, cards `layout="row"`) → load-more row `flex center; padding-top 36px`, button `height 44px; padding 0 24px; 14px ink; transparent; 1px ink-2; radius 2px` 「再看 1,015 件」. No zero-result, mobile filter panel, or non-deal heading on canvas. Site-empty (`data: empty`) replaces every page body with 「還在準備中。」. Header search form + mobile icon → `Browse.dc.html`.

### Key Files

| File | Role | Change Summary |
|------|------|----------------|
| `internal/radar/radar.go` | index | `Filter.Offset`; `Overview.ByCategory` |
| `internal/radar/radar_test.go` | tests | offset paging; by-category count |
| `internal/web/server.go` | API | `offset` param |
| `internal/web/server_test.go` | tests | `/api/listings` offset |
| `internal/web/static/js/app.js` | router | `/browse` route; `go()`; in-app header search |
| `internal/web/static/js/home.js` | Home | export empty-state markup |
| `internal/web/static/js/model.js` | model page | 看全部 comment (present tense) |
| `internal/web/static/kura.css` | styles | `.k-field`, browse head/form/grid/load-more, mobile |
| `internal/web/static/js/browse.js` | new | browse view + listeners |

### Dependencies
- **Requires:** sprint branch HEAD `aa020a0` (iterations 1–4: shell, router, card incl. row modifier, model page, advisor).
- **Blocks:** nothing in-sprint. F1 answer later = one constant in radar `Search`, applied to Home, Browse, advisor together.

### Constraints
- HANDOFF §3.6 no build step / framework → native ES module; §3.5 375 + 16px gutter, no horizontal scroll; §3.7 a11y (≥ 24px targets, visible focus).
- Canvas governs look/IA/copy/states; def governs data semantics. Browse shows only real data (listings, valuations, counts).
- Grades: cards keep the shop's own label + zh meaning (`cardHTML` unchanged); 最低成色 uses Kura's internal rank as a floor only, never shown (def Canvas deviation "Grade filter"; HANDOFF §3.1).
- F1 / F6 / member entry / 大家都在找 not decided: radar predicate + sort constants untouched.
- Client never converts JPY → TWD. Source Hygiene: no projex IDs in comments; non-obvious decisions name the rejected alternative.
- No crawls (serve without crawl flags; never `POST /api/crawl`); commits on the sprint branch in the worktree only.

### Assumptions
- **A1 — URL = state.** Params mirror `/api/listings`: `deals` (`1` = on) · `brand` (canonical name from `ov.brands`) · `category` (key from `ov.categories`) · `min_rank` (`S|A|B`) · `max_twd` (positive integer) · `q` (trimmed) · `sort` (`newest|price_asc|price_desc`; absent/invalid = `deal`). Absent = off / any. Invalid values are ignored (control shows the default; not passed to the API). Filter changes `history.replaceState` — a tweak is not a history entry, so Back leaves Browse (rejected: `pushState` per change → Back steps through every select). Canonical URL omits `sort=deal`; Home's `/browse?deals=1&sort=deal` still parses to the same state. Bare `/browse` = switch off, all active, 低於行情優先 (rejected: absent `deals` = on — a header search would then silently hide every non-deal match).
- **A2 — Heading** (h1, crumb, title): `deals` on → 價格最美 (canvas); off + `sort=newest` → 最新上架 (canvas Home section heading, the page Home's 看全部最新上架 opens); off otherwise → 全部在售 (plan-authored; 在售 = canvas model-tile wording). Canvas draws only the 價格最美 state; a 價格最美 heading over non-deal results would label them as deals. Orchestrator: canvas gap (non-deal / search heading) → canvas pass 2.
- **A3 — 品牌 options** = `ov.brands` entries with `overview.by_brand[name] > 0`, by count desc, label `{name}（{num(count)}）` (canvas format; real list = canvas list). Counts are all active listings, not the current filter (canvas sample = HANDOFF totals); option counts may exceed results under the switch.
- **A4 — 類型 options** = `ov.categories` with `overview.by_category[key] > 0`, by count desc, zh label, no count (canvas shows none). Canvas lists 3 (手提包 / 肩背包 / 托特包); data has 10 non-empty → all 10 (the canvas 3 are the top 3); 5 empty catalog categories hidden (each would open a 0-result list). Needs `by_category` (Step 1). Orchestrator: canvas deviation (option set).
- **A5 — 最低成色 → rank floor** `""` / `S` / `A` / `B` — legacy UI mapping (`legacy/index.html`), matching BRAND OFF definitions (未使用品 = 未使用; ランクA = 輕微使用感; ランクB = 使用感; ランクC below 明顯). The 1 unranked listing drops out once a floor is set (existing radar behaviour) — def L4 「grades with no defensible mapping — excluded from the filter?」 stays open; not decided here.
- **A6 — Paging:** `PAGE = 40` (10 desktop rows of 4 at 1280; plan-authored, canvas sample shows 5 + 「再看 1,015 件」). Load-more = `offset = cards shown`; next page appended. Index order is stable between crawls (no crawl in the mock run); a crawl between pages could shift items — accepted (Risks).
- **A7 — Mobile 篩選與排序** = in-place disclosure (button `aria-expanded` + panel `hidden`) of the same five fields the desktop row draws, stacked full width at 44px. Canvas draws the button only, no open state; reusing its own fields invents no new component (rejected: sheet / dialog — a new surface for canvas pass 2). Orchestrator: canvas gap (mobile filter panel open state).
- **A8 — Zero results:** 「沒有符合條件的包。」 (plan-authored, quiet 16px ink-2 line where the grid sits; count reads `0 件`). Canvas has only site-empty 「還在準備中。」, used here when `overview.active === 0` (canvas `data: empty` covers every page). Orchestrator: canvas gap (designed zero-result state).
- **A9 — Model 看全部 stays in place.** Browse filters by brand / 類型 / 成色 / budget / keyword, not by model group (2,281 groups have no code; keywords any-match across models), so no Browse URL lists exactly a model's N listings and the link's count would not match the page it opens. Model-page deviation stands with this reason; only its comment changes (present tense). Orchestrator: keep as canvas deviation; a `group` filter would be its own change.
- **A10 — Header search in-app:** `app.js` intercepts `form.k-search` submit → `/browse?q=…` (empty → `/browse`), a fresh state (other filters dropped). The form keeps `action`/`method` so it works before the module loads.
- **A11 — Apply on commit**, no apply button (canvas form has none): selects + switch on `change`; text fields on `change` and Enter. `max_twd` keeps digits only (「2,0000」 → 20000; none → unset, field cleared).
- **A12 — Dark × browse** token-driven (no artboard; def Canvas gaps) — every colour via `--k-*` tokens; checked in Step 5.

### Impact Analysis
- **Direct:** new `/browse` view; `/api/listings` accepts `offset`; overview gains `by_category`; header search no longer reloads the page.
- **Adjacent:** Home (empty-state markup moves to a shared export — same output); advisor `search_listings` builds its own `Filter` (no `Offset` → 0, unchanged); `/legacy/` uses `/api/listings` without `offset` (unchanged); card track listener shared (browse cards toggle like Home's).
- **Downstream:** none outside the repo.

---

## Implementation

### Overview
Server paging + category counts (tested) → shell plumbing (route, header search, shared empty state, model comment) → CSS → view → browser matrix incl. an empty scratch DB for the site-empty state.

### [PATCHED] Step 1: Radar offset + per-category counts; API `offset`

**Objective:** S5 paging data, S3 類型 options, S6. **Confidence:** High. **Depends on:** None.
**Do-Projex:** Encouraged

**Files:** `internal/radar/radar.go`, `internal/radar/radar_test.go`, `internal/web/server.go`, `internal/web/server_test.go`.

**Changes:**

```go
// radar.go — Filter: insert before Limit
	Sort        string   `json:"sort,omitempty"` // deal | price_asc | price_desc | newest
	// Offset skips that many sorted matches, for paging; total still counts
	// every match.
	Offset int `json:"offset,omitempty"`
	Limit       int      `json:"limit,omitempty"`

// radar.go — Search, before:
	sortItems(out, f.Sort)
	total = len(out)
	limit := f.Limit
// after:
	sortItems(out, f.Sort)
	total = len(out)
	out = out[min(max(f.Offset, 0), len(out)):]
	limit := f.Limit
```
(`gofmt` aligns the struct; `items = make([]Item, len(out))` already yields `[]` for an empty page.)

```go
// radar.go — Overview struct: after ByBrand
	ByBrand    map[string]int `json:"by_brand"`
	ByCategory map[string]int `json:"by_category"` // active listings per category key
// Overview(): init ByCategory: map[string]int{} beside ByBrand; in the active branch after o.ByBrand[it.Brand]++:
		o.ByCategory[it.Category]++
```

`radar_test.go` — append to `TestSearchFilters`:
```go
	all, _ := e.Search(Filter{Sort: "price_asc"})
	page, total := e.Search(Filter{Sort: "price_asc", Offset: 1, Limit: 1})
	if total != 3 || len(page) != 1 || page[0].ID != all[1].ID {
		t.Errorf("offset page = %d %+v, want %s", total, page, all[1].ID)
	}
	if past, total := e.Search(Filter{Offset: 5}); total != 3 || len(past) != 0 {
		t.Errorf("offset past end = %d %+v", total, past)
	}
	if o := e.Overview(); o.ByCategory["tote"] != 3 {
		t.Errorf("ByCategory = %+v", o.ByCategory)
	}
```
(`row()` sets `Category: "tote"`.)

`server.go` `listings`: `Sort: q.Get("sort"), Offset: int(intParam(r, "offset")), Limit: int(intParam(r, "limit")),`.

`server_test.go` `TestListingAPI` — after the `/api/listings/wt:nope` check (wt:1, wt:2 exist):
```go
	getList := func(qs string) (total int, ids []string) {
		t.Helper()
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/listings?"+qs, nil))
		var lb struct {
			Total int          `json:"total"`
			Items []radar.Item `json:"items"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &lb); err != nil {
			t.Fatal(err)
		}
		for _, it := range lb.Items {
			ids = append(ids, it.ID)
		}
		return lb.Total, ids
	}
	_, first := getList("sort=newest&limit=1")
	if total, second := getList("sort=newest&limit=1&offset=1"); total != 2 || len(second) != 1 || len(first) != 1 || second[0] == first[0] {
		t.Errorf("offset 1: total %d, pages %v / %v", total, first, second)
	}
	if total, past := getList("offset=9"); total != 2 || len(past) != 0 {
		t.Errorf("offset past end: total %d, items %v", total, past)
	}
```

**Rationale:** paging belongs where the sorted match list is (one sort, one predicate); a client that fetched everything and sliced would ship 15k items. `by_category` lets 類型 list only non-empty types from the same index the counts come from.

**Verification:** `go vet ./...`; `go test ./internal/radar/ ./internal/web/`.

**If this fails:** revert the four files.

---

### [PATCHED] Step 2: Shell — route, in-app header search, shared empty state, model comment

**Objective:** S1, A10, S5 site-empty, A9 comment. **Confidence:** High. **Depends on:** None (Step 4 provides `browse.js`).
**Do-Projex:** Encouraged

**Files:** `internal/web/static/js/app.js`, `internal/web/static/js/home.js`, `internal/web/static/js/model.js`.

**Changes:**
- `app.js`:
  - `import { renderBrowse } from "./browse.js";` routes += `[/^\/browse$/, renderBrowse, "價格最美 · Kura 蔵"]` (after `/`).
  - Extract the click handler's tail into `function go(url) { history.pushState(null, "", url); render(); scrollTo(0, 0); }`; click handler calls `go(url.pathname + url.search + url.hash)`.
  - Add:
    ```js
    // The header search opens Browse in-app. A native GET would reload the
    // page and drop what lives only in memory (tracked listings, the advisor
    // conversation); the form keeps action="/browse" so it still works before
    // this module runs. A search starts a fresh Browse state.
    document.addEventListener("submit", (e) => {
      if (!e.target.matches("form.k-search")) return;
      e.preventDefault();
      const q = String(new FormData(e.target).get("q") ?? "").trim();
      go(q ? `/browse?q=${encodeURIComponent(q)}` : "/browse");
    });
    ```
- `home.js`: `export const EMPTY_HTML = \`<section class="k-empty"> <h1>還在準備中。</h1> <p>店裡正在整理，新的包很快就會上架。</p> </section>\`;` (same markup as now); `renderHome` uses `main.innerHTML = EMPTY_HTML; return;`.
- `model.js` listener comment →
  ```js
  // 看全部 reveals the cards already rendered rather than linking to Browse:
  // Browse filters by brand, type, condition, budget and keyword, not by model
  // group, so no Browse URL lists exactly this group and the link's count
  // would not match the page it opens. One document-level listener, installed
  // at module load, as in card.js: a listener per render would stack on
  // re-render.
  ```

**Rationale:** one navigation helper for links and the search form; the empty state stays one markup for every page (canvas `data: empty` covers all pages).

**Verification:** with Step 4 — Home unchanged at `/`; header search Enter → `/browse?q=…` without reload (B7).

**If this fails:** revert the three files.

---

### [PATCHED] Step 3: `kura.css` — browse

**Objective:** S2–S5, S7 look. **Confidence:** High. **Depends on:** None.

**Files:** `internal/web/static/kura.css`.

**Changes** (canvas `browseData` markup + `renderVals` `crumbPad`, `h2Size`, `gridStyle`, `cellStyle`; Tokens `.k-field`). Insert a `/* Browse (價格最美) — reuses .k-crumb, .k-card, .k-empty, .k-error */` block after the 型號頁 block:
1. Field (canvas component sheet): `.k-field { font:inherit; font-size:14px; color:var(--k-ink); background:var(--k-surface); border:1px solid var(--k-rule-strong); border-radius:2px; height:38px; padding:0 10px; box-sizing:border-box; width:100%; }` · `.k-field::placeholder { color:var(--k-ink-3); }`
2. Head: `.k-browse__head { display:flex; flex-wrap:wrap; align-items:baseline; gap:6px 14px; padding:0 0 24px; }` · `.k-browse__head h1 { margin:0; font-family:var(--k-font-display); font-weight:400; font-size:30px; line-height:1.3; }` · `.k-browse__count { font-size:14px; color:var(--k-ink-3); font-variant-numeric:tabular-nums; }`
3. Form: `.k-bf { display:flex; flex-wrap:wrap; align-items:flex-end; gap:12px; padding:0 0 32px; }` · `.k-bf label { display:flex; flex-direction:column; gap:6px; font-size:12px; color:var(--k-ink-3); }` · widths `.k-bf__brand { width:150px; }` `.k-bf__cat { width:120px; }` `.k-bf__cond { width:160px; }` `.k-bf__max { width:130px; }` `.k-bf__q { flex:1 1 180px; }` `.k-bf__sort { width:150px; }` · switch `.k-bf .k-switch { flex-direction:row; align-items:center; gap:8px; height:38px; font-size:14px; color:var(--k-ink); white-space:nowrap; cursor:pointer; }` · `.k-switch input { width:18px; height:18px; margin:0; accent-color:var(--k-accent); }` · `.k-bf--narrow { display:none; }`
4. Grid: `.k-bgrid { display:grid; grid-template-columns:repeat(auto-fill, minmax(240px, 1fr)); gap:48px 28px; }` · `.k-bgrid > * { min-width:0; }` · `.k-bgrid[aria-busy="true"] { opacity:.6; }` · `.k-bnone { margin:0; padding:24px 0 8px; font-size:16px; color:var(--k-ink-2); }`
5. Load-more: `.k-bmore { display:flex; justify-content:center; padding-top:36px; }` · `.k-bmore button { height:44px; padding:0 24px; font:inherit; font-size:14px; color:var(--k-ink); background:transparent; border:1px solid var(--k-ink-2); border-radius:2px; cursor:pointer; }` · `.k-bmore button:disabled { opacity:.6; cursor:default; }`
6. Mobile (inside the existing `@media (max-width: 767px)`):
   - `.k-browse__head h1 { font-size:24px; }`
   - `.k-bf--wide { display:none; }` · `.k-bf--narrow { display:flex; flex-direction:column; align-items:stretch; gap:10px; padding:0 0 12px; }` · `.k-bf--narrow .k-field { height:44px; }` · `.k-bf__bar { display:flex; gap:10px; align-items:center; }` · `.k-bf--narrow .k-switch { height:44px; }`
   - `.k-bf__toggle { flex:1 1 0; height:44px; font:inherit; font-size:14px; color:var(--k-ink); background:var(--k-surface); border:1px solid var(--k-rule-strong); border-radius:2px; display:flex; align-items:center; justify-content:center; gap:8px; cursor:pointer; }`
   - `.k-bf__panel { display:flex; flex-direction:column; gap:10px; }` · `.k-bf__panel[hidden] { display:none; }` (author `display` would otherwise override the `hidden` attribute)
   - `.k-bgrid { display:flex; flex-direction:column; gap:0; }` · `.k-bgrid > * { padding:20px 0; border-bottom:1px solid var(--k-rule); }`
   - Row cards — comment: same values as `.k-card--row`; applied here by width because the card markup is shared across widths, and choosing the layout in JS would leave cards in the wrong layout after a resize across 767px: `.k-bgrid .k-card { display:grid; grid-template-columns:116px minmax(0, 1fr); gap:16px; align-items:start; }` · `.k-bgrid .k-card__model { font-size:16px; }` · `.k-bgrid .k-card__price .k-price { font-size:15px; }`

No 768–1023 rule: `auto-fill` gives 3 columns there.

**Rationale:** transcribed canvas values; `.k-field` is the canvas component-sheet field (first form on the canvas shell); reuses `.k-crumb`, `.k-card`, `.k-empty`.

**Verification:** visual, Step 5.

**If this fails:** remove the block and the mobile additions.

---

### [PATCHED] Step 4: Browse view

**Objective:** S1–S5, S8. **Confidence:** Medium (two synced forms, async paging). **Depends on:** Steps 1–3.
**Verify-Projex:** Encouraged

**Files:** `internal/web/static/js/browse.js` (new).

**Changes:** `export async function renderBrowse(main, ov)`; imports `api, esc, num` (util), `cardHTML` (card), `EMPTY_HTML` (home).

**Module constants + comments:**
```js
const PAGE = 40; // ten desktop rows of four at 1280; load-more fetches the next PAGE
const SORTS = [["deal", "低於行情優先"], ["newest", "最新上架"], ["price_asc", "價格 ↑"], ["price_desc", "價格 ↓"]];
// Canvas labels → a floor on Kura's internal rank, the only scale one filter
// can apply across shops whose grades differ; cards still show each shop's own
// grade. S = 未使用品 and better, A = ランクA (輕微使用感), B = ランクB (使用感).
const CONDS = [["", "任何成色"], ["S", "未使用以上"], ["A", "輕微使用感以上"], ["B", "明顯使用感以上"]];
const FILTER_ICON = `<svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.4" aria-hidden="true"><path d="M2 4h12M4.5 8h7M7 12h2"/></svg>`;
let view = null; // { ov, st, seq } for the browse view on screen
```
Module comment (top): view at `/browse`; the URL query is the state (shareable, reload-safe), filter changes replace the history entry rather than adding one (Back leaves the page instead of stepping through every select); every count comes from the server; the deal list is radar's `deals=1` predicate and `deal` sort, the same query as Home's 價格最美 shelf and count.

**Functions:**
1. `readState(ov, search)` → `{ deals, brand, category, min_rank, max_twd, q, sort }` per A1 (whitelists: `ov.brands[].name`, `ov.categories[].key`, `CONDS`/`SORTS` keys; `max_twd` `/^\d+$/` and > 0; `q` trimmed).
2. `stateQuery(st)` → `URLSearchParams` in order `deals brand category min_rank max_twd q sort`, empty/off omitted, `sort` omitted when `deal`. `apiURL(st, offset)` = `/api/listings?${stateQuery(st)}&sort=${st.sort}&limit=${PAGE}&offset=${offset}` (sort always explicit).
3. `heading(st)` per A2.
4. `fieldsHTML(ov, st, wide)` → object `{ brand, cat, cond, max, sort }` of the five non-keyword field strings (so each form places them in its own order); `wide` adds the width classes `k-bf__brand|cat|cond|max|sort`. Each = `<label class="{cls}">{k}<select class="k-field" name="{name}">…</select></label>`; selected option via `selected`. Brand options per A3 (`ov.overview.by_brand`), 類型 per A4 (`ov.overview.by_category`, labels from `ov.categories`), 成色 `CONDS`, budget `<input class="k-field" name="max_twd" inputmode="numeric" autocomplete="off" placeholder="不限" value="{st.max_twd}">`, sort `SORTS`.
5. `keywordHTML(st, placeholder, cls)` = `<label class="{cls}">關鍵字<input class="k-field" type="search" name="q" autocomplete="off" placeholder="{placeholder}" value="{esc(st.q)}"></label>`; `switchHTML(st)` = `<label class="k-switch"><input type="checkbox" role="switch" name="deals"{st.deals ? " checked" : ""}>只看低於行情</label>`.
6. `cellsHTML(items)` = `items.map((it) => \`<div role="listitem">${cardHTML(it)}</div>\`).join("")`; `resultsHTML(data)` = 0 total → `<p class="k-bnone">沒有符合條件的包。</p>`; else `<div role="list" class="k-bgrid">{cells}</div>` + `moreHTML(data.total - data.items.length)`; `moreHTML(rest)` = rest > 0 ? `<div class="k-bmore"><button type="button" data-browse-more>再看 ${num(rest)} 件</button></div>` : "".

**`renderBrowse(main, ov)`:**
- `if (ov.overview.active === 0) { main.innerHTML = EMPTY_HTML; return; }` (route default title)
- `const st = readState(ov, location.search); const data = await api(apiURL(st, 0)); view = { ov, st, seq: 0 }; const h = heading(st);`
- Markup (all interpolation `esc`, figures `num`):
```html
<div class="k-browse">
  <nav aria-label="位置" class="k-crumb"><a href="/" class="k-link">首頁</a><span aria-hidden="true">/</span><span aria-current="page" data-browse-h>{h}</span></nav>
  <div class="k-browse__head"><h1 data-browse-h>{h}</h1><span class="k-browse__count" aria-live="polite" data-browse-count>{num(data.total)} 件</span></div>
  <form class="k-bf k-bf--wide" aria-label="篩選與排序">{fieldsHTML 品牌/類型/最低成色/預算 (wide)}{keywordHTML(st, "型號、編號，可用日文：ネヴァーフル", "k-bf__q")}{sort field (wide)}{switchHTML}</form>
  <form class="k-bf k-bf--narrow" aria-label="篩選與排序">{keywordHTML(st, "型號、編號，可用日文", "")}
    <div class="k-bf__bar"><button type="button" class="k-bf__toggle" data-browse-toggle aria-expanded="false" aria-controls="bf-panel">{FILTER_ICON}篩選與排序</button>{switchHTML}</div>
    <div class="k-bf__panel" id="bf-panel" hidden>{fieldsHTML 品牌/類型/最低成色/預算/排序 (narrow)}</div></form>
  <div data-browse-results>{resultsHTML(data)}</div>
</div>
```
  (`fieldsHTML` returns the parts separately so the wide form puts 關鍵字 between 預算 and 排序, canvas order.) Return `` `${h} · Kura 蔵` ``.
- Comment on the two forms: the canvas draws a different form per width; both are rendered and CSS shows one (`display: none` also removes the other from the tab order and accessibility tree), kept in step on every change (rejected: one form reordered with CSS `order` — keyboard order would no longer follow the visual order at one of the widths).

**`apply(form)`** (async):
1. `const fd = new FormData(form); const v = (k) => String(fd.get(k) ?? "");` build `next` = `readState(view.ov, "?" + new URLSearchParams({ deals: form.elements.deals.checked ? "1" : "", brand: v("brand"), category: v("category"), min_rank: v("min_rank"), max_twd: v("max_twd").replace(/\D/g, ""), q: v("q"), sort: v("sort") }))` — one whitelist for URL and form input (both forms carry all seven names).
2. `const qs = stateQuery(next).toString(); const url = "/browse" + (qs ? "?" + qs : "");` if `url === location.pathname + location.search` → sync controls, return (idempotent: Enter + change fire once each).
3. `history.replaceState(null, "", url); view.st = next;` set every `[data-browse-h]` text and `document.title` to `heading(next)` (+ ` · Kura 蔵`); `syncControls(next)` — for each `.k-browse [name]`: checkbox `checked = next.deals`, else `value = next[name]` (normalizes budget text; mirrors the other form).
4. `const mine = ++view.seq;` grid (if any) `aria-busy="true"`; `try { data = await api(apiURL(next, 0)) } catch (err) { if (mine === view.seq) results.innerHTML = '<p class="k-error">暫時無法載入資料，請稍後再試。</p>'; console.warn(err); return; }`; `if (mine !== view.seq || !results.isConnected) return;` then results `innerHTML = resultsHTML(data)`, count text `{num(data.total)} 件`.

**`loadMore(btn)`** (async): `btn.disabled = true; const mine = view.seq; const grid = document.querySelector(".k-browse .k-bgrid"); const shown = grid.children.length;` fetch `apiURL(view.st, shown)`; on error `btn.disabled = false`, `console.warn`, return; `if (mine !== view.seq || !grid.isConnected) return;` `grid.insertAdjacentHTML("beforeend", cellsHTML(data.items))`; `rest = data.total - grid.children.length`; rest > 0 → `btn.textContent = \`再看 ${num(rest)} 件\``, `btn.disabled = false`; else `btn.closest(".k-bmore").remove()`; focus `grid.children[shown]?.querySelector(".k-card__model")`. Comment: focus moves to the first new card so keyboard and screen-reader users continue where the new cards start instead of on a removed or relabelled button.

**Listeners** (document-level, installed at module load, as `card.js` / `model.js` — a listener per render would stack):
- `change`: `const form = e.target.closest(".k-browse .k-bf"); if (form) apply(form);`
- `keydown`: Enter in a `.k-browse .k-bf input` (not the checkbox) → `e.preventDefault(); apply(form)` (text fields commit on Enter as well as on blur; neither form has a submit button, so Enter would otherwise do nothing).
- `submit`: `.k-browse .k-bf` → `e.preventDefault()` (no native GET reload).
- `click`: `[data-browse-toggle]` → flip `aria-expanded`, `document.getElementById(btn.getAttribute("aria-controls")).hidden = !open`; `[data-browse-more]` → `loadMore(btn)`.

**Rationale:** URL-as-state reproduces every Home / search entry exactly and survives reload; results re-render without touching the form, so focus and typing stay put; one `apiURL` for first page and load-more keeps the paging query identical to the counted one.

**Verification:** Step 5 rows B1–B14.

**If this fails:** delete `browse.js`, drop the route + import in `app.js` (Browse returns to not-found; header search interception may stay — it still lands on `/browse`).

---

### [PATCHED] Step 5: End-to-end verification

**Objective:** S1–S9 evidence. **Confidence:** Medium (hotlinked photos/fonts need network). **Depends on:** Steps 1–4.
**Verify-Projex:** Encouraged

**Procedure:**
1. `go vet ./...` + `go test ./...`; hygiene grep `grep -rnE "[0-9]{10}-[a-z0-9-]+-(plan|sprint|def|nav|audit|patch)" internal cmd` → none; `git diff aa020a0 -- internal/radar/radar.go` touches no `0.05`, `anomalyPct`, `dealScore`, `confidenceWeight` line (S8).
2. Background `go run ./cmd/kura serve -db kura.db -addr 127.0.0.1:8090` (worktree root; no crawl flags; never `POST /api/crawl`). Empty-state server: `go run ./cmd/kura serve -db <scratchpad>/empty.db -addr 127.0.0.1:8092` (fresh DB).
3. Matrix (browser pane):

| # | Case | URL / action | Check |
|---|---|---|---|
| B1 | 價格最美 from Home | `/` → click 看全部價格最美 N 件 | URL `/browse?deals=1&sort=deal`; h1 + crumb 價格最美; count = Home N (1,471); first 3 card ids = Home shelf; switch on, sort 低於行情優先; 40 cards; 「再看 1,431 件」 |
| B2 | 最新上架 from Home | `/` → 看全部最新上架 | h1 最新上架; switch off; sort 最新上架; count 15,367; first 5 ids = Home 最新上架 row |
| B3 | load-more | B1 → click 再看 ×2 | 80 → 120 cards, button 1,391 → 1,351; no duplicate `data-id`; focus on card 41 / 81 model link |
| B4 | last page | `/browse?brand=GOYARD` → 再看 ×2 | 40 → 80 → 98; button 再看 58 件 → 18 件 → gone |
| B5 | each filter | from B1 / bare `/browse`: brand CHANEL, 類型 旅行包, 最低成色 未使用以上 / 輕微 / 明顯, budget `2,0000`, keyword ネヴァーフル + Enter, sort 價格 ↑ / ↓, switch off/on | URL params per A1 (budget field → `20000`); counts = probe (`min_rank` S 831 / A 10,152 / B 15,322; `max_twd=20000` 2,082; `category=travel` 37; `q` 290 off / 63 on); price sorts monotone on cards; heading flips 價格最美 ↔ 全部在售 with the switch; reload keeps state + controls |
| B6 | zero results | `/browse?deals=1&brand=GOYARD` | 「沒有符合條件的包。」, `0 件`, no load-more |
| B7 | header search | on `/` set `window.__probe = 1`, search ネヴァーフル + Enter; then empty search | `/browse?q=…`, h1 全部在售, keyword field filled, count 290, `window.__probe === 1` (no reload); empty → `/browse` |
| B8 | mobile icon | 375: tap header search icon | `/browse`, 全部在售, 15,367 |
| B9 | history | Home → B1 → change brand → card → detail → Back → Back | Back from detail restores Browse with brand state; next Back → Home (filter tweak added no entry) |
| B10 | invalid params | `/browse?brand=NOPE&sort=x&min_rank=Z&max_twd=-5&deals=yes` | defaults: 全部在售, 15,367, controls at defaults; `/browse/x` → not-found + title |
| B11 | mobile form | 375: B1 → 篩選與排序 → brand CHANEL → switch off | panel opens (`aria-expanded=true`), stays open; results/count update; resize to 1280 → wide form shows CHANEL + switch off |
| B12 | layouts/themes | B1 at 1280 + 375 × `?theme=light|dark` (+ once colorScheme emulation) | canvas `Browse` / `BrowseMobile`; 4 columns at 1280; row cards with rules at 375; tokens switch; `scrollWidth <= innerWidth` at 375 |
| B13 | keyboard | 1280 B1: Tab through crumb → 7 controls → cards → 再看 | visible ring on each; Space toggles switch; Enter in keyword applies |
| B14 | site empty | :8092 `/browse?deals=1` | 「還在準備中。」 |

4. Per view: `read_console_messages` onlyErrors empty; page text has no sample values (`1,020 件`, `1,015`) unless real.
5. Model page regression: `/model/CHANEL/name%3A%E3%83%9E%E3%83%88%E3%83%A9%E3%83%83%E3%82%BB` 看全部 still reveals in place.
6. Stop servers; delete `<scratchpad>/empty.db*`; worktree `git status` clean except committed work.

---

## Verification Plan

### Automated Checks
- [x] `go vet ./...` clean
- [x] `go test ./...` green incl. extended `TestSearchFilters`, `TestListingAPI`

### Manual Verification
- [x] Step 5 matrix B1–B14
- [x] Home, detail, model, advisor unchanged (Home still 5 + 3 cards; model 看全部 in place)
- [x] Source hygiene grep → none

### Acceptance Criteria Validation

| Criterion | How to Verify | Expected Result |
|---|---|---|
| S1 | B1, B2, B7, B8, B10; `document.title` | view on click / search / reload; not-found for `/browse/x` |
| S2 | B1, B2 vs Home + `/api/listings` | heading per A2; count = API total = Home N; shelf order kept |
| S3 | B5, B10, B12 screenshot vs canvas | canvas fields/labels/order; URL ↔ controls ↔ results |
| S4 | B11, B12 (375) | keyword + toggle + switch; panel in place; forms in sync |
| S5 | B1, B3, B4, B6, B14 | 40/page, 再看 count, focus, last page, zero + site-empty states |
| S6 | Step 1 tests | offset paging; `by_category` |
| S7 | B12, B13; console | both themes/widths, no overflow, rings, 0 errors |
| S8 | Step 5.1 diff check; code review of `browse.js` | counts from server only; radar predicate/sort untouched |
| S9 | vet / test / grep | green / none |

**Data-class trace (browse):** cards ← `/api/listings` items (real: listing fields, shop grade, TWD price; valuation badges = radar fair value over current listings, as Home) | count ← API `total` (real) | 價格最美 membership + order ← radar `OnlyDeals` + `deal` sort (the predicate Home and advisor `only_deals` use) | 品牌 counts ← overview `by_brand` (real) | 類型 options ← overview `by_category` (real) | 再看 N ← `total − shown` (real). Real-derived: none shown. Mocked: none.

---

## Rollback Plan

1. `git -C <work-root> log --oneline aa020a0..HEAD` → this plan's commits.
2. `git -C <work-root> revert <sha>…` newest first (sprint branch only); `/browse` returns to the not-found view, header search back to native GET.

---

## Notes

### Split Decision
No split — single scope (one `.projex` folder; `internal/radar` + `internal/web` are one repo scope, as iterations 1–4), within size budget (~420 lines; 5 steps — heuristic needs > 5 steps and > 500 lines). Fits one strategy run.

### Risks
- Two forms kept in step by `syncControls` — a missed control would show stale state at the other width (B11 checks).
- A crawl between page fetches could shift sorted positions (duplicate / skipped card); the mock runs without crawls. Hosted realtime (roadmap Phase 6) would need a cursor or snapshot version.
- `change` + Enter both trigger `apply`; the URL-equality guard makes the second a no-op.
- app.js `current` (last rendered path+query) is not updated by `replaceState`; harmless — `popstate` re-renders whenever the URL differs from it, and a re-render reads the URL. **Disproved** (audit F1: a later `#` step re-rendered Browse, dropping load-more pages) — fixed in 2610130952-kura-browse-audit-fixes-patch.md (`nav.js` `replaceURL`).
- Hotlinked photos / fonts unreachable → same exposure as Home/detail.

### Open Questions
None blocking execution. For the orchestrator to log (sprint nav / def), not decided here:
- A2 heading for non-deal / search states (全部在售), A7 mobile filter panel open state, A8 zero-result copy → def § Canvas gaps (canvas pass 2).
- A4 類型 option set (10 real vs canvas 3), A9 model 看全部 in place (reason now: no group filter) → canvas deviations.
- A5 unranked listings drop out under any 最低成色 floor → def L4 「grades with no defensible mapping」 still open.
- A12 dark × browse built token-driven → def § Canvas gaps tally (dark × browse: built, token-driven).
- F1 / F6 unchanged — Browse inherits whatever radar `OnlyDeals` / `deal` sort become.
