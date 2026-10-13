# Audit: Kura listing detail page (商品頁)

> **Status:** Complete (Accepted)
> **Sprint:** 2610130524-kura-interactive-mock-sprint.md — iteration 2
> **Author:** Claude Opus 5.5 (auditor, sprint subagent)
> **Parent:** 2610130634-kura-listing-detail-patch.md
> **Related Projex:** 2610130624-kura-listing-detail-plan.md (audited plan) | 2610130634-kura-listing-detail-patch.md (execution claims) | 2610130524-kura-interactive-mock-sprint.md | 2610130449-kura-mock-def.md (data semantics, Canvas gaps) | 2610130602-kura-canvas-foundation-home-audit.md (iteration 1 audit; F1/F6 still open there) | 2610130705-kura-listing-detail-audit-polish-patch.md (F1 F4 F5 fix, F2 comment)
> **Audit Date:** 2026-10-13 | **Work Period:** 2026-10-13, commits `766218d` `b12c30a` `596bc28` on `projex/sprint/2610130524-kura-interactive-mock` (range `9c562fb..596bc28`); re-audit 2026-10-13 of polish fix `1dab0e9` + doc `90acc12` (range `891c2c9..90acc12`)
> **Subject:** plan S1–S11 vs code, tests, live app (copy of worktree `kura.db` outside repo + mutated copy for sold/multi-point), canvas `project/Main.dc.html` detail markup + `renderVals`

---

## Audit Summary

**Claim:** `/listing/{id}` renders canvas 商品頁 on real data (crumb, mat, info/CTA/track, 鑑定卡 + ladder, 商品資料, 行情估價 prose per state, 同型號價格分布 chart ↔ table, 售價變化 strip / step line); grade-stat + history TWD computed once in `radar`, read by advisor + UI; listing API returns `scheme`; 404 → not-found view; vet/test green; S1–S11 ✓.

**Verdict:** Verified — every sampled S1–S11 claim reproduced independently, incl. an extra state combination the executor did not run (sold **and** multi-point). No Critical or Significant defect. Findings are Minor polish/latent items + one documented, unfixable S1 edge.
**Re-audit (2026-10-13):** F1 F4 F5 fixed and independently measured; F2 Deferred (trap comment landed); F3 Accepted by orchestrator; no regression on detail or Home (§ Re-audit). Nothing actionable open.

**Assessment:** Completeness: High | Correctness: High | Quality: High | Value: High

**Top Issues:**
1. Distribution at 375 with real density: fixed ±26px jitter in ~60px columns → adjacent grade columns' dots touch/overlap (ランクA right edge vs ランクB left edge 7.8px apart, 10px dots). **Fixable now** (clamp jitter to column width).
2. `dist.js` column building is single-source-safe only: points match by grade *code* across all sources while the ladder/stat are this listing's shop. Latent until a second shop; conflicts with "keep each shop's own grading". **Defer to multi-provider iteration** (already in sprint Deferred).
3. S1 malformed-escape direct load (`/listing/%E0%A4%A`) → Go `net/http` 400 before any handler. Documented by executor; not fixable at app level. **Needs criterion wording accepted** (orchestrator can accept; no code change).

---

## Claims vs Evidence

| Claim | Evidence | Status | Notes |
|---|---|---|---|
| `go vet` / `go test` green | Re-run `go vet ./...` clean; `go test -count=1 ./...` ok advisor/radar/brandoff/store/web; `-count=2 ./internal/web/` ok (global `RegisterScheme("wt")` idempotent — map assign) | ✓ | `gofmt -l` lists files only due to CRLF working copy (`core.autocrlf=true`); committed blobs of all touched Go files gofmt-clean |
| Hygiene grep none | Re-run plan Step 7.1 regex over `internal cmd` → exit 1 (none) | ✓ | |
| Server TWD fields | `radar.PricePoint`, `GradeStat.*TWD` set in `stats()` from same `med/lo/hi`; `toItem` loop; advisor `modelStats`/`getListing` read fields (diff `766218d`) | ✓ | `by_grade` = active listings only (`stats()` branch) — matches 「中位數（在售）」 |
| Listing API `scheme` + 404 | `TestListingAPI` reads scheme/short/history TWD/by_grade TWD + 404; live `/api/listings/brandoff:nope` → 404 | ✓ | |
| Legacy key rename | All 5 `At`/`PriceJPY` sites renamed; no remaining readers of old keys in `static/` | ✓ | Legacy still converts JPY→TWD client-side (pre-existing, out of scope) |
| Patch line refs | Spot-checked `condition.go:15`, `radar.go:53/107/219/574`, `tools.go:250/281`, `server.go:200`, `server_test.go:60`, `app.js:16`, `kura.css:95` → all land on claimed code | ✓ | |
| D1 (same-grade/high) | Live: prose, dl, ladder aria 「店家等級：新品、未使用品、ランクA、ランクB（此件）、ランクC」, ladder text 新品/未使用/A/B/C, stamp, 25 dots = 25 items, 2 medians, table ランクA `8 · NT$31,285 · 16,147 · 38,349` / ランクB `17 · 20,184 · 10,092 · 26,239` = `/api/models/LOUIS VUITTON/M51252` `by_grade` exactly; title 「カルトシエールGM · LOUIS VUITTON · Kura 蔵」 | ✓ | Copy strings checked against canvas detail markup: fine print, 鑑定卡/店鑑, 行情估價 sentence form, 上架以來… note, legend 中位數（在售）, CTA 到 BRAND OFF 商品頁 |
| D6 (409-dot group) | 375 dark: 409 dots = 409 items, axis ランクA 在售 289 / ランクB 在售 120, scrollWidth 375, no `main a/button` < 24px tall, h1 26px, bg `rgb(27,24,21)` | ✓ | Legibility: F1 |
| D10 Home → card → track → back | In-app nav (same `<main>`), title set, detail track → `aria-pressed=true` text 追蹤中 → back → card `aria-pressed=true`, label 追蹤中，點一下取消 | ✓ | |
| D11 not-found | `/listing/brandoff%3Anope` in-app → 找不到這個頁面 view + title; `%E0%A4%A` in-app → not-found; direct `%E0%A4` → 200 shell; direct `%E0%A4%A` → **400 text/plain** | ⚠ | F3 — server-level, documented in patch Notes |
| D12 sold + multi-point | Own mutated copy: (a) `2101220216016` sold **with** 2 points: 最後售價 NT$9,083 ¥45,000, 已找到新主人, no badge/track/#fair, CTA kept, fine print 2nd sentence only, 上架 row date only, step path `M0.00 14.29 H76.54 V85.71 H100`, end label = gone date 10/12, sr list 2 items, dist dot 「此件 · 已售出」; (b) `2101220298944` active price **rise** at 375 dark: path `M0.00 85.71 H69.35 V14.29 H100`, labels inside 18–351px, scrollWidth 375 | ✓ | Executor tested sold-single and drop-active separately; combination + rise now also covered |
| S6 interactions | Hover dot → tip 「ランクA · 在售 NT$36,331」, mouseout hides; toggle → chart hidden/table shown, 以圖表檢視, table box 343 scrolls 560 inside, page 375 | ✓ | |
| S9 focus | Real Tab presses → CTA `:focus-visible` outline 2px solid; `#fair` click scrolls, same `h1` node (no re-render) | ✓ | Patch's Tab-order list omits the 行情 link that precedes CTA when `fair` exists — doc imprecision only |
| S10 banned strings | Page text on D1, sold+multi: no 售出率 / 天找到新主人 / 已售出 N 件 | ✓ | Model `sold`/`sell_through` exist in API but not rendered |
| Worktree / DB untouched | `git status` clean; `kura.db` + `-wal` mtime 05:25 (pre-run) | ✓ | Audit served a copy outside repo; copy deleted |

---

## Objective Verification

### Steps 1–2 (server)
**Evidence:** diff `766218d`, tests, live API. **Actual:** as planned verbatim; extra gofmt realignment of `Item` tags (whitespace). **Verification:** ✓ Verified.

### Steps 3–6 (client)
**Evidence:** diff `b12c30a`, `detail.js`, `dist.js`, live DOM. **Actual:** plan markup + copy table implemented; deviations 2–4 (edge-aligned history labels, extra CSS, 鑑定卡 foot omitted when empty) are justified and within A11/S9. **Missing:** none vs plan. **Verification:** ✓ Verified.

### Step 7 (verification)
**Evidence:** patch D1–D12 table; this audit re-ran D1, D6, D10, D11, D12 (+ combination). **Verification:** ✓ Verified (claims reproduced).

---

## Code/Implementation Inspection

### `internal/web/static/js/dist.js`
**Claimed:** canvas `kuraChart` geometry on real points; server medians/table. **Actual:** matches — Quality: High.
**Issues Found:**
- Fixed `OFFS = [0,-13,13,-26,26]` px regardless of column width; at 375 (plot 299px / 5 cols ≈ 60px) neighbouring columns' outer jitter lanes overlap — Severity: Low (F1)
- `cols` from this listing's `scheme`; `col.pts` filters items by `condition.code` across every source; `col.stat` restricted to `scheme.source` → with a second shop sharing a code, its dots land under this shop's ladder/definition while median/table exclude them — Severity: Low, latent (F2)
- `.k-dist__tip role="status"` sits inside `role="img"` (children presentational) → live region likely never announced; harmless since dots are hover-only — Severity: Info

### `internal/web/static/js/detail.js`
**Actual:** Quality: High. Every interpolation through `esc`/`twd`; group fetch failure tolerated; sold/multi built from card vocabulary (A5/A6).
**Issues Found:** `CONF[f.confidence]` has no fallback (renders `undefined` for an unknown confidence value; `CONF_DD`/`METHOD` do fall back) — Severity: Info.

### `internal/web/static/js/app.js`
**Actual:** path params, returned title, 404 mapping, title inside `mine === seq` guard. Any 404 thrown by a view maps to not-found — correct today (detail's only secondary fetch is `.catch`ed). Quality: High.

### Server (`radar`, `advisor`, `web`, `listing`, `brandoff`)
**Actual:** Quality: High. **Undocumented:** model-group items still carry full `history` (only search strips it) — payload grows slightly with TWD per point; covered by A10 acceptance.

---

## Testing Validation

**Coverage:** `TestDropsAndStats` (TWD stats + history), `TestListingAPI` (scheme, TWD, 404), `TestStaticRoutes`. Client: no automated tests (no build step per HANDOFF §3.6) — browser matrix is the evidence.
**Execution:** All pass, `-count=2` stable. **Flaky:** none seen.
**Missing:** no test for grade `Short` fallback (empty → label) — trivially covered by client `g.short || g.label`.

---

## Documentation Audit

**Accuracy:** patch doc matches implementation; line refs correct. Two small inaccuracies: (a) deviation 5 says the doc commit carries trailers — `596bc28` has no `Projex:` trailer (doc-only, exempt, so harmless); (b) Tab-order list omits the 行情 link. Plan header/criteria correctly marked `[PATCHED]`/✓.
**Canvas read:** artifact file listing reports live `1791840357-ac34` (unchanged from sprint nav); the file read itself was labelled `1791845608-ba84`. Detail copy and `renderVals` values (`crumbPad`, `detailTop`, `heroIcon`, `detailH1`, `detailGap`, `specCols`, `chartPad`, 880px column) matched implementation. Pin 1791839497-46ec still not directly readable.

---

## Source Hygiene

**Scope:** Diff only (`9c562fb..596bc28`, `.projex/**` excluded)

| Rule | Location (symbol) | Quote | Severity |
|------|-------------------|-------|----------|
| 3 | `kura.css` → distribution block header | "同型號價格分布 (shared with the model page)" | Low — model page does not exist yet; true after next iteration |

**Rule 1/2/4:** none (no projex refs, no `file:ln`, no step banners).
**Rule 5:** `kura.css` `[hidden]` "!important: component rules set display, which would otherwise beat the attribute" names the trap ✓.

**Rule 6 gaps:**
- `dist.js` → `distHTML` — `s.source === scheme.source` filter on stats with no comment; the cross-source asymmetry vs `col.pts` (F2) is exactly the trap a reader needs named.
- `detail.js` → `historyHTML` — `Math.min(100, …)` clamp in `x` (an observation after `gone_at`) unexplained — Low.

**Rule 6 vacuity:** none — `detail.js` header (sold keeps CTA, drops valuation/track), `dist.js` header (spans vs buttons → tab stops; server medians vs recompute), document-listener comment (stacked listeners → double flip), `card.js` text-vs-aria-label, `app.js` stale-view guard all name the alternative/trap.

**Rationale promotion:** none needed — load-bearing rationale (one source per stat, hover-only dots, sold variant) sits in code comments; plan-only assumptions A4–A7 are orchestrator log items, not code rationale.

**Commit composition** (informational — fix forward, never rewrite history): 2 code commits checked — typed subject 2/2 | type matches diff 2/2 (`766218d` scope `radar,web` also touches advisor/listing/brandoff/legacy — scope narrow, type correct) | `Projex:` trailer 2/2 (`766218d` trailer after `Co-Authored-By`, still parses). `596bc28` doc-only, exempt.
**Violations:** None actionable.

**Trailer survival on `master`:** 0/6 — N/A (unchanged since iteration 1 audit): all baseline import + doc-only commits; sprint close must carry `Projex:` in its landing message.

---

## Gap Analysis

### Promised But Not Delivered
| Promise | Status | Impact |
|---------|--------|--------|
| S1 "malformed encoding → not-found view" | Partial — direct load of invalid escape gets Go 400 | Low |

### Undocumented Issues
| Issue | Severity | Affects |
|-------|----------|---------|
| F1 cross-column dot overlap at 375 | Low | mobile distribution legibility |
| F2 cross-source grade-code merge in `distHTML` | Low (latent) | multi-provider iteration, model page reuse |

### Unhandled Edge Cases
- Unknown `fair.confidence` value → 「行情 NT$… · undefined」 — Impact: none today (radar emits high/medium/low only).
- Sold + single point keeps forward-looking note 「店家調價時，會畫在這條線上。」 — executor already flagged for canvas pass 2.

---

## Quality Assessment

### Completeness: High
**Strengths:** all 12 matrix cases + mobile/dark + a11y checks; undesigned states built strictly from existing vocabulary and logged as assumptions.
**Gaps:** F3 edge only.

### Correctness: High
**Works:** routing/404/title race guard: Yes | track sync: Yes | TWD single source: Yes | sold/multi states: Yes.
**Bugs:** none functional.

### Code Quality: High
**Positive:** small view modules, escaped interpolation, document-level listeners, server-owned numbers.
**Tech Debt:** fixed jitter (F1) — Low; single-source assumption in `distHTML` (F2) — Low.

### Value Delivered: High
**Intended:** close Home's dead-end click; canvas 商品頁 on real data; reusable distribution. **Actual:** delivered. **Impact:** User: Positive.

---

## Open Findings

### Undocumented Discoveries
- Changes: gofmt tag realignment in `radar.Item` (whitespace, noted in patch).
- Workarounds: executor's scratch-DB date shift (deviation 1) — acceptable, reproduced here with own shifts.

### Impact Analysis
- Downstream: model page reuses `distHTML` with `thisId` null — inherits F1/F2.
- Future: Enabled: model page, advisor answers quoting identical TWD figures. Blocked: none.
- Risks: model-group payload (A10) on hosting — Likelihood: Medium once hosted.

### Improvements
- Could be better: jitter lanes ∝ column width; per-source column keys `{source, code}` once a second shop exists.

---

## Stakeholder Impact

| Stakeholder | Promised | Reality | Impact |
|-------------|----------|---------|--------|
| Reviewer (104 job) | canvas detail flow on real data | delivered, both themes, both widths | Positive |
| Model iteration | reusable `dist.js`, model URL A1 | delivered; F1/F2 carried over | Neutral |

---

## Findings

### Critical (Must Address)
- None.

### Significant (Should Address)
- None.

### Minor (Nice to Fix)
- **F1 — Dot jitter overlaps neighbour columns at 375** (fixable now) → scale `OFFS` by column width (e.g. cap |offset| at `colWidth/2 − 7px`, computed in CSS via `calc()` with a per-column `--w` or in JS from `cols.length`); canvas sample unchanged at 1280. — **[PATCHED]** 2610130705-kura-listing-detail-audit-polish-patch.md (`1dab0e9`): lanes scale to `clamp(0px, (50/cols)% − 8px, 26px)`; 375 halo gap 2px / mark gap 6px; 1280 max centre delta vs old 0.01px.
- **F2 — `distHTML` merges grades across sources by code** (latent; defer) → key columns by `{source, code}` when the multi-provider iteration lands; add a Rule-6 comment on the `scheme.source` filter now if touching the file. — **Deferred** (multi-provider iteration); Rule-6 comment landed in `1dab0e9`.
- **F3 — S1 malformed-escape direct load → Go 400** (needs orchestrator acceptance, no code) → accept S1 as "every URL the server accepts"; executor already documented. — **Accepted** by orchestrator: S1 holds for every URL the server accepts.
- **F4 — Hygiene Low:** `kura.css` "(shared with the model page)" forward reference; `historyHTML` clamp without rationale (fixable now, or resolves itself when model page lands). — **[PATCHED]** `1dab0e9`: header present tense; clamp comment names rejected stretch of `t1`.
- **F5 — `CONF` lacks fallback** (fixable now) → `CONF[f.confidence] ?? ""`. — **[PATCHED]** `1dab0e9`: `?? esc(f.confidence)` (matches `CONF_DD`/`METHOD`).

### Positive
- Executor re-verified table numbers against API to the dollar; anticipated history-label clipping at 375 and fixed it (deviation 2); honest deviation log incl. trailer-order flaw.
- Server-side TWD unification removed the advisor's duplicate conversion — real S10 improvement, not just UI.

---

## Recommendations

**Immediate:** ~~optional polish patch for F1 + F5 (+ F4 clamp comment)~~ — done via 2610130705-kura-listing-detail-audit-polish-patch.md, re-audited.
**Future:** F2 with multi-provider work; orchestrator logs plan A4–A7 + sold-single note copy (patch Notes) into def § Canvas gaps / canvas pass 2 review.
**Process:** state-combination cases (sold × multi-point) belong in the matrix when two undesigned states can co-occur.

---

## Final Verdict

**Status:** Accept

**Overall Assessment:**
- Completeness: High
- Correctness: High
- Quality: High
- Value: High

**Conditions:**
- [x] (non-blocking) Orchestrator accepts F3 wording; F1/F5 polish at orchestrator discretion — F3 accepted; F1/F4/F5 via 2610130705-kura-listing-detail-audit-polish-patch.md — re-audit confirmed

**Sign-off:** Yes — all S1–S11 reproduced; F1/F4/F5 fixed and measured; F2 deferred to multi-provider iteration (sprint Deferred); F3 accepted. Does not block the model-page iteration.

---

## Re-audit (2026-10-13)

**Subject:** 2610130705-kura-listing-detail-audit-polish-patch.md — code `1dab0e9` (`dist.js`, `detail.js`, `kura.css`; 30 lines), doc `90acc12` (patch doc + this audit's finding marks).
**Method:** diff read; `go vet ./...` clean; `go test -count=1 ./...` ok (advisor, radar, brandoff, store, web); fresh binary on a copy of `kura.db` outside the repo; built-in browser; own measurement script (dot centres from `getBoundingClientRect`, column = plot width / axis cells, halo r7 / mark r5, 此件 halo r9 / mark r7); old-formula comparison parses lane + `cx` from each dot's raw `style`.

| Finding | Mark | Evidence | Earned? |
|---|---|---|---|
| F1 jitter overlap | [PATCHED] | 375×812 dark, CHANEL マトラッセ (`brandoff:2104102463379`, 409 dots): plot 299 / 5 cols = 59.8px; ランクA│ランクB outer halo gap **2px**, mark gap **6px** (was 7.8px centre gap → overlap); every regular dot's halo inside its own column (0 out); `scrollWidth` 375; screenshot shows two separated clouds. 1280×900 light: plot 836 (167.2px cols), halo gap 101.2px, **max |new − old centre| 0.01px** over 409 dots, 0 parse failures → desktop/canvas look unchanged | ✓ |
| F4 hygiene | [PATCHED] | `kura.css` header now "(markup from dist.js distHTML)" — present tense; `historyHTML` clamp comment names the rejected alternative (stretch `t1` → line end leaves the day its label names); hygiene grep over touched files: none | ✓ |
| F5 confidence fallback | [PATCHED] | `CONF[f.confidence] ?? esc(f.confidence)`; `esc` maps `undefined` → `""`; same pattern as `CONF_DD`/`METHOD`; live 「行情 NT$20,184 · 高信心」 known-value path intact | ✓ |
| F2 cross-source merge | Deferred | Rule-6 comment on `col.stat` filter names the trap (another shop's "A" under this ladder; median/table exclude it) and the `{source, code}` fix; behaviour unchanged as intended | ✓ |
| F3 malformed escape | Accepted | Orchestrator decision; no code | ✓ |

**Regression (detail, 1280 light, `brandoff:2101220216016`):** title, 行情 link, prose, ladder 新品/未使用/A/B/C, 25 dots, table ランクA `8 · 31,285 · 16,147 · 38,349` / ランクB `17 · 20,184 · 10,092 · 26,239` (unchanged), tooltip 「ランクA · 在售 NT$16,147」, chart ↔ table toggle, history strip, track → 追蹤中. **Home:** in-app crumb → `/`, title 「Kura 蔵 · 行情雷達」, shelves 最新上架 | 價格最美, 8 cards, footer FX line, detail-tracked listing's card `aria-pressed=true`, card track toggles. Console errors: none.

**Residual (Info, not reopened):**
- 此件 dot (r9 halo) on an outermost lane crosses its column edge by ~1px at 375 (measured: 1 dot, ランクB → empty ランクC lane); marks stay clear. Acknowledged in patch Notes — acceptable.
- Formatting slips introduced by the patch: `const x =(t) =>` (`historyHTML`) and `col.stat =stats.by_grade…` (`distHTML`) — missing space after `=`; no JS formatter in repo. Fix forward on next touch.
- Commit composition: `1dab0e9` `fix(web)` typed, matches diff, `Projex: 2610130705-kura-listing-detail-audit-polish`, `Co-Authored-By` last ✓. `90acc12` doc-only, exempt.
- Patch agent edited this audit's finding marks (`90acc12`); marks verified above as earned. Status/verdict lines set by auditor in this re-audit.

**Worktree:** clean; `kura.db` + `-wal` mtime 05:25 unchanged; scratch copy + binary deleted; viewport reset, tab closed.
