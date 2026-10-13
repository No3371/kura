# Audit: Kura browse page (價格最美)

> **Status:** Complete (Accepted)
> **Sprint:** 2610130524-kura-interactive-mock-sprint.md — iteration 5
> **Author:** Claude Opus 5.5 (auditor, sprint subagent)
> **Parent:** 2610130928-kura-browse-page-patch.md
> **Related Projex:** 2610130913-kura-browse-page-plan.md (audited plan, S1–S9, A1–A12) | 2610130928-kura-browse-page-patch.md (execution claims) | 2610130952-kura-browse-audit-fixes-patch.md (F1–F5 fixes, re-audited below) | 2610130524-kura-interactive-mock-sprint.md | 2610130449-kura-mock-def.md (Canvas deviations "Grade filter", Canvas gaps "dark × browse", L4 grade ordering) | 2610130602-kura-canvas-foundation-home-audit.md (F3 link contract; F1 / F6 open)
> **Audit Date:** 2026-10-13 | **Work Period:** 2026-10-13, commits `2f15a43` (radar + API) `265eea0` (web) `b302594` (patch doc) on `projex/sprint/2610130524-kura-interactive-mock` (range `a555587..b302594`); re-audit 2026-10-13 of fix `2d58893` + doc `4bea535` (range `6758d95..4bea535`)
> **Subject:** plan S1–S9 vs code, tests, live app (worktree `kura.db` served read-only on 127.0.0.1:8095, no crawl flags; fresh scratch DB on :8096 for site-empty, deleted after), canvas `project/Main.dc.html` `browseData` block + `renderVals` + Tokens `.k-field` + `Card.dc.html` row layout, read from the served version 1791845608-ba84 (sprint nav records it identical to pin 1791839497-46ec)

---

## Audit Summary

**Claim:** `/browse` built per canvas Browse / BrowseMobile on real data — URL-as-state filters (two forms kept in step), 40-card pages with 「再看 N 件」, in-app header search, server `offset` + overview `by_category`; S1–S9 all ✓.

**Verdict:** Verified — every success criterion holds on independent re-check; one undisclosed defect introduced (F1), which the plan's Risks had dismissed as harmless.

**Assessment:** Completeness: High | Correctness: Medium-High | Quality: High | Value: High

**Top Issues:**
1. **F1** — `replaceState` leaves the router's `current` stale; any later `#` navigation (header 登入 / 加入會員 are `href="#"` on every page) re-renders Browse, discarding load-more pages (reproduced: 80 → 40 cards). Fixable now.
2. **F2** — Back from a listing to Browse always lands on page 1 (40 cards); a user who paged to card 85 loses their place. Needs a decision (URL param vs. session memory vs. accept).
3. **F3 / F4** — two source-hygiene items in `browse.js` (misleading `stateQuery` comment; `apply`'s load-more disable has no rationale). Fixable now.

---

## Claims vs Evidence

| Claim | Evidence | Status | Notes |
|---|---|---|---|
| `go vet` + `go test` green | Clean `go build ./... && go vet ./... && go test -count=1 ./...` → ok advisor, radar, brandoff, store, web | ✓ | IDE diagnostic "unknown field Offset in struct literal of type radar.Filter" is stale — field exists (`radar.Filter.Offset`), build clean |
| `Filter.Offset` slices after `total` | `radar.Search`: `out = out[min(max(f.Offset, 0), len(out)):]` after `total = len(out)` | ✓ | Live: `offset=-5` → first item, total 15,367; `offset=abc` → 0; `offset=15366` → 1 item; past end → `[]` |
| Paging lists exactly the counted matches | Exhaustive walk at limit 40: `deals=1&sort=deal` 1,471 ids / 1,471 unique; `brand=GOYARD&sort=price_asc` 98/98; `category=travel&sort=newest` 37/37 | ✓ | No duplicate / skipped id under a fixed index |
| `overview.by_category` | `/api/overview` → shoulder 6,861 · handbag 4,741 · tote 2,073 · backpack 524 · bag 361 · boston 297 · belt 286 · clutch 126 · business 61 · travel 37 | ✓ | = plan probe |
| Advisor search unchanged | `advisor/tools.go` builds its own `radar.Filter{…}` (no `Offset`); advisor tests green | ✓ | |
| Tests extended | `TestSearchFilters` (offset page = `all[1]`, past end, `ByCategory["tote"]`); `TestListingAPI` (`getList` offset 1 / 9) | ✓ | Meaningful: assert identity of the paged item, not only length |
| B1 價格最美 from Home | Real click on 「看全部價格最美 1,471 件」 → `/browse?deals=1&sort=deal`; h1 + crumb 價格最美 (`aria-current`); `1,471 件`; first 3 ids = Home shelf ids in order; 40 cards; 「再看 1,431 件」; title `價格最美 · Kura 蔵`; `window.__probe` survived (no reload) | ✓ | |
| B3 load-more | Sort → 最新上架, 再看 → 80 cards, 「再看 1,391 件」 | ✓ | |
| B5 budget typing | Real typing `2,0000` + Enter → `?max_twd=20000&q=M40156`, both forms show `20000`, focus stays in field; `0 件` = API total 0 (cheapest M40156 is NT$26,239) | ✓ | Zero-result state 「沒有符合條件的包。」 path exercised |
| B7 header search | Real typing `M40156` + Enter in header → `/browse?q=M40156`, h1 全部在售, keyword filled, switch off, 31 件, `window.__probe` intact (no reload) | ✓ | |
| B10 invalid params | `?brand=NOPE&sort=x&min_rank=Z&max_twd=-5&deals=yes&category=pouch` → 全部在售, 15,367, all 7 controls default; `/browse/x` → 找不到這個頁面 + its title | ✓ | `category=pouch` (catalog type, 0 active) also falls back — the executor's whitelist deviation works |
| B11 / B12 mobile 375 | Wide form `display:none`; keyword 44px; toggle 44px; h1 24px; list `display:flex`, cells `20px 0` + 1px rule; row cards `116px | 211px`; toggle → `aria-expanded=true`, panel 5 fields × 44px × 343px; `scrollWidth 375 = innerWidth` before and after opening | ✓ | Screenshot = canvas BrowseMobile structure (dark) |
| B12 desktop 1280 | Label widths 150 / 120 / 160 / 130 / 293 (flex) / 150 / switch 110; field 38px; 4 columns of 275px; light (emulated) body `#F5F0E6`, dark tokens on screenshot | ✓ | |
| B13 keyboard | Tab from keyword → sort → switch: switch `:focus-visible`, `2px solid` focus token; Space → switch off, URL `/browse`, h1 全部在售, 15,367 | ✓ | |
| B14 site empty | Fresh DB `/browse?deals=1` → 「還在準備中。」 (title stays route default, as planned) | ✓ | |
| Console | `read_console_messages` onlyErrors → none across the run | ✓ | |
| S8 radar untouched | Diff of `radar.go` touches only `Filter`, `Search` slice, `Overview` map; no `0.05` / `anomalyPct` / `dealScore` / `confidenceWeight` line | ✓ | F1 / F6 stay open |
| Hygiene grep | No projex IDs in `internal/` / `cmd/` | ✓ | |

Cases B2, B4, B6, B8, B9 not re-driven; code paths shared with re-driven cases (B4/B6 by the exhaustive API walk and zero-result path above; B9 by the router read below).

---

## Objective Verification

### S1 Route — ✓ Verified
`routes` += `/^\/browse$/`; Home links, header form (`submit` listener → `go()`), mobile icon (`a[href^='/']`), direct load all render it; `/browse/x` not-found. **Issue:** F1 (router state after `replaceState`).

### S2 Head — ✓ Verified
Crumb / h1 / count / title per A2; count = API `total`; shelf parity with Home confirmed by id.

### S3 Desktop filters — ✓ Verified
Canvas order, labels, widths, placeholders transcribed exactly — checked against canvas markup itself, not the plan: labels 品牌 / 類型 / 最低成色 / 預算上限（NT$） / 關鍵字 / 排序 + switch; widths 150 / 120 / 160 / 130 / `flex 1 1 180px` / 150; placeholders 不限, 「型號、編號，可用日文：ネヴァーフル」; options 全部品牌 + `{name}（{count}）`, 任何成色 / 未使用以上 / 輕微使用感以上 / 明顯使用感以上, 低於行情優先 / 最新上架 / 價格 ↑ / 價格 ↓; form `gap 12px; padding 0 0 32px`; switch label `height 38px; gap 8px; 14px ink; nowrap; cursor pointer`, input 18×18 `accent-color: accent`. Brand list = canvas `KURA_BRANDS` set; 類型 10 real (canvas 3 = top 3, A4 deviation).

### S4 Mobile — ✓ Verified
Canvas mobile form (`column; gap 10px; padding 0 0 12px`, keyword 44px, button `flex 1 1 0; 44px; surface; rule-strong; radius 2px; gap 8px` + icon path `M2 4h12M4.5 8h7M7 12h2`, switch label 44px) matches. Canvas mobile switch label sets no `color` / `cursor`; impl inherits the desktop rule (ink, pointer) — no visible difference. Panel open state is plan-authored (A7, canvas gap).

### S5 Results — ✓ Verified
Grid `repeat(auto-fill, minmax(240px, 1fr)); gap 48px 28px`; mobile `flex column; gap 0`, cell `padding 20px 0; border-bottom 1px rule`; row cards = `Card.dc.html` row values (`116px minmax(0,1fr)`, gap 16, model 16px, price 15px) applied by width; load-more `padding-top 36px`, button `44px; 0 24px; transparent; 1px ink-2; radius 2px`. Copy 「再看 {n} 件」 = canvas.

### S6 Server — ✓ Verified (see Claims).

### S7 Responsive / themes — ✓ Verified (1280 + 375, light + dark, no overflow, focus ring, 0 errors).

### S8 One source per stat — ✓ Verified
Every figure is API `total`, `total − shown`, or overview `by_brand`; client computes no stat, converts no currency. Grade rule (def Canvas deviation "Grade filter", L4): cards render the shop's own grade (`cardHTML` untouched); 最低成色 maps canvas labels → internal rank floor `S` / `A` / `B` only as API params, never displayed. Live totals S 831 / A 10,152 / B 15,322 = grade tallies (596+235 / +9,321 / all − 44 ランクC − 1 unranked).

### S9 — ✓ Verified (vet / test / grep).

---

## Code/Implementation Inspection

### `internal/web/static/js/browse.js` (new, 260 lines)
**Claimed:** view + two synced forms + apply/load-more + document listeners. **Actual:** as claimed; Quality High — one whitelist (`readState`) serves URL and form input, `apiURL` shares the counted query, seq guard + `isConnected` on both async paths, IME-composition guard on Enter, load-more disabled during `apply`.

**Issues:**
- `apply` → `history.replaceState` without informing the router → **F1** (Medium).
- `renderBrowse` always renders page 1; no paging state in the URL → **F2** (Low, decision).
- `readState` `max_twd`: `String(Number(max))` turns ≥ 1e21 digit strings into `1e+21` (URL + field show exponent form, API ignores it, reload clears it) → **F5** (Low).
- Comment hygiene → F3, F4.

### `internal/web/static/js/app.js`
`go(url)` extracted; header-search `submit` listener (A10) with rationale naming the rejected native GET. Router `current` only updated in `render()` — the root of F1 (pre-existing design, newly exposed by Browse's `replaceState`).

### `internal/radar/radar.go`, `server.go`, tests — Quality High; minimal, gofmt'd, clamped.

### `kura.css` — browse block + mobile additions match canvas; `.k-bf__panel[hidden]` omission correct (base `[hidden] { display: none !important; }`).

### `home.js` / `model.js` — `EMPTY_HTML` export identical markup; model 看全部 comment present tense with the reason (no group filter).

**Undocumented:** none beyond the patch's own deviation list (all six verified as described).

---

## Testing Validation

**Execution:** all packages pass with `-count=1`; no flaky tests observed.
**Coverage:** Go side covers offset (identity, past end) and `by_category`; negative offset is clamped in code but not asserted (live-checked here). Browse JS has no automated test (repo has no JS test harness — consistent with iterations 1–4; HANDOFF §3.6 no build step).
**Missing:** a router test for "fragment navigation after `replaceState`" would have caught F1 — none possible without a JS harness; covered by manual case below.

---

## Documentation Audit

Patch doc accurate in every checked claim (counts, matrix rows re-driven, deviations). One inaccuracy carried from the plan: Risks bullet "app.js `current` … is not updated by `replaceState`; harmless" — disproved (F1). Line refs (`l.…`) in the patch are point-in-time and fine in a projex doc.

---

## Source Hygiene

**Scope:** Diff only (`a555587..b302594`, `.projex/` excluded).

| Rule | Location (symbol) | Quote | Severity |
|---|---|---|---|
| 3 (accuracy) | `browse.js` → `stateQuery` | "Canonical query: fixed key order, defaults omitted (sort=deal included)." | Low — parenthetical reads as "sort=deal is included in the query"; intent is the opposite (F3) |

No rule 1 / 2 / 4 / 5 violations: no projex IDs, no `file:ln`, no step banners; "One form reordered with CSS order was rejected …", row-card CSS comment, replaceState comment, header-search comment all name the rejected alternative.

**Rule 6 gaps:**
- `browse.js` → `apply` — `results.querySelector("[data-browse-more]")?.setAttribute("disabled", "")` has no comment; the reason (a 再看 click during a filter fetch would otherwise append the old grid's next page — the seq guard does not catch it because `loadMore` reads `view.seq` after `apply` bumped it) lives only in the patch doc's Notes (F4).

**Rule 6 vacuity:** none found.

**Rationale promotion:** no `**Rationale promoted:**` line in the patch doc. Code-level rationale (A1 URL state, A2 heading, A5 rank floor, A6 page size, A9 model 看全部, A10 header search, A11 commit-on-Enter, IME) is in comments — supported by diff. Def-level items (A2 / A4 / A7 / A8 / A12 canvas gaps + deviations, A5 unranked drop-out under L4) are **unpromoted** — they belong in `2610130449-kura-mock-def.md` § Canvas deviations / Canvas gaps via the orchestrator's absorb step (patch correctly left the def untouched).

**Commit composition** (informational — fix forward, never rewrite history): 3 commits — typed subject 2/2 code commits | type matches diff 2/2 (`2f15a43` `feat(radar)` also edits `internal/web/server.go` + test — scope narrower than diff, same pattern noted in the iteration-1 audit) | `Projex:` trailer 3/3 (`2610130928-kura-browse-page`).
**Violations:** None.

**Trailer survival on `master`:** 0/30 — N/A, not Critical: `master` holds only the baseline import + doc-only `projex(...)` commits (sprint unmerged). Sprint branch: every code commit since iteration 1 carries `Projex:`. Re-check after the sprint lands.

---

## Gap Analysis

### Promised But Not Delivered
None — all S1–S9 delivered.

### Undocumented Issues
| Issue | Severity | Affects |
|---|---|---|
| F1 fragment nav re-renders Browse after any filter change | Medium | Browse users who page with 再看 then hit 登入 / 加入會員 (or any future `#` link) |
| F2 Back restores page 1 only | Low | Browse → detail → Back after paging |
| F5 `max_twd` ≥ 1e21 exponent form | Low | pathological input only |

### Unhandled Edge Cases
- Crawl between page fetches can shift positions — plan-accepted risk (mock runs without crawls).

---

## Quality Assessment

### Completeness: High
**Strengths:** every canvas Browse value transcribed exactly; all three Home/header dead ends closed; site-empty and zero-result states.
**Gaps:** F2 (paging not restorable).

### Correctness: Medium-High
**Works:** counts, paging integrity, filters, sync, history (replace vs push), header search, themes, a11y focus.
**Bugs:** F1 — Medium.

### Code Quality: High
**Positive:** one whitelist, one query builder, guarded async, listeners installed once, comments name rejected alternatives.
**Tech Debt:** router `current` cannot be updated by views — Low (F1 fix removes it).

### Value Delivered: High
**Intended:** last canvas flow not blocked by an open question. **Actual:** delivered on real data; reviewer can now reach every Home entry point. **Impact:** User Positive.

---

## Open Findings

### Undocumented Discoveries
- Exhaustive paging walk confirms stable `sortItems` ordering across offsets for deal / price / newest sorts.
- Brand / 類型 whitelist against *offered* options (executor deviation) also neutralises catalog types with 0 listings (`category=pouch`) — verified live.

### Impact Analysis
- Downstream: advisor, `/legacy/`, Home unaffected (no `offset` → 0; empty-state markup identical).
- Future: a member layer that turns 登入 / 加入會員 into real routes removes the most common F1 trigger but not the bug (detail's `#fair` link shows fragment links are used in-app).
- Risks: F1 likelihood Medium for a reviewer exploring the header after paging.

### Improvements
- Persist shown-count (e.g. `?n=80` read on render, or `history.state`) to make Back land where the user left (F2).

---

## Stakeholder Impact

| Stakeholder | Promised | Reality | Impact |
|---|---|---|---|
| Job-application reviewer | Walk Browse end-to-end on real data | Delivered; F1 can silently drop loaded pages | Positive |
| User (BA) | Canvas look, real data only, shop grades verbatim | Held; no mocked figure, grade rule honoured | Positive |

---

## Findings

### Critical (Must Address)
None.

### Significant (Should Address)
- **F1 — Stale router `current` after `replaceState` (fixable now).** `app.js` sets `current` only in `render()`; `browse.js` `apply` changes the URL with `history.replaceState`, so the `popstate` guard (`location.pathname + location.search !== current`) sees a "new" path on the next fragment navigation and re-renders the whole view. Reproduced live: `/browse?deals=1&sort=deal` → sort 最新上架 → 再看 (80 cards) → click header 登入 (`href="#"`) → grid re-rendered, 40 cards, 「再看 1,431 件」. Also collapses an open mobile panel and resets scroll. → Remediation: let views replace the URL through the router — e.g. a small `nav.js` module (imported by both `app.js` and `browse.js`, no import cycle) holding `current` with `go(url)` and `replaceURL(url)` (`history.replaceState` + `current = pathname + search`); `app.js` popstate compares against it. Re-verify: same repro keeps 80 cards; Back/forward and B9 unchanged. Plan Risks bullet should be marked disproved in the fix's patch doc.

### Minor (Nice to Fix)
- **F2 — Back to Browse returns to page 1 (needs decision).** `renderBrowse` always fetches offset 0; paging is not part of the URL state. → Options: (a) accept for the mock; (b) record shown count in `history.state` via `replaceState` on load-more and re-fetch `limit = n` on render (API caps `limit` at 100 → would need several requests or a higher cap); (c) `?n=` URL param (changes the canonical URL contract). Orchestrator / human call; not a canvas question.
- **F3 — `stateQuery` comment ambiguous (fixable now).** → "Canonical query: fixed key order; defaults omitted, sort=deal among them."
- **F4 — Rule 6 gap in `apply` (fixable now).** → One-line comment on the load-more disable naming the race the seq guard misses.
- **F5 — `max_twd` exponent form (fixable now).** → Bound the digits (e.g. `/^\d{1,9}$/` in `readState`, and slice the stripped digits in `apply`) so the field, URL and API agree.
- **F6 — Def absorb pending (orchestrator, not executor).** A2 heading 全部在售 / 最新上架, A7 mobile panel open state, A8 zero-result copy → def § Canvas gaps (canvas pass 2); A4 類型 option set, A9 model 看全部 → § Canvas deviations; A12 dark × browse built token-driven; A5 unranked listing excluded under any floor → L4 open item. Copy choices (全部在售, 沒有符合條件的包。) are plan-authored and want human / canvas-pass-2 confirmation.

### Positive
- Canvas transcription exact on every value checked against the canvas file itself (not just the plan).
- Executor deviations all improve correctness (offered-option whitelist, IME guard, load-more disable during apply) and were disclosed.
- Real-input verification (typing, Enter, Space) on the cases about input.

---

## Recommendations

**Immediate:** fix F1 + F3 + F4 + F5 in one patch on the sprint branch; re-run the F1 repro, B3, B9, B13.
**Future:** F2 decision; F6 absorb into def / sprint nav.
**Process:** plan Risks that call a known divergence "harmless" should carry a repro step in the verification matrix (here: fragment navigation after a filter change).

---

## Re-audit: F1–F5 fixes (`2d58893`, doc `4bea535`)

**Orchestrator decision on F2:** restore the shown count on Back via `history.state`; URL contract unchanged. F6 stays the orchestrator's logging item.

**Change read:** new `nav.js` owns `current` + `viewState` and every history write (`go`, `replaceURL(url, state)`, `keepViewState`); `app.js` popstate → `moved()` ? `render()` : `keepViewState()`; `render()` → `markRendered()`. `browse.js`: `apply` → `replaceURL(url)` (null state = first page); `loadMore` → `replaceURL(location.href, { shown })` after appending, with a URL guard so a navigation started mid-fetch cannot receive the count; `renderBrowse` → `firstCards(st, shownBefore())` in `API_MAX` (100) chunks; F3 / F4 comments; F5 `readState` `/^[1-9]\d{0,8}$/`, `apply` digits → strip leading zeros → `slice(0, 9)`. Only `nav.js` calls `history.pushState` / `replaceState` (grep). No import cycle (`browse.js` → `nav.js` ← `app.js`).

**Checks** (clean `go build` / `go vet` / `go test -count=1 ./...` all ok; `node --check` on `nav.js`, `browse.js`, `app.js`; worktree `kura.db` served read-only on :8098, no crawl flags; built-in browser, real clicks / typing / keys):

| Item | Re-verification | Status |
|---|---|---|
| F1 desktop 1280 | Original repro: Home link → sort 最新上架 → 再看 (80, state `{shown:80}`) → header 登入 (`#`) → URL `…#`, same grid element, 80 cards, 「再看 1,391 件」, state carried; Back to the non-`#` entry → same grid, 80 | ✓ Fixed |
| F1 mobile 375, panel open | Toggle → panel open → sort 價格 ↑ via mobile select (wide form mirrors `price_asc`) → 再看 (80) → 登入 → same grid, 80 cards, `aria-expanded=true`, panel visible, `scrollWidth 375 = innerWidth` | ✓ Fixed |
| F2 | 再看 → 120 (120 unique) → real click on card 100 (`ダヌーブ`) → browser Back → 120 cards, ids identical in order, 「再看 1,351 件」, no reload (`__probe` intact), scroll restored (≈ 12,100px) | ✓ Fixed |
| F3 | `stateQuery`: "Canonical query: fixed key order; defaults omitted, sort=deal among them." | ✓ |
| F4 | `apply`: comment names the race (new filter fetched at the old grid's length, appended to the old grid) and why the seq check misses it | ✓ |
| F5 | Direct `?max_twd=1000000000000000000000` → field empty, 15,367; typing the same 22 digits + Enter → `?max_twd=100000000`, both forms `100000000` | ✓ |
| Back/forward chain | detail ⇄ Browse ⇄ Home via `history.back/forward` ×5: each step renders the right view + title; Browse restored at 120 both directions | ✓ |
| Detail `#fair` | Click → same view element (no re-render), scrolls to section; Back → same element, `#` removed | ✓ |
| Detail → model → Back → Back → forward ×2 | model title `ダヌーブ モノグラム 型號 · LOUIS VUITTON · Kura 蔵`; Back → detail; Back → Browse (120); forward ×2 → model | ✓ |
| Model 看全部 | Reveals in place (hidden 13 → 2), URL unchanged | ✓ |
| Advisor | Dock opens the dialog; URL unchanged | ✓ |
| Console | onlyErrors → none | ✓ |

**Source hygiene (fix diff):** no projex IDs / `file:ln` / banners. `nav.js` module comment names the constraint (import cycle) and the trap (bare `replaceState` → stale `current`); `API_MAX` names the server constraint (`limit > 100` → 30); F5 comment names the failure (`1e+21`). Commit `2d58893` `fix(web)` matches the diff; `Projex: 2610130952-kura-browse-audit-fixes` + Co-Authored-By present. Plan Risks bullet marked Disproved (`4bea535`).

**Residual observations (informational, no action required):**
- **R1** — Two adjacent entries with the same URL (e.g. `/browse` opened twice from the mobile icon): Back between them does not re-render (`moved()` false), and `keepViewState` copies the newer entry's `shown` onto the older one. Same list either way; benign.
- **R2** — A budget over 9 digits is cut to its first 9 digits (22 typed digits → 100,000,000) rather than rejected. Still far above any listing, so in effect unlimited; field, URL and API agree.
- **R3** — An invalid param in a directly loaded URL stays in the address bar until the first filter change (A1: ignored, control shows default). Behaviour from before this fix.
- **R4** — Restoring a large shown count makes `ceil(n/100)` requests in parallel; bounded by how far the user paged (patch Notes disclose it).

**Still open:** F6 only (def / sprint-nav absorb of A2, A4, A5, A7, A8, A9, A12 — orchestrator).

---

## Final Verdict

**Status:** Accept

**Overall Assessment:**
- Completeness: High
- Correctness: High
- Quality: High
- Value: High

**Conditions:**
- [x] F1 fixed and re-verified (fragment nav after filter change + load-more keeps the grid) — `2d58893`, desktop + 375 with panel open
- [x] F3 / F4 comment fixes (F5 recommended in the same patch) — all three in `2d58893`
- [x] F2 + F6 routed to the orchestrator (decision / def absorb), not blocking — F2 decided (`history.state`) and fixed; F6 with orchestrator

**Sign-off:** Yes — S1–S9 independently verified; F1–F5 fixed and re-verified live with no router regression across Home / Browse / detail / model / advisor and Back / forward chains. F6 (def absorb) remains the orchestrator's item.
