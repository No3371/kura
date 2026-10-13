# Kura footer IA + 行情怎麼算

> **Sprint:** 2610130524-kura-interactive-mock-sprint.md — iteration 7
> **Status:** Complete
> **Author:** Claude Opus 5.5 (planner, sprint subagent)
> **Parent:** 2610130524-kura-interactive-mock-sprint.md
> **Source:** Sprint derive, iteration 7 (sprint nav § Current Position + Revision Log, commit `f3a6c04`)
> **Related Projex:** 2610130524-kura-interactive-mock-sprint.md (footer IA evidence, judgment, constraints) | 2610130449-kura-mock-def.md (Real / Real-derived / Mocked; 行情怎麼算 hosts the disclosure; L4 grade ordering + freshness open) | 2610130420-kura-roadmap-nav.md (Phase 3: 行情怎麼算 under canvas pass 2) | 2610130602-kura-canvas-foundation-home-audit.md (F1 / F6 open) | 2610131010-kura-member-notifications-plan.md (mocked feed + watchlist semantics, A1–A9) | 2610130722-kura-model-page-plan.md (`MinSoldHistory` gate, 平均找到新主人 = median) | 2610131121-kura-methodology-footer-patch.md (execution)
> Partial Execution: Steps 1–4 (all) completed via 2610131121-kura-methodology-footer-patch.md — code commit `c19a8a7`; nothing remains open
> **Worktree:** Yes — existing sprint worktree `kura/.projexwt/2610130524-kura-interactive-mock`, branch `projex/sprint/2610130524-kura-interactive-mock`; execute via patch-projex plan-execution mode, no new branch

---

## Summary

Make the footer IA reachable and build **行情怎麼算** at `/methodology`: the def's methodology + real / real-derived / mocked disclosure, composed only from canvas evidence already in the app (page shell, `.k-crumb`, `.k-dsec` brass-ruled h2, 40em prose at line-height 1.85, key–value `dl`, `.k-badge` / `.k-grade`). Every threshold the page states is read from the radar's own constants (new `method` block on `/api/overview`) or from the card badge band's single definition (`card.js` `NEAR_BAND`) — never typed into copy, so an F1 / F6 answer flows through. The other five footer slugs open an honest "still being prepared" page (no invented content). Detail 行情估價's 行情怎麼算 link goes live; detail prose and the advisor prompt stop repeating the same thresholds as literals.

**Scope:** `internal/radar` (export constants), `internal/listing` (`Schemes()`), `internal/web` (`/api/overview` `method` + `schemes`; tests; static: new `js/pages.js`, `app.js`, `card.js`, `detail.js`, `kura.css`), `internal/advisor` (prompt thresholds from constants + test). One `.projex` scope (repo root).
**Estimated Changes:** 11 files — `radar.go`, `condition.go`, `server.go`, `server_test.go`, `prompt.go`, `agent_test.go`, `app.js`, `card.js`, `detail.js`, `kura.css` edited; `pages.js` added.

---

## Objective

### Problem / Gap / Need
Footer `nav aria-label="頁尾"` links `/about` `/methodology` `/faq` `/contact` `/privacy` `/terms` and detail 行情估價 「行情怎麼算」 (`detail.js` `fairHTML`) all land on `renderNotFound` (「找不到這個頁面」) — the last canvas-linked surfaces not reachable. Since iteration 6 the mocked set (provisional 登入 → 林小姐, persona watchlist, 通知 feed) is live and unlabelled with no disclosure anywhere → the def's precondition for unlabelled mocks (行情怎麼算 explains real vs mocked; red team F19) is unmet. Thresholds the page must explain exist only as unexported literals (`IsDeal` `0.05`, `anomalyPct`, `minComps`) or a client literal (`card.js` `0.03`), and are already repeated as copy in `detail.js` ("不足 3 件", "低一半以上") and `prompt.go` ("at least 5%", "低一半以上").

### Success Criteria
- [x] **S1 Routes:** `/methodology` renders 行情怎麼算 (title `行情怎麼算 · Kura 蔵`); `/about` `/faq` `/contact` `/privacy` `/terms` render the pending view, h1 = the footer link's own text, title `{text} · Kura 蔵`. In-app click, direct load / reload (SPA fallback) and Back all work. Any other unknown path still renders not-found.
- [x] **S2 Detail link live:** detail 行情估價 「行情怎麼算」 opens `/methodology` in-app.
- [x] **S3 Content:** `/methodology` = crumb (首頁 / 行情怎麼算) + h1 + lead + five sections `#source` 店家資料 · `#fair` 行情估價 · `#badges` 標籤怎麼看 · `#history` 需要時間累積的數字 · `#demo` 哪些是示範, copy per Step 2. `#source` lists each registered shop's grade ladder verbatim (label + 中文 + 「日文原文」) with its scheme name, grader and a link to the shop's definition page. `#badges` renders the six live badges via `badgeHTML`.
- [x] **S4 Disclosure = def's mocked set exactly:** sign-in + demo member | persona watchlist | 通知 feed (who-tracks mocked; events/figures real, not delivered) | demand 大家都在找 (not shown). Everything else stated as shop data or computed from it; advisor shares the page's computations; advisor 示範模式 meaning stated. No new in-UI label anywhere else.
- [x] **S5 One source per threshold:** every number on `/methodology` comes from `ov.method`, `ov.overview`, `ov.schemes` or `NEAR_BAND` — `pages.js` contains no threshold literal. `detail.js` fair prose reads `ov.method`; `prompt.go` formats `radar.DealMinPct` / `radar.AnomalyPct`. Tests prove `/api/overview` `method` = radar constants and the advisor prompt states them. On the worktree DB the page reads 3 / 5 / 50 / 14 (band / 價格最美 / anomaly / history days) and min comps 3.
- [x] **S6 Not stated:** no internal rank codes, factors or cross-shop mapping (def L4 open); no data-age / last-updated / crawl-cadence claim (def L4 freshness open); no contact, company or legal content; F1 / F6 constant values unchanged (`DealMinPct` 0.05, `AnomalyPct` 0.5).
- [x] **S7 Look:** only existing tokens and classes plus a small `.k-method` block composed from them; 1280 + 375 × light/dark (no artboard → token-driven); no horizontal page scroll at 375; targets ≥ 24 px; visible focus ring; 0 console errors.
- [x] **S8** `go vet ./...` + `go test ./...` green; Source Hygiene grep clean.

### Out of Scope
- Content of 關於 Kura / 常見問題 / 聯絡我們 / 隱私權政策 / 服務條款 — sprint Deferred (needs real contact / legal / company facts from the human + canvas pass 2)
- 資料管線 page, HANDOFF §4.A status line / 近期降價 — sprint Deferred
- Deciding F1 (價格最美 bar), F6 (anomaly-boundary ordering), def L2 unlabelled-mock confirmation, L3 ranking definitions / persona wording, L4 grade-mapping disclosure, L4 freshness — constants keep today's values; the page describes what the code does
- Overview `Deals` 10% bar (`radar.Overview`) — not stated on the page; unchanged
- Advisor tool `min_rank` description (model-facing, not UI) — unchanged
- Router-wide focus on route change, `aria-current` in the footer nav — carry-over shell a11y (iter 4)
- Deep link `/methodology#fair` — `go()` scrolls to top and ignores hashes; plain `/methodology` kept
- Canvas pass 2 redesign of the page look / copy

---

## Context

### Current State
- **Router** (`app.js`): 4 routes (`/`, `/browse`, `/listing/{id}`, `/model/{b}/{k}`); else `renderNotFound` (`<section class="k-notfound">` h1 + 回到首頁). Document click → `go()` for any same-origin `a[href^='/']` not under `/legacy|dev|api` → footer links already navigate in-app. Views: `view(el, ov, ...params)` → optional title string.
- **Footer** (`index.html`): six `<a href="/{slug}" class="k-link">` with canvas labels; bottom line 運費 / 真偽 + `#fx` rate span (filled by `app.js`).
- **Server:** `spaFallback` answers unknown GET paths with `index.html` → `/methodology` etc. already load the shell. `/api/overview` → `{overview, fx, advisor, runs, crawling, can_crawl, brands, categories}`; `overview` carries `since` (epoch), `sold_stats_ready`.
- **Radar constants:** `minComps = 3` (unexported const block), `anomalyPct = 0.5` (after `value`), `IsDeal` literal `DealPct >= 0.05`; `MinSoldHistory = 14 * 24 * time.Hour` exported. Confidence: same-grade ≥ `minComps` → medium, ≥ 6 → high; rank-adjusted → low, ≥ 6 → medium; dispersion > 0.4 → one level down, > 0.6 → low; `name:` group key → low. Comparables: same brand + group, gone listings older than `compWindow` (180 d) excluded. `dealScore` = `DealPct × confidenceWeight` (medium 0.7), anomaly → 0. Sold stats: `Sold` counts every non-active listing; days-to-sell only for listings first seen after epoch + 24 h (`watched`), median.
- **Card band** (`card.js` `valuation`): anomaly → `anomaly`; `deal_pct >= 0.03` → `below` (`low` if confidence low); `<= -0.03` → `above`; else `near`; no fair → `none`. `badgeHTML(v)` renders the six badge states.
- **Detail** (`detail.js` `fairHTML`): 行情怎麼算 link `<a href="/methodology" class="k-link">` in the 方法/樣本/信心 `dl`; literals "同型號的比較樣本不足 3 件" (no fair), "不到 3 件" (rank-adjusted), "低一半以上" (anomaly).
- **Advisor** (`prompt.go`): system prompt line "price_anomaly 為 true 的商品比行情低一半以上"; `search_listings.only_deals` description "Only listings at least 5% below fair value". `buildSystemPrompt` already substitutes `{{categories}}` / `{{brands}}` via `strings.NewReplacer`.
- **Grades:** `listing.Scheme{Source, Name, Shop, Graded, Ref, Grades}` registered per source (`brandoff` init: "BRAND OFF 商品ランク", 店家鑑定, FAQ URL, 5 grades best first). Registry has `SchemeFor` only — no list accessor.
- **Member mock** (`member.go`, `member.js`): sign-in → 林小姐 directly; state in `sessionStorage` `kura.watch` (new tab = guest); watchlist picked server-side from live listings/models; events = observed 降價 / 已找到新主人 + model's newest active listing (新上架); feed static, nothing delivered.
- **Canvas** (`project/Main.dc.html`, re-checked this plan via Artifact read, served 1791845608-ba84): footer six links `href="#"`; detail 行情估價 `dl` ends in 行情怎麼算 (`#`); prose `font-size: 16px; line-height: 1.85; max-width: 40em`; site-empty h1 「還在準備中。」 (display font 40px); sample note "sample values for layout only; real values need ~2 weeks of history". No artboard or copy for any footer page.
- **CSS:** `.k-dsec > h2` (brass top rule, display font 22px), `.k-fair p` (16px / 1.85 / 40em), `.k-empty, .k-notfound` (+ h1 40px display, p 16px / 1.8 ink-2), `.k-crumb`, `.k-badge*`, `.k-grade`; tokens `--k-ink-2/3`, `--k-rule`, `--k-font-ja`, `--k-font-display`. Mobile ≤ 767: `.k-wrap` 16 px gutter; detail/model h1 26px.

### Key Files

| File | Role | Change Summary |
|------|------|----------------|
| `internal/radar/radar.go` | valuation + deal predicate | export `MinComps`, `DealMinPct`, `AnomalyPct`; `IsDeal` / `value` use them |
| `internal/listing/condition.go` | grade schemes | `Schemes()` — all registered schemes, by source |
| `internal/web/server.go` | API | `/api/overview` + `method` + `schemes` |
| `internal/web/server_test.go` | tests | static routes for new pages/JS; `TestOverviewMethod` |
| `internal/advisor/prompt.go` | advisor prompt | anomaly line + `only_deals` description from radar constants |
| `internal/advisor/agent_test.go` | tests | prompt states the constants |
| `internal/web/static/js/pages.js` | new | `renderMethodology`, `renderPending` |
| `internal/web/static/js/app.js` | router | two routes |
| `internal/web/static/js/card.js` | badge | `NEAR_BAND` export; comment names radar constants |
| `internal/web/static/js/detail.js` | detail | fair prose reads `ov.method` |
| `internal/web/static/kura.css` | styles | `.k-method*`, `.k-pending` |

### Dependencies
- **Requires:** sprint branch HEAD `f3a6c04` (iterations 1–6).
- **Blocks:** iteration 8 closing pass (goal walk needs every canvas link reachable). A later F1 / F6 / L4 answer changes a constant or adds a section without touching the page's numbers.

### Constraints
- Canvas governs look / IA / copy; no canvas copy or layout exists for these pages → look composed only from existing canvas evidence, copy plan-authored and provisional until canvas pass 2 (sprint derive-7 judgment).
- Def governs data semantics: mocked list = exactly def's; one source per stat; mocked surfaces stay unlabelled elsewhere; advisor 示範模式 remains the only in-UI label.
- Never fabricate contact / company / legal facts; never decide open human questions (F1, F6, def L2 / L3 / L4).
- HANDOFF §3.6 no build step (native ES modules); §3.5 375 + 16 px gutter; §3.7 targets ≥ 24 px, visible focus.
- Source Hygiene (SKILL.md): no projex IDs in comments; non-obvious decisions name the rejected alternative; present tense.
- No crawls (serve without crawl flags; never `POST /api/crawl`); commits on the sprint branch in the worktree only.

### Assumptions
- **A1 — Other five slugs → pending page (plan's call per derive 7):** h1 = footer link text, 「這一頁的內容還在準備中。」 + link to 行情怎麼算. Rejected: keep not-found — a footer link answering 「找不到這個頁面」 reads as a broken site in a reviewer-facing demo, and def L2 makes the canvas footer pages reachable; remove the links — contradicts canvas IA. Logged as canvas gap + deviation (page body undesigned).
- **A2 — Constants travel on `/api/overview`** (`method` block), not a new endpoint. Rejected: `GET /api/methodology` — detail's fair prose needs `min_comps` / `anomaly_pct` too and already holds `ov`; a second endpoint means a second fetch per detail render for four numbers. Overview already bundles static config (`brands`, `categories`, `advisor`).
- **A3 — Badge band stays client-owned** as `card.js` `NEAR_BAND`, imported by the page. Rejected: server constant — no server code reads the band (only the badge does), and `valuation(it)` runs per card without `ov`. Deviation from the nav's "server constants via API" wording for this one value; single source kept.
- **A4 — Copy plan-authored, provisional** (Step 2 text) → canvas pass 2 copy review; recorded as def § Canvas gaps 行情怎麼算 content: built (provisional).
- **A5 — Rank bridge described, mapping not:** page says cross-grade estimates convert other grades of the same model to this grade and never replace the shop's grade; no rank codes, factors or ordering across shops. The Browse 最低成色 filter's use of the bridge is **not** described (def L4: disclose the mapping? / unmappable grades — open).
- **A6 — Grade ladders = registered schemes** (`listing.Schemes()`), each in its shop's own order (best first, as the shop lists it) — one ladder per shop, never merged. Registered = sources the binary crawls (BRAND OFF only). `Grade.rank` already ships in every listing payload; the page does not render it.
- **A7 — 已找到新主人 disclosed as Kura's inference** (listing left the shop's pages; shops do not publish sales). States current code behaviour; decides nothing.
- **A8 — History state from `sold_stats_ready` only:** page says whether the history has reached `min_sold_history_days`; it states no day count so far, no last-crawl time (def L4 freshness open). `since` (record start, same as the model page's 「自 M/D 起記錄」) is stated; omitted when zero (empty DB).
- **A9 — Adjacent literals moved** (detail prose, advisor prompt) so the 行情估價 section and the page it links to cannot disagree after a constant changes. Wording: "低一半以上" → "低超過 N%" (strict `>` matches `DealPct > AnomalyPct`).
- **A10 — Pending page label read from the footer link** (`index.html` holds the canvas copy once). Rejected: label map in JS — a second copy of six canvas strings.
- **A11 — 平均找到新主人 = median** disclosed on the page (label is canvas copy; value is the median — model-page comment). Relabelling stays a canvas decision.
- **A12 — Demo member not named** on the page (「示範會員」): the name comes from `/api/member`; the page does not fetch it.

### Impact Analysis
- **Direct:** two new client routes; `/api/overview` gains two keys; radar constants exported (values unchanged).
- **Adjacent:** `IsDeal` callers (Search, member pick, advisor `only_deals`) — same predicate, named constant. `valuation()` callers (card, detail, advisor recs, notify) — same band, named constant. `detail.js` fair prose wording ("一半以上" → "超過 50%"). Advisor system prompt text changes (cache key changes once per process — harmless). Tests registering schemes (`wt`, `wm`, new `wo`) share the global registry → `Schemes()` assertions use "contains".
- **Downstream:** none outside the repo.

---

## Implementation

### Overview
Server first (exported constants, scheme list, overview payload, advisor prompt — all unit-tested), then the page + router + styles, then detail's prose, then the browser matrix.

### Step 1: Constants on the server; overview `method` + `schemes`; advisor prompt [PATCHED] → 2610131121-kura-methodology-footer-patch.md

**Objective:** S5 / S6 server side. **Confidence:** High. **Depends on:** None.
**Verify-Projex:** Encouraged

**Files:** `internal/radar/radar.go`, `internal/listing/condition.go`, `internal/web/server.go`, `internal/web/server_test.go`, `internal/advisor/prompt.go`, `internal/advisor/agent_test.go`

**Changes:**

```go
// radar.go — Before:
const (
	minComps = 3
	// Gone listings older than this stop counting as comparables.
	compWindow = 180 * 24 * time.Hour
)
// ... value(): len(same) >= minComps / len(adjusted) >= minComps ... v.Anomaly = v.DealPct > anomalyPct
const anomalyPct = 0.5
// IsDeal: ... it.Fair.DealPct >= 0.05 ...

// radar.go — After (values unchanged; comment wording the executor's, hygiene rules apply):
const (
	// MinComps is the fewest comparables a fair value is estimated from; with
	// fewer a listing has no valuation.
	MinComps = 3
	// Gone listings older than this stop counting as comparables.
	compWindow = 180 * 24 * time.Hour
)

// DealMinPct is how far below fair value a listing must be priced to count as
// a deal (IsDeal). AnomalyPct: further below than this is flagged Anomaly —
// half price against a solid comparable set is rarely a bargain; it's usually
// damage or missing parts the title doesn't mention.
// The site's 行情怎麼算 page, detail prose and the advisor prompt state these
// values from here; copy that typed them would go stale when they change.
const (
	DealMinPct = 0.05
	AnomalyPct = 0.5
)
// value(): MinComps in both cases; v.Anomaly = v.DealPct > AnomalyPct
//   (drop the old `const anomalyPct` and move its comment as above)
// IsDeal: it.Fair.DealPct >= DealMinPct; doc comment says "at least DealMinPct under fair value"
```

```go
// condition.go — After (add; imports slices, strings):
// Schemes returns every registered grading scheme, ordered by source.
func Schemes() []Scheme {
	schemesMu.RLock()
	defer schemesMu.RUnlock()
	out := make([]Scheme, 0, len(schemes))
	for _, s := range schemes {
		out = append(out, s)
	}
	slices.SortFunc(out, func(a, b Scheme) int { return strings.Compare(a.Source, b.Source) })
	return out
}
```

```go
// server.go overview() — Before:
	return map[string]any{
		"overview": s.Radar.Overview(), "fx": s.FX.Current(), "advisor": s.Advisor.Mode(),
		"runs": runs, "crawling": s.crawling.Load(), "can_crawl": s.Crawl != nil,
		"brands": brands, "categories": cats,
	}, 0, nil

// After:
	return map[string]any{
		"overview": s.Radar.Overview(), "fx": s.FX.Current(), "advisor": s.Advisor.Mode(),
		"runs": runs, "crawling": s.crawling.Load(), "can_crawl": s.Crawl != nil,
		"brands": brands, "categories": cats,
		// The radar's own thresholds, for pages that explain them (行情怎麼算,
		// the detail page's 行情估價 prose): stated from here, never typed in.
		"method": map[string]any{
			"deal_min_pct": radar.DealMinPct, "anomaly_pct": radar.AnomalyPct,
			"min_comps": radar.MinComps, "min_sold_history_days": int(radar.MinSoldHistory / (24 * time.Hour)),
		},
		"schemes": listing.Schemes(), // each shop's grade ladder, verbatim
	}, 0, nil
```

```go
// prompt.go — Before:
//   systemPrompt: "- price_anomaly 為 true 的商品比行情低一半以上，多半有標題沒寫的瑕疵或缺件：不要當成好價推薦，"
//   buildSystemPrompt: strings.NewReplacer("{{categories}}", ..., "{{brands}}", ...)
//   toolDefs: "only_deals": {"description": "Only listings at least 5% below fair value"}
// After:
//   systemPrompt: "- price_anomaly 為 true 的商品比行情低超過 {{anomaly_pct}}%，多半有標題沒寫的瑕疵或缺件：不要當成好價推薦，"
//   buildSystemPrompt: add "{{anomaly_pct}}", pctText(radar.AnomalyPct) to the replacer
//   toolDefs: "description": "Only listings at least " + pctText(radar.DealMinPct) + "% below fair value"
//   + helper (imports math, strconv, radar):
// pctText formats a fraction as a whole percent, e.g. 0.05 → "5".
func pctText(f float64) string { return strconv.Itoa(int(math.Round(f * 100))) }
```

Tests:
- `server_test.go` `TestStaticRoutes` add: `{"/methodology", "text/html", `id="main"`, 200}`, `{"/about", "text/html", `id="main"`, 200}`, `{"/js/pages.js", "javascript", "renderMethodology", 200}`.
- `server_test.go` new `TestOverviewMethod`: register scheme `wo` (Shop `テスト店`, one grade); temp store (no rows needed); server as `TestListingAPI` plus `Advisor: &advisor.Mock{}` (overview calls `Advisor.Mode()`); GET `/api/overview` → 200; decode `method` → `deal_min_pct == radar.DealMinPct`, `anomaly_pct == radar.AnomalyPct`, `min_comps == radar.MinComps`, `min_sold_history_days == int(radar.MinSoldHistory/(24*time.Hour))`; `schemes` contains source `wo` with shop `テスト店`.
- `agent_test.go` new `TestPromptStatesThresholds`: `buildSystemPrompt()` contains `"低超過 " + pctText(radar.AnomalyPct) + "%"` and no `{{`; the `search_listings` tool's `only_deals` description (`OfTool.InputSchema.Properties` → `map[string]any` → `"only_deals"` → `"description"`) contains `"at least " + pctText(radar.DealMinPct) + "%"`.

**Rationale:** One definition per threshold, read by every consumer that states it. Values untouched → F1 / F6 undecided.
**Verification:** `go vet ./...`; `go test ./internal/radar ./internal/listing ./internal/web ./internal/advisor` green (existing `TestIsDeal` still passes at the 0.05 / 0.049 boundary); `curl -s 127.0.0.1:<port>/api/overview` shows `method` `{deal_min_pct:0.05, anomaly_pct:0.5, min_comps:3, min_sold_history_days:14}` and one `BRAND OFF` scheme.
**If this fails:** `git -C <work-root> checkout -- <files of this step>`; constants are a rename only — a compile error names the missed call site.

---

### Step 2: `/methodology` + pending pages; router; styles [PATCHED] → 2610131121-kura-methodology-footer-patch.md

**Objective:** S1, S3, S4, S5 (page), S6, S7. **Confidence:** Medium (copy + layout judgement). **Depends on:** Step 1.

**Files:** `internal/web/static/js/pages.js` (new), `internal/web/static/js/app.js`, `internal/web/static/js/card.js`, `internal/web/static/kura.css`

**Changes:**

`card.js` — Before: `if (f.deal_pct >= 0.03) …; if (f.deal_pct <= -0.03) …` with the header comment naming "deal_pct >= 0.05". After: `export const NEAR_BAND = 0.03;` used in both comparisons; comment: within `NEAR_BAND` of fair value reads as 接近行情 (legacy badge band); the band lives here, not on the server, because only the badge reads it and `valuation()` runs per card without the overview; 行情怎麼算 states it from this export. Keep the "not the deal predicate" paragraph but name `radar.IsDeal` / `radar.DealMinPct` instead of the literal 0.05 (Overview's 10% bar stays named as is).

`app.js` — After:
```js
import { renderMethodology, renderPending } from "./pages.js";
// routes += (before the closing bracket)
  [/^\/methodology$/, renderMethodology, "行情怎麼算 · Kura 蔵"],
  [/^\/(about|faq|contact|privacy|terms)$/, renderPending, "Kura 蔵"],
```

`pages.js` — new. Structure (exact markup/helpers the executor's; copy below is the spec — keep wording, fill `{…}` from the named source, `esc()` every interpolated string):

```js
import { esc, ymd, pct } from "./util.js";
import { badgeHTML, NEAR_BAND } from "./card.js";

// Header comment (rationale, hygiene rules apply): every threshold this page
// states comes from the overview's method block (the radar's constants) or
// card.js NEAR_BAND; a number typed into the copy would keep describing the old
// rule after the code changed. The page names no internal rank and makes no
// claim about how fresh the data is.

export function renderMethodology(el, ov) { … }   // returns nothing → route default title
export function renderPending(el, ov, slug) { … } // returns `${label} · Kura 蔵`
```

Values: `m = ov.method`; `band = pct(NEAR_BAND)`; `deal = pct(m.deal_min_pct)`; `anomaly = pct(m.anomaly_pct)`; `n = m.min_comps`; `days = m.min_sold_history_days`; `shops` = unique `ov.schemes[].shop`; `since` shown only when `Date.parse(ov.overview.since) > 0` (formatted `ymd`).

Page markup: `<div class="k-method">` → crumb `<nav aria-label="位置" class="k-crumb"><a href="/" class="k-link">首頁</a><span aria-hidden="true">/</span><span aria-current="page">行情怎麼算</span></nav>` → `<h1>行情怎麼算</h1>` → `<p class="k-method__lead">` → five `<section class="k-dsec" id="…" aria-labelledby="…-h"><h2 id="…-h">…</h2>…</section>` (last one `k-dsec k-dsec--last`).

Copy:

- **Lead:** Kura 上的數字分三種：店家刊登的資料、由這些資料算出來的行情與統計，以及為了展示設計而做的示範內容。這一頁說明每一種怎麼來。
- **`#source` 店家資料**
  - p: [if shops] 目前的資料來自 {shops 以「、」連接} 的線上商店。 商品、售價、照片、成色等級與等級說明，都照店家刊登的內容呈現。 [if since] Kura 自 {ymd(since)} 起記錄。
  - p: 售價以日圓刊登，Kura 用頁尾標示的匯率換算成新台幣，未含國際運費與進口稅費。Kura 不販售商品：每件商品都連到店家自己的商品頁，購買、付款與寄送都在店家完成。
  - per scheme: p `{name}（{graded}）` + [if ref] ` · <a class="k-link" href="{ref}" target="_blank" rel="noopener noreferrer">店家的等級說明<span class="k-sr">（另開視窗）</span></a>`; then `<dl class="k-method__dl">` rows `<div><dt><span class="k-grade">{label}</span></dt><dd>{desc_zh}<span class="k-method__ja" lang="ja">「{desc_ja}」</span></dd></div>` (omit the ja span when empty). Grade order = scheme order. Never render `rank`.
- **`#fair` 行情估價**
  - p: 行情是同品牌、同型號（有型號編號時以編號為準）的在架與近期下架商品售價的中位數。比較時優先找同一家店、同一個等級的商品：同店同等級有 {n} 件以上，就取它們的中位數（同店同等級中位數）。
  - p: 不到 {n} 件時，Kura 把同型號各等級的價格換算到這件的等級，再取中位數（跨等級換算）。這個換算只是估價用的參考，不會取代店家的等級：頁面上的等級一律是店家自己的等級與說明。同型號可以比較的商品不到 {n} 件時，不估計行情。
  - p: 信心表示這個估計有多可靠。同店同等級的樣本越多、價格越集中，信心越高；跨等級換算的估計最高是「中」。價格很分散（可能混有不同皮革、尺寸或限定版）時信心會調降；沒有型號編號、只能用名稱分組的估計一律是「低」，僅供參考。
  - p: 商品頁與型號頁的「同型號價格分布」，畫的是同型號在架商品依店家等級分組的售價；中位數、表格與顧問引用的是同一組數字。
- **`#badges` 標籤怎麼看** — `<dl class="k-method__dl">`, dt = `badgeHTML(...)`, dd:
  - `{state:"below", pct:"N"}` → 售價比行情低 {band}% 以上；N 是實際低多少。
  - `{state:"low"}` → 也低了 {band}% 以上，但估價信心低，僅供參考。
  - `{state:"near"}` → 和行情相差不到 {band}%。
  - `{state:"above", pct:"N"}` → 比行情高 {band}% 以上。
  - `{state:"anomaly"}` → 比行情低超過 {anomaly}%。低這麼多，通常代表標題沒寫的瑕疵、缺件或版本差異，購買前請先看店家的照片與備註。
  - `{state:"none"}` → 同型號可以比較的商品不到 {n} 件，暫不估計。
  - p after the dl: 「價格最美」收錄比行情低 {deal}% 以上、估價信心不是「低」、也不是異常低價的在售商品，依低於行情的幅度排序；信心「中」的估計排序時會打折。 [if band < deal] 所以標示「低於行情」但低不到 {deal}% 的商品，不會出現在價格最美。
- **`#history` 需要時間累積的數字**
  - p: 已售出件數、找到新主人的天數與售出率，要看商品從上架到售出的過程，需要累積記錄。Kura 的記錄滿 {days} 天之前，這些數字一律顯示「資料不足」：不是 0，也不代表賣不掉。 + (`ov.overview.sold_stats_ready` ? 目前的記錄已滿 {days} 天，這些數字已經開始計算。 : 目前的記錄還不到 {days} 天。)
  - p: 商品從店家的頁面下架時，Kura 把它標為「已找到新主人」；店家通常不公開成交，這是 Kura 的推定。找到新主人的天數只計算 Kura 看著它上架的商品：開始記錄時已經在架上的商品，不知道真正的上架日，不列入。型號頁的「平均找到新主人」顯示的是這些天數的中位數，不會被少數放了很久的商品拉高。
  - p: 商品頁的「售價變化」是 Kura 每次查看店家時記下的售價；店家調價，就會畫在這條線上。
- **`#demo` 哪些是示範** (`k-dsec--last`)
  - p: Kura 還沒有真的會員服務。以下內容是為了展示設計而做的示範，不是真實使用者的資料：
  - ul:
    - **登入與示範會員**：沒有帳號系統。按「登入」會直接以示範會員進入，狀態只保存在目前的瀏覽器分頁，新開的分頁會回到訪客。
    - **追蹤清單**：示範會員追蹤的型號與商品，是 Kura 從目前在架的商品裡挑出來的；你按的「追蹤」也只保存在這個分頁。
    - **通知**：誰在追蹤是示範的，但每則通知說的事和數字，都和連到的商品頁一致：新上架是這個型號在店家最新的一件，降價與已找到新主人只在 Kura 真的記錄到時出現。通知不會寄出。
    - **大家都在找**：需要真實使用者的瀏覽與追蹤資料。Kura 還沒有這些資料，所以目前不顯示。
  - p: 除此之外，頁面上的數字都來自店家資料，或由它算出。顧問回答裡的價格、行情與件數，和頁面用的是同一套計算；顧問標示「示範模式」時，回答是依規則搜尋的結果，沒有經過 AI 分析。

`renderPending(el, ov, slug)`: `label` = `document.querySelector(`.k-footer nav a[href="/${slug}"]`)?.textContent.trim() || "Kura 蔵"` (slug is one of five literals from the route regex). Markup: `<section class="k-empty k-pending"><h1>{label}</h1><p>這一頁的內容還在準備中。想知道 Kura 的數字怎麼來，可以先看<a href="/methodology" class="k-link">行情怎麼算</a>。</p></section>`. Comment (rationale): these pages' content — what Kura is, FAQ, contact details, legal terms — is the owner's to write and invented text would read as real; each slug still opens a page because the footer promises one, and the not-found view would read as a broken link (rejected).

`kura.css` — add after the `.k-empty` / `.k-notfound` block (tokens only; values the executor tunes against the matrix):
```css
/* 行情怎麼算 — reuses .k-crumb, .k-dsec, .k-badge, .k-grade */
.k-method { max-width: 880px; }
.k-method h1 { margin: 12px 0 16px; font-family: var(--k-font-display); font-weight: 400; font-size: 40px; line-height: 1.3; }
.k-method p, .k-method li { font-size: 16px; line-height: 1.85; max-width: 40em; }
.k-method p { margin: 0 0 18px; }
.k-method__lead { color: var(--k-ink-2); padding-bottom: 32px; }
.k-method ul { margin: 0 0 18px; padding-left: 1.2em; }
.k-method .k-dsec a { color: var(--k-ink-2); } /* scoped: .k-crumb a keeps ink-3 */
.k-method__dl { margin: 0 0 18px; max-width: 46em; }
.k-method__dl > div { display: grid; grid-template-columns: 13.5em minmax(0, 1fr); gap: 4px 20px; align-items: baseline; padding: 10px 0; border-top: 1px solid var(--k-rule); }
.k-method__dl dd { margin: 0; font-size: 15px; line-height: 1.7; color: var(--k-ink-2); }
.k-method__ja { display: block; font-family: var(--k-font-ja); font-size: 13px; color: var(--k-ink-3); }
.k-pending a { color: var(--k-ink-2); }
/* in @media (max-width: 767px): */
  .k-method h1 { font-size: 26px; }
  .k-method__dl > div { grid-template-columns: minmax(0, 1fr); }
```
First column must fit the widest badge (異常低價 · 先確認狀況) on one line at 1280; adjust the em value if not.

**Rationale:** Page body built from the app's existing canvas-derived parts (no new component or visual language); copy derived from the code paths it describes (Context § Current State), so each sentence is checkable; numbers interpolated so an F1 / F6 answer flows through.
**Verification:** local serve (`go run ./cmd/kura serve -db kura.db -addr 127.0.0.1:<port>`): footer → each of six slugs; `/methodology` shows 3 / 5 / 50 / 14 and 3 件 values, BRAND OFF ladder (5 grades, ja lines), six badges, 目前的記錄還不到 14 天 (worktree DB: one sweep), `Kura 自 2026/10/13 起記錄`; direct load + reload of `/methodology` and `/privacy`; `/nope` still not-found; `grep -nE "[0-9]+ ?%|[0-9]+ ?件|[0-9]+ ?天" internal/web/static/js/pages.js` → only interpolations (no literal threshold).
**If this fails:** revert `pages.js` (delete), `app.js`, `card.js`, `kura.css` edits; detail link returns to not-found (pre-plan state).

---

### Step 3: Detail 行情估價 prose from `ov.method` [PATCHED] → 2610131121-kura-methodology-footer-patch.md

**Objective:** S2, S5 (detail). **Confidence:** High. **Depends on:** Step 1.
**Do-Projex:** Encouraged

**Files:** `internal/web/static/js/detail.js`

**Changes:**
```js
// Before:
import { api, esc, twd, jpy, ymd, md } from "./util.js";
    ${sold ? "" : fairHTML(it, v, name)}
function fairHTML(it, v, name) {
    body = `<p>同型號的比較樣本不足 3 件，暫不估計行情。</p><dl>${how}</dl>`;
      : `同一家店、同樣是 ${g} 的 ${n} 不到 3 件，所以把同型號各等級的 ${f.comps} 件換算到 ${g} 來比較，估計行情是 <strong>${twd(f.fair_twd)}</strong>。`;
    if (v.state === "anomaly") s += "價格比同型號同成色的行情低一半以上，通常代表有標題沒寫的瑕疵、缺件或版本差異，購買前請務必確認商品頁的照片與備註。";

// After:
import { api, esc, twd, jpy, ymd, md, pct } from "./util.js";
    ${sold ? "" : fairHTML(it, v, name, ov.method)}
// m: the radar's thresholds from the overview, so this prose and 行情怎麼算 state one value.
function fairHTML(it, v, name, m) {
    body = `<p>同型號的比較樣本不足 ${m.min_comps} 件，暫不估計行情。</p><dl>${how}</dl>`;
      : `同一家店、同樣是 ${g} 的 ${n} 不到 ${m.min_comps} 件，所以把同型號各等級的 ${f.comps} 件換算到 ${g} 來比較，估計行情是 <strong>${twd(f.fair_twd)}</strong>。`;
    if (v.state === "anomaly") s += `價格比同型號同成色的行情低超過 ${pct(m.anomaly_pct)}%，通常代表有標題沒寫的瑕疵、缺件或版本差異，購買前請務必確認商品頁的照片與備註。`;
```
**Rationale:** The 行情估價 section and the page its link opens state the same thresholds from one source.
**Verification:** a detail page with a fair value renders unchanged except numbers now from `ov.method` (same values); a no-fair listing reads 「…不足 3 件…」; an anomaly listing (`/api/listings?sort=deal` → any `fair.anomaly: true`, or skip if none on DB — the string is checked by reading the code path) reads 「低超過 50%」; 行情怎麼算 link → `/methodology`; `grep -n "3 件\|一半" detail.js` → no literal left.
**If this fails:** `git -C <work-root> checkout -- internal/web/static/js/detail.js`.

---

### Step 4: End-to-end verification [PATCHED] → 2610131121-kura-methodology-footer-patch.md

**Objective:** S1–S8 together. **Confidence:** High. **Depends on:** Steps 1–3.
**Verify-Projex:** Encouraged

**Files:** none (evidence in the patch / walkthrough).

**Changes:** none. Run:
1. `go vet ./...`; `go test ./...`.
2. Serve worktree DB (no crawl flags), browser at 1280 and 375, light and dark (`?theme=dark` on the first load): footer → 關於 Kura / 行情怎麼算 / 常見問題 / 聯絡我們 / 隱私權政策 / 服務條款; Home → card → detail → 行情估價 「行情怎麼算」; Back / Forward; reload on each slug.
3. Per view: document title; no horizontal scroll at 375 (`document.documentElement.scrollWidth <= innerWidth`); badges legend and grade ladder rows legible, first column not wrapping at 1280, stacking at 375; focus ring on crumb + 店家的等級說明 + pending link; 0 console errors.
4. Content audit against Context § Current State: each sentence's claim matches the code path (fair method, confidence rules, band, IsDeal, anomaly, sold gate, watched rule, median label, member mock facts); mocked list = def's four; no N–C rank / factor, no data-age claim, no contact/legal text.
5. Source Hygiene grep over changed sources: no `\.projex|projex|[0-9]{10}-` IDs in comments; no `// Step`; present tense.

**Verification:** every S-criterion row in § Acceptance Criteria Validation ✓.
**If this fails:** fix in the owning step's files; re-run 1–5.

---

## Verification Plan

### Automated Checks
- [x] `go vet ./...`
- [x] `go test ./...` (radar, listing, web incl. `TestStaticRoutes` + `TestOverviewMethod`, advisor incl. `TestPromptStatesThresholds`)

### Manual Verification
- [x] Browser matrix 1280 / 375 × light / dark over `/methodology` + five pending slugs + detail link (Step 4)
- [x] Content audit of every sentence against code (Step 4.4)

### Acceptance Criteria Validation
| Criterion | How to Verify | Expected Result |
|-----------|---------------|-----------------|
| S1 Routes | click each footer link; reload; Back; `/nope` | 6 pages render with titles; not-found only for `/nope` |
| S2 Detail link | detail 行情估價 → 行情怎麼算 | `/methodology` in-app, no reload |
| S3 Content | read `/methodology` | crumb, h1, lead, 5 sections w/ ids; BRAND OFF ladder verbatim; 6 badges |
| S4 Disclosure | `#demo` vs def § Boundaries Mocked | exactly sign-in/member, watchlist, 通知, 大家都在找; advisor line present |
| S5 One source | `TestOverviewMethod`, `TestPromptStatesThresholds`; grep `pages.js` / `detail.js` for literal thresholds | tests pass; no literals; page shows 3/5/50/14, 3 件 |
| S6 Not stated | read page; grep `pages.js` for `rank`, 更新, 最後 | absent; `DealMinPct` 0.05 + `AnomalyPct` 0.5 unchanged |
| S7 Look | matrix (Step 4.2–4.3) | no h-scroll at 375, focus visible, 0 console errors |
| S8 Build | `go vet`, `go test`, hygiene grep | green / clean |

---

## Rollback Plan

1. Revert the patch commit(s) on the sprint branch: `git -C <work-root> revert <sha>` (no history rewrite on the sprint branch).
2. Footer slugs and the detail link fall back to the not-found view (pre-plan state); `/api/overview` loses `method` / `schemes` (no other consumer).

---

## Notes

### Split Decision
`No split — single scope, within size budget` (4 steps, ~400 lines; server + client changes are one vertical slice: the page cannot state server constants that are not exposed).

### Risks
- **Copy accuracy drift:** prose describes confidence / comparables rules qualitatively; a later radar change (e.g. confidence rules) could outdate a sentence without a test failing. Mitigation: numbers interpolated; qualitative claims audited in Step 4.4; radar constant comment names the page.
- **Badge column width:** CJK badge widths vary by font fallback; 13.5em may wrap on some systems. Mitigation: Step 4.3 check; tune em.
- **F19 still open:** the page is the def's disclosure channel, but whether unlabelled mocks are acceptable at all is a pending user confirmation (def L2). Page does not settle it.
- **Pending pages look like content gaps** to a reviewer. Accepted: honest state; content needs the human (sprint Deferred).
- **R1 carry-over** (genuinely new listing inside epoch+24 h sorts last) unaffected; page's 新上架 wording ("在店家最新的一件") stays true either way.

### Open Questions
- (none blocking execution — human-owned questions F1, F6, def L2 F19 confirmation, L3, L4 are out of scope and not decided here)
