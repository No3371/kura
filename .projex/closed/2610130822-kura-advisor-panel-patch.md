# Patch: Kura advisor panel (顧問)

> **Sprint:** 2610130524-kura-interactive-mock-sprint.md — iteration 4
> **Status:** Complete
> **Author:** Claude Opus 5.5 (patch-projex plan-execution, sprint subagent)
> **Parent:** 2610130815-kura-advisor-panel-plan.md
> **Directive:** Execute `2610130815-kura-advisor-panel-plan.md` in full (patch-projex plan-execution mode) on the sprint branch; verify per the plan's own checks.
> **Source Plan:** 2610130815-kura-advisor-panel-plan.md (Steps 1–4, all objectives)
> **Result:** Success
> **Commits:** `5e65a9a` fix(advisor) — Step 1 | `912a354` feat(web) — Steps 2–3 | doc commit (this file + plan close)
> **Related Projex:** 2610130815-kura-advisor-panel-plan.md | 2610130524-kura-interactive-mock-sprint.md | 2610130449-kura-mock-def.md | 2610130748-kura-model-page-audit.md (carry-over H2 closed here)

---

## Summary

Canvas 顧問 built into the new shell: fixed 「蔵 找包顧問」 dock on every route → modal `<dialog>` floating panel (680 × 640, bottom 24, backdrop) ⇄ full page (720 column) → closed; Esc / 關閉顧問 return focus to dock. Panel streams `POST /api/chat` SSE: user bubble → 「助手的進度」 lines from real tool events (行情 / 成色; ✓ + server summary on result; failed call's line removed) → markdown answer with `[[id]]` chips + blinking cursor → at stream end ≤ 4 recommendation figures (photo, brand, name link, grade + 定義, NT$ + card badge, same-grade position bar from model API `by_grade`, 追蹤 / 看詳細 / 商品頁). In-app links close panel; conversation persists for page lifetime. Error fallback = plain paragraph (server `error` text or shell copy). Carry-over **H2** closed: advisor tools omit sold count / sell-through / days-to-sell + add `sold_stats: insufficient_history` until `radar.Engine.SoldStatsReady()`.

---

## Changes

### Sold-stat gate (Step 1, `5e65a9a`)

**File:** `internal/advisor/tools.go` — Modified
- `soldUnknown` const + rationale comment (l.135–140)
- `modelStats`: `sold_observed` / `sell_through` / `median_days_to_sell` + summary 「、已售出 N 件」 only when `t.Radar.SoldStatsReady()`; else `sold_stats` marker (l.273–278)
- `getListing`: `model_sold_observed` gated, else marker (l.300–303)
- `overview`: `o := Overview()` moved before loop; `model.Sold` → `*int` `omitempty`, set only when `o.SoldStatsReady`; top-level `sold_observed` gated, else marker (l.333–344)

**File:** `internal/advisor/prompt.go` — Modified
- System prompt 呈現方式 bullet on `sold_stats: insufficient_history` → 資料不足, unknown not zero (l.33–34)
- `model_price_stats` description sentence on omission (l.83)

**File:** `internal/advisor/tools_test.go` — Created (107 lines)
- `TestSoldStatsGatedUntilHistory`: not ready → all gated keys absent across 3 tools, marker present, summary lacks 已售出; then a listing first seen −15 d + `MarkGone` now + `Refresh` → `sold_observed == 1`, `sell_through` present, no marker, summary 「已售出 1 件」, `model_sold_observed == 1`, overview + `top_models[0]` `sold_observed == 1`.

**Why:** def 「One source per stat」 — model page shows 資料不足 from the same gate; omit + marker cannot be read as 0.

### Shell markup + CSS (Step 2, `912a354`)

**File:** `internal/web/static/index.html` — Modified (l.56–86)
- Dock `<button class="k-adv-dock" data-adv-open aria-haspopup="dialog">` + `<dialog class="k-adv" aria-label="Kura 顧問">`: header (seal, h2, hidden 示範模式 pill w/ canvas title, max/shrink icon button, close), body (canvas intro + 4 `data-adv-prompt` chips), form (sr-only 問顧問, textarea `autofocus`, 送出, hint). Copy + SVG paths verbatim from canvas `Main.dc.html` 顧問 block.

**File:** `internal/web/static/kura.css` — Modified (block l.363–456; mobile l.489–492; header comment)
- All plan-table selectors, values from canvas `winStyle` / inline styles; `@keyframes k-blink` + reduced-motion off; `::backdrop` canvas literal, transparent in full; ≤ 767 px: dock bottom 16, panel 8/8/8 × min(600, 100dvh − 16), median label moved to `.k-adv__medrow`.

### Behaviour module + wiring (Step 3, `912a354`)

**File:** `internal/web/static/js/advisor.js` — Created (344 lines)
- `STEP` tool→[helper, pending copy]; `MAX_FIGURES = 4`; session `crypto.randomUUID()` per load (rationale comments name rejected alternatives: figure-per-reference, `sessionStorage`).
- States (`setFull`, dock → `showModal`, close event → floating + dock focus), nav-close click handler, prompt chips, submit, Enter / Shift+Enter / `isComposing`.
- `ask` (busy guard, turn markup, event switch, `finally` final paint + error paragraph + `fillPositions`), `events` SSE async generator, `blocks` / `inline` / `chip` markdown port, `streamingHTML` (cursor in last block), `finalHTML` (figure after first-reference block, active + known only), `figure`, cached `modelData`, `fillPositions` (A5 rule, 25–75 % median-label clamp).

**File:** `internal/web/static/js/app.js` — Modified: import + `initAdvisor(overview())` before `render()` (l.5, l.97).

**File:** `internal/web/server_test.go` — Modified: `TestStaticRoutes` + `{"/", …, "找包顧問"}`, `{"/js/advisor.js", …, "initAdvisor"}` (l.31, l.33).

---

## Verification

**Method:** `go vet ./...`, `go test ./...`, hygiene grep, `node --check`; live server `go run ./cmd/kura serve -db kura.db -addr 127.0.0.1:8094` (`ANTHROPIC_API_KEY` / `KURA_ADVISOR` unset → log `advisor=mock`); Browser pane matrix V1–V12 via DOM/geometry/computed-style inspection + screenshots where capture rendered.

**Result:**
```
go vet ./...            → clean
go test ./...           → ok advisor | radar | brandoff | store | web (TestSoldStatsGatedUntilHistory, TestStaticRoutes incl. new cases, fake-API agent tests)
hygiene grep            → no matches (exit 1)
log                     → msg=serving url=http://127.0.0.1:8094 advisor=mock ; /api/overview sold_stats_ready:false
```

| # | Evidence | ✓/✗ |
|---|---|---|
| V1 | `/`, `/listing/brandoff%3A2101220216016`, `/model/LOUIS%20VUITTON/M40156`, `/nope`: dock centred, bottom 28, visible; at page end footer text bottom 696 < dock top 716 (596 on not-found) | ✓ |
| V2 | floating 680 × 640 @ bottom 24, `:modal`, backdrop `rgba(30,25,21,.18)`; full 0/0 inset, column 720, backdrop transparent, label 縮小成視窗 + shrink icon; shrink → 680 / 放大成整頁; focus textarea on open, dock on close (after close event); dock `visibility: hidden` while open; header link not focusable while modal (inert) | ✓ |
| V3 | pill visible in mock, title 「未設定 API 金鑰，顧問以範例資料回答」 | ✓ |
| V4 | prompt chip: chips removed; bubble; one line 行情 「找到 684 件符合條件的商品，回傳 6 件」 ✓; Mock text paragraphs + `<ul>`; chips `バティニョール ヴェルティカル · M51153` → `/listing/brandoff%3A2107400261920` …; 4 figures after the list (photo alt, LOUIS VUITTON, name link, ランクB + 定義, NT$16,147 + 低於行情 43% = deal_pct 0.4286) | ✓ |
| V5 | all 4 figures: caption count, min/median/max ticks + aria-label = `/api/models/…` `by_grade` {brandoff, ランクB}; dot left 0 % = (price − min)/(max − min); median 74.9985 % vs 75.00 expected; HERMES 060991CK ランクB count 1 → bar hidden | ✓ |
| V6 | 「哪些愛馬仕包明顯低於行情？」 then 「100 萬以內的 GOYARD 腰包」: turns appended in order; GOYARD turn = no-result text, 0 figures, no error | ✓ |
| V7 | blank Enter → prevented, no send; `isComposing` Enter → not prevented, no send; Shift+Enter → not prevented; Enter → sends, textarea cleared, 送出 disabled; Enter while streaming → no new turn, text kept | ✓ |
| V8 | figure 追蹤 → `aria-pressed=true` 追蹤中; 看詳細 → `/listing/brandoff%3A2107400261920`, panel closed, detail 追蹤這件 shows 追蹤中 (shared set); reopen → 1 turn, 4 figures, figure still 追蹤中 | ✓ |
| V9 | chip → `/listing/brandoff%3A2107601166086` panel closed; name → `/listing/brandoff%3A2101220221119` panel closed; shop link (`target=_blank`, `noopener noreferrer`; navigation suppressed in test) → panel stays open | ✓ |
| V10 | fetch-stub replay: mid-stream cursor present, 0 figures, `aria-busy`; final steps 行情 「カルトシエールGM：在售 9 件」 ✓ + 成色 「讀取商品 カルトシエールGM」 ✓ (is_error line removed), no 風格; `P` (chip) → `FIGURE` → `P` (**NT$10092** bold) → `P.k-adv__err` 「Claude API 目前負載過高，請稍後再試」; cursor gone; 送出 re-enabled; position bar shown | ✓ |
| V11 | rejected fetch and HTTP 400 → 「暫時無法載入資料，請稍後再試。」 paragraph, composer enabled | ✓ |
| V12 | `/legacy/` send → steps ⋯/✓, Mock text, 6 cards, no error | ✓ |
| 375 | panel 8/204/359 × 600 (bottom 8); dock bottom 16; `scrollWidth` 375 closed / open / with conversation; panel body no x-overflow; median label on own row (222–312 px inside track), ticks-row copy `display: none`; full 0/0 375 × 812; no visible button/link < 24 px (chips inline, exempt) | ✓ |
| Themes | dark computed: panel `rgb(36,32,28)`, ink `rgb(238,231,219)`, chip `rgb(227,161,166)` on `rgb(58,31,35)`, mat/shop/textarea tokens; screenshots 1280 light (floating conversation) + 375 light/dark match canvas `AdvisorOpen` / `AdvisorMobile` | ✓ |
| Console | `read_console_messages` onlyErrors → none | ✓ |

**Status:** PASS

### Per-criterion

| Criterion | Result | Evidence |
|---|---|---|
| S1 Dock + states | ✓ | V1, V2, 375 row |
| S2 Header | ✓ | V2, V3 |
| S3 Intro | ✓ | V4 (chip sends, chips removed) |
| S4 Composer | ✓ | V7 |
| S5 Turn | ✓ | V4, V10 |
| S6 Figures | ✓ | V4, V5, V10 |
| S7 Navigation | ✓ | V8, V9 |
| S8 Errors | ✓ | V10, V11 |
| S9 One source per stat | ✓ | `TestSoldStatsGatedUntilHistory`; V5 numbers = API; `advisor.js` does no math beyond bar positions + label clamp |
| S10 Responsive/themes/a11y | ✓ (caveats in Notes) | 375 row, Themes row, inert check, global `:focus-visible` rule, reduced-motion rule in CSSOM |
| S11 vet/test/hygiene/legacy | ✓ | Result block, V12 |

---

## Impact on Related Projex

| Document | Relationship | Update Made |
|---|---|---|
| 2610130815-kura-advisor-panel-plan.md | Source plan | All steps `[PATCHED]` → this doc; Status `Complete`; moved to `closed/` |
| 2610130524-kura-interactive-mock-sprint.md | Sprint nav | None (orchestrator owns nav; iteration 4 result returned to it) |
| 2610130449-kura-mock-def.md | Def | None here — plan's provisional choices (A1, A10) + Deferred table are for the orchestrator to log |

---

## Notes

**Deviations from plan (all minor, behaviour per plan intent):**
- `data-adv-pos` attribute added to the figure's `.k-adv__pos` (plan Step 8 markup omitted it; Step 9 selects on it).
- `modelData` evicts a failed model fetch from the cache (plan: cache for page lifetime) — a later turn can retry instead of a permanent missing bar.
- `fillPositions` re-pins scroll when a bar appears and the reader was at the end (A12 applied to late content).
- CSS beyond the plan table, all canvas-sourced: `.k-adv__grade .k-grade` ink colour, compact figure badge (card size), figure price weight 400, shop-link arrow 11 px, dock `text-align: left`; `flex: none` on head/form; `overflow-wrap` on answer / name; `box-sizing` on dialog / buttons.
- Header: canvas wraps the title row in an extra column `div`; plan's flat `.k-adv__title` used (no visual difference).

**Verification caveats:**
- Browser pane was hidden/behind during the run: rendering frames ran only intermittently, so `close` event (and its focus return) fired only on the next frame; full-mode 1280 and dark 1280 screenshots captured a cropped region. Those states verified by geometry + computed styles instead of images.
- Reduced-motion emulation unavailable → verified the `@media (prefers-reduced-motion: reduce)` rule exists in CSSOM; IME verified with a synthetic `isComposing` keydown.
- Full mode at 1280 measures 1265 wide: the page's classic scrollbar (15 px) stays visible beside the full-page panel. Not in canvas (artboards have no scrollbar); left as is — locking body scroll would shift layout on open. Candidate for audit.
- Dark backdrop uses the light canvas literal (def gap 「token-driven」, plan A8).

**Deferred (from plan, unchanged):** designed error state | spend-cap state | 5-photo gallery | 風格 helper — blockers per plan § Deferred.

**Cleanup:** server (`kura.exe` child of `go run`) stopped; Browser tabs closed, viewport reset; scratch log only in session scratchpad; worktree `kura.db*` untouched by writes (read-only serve).
