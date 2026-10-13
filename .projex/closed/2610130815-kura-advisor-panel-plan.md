# Kura advisor panel (顧問)

> **Sprint:** 2610130524-kura-interactive-mock-sprint.md — iteration 4
> **Status:** Complete
> **Author:** Claude Opus 5.5 (planner, sprint subagent)
> **Parent:** 2610130524-kura-interactive-mock-sprint.md
> **Source:** Sprint derive, iteration 4 (sprint nav § Current Position + Revision Log, commit `bbd41fb`)
> **Related Projex:** 2610130524-kura-interactive-mock-sprint.md | 2610130449-kura-mock-def.md (one source per stat, advisor demo label exception, Canvas gaps 「Advisor: error state, spend-cap-reached state」, L3 advisor helpers/badge/cap open) | 2610130420-kura-roadmap-nav.md (Phase 3) | 2610130748-kura-model-page-audit.md (carry-over H2) | 2610130722-kura-model-page-plan.md (`radar.MinSoldHistory`, `soldKnown`) | 2610130624-kura-listing-detail-plan.md (`dist.js` by_grade source, single photo) | 2610130535-kura-canvas-foundation-home-plan.md (shell, `card.js`)
> Partial Execution: Steps 1–4 (all objectives) completed via 2610130822-kura-advisor-panel-patch.md — commits `5e65a9a`, `912a354`; nothing remains open in this plan
> **Worktree:** Yes — existing sprint worktree `kura/.projexwt/2610130524-kura-interactive-mock`, branch `projex/sprint/2610130524-kura-interactive-mock`; execute via patch-projex plan-execution mode, no new branch

---

## Summary

Build the canvas 顧問 (pinned 1791839497-46ec; markup `project/Main.dc.html` 顧問 block; artboards `AdvisorOpen` 1280 claude open, `AdvisorFull` 1280 claude full + `chat: conversation`, `AdvisorMobile` 375 demo open) in the new shell on every route: 「蔵 找包顧問」 dock → centered panel (modal, backdrop) ⇄ full page → closed. Panel talks to the existing `POST /api/chat` SSE (`Mock` here — no API key; `Claude` shape covered by fake-API tests + a browser replay). Conversation = user bubble → 「助手的進度」 list from **real** tool events → streamed answer with `[[id]]` chips → recommendation figures (photo, grade, TWD, badge, same-grade position bar from server `by_grade`, 追蹤 / 看詳細 / 商品頁). Closes carry-over **H2**: advisor tools stop stating sold count / sell-through / days-to-sell until `radar.Engine.SoldStatsReady()` — the gate the model page uses.

**Scope:** `internal/advisor` (sold-stat gate, prompt, test) + `internal/web` (shell markup, CSS, new `advisor.js`, router wiring, static-route test). One `.projex` scope (repo root `.projex/`).
**Estimated Changes:** 8 files — `tools.go`, `prompt.go`, `index.html`, `kura.css`, `app.js`, `server_test.go` edited; `tools_test.go`, `js/advisor.js` added.

---

## Objective

### Problem / Gap / Need
Advisor exists only in `/legacy/`; new shell has no 顧問 entry (largest unbuilt canvas flow not blocked by an open question — sprint nav blocking map). Advisor tools leak ungated sold stats: `model_price_stats` (`sold_observed`, `sell_through`, `median_days_to_sell`, summary 「已售出 N 件」), `get_listing` (`model_sold_observed`), `market_overview` (`sold_observed` top + per model) — while the model page shows 資料不足 (`MinSoldHistory` 14 d gate) → violates def 「One source per stat」 (iteration 3 audit H2). The summary string reaches the UI as a progress line, so the conflict is visible, not only model-side.

### Success Criteria
- [x] **S1 Dock + states:** every shell route (`/`, `/listing/…`, `/model/…`, not-found) shows the canvas dock (fixed, bottom-centre; seal 蔵 + 「找包顧問」). Click → floating panel (`role` dialog, name 「Kura 顧問」, backdrop); 放大成整頁 → full page (inner column 720 px); 縮小成視窗 → floating; 關閉顧問 / Esc → closed, focus back on the dock. Dock hidden while panel open. Geometry: desktop 680 × 640 bottom 24 centred; ≤ 767 px left/right/bottom 8, height 600 (both capped to viewport).
- [x] **S2 Header:** seal, h2 「Kura 顧問」, 示範模式 pill (title 「未設定 API 金鑰，顧問以範例資料回答」) iff `/api/overview` `advisor === "mock"`; max/shrink icon swaps with label.
- [x] **S3 Intro:** canvas intro paragraph + 4 `KURA_PROMPTS` chips; a chip sends its text; chips gone after first send (canvas `chat: intro` → `conversation`).
- [x] **S4 Composer:** textarea (label 問顧問 sr-only, placeholder 「想找什麼樣的包？」), 送出 button, hint 「Enter 送出 · Shift+Enter 換行」. Enter sends (not during IME composition), Shift+Enter newline, blank ignored, one turn in flight (送出 disabled while streaming).
- [x] **S5 Turn:** user bubble (right, sunk) → 「助手的進度」 `<ol>` (only when ≥ 1 tool event): per tool call a line `{helper} {TOOL_ZH}⋯` (A1/A2) that becomes `{helper} {server summary} ✓` on its result → answer streams with blinking cursor (static under reduced motion), simple markdown (paragraphs, `- ` bullets, `**bold**`), `[[id]]` → chip `{model} · {code}` linking `/listing/{id}` (A6). No 風格 line.
- [x] **S6 Figures:** after the stream ends, each referenced **active** listing (first-reference order, max 4) gets a figure after the block that first references it (A3): photo mat (single photo / 店家照片準備中, A4), brand, model link → `/listing/{id}`, grade tag + `desc_zh`, `NT$` price + card badge (`valuation`/`badgeHTML`), position bar 「在同款 {grade} 的 {n} 件裡的位置」 min / 中位數 / max from `/api/models/{brand}/{group_key}` `by_grade` for {source, code} (A5; omitted when n < 2 or min = max or fetch fails), 追蹤 toggle (shared tracked set), 看詳細, 「{shop} 商品頁」 (new tab).
- [x] **S7 Navigation:** any in-app link in the panel navigates via the router and closes the panel; reopening shows the same conversation (A7). External shop link leaves panel open.
- [x] **S8 Errors (functional fallback, A9):** SSE `error` event → its text as a plain paragraph after the partial answer; network / non-200 → shell copy 「暫時無法載入資料，請稍後再試。」; composer re-enabled. No new visual state invented.
- [x] **S9 One source per stat:** figure numbers = server fields (`price_twd`, `fair` via card `valuation`, `by_grade` `count/min_twd/median_twd/max_twd`) — no client math beyond bar positions. Advisor tools omit `sold_observed`, `sell_through`, `median_days_to_sell`, `model_sold_observed` and per-model `sold_observed` and add `"sold_stats": "insufficient_history"` while `SoldStatsReady()` is false; `model_price_stats` summary drops 「已售出 N 件」 then. System prompt + tool description tell the model the fields are unknown, not zero.
- [x] **S10 Responsive/themes/a11y:** 1280 + 375 × light/dark — tokens only (backdrop = canvas literal); no horizontal page scroll at 375; visible focus ring; targets ≥ 24 px; modal keeps focus inside panel.
- [x] **S11** `go vet ./...` + `go test ./...` green (new `tools_test.go`, extended `TestStaticRoutes`); no projex references in source comments; `/legacy/` advisor still works.

### Out of Scope
- Designed error + spend-cap states (def § Canvas gaps) → Deferred (Notes) — only the A9 functional fallback ships
- Spend cap / hosted demo-fallback policy (def L3 open; hosting = roadmap Phase 6, out of sprint)
- Reviewer-facing demo-badge copy + `Mock` preface wording (def L3 open) — canvas copy verbatim, `mock.go` text unchanged (A10)
- Scripted helper lines / 風格 helper (def L3 open; 風格 grounds on mocked 大家都在找) — not rendered (A1)
- 5-photo gallery (def L3 gallery open; provider gives one photo)
- HANDOFF §4.A header status line (advisor mode) — remains a def canvas gap
- Real Claude calls (no key); conversation persistence across reloads; advisor search/ranking changes; F1 `only_deals` threshold (5 %) — untouched
- Browse, member, footer pages; `/legacy/` UI

---

## Context

### Current State
- **Backend:** `POST /api/chat` (`internal/web/server.go` `chat`) → JSON `{session, message}` (16 KB cap; 400 on blank) → SSE `data: {Event}\n\n`, types `text | tool_call | tool_result | listings | error | done` (`internal/advisor/agent.go` `Event`). `Claude.Chat`: one in-flight turn per session (`TryLock` → zh error), tool loop ≤ 8 rounds, emits `tool_call` → `tool_result` (Summary, `is_error`) → `listings` (radar `Item`s) per tool, text deltas; zh-TW error strings (`describeAPIError`, refusal, max-tokens). `Mock.Chat` (no key): one `search_listings` → `tool_call`, `tool_result` 「找到 N 件符合條件的商品，回傳 M 件」, one `text` (preface 「（示範模式：未設定 ANTHROPIC_API_KEY…）」 + `- …[[id]]` bullets + FX line), then `listings`. Mode: `mock` | `claude:{model}` in `/api/overview` `advisor`. Wiring `cmd/kura/main.go` (Mock unless `ANTHROPIC_API_KEY` / `KURA_ADVISOR=claude`).
- **Tools (H2):** `internal/advisor/tools.go` `modelStats` returns `sold_observed`, `sell_through`, `median_days_to_sell` ungated, summary `%s：在售 %d 件、已售出 %d 件`; `getListing` `model_sold_observed`; `overview` `sold_observed` + per-model `Sold` (`model` struct). `radar.Engine.SoldStatsReady()` (`internal/radar/radar.go`) = epoch→latest span ≥ `MinSoldHistory`; model page reads it via overview `sold_stats_ready` (`model.js` `soldKnown`). Worktree DB: one sweep → not ready.
- **Shell:** `index.html` header/main/footer; footer already pads bottom 104 / 96 px (≤ 767) for the dock. `app.js` router (3 routes; document click handler intercepts `a[href^='/']` except `/legacy|dev|api`; `overview()` memoised promise). `card.js` exports `valuation`, `badgeHTML`, `isTracked`, `HEART`, `OUT_ICON`, `PENDING_ICON`; document `[data-track]` listener toggles an in-memory set and relabels buttons via `[data-track-text]`. `util.js` `api` (throws with `status`), `esc`, `num`, `twd`. `dist.js` matches `by_grade` by `{source, grade.code}`. No `@keyframes` in `kura.css`; global `:focus-visible` ring; mobile breakpoint 767.
- **Legacy client** (`static/legacy/app.js` advisor block): SSE reader over `fetch` + `getReader`, `renderMD` (escape first → bold, bullets, paragraphs, `[[([a-z]+:[\w-]+)]]` chips), `TOOL_ZH` map — pattern to port, file untouched.
- **Canvas** (read 2026-10-13, served 1791845608-ba84, pin governs; artboards identical): `Main.dc.html` lines ~499–623 markup, `renderVals` (`adv` closed|open|full, `toggleMax`, `winStyle`, `KURA_PROMPTS`, gallery state). Conversation sample: user bubble 「預算 8 萬，通勤可放 A4 的 LV 托特包」, `<ol aria-label="助手的進度">` lines 行情 ✓ / 成色 ✓ / 風格 ⋯ (ink, no ✓), answer + inline chip 「ネヴァーフルMM · M40156」, `<figure aria-label="推薦：…">`, blinking cursor (`k-blink`).

### Key Files

> Quick reference — detailed changes in Implementation.

| File | Role | Change Summary |
|------|------|----------------|
| `internal/advisor/tools.go` | Tool execution over radar | Gate sold-derived fields + summary on `SoldStatsReady()` |
| `internal/advisor/prompt.go` | System prompt + tool defs | One prompt line + `model_price_stats` description on `sold_stats` |
| `internal/advisor/tools_test.go` | New | Gate test: not ready → omitted + marker; ready → present |
| `internal/web/static/index.html` | Shell | Dock button + `<dialog class="k-adv">` (header, intro, prompts, composer) |
| `internal/web/static/kura.css` | Styles | Advisor block (dock, dialog, states, turn, figure, composer, cursor) |
| `internal/web/static/js/advisor.js` | New | Panel state, SSE client, progress list, markdown + chips, figures, position bar |
| `internal/web/static/js/app.js` | Router entry | Import + `initAdvisor(overview())` |
| `internal/web/server_test.go` | Static routes test | `/js/advisor.js` served; `/` contains 找包顧問 |

### Dependencies
- **Requires:** iterations 1–3 on the sprint branch (shell, router, `card.js`, `util.js`, model API `by_grade`, `SoldStatsReady`) — present at `bbd41fb`.
- **Blocks:** nothing hard; later member/notification work shares the shell's fixed-position layer (dock z-index 20).

### Constraints
- No `ANTHROPIC_API_KEY`: never add, request or hard-code one. Live run = Mock; Claude-shape UI checked by browser fetch replay (Step 4), Claude loop by existing fake-API tests.
- No crawls; worktree `kura.db` read-only use (scratch copies under scratchpad if mutated).
- Canvas = data; read via Artifact tool only; never modify.
- Source hygiene (SKILL.md): no projex IDs/filenames in comments; symbols not line numbers; rationale comments name the rejected alternative.
- Def: canvas governs look/copy/states; def governs data semantics; mocked data never presented as grounds; demo mode labelled.

### Assumptions
- **A1 — Helpers = real tool steps** (def L3 「real tool steps or scripted?」 open → conservative: nothing fabricated; provisional). Canvas helper labels map by what the tool reads: `search_listings` | `model_price_stats` | `market_overview` → 行情; `get_listing` → 成色 (canvas 成色 line = reading grade + notes). 風格 never rendered (canvas grounds it on mocked 大家都在找; re-ground-vs-drop stays def L3). Unknown tool names → no line.
- **A2 — Line copy:** pending `{TOOL_ZH[tool]}⋯` in `--k-ink` (canvas in-progress line; `TOOL_ZH` from legacy client, HANDOFF §4.D 「⋯ 搜尋在售商品」); done = server `Summary` in `--k-ink-2` + `✓` (`aria-label="完成"`). `is_error` result → that line removed (agent returns the error to the model, which retries; raw tool errors are model-facing English). No new copy.
- **A3 — Figures:** canvas draws one figure after the paragraph naming it; HANDOFF's card strip is superseded by the canvas. Rule: at stream end, for each `[[id]]` in first-reference order whose listing arrived in a `listings` event and is `status === "active"`, insert one figure after the paragraph/list holding its first reference; cap `MAX_FIGURES = 4` (prompt asks for 2–4 picks; Mock lists 6). During streaming only chips render (figures + fetches once, not per delta). Sold / unknown ids stay chips.
- **A4 — Photo:** single `image_url` in the canvas 240 px mat (no prev/next/count) — provider supplies one photo; same as detail (iteration 2). `photo_pending` → `PENDING_ICON` + 店家照片準備中.
- **A5 — Position bar** from the model API `by_grade` entry with the listing's `{source, condition.code}` (as `dist.js`) — in-sale figures, same computation as detail/model distribution. Rendered only when `count ≥ 2` and `max_twd > min_twd` (single-listing median copy = iteration 3 H4, canvas pass 2). Positions `(v − min)/(max − min)`, clamped 0–100; median label clamped to 25–75 % so it clears the min/max labels; ≤ 767 px the median label sits on its own row above the track. Model responses cached per `{brand, group_key}` for the page lifetime.
- **A6 — Chip** (canvas `<button>` with no target) → `<a href="/listing/{id}">` styled as the canvas chip, text `{model_name || title} · {model_code}` (code part only if present); unknown id → 「查看商品」 (legacy copy). Rationale: the figure carries the detail link too, but chips also appear for listings without a figure.
- **A7 — Panel ↔ navigation:** in-app link click inside the panel (plain left click) closes the panel before the router renders, so the page is visible; DOM conversation persists for the page lifetime. Session id = `crypto.randomUUID()` per page load (not `sessionStorage` like legacy): after a reload the UI is empty, and a reused id would let `Claude` answer from history the user can no longer see.
- **A8 — `<dialog>` + `showModal()`** for floating and full (canvas: `section role=dialog` + backdrop div): native modal gives focus containment, Esc, inert page; `aria-label="Kura 顧問"`. Backdrop via `::backdrop` with the canvas literal `rgba(30,25,21,.18)` (light-only artboards; dark = def gap 「token-driven」); transparent in full mode. Backdrop click does nothing (canvas has no handler).
- **A9 — Error fallback, not a design:** canvas has no error/spend-cap state (def § Canvas gaps). Shipped: server `error` text (already user-facing zh-TW from `agent.go`) or shell copy 「暫時無法載入資料，請稍後再試。」 as a plain `--k-ink-2` paragraph. Designed states → Deferred (blocker: canvas pass 2; spend cap also def L3 hosted-mode policy + Phase 6).
- **A10 — Demo labelling** verbatim canvas (pill + title). `Mock` preface text untouched. Both developer-facing per def L3 → reviewer wording open, provisional.
- **A11 — Sold-stat gate (H2):** one predicate `t.Radar.SoldStatsReady()`; not ready → fields omitted (not 0, not null) + `"sold_stats": "insufficient_history"`; ready → today's fields unchanged. `market_overview` uses `Overview().SoldStatsReady` (same lock snapshot as its counts). Ordering of `Models` (active + sold) unchanged — ordering states no number.
- **A12 — Chat UI scroll:** body auto-scrolls to bottom only if it was within 120 px of the bottom before the update (reading earlier text is not yanked).
- `crypto.randomUUID` available (127.0.0.1 is a secure context; hosting will be https).

### Impact Analysis
- **Direct:** advisor tool JSON + `model_price_stats` summary (Claude + `/legacy/` progress lines); every shell page gains the dock + dialog.
- **Adjacent:** `card.js` track listener now also serves figure buttons (`data-track-text="追蹤"`); router click handler runs after the panel's close handler; model API called by the panel (cached).
- **Downstream:** none outside the repo. `Mock` path unchanged (never calls gated tools).

---

## Implementation

### Overview
Step 1 fixes the data contract (Go, tested). Step 2 lays the static markup + CSS for all canvas states. Step 3 adds the behaviour module and wires it. Step 4 verifies end-to-end in the browser (Mock live + Claude-shaped replay).

### Step 1: Gate advisor sold stats on `SoldStatsReady` (H2) [PATCHED → 2610130822-kura-advisor-panel-patch.md]

**Objective:** S9 advisor half. **Confidence:** High. **Depends on:** None.
**Do-Projex:** Encouraged
**Verify-Projex:** Encouraged

**Files:** `internal/advisor/tools.go`, `internal/advisor/prompt.go`, `internal/advisor/tools_test.go` (new)

**Changes:**

```go
// tools.go — new, near ErrInvalidInput:

// soldUnknown marks results whose sold-derived figures (sold count,
// sell-through, days to sell) are withheld because crawl history is shorter
// than radar.MinSoldHistory. Sending them as 0 would read as "never sells",
// and the model page shows 資料不足 for the same figures from the same gate
// (radar.Engine.SoldStatsReady), so the advisor must not state a number there.
const soldUnknown = "insufficient_history"
```

```go
// modelStats — Before:
	res := map[string]any{
		"brand": ..., "group_key": stats.GroupKey,
		"active_listings": stats.Active, "sold_observed": stats.Sold,
		"sell_through": round2(stats.SellThrough), "median_days_to_sell": stats.MedianDaysToSell,
		"median_twd": ..., "by_condition_grade": grades, "cheapest_active": cheapest, "fx": t.fxNote(),
	}
	return ToolResult{JSON: marshal(res), Listings: cards,
		Summary: fmt.Sprintf("%s：在售 %d 件、已售出 %d 件", stats.ModelName, stats.Active, stats.Sold)}, nil

// After:
	res := map[string]any{
		"brand": ..., "group_key": stats.GroupKey, "active_listings": stats.Active,
		"median_twd": ..., "by_condition_grade": grades, "cheapest_active": cheapest, "fx": t.fxNote(),
	}
	summary := fmt.Sprintf("%s：在售 %d 件", stats.ModelName, stats.Active)
	if t.Radar.SoldStatsReady() {
		res["sold_observed"], res["sell_through"], res["median_days_to_sell"] =
			stats.Sold, round2(stats.SellThrough), stats.MedianDaysToSell
		summary += fmt.Sprintf("、已售出 %d 件", stats.Sold)
	} else {
		res["sold_stats"] = soldUnknown
	}
	return ToolResult{JSON: marshal(res), Listings: cards, Summary: summary}, nil
```

```go
// getListing — Before:
	if stats != nil {
		res["model_active_listings"] = stats.Active
		res["model_sold_observed"] = stats.Sold
	}
// After:
	if stats != nil {
		res["model_active_listings"] = stats.Active
		if t.Radar.SoldStatsReady() {
			res["model_sold_observed"] = stats.Sold
		} else {
			res["sold_stats"] = soldUnknown
		}
	}
```

```go
// overview — Before: model.Sold int `json:"sold_observed"`, positional literal, o := t.Radar.Overview() after the loop,
// res has "sold_observed": o.Sold.
// After:
	o := t.Radar.Overview()
	type model struct {
		Model     string `json:"model"`
		Brand     string `json:"brand"`
		GroupKey  string `json:"group_key"`
		Category  string `json:"category"`
		Active    int    `json:"active"`
		Sold      *int   `json:"sold_observed,omitempty"` // nil until sold stats are ready
		MedianTWD int64  `json:"median_twd"`
		MinTWD    int64  `json:"min_twd"`
	}
	var models []model
	for _, m := range t.Radar.Models(brand, 15) {
		md := model{Model: m.ModelName, Brand: m.Brand, GroupKey: m.GroupKey, Category: listing.CategoryZH(m.Category),
			Active: m.Active, MedianTWD: t.FX.TWD(m.MedianJPY), MinTWD: t.FX.TWD(m.MinJPY)}
		if o.SoldStatsReady {
			n := m.Sold
			md.Sold = &n
		}
		models = append(models, md)
	}
	res := map[string]any{"active_listings": o.Active, "by_brand": o.ByBrand,
		"tracking_since": o.Since.Format(time.DateOnly), "top_models": models, "fx": t.fxNote()}
	if o.SoldStatsReady {
		res["sold_observed"] = o.Sold
	} else {
		res["sold_stats"] = soldUnknown
	}
```

```go
// prompt.go — systemPrompt, 呈現方式 list, after the 「行情價」 bullet, add:
- 工具結果帶有 sold_stats: insufficient_history 時，代表 Kura 的記錄時間還不夠：已售出件數、售出率與售出天數都未知（不是 0）。
  請說「資料不足」，不要推測，也不要當成賣不掉。

// toolDefs — model_price_stats Description: append
" Sold count, sell-through and days to sell are omitted (sold_stats: insufficient_history) until Kura has enough crawl history; they are then unknown, not zero."
```

```go
// tools_test.go (new, package advisor) — reuses testTools from agent_test.go
func TestSoldStatsGatedUntilHistory(t *testing.T) {
	tools := testTools(t) // 4 HERMES 056289CK listings, all first seen now → history span 0
	run := func(name, in string) (map[string]any, string) {
		t.Helper()
		res, err := tools.Run(name, in)
		if err != nil {
			t.Fatal(err)
		}
		var m map[string]any
		if err := json.Unmarshal([]byte(res.JSON), &m); err != nil {
			t.Fatal(err)
		}
		return m, res.Summary
	}
	// Not ready: no sold figure anywhere, marker present, summary silent on sales.
	// assert for model_price_stats {"brand":"HERMES","model_code":"056289CK"}:
	//   keys sold_observed, sell_through, median_days_to_sell absent; sold_stats == "insufficient_history";
	//   summary lacks "已售出"
	// get_listing {"id":"brandoff:1000"}: model_sold_observed absent; sold_stats marker
	// market_overview {}: sold_observed absent; top_models[0] has no sold_observed; sold_stats marker

	// Ready: a listing first seen 15 days ago, gone now → span ≥ MinSoldHistory, 1 sold.
	ctx := context.Background()
	old := listing.Listing{Source: "brandoff", SourceID: "2000", URL: "https://example.test", Brand: "HERMES",
		Title: "エルメス ピコタンロックPM", ModelName: "ピコタンロックPM", ModelCode: "056289CK",
		Category: "handbag", Rank: listing.RankA, PriceJPY: 310000}
	// RecordSeen(ctx, "HERMES", old, time.Now().Add(-15*24*time.Hour)); MarkGone(ctx, "brandoff:2000", time.Now());
	// tools.Radar.Refresh(ctx) — each error → t.Fatal
	// assert model_price_stats: sold_observed == 1 (float64), sell_through present, no sold_stats; summary contains "已售出 1 件"
	// get_listing brandoff:1000: model_sold_observed == 1; market_overview: sold_observed == 1, top_models[0].sold_observed == 1
}
```

**Rationale:** Omit + marker (not `null`/0): a missing key cannot be read as a value; the marker plus prompt line lets the model say 資料不足 like the page. One predicate on the engine — no second history rule in the advisor.

**Verification:** `go test ./internal/advisor/...` (new + existing agent tests green); `go vet ./...`.

**If this fails:** `git checkout -- internal/advisor` (uncommitted) or revert the step's commit; UI steps do not depend on it at runtime.

---

### Step 2: Shell markup + advisor CSS [PATCHED → 2610130822-kura-advisor-panel-patch.md]

**Objective:** S1–S4, S10 static structure/looks. **Confidence:** Medium (native `<dialog>` positioning in top layer). **Depends on:** None.

**Files:** `internal/web/static/index.html`, `internal/web/static/kura.css`

**Changes — `index.html`**, after `</footer>`, before `</body>` (copy verbatim from canvas; SVG paths from canvas):

```html
<button type="button" class="k-adv-dock" data-adv-open aria-haspopup="dialog">
  <span class="k-adv__seal" aria-hidden="true">蔵</span><span class="k-adv-dock__label">找包顧問</span>
</button>
<dialog class="k-adv" aria-label="Kura 顧問">
  <header class="k-adv__head"><div class="k-adv__in k-adv__bar">
    <span class="k-adv__seal k-adv__seal--lg" aria-hidden="true">蔵</span>
    <div class="k-adv__title"><h2>Kura 顧問</h2><span class="k-adv__demo" title="未設定 API 金鑰，顧問以範例資料回答" hidden>示範模式</span></div>
    <span class="k-adv__sp"></span>
    <button type="button" class="k-adv__icon" data-adv-max aria-label="放大成整頁" title="放大成整頁">
      <svg class="k-adv__expand" …canvas expand path "M9.5 2.5h4v4M13.5 2.5 9 7M6.5 13.5h-4v-4M2.5 13.5 7 9"…/>
      <svg class="k-adv__shrink" …canvas shrink path "M13.5 6.5h-4v-4M9.5 6.5 14 2M2.5 9.5h4v4M6.5 9.5 2 14"…/>
    </button>
    <button type="button" class="k-adv__icon" data-adv-close aria-label="關閉顧問" title="關閉"><svg …"M3.5 3.5l9 9M12.5 3.5l-9 9"…/></button>
  </div></header>
  <div class="k-adv__body"><div class="k-adv__in k-adv__log">
    <p class="k-adv__intro">你好。跟我說說你想要的包：預算、什麼場合用、喜歡哪個牌子都可以。我會請助手們去店裡翻一翻，挑幾件給你，說清楚好在哪、要注意什麼。</p>
    <div class="k-adv__prompts">
      <button type="button" data-adv-prompt>預算 8 萬，通勤可放 A4 的 LV 托特包</button>
      <button type="button" data-adv-prompt>哪些愛馬仕包明顯低於行情？</button>
      <button type="button" data-adv-prompt>香奈兒菱格紋鏈條包的行情</button>
      <button type="button" data-adv-prompt>5 萬以內的入門經典款</button>
    </div>
  </div></div>
  <form class="k-adv__form"><div class="k-adv__in">
    <div class="k-adv__row">
      <label class="k-adv__field"><span class="k-sr">問顧問</span><textarea name="q" rows="2" placeholder="想找什麼樣的包？" autofocus></textarea></label>
      <button type="submit" class="k-adv__send" aria-label="送出"><svg …"M8 13V3M3.5 7.5 8 3l4.5 4.5"…/></button>
    </div>
    <p class="k-adv__hint">Enter 送出 · Shift+Enter 換行</p>
  </div></form>
</dialog>
```
All SVGs `aria-hidden="true"`, `fill="none" stroke="currentColor"`, sizes/stroke widths as canvas (expand/shrink 18 / 1.4, close 16 / 1.4, send 18 / 1.5).

**Changes — `kura.css`**, new block before the `@media (max-width: 1023px)` block, values from canvas `winStyle` / inline styles:

| Selector | Key declarations |
|---|---|
| `.k-adv-dock` | `position: fixed; left: 50%; bottom: 28px; transform: translateX(-50%); z-index: 20; height: 56px; padding: 0 10px; display: flex; align-items: center; gap: 12px; white-space: nowrap; border: 1px solid var(--k-brass); border-radius: 999px; background: var(--k-surface); color: var(--k-ink); box-shadow: var(--k-shadow); font: inherit; cursor: pointer` |
| `body:has(.k-adv[open]) .k-adv-dock` | `visibility: hidden` (canvas: dock absent while open) |
| `.k-adv__seal` / `--lg` | 38 px / 40 px circle, `background: var(--k-accent); color: var(--k-on-accent); font-family: var(--k-font-ja)`; 18 / 19 px; flex centre; `flex: none` |
| `.k-adv-dock__label` | `font-family: var(--k-font-display); font-size: 16px; padding-right: 12px` |
| `.k-adv` | `position: fixed; inset: auto; left: 50%; bottom: 24px; transform: translateX(-50%); margin: 0; padding: 0; width: min(680px, 100vw - 16px); height: min(640px, 100dvh - 48px); max-width: none; max-height: none; border: 1px solid var(--k-brass); border-radius: 16px; background: var(--k-surface); color: var(--k-ink); font-family: var(--k-font-ui); box-shadow: var(--k-shadow); overflow: hidden` |
| `.k-adv[open]` | `display: flex; flex-direction: column` (closed dialog keeps UA `display: none`) |
| `.k-adv::backdrop` / `.k-adv.is-full::backdrop` | `background: rgba(30,25,21,.18)` / `transparent` |
| `.k-adv.is-full` | `inset: 0; transform: none; width: 100%; height: 100%; border: 0; border-radius: 0` |
| `.k-adv.is-full .k-adv__in` | `max-width: 720px; margin: 0 auto` |
| `.k-adv__head` | `border-bottom: 1px solid var(--k-rule); padding: 14px 12px 14px 20px` |
| `.k-adv__in` | `width: 100%; box-sizing: border-box` |
| `.k-adv__bar` | `display: flex; align-items: center; gap: 12px` |
| `.k-adv__title` | `display: flex; align-items: center; gap: 8px; flex-wrap: wrap; min-width: 0`; `h2`: `margin: 0; font-family: var(--k-font-display); font-weight: 400; font-size: 18px; white-space: nowrap` |
| `.k-adv__demo` | `font-size: 11px; color: var(--k-ink-2); border: 1px dashed var(--k-rule-strong); padding: 0 7px; border-radius: 999px` |
| `.k-adv__sp` | `flex: 1` |
| `.k-adv__icon` | `flex: none; width: 36px; height: 36px; padding: 0; border: 0; background: transparent; color: var(--k-ink-2); display: flex; align-items: center; justify-content: center; cursor: pointer; border-radius: 2px` |
| `.k-adv__shrink`, `.k-adv.is-full .k-adv__expand` / `.k-adv.is-full .k-adv__shrink` | `display: none` / `display: block` |
| `.k-adv__body` | `flex: 1 1 auto; overflow-y: auto; overscroll-behavior: contain; padding: 26px 20px` |
| `.k-adv__log`, `.k-adv__turn` | `display: flex; flex-direction: column; gap: 20px` |
| `.k-adv__intro` | `margin: 0; font-size: 15px; line-height: 1.85` |
| `.k-adv__prompts` / `button` | `display: flex; flex-wrap: wrap; gap: 8px` / `font: inherit; font-size: 14px; line-height: 1.5; color: var(--k-ink); background: transparent; border: 1px solid var(--k-rule-strong); border-radius: 999px; padding: 7px 14px; cursor: pointer; text-align: left` |
| `.k-adv__user` | `align-self: flex-end; max-width: 80%; background: var(--k-paper-sunk); border-radius: 14px 14px 2px 14px; padding: 10px 14px; font-size: 15px; line-height: 1.6; white-space: pre-wrap; overflow-wrap: anywhere` |
| `.k-adv__steps` / `li` / `li.is-done` | `list-style: none; margin: 0; padding: 0 0 0 14px; border-left: 1px solid var(--k-brass); display: flex; flex-direction: column; gap: 10px` / `display: flex; gap: 10px; align-items: baseline; font-size: 14px; color: var(--k-ink)` / `color: var(--k-ink-2)` |
| `.k-adv__who`, `.k-adv__ok` | `font-size: 12px; color: var(--k-ink-3); white-space: nowrap` / `color: var(--k-ink-3)` |
| `.k-adv__answer` | `display: flex; flex-direction: column; gap: 12px; font-size: 15px; line-height: 1.85`; `p` margin 0; `ul` `margin: 0; padding-left: 1.2em` |
| `.k-adv__chip` | `font-family: var(--k-font-ja); font-size: 13px; color: var(--k-accent-ink); background: var(--k-accent-wash); border: 1px solid var(--k-accent); border-radius: 2px; padding: 0 6px; text-decoration: none` |
| `.k-adv__cursor` | `display: inline-block; width: 2px; height: 1.1em; background: var(--k-ink); vertical-align: -3px; margin-left: 2px; animation: k-blink 1s steps(1) infinite`; `@keyframes k-blink { 50% { opacity: 0 } }`; `@media (prefers-reduced-motion: reduce) { .k-adv__cursor { animation: none } }` |
| `.k-adv__err` | `margin: 0; color: var(--k-ink-2)` |
| `.k-adv__rec` | `margin: 4px 0; display: flex; flex-direction: column; gap: 14px` |
| `.k-adv__mat` (+ `img`) | `width: min(240px, 100%); aspect-ratio: 1 / 1; box-sizing: border-box; padding: 10%` + `.k-card__mat` mat/frame/flex declarations; `img` as `.k-card__mat img` |
| `.k-adv__recmeta` | `display: flex; flex-wrap: wrap; justify-content: space-between; align-items: flex-end; gap: 10px 20px` |
| `.k-adv__recid` / `.k-adv__brand` / `.k-adv__name` / `.k-adv__grade` | column gap 5 `min-width: 0` / `font-family: var(--k-font-num); font-size: 12px; letter-spacing: .16em; color: var(--k-ink-3)` / `font-family: var(--k-font-ja); font-size: 19px; font-weight: 600; color: var(--k-ink); text-decoration: none; line-height: 1.3` / `display: flex; gap: 8px; align-items: baseline; font-size: 13px; color: var(--k-ink-2)` |
| `.k-adv__recprice` / `.k-price` | `display: flex; align-items: center; gap: 10px` / `font-size: 20px` |
| `.k-adv__pos` | `display: flex; flex-direction: column; gap: 6px`; `.k-adv__poscap` 12 px `--k-ink-3`; `.k-adv__track` `position: relative; height: 18px` + `::before` rule line (top 8, height 2, `--k-rule-strong`); `.k-adv__med` (abs, top 2, 2 × 14, `margin-left: -1px`, `--k-ink`); `.k-adv__this` (abs, top 3, 12 × 12, `margin-left: -6px`, round, `--k-accent-ink`, `box-shadow: 0 0 0 2px var(--k-surface)`); `.k-adv__ticks` `position: relative; height: 16px; font-size: 11px; color: var(--k-ink-3); font-variant-numeric: tabular-nums` — first span `left: 0`, last `right: 0`, `.k-adv__medlabel` abs `transform: translateX(-50%); color: var(--k-ink-2)` (all spans abs, nowrap) |
| `.k-adv__acts` / `.k-adv__btn` | `display: flex; flex-wrap: wrap; gap: 8px` / `height: 36px; padding: 0 14px; display: inline-flex; align-items: center; gap: 6px; font: inherit; font-size: 13px; color: var(--k-ink); background: transparent; border: 1px solid var(--k-ink-2); border-radius: 999px; text-decoration: none; cursor: pointer`; `svg` 15 px |
| `.k-adv__btn[aria-pressed="true"]` (+ `svg`) | as `.k-dtrack[aria-pressed="true"]` (accent-ink / wash / accent border; heart filled) |
| `.k-adv__btn--shop` | `color: var(--k-on-accent); background: var(--k-accent); border-color: var(--k-accent)` |
| `.k-adv__form` | `border-top: 1px solid var(--k-rule); padding: 12px 12px 10px` |
| `.k-adv__row` / `.k-adv__field` / `textarea` | `display: flex; gap: 8px; align-items: flex-end` / `flex: 1 1 auto; display: flex` / `flex: 1 1 auto; resize: none; font: inherit; font-size: 15px; line-height: 1.5; color: var(--k-ink); background: var(--k-paper); border: 1px solid var(--k-rule-strong); border-radius: 2px; padding: 8px 10px; box-sizing: border-box` |
| `.k-adv__send` (+ `:disabled`) | `flex: none; width: 44px; height: 44px; padding: 0; border: 0; border-radius: 50%; background: var(--k-accent); color: var(--k-on-accent); display: flex; align-items: center; justify-content: center; cursor: pointer` (+ `opacity: .5; cursor: default`) |
| `.k-adv__hint` | `margin: 6px 0 0; font-size: 12px; color: var(--k-ink-3)` |

In the existing `@media (max-width: 767px)` block add: `.k-adv-dock { bottom: 16px; }` · `.k-adv:not(.is-full) { left: 8px; right: 8px; bottom: 8px; transform: none; width: auto; height: min(600px, 100dvh - 16px); }` · `.k-adv__ticks .k-adv__medlabel { display: none; }` + `.k-adv__medrow { display: block; }` (A5: Step 3 emits the median label twice — in the ticks row and in a `.k-adv__medrow` above the track, `position: relative; height: 16px; font-size: 11px; color: var(--k-ink-2)`, label abs at the same clamped left with `translateX(-50%)`). Base rule: `.k-adv__medrow { display: none; }`.

**Rationale:** Static markup keeps canvas copy in one file and lets `TestStaticRoutes` assert it; `<dialog>` per A8. Values transcribed from canvas so 1280/375 artboards match.

**Verification:** load `/` (server from Step 4 or any) → dock visible bottom-centre in both themes; `document.querySelector('dialog.k-adv').showModal()` from console renders the empty panel at canvas geometry (desktop + 375).

**If this fails:** revert the two files; dialog top-layer quirk → fall back to `section role=dialog` + own backdrop div (canvas markup) with manual focus trap — record as deviation.

---

### Step 3: `advisor.js` — panel behaviour, SSE client, rendering; router wiring; static test [PATCHED → 2610130822-kura-advisor-panel-patch.md]

**Objective:** S1–S9 behaviour. **Confidence:** Medium. **Depends on:** Step 2 (classes/markup); Step 1 only for correct summaries.

**Files:** `internal/web/static/js/advisor.js` (new), `internal/web/static/js/app.js`, `internal/web/server_test.go`

**Changes — `advisor.js`** (module; DOM present — module scripts run after parsing):

```js
import { api, esc, num, twd } from "./util.js";
import { valuation, badgeHTML, isTracked, HEART, OUT_ICON, PENDING_ICON } from "./card.js";

// Tool → canvas helper label + in-progress copy. Lines come from real tool
// events only; the canvas 風格 helper is not drawn because its stated grounds
// (大家都在找) are mocked demand, not data.
const STEP = {
  search_listings: ["行情", "搜尋在售商品"], model_price_stats: ["行情", "查詢型號行情"],
  market_overview: ["行情", "查看市場概況"], get_listing: ["成色", "讀取商品詳情"],
};
const MAX_FIGURES = 4; // <rationale comment: prompt asks 2–4 picks; Mock lists 6; figures after that would bury the answer>
const REF = /\[\[([a-z]+:[\w-]+)\]\]/g;
const NET_ERR = "暫時無法載入資料，請稍後再試。";
const CURSOR = `<span class="k-adv__cursor" aria-hidden="true"></span>`;

const dock = document.querySelector("[data-adv-open]");
const dlg = document.querySelector("dialog.k-adv");
const scroller = dlg.querySelector(".k-adv__body");
const log = dlg.querySelector(".k-adv__log");
const form = dlg.querySelector(".k-adv__form");
const input = form.elements.q;
const sendBtn = form.querySelector(".k-adv__send");
const maxBtn = dlg.querySelector("[data-adv-max]");
const session = crypto.randomUUID(); // per page load — see A7 rationale, as a comment
let busy = false, navigating = false;

export function initAdvisor(overviewP) {
  overviewP.then((ov) => { dlg.querySelector(".k-adv__demo").hidden = ov.advisor !== "mock"; }).catch(() => {});
}
```

Behaviour (functions; names indicative):

1. **States.** `setFull(full)`: toggle `dlg.classList` `is-full`; `maxBtn` `aria-label`/`title` = `full ? "縮小成視窗" : "放大成整頁"`. Dock click → `setFull(false); dlg.showModal()`. `[data-adv-max]` → `setFull(!dlg.classList.contains("is-full"))`. `[data-adv-close]` → `dlg.close()`. `dlg` `close` event → `setFull(false)` (canvas: reopen = floating); `if (!navigating) dock.focus()`; `navigating = false`. Esc = native `cancel` → `close`.
2. **Navigation close (A7).** `dlg` click listener: plain left click (no modifiers, `button === 0`) on `a[href^='/']` without `target` → `navigating = true; dlg.close()`; no `preventDefault` (router's document listener then navigates).
3. **Send.** Prompt chips → `ask(btn.textContent)`. `form` submit → `preventDefault(); ask(input.value)`. `input` keydown: `Enter && !shiftKey && !isComposing` → `preventDefault(); form.requestSubmit()`.
4. **`ask(text)`:** trim; return if empty or `busy`. `busy = true; sendBtn.disabled = true; input.value = ""`. Remove `.k-adv__prompts`. Append `<div class="k-adv__user">${esc(text)}</div>` and a turn `<div class="k-adv__turn"><ol class="k-adv__steps" aria-label="助手的進度" hidden></ol><div class="k-adv__answer" aria-live="polite" aria-busy="true"></div></div>`. State `raw = ""`, `known = new Map()`, `errText = ""`. For each event from `events(text)` — compute `pinned = nearBottom()` first, then: `text` → `raw += ev.text`, paint streaming; `tool_call` → if `STEP[ev.tool]`, unhide `ol`, append `<li data-tool>` `<span class="k-adv__who">{label}</span><span>{pending}⋯</span>`; `tool_result` → last `li[data-tool="{tool}"]:not(.is-done)`: `is_error` → remove it; else replace text span with `esc(summary)`, append `<span class="k-adv__ok" aria-label="完成">✓</span>`, add `is-done`; `listings` → `known.set(it.id, it)` each; `error` → `errText = ev.text`; `done` → ignore. Then `if (pinned) scroller.scrollTop = scroller.scrollHeight`. `catch` (fetch/parse/non-200) → `console.warn(err); errText ||= NET_ERR`. `finally`: final paint (figures), `answer.removeAttribute("aria-busy")`, `errText` → append `<p class="k-adv__err">${esc(errText)}</p>`, `fillPositions(answer)` (not awaited), `busy = false; sendBtn.disabled = false`, scroll if pinned. `nearBottom = () => scroller.scrollHeight - scroller.scrollTop - scroller.clientHeight < 120` (A12).
5. **`events(message)`** — async generator: `fetch("/api/chat", {method: "POST", headers: {"Content-Type": "application/json"}, body: JSON.stringify({session, message})})`; `!r.ok || !r.body` → throw; `getReader` + `TextDecoder` (`decode(value, {stream: true})`); split on `"\n\n"`; yield `JSON.parse(chunk.slice(6))` for chunks starting `"data: "` (legacy loop).
6. **Markdown** (`blocks(src, known)` → `[{html, ids}]`, port of legacy `renderMD`): `esc(src).split("\n")`; bullet regex `/^\s*(?:[-*•]|\d+\.)\s+(.*)$/` groups consecutive lines into one `<ul>` block; other non-blank lines → `<p>` (leading `#+\s*` stripped); inline: `**x**` → `<strong>`, `REF` → `chip(id)` pushing `id` into the block's `ids`. `chip(id)`: `<a class="k-adv__chip" href="/listing/${encodeURIComponent(id)}">` + text per A6.
7. **Paint.** Streaming: `blocks.map(b => b.html).join("")` with `CURSOR` inserted before the last block's closing `</p>` / `</li></ul>` (no blocks → `CURSOR` alone). Final: after each block, `figure(it)` for its ids passing A3 (known, `status === "active"`, not shown yet, `< MAX_FIGURES`).
8. **`figure(it)`** — `name = it.model_name || it.title`, `href = /listing/{enc id}`, `shop = it.source_name || it.source`:
   `<figure class="k-adv__rec" aria-label="推薦：{name}{ code}">` → `.k-adv__mat` (img alt `{name} 商品照片（店家白底照）` | `PENDING_ICON` + `<span class="k-card__pending">店家照片準備中</span>`) → `.k-adv__recmeta` [`.k-adv__recid`: brand span, `<a class="k-adv__name k-link" href>`{name}, grade line `<span class="k-grade">{label}</span>{desc_zh}` when label] [`.k-adv__recprice`: `<span class="k-price">{twd(price_twd)}</span>` + `badgeHTML(valuation(it))`] → `<div class="k-adv__pos" hidden data-brand data-group data-source data-code data-label data-price>` → `.k-adv__acts`: `<button type="button" class="k-adv__btn" data-track="{id}" aria-pressed="{isTracked}">{HEART}<span data-track-text="追蹤">{追蹤中|追蹤}</span></button>`, `<a class="k-adv__btn" href>看詳細</a>`, `<a class="k-adv__btn k-adv__btn--shop" href="{url}" target="_blank" rel="noopener noreferrer">{shop} 商品頁<span class="k-sr">（另開視窗）</span>{OUT_ICON}</a>`. Every interpolation `esc`'d.
9. **`fillPositions(root)`** (A5): per `[data-adv-pos]` with `group` + `code`: `model = await cached(brand, group)` (`Map` of promises of `api("/api/models/{enc brand}/{enc group}").catch(() => null)`); `s = model?.model.by_grade?.find(x => x.source === source && x.grade.code === code)`; skip unless `s && s.count >= 2 && s.max_twd > s.min_twd`. `at(v) = clamp 0–100 of (v − min)/(max − min) × 100`; `role="img"`, `aria-label` `同款 {label} 在售價格 {twd min} 至 {twd max}，中位數 {twd median}，這件 {twd price}`; inner: `<span class="k-adv__poscap">在同款 {label} 的 {num(count)} 件裡的位置</span>`, `.k-adv__medrow` (mobile median label), `.k-adv__track` (`<i class="k-adv__med" style="left:{at(median)}%">`, `<i class="k-adv__this" style="left:{at(price)}%">`), `.k-adv__ticks` (min, `<span class="k-adv__medlabel" style="left:{clamp 25–75 of at(median)}%">中位數 {twd}</span>`, max); unhide.

**Changes — `app.js`:**

```js
// Before:
import { renderModel } from "./model.js";
// After:
import { renderModel } from "./model.js";
import { initAdvisor } from "./advisor.js";
// … and after the fx footer block, before the final render():
initAdvisor(overview());
```

**Changes — `server_test.go` `TestStaticRoutes` cases:** add `{"/", "text/html", "找包顧問", 200}` and `{"/js/advisor.js", "javascript", "initAdvisor", 200}`.

**Rationale:** One module owns the panel (shell-level, not a route view — it outlives navigation). Reuses `card.js` badge/track and the `dist.js` `by_grade` lookup so every number has one source. Figures render once at stream end to avoid refetching per text delta.

**Verification:** `go test ./internal/web/...`; console free of errors on `/`; quick Mock send renders steps → text → figures (full matrix Step 4).

**If this fails:** remove the `initAdvisor` import/call — shell still renders (dock inert); revert files.

---

### Step 4: End-to-end verification [PATCHED → 2610130822-kura-advisor-panel-patch.md]

**Objective:** S1–S11 evidence. **Confidence:** Medium (hotlinked photos/fonts need network). **Depends on:** Steps 1–3.
**Verify-Projex:** Encouraged

**Procedure:**
1. `go vet ./...`; `go test ./...`; hygiene grep `grep -rnE "[0-9]{10}-[a-z0-9-]+-(plan|sprint|def|nav|audit|patch)" internal cmd` → none.
2. Background `go run ./cmd/kura serve -db kura.db -addr 127.0.0.1:8094` from the worktree root, `ANTHROPIC_API_KEY` unset, no crawl flags → log `advisor=mock`.
3. Browser pane matrix:

| # | Case | Check |
|---|---|---|
| V1 | dock on `/`, `/listing/{id}`, `/model/LOUIS%20VUITTON/M40156`-type, `/nope` | dock bottom-centre; footer not covered at page end |
| V2 | open → max → shrink → close; reopen → Esc | floating 680 × 640 + backdrop; full = inset 0, 720 column, no backdrop; labels swap; focus to textarea on open, dock on close; dock hidden while open; Tab stays inside panel |
| V3 | 示範模式 | pill visible, `title` canvas copy |
| V4 | chip 「預算 8 萬，通勤可放 A4 的 LV 托特包」 | chips vanish; bubble; 1 step line 行情 「找到 N 件符合條件的商品，回傳 M 件」 ✓; answer = Mock text with bullets; chips `{model} · {code}`; ≤ 4 figures after the list, each: photo/pending, brand, name link, grade + 定義, price + badge equal to that listing's detail page badge |
| V5 | position bar | for one figure, `/api/models/{brand}/{group}` `by_grade` {source, code}: caption count, min/median/max labels and aria-label equal; dot left = `(price−min)/(max−min)`; absent when that grade has < 2 in sale |
| V6 | 「哪些愛馬仕包明顯低於行情？」, then free text 「100 萬以內的 GOYARD 腰包」 | second turn appends below first; no-result text, no figures, no crash |
| V7 | composer | Enter sends, Shift+Enter newline, blank Enter ignored, zh IME composition Enter does not send; second send while streaming ignored, 送出 disabled |
| V8 | figure 追蹤 → 看詳細 | 追蹤中 + `aria-pressed`; 看詳細 navigates, panel closes, detail 追蹤這件 shows 追蹤中 (shared set); dock → reopen → conversation intact |
| V9 | chip click / name link / shop link | chip + name → `/listing/{id}` (panel closes); shop link → new tab, panel stays |
| V10 | Claude-shaped replay (inspection only, no source change): `javascript_tool` stubs `window.fetch` for `/api/chat` with a `ReadableStream` of scripted SSE: `tool_call model_price_stats` → `tool_result` 「{model}：在售 N 件」 → `tool_call get_listing` → `tool_result is_error` → `tool_call get_listing` → `tool_result` → `listings` [one real item from `/api/listings?limit=1&sort=deal`] → `text` deltas 「我先推薦這一件：[[id]]」 / 「**NT$…**，…」 → `error` 「Claude API 目前負載過高，請稍後再試」 → `done` | lines 行情 ✓, 成色 ✓ (failed call's line removed), no 風格; chip in paragraph 1 then figure after it; bold; error paragraph after answer; cursor gone; composer re-enabled |
| V11 | network failure replay: stub rejects | 「暫時無法載入資料，請稍後再試。」 paragraph |
| V12 | `/legacy/` advisor | send a prompt → still renders (unchanged client) |

   V2 + V4 at 1280 and 375 × `?theme=light|dark` (screenshots vs canvas `AdvisorOpen`, `AdvisorFull`, `AdvisorMobile`; dark = token-driven gap); others 1280 light (V5, V8 also 375).
4. Per view: `read_console_messages` onlyErrors empty; at 375 `document.documentElement.scrollWidth <= innerWidth` with panel open and closed; reduced-motion emulation (if available) → cursor static.
5. Stop server; worktree `git status` clean except committed work; scratch files deleted.

---

## Verification Plan

### Automated Checks
- [x] `go vet ./...` clean
- [x] `go test ./...` green incl. `TestSoldStatsGatedUntilHistory`, extended `TestStaticRoutes`, existing fake-API agent tests

### Manual Verification
- [x] Step 4 matrix V1–V12
- [x] Home / detail / model pages unchanged apart from the dock
- [x] Source hygiene grep → none

### Acceptance Criteria Validation

| Criterion | How to Verify | Expected Result |
|---|---|---|
| S1 | V1, V2 (both widths) | states + geometry per canvas; focus round-trip |
| S2 | V2, V3 | header per canvas; pill only in mock |
| S3 | V4 | chips send; gone after first send |
| S4 | V7 | Enter/Shift+Enter/IME/busy rules |
| S5 | V4, V10 | real steps, ✓ on result, no 風格; markdown + chips; cursor while streaming |
| S6 | V4, V5, V10 | ≤ 4 active figures after first-reference block; fields = API |
| S7 | V8, V9 | in-app link closes panel; conversation kept |
| S8 | V10, V11 | error text paragraph; network copy; composer usable |
| S9 | `tools_test.go`; V5 vs API; code review `advisor.js` | gated fields absent + marker; figure numbers = server fields |
| S10 | 375 scrollWidth; dark screenshots; Tab | no overflow; tokens switch; focus contained + ring visible |
| S11 | vet/test/grep; V12 | green / none / legacy works |

**Data-class trace (advisor):** figure brand / name / grade / definition / photo / price / shop link ← listing (real) | badge ← `fair` via `valuation` (real; same as cards) | position bar ← `by_grade` count/min/median/max (real; same as distribution) | progress lines ← tool `Summary` (real; sold count only when `SoldStatsReady`) | answer ← Mock template over real search / Claude over tool JSON (sold-derived fields gated, real-derived → 資料不足) | 示範模式 ← overview `advisor` mode. Mocked: none rendered (風格 helper omitted). 追蹤 = in-memory UI state, no number.

---

## Rollback Plan

1. `git -C <work-root> log --oneline <plan-commit>..HEAD` → this plan's execution commits.
2. `git -C <work-root> revert <sha>…` newest first (sprint branch only): dock disappears; advisor tools return to ungated fields (H2 reopens).

---

## Notes

### Split Decision
No split — single scope (one `.projex` folder; `internal/advisor` + `internal/web` in one repo, no spec/impl layering), ~535 lines / ~53 KB (size half of the heuristic met: CSS table + code sketches) but 4 steps — heuristic needs both, so not tripped; Steps 2–3 are tightly coupled (classes ↔ markup ↔ module) and would not split cleanly. Fits one strategy run: one shell component + one backend gate, comparable to iterations 2–3.

### Risks
- Native `<dialog>` top-layer + `transform` centring: some engines mis-place a transformed modal → Step 2 fallback (canvas `section` + own backdrop + focus trap), logged as deviation.
- Model API for large groups is heavy (up to ~477 KB for 409 listings) per figure; cached per group, ≤ 4 per turn — acceptable locally; a stats-only endpoint is a later optimisation.
- Mock text preface names `ANTHROPIC_API_KEY` (developer-facing) beside the 示範模式 pill — kept (A10); reviewer wording is def L3.
- `aria-live` + `aria-busy` announce the whole answer at once; long Mock lists are verbose for screen readers.
- Claude may cite a sold figure from its own reasoning despite the gate; prompt line mitigates, cannot guarantee.

### Deferred (for the orchestrator to record in the sprint nav — not decided here)
| Item | Blocker |
|---|---|
| Advisor **error state** (designed) | Canvas pass 2 — no canvas evidence; A9 functional fallback only |
| Advisor **spend-cap-reached state** + demo fallback | Def L3 hosted-mode policy (open) + canvas pass 2; hosting = Phase 6 (out of sprint) |
| 5-photo recommendation gallery | Def L3 gallery question (open) + single provider photo |
| 風格 helper line | Def L3 re-ground-vs-drop (open) |

### Open Questions
None blocking execution. Provisional choices for orchestrator logging (sprint nav / def), not decided here:
- A1 helpers = real tool steps + tool→label map; 風格 omitted → def L3 「real or scripted」 / 風格 grounding.
- A10 demo pill + Mock preface copy verbatim → def L3 reviewer-facing wording.
- Canvas deviations: A3 figure count/placement rule (canvas draws one), A4 single photo, A5 bar omitted for < 2 / median-label clamp + mobile row, A6 chip as link, A7 close-on-navigate + per-load session, A8 native dialog, A9 error fallback.
- Def § Canvas gaps tally: 「Advisor: error state, spend-cap-reached state」 → Deferred (table above); 「dark × advisor」 combo → built token-driven (verify V2/V4 dark).
- H2 closed by Step 1 (one gate: `Engine.SoldStatsReady`); def L4 「Day-1 appearance」 rule still replaces `MinSoldHistory` if answered.
